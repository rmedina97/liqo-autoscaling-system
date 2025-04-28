package handler

import (
	"log"
	"net/http"
	util "nodegroupController/util"
)

// List of GPU labels
var gpuLabelsList = []string{"first type", "second type"}

// label for GPU node
var gpuLabel string = "GPU node"

func HandleConnection(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/nodegroup":

		// Get all the nodegroup
		util.GetAllNodegroups(w)

	case "/nodegroup/ownership":

		// Get the nodegroup of a specific node
		util.GetNodegroupForNode(w, r)

	case "/nodegroup/current-size":

		// Get the current size of a specific nodegroup
		util.GetCurrentSize(w, r)

	case "/gpu/label":

		//TODO CRUD functions for gpu label or a search
		util.WriteGetResponse(w, http.StatusOK, gpuLabel, "")

	case "/gpu/types":

		//TODO CRUD functions for gpu label or a search
		util.WriteGetResponse(w, http.StatusOK, gpuLabelsList, "")

	case "/nodegroup/nodes":

		// Get all the nodes of a nodegroup
		util.GetNodegroupNodes(w, r)

	case "/nodegroup/create":

		// Create a new nodegroup
		util.CreateNodegroup(w, r)

	case "/nodegroup/destroy":

		// Delete the target nodegroup
		util.DeleteNodegroup(w, r)

	case "/nodegroup/scaleup":

		// Scale up the nodegroup of a certain amount
		util.ScaleUpNodegroup(w, r)

	case "/nodegroup/scaledown":

		// Scale down the nodegroup killing a certain node
		util.ScaleDownNodegroup(w, r)

	default:
		log.Printf("wrong request")
	}
}
