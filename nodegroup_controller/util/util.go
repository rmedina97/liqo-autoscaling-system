package util

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	types "nodegroupController/types"
	"os"
	"os/exec"
)

// key is id nodegroup, value is nodegroup
// TODO change from hardcoded data to a dynamic one
var mapNodegroup = map[string]types.Nodegroup{
	"SINGLE": {
		Id:          "SINGLE",
		MaxSize:     3,
		MinSize:     1,
		CurrentSize: 1,
		Nodes:       []string{"rmedina"},
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

	// TODO implement https server that send resources needed

	queryParams := r.URL.Query()
	nodegroupId := queryParams.Get("id")
	//numberToAdd := queryParams.Get("deltaInt")
	log.Printf("ScaleUpNodegroup called with query params: %v", queryParams)

	client, err := newClient()
	if err != nil {
		log.Printf("failed to create a client: %v", err)
	}

	// Send a GET request to the nodegroup controller
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
	log.Printf("NodeGroupNodes: %d lunghezza mappa", len(clusterList))

	// Convert the response to the protos format
	// QUI invece fai il peering, serve il percorso poichè non lo puoi buttare dentro (o forse si)
	/*kubeconfig := clusterList[0].Kubeconfig
	path := "/home/rmedina/kubeconfig1"

	// Scrive il file con permessi 0600 (solo lettura/scrittura per l'utente)
	erro := os.WriteFile(path, []byte(kubeconfig), 0777)
	if erro != nil {
		fmt.Println("Errore durante il salvataggio del kubeconfig:", err)
		return
	}*/
	log.Printf("CLUSTER SCELTO: %s", clusterList[0].Name)
	cmd := exec.Command(
		"liqoctl", "peer", "--remote-kubeconfig", "/home/rmedina/provider.kubeconfig", "--skip-confirm",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error during SSH:%v", err)
	}
	log.Printf("Output: %s ", output)
	mapNode["remote"] = types.Node{Id: "remote", NodegroupId: "SINGLE", InstanceStatus: types.InstanceStatus{InstanceState: 1, InstanceErrorInfo: ""}}
	nodegroup := mapNodegroup[nodegroupId]
	nodegroup.CurrentSize++
	nodegroup.Nodes = append(nodegroup.Nodes, "remote")
	mapNodegroup[nodegroupId] = nodegroup
	WriteGetResponse(w, http.StatusOK, nil, "")

	/*queryParams := r.URL.Query()
	//numberToAdd := queryParams.Get("deltaInt")
	log.Printf("ScaleUpNodegroup called with query params: %v", queryParams)
	nodegroupId := queryParams.Get("id")
	//"ssh","rmedina@192.168.11.114",
	cmd := exec.Command(
		"liqoctl", "peer", "--remote-kubeconfig", "/home/rmedina/kubeconfig", "--skip-confirm",
	)
	output, err:= cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error during SSH:%v", err)
	}
	log.Printf("Output: %s ", output)
	mapNode["remote"] = types.Node{Id: "remote", NodegroupId: "SINGLE", InstanceStatus: types.InstanceStatus{InstanceState: 1, InstanceErrorInfo: ""}}
	nodegroup := mapNodegroup[nodegroupId]
	nodegroup.CurrentSize++
	nodegroup.Nodes = append(nodegroup.Nodes, "remote")
	mapNodegroup[nodegroupId] = nodegroup
	WriteGetResponse(w, http.StatusOK, nil, "")*/
}

// scaleDownNodegroup scale down the nodegroup killing a certain node
func ScaleDownNodegroup(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	nodegroupId := queryParams.Get("nodegroupid")
	nodeId := queryParams.Get("id")
	log.Printf("ScaleDownNodegroup CALLED ON: %s", nodeId)
	cmd := exec.Command(
		"liqoctl", "unpeer", "--remote-kubeconfig", "/home/rmedina/provider.kubeconfig", "--skip-confirm",
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
