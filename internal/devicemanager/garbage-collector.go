package devicemanager

import (
	"context"
	"fmt"
	"log"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func (dm *DeviceManager) runGarbageCollector() {
	for {
		dm.garbageCollectPodQuotas()
		time.Sleep(2 * time.Second)
	}
}

func (dm *DeviceManager) garbageCollectPodQuotas() {
	for _, device := range dm.devices {
		device.mut.RLock()
		pods := []string{}
		for _, pod := range device.Pods {
			pods = append(pods, pod.Id)
		}
		device.mut.RUnlock()

		runningPods, err := getRunningPods(pods)
		if err != nil {
			log.Printf("[GC] Error getting running pods: %v", err)
			continue
		}

		for _, podId := range pods {
			if !runningPods[podId] {
				log.Printf("[GC] Pod %s is not running, removing it from device %s", podId, device.Id)
				dm.unreservePodQuota(device.Id, podId)
			}
		}
	}
}

func (dm *DeviceManager) unreservePodQuota(deviceId, podId string) {
	device := dm.devices[deviceId]
	if device == nil {
		return
	}

	device.mut.Lock()
	defer device.mut.Unlock()

	pod := device.Pods[podId]
	if pod == nil {
		return
	}

	dm.schedulerPerDevice[deviceId].UnreservePodQuota(podId)
	device.MemoryBUsed -= pod.MemoryBUsed
	delete(device.Pods, podId)
}

func getRunningPods(podName []string) (map[string]bool, error) {
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
		metav1.ListOptions{LabelSelector: "sharedev=true"},
	)
	if err != nil {
		return nil, err
	}

	runningPods := make(map[string]bool)
	for _, pod := range pods.Items {
		runningPods[pod.Name] = pod.Status.Phase == "Running" || pod.Status.Phase == "Pending"
	}
	return runningPods, nil
}
