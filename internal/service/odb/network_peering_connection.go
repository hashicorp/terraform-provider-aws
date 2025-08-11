//Copyright © 2025, Oracle and/or its affiliates. All rights reserved.

package odb

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_odb_network_peering_connection", name="Network Peering Connection")
// @Tags(identifierAttribute="arn")
func newResourceNetworkPeeringConnection(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceNetworkPeeringConnection{}

	r.SetDefaultCreateTimeout(24 * time.Hour)
	r.SetDefaultUpdateTimeout(24 * time.Hour)
	r.SetDefaultDeleteTimeout(24 * time.Hour)

	return r, nil
}

const (
	ResNameNetworkPeeringConnection = "Network Peering Connection"
)

var OdbNetworkPeeringConnection = newResourceNetworkPeeringConnection

type resourceNetworkPeeringConnection struct {
	framework.ResourceWithModel[odbNetworkPeeringConnectionResourceModel]
	framework.WithTimeouts
}

func (r *resourceNetworkPeeringConnection) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Information about an ODB network. Also refer odb_network_peering resource : A peering connection between an ODB network and either another ODB network or a customer-owned VPC.",
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			"odb_network_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Required field. The unique identifier of the ODB network that initiates the peering connection. " +
					"A sample ID is odbpcx-abcdefgh12345678. Changing this will force terraform to create new resource.",
			},
			"peer_network_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Required field. The unique identifier of the ODB peering connection. Changing this will force terraform to create new resource",
			},

			"display_name": schema.StringAttribute{
				Description: "Display name of the odb network peering connection. Changing this will force terraform to create new resource",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"status": schema.StringAttribute{
				Description: "Status of the odb network peering connection.",
				CustomType:  fwtypes.StringEnumType[odbtypes.ResourceStatus](),
				Computed:    true,
			},
			"status_reason": schema.StringAttribute{
				Description: "The reason for the current status of the ODB peering connection..",
				Computed:    true,
			},

			"odb_network_arn": schema.StringAttribute{
				Description: "ARN of the odb network peering connection.",
				Computed:    true,
			},

			"peer_network_arn": schema.StringAttribute{
				Description: "ARN of the peer network peering connection.",
				Computed:    true,
			},
			"odb_peering_connection_type": schema.StringAttribute{
				Description: "Type of the odb peering connection.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Created time of the odb network peering connection.",
				Computed:    true,
			},
			"percent_progress": schema.Float32Attribute{
				Description: "Progress of the odb network peering connection.",
				Computed:    true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceNetworkPeeringConnection) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ODBClient(ctx)

	var plan odbNetworkPeeringConnectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := odb.CreateOdbPeeringConnectionInput{
		OdbNetworkId:  plan.OdbNetworkId.ValueStringPointer(),
		PeerNetworkId: plan.PeerNetworkId.ValueStringPointer(),
		DisplayName:   plan.DisplayName.ValueStringPointer(),
		Tags:          getTagsIn(ctx),
	}

	out, err := conn.CreateOdbPeeringConnection(ctx, &input)
	if err != nil {

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameNetworkPeeringConnection, plan.DisplayName.ValueString(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.OdbPeeringConnectionId == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameNetworkPeeringConnection, plan.DisplayName.ValueString(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	createdPeeredConnection, err := waitNetworkPeeringConnectionCreated(ctx, conn, plan.OdbPeeringConnectionId.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForCreation, ResNameNetworkPeeringConnection, plan.DisplayName.ValueString(), err),
			err.Error(),
		)
		return
	}
	plan.CreatedAt = types.StringValue(createdPeeredConnection.CreatedAt.Format(time.RFC3339))
	resp.Diagnostics.Append(flex.Flatten(ctx, createdPeeredConnection, &plan, flex.WithIgnoredFieldNamesAppend("CreatedAt"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceNetworkPeeringConnection) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ODBClient(ctx)
	var state odbNetworkPeeringConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findNetworkPeeringConnectionByID(ctx, conn, state.OdbPeeringConnectionId.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameNetworkPeeringConnection, state.OdbPeeringConnectionId.ValueString(), err),
			err.Error(),
		)
		return
	}

	odbNetworkARNParsed, err := arn.Parse(*out.OdbNetworkArn)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameNetworkPeeringConnection, state.OdbPeeringConnectionId.ValueString(), err),
			err.Error(),
		)
		return
	}

	peerVpcARN, err := arn.Parse(*out.PeerNetworkArn)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameNetworkPeeringConnection, state.OdbPeeringConnectionId.ValueString(), err),
			err.Error(),
		)
		return
	}
	state.PeerNetworkId = types.StringValue(strings.Split(peerVpcARN.Resource, "/")[1])
	state.CreatedAt = types.StringValue(out.CreatedAt.Format(time.RFC3339))
	state.OdbNetworkId = types.StringValue(strings.Split(odbNetworkARNParsed.Resource, "/")[1])

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state, flex.WithIgnoredFieldNamesAppend("CreatedAt"),
		flex.WithIgnoredFieldNamesAppend("PeerNetworkId"), flex.WithIgnoredFieldNamesAppend("OdbNetworkId"))...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceNetworkPeeringConnection) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().ODBClient(ctx)

	var plan, state odbNetworkPeeringConnectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	updatedOdbNetPeeringConn, err := waitNetworkPeeringConnectionUpdated(ctx, conn, plan.OdbPeeringConnectionId.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForUpdate, ResNameNetworkPeeringConnection, plan.OdbPeeringConnectionId.ValueString(), err),
			err.Error(),
		)
		return
	}
	odbNetworkARNParsed, err := arn.Parse(*updatedOdbNetPeeringConn.OdbNetworkArn)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameNetworkPeeringConnection, state.OdbPeeringConnectionId.ValueString(), err),
			err.Error(),
		)
		return
	}

	peerVpcARN, err := arn.Parse(*updatedOdbNetPeeringConn.PeerNetworkArn)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameNetworkPeeringConnection, state.OdbPeeringConnectionId.ValueString(), err),
			err.Error(),
		)
		return
	}
	state.PeerNetworkId = types.StringValue(strings.Split(peerVpcARN.Resource, "/")[1])
	state.CreatedAt = types.StringValue(updatedOdbNetPeeringConn.CreatedAt.Format(time.RFC3339))
	state.OdbNetworkId = types.StringValue(strings.Split(odbNetworkARNParsed.Resource, "/")[1])
	resp.Diagnostics.Append(flex.Flatten(ctx, updatedOdbNetPeeringConn, &plan, flex.WithIgnoredFieldNamesAppend("CreatedAt"), flex.WithIgnoredFieldNamesAppend("CreatedAt"),
		flex.WithIgnoredFieldNamesAppend("PeerNetworkId"), flex.WithIgnoredFieldNamesAppend("OdbNetworkId"))...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceNetworkPeeringConnection) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ODBClient(ctx)
	var state odbNetworkPeeringConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := odb.DeleteOdbPeeringConnectionInput{
		OdbPeeringConnectionId: state.OdbPeeringConnectionId.ValueStringPointer(),
	}

	_, err := conn.DeleteOdbPeeringConnection(ctx, &input)

	if err != nil {
		if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionDeleting, ResNameNetworkPeeringConnection, state.OdbPeeringConnectionId.ValueString(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitNetworkPeeringConnectionDeleted(ctx, conn, state.OdbPeeringConnectionId.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForDeletion, ResNameNetworkPeeringConnection, state.OdbPeeringConnectionId.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceNetworkPeeringConnection) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func waitNetworkPeeringConnectionCreated(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.OdbPeeringConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(odbtypes.ResourceStatusProvisioning),
		Target:                    enum.Slice(odbtypes.ResourceStatusAvailable, odbtypes.ResourceStatusFailed),
		Refresh:                   statusNetworkPeeringConnection(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odbtypes.OdbPeeringConnection); ok {
		return out, err
	}

	return nil, err
}

func waitNetworkPeeringConnectionUpdated(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.OdbPeeringConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(odbtypes.ResourceStatusUpdating),
		Target:                    enum.Slice(odbtypes.ResourceStatusAvailable, odbtypes.ResourceStatusFailed),
		Refresh:                   statusNetworkPeeringConnection(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odbtypes.OdbPeeringConnection); ok {
		return out, err
	}

	return nil, err
}

func waitNetworkPeeringConnectionDeleted(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.OdbPeeringConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(odbtypes.ResourceStatusTerminating),
		Target:  []string{},
		Refresh: statusNetworkPeeringConnection(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odbtypes.OdbPeeringConnection); ok {
		return out, err
	}

	return nil, err
}

func statusNetworkPeeringConnection(ctx context.Context, conn *odb.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findNetworkPeeringConnectionByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findNetworkPeeringConnectionByID(ctx context.Context, conn *odb.Client, id string) (*odbtypes.OdbPeeringConnection, error) {
	input := odb.GetOdbPeeringConnectionInput{
		OdbPeeringConnectionId: &id,
	}

	out, err := conn.GetOdbPeeringConnection(ctx, &input)
	if err != nil {
		if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		return nil, err
	}

	if out == nil || out.OdbPeeringConnection == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return out.OdbPeeringConnection, nil
}

type odbNetworkPeeringConnectionResourceModel struct {
	framework.WithRegionModel
	OdbNetworkId             types.String                                `tfsdk:"odb_network_id"`
	PeerNetworkId            types.String                                `tfsdk:"peer_network_id"`
	OdbPeeringConnectionId   types.String                                `tfsdk:"id"`
	DisplayName              types.String                                `tfsdk:"display_name"`
	Status                   fwtypes.StringEnum[odbtypes.ResourceStatus] `tfsdk:"status"`
	StatusReason             types.String                                `tfsdk:"status_reason"`
	OdbPeeringConnectionArn  types.String                                `tfsdk:"arn"`
	OdbNetworkArn            types.String                                `tfsdk:"odb_network_arn"`
	PeerNetworkArn           types.String                                `tfsdk:"peer_network_arn"`
	OdbPeeringConnectionType types.String                                `tfsdk:"odb_peering_connection_type"`
	CreatedAt                types.String                                `tfsdk:"created_at"`
	PercentProgress          types.Float32                               `tfsdk:"percent_progress"`
	Timeouts                 timeouts.Value                              `tfsdk:"timeouts"`
	Tags                     tftags.Map                                  `tfsdk:"tags"`
	TagsAll                  tftags.Map                                  `tfsdk:"tags_all"`
}

func sweepNetworkPeeringConnections(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := odb.ListOdbPeeringConnectionsInput{}
	conn := client.ODBClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := odb.NewListOdbPeeringConnectionsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.OdbPeeringConnections {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceNetworkPeeringConnection, client,
				sweepfw.NewAttribute(names.AttrID, aws.ToString(v.OdbPeeringConnectionId))),
			)
		}
	}

	return sweepResources, nil
}
