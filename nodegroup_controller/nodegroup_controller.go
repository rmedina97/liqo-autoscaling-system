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

// TODO transfer case code inside functions
func handleConnection(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/nodegroup": //Asks about all the nodegroup
		// No existing Nodegroups
		if len(mapNodegroup) == 0 {
			w.WriteHeader(http.StatusNoContent)
		} else {
			// Check if it is cached
			if !nodegroupIdsCached {
				for _, nodegroup := range mapNodegroup {
					nodegroupList = append(nodegroupList, nodegroup)
				}
				nodegroupIdsCached = true
			}
			// Send response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(nodegroupList); err != nil {
				http.Error(w, fmt.Sprintf("Errore encoding JSON: %v", err), http.StatusInternalServerError)
			}

		}
	case "/nodegroup/ownership":
		//ricerca tramite mappa chiave nodo valore nodegroup
		queryParams := r.URL.Query()
		node, exist := mapNode[queryParams.Get("id")]
		if !exist { //TODO error node id doesn't exist
			w.WriteHeader(http.StatusNoContent)
			break
		}
		if node.NodegroupId != "" {
			nodegroup := mapNodegroup[node.NodegroupId]
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(nodegroup); err != nil {
				http.Error(w, fmt.Sprintf("Errore encoding JSON: %v", err), http.StatusInternalServerError)
			}

		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	case "/nodegroup/current-size":
		queryParams := r.URL.Query()
		nodegroup, exist := mapNodegroup[queryParams.Get("id")]
		if !exist { //TODO error node id doesn't exist
			w.WriteHeader(http.StatusNoContent)
			break
		} else {
			nodegroupCurrentSize := NodegroupCurrentSize{CurrentSize: nodegroup.CurrentSize}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(nodegroupCurrentSize); err != nil {
				http.Error(w, fmt.Sprintf("Errore encoding JSON: %v", err), http.StatusInternalServerError)
			}
		}
	case "/gpu/label":
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(gpuLabel); err != nil {
			http.Error(w, fmt.Sprintf("Errore encoding JSON: %v", err), http.StatusInternalServerError)
		}
	case "/gpu/types":
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(gpuLabelsList); err != nil {
			http.Error(w, fmt.Sprintf("Errore encoding JSON: %v", err), http.StatusInternalServerError)
		}
	case "/nodegroup/nodes":
		nodegroup, exist := mapNodegroup[r.URL.Query().Get("id")]
		log.Printf("nodegroup %s", nodegroup.Id)
		if !exist {
			w.WriteHeader(http.StatusNoContent)
		}
		for _, node := range nodegroup.Nodes {
			log.Printf("node %s che stampa", node)
			nodeList = append(nodeList, mapNode[node])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(nodeList); err != nil {
			http.Error(w, fmt.Sprintf("Errore encoding JSON: %v", err), http.StatusInternalServerError)
		}
	case "/nodegroup/create":
		var newNodegroup Nodegroup
		if err := json.NewDecoder(r.Body).Decode(&newNodegroup); err != nil {
			http.Error(w, fmt.Sprintf("Errore decoding JSON: %v", err), http.StatusBadRequest)
			return
		}
		if _, exists := mapNodegroup[newNodegroup.Id]; exists {
			http.Error(w, "Nodegroup already exists", http.StatusConflict)
			return
		}
		mapNodegroup[newNodegroup.Id] = newNodegroup
		// TODO check if we need a lock on nodegroupiscached for the get all nodegroups
		nodegroupList = append(nodegroupList, newNodegroup)
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(newNodegroup); err != nil {
			http.Error(w, fmt.Sprintf("Errore encoding JSON: %v", err), http.StatusInternalServerError)
		}
	case "/nodegroup/destroy":
		queryParams := r.URL.Query()
		nodegroupId := queryParams.Get("id")
		if _, exists := mapNodegroup[nodegroupId]; !exists {
			http.Error(w, "Nodegroup not found", http.StatusNotFound)
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
		w.WriteHeader(http.StatusOK)
	case "/nodegroup/scaleup":
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
		w.WriteHeader(http.StatusOK)
	case "/nodegroup/scaledown":
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
		w.WriteHeader(http.StatusOK)
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
func startPeriodicFunction() {
	panic("unimplemented")
}

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
