package util

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	types "nodegroupController/types"
	"os"
	"os/exec"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// key is id nodegroup, value is nodegroup
// TODO change from hardcoded data to a dynamic one
var mapNodegroup = map[string]types.Nodegroup{
	"STANDARD": {
		Id:          "STANDARD",
		MaxSize:     3,
		MinSize:     1,
		CurrentSize: 1,
		Nodes:       []string{"rmedina"},
	},
	"GPU": {
		Id:          "GPU",
		MaxSize:     3,
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
		Id:             "rmedina",
		NodegroupId:    "STANDARD",
		InstanceStatus: types.InstanceStatus{InstanceState: 1, InstanceErrorInfo: ""},
	},
}

// Template map
var mapNodegroupTemplate = map[string]types.NodegroupTemplate{
	"GPU": {
		NodegroupId: "GPU",
		Resources: types.ResourceRange{
			Min: v1.ResourceList{
				v1.ResourceCPU:    *resource.NewQuantity(4000, resource.DecimalSI),
				v1.ResourceMemory: *resource.NewQuantity(4096, resource.DecimalSI),
				v1.ResourcePods:   *resource.NewQuantity(110, resource.DecimalSI),
				"nvidia.com/gpu":  *resource.NewQuantity(1, resource.DecimalSI),
			},
			Max: v1.ResourceList{
				v1.ResourceCPU:    *resource.NewQuantity(7000, resource.DecimalSI),
				v1.ResourceMemory: *resource.NewQuantity(8192, resource.DecimalSI),
				v1.ResourcePods:   *resource.NewQuantity(110, resource.DecimalSI),
				"nvidia.com/gpu":  *resource.NewQuantity(4, resource.DecimalSI),
			},
		},
		Labels: map[string]string{
			"country":  "france",
			"provider": "liqo",
			"city":     "paris",
		},
	},
	"STANDARD": {
		NodegroupId: "STANDARD",
		Resources: types.ResourceRange{
			Min: v1.ResourceList{
				v1.ResourceCPU:    *resource.NewQuantity(1500, resource.DecimalSI),
				v1.ResourceMemory: *resource.NewQuantity(1024, resource.DecimalSI),
				v1.ResourcePods:   *resource.NewQuantity(110, resource.DecimalSI),
				"nvidia.com/gpu":  *resource.NewQuantity(0, resource.DecimalSI),
			},
			Max: v1.ResourceList{
				v1.ResourceCPU:    *resource.NewQuantity(3000, resource.DecimalSI),
				v1.ResourceMemory: *resource.NewQuantity(3072, resource.DecimalSI),
				v1.ResourcePods:   *resource.NewQuantity(110, resource.DecimalSI),
				"nvidia.com/gpu":  *resource.NewQuantity(0, resource.DecimalSI),
			},
		},
		Labels: map[string]string{
			"country":  "italy",
			"provider": "liqo",
			"city":     "turin",
		},
	},
}

// List of function inside the handle connection -------------------------------------------------------

