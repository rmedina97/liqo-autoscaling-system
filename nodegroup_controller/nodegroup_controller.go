package main

import (
	"encoding/json"
	"fmt"
	"log"

	"net/http"
)

var nodegroupIdsCached = false
var kubeconfig string
var gpuLabel string = "GPU node"

var keyPem string = "C:/Users/ricca/Desktop/server_grpc_1.30/gRPC_server/nodegroup_controller/key.pem"

var certPem string = "C:/Users/ricca/Desktop/server_grpc_1.30/gRPC_server/nodegroup_controller/cert.pem"

// List of GPU labels
var gpuLabelsList []string = make([]string, 0, 5)

// key is id node, value is node
var mapNode = make(map[string]Node)

// key is id nodegroup, value is nodegroup
var mapNodegroup = make(map[string]Nodegroup)

// TODO change errorInfo in a struct
type InstanceStatus struct {
	InstanceState     int32 //from zero to three
	InstanceErrorInfo string
}

type Node struct {
	Id          string `json:"id"`
	NodegroupId string `json:"nodegroupId"`
	//TODO other info
	InstanceStatus InstanceStatus `json:"--"`
}

type Nodegroup struct {
	Id          string   `json:"id"`
	CurrentSize int32    `json:"currentSize"` //TODO struct only with the required field
	MaxSize     int32    `json:"maxSize"`
	MinSize     int32    `json:"minSize"`
	Nodes       []string `json:"nodes"` //TODO maybe put only ids of the nodes?
}

// HERE START CUSTOM OBJECTS TO ADHERE GRPC TYPES

type NodegroupMinInfo struct {
	Id      string `json:"id"`
	MaxSize int32  `json:"maxSize"`
	MinSize int32  `json:"minSize"`
}

var nodegroupListMinInfo []NodegroupMinInfo = make([]NodegroupMinInfo, 0, 5)

type NodegroupCurrentSize struct {
	CurrentSize int32 `json:"currentSize"`
}

type NodeMinInfo struct {
	Id             string         `json:"id"`
	InstanceStatus InstanceStatus `json:"--"`
}

// Node list
var nodeMinInfoList []NodeMinInfo = make([]NodeMinInfo, 0, 20)

// HERE END CUSTOM OBJECTS TO ADHERE GRPC TYPES

// Nodegroup list with all fields
//var nodegroupList []Nodegroup = make([]Nodegroup, 0, 6)

// Node list
//var nodeList []Node = make([]Node, 0, 20)

// List of function inside the handle connection -------------------------------------------------------

// getAllNodegroups get all the nodegroups
func getAllNodegroups(w http.ResponseWriter) {

	// No existing Nodegroups
	if len(mapNodegroup) == 0 {
		writeGetResponse(w, http.StatusNotFound, nil, "Nodegroups not found")
		w.WriteHeader(http.StatusNoContent)
	} else {
		// Check if it is cached
		// TODO add a log, cache flag is necessary when will exist the cr and there will be a function to search if cr already exist
		if !nodegroupIdsCached {
			for _, nodegroup := range mapNodegroup {

				nodegroupMinInfo := NodegroupMinInfo{Id: nodegroup.Id, MaxSize: nodegroup.MaxSize, MinSize: nodegroup.MinSize}
				nodegroupListMinInfo = append(nodegroupListMinInfo, nodegroupMinInfo)
			}
			nodegroupIdsCached = true
		}
		writeGetResponse(w, http.StatusOK, nodegroupListMinInfo, "")
	}
}

// nodegroupForNode get the nodegroup for a specific node
func getNodegroupForNode(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	node, exist := mapNode[queryParams.Get("id")]
	if !exist {
		writeGetResponse(w, http.StatusNotFound, nil, "Node not found")
		return
	}
	if node.NodegroupId != "" {
		nodegroup := mapNodegroup[node.NodegroupId]
		nodegroupMinInfo := NodegroupMinInfo{Id: nodegroup.Id, MaxSize: nodegroup.MaxSize, MinSize: nodegroup.MinSize}
		writeGetResponse(w, http.StatusOK, nodegroupMinInfo, "")
	} else {
		writeGetResponse(w, http.StatusNotFound, nil, "Nodegroup not found")
	}

}

