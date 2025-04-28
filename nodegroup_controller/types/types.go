package types

var KeyPem string = "C:/Users/ricca/Desktop/server_grpc_1.30/gRPC_server/nodegroup_controller/key.pem"

var CertPem string = "C:/Users/ricca/Desktop/server_grpc_1.30/gRPC_server/nodegroup_controller/cert.pem"

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