// getAllNodegroups get all the nodegroups
/*func GetAllNodegroups(w http.ResponseWriter) {

	// No existing Nodegroups
	if len(mapNodegroup) == 0 {
		WriteGetResponse(w, http.StatusNotFound, nil, "Nodegroups not found")
		w.WriteHeader(http.StatusNoContent)
	} else {
		// Check if it is cached
		// TODO add a log, cache flag is necessary when will exist the cr and there will be a function to search if cr already exist
		if !nodegroupIdsCached {
			for _, nodegroup := range mapNodegroup {

				nodegroupMinInfo := types.NodegroupMinInfo{Id: nodegroup.Id, MaxSize: nodegroup.MaxSize, MinSize: nodegroup.MinSize}
				nodegroupListMinInfo = append(nodegroupListMinInfo, nodegroupMinInfo)
			}
			nodegroupIdsCached = true
		}
		WriteGetResponse(w, http.StatusOK, nodegroupListMinInfo, "")
	}
}*/

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
// func GetNodegroupForNode(w http.ResponseWriter, r *http.Request) {
// 	queryParams := r.URL.Query()
// 	//id := queryParams.Get("id")
// 	//idWithoutPrefix := strings.TrimPrefix(id, "k3s://")
// 	node, exist := mapNode[queryParams.Get("id")]
// 	if !exist {
// 		WriteGetResponse(w, http.StatusNotFound, nil, "Node not found")
// 		return
// 	}
// 	if node.NodegroupId != "" {
// 		nodegroup := mapNodegroup[node.NodegroupId]
// 		nodegroupMinInfo := types.NodegroupMinInfo{Id: nodegroup.Id, MaxSize: nodegroup.MaxSize, MinSize: nodegroup.MinSize}
// 		WriteGetResponse(w, http.StatusOK, nodegroupMinInfo, "")
// 	} else {
// 		WriteGetResponse(w, http.StatusNotFound, nil, "Nodegroup not found")
// 	}
// }

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
// func GetCurrentSize(w http.ResponseWriter, r *http.Request) {
// 	queryParams := r.URL.Query()
// 	nodegroup, exist := mapNodegroup[queryParams.Get("id")]
// 	if !exist {
// 		WriteGetResponse(w, http.StatusNotFound, nil, "Nodegroup not found")
// 		return
// 	} else {
// 		nodegroupCurrentSize := types.NodegroupCurrentSize{CurrentSize: nodegroup.CurrentSize}
// 		WriteGetResponse(w, http.StatusOK, nodegroupCurrentSize, "")
// 	}
// }

// getCurrentSize get the current size of a nodegroup
func GetCurrentSize(id string) (*types.NodegroupCurrentSize, error) {
	nodegroup, exist := mapNodegroup[id]
	if !exist {
		return nil, nil
	}
	nodegroupCurrentSize := types.NodegroupCurrentSize{CurrentSize: nodegroup.CurrentSize}
	return &nodegroupCurrentSize, nil
}

// // getNodegroupNodes get all the nodes of a nodegroup
// func GetNodegroupNodes(w http.ResponseWriter, r *http.Request) {

// 	// Clear the list before adding the new nodes
// 	nodeMinInfoList = nodeMinInfoList[:0]
// 	nodegroup, exist := mapNodegroup[r.URL.Query().Get("id")]
// 	if !exist {
// 		WriteGetResponse(w, http.StatusNotFound, nil, "Nodegroup not found")
// 	}
// 	if len(nodegroup.Nodes) == 0 {
// 		WriteGetResponse(w, http.StatusNotFound, nil, "Nodes not found")
// 	} else {
// 		for _, nodeId := range nodegroup.Nodes {
// 			nodeX := mapNode[nodeId]
// 			nodeMinInfoList = append(nodeMinInfoList, types.NodeMinInfo{Id: nodeX.Id, InstanceStatus: nodeX.InstanceStatus})
// 		}
// 		log.Printf("nodeMinInfoList: %v", nodeMinInfoList)
// 		WriteGetResponse(w, http.StatusOK, nodeMinInfoList, "")
// 	}
// }

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
		log.Printf("nodeMinInfoList: %v", nodeMinInfoList)
		return &nodeMinInfoList, nil
	}
}

// // WriteGetResponse write the response of a get request
// func WriteGetResponse(w http.ResponseWriter, statusCode int, data any, errMsg string) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(statusCode)
// 	if errMsg != "" {
// 		http.Error(w, errMsg, statusCode)
// 		return
// 	}
// 	if err := json.NewEncoder(w).Encode(data); err != nil {
// 		http.Error(w, fmt.Sprintf("Errore encoding JSON: %v", err), http.StatusInternalServerError)
// 	}
// }

