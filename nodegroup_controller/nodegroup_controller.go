package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var nodegroupIdsCached = false
var kubeconfig string
var mapNode map[string]string
var mapNodegroup map[string][]string

type Node struct {
	Id string `json:"id"`
	//TODO other info
}

type Nodegroup struct {
	Id          string `json:"id"`
	CurrentSize int    `json:"currentSize"`
	MaxSize     int    `json:"maxSize"`
	MinSize     int    `json:"minSize"`
	Nodes       []Node
}

type NodegroupId struct {
	Id string `json:"id"`
}

// Nodegroup list with all fields
var nodegroupList []Nodegroup = make([]Nodegroup, 0, 5)

// Nodegroup list with only the ids
var nodegroupIdsList []NodegroupId = make([]NodegroupId, 0, 5)

func handleConnection(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/nodegroup": //Asks about all the nodegroup
		// No existing Nodegroups
		if len(nodegroupList) == 0 {
			//TODO RETURN MESSAGE
			w.WriteHeader(http.StatusNoContent)
		} else {
			//TODO RETURN MESSAGE
			// Check if it is cached
			log.Printf("esiste la cache? %t", nodegroupIdsCached)
			if !nodegroupIdsCached {
				for _, nodegroup := range nodegroupList {
					nodegroupIdsList = append(nodegroupIdsList, NodegroupId{Id: nodegroup.Id})
				}
				nodegroupIdsCached = true
				log.Printf("caso not cached")
			}
			// Send response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(nodegroupIdsList); err != nil {
				http.Error(w, fmt.Sprintf("Errore encoding JSON: %v", err), http.StatusInternalServerError)
			}

		}
	case "/nodegruop/ownership":
		//ricerca tramite mappa chiave nodo valore nodegroup
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

}*/

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

}

// Controller pattern
func startPeriodicFunction() {
	panic("unimplemented")
}

func main() {

	//go startPeriodicFunction()
	//listNodegroups()
	//listNodeOfNodegroup()
	nodegroupList = append(nodegroupList, Nodegroup{Id: "uno", MaxSize: 3, MinSize: 1})
	nodegroupList = append(nodegroupList, Nodegroup{Id: "due", MaxSize: 3, MinSize: 1})

	http.HandleFunc("/", handleConnection)

	err := http.ListenAndServe(":9009", nil)
	if err != nil {
		log.Fatalf("failed to start server, %v ", err)
	}
}