// getCurrentSize get the current size of a nodegroup
func getCurrentSize(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	nodegroup, exist := mapNodegroup[queryParams.Get("id")]
	if !exist {
		writeGetResponse(w, http.StatusNotFound, nil, "Nodegroup not found")
		return
	} else {
		nodegroupCurrentSize := NodegroupCurrentSize{CurrentSize: nodegroup.CurrentSize}
		writeGetResponse(w, http.StatusOK, nodegroupCurrentSize, "")
	}
}

// getNodegroupNodes get all the nodes of a nodegroup
func getNodegroupNodes(w http.ResponseWriter, r *http.Request) {

	// Clear the list before adding the new nodes
	nodeMinInfoList = nodeMinInfoList[:0]
	nodegroup, exist := mapNodegroup[r.URL.Query().Get("id")]
	if !exist {
		writeGetResponse(w, http.StatusNotFound, nil, "Nodegroup not found")
	}
	if len(nodegroup.Nodes) == 0 {
		writeGetResponse(w, http.StatusNotFound, nil, "Nodes not found")
	} else {
		for _, nodeId := range nodegroup.Nodes {
			nodeX := mapNode[nodeId]
			nodeMinInfoList = append(nodeMinInfoList, NodeMinInfo{Id: nodeX.Id, InstanceStatus: nodeX.InstanceStatus})
		}
		writeGetResponse(w, http.StatusOK, nodeMinInfoList, "")
	}
}

