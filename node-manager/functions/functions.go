package functions

import (
	"context"
	"fmt"
	"log"
	types "node-manager/types"
	"regexp"
	"strings"
	"time"

	util "node-manager/util"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

// key is id nodegroup, value is nodegroup
// TODO change from hardcoded data to a dynamic one
var mapNodegroup = map[string]types.Nodegroup{
	"STANDARD": {
		Id:          "STANDARD",
		MaxSize:     1,
		MinSize:     0,
		CurrentSize: 0,
		//Nodes:       []string{"rmedina"},
		Nodes: []string{},
	},
	"GPU": {
		Id:          "GPU",
		MaxSize:     0,
		MinSize:     0,
		CurrentSize: 0,
		Nodes:       []string{},
	},
}

// flag for cache
var nodegroupIdsCached = false

// Node list
var nodeMinInfoList []types.NodeMinInfo = make([]types.NodeMinInfo, 0, 20)

// Nodegroup list
var nodegroupListMinInfo []types.NodegroupMinInfo = make([]types.NodegroupMinInfo, 0, 5)

// key is id node, value is node
// TODO change from hardcoded data to a dynamic one
var mapNode = map[string]types.Node{
	"rmedina": {
		Id: "rmedina",
		//NodegroupId:    "STANDARD",
		NodegroupId:    "",
		InstanceStatus: types.InstanceStatus{InstanceState: 1, InstanceErrorInfo: ""},
	},
}

var mapRemoteClusters = make(map[string]string)

// Template map
var mapNodegroupTemplate = map[string]types.NodegroupTemplate{
	"GPU": {
		NodegroupId: "GPU",
		Resources: types.ResourceRange{
			Min: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("3.5"),
				corev1.ResourceMemory: *resource.NewQuantity(4*1024*1024*1024, resource.BinarySI),
				corev1.ResourcePods:   *resource.NewQuantity(110, resource.DecimalSI),
				"nvidia.com/gpu":      *resource.NewQuantity(1, resource.DecimalSI),
			},
			Max: corev1.ResourceList{
				//corev1.ResourceCPU:    *resource.NewQuantity(4, resource.DecimalSI),
				corev1.ResourceCPU:    resource.MustParse("3.5"),
				corev1.ResourceMemory: *resource.NewQuantity(8*1024*1024*1024, resource.BinarySI),
				corev1.ResourcePods:   *resource.NewQuantity(110, resource.DecimalSI),
				"nvidia.com/gpu":      *resource.NewQuantity(4, resource.DecimalSI),
			},
		},
		Labels: map[string]string{
			"country":  "france",
			"provider": "liqo",
			"city":     "paris",
		},
		Cost: 25.0,
	},
	"STANDARD": {
		NodegroupId: "STANDARD",
		Resources: types.ResourceRange{
			Min: corev1.ResourceList{
				corev1.ResourceCPU:    *resource.NewQuantity(2, resource.DecimalSI),
				corev1.ResourceMemory: *resource.NewQuantity(2*1024*1024*1024, resource.BinarySI),
				corev1.ResourcePods:   *resource.NewQuantity(110, resource.DecimalSI),
				"nvidia.com/gpu":      *resource.NewQuantity(0, resource.DecimalSI),
			},
			Max: corev1.ResourceList{
				corev1.ResourceCPU:    *resource.NewQuantity(2, resource.DecimalSI),
				corev1.ResourceMemory: *resource.NewQuantity(3*1024*1024*1024, resource.BinarySI),
				corev1.ResourcePods:   *resource.NewQuantity(110, resource.DecimalSI),
				"nvidia.com/gpu":      *resource.NewQuantity(0, resource.DecimalSI),
			},
		},
		Labels: map[string]string{
			"country":  "italy",
			"provider": "liqo",
			"city":     "turin",
		},
		Cost: 1.0,
	},
}

// List of function inside the handle connection -------------------------------------------------------

// getAllNodegroups get all the nodegroups
func GetAllNodegroups() ([]types.NodegroupMinInfo, error) {

	// No existing Nodegroups
	if len(mapNodegroup) == 0 {
		return nil, nil
	} else {
		// Check if it is cached
		// TODO Design decision required to determine where these informations should be stored
		if !nodegroupIdsCached {
			for _, nodegroup := range mapNodegroup {
				nodegroupMinInfo := types.NodegroupMinInfo{Id: nodegroup.Id, MaxSize: nodegroup.MaxSize, MinSize: nodegroup.MinSize}
				nodegroupListMinInfo = append(nodegroupListMinInfo, nodegroupMinInfo)
			}
			nodegroupIdsCached = true
		}
		return nodegroupListMinInfo, nil
	}
}

