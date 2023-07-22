package devicemanager

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	DeviceCleanupPeriod = 1 * time.Minute
)

func (dm *DeviceManager) runGarbageCollector() {
	for {
		wg := &sync.WaitGroup{}
		wg.Add(2)

		dm.garbageCollectPodQuotas(wg)
		dm.garbageCollectDevices(wg)

		wg.Wait()

		time.Sleep(2 * time.Second)
	}
}

func (dm *DeviceManager) garbageCollectDevices(wg *sync.WaitGroup) {
	defer wg.Done()

	runningAllocators, err := getRunningPods(PodTypeAllocator)
	if err != nil {
		log.Printf("[GC] Error getting running pods: %v", err)
		return
	}

	dm.lock.Lock()
	defer dm.lock.Unlock()

	for _, device := range dm.devices {
		if pod, ok := runningAllocators[device.AllocatorPodId]; !ok {
			log.Printf("[GC] Allocator %s is not running, removing device %s", device.AllocatorPodId, device.Id)
			dm.deregisterDevice(device.Id)
		} else if len(device.Pods) == 0 && time.Now().After(device.LastUsedAt.Add(DeviceCleanupPeriod)) {
			log.Printf("[GC] Device %s has not been used for 5 minutes, removing it", device.Id)
			err := dm.deleteAllocatorDeployment(pod)
			if err != nil {
				log.Printf("[GC] Error deleting allocator deployment %s: %v", device.AllocatorPodId, err)
			}
		}
	}
}

func (dm *DeviceManager) deleteAllocatorDeployment(allocatorPod *v1.Pod) error {
	deploymentName, ok := allocatorPod.Labels["app"]
	if !ok {
		return fmt.Errorf("allocator pod %s does not have a deployment name", allocatorPod.Name)
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	err = clientset.AppsV1().Deployments("default").Delete(context.Background(), deploymentName, metav1.DeleteOptions{})
	return err
}

func (dm *DeviceManager) deregisterDevice(deviceId string) {
	if device, ok := dm.devices[deviceId]; ok {
		device.lock.Lock()
		defer device.lock.Unlock()

		device.sch.Stop()
		delete(dm.devices, deviceId)
	}
}

func (dm *DeviceManager) garbageCollectPodQuotas(wg *sync.WaitGroup) {
	defer wg.Done()

	runningPods, err := getRunningPods(PodTypeClient)
	if err != nil {
		log.Printf("[GC] Error getting running pods: %v", err)
		return
	}

	dm.lock.RLock()
	defer dm.lock.RUnlock()

	for _, device := range dm.devices {
		device.lock.Lock()
		for podId := range device.Pods {
			if _, ok := runningPods[podId]; !ok {
				log.Printf("[GC] Pod %s is not running, removing it from device %s", podId, device.Id)
				dm.unreservePodQuota(device.Id, podId)
			}
		}
		device.lock.Unlock()
	}
}

func (dm *DeviceManager) unreservePodQuota(deviceId, podId string) {
	device := dm.devices[deviceId]
	if device == nil {
		return
	}

	if !device.Pods[podId] {
		return
	}

	device.sch.UnreservePodQuota(podId)
	device.mm.UnreservePodQuota(podId)
	delete(device.Pods, podId)

	if len(device.Pods) == 0 {
		device.LastUsedAt = time.Now()
	}
}

type ShareDevPodType string

const (
	PodTypeClient    ShareDevPodType = "client"
	PodTypeAllocator ShareDevPodType = "allocator"
)

func getRunningPods(podType ShareDevPodType) (map[string]*v1.Pod, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		// skip eviction if not running in a k8s cluster
		return nil, nil
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	// TODO: Refactor all the code to use namespaces for Pods
	pods, err := clientset.CoreV1().Pods("default").List(
		context.Background(),
		metav1.ListOptions{LabelSelector: fmt.Sprintf("sharedev=%s", podType)},
	)
	if err != nil {
		return nil, err
	}

	runningPods := make(map[string]*v1.Pod)
	for _, pod := range pods.Items {
		if pod.Status.Phase == "Running" || pod.Status.Phase == "Pending" {
			runningPods[pod.Name] = &pod
		}
	}
	return runningPods, nil
}
