// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_networkfirewall_vpc_endpoint_association", name="VPC Endpoint Association")
// @Tags(identifierAttribute="firewall_arn")
func newVPCEndpointAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceVPCEndpointAssociation{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameVPCEndpointAssociation = "VPC Endpoint Association"
)

type resourceVPCEndpointAssociation struct {
	framework.ResourceWithModel[resourceVPCEndpointAssociationModel]
	framework.WithTimeouts
	framework.WithNoUpdate
}

func (r *resourceVPCEndpointAssociation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"firewall_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"vpc_endpoint_association_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"vpc_endpoint_association_arn": framework.ARNAttributeComputedOnly(),
			names.AttrVPCID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"subnet_mapping": schema.ListNestedBlock{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[subnetMappingModel](ctx),
				Description: "A list of subnet mappings for the VPC endpoint association.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrIPAddressType: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.IPAddressType](),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						names.AttrSubnetID: schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			"vpc_endpoint_association_status": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[vpcEndpointAssociationStatusModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrStatus: schema.StringAttribute{
							Computed: true,
						},
					},
					Blocks: map[string]schema.Block{
						"association_sync_state": schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[AZSyncStateModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrAvailabilityZone: schema.StringAttribute{
										Computed: true,
									},
								},
								Blocks: map[string]schema.Block{
									"attachment": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[attachmentModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"endpoint_id": schema.StringAttribute{
													Computed: true,
												},
												names.AttrSubnetID: schema.StringAttribute{
													Computed: true,
												},
												names.AttrStatus: schema.StringAttribute{
													Computed: true,
												},
												names.AttrStatusMessage: schema.StringAttribute{
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceVPCEndpointAssociation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var plan resourceVPCEndpointAssociationModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input networkfirewall.CreateVpcEndpointAssociationInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)
	out, err := conn.CreateVpcEndpointAssociation(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionCreating, ResNameVPCEndpointAssociation, "", err),
			err.Error(),
		)
		return
	}
	if out == nil || out.VpcEndpointAssociation == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionCreating, ResNameVPCEndpointAssociation, "", nil),
			errors.New("empty output").Error(),
		)
		return
	}
	arn := aws.ToString(out.VpcEndpointAssociation.VpcEndpointAssociationArn)
	resp.Diagnostics.Append(flex.Flatten(ctx, out.VpcEndpointAssociation, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out.VpcEndpointAssociationStatus, &plan.VpcEndpointAssociationStatus, flex.WithIgnoredFieldNamesAppend("AssociationSyncState"))...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.VPCEndpointAssociationARN = flex.StringValueToFramework(ctx, arn)
	plan.setID()

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	created, err := waitVPCEndpointAssociationCreated(ctx, conn, arn, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionWaitingForCreation, ResNameVPCEndpointAssociation, arn, err),
			err.Error(),
		)
		return
	}
	// AZState is a map and needs to be flattened into a set of objects
	vpcEndpointAssociationStatusModel, d := plan.VpcEndpointAssociationStatus.ToPtr(ctx)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(vpcEndpointAssociationStatusModel.flattenAZSyncState(ctx, created.VpcEndpointAssociationStatus.AssociationSyncState)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceVPCEndpointAssociation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var state resourceVPCEndpointAssociationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := findVPCEndpointAssociationByID(ctx, conn, state.ID.ValueString())

	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionReading, ResNameVPCEndpointAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state, flex.WithIgnoredFieldNamesAppend("VpcEndpointAssociationStatus"))...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out.VpcEndpointAssociationStatus, &state.VpcEndpointAssociationStatus, flex.WithIgnoredFieldNamesAppend("AssociationSyncState"))...)
	if resp.Diagnostics.HasError() {
		return
	}
	vpcEndpointAssociationStatusModel, d := state.VpcEndpointAssociationStatus.ToPtr(ctx)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(vpcEndpointAssociationStatusModel.flattenAZSyncState(ctx, out.VpcEndpointAssociationStatus.AssociationSyncState)...)

	setTagsOut(ctx, out.VpcEndpointAssociation.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceVPCEndpointAssociation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var state resourceVPCEndpointAssociationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := networkfirewall.DeleteVpcEndpointAssociationInput{
		VpcEndpointAssociationArn: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteVpcEndpointAssociation(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionDeleting, ResNameVPCEndpointAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitVPCEndpointAssociationDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionWaitingForDeletion, ResNameVPCEndpointAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceVPCEndpointAssociation) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) { // nosemgrep:ci.semgrep.framework.with-import-by-id
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(""), request.ID)...)
}

