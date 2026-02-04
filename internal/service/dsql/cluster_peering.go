// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package dsql

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dsql"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dsql/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_dsql_cluster_peering", name="Cluster Peering")
func newClusterPeeringResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &clusterPeeringResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)

	return r, nil
}

type clusterPeeringResource struct {
	framework.ResourceWithModel[clusterPeeringResourceModel]
	framework.WithTimeouts
	framework.WithNoUpdate
	framework.WithNoOpDelete
}

func (r *clusterPeeringResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrIdentifier: schema.StringAttribute{
				Required: true,
			},
			"clusters": schema.SetAttribute{
				CustomType:  fwtypes.SetOfARNType,
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"witness_region": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					fwvalidators.AWSRegion(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}
}

func (r *clusterPeeringResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data clusterPeeringResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DSQLClient(ctx)

	// Check if the cluster exists and is in a valid state to create a peering connection.
	id := fwflex.StringValueFromFramework(ctx, data.Identifier)
	output, err := findClusterByID(ctx, conn, id)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Aurora DSQL Cluster (%s)", id), err.Error())

		return
	}

	if status := output.Status; status != awstypes.ClusterStatusPendingSetup {
		response.Diagnostics.AddError(fmt.Sprintf("Aurora DSQL Cluster (%s) is not in a valid state to create a peering", id), string(status))

		return
	}

	input := dsql.UpdateClusterInput{
		ClientToken:           aws.String(sdkid.UniqueId()),
		Identifier:            aws.String(id),
		MultiRegionProperties: new(awstypes.MultiRegionProperties),
	}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input.MultiRegionProperties)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err = conn.UpdateCluster(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Aurora DSQL Cluster (%s) peering", id), err.Error())

		return
	}

	output, err = waitClusterPeeringCreated(ctx, conn, data.Identifier.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err == nil && output.MultiRegionProperties == nil {
		err = tfresource.NewEmptyResultError()
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Aurora DSQL Cluster (%s) peering create", id), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *clusterPeeringResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data clusterPeeringResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DSQLClient(ctx)

	output, err := findClusterByID(ctx, conn, data.Identifier.ValueString())

	if err == nil && output.MultiRegionProperties == nil {
		err = tfresource.NewEmptyResultError()
	}

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating Aurora DSQL Cluster (%s)", data.Identifier.ValueString()), err.Error())

		return
	}

	properties := normalizeMultiRegionProperties(output)
	response.Diagnostics.Append(fwflex.Flatten(ctx, properties, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *clusterPeeringResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrIdentifier), request, response)
}

func waitClusterPeeringCreated(ctx context.Context, conn *dsql.Client, id string, timeout time.Duration) (*dsql.GetClusterOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ClusterStatusUpdating, awstypes.ClusterStatusPendingSetup, awstypes.ClusterStatusCreating),
		Target:                    enum.Slice(awstypes.ClusterStatusActive),
		Refresh:                   statusCluster(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dsql.GetClusterOutput); ok {
		return output, err
	}

	return nil, err
}

type clusterPeeringResourceModel struct {
	framework.WithRegionModel
	Clusters      fwtypes.SetOfARN `tfsdk:"clusters"`
	Identifier    types.String     `tfsdk:"identifier"`
	Timeouts      timeouts.Value   `tfsdk:"timeouts"`
	WitnessRegion types.String     `tfsdk:"witness_region"`
}
