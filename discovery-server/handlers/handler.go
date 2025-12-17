package handler

import (
	fun "discovery_server/functions"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func HandleConnection(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {

	case "/list":

		// Send all the possible remote clusters
		result, err := fun.ReturnList()
		log.Printf("Sent list to the Node Manager")
		WriteGetResponse(w, result, err)

	//case "/update":

	// Update the list of remote clusters
	//TODO: unwrap https request and pass only the raw data
	//err := fun.UpdateList()
	//WriteGetResponse(w, "", err)

	default:
		log.Printf("wrong request")
	}
}

// WriteGetResponse write the response of a get request
// TODO: change the status code, add the case for no data but no error
func WriteGetResponse(w http.ResponseWriter, data any, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if encodeErr := json.NewEncoder(w).Encode(data); encodeErr != nil {
		http.Error(w, fmt.Sprintf("Errore encoding JSON: %v", encodeErr), http.StatusInternalServerError)
	}
}
