package types

import (
	v1 "k8s.io/api/core/v1"
)

var KeyPem string = "./key.pem"

var CertPem string = "./cert.pem"

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

type NodegroupCurrentSize struct {
	CurrentSize int32 `json:"currentSize"`
}

type NodeMinInfo struct {
	Id             string         `json:"id"`
	InstanceStatus InstanceStatus `json:"--"`
}

// HERE END CUSTOM OBJECTS TO ADHERE GRPC TYPES

// Nodegroup list with all fields
//var nodegroupList []Nodegroup = make([]Nodegroup, 0, 6)

// Node list
//var nodeList []Node = make([]Node, 0, 20)

type Cluster struct {
	Name       string          `json:"name"`
	Kubeconfig string          `json:"kubeconfig"`
	Resources  v1.ResourceList `json:"resources"` // Number of resources available in the cluster
}
