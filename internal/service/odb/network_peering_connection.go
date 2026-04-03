// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package odb

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
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

var OracleDBNetworkPeeringConnection = newResourceNetworkPeeringConnection
var ResourceNetworkPeeringConnection resourceNetworkPeeringConnection

type resourceNetworkPeeringConnection struct {
	framework.ResourceWithModel[odbNetworkPeeringConnectionResourceModel]
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *resourceNetworkPeeringConnection) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A peering connection between an ODB network and either another ODB network or a customer-owned VPC.",
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			"odb_network_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
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

			names.AttrDisplayName: schema.StringAttribute{
				Description: "Display name of the odb network peering connection. Changing this will force terraform to create new resource",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			names.AttrStatus: schema.StringAttribute{
				Description: "Status of the odb network peering connection.",
				CustomType:  fwtypes.StringEnumType[odbtypes.ResourceStatus](),
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrStatusReason: schema.StringAttribute{
				Description: "The reason for the current status of the ODB peering connection..",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"odb_network_arn": schema.StringAttribute{
				Description: "ARN of the odb network peering connection.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"peer_network_arn": schema.StringAttribute{
				Description: "ARN of the peer network peering connection.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"peer_network_cidrs": schema.SetAttribute{
				Description: "List of peered network cidrs.",
				CustomType:  fwtypes.SetOfStringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"odb_peering_connection_type": schema.StringAttribute{
				Description: "Type of the odb peering connection.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrCreatedAt: schema.StringAttribute{
				Description: "Created time of the odb network peering connection.",
				Computed:    true,
				CustomType:  timetypes.RFC3339Type{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"percent_progress": schema.Float32Attribute{
				Description: "Progress of the odb network peering connection.",
				Computed:    true,
				PlanModifiers: []planmodifier.Float32{
					float32planmodifier.UseStateForUnknown(),
				},
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

func (r *resourceNetworkPeeringConnection) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data odbNetworkPeeringConnectionResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	//Neither is present
	if data.OdbNetworkId.IsNull() && data.OdbNetworkArn.IsNull() {
		err := errors.New("either odb_network_id or odb_network_arn must be present. Neither is present.")
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameNetworkPeeringConnection, data.DisplayName.String(), err),
			err.Error(),
		)
		return
	}

	//Both are present
	if !data.OdbNetworkId.IsNull() && !data.OdbNetworkArn.IsNull() {
		err := errors.New("either odb_network_id or odb_network_arn must be present. Both are present.")
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameNetworkPeeringConnection, data.DisplayName.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceNetworkPeeringConnection) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ODBClient(ctx)
	var plan odbNetworkPeeringConnectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	odbNetwork := plan.OdbNetworkArn
	if odbNetwork.IsNull() || odbNetwork.IsUnknown() {
		odbNetwork = plan.OdbNetworkId
	}
	//Validation : check is there any peer cidr for removal
	if len(plan.PeerNetworkCidrs.Elements()) > 0 {
		err := errors.New("during creation add / removal of peer network cidr is not supported")
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameNetworkPeeringConnection, plan.DisplayName.ValueString(), err),
			err.Error(),
		)
		return
	}
	input := odb.CreateOdbPeeringConnectionInput{
		OdbNetworkId:  odbNetwork.ValueStringPointer(),
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
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), aws.ToString(out.OdbPeeringConnectionId))...)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForCreation, ResNameNetworkPeeringConnection, plan.DisplayName.ValueString(), err),
			err.Error(),
		)
		return
	}

	odbNetworkARNParsed, err := arn.Parse(*createdPeeredConnection.OdbNetworkArn)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameNetworkPeeringConnection, plan.DisplayName.ValueString(), err),
			err.Error(),
		)
		return
	}
	peerVpcARN, err := arn.Parse(*createdPeeredConnection.PeerNetworkArn)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameNetworkPeeringConnection, plan.DisplayName.ValueString(), err),
			err.Error(),
		)
		return
	}
	plan.PeerNetworkId = types.StringValue(strings.Split(peerVpcARN.Resource, "/")[1])
	plan.OdbNetworkId = types.StringValue(strings.Split(odbNetworkARNParsed.Resource, "/")[1])
	resp.Diagnostics.Append(flex.Flatten(ctx, createdPeeredConnection, &plan)...)
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
	if retry.NotFound(err) {
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
	state.OdbNetworkId = types.StringValue(strings.Split(odbNetworkARNParsed.Resource, "/")[1])

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
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
	diff, d := flex.Diff(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	if diff.HasChanges() {
		var input odb.UpdateOdbPeeringConnectionInput
		var planPeeredCidrs []string
		plan.PeerNetworkCidrs.ElementsAs(ctx, &planPeeredCidrs, false)
		var statePeeredCidrs []string
		state.PeerNetworkCidrs.ElementsAs(ctx, &statePeeredCidrs, false)
		addedRemovedCidrs := r.FindAddRemovePeeredNetworkCIDR(planPeeredCidrs, statePeeredCidrs)
		if len(addedRemovedCidrs) > 0 {
			var addedPeeredCidrs []string
			var removedPeeredCidrs []string

			for k, v := range addedRemovedCidrs {
				switch v {
				case -1:
					removedPeeredCidrs = append(removedPeeredCidrs, k)
				case 1:
					addedPeeredCidrs = append(addedPeeredCidrs, k)
				}
			}
			if len(removedPeeredCidrs) > 0 {
				input.PeerNetworkCidrsToBeRemoved = removedPeeredCidrs
			}
			if len(addedPeeredCidrs) > 0 {
				input.PeerNetworkCidrsToBeAdded = addedPeeredCidrs
			}
		}

		input.OdbPeeringConnectionId = state.OdbPeeringConnectionId.ValueStringPointer()
		out, err := conn.UpdateOdbPeeringConnection(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionUpdating, ResNameNetworkPeeringConnection, state.OdbPeeringConnectionId.ValueString(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.OdbPeeringConnectionId == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionUpdating, ResNameNetwork, state.OdbPeeringConnectionId.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
	}
	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	updatedNetworkPeeringConnections, err := waitNetworkPeeringConnectionUpdated(ctx, conn, state.OdbPeeringConnectionId.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForUpdate, ResNameNetworkPeeringConnection, state.OdbPeeringConnectionId.String(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, updatedNetworkPeeringConnections, &plan)...)
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

func (r *resourceNetworkPeeringConnection) FindAddRemovePeeredNetworkCIDR(planCIDRs, sateCiders []string) map[string]int {
	addedRemovedCidrs := make(map[string]int)
	//1 indicates newly added cidrs. Here we are assuming that all cidrs are new.
	for _, nCidr := range planCIDRs {
		addedRemovedCidrs[nCidr] = 1
	}
	//Now lets remove those which are not present
	for _, oCidr := range sateCiders {
		//if cidr is present in the map; that means no change is required for that cidr so remove it
		_, ok := addedRemovedCidrs[oCidr]
		if ok {
			delete(addedRemovedCidrs, oCidr)
		} else {
			addedRemovedCidrs[oCidr] = -1
		}
	}
	return addedRemovedCidrs
}

func waitNetworkPeeringConnectionCreated(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.OdbPeeringConnection, error) {
	stateConf := &sdkretry.StateChangeConf{
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
	stateConf := &sdkretry.StateChangeConf{
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
	stateConf := &sdkretry.StateChangeConf{
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

func statusNetworkPeeringConnection(ctx context.Context, conn *odb.Client, id string) sdkretry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findNetworkPeeringConnectionByID(ctx, conn, id)
		if retry.NotFound(err) {
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
			return nil, &sdkretry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		return nil, err
	}

	if out == nil || out.OdbPeeringConnection == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out.OdbPeeringConnection, nil
}

type odbNetworkPeeringConnectionResourceModel struct {
	framework.WithRegionModel
	OdbNetworkId             types.String                                `tfsdk:"odb_network_id" autoflex:",noflatten"`
	PeerNetworkId            types.String                                `tfsdk:"peer_network_id" autoflex:",noflatten"`
	OdbPeeringConnectionId   types.String                                `tfsdk:"id"`
	DisplayName              types.String                                `tfsdk:"display_name"`
	Status                   fwtypes.StringEnum[odbtypes.ResourceStatus] `tfsdk:"status"`
	StatusReason             types.String                                `tfsdk:"status_reason"`
	OdbPeeringConnectionArn  types.String                                `tfsdk:"arn"`
	OdbNetworkArn            types.String                                `tfsdk:"odb_network_arn"`
	PeerNetworkArn           types.String                                `tfsdk:"peer_network_arn"`
	PeerNetworkCidrs         fwtypes.SetValueOf[types.String]            `tfsdk:"peer_network_cidrs"`
	OdbPeeringConnectionType types.String                                `tfsdk:"odb_peering_connection_type"`
	CreatedAt                timetypes.RFC3339                           `tfsdk:"created_at"`
	PercentProgress          types.Float32                               `tfsdk:"percent_progress"`
	Timeouts                 timeouts.Value                              `tfsdk:"timeouts"`
	Tags                     tftags.Map                                  `tfsdk:"tags"`
	TagsAll                  tftags.Map                                  `tfsdk:"tags_all"`
}