// // createNodegroup create a new nodegroup
// func CreateNodegroup(w http.ResponseWriter, r *http.Request) {
// 	var newNodegroup types.Nodegroup
// 	if err := json.NewDecoder(r.Body).Decode(&newNodegroup); err != nil {
// 		http.Error(w, fmt.Sprintf("Errore decoding JSON: %v", err), http.StatusBadRequest)
// 		return
// 	}
// 	if _, exists := mapNodegroup[newNodegroup.Id]; exists {
// 		WriteGetResponse(w, http.StatusConflict, nil, "Nodegroup already exists")
// 		return
// 	}
// 	mapNodegroup[newNodegroup.Id] = newNodegroup
// 	// TODO check if we need a lock on nodegroupiscached for the get all nodegroups
// 	// Update the list of nodegroups
// 	newNodegroupMinInfo := types.NodegroupMinInfo{Id: newNodegroup.Id, MaxSize: newNodegroup.MaxSize, MinSize: newNodegroup.MinSize}
// 	nodegroupListMinInfo = append(nodegroupListMinInfo, newNodegroupMinInfo)
// 	//nodegroupList = append(nodegroupList, newNodegroup)
// 	WriteGetResponse(w, http.StatusCreated, nil, "")
// }

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

// // deleteNodegroup delete the target nodegroup
// func DeleteNodegroup(w http.ResponseWriter, r *http.Request) {
// 	queryParams := r.URL.Query()
// 	nodegroupId := queryParams.Get("id")
// 	if _, exists := mapNodegroup[nodegroupId]; !exists {
// 		WriteGetResponse(w, http.StatusNotFound, nil, "Nodegroup not found")
// 		return
// 	}
// 	delete(mapNodegroup, nodegroupId)
// 	// Remove the nodegroup from the list
// 	for i, nodegroup := range nodegroupListMinInfo {
// 		if nodegroup.Id == nodegroupId {
// 			nodegroupListMinInfo = append(nodegroupListMinInfo[:i], nodegroupListMinInfo[i+1:]...)
// 			break
// 		}
// 	}
// 	// Update the cache flag if the list is empty
// 	if len(nodegroupListMinInfo) == 0 {
// 		nodegroupIdsCached = false
// 	}
// 	WriteGetResponse(w, http.StatusNoContent, nil, "")
// }

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

// // scaleUpNodegroup scale up the nodegroup of a certain amount
// func ScaleUpNodegroup(w http.ResponseWriter, r *http.Request) {

// 	// TODO implement https server that send resources needed

// 	queryParams := r.URL.Query()
// 	nodegroupId := queryParams.Get("id")
// 	//numberToAdd := queryParams.Get("deltaInt")
// 	log.Printf("ScaleUpNodegroup called with query params: %v", queryParams)

// 	client, err := newClient()
// 	if err != nil {
// 		log.Printf("failed to create a client: %v", err)
// 	}

// 	// Send a GET request to the discovery server
// 	reply, err := client.Get("https://localhost:9010/list") // TODO create a parameter
// 	if err != nil {
// 		log.Printf("failed to execute get query: %v", err)
// 	}
// 	defer reply.Body.Close()

// 	// Check the response status code
// 	if reply.StatusCode == http.StatusNotFound {
// 		log.Printf("remote cluster not found")
// 	} else if reply.StatusCode != http.StatusOK {
// 		log.Printf("server responded with status %d", reply.StatusCode)
// 	}

// 	// Decode the JSON response
// 	var clusterList []types.Cluster
// 	if err := json.NewDecoder(reply.Body).Decode(&clusterList); err != nil {
// 		log.Printf("error decoding JSON: %v", err)
// 	}
// 	log.Printf("Cluster chosen: %s", clusterList[0].Name)
// 	cmd := exec.Command(
// 		"liqoctl", "peer", "--remote-kubeconfig", "/home/rmedina/provider.kubeconfig", "--skip-confirm",
// 	)
// 	output, err := cmd.CombinedOutput()
// 	if err != nil {
// 		log.Printf("Error during SSH:%v", err)
// 	}
// 	log.Printf("Output: %s ", output)
// 	mapNode["remote"] = types.Node{Id: "remote", NodegroupId: "SINGLE", InstanceStatus: types.InstanceStatus{InstanceState: 1, InstanceErrorInfo: ""}}
// 	nodegroup := mapNodegroup[nodegroupId]
// 	nodegroup.CurrentSize++
// 	nodegroup.Nodes = append(nodegroup.Nodes, "remote")
// 	mapNodegroup[nodegroupId] = nodegroup
// 	WriteGetResponse(w, http.StatusOK, nil, "")

