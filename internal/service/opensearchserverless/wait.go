package opensearchserverless

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
)

func waitVPCEndpointCreated(ctx context.Context, conn *opensearchserverless.Client, id string, timeout time.Duration) (*types.VpcEndpointDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   enum.Slice(types.VpcEndpointStatusPending),
		Target:                    enum.Slice(types.VpcEndpointStatusActive),
		Refresh:                   statusVPCEndpoint(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.VpcEndpointDetail); ok {
		return out, err
	}

	return nil, err
}

func waitVPCEndpointUpdated(ctx context.Context, conn *opensearchserverless.Client, id string, timeout time.Duration) (*types.VpcEndpointDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   enum.Slice(types.VpcEndpointStatusPending),
		Target:                    enum.Slice(types.VpcEndpointStatusActive),
		Refresh:                   statusVPCEndpoint(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.VpcEndpointDetail); ok {
		return out, err
	}

	return nil, err
}

func waitVPCEndpointDeleted(ctx context.Context, conn *opensearchserverless.Client, id string, timeout time.Duration) (*types.VpcEndpointDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: enum.Slice(types.VpcEndpointStatusDeleting, types.VpcEndpointStatusActive),
		Target:  []string{},
		Refresh: statusVPCEndpoint(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.VpcEndpointDetail); ok {
		return out, err
	}

	return nil, err
}
