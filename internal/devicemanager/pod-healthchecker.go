package devicemanager

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func (dm *DeviceManager) runGarbageCollector() {
	for {
		err := dm.garbageCollectPodQuotas()
		if err != nil {
			fmt.Printf("Error during garbage collection: %v", err)
		}

		time.Sleep(30 * time.Second)
	}
}

func (dm *DeviceManager) garbageCollectPodQuotas() error {
	for _, device := range dm.devices {
		device.mut.Lock()

		pods := []string{}
		for _, pod := range device.Pods {
			pods = append(pods, pod.Id)
		}

		runningPods, err := getRunningPods(pods)
		if err != nil {
			return err
		}

		for _, podId := range pods {
			if !runningPods[podId] {
				dm.unreservePodQuota(device.Id, podId)
			}
		}

		device.mut.Unlock()
	}

	return nil
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
