package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/externalgrpc/protos"

	"net/http"
)

// HERE START PERSONAL STRUCTS---------------// HERE START PERSONAL STRUCTS	---------------// HERE START PERSONAL STRUCTS
type Nodegroup struct {
	Id          string   `json:"id"`
	CurrentSize int32    `json:"currentSize"` //TODO struct only with the required field
	MaxSize     int32    `json:"maxSize"`
	MinSize     int32    `json:"minSize"`
	Nodes       []string `json:"nodes"` //TODO maybe put only ids of the nodes?
}

type Node struct {
	Id          string `json:"id"`
	NodegroupId string `json:"nodegroupId"`
}

// HERE END PERSONAL STRUCTS---------------// HERE END PERSONAL STRUCTS	---------------// HERE END PERSONAL STRUCTS

// protos->package with the files definition (_grpc.pb.go and pb.go) born from the proto file
type cloudProviderServer struct {
	protos.UnimplementedCloudProviderServer
}

// HERE START HARDCODED COMPONENTS---------------// HERE START HARDCODED COMPONENTS	---------------// HERE START HARDCODED COMPONENTS
// Hardcoded invented type
type GPUTypes struct {
	Name  string
	Specs string
}

// hardcoded flag for increase size
var test int = 1

// HERE END HARDCODED COMPONENTS---------------// HERE END HARDCODED COMPONENTS	---------------// HERE END HARDCODED COMPONENTS

// NodeGroups returns all node groups configured for this cloud provider.
func (s *cloudProviderServer) NodeGroups(ctx context.Context, req *protos.NodeGroupsRequest) (*protos.NodeGroupsResponse, error) {

	// Send a GET request to the nodegroup controller
	reply, err := http.Get("http://localhost:9009/nodegroup") // TODO create a parameter
	if err != nil {
		log.Printf("Error during HTTP request: %v", err)
		//return nil,err // TODO probably there is a specific error
	}
	defer reply.Body.Close()

	// Check the response status code
	if reply.StatusCode == http.StatusNoContent {
		return &protos.NodeGroupsResponse{}, nil //TODO probably there is a specific error
	} else if reply.StatusCode != http.StatusOK {
		log.Printf("errore: server ha risposto con status %d", reply.StatusCode)
		return nil, nil
	}

	// Decode the JSON response
	var nodeGroups []Nodegroup
	if err := json.NewDecoder(reply.Body).Decode(&nodeGroups); err != nil {
		return nil, fmt.Errorf("errore nel decoding JSON: %v", err)
	}

	// Convert the response to the protos format
	protoNodeGroups := make([]*protos.NodeGroup, len(nodeGroups))
	for i, nodegroup := range nodeGroups {
		log.Printf("iterazione n %d, nodegroup %v", i, nodegroup)
		protoNodeGroups[i] = &protos.NodeGroup{
			Id:      nodegroup.Id,
			MinSize: nodegroup.MinSize,
			MaxSize: nodegroup.MaxSize,
		}
	}
	// Return the response
	return &protos.NodeGroupsResponse{
		NodeGroups: protoNodeGroups,
	}, nil

}

// NodeGroupForNode returns the node group for the given node.
// The node group id is an empty string if the node should not
// be processed by cluster autoscaler.
func (c *cloudProviderServer) NodeGroupForNode(ctx context.Context, req *protos.NodeGroupForNodeRequest) (*protos.NodeGroupForNodeResponse, error) {

	//here TODO the real computations

	// Take the parameter
	nodeId := req.Node.ProviderID
	url := fmt.Sprintf("http://localhost:9009/nodegroup/ownership?id=%s", nodeId)
	// Send a GET request to the nodegroup controller
	reply, err := http.Get(url) // TODO create a better parameter, maybe using something more complex like DefaultClient
	if err != nil {
		log.Printf("Error during HTTP request: %v", err)
		//return nil,err // TODO probably there is a specific error
	}
	defer reply.Body.Close()

	// Check the response status code
	if reply.StatusCode == http.StatusNoContent {
		return &protos.NodeGroupForNodeResponse{}, nil //TODO probably there is a specific error
	} else if reply.StatusCode != http.StatusOK {
		log.Printf("errore: server ha risposto con status %d", reply.StatusCode)
		return nil, nil
	}

	// Decode the JSON response
	var nodeGroup Nodegroup
	if err := json.NewDecoder(reply.Body).Decode(&nodeGroup); err != nil {
		return nil, fmt.Errorf("errore nel decoding JSON: %v", err)
	}

	// Convert the response to the protos format
	protoNodeGroup := &protos.NodeGroup{
		Id:      nodeGroup.Id,
		MinSize: nodeGroup.MinSize,
		MaxSize: nodeGroup.MaxSize,
	}

	// Return the response
	return &protos.NodeGroupForNodeResponse{
		NodeGroup: protoNodeGroup,
	}, nil
}