// nodegroupForNode get the nodegroup for a specific node
func GetNodegroupForNode(id string) (*types.NodegroupMinInfo, error) {
	node, exist := mapNode[id]
	if !exist {
		return nil, nil //TODO change with error
	}
	if node.NodegroupId != "" { // Constraint imposed by CA; if empty, it should not be processed
		nodegroup := mapNodegroup[node.NodegroupId]
		nodegroupMinInfo := &types.NodegroupMinInfo{Id: nodegroup.Id, MaxSize: nodegroup.MaxSize, MinSize: nodegroup.MinSize}
		return nodegroupMinInfo, nil
	}
	return &types.NodegroupMinInfo{}, nil

}

// getCurrentSize get the current size of a nodegroup
func GetCurrentSize(id string) (*types.NodegroupCurrentSize, error) {
	nodegroup, exist := mapNodegroup[id]
	if !exist {
		return nil, nil
	}
	nodegroupCurrentSize := types.NodegroupCurrentSize{CurrentSize: nodegroup.CurrentSize}
	return &nodegroupCurrentSize, nil
}

// getNodegroupNodes get all the nodes of a nodegroup
func GetNodegroupNodes(id string) (*[]types.NodeMinInfo, error) {

	// Clear the list before adding the new nodes
	nodeMinInfoList = nodeMinInfoList[:0]
	nodegroup, exist := mapNodegroup[id]
	if !exist {
		return nil, nil //TODO change with error
	}
	if len(nodegroup.Nodes) == 0 {
		return nil, nil
	} else {
		for _, nodeId := range nodegroup.Nodes {
			nodeX := mapNode[nodeId]
			nodeMinInfoList = append(nodeMinInfoList, types.NodeMinInfo{Id: nodeX.Id, InstanceStatus: nodeX.InstanceStatus})
		}
		return &nodeMinInfoList, nil
	}
}

// createNodegroup create a new nodegroup
func CreateNodegroup(newNodegroup types.Nodegroup) (success bool, err error) {

	if _, exists := mapNodegroup[newNodegroup.Id]; exists {
		return false, fmt.Errorf("nodegroup already exists")
	}
	mapNodegroup[newNodegroup.Id] = newNodegroup
	// TODO check if we need a lock on nodegroupiscached for the get all nodegroups
	// Update the list of nodegroups
	newNodegroupMinInfo := types.NodegroupMinInfo{Id: newNodegroup.Id, MaxSize: newNodegroup.MaxSize, MinSize: newNodegroup.MinSize}
	nodegroupListMinInfo = append(nodegroupListMinInfo, newNodegroupMinInfo)
	//nodegroupList = append(nodegroupList, newNodegroup)
	return true, nil
}

// deleteNodegroup delete the target nodegroup
func DeleteNodegroup(nodegroupId string) (success bool, err error) {

	if _, exists := mapNodegroup[nodegroupId]; !exists {
		return false, fmt.Errorf("nodegroup not found")
	}
	delete(mapNodegroup, nodegroupId)
	// Remove the nodegroup from the list
	for i, nodegroup := range nodegroupListMinInfo {
		if nodegroup.Id == nodegroupId {
			nodegroupListMinInfo = append(nodegroupListMinInfo[:i], nodegroupListMinInfo[i+1:]...)
			break
		}
	}
	// Update the cache flag if the list is empty
	if len(nodegroupListMinInfo) == 0 {
		nodegroupIdsCached = false
	}
	return true, nil
}

