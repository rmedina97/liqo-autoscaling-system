package util

import (
	"fmt"
)

var clusterMap = map[string]string{
	//metti cluster1,kubeconfig,resources?
}

func ReturnList() (map[string]string, error) {
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
	clusterMap[name] = fmt.Sprintf("kubeconfig: %s, resources: %d", kubeconfig, resources)
	return nil
}