func waitVPCEndpointAssociationCreated(ctx context.Context, conn *networkfirewall.Client, arn string, timeout time.Duration) (*networkfirewall.DescribeVpcEndpointAssociationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.FirewallStatusValueProvisioning),
		Target:                    enum.Slice(awstypes.FirewallStatusValueReady),
		Refresh:                   statusVPCEndpointAssociation(ctx, conn, arn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*networkfirewall.DescribeVpcEndpointAssociationOutput); ok {
		return out, err
	}

	return nil, err
}

func waitVPCEndpointAssociationDeleted(ctx context.Context, conn *networkfirewall.Client, arn string, timeout time.Duration) (*networkfirewall.DescribeVpcEndpointAssociationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.FirewallStatusValueReady, awstypes.FirewallStatusValueDeleting),
		Target:                    []string{},
		Refresh:                   statusVPCEndpointAssociation(ctx, conn, arn),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*networkfirewall.DescribeVpcEndpointAssociationOutput); ok {
		return out, err
	}

	return nil, err
}

func statusVPCEndpointAssociation(ctx context.Context, conn *networkfirewall.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findVPCEndpointAssociationByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.VpcEndpointAssociationStatus.Status), nil
	}
}

func findVPCEndpointAssociationByID(ctx context.Context, conn *networkfirewall.Client, arn string) (*networkfirewall.DescribeVpcEndpointAssociationOutput, error) {
	input := networkfirewall.DescribeVpcEndpointAssociationInput{
		VpcEndpointAssociationArn: aws.String(arn),
	}

	out, err := conn.DescribeVpcEndpointAssociation(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return out, nil
}

type resourceVPCEndpointAssociationModel struct {
	framework.WithRegionModel
	Description                  types.String                                                       `tfsdk:"description"`
	FirewallARN                  fwtypes.ARN                                                        `tfsdk:"firewall_arn"`
	SubnetMapping                fwtypes.ListNestedObjectValueOf[subnetMappingModel]                `tfsdk:"subnet_mapping"`
	Tags                         tftags.Map                                                         `tfsdk:"tags"`
	TagsAll                      tftags.Map                                                         `tfsdk:"tags_all"`
	Timeouts                     timeouts.Value                                                     `tfsdk:"timeouts"`
	VPCEndpointAssociationARN    types.String                                                       `tfsdk:"vpc_endpoint_association_arn"`
	VPCEndpointAssociationID     types.String                                                       `tfsdk:"vpc_endpoint_association_id"`
	VpcEndpointAssociationStatus fwtypes.ListNestedObjectValueOf[vpcEndpointAssociationStatusModel] `tfsdk:"vpc_endpoint_association_status"`
	VPCID                        types.String                                                       `tfsdk:"vpc_id"`
}

type subnetMappingModel struct {
	SubnetId      types.String                               `tfsdk:"subnet_id"`
	IPAddressType fwtypes.StringEnum[awstypes.IPAddressType] `tfsdk:"ip_address_type"`
}

type vpcEndpointAssociationStatusModel struct {
	Status               types.String                                     `tfsdk:"status"`
	AssociationSyncState fwtypes.SetNestedObjectValueOf[AZSyncStateModel] `tfsdk:"association_sync_state"`
}

type AZSyncStateModel struct {
	Attachment       fwtypes.ListNestedObjectValueOf[attachmentModel] `tfsdk:"attachment"`
	AvailabilityZone types.String                                     `tfsdk:"availability_zone"`
}

type attachmentModel struct {
	EndpointId    types.String `tfsdk:"endpoint_id"`
	SubnetId      types.String `tfsdk:"subnet_id"`
	Status        types.String `tfsdk:"status"`
	StatusMessage types.String `tfsdk:"status_message"`
}

func (m *vpcEndpointAssociationStatusModel) flattenAZSyncState(ctx context.Context, azSyncStateMap map[string]awstypes.AZSyncState) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(azSyncStateMap) == 0 {
		return diags
	}
	azSyncStates := make([]*AZSyncStateModel, 0, len(azSyncStateMap))
	for az, syncState := range azSyncStateMap {
		azSyncState := &AZSyncStateModel{
			AvailabilityZone: flex.StringValueToFramework(ctx, az),
		}
		var attachment attachmentModel
		diags.Append(flex.Flatten(ctx, syncState.Attachment, &attachment)...)
		if diags.HasError() {
			return diags
		}
		azSyncState.Attachment = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &attachment)
		azSyncStates = append(azSyncStates, azSyncState)
	}
	m.AssociationSyncState = fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, azSyncStates)

	return diags
}