// }

// scaleUpNodegroup scale up the nodegroup of a certain amount
func ScaleUpNodegroup(nodegroupId string) (success bool, err error) {

	//numberToAdd := queryParams.Get("deltaInt")
	log.Printf("ScaleUpNodegroup called with query params: %s", nodegroupId)

	client, err := newClient()
	if err != nil {
		log.Printf("failed to create a client: %v", err)
	}

	// Send a GET request to the discovery server
	reply, err := client.Get("https://localhost:9010/list") // TODO create a parameter
	if err != nil {
		log.Printf("failed to execute get query: %v", err)
	}
	defer reply.Body.Close()

	// Check the response status code
	if reply.StatusCode == http.StatusNotFound {
		log.Printf("remote cluster not found")
	} else if reply.StatusCode != http.StatusOK {
		log.Printf("server responded with status %d", reply.StatusCode)
	}

	// Decode the JSON response
	var clusterList []types.Cluster
	if err := json.NewDecoder(reply.Body).Decode(&clusterList); err != nil {
		log.Printf("error decoding JSON: %v", err)
	}

	//divide in nodegroup
	clusterchosen := types.Cluster{}
	for _, c := range clusterList {
		if c.Labels["hasGPU"] == "true" && nodegroupId == "GPU" {
			clusterchosen = c
			break
		}
		if c.Labels["hasGPU"] == "false" && nodegroupId == "STANDARD" {
			clusterchosen = c
			break
		}
	}

	log.Printf("Cluster chosen: %s", clusterchosen.Name)
	// -----------------------------------------
	// Decodifica
	kubeconfigBytes, err := base64.StdEncoding.DecodeString(clusterchosen.Kubeconfig)
	if err != nil {
		panic(err)
	}
	log.Printf("preso new cluster %s", clusterchosen.Name)

	// Crea file temporaneo
	tmpFile, err := os.CreateTemp("", "kubeconfig-*.yaml")
	if err != nil {
		panic(err)
	}

	defer os.Remove(tmpFile.Name())

	// Scrive il contenuto
	if _, err := tmpFile.Write(kubeconfigBytes); err != nil {
		panic(err)
	}
	tmpFile.Close()

	// Ottieni il path
	kubeconfigPath := tmpFile.Name()
	fmt.Println("Kubeconfig salvato in:", kubeconfigPath)
	// -----------------------------------------
	cmd := exec.Command(
		"liqoctl", "peer", "--remote-kubeconfig", kubeconfigPath, "--skip-confirm",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error during SSH:%v", err)
	}
	log.Printf("Output: %s ", output)
	mapNode[clusterchosen.Name] = types.Node{Id: clusterchosen.Name, NodegroupId: nodegroupId, InstanceStatus: types.InstanceStatus{InstanceState: 1, InstanceErrorInfo: ""}}
	nodegroup := mapNodegroup[nodegroupId]
	nodegroup.CurrentSize++
	nodegroup.Nodes = append(nodegroup.Nodes, clusterchosen.Name)
	mapNodegroup[nodegroupId] = nodegroup
	return true, nil

}

// // scaleDownNodegroup scale down the nodegroup killing a certain node
// func ScaleDownNodegroup(w http.ResponseWriter, r *http.Request) {
// 	queryParams := r.URL.Query()
// 	nodegroupId := queryParams.Get("nodegroupid")
// 	nodeId := queryParams.Get("id")
// 	//log.Printf("ScaleDownNodegroup called on first: %s", nodeId)
// 	nodeId = "remote"
// 	cmd := exec.Command(
// 		"liqoctl", "unpeer", "--remote-kubeconfig", "/home/rmedina/provider.kubeconfig", "--skip-confirm",
// 	)
// 	output, err := cmd.CombinedOutput()
// 	log.Printf("Fine SSH, %s %s", output, err)
// 	if err != nil {
// 		log.Printf("Error during SSH: %v", err)
// 	}
// 	nodegroup := mapNodegroup[nodegroupId]
// 	for i, node := range nodegroup.Nodes {
// 		if node == nodeId { // Remove the node from the list
// 			nodegroup.Nodes = append(nodegroup.Nodes[:i], nodegroup.Nodes[i+1:]...)
// 			break
// 		}
// 	}
// 	nodegroup.CurrentSize--
// 	mapNodegroup[nodegroupId] = nodegroup
// 	delete(mapNode, nodeId)
// 	WriteGetResponse(w, http.StatusOK, nil, "")
// }

