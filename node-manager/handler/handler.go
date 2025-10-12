package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	types "nodegroupController/types"
	util "nodegroupController/util"
	"strconv"
)

// List of GPU labels
var gpuLabelsList = []string{"nvidia.com/gpu", "second type"}

// label for GPU node
var gpuLabel string = "nvidia.com/gpu"

func HandleConnection(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/nodegroup":

		// Get all the nodegroup
		result, err := util.GetAllNodegroups()
		WriteGetResponse(w, result, err)

	case "/nodegroup/ownership":

		// Get the nodegroup of a specific node
		queryParams := r.URL.Query()
		id := queryParams.Get("id")
		result, err := util.GetNodegroupForNode(id)
		WriteGetResponse(w, result, err)

	case "/nodegroup/current-size":

		// Get the current size of a specific nodegroup
		queryParams := r.URL.Query()
		id := queryParams.Get("id")
		result, err := util.GetCurrentSize(id)
		WriteGetResponse(w, result, err)

	case "/gpu/label":

		//TODO Design required to search
		WriteGetResponse(w, gpuLabel, nil)

	case "/gpu/types":

		//TODO Design required to search
		WriteGetResponse(w, gpuLabelsList, nil)

	case "/nodegroup/nodes":

		// Get all the nodes of a nodegroup
		queryParams := r.URL.Query()
		id := queryParams.Get("id")
		result, err := util.GetNodegroupNodes(id)
		WriteGetResponse(w, result, err)

	case "/nodegroup/create":

		// Create a new nodegroup
		var newNodegroup types.Nodegroup
		if err := json.NewDecoder(r.Body).Decode(&newNodegroup); err != nil {
			http.Error(w, fmt.Sprintf("Errore decoding JSON: %v", err), http.StatusBadRequest)
			WriteGetResponse(w, nil, err)
			return
		}
		result, err := util.CreateNodegroup(newNodegroup)
		WriteGetResponse(w, result, err)

	case "/nodegroup/destroy":

		// Delete the target nodegroup
		queryParams := r.URL.Query()
		id := queryParams.Get("id")
		result, err := util.DeleteNodegroup(id)
		WriteGetResponse(w, result, err)

	case "/nodegroup/scaleup":

		// Scale up the nodegroup of a certain amount
		queryParams := r.URL.Query()
		id := queryParams.Get("id")
		countstr := queryParams.Get("count")
		count, _ := strconv.Atoi(countstr)
		log.Printf("Count------------------------------------------------ %d for nodegroup %s", count, id)
		result, err := util.ScaleUpNodegroup(id, count)
		WriteGetResponse(w, result, err)

	case "/nodegroup/scaledown":

		// Scale down the nodegroup killing a certain node
		queryParams := r.URL.Query()
		nodegroupId := queryParams.Get("nodegroupid")
		nodeId := queryParams.Get("id")
		log.Printf("NodegroupId scale down------------------------------------------------ %s", nodegroupId)
		log.Printf("NodeId to be cancelled------------------------------------------------ %s", nodeId)
		util.ScaleDownNodegroup(nodegroupId, nodeId)

	case "/nodegroup/template":

		// Get all the nodegroup
		queryParams := r.URL.Query()
		nodegroupId := queryParams.Get("id")
		result, err := util.GetTemplateNodegroup(nodegroupId)
		WriteGetResponse(w, result, err)

	case "/nodegroup/price":

		// Get all the nodegroup
		queryParams := r.URL.Query()
		nodegroupId := queryParams.Get("id")
		result, err := util.GetPriceNodegroup(nodegroupId)
		WriteGetResponse(w, result, err)

	default:
		log.Printf("wrong request")
	}
}

// WriteGetResponse write the response of a get request
// TODO Design decision required to determine the response code
func WriteGetResponse(w http.ResponseWriter, data any, err error) {
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if data == nil {
		http.Error(w, "No data found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)

	if encodeErr := json.NewEncoder(w).Encode(data); encodeErr != nil {
		http.Error(w, fmt.Sprintf("Errore encoding JSON: %v", encodeErr), http.StatusInternalServerError)
	}
}
