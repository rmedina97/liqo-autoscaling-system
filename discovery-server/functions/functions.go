package util

import (
	"fmt"
	"log"

	watcher "discovery_server/watcher"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Secret struct {
	Data map[string]string `json:"data"`
}

func ReturnList() ([]watcher.Cluster, error) {
	// No existing Nodegroups

	clusterList := watcher.GetClusters()
	if len(clusterList) == 0 {
		return nil, fmt.Errorf("no clusters found")
	} else {
		return clusterList, nil
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

func CreateKubernetesClient() (*kubernetes.Clientset, error) {
	// Client clientset for core resources
	//cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Errore caricando kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return clientset, nil

}