// scaleDownNodegroup scale down the nodegroup killing a certain node
func ScaleDownNodegroup(nodegroupId string, nodeId string) (success bool, err error) {

	// TODO Sent request to discovery server for only the kubeconfig of that node
	client, err := newClient()
	if err != nil {
		log.Printf("failed to create a client: %v", err)
	}

	// Send a GET request to the discovery server
	reply, err := client.Get("https://localhost:9010/list") // TODO create a parameter
	if err != nil {
		log.Printf("failed to execute get query: %v", err)
	}
	defer reply.Body.Close()

	// Check the response status code
	if reply.StatusCode == http.StatusNotFound {
		log.Printf("remote cluster not found")
	} else if reply.StatusCode != http.StatusOK {
		log.Printf("server responded with status %d", reply.StatusCode)
	}

	// Decode the JSON response
	var clusterList []types.Cluster
	if err := json.NewDecoder(reply.Body).Decode(&clusterList); err != nil {
		log.Printf("error decoding JSON: %v", err)
	}

	//divide in nodegroup
	clusterchosen := types.Cluster{}
	for _, c := range clusterList {
		if c.Name == nodeId {
			clusterchosen = c
			break
		}
	}

	log.Printf("Cluster chosen: %s", clusterchosen.Name)
	// -----------------------------------------
	// Decodifica
	kubeconfigBytes, err := base64.StdEncoding.DecodeString(clusterchosen.Kubeconfig)
	if err != nil {
		panic(err)
	}
	log.Printf("preso cluster %s", clusterchosen.Name)

	// Crea file temporaneo
	tmpFile, err := os.CreateTemp("", "kubeconfig-*.yaml")
	if err != nil {
		panic(err)
	}

	defer os.Remove(tmpFile.Name())

	// Scrive il contenuto
	if _, err := tmpFile.Write(kubeconfigBytes); err != nil {
		panic(err)
	}
	tmpFile.Close()

	// Ottieni il path
	kubeconfigPath := tmpFile.Name()
	fmt.Println("Kubeconfig salvato in:", kubeconfigPath)
	// -----------------------------------------

	//log.Printf("ScaleDownNodegroup called on first: %s", nodeId)
	cmd := exec.Command(
		"liqoctl", "unpeer", "--remote-kubeconfig", kubeconfigPath, "--skip-confirm",
	)
	output, err := cmd.CombinedOutput()
	log.Printf("Fine SSH, %s %s", output, err)
	if err != nil {
		log.Printf("Error during SSH: %v", err)
	}
	nodegroup := mapNodegroup[nodegroupId]
	for i, node := range nodegroup.Nodes {
		if node == nodeId { // Remove the node from the list
			nodegroup.Nodes = append(nodegroup.Nodes[:i], nodegroup.Nodes[i+1:]...)
			break
		}
	}
	nodegroup.CurrentSize--
	mapNodegroup[nodegroupId] = nodegroup
	delete(mapNode, nodeId)
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

// Create a new client
// TODO Search if someone still uses 509 cert without san, if yes use VerifyPeerCertificate to custom accept them
func newClient() (*http.Client, error) {
	certPool := x509.NewCertPool()
	certData, err := os.ReadFile("cert.pem")

	if err != nil {
		return nil, fmt.Errorf("error reading certificate: %v", err)
	}

	if !certPool.AppendCertsFromPEM(certData) {
		return nil, fmt.Errorf("failed to append certificate")
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:   &tls.Config{RootCAs: certPool},
			ForceAttemptHTTP2: true,
		},
	}
	return client, nil

}