// PricingNodePrice returns a theoretical minimum price of running a node for
// a given period of time on a perfectly matching machine.
// Implementation optional: if unimplemented return error code 12 (for `Unimplemented`)
func (c *cloudProviderServer) PricingNodePrice(ctx context.Context, req *protos.PricingNodePriceRequest) (*protos.PricingNodePriceResponse, error) {
	//here TODO the real computations
	return nil, status.Error(codes.Unimplemented, "function PricingNodePrice is Unimplemented")
}

// PricingPodPrice returns a theoretical minimum price of running a pod for a given
// period of time on a perfectly matching machine.
// Implementation optional: if unimplemented return error code 12 (for `Unimplemented`)
func (c *cloudProviderServer) PricingPodPrice(ctx context.Context, req *protos.PricingPodPriceRequest) (*protos.PricingPodPriceResponse, error) {
	//here TODO the real computations
	return nil, status.Error(codes.Unimplemented, "function PricingPodPrice is Unimplemented")
}

// GPULabel returns the label added to nodes with GPU resource.
func (c *cloudProviderServer) GPULabel(ctx context.Context, req *protos.GPULabelRequest) (*protos.GPULabelResponse, error) {
	//here TODO the real computations

	// Send a GET request to the nodegroup controller
	reply, err := http.Get("http://localhost:9009/gpu/label") // TODO create a parameter
	if err != nil {
		log.Printf("Error during HTTP request: %v", err)
		//return nil,err // TODO probably there is a specific error
	}
	defer reply.Body.Close()

	// Check the response status code
	if reply.StatusCode == http.StatusNoContent {
		return &protos.GPULabelResponse{}, nil //TODO probably there is a specific error
	} else if reply.StatusCode != http.StatusOK {
		log.Printf("errore: server ha risposto con status %d", reply.StatusCode)
		return nil, nil
	}

	// Decode the JSON response
	var gpuLabel string
	if err := json.NewDecoder(reply.Body).Decode(&gpuLabel); err != nil {
		return nil, fmt.Errorf("errore nel decoding JSON: %v", err)
	}

	return &protos.GPULabelResponse{
		Label: gpuLabel,
	}, nil
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (c *cloudProviderServer) GetAvailableGPUTypes(ctx context.Context, req *protos.GetAvailableGPUTypesRequest) (*protos.GetAvailableGPUTypesResponse, error) {
	//here TODO the real computations

	// Send a GET request to the nodegroup controller
	reply, err := http.Get("http://localhost:9009/gpu/types") // TODO create a parameter
	if err != nil {
		log.Printf("Error during HTTP request: %v", err)
		//return nil,err // TODO probably there is a specific error
	}
	defer reply.Body.Close()

	// Check the response status code
	if reply.StatusCode == http.StatusNoContent {
		return &protos.GetAvailableGPUTypesResponse{}, nil //TODO probably there is a specific error
	} else if reply.StatusCode != http.StatusOK {
		log.Printf("errore: server ha risposto con status %d", reply.StatusCode)
		return nil, nil
	}

	// Decode the JSON response
	var gpuLabels []string
	if err := json.NewDecoder(reply.Body).Decode(&gpuLabels); err != nil {
		return nil, fmt.Errorf("errore nel decoding JSON: %v", err)
	}

	// Convert the response to the protos format
	mapGpu := map[string]*anypb.Any{}
	for _, label := range gpuLabels {
		mapGpu[label] = &anypb.Any{}
	}

	return &protos.GetAvailableGPUTypesResponse{
		GpuTypes: mapGpu,
	}, nil
}

// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
func (c *cloudProviderServer) Cleanup(ctx context.Context, req *protos.CleanupRequest) (*protos.CleanupResponse, error) {
	//here TODO the real computations
	return &protos.CleanupResponse{}, nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
func (c *cloudProviderServer) Refresh(ctx context.Context, req *protos.RefreshRequest) (*protos.RefreshResponse, error) {
	//here TODO the real computations
	return &protos.RefreshResponse{}, nil
}

// NodeGroupTargetSize returns the current target size of the node group. It is possible
// that the number of nodes in Kubernetes is different at the moment but should be equal
// to the size of a node group once everything stabilizes (new nodes finish startup and
// registration or removed nodes are deleted completely).
func (c *cloudProviderServer) NodeGroupTargetSize(ctx context.Context, req *protos.NodeGroupTargetSizeRequest) (*protos.NodeGroupTargetSizeResponse, error) {
	//here TODO the real computations

	// Take the parameter
	nodeId := req.Id
	url := fmt.Sprintf("http://localhost:9009/nodegroup/current-size?id=%s", nodeId)

	// Send a GET request to the nodegroup controller
	reply, err := http.Get(url) // TODO create a parameter
	if err != nil {
		log.Printf("Error during HTTP request: %v", err)
		//return nil,err // TODO probably there is a specific error
	}
	defer reply.Body.Close()

	// Check the response status code
	if reply.StatusCode == http.StatusNoContent {
		return &protos.NodeGroupTargetSizeResponse{}, nil //TODO probably there is a specific error
	} else if reply.StatusCode != http.StatusOK {
		log.Printf("errore: server ha risposto con status %d", reply.StatusCode)
		return nil, nil
	}

	// Decode the JSON response
	var currentSize Nodegroup
	if err := json.NewDecoder(reply.Body).Decode(&currentSize); err != nil {
		return nil, fmt.Errorf("errore nel decoding JSON: %v", err)
	}

	return &protos.NodeGroupTargetSizeResponse{
		TargetSize: currentSize.CurrentSize,
	}, nil
}

// NodeGroupIncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use NodeGroupDeleteNodes. This function should wait until
// node group size is updated.
func (c *cloudProviderServer) NodeGroupIncreaseSize(ctx context.Context, req *protos.NodeGroupIncreaseSizeRequest) (*protos.NodeGroupIncreaseSizeResponse, error) {
	//here TODO the real computations

	// Take the parameter
	nodegroupId := req.Id
	log.Printf("l'id è %s e %s", nodegroupId, req.Id)
	url := fmt.Sprintf("http://localhost:9009/nodegroup/scaleup?id=%s", nodegroupId)

	// Send a GET request to the nodegroup controller
	log.Printf("l'url è %s", url)
	reply, err := http.Get(url) // TODO create a parameter
	if err != nil {
		log.Printf("Error during HTTP request: %v", err)
		//return nil,err // TODO probably there is a specific error
	}
	defer reply.Body.Close()

	// Check the response status code
	if reply.StatusCode == http.StatusNoContent {
		return &protos.NodeGroupIncreaseSizeResponse{}, nil //TODO probably there is a specific error
	} else if reply.StatusCode != http.StatusOK {
		log.Printf("status code è %d", reply.StatusCode)
		log.Printf("errore: server ha risposto con status %d", reply.StatusCode)
		return nil, nil
	}
	return &protos.NodeGroupIncreaseSizeResponse{}, nil
}

// NodeGroupDeleteNodes deletes nodes from this node group (and also decreasing the size
// of the node group with that). Error is returned either on failure or if the given node
// doesn't belong to this node group. This function should wait until node group size is updated.
func (c *cloudProviderServer) NodeGroupDeleteNodes(ctx context.Context, req *protos.NodeGroupDeleteNodesRequest) (*protos.NodeGroupDeleteNodesResponse, error) {
	//here TODO the real computations

	// Take the parameter
	nodeId := req.Nodes[0].ProviderID
	nodegroupId := req.Id
	log.Printf("l'id è %s e %s", nodegroupId, req.Id)
	url := fmt.Sprintf("http://localhost:9009/nodegroup/scaledown?id=%s&nodegroupid=%s", nodeId, nodegroupId)

	// Send a GET request to the nodegroup controller
	log.Printf("l'url è %s", url)
	reply, err := http.Get(url) // TODO create a parameter
	if err != nil {
		log.Printf("Error during HTTP request: %v", err)
		//return nil,err // TODO probably there is a specific error
	}
	defer reply.Body.Close()

	// Check the response status code
	if reply.StatusCode == http.StatusNoContent {
		return &protos.NodeGroupDeleteNodesResponse{}, nil //TODO probably there is a specific error
	} else if reply.StatusCode != http.StatusOK {
		log.Printf("status code è %d", reply.StatusCode)
		log.Printf("errore: server ha risposto con status %d", reply.StatusCode)
		return nil, nil
	}
	return &protos.NodeGroupDeleteNodesResponse{}, nil
}

/*if test == 2 {
	test = 1
	log.Printf("DECREASE SIZE of %s NODEGROUP, DELETING NODE %s", req.Id, req.Nodes[0].Name)
	log.Printf("list size is %d", len(req.Nodes))
	for i, node := range req.Nodes {
		log.Printf("Node %d: %s", i, node.Name)
	}
	log.Printf("Node fine loop")
	cmd := exec.Command(
		"ssh",
		"-J", "bastion@ssh.crownlabs.polito.it",
		"crownlabs@10.97.97.14",
		"liqoctl", "unpeer", "remoto", "--skip-confirm",
	)
	output, err := cmd.CombinedOutput()
	log.Printf("Fine SSH")
	if err != nil {
		log.Printf("Error during SSH: %v", err)
		//return nil,err
	}
	log.Printf(" %s", output)
}*/

// NodeGroupDecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the request
// for new nodes that have not been yet fulfilled. Delta should be negative. It is assumed
// that cloud provider will not delete the existing nodes if the size when there is an option
// to just decrease the target.
func (c *cloudProviderServer) NodeGroupDecreaseTargetSize(ctx context.Context, req *protos.NodeGroupDecreaseTargetSizeRequest) (*protos.NodeGroupDecreaseTargetSizeResponse, error) {
	//here TODO the real computations
	log.Printf("decrease SIZE target why? of %s", req.Id)
	return &protos.NodeGroupDecreaseTargetSizeResponse{}, nil
}

// NodeGroupNodes returns a list of all nodes that belong to this node group.
func (c *cloudProviderServer) NodeGroupNodes(ctx context.Context, req *protos.NodeGroupNodesRequest) (*protos.NodeGroupNodesResponse, error) {
	//here TODO the real computations

	// Take the parameter
	nodegroupId := req.Id
	url := fmt.Sprintf("http://localhost:9009/nodegroup/nodes?id=%s", nodegroupId)

	// Send a GET request to the nodegroup controller
	reply, err := http.Get(url) // TODO create a parameter
	if err != nil {
		log.Printf("Error during HTTP request: %v", err)
		//return nil,err // TODO probably there is a specific error
	}
	defer reply.Body.Close()

	// Check the response status code
	if reply.StatusCode == http.StatusNoContent {
		return &protos.NodeGroupNodesResponse{}, nil //TODO probably there is a specific error
	} else if reply.StatusCode != http.StatusOK {
		log.Printf("errore: server ha risposto con status %d", reply.StatusCode)
		return nil, nil
	}

	// Decode the JSON response
	var nodeList []Node
	if err := json.NewDecoder(reply.Body).Decode(&nodeList); err != nil {
		return nil, fmt.Errorf("errore nel decoding JSON: %v", err)
	}

	// Convert the response to the protos format
	protoNodes := make([]*protos.Instance, len(nodeList))
	for i, node := range nodeList {
		protoNodes[i] = &protos.Instance{
			Id: node.Id,
			Status: &protos.InstanceStatus{
				InstanceState: 1,
			},
		}
	}

	return &protos.NodeGroupNodesResponse{
		Instances: protoNodes,
	}, nil
}

/*log.Printf("INFO ABOUT NODES OF NODEGROUPS, per nodegropu %s", req.Id)
	if test == 1 {
		log.Printf("return sud-> 1 node")
		return &protos.NodeGroupNodesResponse{
			Instances: []*protos.Instance{
				{
					Id: "instance-zf6d5",
					Status: &protos.InstanceStatus{
						InstanceState: 1,
					},
				},
			},
		}, nil
	} else {
		log.Printf("return sud-> 2 nodes")
		return &protos.NodeGroupNodesResponse{
			Instances: []*protos.Instance{
				{
					Id: "instance-zf6d5",
					Status: &protos.InstanceStatus{
						InstanceState: 1,
					},
				},
				{
					Id: "liqo-remoto",
					Status: &protos.InstanceStatus{
						InstanceState: 1,
					},
				},
			},
		}, nil
	}
}*/

// NodeGroupTemplateNodeInfo returns a structure of an empty (as if just started) node,
// with all of the labels, capacity and allocatable information. This will be used in
// scale-up simulations to predict what would a new node look like if a node group was expanded.
// Implementation optional: if unimplemented return error code 12 (for `Unimplemented`)
func (c *cloudProviderServer) NodeGroupTemplateNodeInfo(ctx context.Context, req *protos.NodeGroupTemplateNodeInfoRequest) (*protos.NodeGroupTemplateNodeInfoResponse, error) {
	log.Printf("INFO templateeeeeeeeeeee")
	return nil, status.Error(codes.Unimplemented, "function NodeGroupTemplateNodeInfo is Unimplemented")
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup.
// Implementation optional: if unimplemented return error code 12 (for `Unimplemented`)
func (c *cloudProviderServer) NodeGroupGetOptions(ctx context.Context, req *protos.NodeGroupAutoscalingOptionsRequest) (*protos.NodeGroupAutoscalingOptionsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "function NodeGroupGetOptions is Unimplemented")
}

func main() {

	lis, err := net.Listen("tcp", ":9007")
	if err != nil {
		log.Fatalf("failed to start server, %v ", err)
	}

	server := grpc.NewServer()
	service := &cloudProviderServer{}

	protos.RegisterCloudProviderServer(server, service)
	//require.NoError(t, err)

	log.Printf("server partito")
	server.Serve(lis)

}
