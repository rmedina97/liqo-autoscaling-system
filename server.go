package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/externalgrpc/protos"
)

// protos->package with the files definition (_grpc.pb.go and pb.go) born from the proto file
type cloudProviderServer struct {
	protos.UnimplementedCloudProviderServer
}

// NodeGroups returns all node groups configured for this cloud provider.
func (s *cloudProviderServer) NodeGroups(ctx context.Context, req *protos.NodeGroupsRequest) (*protos.NodeGroupsResponse, error) {
	//here TODO the real computations
	return &protos.NodeGroupsResponse{
		NodeGroups: []*protos.NodeGroup{
			{
				Id:      "Nord",
				MinSize: 1,
				MaxSize: 1,
			},
			{
				Id:      "Sud",
				MinSize: 1,
				MaxSize: 1,
			},
		},
	}, nil
}

// NodeGroupForNode returns the node group for the given node.
// The node group id is an empty string if the node should not
// be processed by cluster autoscaler.
func (c *cloudProviderServer) NodeGroupForNode(ctx context.Context, req *protos.NodeGroupForNodeRequest) (*protos.NodeGroupForNodeResponse, error) {
	//here TODO the real computations
	return &protos.NodeGroupForNodeResponse{
		NodeGroup: &protos.NodeGroup{
			Id:      "Sud",
			MinSize: 1,
			MaxSize: 1,
		},
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
	return &protos.GPULabelResponse{
		Label: "gpu=yes",
	}, nil
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (c *cloudProviderServer) GetAvailableGPUTypes(ctx context.Context, req *protos.GetAvailableGPUTypesRequest) (*protos.GetAvailableGPUTypesResponse, error) {
	//here TODO the real computations
	return nil, nil
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
	return &protos.NodeGroupTargetSizeResponse{
		TargetSize: 1,
	}, nil
}

// NodeGroupIncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use NodeGroupDeleteNodes. This function should wait until
// node group size is updated.
func (c *cloudProviderServer) NodeGroupIncreaseSize(ctx context.Context, req *protos.NodeGroupIncreaseSizeRequest) (*protos.NodeGroupIncreaseSizeResponse, error) {
	//here TODO the real computations
	return &protos.NodeGroupIncreaseSizeResponse{}, nil
}

// NodeGroupDeleteNodes deletes nodes from this node group (and also decreasing the size
// of the node group with that). Error is returned either on failure or if the given node
// doesn't belong to this node group. This function should wait until node group size is updated.
func (c *cloudProviderServer) NodeGroupDeleteNodes(ctx context.Context, req *protos.NodeGroupDeleteNodesRequest) (*protos.NodeGroupDeleteNodesResponse, error) {
	//here TODO the real computations
	return &protos.NodeGroupDeleteNodesResponse{}, nil
}

// NodeGroupDecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the request
// for new nodes that have not been yet fulfilled. Delta should be negative. It is assumed
// that cloud provider will not delete the existing nodes if the size when there is an option
// to just decrease the target.
func (c *cloudProviderServer) NodeGroupDecreaseTargetSize(ctx context.Context, req *protos.NodeGroupDecreaseTargetSizeRequest) (*protos.NodeGroupDecreaseTargetSizeResponse, error) {
	//here TODO the real computations
	return &protos.NodeGroupDecreaseTargetSizeResponse{}, nil
}

// NodeGroupNodes returns a list of all nodes that belong to this node group.
func (c *cloudProviderServer) NodeGroupNodes(ctx context.Context, req *protos.NodeGroupNodesRequest) (*protos.NodeGroupNodesResponse, error) {
	//here TODO the real computations
	return &protos.NodeGroupNodesResponse{}, nil
}

// NodeGroupTemplateNodeInfo returns a structure of an empty (as if just started) node,
// with all of the labels, capacity and allocatable information. This will be used in
// scale-up simulations to predict what would a new node look like if a node group was expanded.
// Implementation optional: if unimplemented return error code 12 (for `Unimplemented`)
func (c *cloudProviderServer) NodeGroupTemplateNodeInfo(ctx context.Context, req *protos.NodeGroupTemplateNodeInfoRequest) (*protos.NodeGroupTemplateNodeInfoResponse, error) {
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