// writeGetResponse write the response of a get request
func writeGetResponse(w http.ResponseWriter, statusCode int, data any, errMsg string) {
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
func createNodegroup(w http.ResponseWriter, r *http.Request) {
	var newNodegroup Nodegroup
	if err := json.NewDecoder(r.Body).Decode(&newNodegroup); err != nil {
		http.Error(w, fmt.Sprintf("Errore decoding JSON: %v", err), http.StatusBadRequest)
		return
	}
	if _, exists := mapNodegroup[newNodegroup.Id]; exists {
		writeGetResponse(w, http.StatusConflict, nil, "Nodegroup already exists")
		return
	}
	mapNodegroup[newNodegroup.Id] = newNodegroup
	// TODO check if we need a lock on nodegroupiscached for the get all nodegroups
	// Update the list of nodegroups
	newNodegroupMinInfo := NodegroupMinInfo{Id: newNodegroup.Id, MaxSize: newNodegroup.MaxSize, MinSize: newNodegroup.MinSize}
	nodegroupListMinInfo = append(nodegroupListMinInfo, newNodegroupMinInfo)
	//nodegroupList = append(nodegroupList, newNodegroup)
	writeGetResponse(w, http.StatusCreated, nil, "")
}

// deleteNodegroup delete the target nodegroup
func deleteNodegroup(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	nodegroupId := queryParams.Get("id")
	if _, exists := mapNodegroup[nodegroupId]; !exists {
		writeGetResponse(w, http.StatusNotFound, nil, "Nodegroup not found")
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
	writeGetResponse(w, http.StatusNoContent, nil, "")
}

// scaleUpNodegroup scale up the nodegroup of a certain amount
func scaleUpNodegroup(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	//numberToAdd := queryParams.Get("deltaInt")
	nodegroupId := queryParams.Get("id")
	/*cmd := exec.Command(
		"ssh",
		"-J", "bastion@ssh.crownlabs.polito.it",
		"crownlabs@10.97.97.14",
		"liqoctl", "unpeer", "remoto", "--skip-confirm",
	)
	output, err := cmd.CombinedOutput()
	log.Printf("Fine SSH")*/
	mapNode["cinque"] = Node{Id: "cinque", NodegroupId: "secondo nodegroup"}
	nodegroup := mapNodegroup[nodegroupId]
	nodegroup.CurrentSize++
	nodegroup.Nodes = append(nodegroup.Nodes, "cinque")
	mapNodegroup[nodegroupId] = nodegroup
	writeGetResponse(w, http.StatusOK, nil, "")
}

// scaleDownNodegroup scale down the nodegroup killing a certain node
func scaleDownNodegroup(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	nodegroupId := queryParams.Get("nodegroupid")
	nodeId := queryParams.Get("id")
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
	writeGetResponse(w, http.StatusOK, nil, "")
}

// End of the list of function inside the handle connection------------------------------------------------

// TODO transfer case code inside functions
func handleConnection(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/nodegroup":

		// Get all the nodegroup
		getAllNodegroups(w)

	case "/nodegroup/ownership":

		// Get the nodegroup of a specific node
		getNodegroupForNode(w, r)

	case "/nodegroup/current-size":

		// Get the current size of a specific nodegroup
		getCurrentSize(w, r)

	case "/gpu/label":

		//TODO CRUD functions for gpu label or a search
		writeGetResponse(w, http.StatusOK, gpuLabel, "")

	case "/gpu/types":

		//TODO CRUD functions for gpu label or a search
		writeGetResponse(w, http.StatusOK, gpuLabelsList, "")

	case "/nodegroup/nodes":

		// Get all the nodes of a nodegroup
		getNodegroupNodes(w, r)

	case "/nodegroup/create":

		// Create a new nodegroup
		createNodegroup(w, r)

	case "/nodegroup/destroy":

		// Delete the target nodegroup
		deleteNodegroup(w, r)

	case "/nodegroup/scaleup":

		// Scale up the nodegroup of a certain amount
		scaleUpNodegroup(w, r)

	case "/nodegroup/scaledown":

		// Scale down the nodegroup killing a certain node
		scaleDownNodegroup(w, r)

	default:
		log.Printf("wrong request")
	}
}

/*func listNodegroups() {

	//read the kubeconfig locally
	log.Printf("all node request start")
	cfg, err := clientcmd.BuildConfigFromFlags("https://localhost:6443", "C:/Users/ricca/Desktop/kubeconfig")
	if err != nil {
		log.Fatalf("Errore nel recupero della configurazione di Kubernetes: %v", err)
	}

	log.Printf("create clientset")
	//create the set of clients from the previous config, set for every Kubernetes group (core, authentication, extensions...)
	clientset, err := kubernetes.NewForConfig(cfg)

	ctx := context.Background()

	//search all nodes
	log.Printf("start search nodes")
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("errore nel recuperare i nodi: %v", err)
	}

	//Stamp all nodes found
	for _, node := range nodes.Items {
		log.Printf("%s", node.Name)
	}

}

func listNodeOfNodegroup() {
	//read the kubeconfig locally
	log.Printf("node of a particular nodegroup")
	cfg, err := clientcmd.BuildConfigFromFlags("https://localhost:6443", "C:/Users/ricca/Desktop/kubeconfig")
	if err != nil {
		log.Fatalf("Errore nel recupero della configurazione di Kubernetes: %v", err)
	}

	log.Printf("create clientset")
	//create the set of clients from the previous config, set for every Kubernetes group (core, authentication, extensions...)
	clientset, err := kubernetes.NewForConfig(cfg)

	ctx := context.Background()
	log.Printf("start search nodes")
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{LabelSelector: "liqo.io/type=virtual-node"})
	if err != nil {
		log.Printf("errore nel recuperare i nodi: %v", err)
	}

	//Stamp all nodes found
	for _, node := range nodes.Items {
		log.Printf("%s", node.Name)
	}

}*/

// Controller pattern
/*func startPeriodicFunction() {
	panic("unimplemented")
}*/

func main() {

	//go startPeriodicFunction()
	mapNodegroup["primonodegroup"] = Nodegroup{Id: "primonodegroup", MaxSize: 3, MinSize: 1, CurrentSize: 2, Nodes: []string{"uno", "tre"}}
	mapNodegroup["secondonodegroup"] = Nodegroup{Id: "secondonodegroup", MaxSize: 3, MinSize: 1, CurrentSize: 2, Nodes: []string{"quattro", "due"}}
	mapNode["uno"] = Node{Id: "uno", NodegroupId: "primonodegroup", InstanceStatus: InstanceStatus{InstanceState: 1, InstanceErrorInfo: ""}}
	mapNode["due"] = Node{Id: "due", NodegroupId: "secondonodegroup", InstanceStatus: InstanceStatus{InstanceState: 1, InstanceErrorInfo: ""}}
	mapNode["tre"] = Node{Id: "tre", NodegroupId: "primonodegroup", InstanceStatus: InstanceStatus{InstanceState: 1, InstanceErrorInfo: ""}}
	mapNode["quattro"] = Node{Id: "quattro", NodegroupId: "secondonodegroup", InstanceStatus: InstanceStatus{InstanceState: 1, InstanceErrorInfo: ""}}
	gpuLabelsList = append(gpuLabelsList, "first type")
	gpuLabelsList = append(gpuLabelsList, "second type")

	mux := http.NewServeMux()
	//TODO use different handler for different routes
	mux.HandleFunc("/", handleConnection)
	err := http.ListenAndServeTLS(":9009", certPem, keyPem, mux)
	if err != nil {
		log.Fatalf("failed to start server, %v ", err)
	}
}
