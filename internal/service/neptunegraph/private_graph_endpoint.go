// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package neptunegraph

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/neptunegraph"
	awstypes "github.com/aws/aws-sdk-go-v2/service/neptunegraph/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_neptunegraph_private_graph_endpoint", name="Private Graph Endpoint")
func newResourcePrivateGraphEndpoint(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourcePrivateGraphEndpoint{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNamePrivateGraphEndpoint = "Private Graph Endpoint"
)

type resourcePrivateGraphEndpoint struct {
	framework.ResourceWithModel[resourcePrivateGraphEndpointModel]
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *resourcePrivateGraphEndpoint) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"graph_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"private_graph_endpoint_identifier": schema.StringAttribute{
				Computed: true,
			},
			names.AttrSecurityGroupIDs: schema.SetAttribute{
				CustomType: fwtypes.SetOfStringType,
				Optional:   true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			names.AttrSubnetIDs: schema.SetAttribute{
				CustomType: fwtypes.SetOfStringType,
				Optional:   true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			names.AttrVPCID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrVPCEndpointID: schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourcePrivateGraphEndpoint) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().NeptuneGraphClient(ctx)

	var plan resourcePrivateGraphEndpointModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input neptunegraph.CreatePrivateGraphEndpointInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreatePrivateGraphEndpoint(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.GraphIdentifier.ValueString()+"_"+plan.VpcId.ValueString())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.GraphIdentifier.ValueString()+"_"+plan.VpcId.ValueString())
		return
	}

	plan.Id = types.StringValue(plan.GraphIdentifier.ValueString() + "_" + aws.ToString(out.VpcId))
	plan.PrivateGraphEndpointIdentifier = types.StringValue(plan.GraphIdentifier.ValueString() + "_" + aws.ToString(out.VpcId))

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitPrivateGraphEndpointAvailable(ctx, conn, plan.GraphIdentifier.ValueString(), plan.VpcId.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Id.ValueString())
		return
	}

	out2, err := findPrivateGraphEndpointByID(ctx, conn, plan.Id.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Id.ValueString())
		return
	}

	plan.VpcEndpointId = types.StringValue(aws.ToString(out2.VpcEndpointId))

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourcePrivateGraphEndpoint) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().NeptuneGraphClient(ctx)

	var state resourcePrivateGraphEndpointModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findPrivateGraphEndpointByID(ctx, conn, state.Id.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Id.ValueString())
		return
	}

	parts := strings.Split(state.Id.ValueString(), "_")
	if len(parts) == 2 {
		state.GraphIdentifier = types.StringValue(parts[0])
		state.VpcId = types.StringValue(parts[1])
	}

	state.PrivateGraphEndpointIdentifier = state.Id
	state.VpcEndpointId = types.StringValue(aws.ToString(out.VpcEndpointId))

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourcePrivateGraphEndpoint) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().NeptuneGraphClient(ctx)

	var state resourcePrivateGraphEndpointModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := neptunegraph.DeletePrivateGraphEndpointInput{
		GraphIdentifier: state.GraphIdentifier.ValueStringPointer(),
		VpcId:           state.VpcId.ValueStringPointer(),
	}

	_, err := conn.DeletePrivateGraphEndpoint(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Id.ValueString())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitPrivateGraphEndpointDeleted(ctx, conn, state.GraphIdentifier.ValueString(), state.VpcId.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Id.ValueString())
		return
	}
}

func findPrivateGraphEndpointByID(ctx context.Context, conn *neptunegraph.Client, id string) (*neptunegraph.GetPrivateGraphEndpointOutput, error) {
	parts := strings.Split(id, "_")
	if len(parts) != 2 {
		return nil, errors.New("invalid ID format, expected graph_id_vpc_id")
	}

	input := neptunegraph.GetPrivateGraphEndpointInput{
		GraphIdentifier: aws.String(parts[0]),
		VpcId:           aws.String(parts[1]),
	}

	out, err := conn.GetPrivateGraphEndpoint(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&sdkretry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

func statusPrivateGraphEndpoint(conn *neptunegraph.Client, graphID, vpcID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findPrivateGraphEndpointByID(ctx, conn, graphID+"_"+vpcID)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}
		return output, string(output.Status), nil
	}
}

func waitPrivateGraphEndpointAvailable(ctx context.Context, conn *neptunegraph.Client, graphID, vpcID string, timeout time.Duration) (*neptunegraph.GetPrivateGraphEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PrivateGraphEndpointStatusCreating),
		Target:  enum.Slice(awstypes.PrivateGraphEndpointStatusAvailable),
		Refresh: statusPrivateGraphEndpoint(conn, graphID, vpcID),
		Timeout: timeout,
	}
	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*neptunegraph.GetPrivateGraphEndpointOutput); ok {
		return output, err
	}
	return nil, err
}

func waitPrivateGraphEndpointDeleted(ctx context.Context, conn *neptunegraph.Client, graphID, vpcID string, timeout time.Duration) (*neptunegraph.GetPrivateGraphEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PrivateGraphEndpointStatusDeleting),
		Target:  []string{},
		Refresh: statusPrivateGraphEndpoint(conn, graphID, vpcID),
		Timeout: timeout,
	}
	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*neptunegraph.GetPrivateGraphEndpointOutput); ok {
		return output, err
	}
	return nil, err
}

type resourcePrivateGraphEndpointModel struct {
	framework.WithRegionModel
	GraphIdentifier                types.String        `tfsdk:"graph_identifier"`
	Id                             types.String        `tfsdk:"id"`
	PrivateGraphEndpointIdentifier types.String        `tfsdk:"private_graph_endpoint_identifier"`
	SecurityGroupIDs               fwtypes.SetOfString `tfsdk:"security_group_ids"`
	SubnetIDs                      fwtypes.SetOfString `tfsdk:"subnet_ids"`
	Timeouts                       timeouts.Value      `tfsdk:"timeouts"`
	VpcId                          types.String        `tfsdk:"vpc_id"`
	VpcEndpointId                  types.String        `tfsdk:"vpc_endpoint_id"`
}
