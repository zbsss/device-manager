package scheduler

import (
	"context"
	"fmt"

	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func evictPod(podName string, namespace string) error {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		// skip eviction if not running in a k8s cluster
		return nil
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	// set eviction policy
	eviction := &policyv1.Eviction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: "default",
		},
	}

	// evict pod
	err = clientset.PolicyV1().Evictions(eviction.Namespace).Evict(context.Background(), eviction)
	if err != nil {
		return fmt.Errorf("failed to evict pod: %w", err)
	}

	return nil
}
