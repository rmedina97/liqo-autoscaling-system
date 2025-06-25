package util

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	types "nodegroupController/types"
	"os/exec"
)

// key is id nodegroup, value is nodegroup
// var mapNodegroup = make(map[string]types.Nodegroup)
var mapNodegroup = map[string]types.Nodegroup{
	"SINGLE": {
		Id:          "SINGLE",
		MaxSize:     3,
		MinSize:     1,
		CurrentSize: 1,
		Nodes:       []string{"instance-zf6d5"},
	},
}

// flag for cache
var nodegroupIdsCached = false

// Node list
var nodeMinInfoList []types.NodeMinInfo = make([]types.NodeMinInfo, 0, 20)

// Nodegroup list
var nodegroupListMinInfo []types.NodegroupMinInfo = make([]types.NodegroupMinInfo, 0, 5)

// key is id node, value is node
// var mapNode = make(map[string]types.Node)
var mapNode = map[string]types.Node{
	"instance-zf6d5": {
		Id:             "instance-zf6d5",
		NodegroupId:    "SINGLE",
		InstanceStatus: types.InstanceStatus{InstanceState: 1, InstanceErrorInfo: ""},
	},
}

// List of function inside the handle connection -------------------------------------------------------

// getAllNodegroups get all the nodegroups
func GetAllNodegroups(w http.ResponseWriter) {

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
}

// nodegroupForNode get the nodegroup for a specific node
func GetNodegroupForNode(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	//id := queryParams.Get("id")
	//idWithoutPrefix := strings.TrimPrefix(id, "k3s://")
	node, exist := mapNode[queryParams.Get("id")]
	if !exist {
		WriteGetResponse(w, http.StatusNotFound, nil, "Node not found")
		return
	}
	if node.NodegroupId != "" {
		nodegroup := mapNodegroup[node.NodegroupId]
		nodegroupMinInfo := types.NodegroupMinInfo{Id: nodegroup.Id, MaxSize: nodegroup.MaxSize, MinSize: nodegroup.MinSize}
		WriteGetResponse(w, http.StatusOK, nodegroupMinInfo, "")
	} else {
		WriteGetResponse(w, http.StatusNotFound, nil, "Nodegroup not found")
	}

}

// getCurrentSize get the current size of a nodegroup
func GetCurrentSize(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	nodegroup, exist := mapNodegroup[queryParams.Get("id")]
	if !exist {
		WriteGetResponse(w, http.StatusNotFound, nil, "Nodegroup not found")
		return
	} else {
		nodegroupCurrentSize := types.NodegroupCurrentSize{CurrentSize: nodegroup.CurrentSize}
		WriteGetResponse(w, http.StatusOK, nodegroupCurrentSize, "")
	}
}

// getNodegroupNodes get all the nodes of a nodegroup
func GetNodegroupNodes(w http.ResponseWriter, r *http.Request) {

	// Clear the list before adding the new nodes
	nodeMinInfoList = nodeMinInfoList[:0]
	nodegroup, exist := mapNodegroup[r.URL.Query().Get("id")]
	if !exist {
		WriteGetResponse(w, http.StatusNotFound, nil, "Nodegroup not found")
	}
	if len(nodegroup.Nodes) == 0 {
		WriteGetResponse(w, http.StatusNotFound, nil, "Nodes not found")
	} else {
		for _, nodeId := range nodegroup.Nodes {
			nodeX := mapNode[nodeId]
			nodeMinInfoList = append(nodeMinInfoList, types.NodeMinInfo{Id: nodeX.Id, InstanceStatus: nodeX.InstanceStatus})
		}
		log.Printf("nodeMinInfoList: %v", nodeMinInfoList)
		WriteGetResponse(w, http.StatusOK, nodeMinInfoList, "")
	}
}

// WriteGetResponse write the response of a get request
func WriteGetResponse(w http.ResponseWriter, statusCode int, data any, errMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if errMsg != "" {
		http.Error(w, errMsg, statusCode)
		return
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("Errore encoding JSON: %v", err), http.StatusInternalServerError)
	}
}

// createNodegroup create a new nodegroup
func CreateNodegroup(w http.ResponseWriter, r *http.Request) {
	var newNodegroup types.Nodegroup
	if err := json.NewDecoder(r.Body).Decode(&newNodegroup); err != nil {
		http.Error(w, fmt.Sprintf("Errore decoding JSON: %v", err), http.StatusBadRequest)
		return
	}
	if _, exists := mapNodegroup[newNodegroup.Id]; exists {
		WriteGetResponse(w, http.StatusConflict, nil, "Nodegroup already exists")
		return
	}
	mapNodegroup[newNodegroup.Id] = newNodegroup
	// TODO check if we need a lock on nodegroupiscached for the get all nodegroups
	// Update the list of nodegroups
	newNodegroupMinInfo := types.NodegroupMinInfo{Id: newNodegroup.Id, MaxSize: newNodegroup.MaxSize, MinSize: newNodegroup.MinSize}
	nodegroupListMinInfo = append(nodegroupListMinInfo, newNodegroupMinInfo)
	//nodegroupList = append(nodegroupList, newNodegroup)
	WriteGetResponse(w, http.StatusCreated, nil, "")
}

// deleteNodegroup delete the target nodegroup
func DeleteNodegroup(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	nodegroupId := queryParams.Get("id")
	if _, exists := mapNodegroup[nodegroupId]; !exists {
		WriteGetResponse(w, http.StatusNotFound, nil, "Nodegroup not found")
		return
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
	WriteGetResponse(w, http.StatusNoContent, nil, "")
}

// scaleUpNodegroup scale up the nodegroup of a certain amount
func ScaleUpNodegroup(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	//numberToAdd := queryParams.Get("deltaInt")
	nodegroupId := queryParams.Get("id")
	cmd := exec.Command(
		"ssh",
		"-J", "bastion@ssh.crownlabs.polito.it",
		"crownlabs@10.97.97.14",
		"liqoctl", "peer", "out-of-band", "remoto",
		"--auth-url", "https://172.16.203.62:32473",
		"--cluster-id", "1b5f548d-630b-4a95-90e2-9157b5a560ba",
		"--auth-token", "dea56520895f222a8575f58270f08df46a8249d7180da6b5b747dd9cd2d62261e704a3c1c9b21abfdb0094eb02e2e5401776634e64f6aca480549c423fbca936",
	)
	output, err := cmd.CombinedOutput()
	log.Printf("Output: %s %s", output, err)
	mapNode["liqo-remoto"] = types.Node{Id: "liqo-remoto", NodegroupId: "SINGLE", InstanceStatus: types.InstanceStatus{InstanceState: 1, InstanceErrorInfo: ""}}
	nodegroup := mapNodegroup[nodegroupId]
	nodegroup.CurrentSize++
	nodegroup.Nodes = append(nodegroup.Nodes, "liqo-remoto")
	mapNodegroup[nodegroupId] = nodegroup
	WriteGetResponse(w, http.StatusOK, nil, "")
}

// scaleDownNodegroup scale down the nodegroup killing a certain node
func ScaleDownNodegroup(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	nodegroupId := queryParams.Get("nodegroupid")
	nodeId := queryParams.Get("id")
	cmd := exec.Command(
		"ssh",
		"-J", "bastion@ssh.crownlabs.polito.it",
		"crownlabs@10.97.97.14",
		"liqoctl", "unpeer", "remoto", "--skip-confirm",
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
	WriteGetResponse(w, http.StatusOK, nil, "")
}
