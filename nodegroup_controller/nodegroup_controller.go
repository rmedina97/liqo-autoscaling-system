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

// List of GPU labels
var gpuLabelsList []string = make([]string, 0, 5)

// key is id node, value is node
var mapNode = make(map[string]Node)

// key is id nodegroup, value is nodegroup
var mapNodegroup = make(map[string]Nodegroup)

//TODO change all int in int32

// TODO change errorInfo in a struct
type InstanceStatus struct {
	InstanceState     int32 //from zero to three
	InstanceErrorInfo string
}

type Node struct {
	Id          string `json:"id"`
	NodegroupId string `json:"nodegroupId"`
	//TODO other info
	InstanceStatus InstanceStatus
}

type Nodegroup struct {
	Id          string   `json:"id"`
	CurrentSize int      `json:"currentSize"` //TODO struct only with the required field
	MaxSize     int      `json:"maxSize"`
	MinSize     int      `json:"minSize"`
	Nodes       []string `json:"nodes"` //TODO maybe put only ids of the nodes?
}

type NodegroupCurrentSize struct {
	CurrentSize int `json:"currentSize"`
}

// Nodegroup list with all fields
var nodegroupList []Nodegroup = make([]Nodegroup, 0, 5)

// Node list
var nodeList []Node = make([]Node, 0, 5) // TODO what starting capacity is the best?

// List of function inside the handle connection -------------------------------------------------------

// getAllNodegroups get all the nodegroups
func getAllNodegroups(w http.ResponseWriter, r *http.Request) {
	// No existing Nodegroups
	if len(mapNodegroup) == 0 {
		writeGetResponse(w, http.StatusNotFound, nil, "Nodegroups not found")
		w.WriteHeader(http.StatusNoContent)
	} else {
		// Check if it is cached
		if !nodegroupIdsCached {
			for _, nodegroup := range mapNodegroup {
				nodegroupList = append(nodegroupList, nodegroup)
			}
			nodegroupIdsCached = true
		}
		writeGetResponse(w, http.StatusOK, nodegroupList, "")
	}
}

// nodegroupForNode get the nodegroup for a specific node
func getNodegroupForNode(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	node, exist := mapNode[queryParams.Get("id")]
	if !exist { //TODO add different errors for node doesn't exist and he hasn't a nodegroup
		writeGetResponse(w, http.StatusNotFound, nil, "Node not found")
		return
	}
	if node.NodegroupId != "" {
		nodegroup := mapNodegroup[node.NodegroupId]
		writeGetResponse(w, http.StatusOK, nodegroup, "")
	} else {
		writeGetResponse(w, http.StatusNotFound, nil, "Nodegroup not found")
	}

}

// getCurrentSize get the current size of a nodegroup
func getCurrentSize(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	nodegroup, exist := mapNodegroup[queryParams.Get("id")]
	if !exist { //TODO error node id doesn't exist
		writeGetResponse(w, http.StatusNotFound, nil, "Nodegroup not found")
		return
	} else {
		nodegroupCurrentSize := NodegroupCurrentSize{CurrentSize: nodegroup.CurrentSize}
		writeGetResponse(w, http.StatusOK, nodegroupCurrentSize, "")
	}
}

// getNodegroupNodes get all the nodes of a nodegroup
func getNodegroupNodes(w http.ResponseWriter, r *http.Request) {
	nodegroup, exist := mapNodegroup[r.URL.Query().Get("id")]
	if !exist {
		writeGetResponse(w, http.StatusNotFound, nil, "Nodegroup not found")
	}
	for _, node := range nodegroup.Nodes {
		nodeList = append(nodeList, mapNode[node])
	}
	// TODO need to differentiate empty and full?
	writeGetResponse(w, http.StatusOK, nodeList, "")
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
	nodegroupList = append(nodegroupList, newNodegroup)
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
	for i, nodegroup := range nodegroupList {
		if nodegroup.Id == nodegroupId {
			nodegroupList = append(nodegroupList[:i], nodegroupList[i+1:]...)
			break
		}
	}
	// Update the cache flag if the list is empty
	if len(nodegroupList) == 0 {
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
	nodegroup.Nodes = append(nodegroup.Nodes, "cinque")
	mapNodegroup[nodegroupId] = nodegroup
	writeGetResponse(w, http.StatusOK, nil, "")
}

// scaleDownNodegroup scale down the nodegroup killing a certain node
func scaleDownNodegroup(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	nodegroupId := queryParams.Get("id")
	nodeId := queryParams.Get("nodeId")
	nodegroup := mapNodegroup[nodegroupId]
	for i, node := range nodegroup.Nodes {
		if node == nodeId { // Remove the node from the list
			nodegroup.Nodes = append(nodegroup.Nodes[:i], nodegroup.Nodes[i+1:]...)
			break
		}
	}
	mapNodegroup[nodegroupId] = nodegroup
	delete(mapNode, nodeId)
	writeGetResponse(w, http.StatusNoContent, nil, "")
}

// End of the list of function inside the handle connection------------------------------------------------

// TODO transfer case code inside functions
func handleConnection(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/nodegroup":

		// Get all the nodegroup
		getAllNodegroups(w, r)

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
	mapNodegroup["primo nodegroup"] = Nodegroup{Id: "primo nodegroup", MaxSize: 3, MinSize: 1, CurrentSize: 2, Nodes: []string{"uno", "tre"}}
	mapNodegroup["secondo nodegroup"] = Nodegroup{Id: "secondo nodegroup", MaxSize: 3, MinSize: 1, CurrentSize: 2, Nodes: []string{"quattro", "due"}}
	mapNode["uno"] = Node{Id: "uno", NodegroupId: "primo nodegroup"}
	mapNode["due"] = Node{Id: "due", NodegroupId: "secondo nodegroup"}
	mapNode["tre"] = Node{Id: "tre", NodegroupId: "primo nodegroup"}
	mapNode["quattro"] = Node{Id: "quattro", NodegroupId: "secondo nodegroup"}
	gpuLabelsList = append(gpuLabelsList, "first type")
	gpuLabelsList = append(gpuLabelsList, "second type")

	http.HandleFunc("/", handleConnection)

	err := http.ListenAndServe(":9009", nil)
	if err != nil {
		log.Fatalf("failed to start server, %v ", err)
	}
}
