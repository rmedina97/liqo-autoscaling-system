package util

import (
	"fmt"
	"log"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var clusterMap = map[string]Cluster{
	"remote": {
		Name:       "remote",
		Kubeconfig: "provider.kubeconfig",
		Resources: v1.ResourceList{
			v1.ResourceCPU:    *resource.NewQuantity(1000, resource.DecimalSI),
			v1.ResourceMemory: *resource.NewQuantity(2048, resource.DecimalSI),
			v1.ResourcePods:   *resource.NewQuantity(1, resource.DecimalSI), //test per scegliere il secondo
		}, // Example resources
	},
	"remote2": {
		Name:       "remote2",
		Kubeconfig: "kubeconfig2",
		Resources: v1.ResourceList{
			v1.ResourceCPU:    *resource.NewQuantity(1000, resource.DecimalSI),
			v1.ResourceMemory: *resource.NewQuantity(2048, resource.DecimalSI),
			v1.ResourcePods:   *resource.NewQuantity(10, resource.DecimalSI)},
	},
}

type Cluster struct {
	Name       string          `json:"name"`
	Kubeconfig string          `json:"kubeconfig"`
	Resources  v1.ResourceList `json:"resources"`
}

func ReturnList() (map[string]Cluster, error) {
	// No existing Nodegroups
	if len(clusterMap) == 0 {
		return nil, fmt.Errorf("no clusters found")
	} else {
		return clusterMap, nil
	}
}

func UpdateList(name string, kubeconfig string, resources int) error {
	// Update the list of remote clusters
	if name == "" || kubeconfig == "" || resources <= 0 {
		return fmt.Errorf("invalid parameters for updating cluster list")
	}

	// Simulate updating the cluster map
	log.Printf("kubeconfig: %s, resources: %d", kubeconfig, resources)
	return nil
}
