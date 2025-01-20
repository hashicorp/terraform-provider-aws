// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"
	"errors"
	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_vpclattice_resource_configuration", name="Resource Configuration")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/vpclattice;vpclattice.GetResourceConfigurationOutput")
func newResourceResourceConfiguration(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceResourceConfiguration{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameResourceConfiguration = "Resource Configuration"
)

type resourceResourceConfiguration struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourceResourceConfiguration) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_vpclattice_resource_configuration"
}

func (r *resourceResourceConfiguration) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"allow_association_to_shareable_service_network": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"arn": framework.ARNAttributeComputedOnly(),
			"id":  framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"port_ranges": schema.ListAttribute{
				CustomType: fwtypes.ListOfStringType,
				Optional:   true,
				Computed:   true,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(regexache.MustCompile("^((\\d{1,5}\\-\\d{1,5})|(\\d+))$"), "must contain one port number between 1 and 65535 or two seperated by hyphen.")),
				},
			},
			"protocol": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ProtocolType](),
				Optional:   true,
				Computed:   true,
				Default:    stringdefault.StaticString(string(awstypes.ProtocolTypeTcp)),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_gateway_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_configuration_group_id": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{stringvalidator.ConflictsWith(
					path.MatchRelative().AtParent().AtName("resource_gateway_identifier"),
				)},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ResourceConfigurationType](),
				Optional:   true,
				Computed:   true,
				Default:    stringdefault.StaticString(string(awstypes.ResourceConfigurationTypeSingle)),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"resource_configuration_definition": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[resourceConfigurationDefinitionModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"arn_resource": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[arnResourceModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrARN: schema.StringAttribute{
										Required: true,
									},
								},
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("arn_resource"),
									path.MatchRelative().AtParent().AtName("ip_resource"),
									path.MatchRelative().AtParent().AtName("dns_resource"),
								),
							},
						},
						"dns_resource": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[dnsResourceModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"domain_name": schema.StringAttribute{
										Required: true,
									},
									"ip_address_type": schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.IpAddressType](),
									},
								},
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
						},
						"ip_resource": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[ipResourceModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"ip_address": schema.StringAttribute{
										Required: true,
									},
								},
							},
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
						},
					},
				},
			},
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceResourceConfiguration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().VPCLatticeClient(ctx)

	var plan resourceResourceConfigurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input vpclattice.CreateResourceConfigurationInput

	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateResourceConfiguration(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VPCLattice, create.ErrActionCreating, ResNameResourceConfiguration, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VPCLattice, create.ErrActionCreating, ResNameResourceConfiguration, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ResourceGatewayIdentifier = flex.StringToFramework(ctx, out.ResourceGatewayId)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitResourceConfigurationCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VPCLattice, create.ErrActionWaitingForCreation, ResNameResourceConfiguration, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceResourceConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().VPCLatticeClient(ctx)

	var state resourceResourceConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findResourceConfigurationByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VPCLattice, create.ErrActionSetting, ResNameResourceConfiguration, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.ResourceGatewayIdentifier = flex.StringToFramework(ctx, out.ResourceGatewayId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceResourceConfiguration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().VPCLatticeClient(ctx)

	var plan, state resourceResourceConfigurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.PortRanges.Equal(state.PortRanges) ||
		!plan.ResourceConfigurationDefinition.Equal(state.ResourceConfigurationDefinition) {

		var input vpclattice.UpdateResourceConfigurationInput
		input.ResourceConfigurationIdentifier = plan.ID.ValueStringPointer()
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		rdc, diags := plan.ResourceConfigurationDefinition.ToPtr(ctx)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if rdc.IpResource.IsNull() {
			// DNS and ARN resources cannot be updated and must not be passed to update
			input.ResourceConfigurationDefinition = nil
		}

		out, err := conn.UpdateResourceConfiguration(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.VPCLattice, create.ErrActionUpdating, ResNameResourceConfiguration, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.VPCLattice, create.ErrActionUpdating, ResNameResourceConfiguration, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)

		if resp.Diagnostics.HasError() {
			return
		}
		plan.ResourceGatewayIdentifier = flex.StringToFramework(ctx, out.ResourceGatewayId)
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	updated, err := waitResourceConfigurationUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	resp.Diagnostics.Append(flex.Flatten(ctx, updated, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VPCLattice, create.ErrActionWaitingForUpdate, ResNameResourceConfiguration, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceResourceConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().VPCLatticeClient(ctx)

	var state resourceResourceConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := vpclattice.DeleteResourceConfigurationInput{
		ResourceConfigurationIdentifier: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteResourceConfiguration(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VPCLattice, create.ErrActionDeleting, ResNameResourceConfiguration, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitResourceConfigurationDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VPCLattice, create.ErrActionWaitingForDeletion, ResNameResourceConfiguration, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceResourceConfiguration) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func (r *resourceResourceConfiguration) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func waitResourceConfigurationCreated(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.GetResourceConfigurationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ResourceConfigurationStatusCreateInProgress),
		Target:                    enum.Slice(awstypes.ResourceConfigurationStatusActive),
		Refresh:                   statusResourceConfiguration(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*vpclattice.GetResourceConfigurationOutput); ok {
		return out, err
	}

	return nil, err
}

func waitResourceConfigurationUpdated(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.GetResourceConfigurationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ResourceConfigurationStatusUpdateInProgress),
		Target:                    enum.Slice(awstypes.ResourceConfigurationStatusActive),
		Refresh:                   statusResourceConfiguration(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*vpclattice.GetResourceConfigurationOutput); ok {
		return out, err
	}

	return nil, err
}

func waitResourceConfigurationDeleted(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.GetResourceConfigurationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceConfigurationStatusActive, awstypes.ResourceConfigurationStatusDeleteInProgress),
		Target:  []string{},
		Refresh: statusResourceConfiguration(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*vpclattice.GetResourceConfigurationOutput); ok {
		return out, err
	}

	return nil, err
}

func statusResourceConfiguration(ctx context.Context, conn *vpclattice.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findResourceConfigurationByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findResourceConfigurationByID(ctx context.Context, conn *vpclattice.Client, id string) (*vpclattice.GetResourceConfigurationOutput, error) {
	in := &vpclattice.GetResourceConfigurationInput{
		ResourceConfigurationIdentifier: aws.String(id),
	}

	out, err := conn.GetResourceConfiguration(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceResourceConfigurationModel struct {
	ARN                                       types.String                                                          `tfsdk:"arn"`
	AllowAssociationToShareableServiceNetwork types.Bool                                                            `tfsdk:"allow_association_to_shareable_service_network"`
	ID                                        types.String                                                          `tfsdk:"id"`
	Name                                      types.String                                                          `tfsdk:"name"`
	PortRanges                                fwtypes.ListOfString                                                  `tfsdk:"port_ranges"`
	Protocol                                  fwtypes.StringEnum[awstypes.ProtocolType]                             `tfsdk:"protocol"`
	ResourceConfigurationDefinition           fwtypes.ListNestedObjectValueOf[resourceConfigurationDefinitionModel] `tfsdk:"resource_configuration_definition"`
	ResourceGatewayIdentifier                 types.String                                                          `tfsdk:"resource_gateway_identifier"`
	ResourceConfigurationGroupId              types.String                                                          `tfsdk:"resource_configuration_group_id"`
	Tags                                      tftags.Map                                                            `tfsdk:"tags"`
	TagsAll                                   tftags.Map                                                            `tfsdk:"tags_all"`
	Timeouts                                  timeouts.Value                                                        `tfsdk:"timeouts"`
	Type                                      fwtypes.StringEnum[awstypes.ResourceConfigurationType]                `tfsdk:"type"`
}

type resourceConfigurationDefinitionModel struct {
	ArnResource fwtypes.ListNestedObjectValueOf[arnResourceModel] `tfsdk:"arn_resource"`
	IpResource  fwtypes.ListNestedObjectValueOf[ipResourceModel]  `tfsdk:"ip_resource"`
	DnsResource fwtypes.ListNestedObjectValueOf[dnsResourceModel] `tfsdk:"dns_resource"`
}

func (r *resourceConfigurationDefinitionModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.ResourceConfigurationDefinitionMemberIpResource:
		var model ipResourceModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		r.IpResource = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.ResourceConfigurationDefinitionMemberDnsResource:
		var model dnsResourceModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		r.DnsResource = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	case awstypes.ResourceConfigurationDefinitionMemberArnResource:
		var model arnResourceModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		r.ArnResource = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
		return diags

	default:
		return diags

	}
}

func (r resourceConfigurationDefinitionModel) Expand(ctx context.Context) (results any, diags diag.Diagnostics) {

	switch {
	case !r.IpResource.IsNull():
		ipAddressData, d := r.IpResource.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var rdc awstypes.ResourceConfigurationDefinitionMemberIpResource
		diags.Append(flex.Expand(ctx, ipAddressData, &rdc.Value)...)

		return &rdc, diags
	case !r.DnsResource.IsNull():
		DnsResourceData, d := r.DnsResource.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var rdc awstypes.ResourceConfigurationDefinitionMemberDnsResource
		diags.Append(flex.Expand(ctx, DnsResourceData, &rdc.Value)...)

		return &rdc, diags

	case !r.ArnResource.IsNull():
		ArnResourceData, d := r.ArnResource.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var rdc awstypes.ResourceConfigurationDefinitionMemberArnResource
		diags.Append(flex.Expand(ctx, ArnResourceData, &rdc.Value)...)

		return &rdc, diags

	}

	return nil, diags
}

var (
	_ flex.Expander  = resourceConfigurationDefinitionModel{}
	_ flex.Flattener = &resourceConfigurationDefinitionModel{}
)

type ipResourceModel struct {
	IpAddress types.String `tfsdk:"ip_address"`
}

type dnsResourceModel struct {
	DomainName    types.String                                                    `tfsdk:"domain_name"`
	IpAddressType fwtypes.StringEnum[awstypes.ResourceConfigurationIpAddressType] `tfsdk:"ip_address_type"`
}

type arnResourceModel struct {
	ARN fwtypes.ARN `tfsdk:"arn"`
}