// scaleUpNodegroup scale up the nodegroup of a certain amount
func ScaleUpNodegroup(nodegroupId string, count int) (success bool, err error) {

	clusterList, err := util.GetClusterList()
	if err != nil {
		panic(err)
	}

	//--------------------------------------------------
	// Choose the first cluster that match the template (two template, one for GPU and one for standard)
	// TODO implement a better choice algorithm
	//--------------------------------------------------
	for i := 0; i < count; i++ {
		clusterchosen := types.Cluster{}
		for _, c := range clusterList {
			if c.Labels["hasGPU"] == "true" && nodegroupId == "GPU" {
				if _, exists := mapRemoteClusters[c.Name]; !exists {
					clusterchosen = c
					mapRemoteClusters[c.Name] = c.Name
					break
				}
			}
			if c.Labels["hasGPU"] == "false" && nodegroupId == "STANDARD" {
				if _, exists := mapRemoteClusters[c.Name]; !exists {
					clusterchosen = c
					mapRemoteClusters[c.Name] = c.Name
					break
				}
			}
		}

		log.Printf("Cluster chosen: %s", clusterchosen.Name)

		ip, kubeconfigPathRemote, tmp, err := util.DecodeKubeconfig(clusterchosen.Kubeconfig)
		if err != nil {
			return false, fmt.Errorf("kubeconfig decode error: %w", err)
		}

		// Wait until the node exists

		_, clientset := util.CreateKubernetesClient("ordinary")
		errPeering := util.PeeringWithLiqoctl(clusterchosen, nodegroupId, kubeconfigPathRemote, tmp, ip)
		if errPeering != nil {
			return false, fmt.Errorf("peering error: %w", errPeering)
		} else {
			log.Printf("Peering created successfully with cluster %s", clusterchosen.Name)
		}

		for {
			_, err := clientset.CoreV1().Nodes().Get(context.Background(), clusterchosen.Name, metav1.GetOptions{})
			if err != nil {
				log.Printf("Node %s not created yet, waiting 1s...", clusterchosen.Name)
				time.Sleep(1 * time.Second)
				continue
			}
			break
		}
		log.Printf("Virtual node %s created", clusterchosen.Name)

		// Wait until the node is Ready (max 30 seconds)
		timeout := time.After(30 * time.Second)
		ticker := time.Tick(1 * time.Second)

	waitLoop:
		for {
			select {
			case <-timeout:
				log.Fatalf("Timeout waiting for node %s to become Ready", clusterchosen.Name)

			case <-ticker:
				node, err := clientset.CoreV1().Nodes().Get(context.Background(), clusterchosen.Name, metav1.GetOptions{})
				if err != nil {
					continue
				}

				for _, cond := range node.Status.Conditions {
					if cond.Type == corev1.NodeReady && cond.Status == corev1.ConditionTrue {
						break waitLoop
					}
				}
			}
		}
		log.Printf("Virtual node %s is ready", clusterchosen.Name)

		// Patch node with providerID
		providerId := "liqo://" + clusterchosen.Name
		patch := fmt.Sprintf(`{"spec":{"providerID":"%s"}}`, providerId)

		_, err = clientset.CoreV1().Nodes().Patch(
			context.Background(),
			clusterchosen.Name,
			k8stypes.MergePatchType,
			[]byte(patch),
			metav1.PatchOptions{},
		)
		if err != nil {
			log.Fatalf("Patch failed: %v", err)
		}

		//--------------------------------------------------
		// Update internal structures
		//--------------------------------------------------
		mapNode[clusterchosen.Name] = types.Node{Id: clusterchosen.Name, NodegroupId: nodegroupId, InstanceStatus: types.InstanceStatus{InstanceState: 1, InstanceErrorInfo: ""}}
		nodegroup := mapNodegroup[nodegroupId]
		nodegroup.CurrentSize++
		log.Printf("Nodegroup size is %d after adding node %s", nodegroup.CurrentSize, clusterchosen.Name)
		nodegroup.Nodes = append(nodegroup.Nodes, clusterchosen.Name)
		mapNodegroup[nodegroupId] = nodegroup
	}
	return true, nil

}

// scaleDownNodegroup scale down the nodegroup killing a certain node
func ScaleDownNodegroup(nodegroupId string, nodeId string) (success bool, err error) {

	clusterList, err := util.GetClusterList()
	if err != nil {
		panic(err)
	}

	//divide in nodegroup
	cleanNodeId := strings.TrimPrefix(nodeId, "liqo://")
	clusterchosen := types.Cluster{}
	for _, c := range clusterList {
		if c.Name == cleanNodeId {
			clusterchosen = c
			break
		}
	}

	log.Printf("preso cluster %s", clusterchosen.Name)

	_, kubeconfigPath, _, error := util.DecodeKubeconfig(clusterchosen.Kubeconfig)
	if error != nil {
		return false, fmt.Errorf("kubeconfig decode error: %w", error)
	}

	err = util.UnPeeringWithLiqoctl(kubeconfigPath)
	if err != nil {
		return false, fmt.Errorf("unpeering error: %w", err)
	}
	nodegroup := mapNodegroup[nodegroupId]
	for i, node := range nodegroup.Nodes {
		if node == cleanNodeId { // Remove the node from the list
			nodegroup.Nodes = append(nodegroup.Nodes[:i], nodegroup.Nodes[i+1:]...)
			break
		}
	}
	nodegroup.CurrentSize--
	log.Printf("Nodegroup size is %d after removing node %s", nodegroup.CurrentSize, cleanNodeId)
	mapNodegroup[nodegroupId] = nodegroup
	delete(mapNode, cleanNodeId)
	delete(mapRemoteClusters, cleanNodeId)
	return true, nil
}

// // getTemplateNodegroup get the template of a nodegroup
func GetTemplateNodegroup(id string) (*types.NodegroupTemplate, error) {

	// No existing Nodegrouptemplate
	log.Printf("GetTemplateNodegroup called with id: %s", id)
	template, exist := mapNodegroupTemplate[id]
	if !exist {
		log.Printf("Nodegroup template not found")
		return nil, nil //TODO change with error
	}
	return &template, nil
}

// getPriceNodegroup get the price of a nodegroup
func GetPriceNodegroup(id string) (*float64, error) {

	// No existing Nodegrouptemplate
	re := regexp.MustCompile(`for-([^-]+)-`)
	templateName := re.FindStringSubmatch(id)
	log.Printf("Extracted template name: %s", templateName[1])
	template, exist := mapNodegroupTemplate[templateName[1]]
	if !exist {
		log.Printf("Nodegroup template not found")
		return nil, nil //TODO change with error
	}
	return &template.Cost, nil
}

// getPriceNodegroup get the price of a nodegroup
func GetPricePod() (*float64, error) {

	// Assuming same price
	var podprice float64 = 1.00
	return &podprice, nil
}
