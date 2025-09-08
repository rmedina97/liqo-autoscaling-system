package main

import (
	//"encoding/json"
	//"fmt"
	"log"
	"net/http"

	//"os/exec"
	handler "nodegroupController/handler"
	types "nodegroupController/types"
)

func main() {

	//go startPeriodicFunction()
	/*mapNodegroup["primonodegroup"] = Nodegroup{Id: "primonodegroup", MaxSize: 3, MinSize: 1, CurrentSize: 2, Nodes: []string{"uno", "tre"}}
	mapNodegroup["secondonodegroup"] = Nodegroup{Id: "secondonodegroup", MaxSize: 3, MinSize: 1, CurrentSize: 2, Nodes: []string{"quattro", "due"}}
	mapNode["uno"] = Node{Id: "uno", NodegroupId: "primonodegroup", InstanceStatus: InstanceStatus{InstanceState: 1, InstanceErrorInfo: ""}}
	mapNode["due"] = Node{Id: "due", NodegroupId: "secondonodegroup", InstanceStatus: InstanceStatus{InstanceState: 1, InstanceErrorInfo: ""}}
	mapNode["tre"] = Node{Id: "tre", NodegroupId: "primonodegroup", InstanceStatus: InstanceStatus{InstanceState: 1, InstanceErrorInfo: ""}}
	mapNode["quattro"] = Node{Id: "quattro", NodegroupId: "secondonodegroup", InstanceStatus: InstanceStatus{InstanceState: 1, InstanceErrorInfo: ""}}
	*/
	/*mapNodegroup["SINGLE"] = Nodegroup{Id: "SINGLE", MaxSize: 3, MinSize: 1, CurrentSize: 1, Nodes: []string{"instance-zf6d5"}}
	mapNode["instance-zf6d5"] = Node{Id: "instance-zf6d5", NodegroupId: "SINGLE", InstanceStatus: InstanceStatus{InstanceState: 1, InstanceErrorInfo: ""}}
	gpuLabelsList = append(gpuLabelsList, "first type")
	gpuLabelsList = append(gpuLabelsList, "second type")
	*/
	// TODO maybe search local cluster and adds all the nodes inside the same nodegroup, with the label to avoid scaledown
	mux := http.NewServeMux()
	//TODO use different handler for different routes
	mux.HandleFunc("/", handler.HandleConnection)
	err := http.ListenAndServeTLS(":9009", types.CertPem, types.KeyPem, mux)
	if err != nil {
		log.Fatalf("failed to start server, %v ", err)
	}
}
