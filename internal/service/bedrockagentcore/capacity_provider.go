// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrockagentcore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_capacity_provider", name="Capacity Provider")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol;bedrockagentcorecontrol;bedrockagentcorecontrol.GetCapacityProviderOutput")
// @Testing(preCheck="testAccCapacityProviderPreCheck")
// @Testing(generator="testAccRandomAgentRuntimeName(t)")
// @Testing(hasNoPreExistingResource=true)
func newCapacityProviderResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &capacityProviderResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type capacityProviderResource struct {
	framework.ResourceWithModel[capacityProviderResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *capacityProviderResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					validResourceName,
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 4096),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"compute_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[capacityProviderComputeModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},

				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"ec2_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[capacityProviderEC2Model](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.IsRequired(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"lifecycle_configuration": framework.ResourceOptionalComputedSingleNestedObjectAttribute[capacityProviderLifecycleModel](ctx),
									"root_volume":             framework.ResourceOptionalComputedSingleNestedObjectAttribute[capacityProviderRootVolumeModel](ctx),
								},
								Blocks: map[string]schema.Block{
									"launch_template_source": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[capacityProviderLaunchTemplateModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
											listvalidator.IsRequired(),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"launch_parameters": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[capacityProviderLaunchParametersModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
														listvalidator.IsRequired(),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"operating_system": schema.StringAttribute{
																Required:   true,
																CustomType: fwtypes.StringEnumType[awstypes.OperatingSystem](),
															},
															"instance_profile_arn": schema.StringAttribute{
																Optional:   true,
																CustomType: fwtypes.ARNType,
															},
															"monitoring": schema.StringAttribute{
																Optional:      true,
																Computed:      true,
																CustomType:    fwtypes.StringEnumType[awstypes.Monitoring](),
																PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
															},
															"propagated_tags": schema.MapAttribute{
																Optional:   true,
																CustomType: fwtypes.MapOfStringType,
															},
															"ssh_key_name": schema.StringAttribute{
																Optional: true,
															},
														},
														Blocks: map[string]schema.Block{
															"instance_requirements": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[capacityProviderInstanceRequirementsModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																	listvalidator.IsRequired(),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"allowed_instance_types": schema.SetAttribute{
																			Required:   true,
																			CustomType: fwtypes.SetOfStringType,
																		},
																	},
																},
															},
															"capacity_reservation_specification": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[capacityProviderReservationModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"capacity_reservation_preference": schema.StringAttribute{
																			Optional:      true,
																			Computed:      true,
																			CustomType:    fwtypes.StringEnumType[awstypes.CapacityReservationPreference](),
																			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
																		},
																	},
																	Blocks: map[string]schema.Block{
																		"capacity_reservation_target": schema.ListNestedBlock{
																			CustomType: fwtypes.NewListNestedObjectTypeOf[capacityProviderReservationTargetModel](ctx),
																			Validators: []validator.List{
																				listvalidator.SizeAtMost(1),
																			},
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					"capacity_reservation_id": schema.StringAttribute{
																						Optional: true,
																					},
																					"capacity_reservation_resource_group_arn": schema.StringAttribute{
																						Optional:   true,
																						CustomType: fwtypes.ARNType,
																					},
																				},
																			},
																		},
																	},
																},
															},
															"ephemeral_volume": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[capacityProviderEphemeralVolumeModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtMost(5),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		names.AttrDeviceName: schema.StringAttribute{
																			Optional: true,
																		},
																		names.AttrVirtualName: schema.StringAttribute{
																			Optional: true,
																		},
																	},
																},
															},
															"license_specification": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[capacityProviderLicenseModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtMost(5),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"license_configuration_arn": schema.StringAttribute{
																			Required:   true,
																			CustomType: fwtypes.ARNType,
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
									names.AttrVPCConfiguration: schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[capacityProviderVPCModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
											listvalidator.IsRequired(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrSecurityGroups: schema.SetAttribute{
													Required:   true,
													CustomType: fwtypes.SetOfStringType,
												},
												names.AttrSubnets: schema.SetAttribute{
													Required:   true,
													CustomType: fwtypes.SetOfStringType,
												},
											},
										},
									},
									"volume": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[capacityProviderVolumeModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(5),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"ebs_configuration": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[capacityProviderEBSModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
														listvalidator.IsRequired(),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrName: schema.StringAttribute{
																Required: true,
															},
															"size_gib": schema.Int32Attribute{
																Required: true,
															},
															names.AttrEncrypted: schema.BoolAttribute{
																Optional:      true,
																Computed:      true,
																PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
															},
															names.AttrIOPS: schema.Int32Attribute{
																Optional:      true,
																Computed:      true,
																PlanModifiers: []planmodifier.Int32{int32planmodifier.UseStateForUnknown()},
															},
															names.AttrKMSKeyID: schema.StringAttribute{
																Optional:      true,
																Computed:      true,
																PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
															},
															names.AttrSnapshotID: schema.StringAttribute{
																Optional: true,
															},
															names.AttrThroughput: schema.Int32Attribute{
																Optional:      true,
																Computed:      true,
																PlanModifiers: []planmodifier.Int32{int32planmodifier.UseStateForUnknown()},
															},
															names.AttrVolumeType: schema.StringAttribute{
																Optional:      true,
																Computed:      true,
																CustomType:    fwtypes.StringEnumType[awstypes.EbsVolumeType](),
																PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
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
					},
				},
			},
			"permissions_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[capacityProviderPermissionsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"capacity_provider_operator_role_arn": schema.StringAttribute{
							Required:   true,
							CustomType: fwtypes.ARNType,
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *capacityProviderResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	var plan, state capacityProviderResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	// Compare after nested plan modifiers have preserved AWS defaults for unchanged configuration.
	if !plan.ComputeConfiguration.Equal(state.ComputeConfiguration) {
		resp.RequiresReplace = append(resp.RequiresReplace, path.Root("compute_configuration"))
	}
}

func (r *capacityProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan capacityProviderResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input bedrockagentcorecontrol.CreateCapacityProviderInput

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("CapacityProvider")))
	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)
	input.ClientToken = aws.String(create.UniqueId(ctx))
	var (
		out *bedrockagentcorecontrol.CreateCapacityProviderOutput
		err error
	)
	err = tfresource.Retry(ctx, propagationTimeout, func(ctx context.Context) *tfresource.RetryError {
		out, err = conn.CreateCapacityProvider(ctx, &input)
		// IAM propagation.
		if tfawserr.ErrMessageContains(err, errCodeValidationException, "Role validation failed") ||
			tfawserr.ErrMessageContains(err, errCodeValidationException, "The operator role does not have permission") {
			return tfresource.RetryableError(err)
		}
		if err != nil {
			return tfresource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan, flex.WithFieldNamePrefix("CapacityProvider")))
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve the identifier so Terraform can clean up a creation that fails while waiting.
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
	if resp.Diagnostics.HasError() {
		return
	}
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	created, err := waitCapacityProviderCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, created, &plan))
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *capacityProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state capacityProviderResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findCapacityProviderByID(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *capacityProviderResource) flatten(ctx context.Context, capacityProvider *bedrockagentcorecontrol.GetCapacityProviderOutput, data *capacityProviderResourceModel) (diags diag.Diagnostics) {
	diags.Append(flex.Flatten(ctx, capacityProvider, data, flex.WithFieldNamePrefix("CapacityProvider"))...)
	return diags
}

func (r *capacityProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan, state capacityProviderResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) {
		input := bedrockagentcorecontrol.UpdateCapacityProviderInput{
			CapacityProviderId: state.ID.ValueStringPointer(),
			Description:        &awstypes.UpdatedDescription{OptionalValue: plan.Description.ValueStringPointer()},
		}
		_, err := conn.UpdateCapacityProvider(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
			return
		}
		_, err = waitCapacityProviderUpdated(ctx, conn, state.ID.ValueString(), r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *capacityProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state capacityProviderResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := bedrockagentcorecontrol.DeleteCapacityProviderInput{
		CapacityProviderId: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteCapacityProvider(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitCapacityProviderDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
}

func waitCapacityProviderCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetCapacityProviderOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.CapacityProviderStatusCreating),
		Target:                    enum.Slice(awstypes.CapacityProviderStatusReady),
		Refresh:                   statusCapacityProvider(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetCapacityProviderOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.StatusReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitCapacityProviderUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetCapacityProviderOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.CapacityProviderStatusUpdating),
		Target:                    enum.Slice(awstypes.CapacityProviderStatusReady),
		Refresh:                   statusCapacityProvider(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetCapacityProviderOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.StatusReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitCapacityProviderDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetCapacityProviderOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CapacityProviderStatusDeleting, awstypes.CapacityProviderStatusReady),
		Target:  []string{},
		Refresh: statusCapacityProvider(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetCapacityProviderOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.StatusReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusCapacityProvider(conn *bedrockagentcorecontrol.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findCapacityProviderByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func findCapacityProviderByID(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string) (*bedrockagentcorecontrol.GetCapacityProviderOutput, error) {
	input := bedrockagentcorecontrol.GetCapacityProviderInput{
		CapacityProviderId: aws.String(id),
	}

	out, err := conn.GetCapacityProvider(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type capacityProviderResourceModel struct {
	framework.WithRegionModel
	ARN                      types.String                                                      `tfsdk:"arn"`
	ID                       types.String                                                      `tfsdk:"id"`
	Name                     types.String                                                      `tfsdk:"name"`
	Description              types.String                                                      `tfsdk:"description"`
	ComputeConfiguration     fwtypes.ListNestedObjectValueOf[capacityProviderComputeModel]     `tfsdk:"compute_configuration"`
	PermissionsConfiguration fwtypes.ListNestedObjectValueOf[capacityProviderPermissionsModel] `tfsdk:"permissions_configuration"`
	Tags                     tftags.Map                                                        `tfsdk:"tags"`
	TagsAll                  tftags.Map                                                        `tfsdk:"tags_all"`
	Timeouts                 timeouts.Value                                                    `tfsdk:"timeouts"`
}

func sweepCapacityProviders(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := bedrockagentcorecontrol.ListCapacityProvidersInput{}
	conn := client.BedrockAgentCoreClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := bedrockagentcorecontrol.NewListCapacityProvidersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.CapacityProviders {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newCapacityProviderResource, client,
				sweepfw.NewAttribute(names.AttrID, aws.ToString(v.CapacityProviderId))),
			)
		}
	}

	return sweepResources, nil
}

type capacityProviderComputeModel struct {
	EC2Configuration fwtypes.ListNestedObjectValueOf[capacityProviderEC2Model] `tfsdk:"ec2_configuration"`
}

type capacityProviderEC2Model struct {
	LaunchTemplateSource   fwtypes.ListNestedObjectValueOf[capacityProviderLaunchTemplateModel] `tfsdk:"launch_template_source"`
	VpcConfiguration       fwtypes.ListNestedObjectValueOf[capacityProviderVPCModel]            `tfsdk:"vpc_configuration"`
	LifecycleConfiguration fwtypes.ListNestedObjectValueOf[capacityProviderLifecycleModel]      `tfsdk:"lifecycle_configuration"`
	RootVolume             fwtypes.ListNestedObjectValueOf[capacityProviderRootVolumeModel]     `tfsdk:"root_volume"`
	Volumes                fwtypes.ListNestedObjectValueOf[capacityProviderVolumeModel]         `tfsdk:"volume"`
}

type capacityProviderLaunchTemplateModel struct {
	LaunchParameters fwtypes.ListNestedObjectValueOf[capacityProviderLaunchParametersModel] `tfsdk:"launch_parameters"`
}

type capacityProviderLaunchParametersModel struct {
	InstanceRequirements             fwtypes.ListNestedObjectValueOf[capacityProviderInstanceRequirementsModel] `tfsdk:"instance_requirements"`
	OperatingSystem                  fwtypes.StringEnum[awstypes.OperatingSystem]                               `tfsdk:"operating_system"`
	CapacityReservationSpecification fwtypes.ListNestedObjectValueOf[capacityProviderReservationModel]          `tfsdk:"capacity_reservation_specification"`
	EphemeralVolumes                 fwtypes.ListNestedObjectValueOf[capacityProviderEphemeralVolumeModel]      `tfsdk:"ephemeral_volume"`
	InstanceProfileARN               fwtypes.ARN                                                                `tfsdk:"instance_profile_arn"`
	LicenseSpecifications            fwtypes.ListNestedObjectValueOf[capacityProviderLicenseModel]              `tfsdk:"license_specification"`
	Monitoring                       fwtypes.StringEnum[awstypes.Monitoring]                                    `tfsdk:"monitoring"`
	PropagatedTags                   fwtypes.MapOfString                                                        `tfsdk:"propagated_tags"`
	SSHKeyName                       types.String                                                               `tfsdk:"ssh_key_name"`
}

type capacityProviderInstanceRequirementsModel struct {
	AllowedInstanceTypes fwtypes.SetOfString `tfsdk:"allowed_instance_types"`
}

type capacityProviderVPCModel struct {
	SecurityGroups fwtypes.SetOfString `tfsdk:"security_groups"`
	Subnets        fwtypes.SetOfString `tfsdk:"subnets"`
}

type capacityProviderPermissionsModel struct {
	CapacityProviderOperatorRoleARN fwtypes.ARN `tfsdk:"capacity_provider_operator_role_arn"`
}

type capacityProviderLifecycleModel struct {
	IdleInstanceTimeout types.Int32 `tfsdk:"idle_instance_timeout"`
	MaxLifetime         types.Int32 `tfsdk:"max_lifetime"`
}

type capacityProviderRootVolumeModel struct {
	Encrypted    types.Bool                                 `tfsdk:"encrypted"`
	FreeSpaceGiB types.Int32                                `tfsdk:"free_space_gib"`
	IOPS         types.Int32                                `tfsdk:"iops"`
	KMSKeyID     types.String                               `tfsdk:"kms_key_id"`
	Throughput   types.Int32                                `tfsdk:"throughput"`
	VolumeType   fwtypes.StringEnum[awstypes.EbsVolumeType] `tfsdk:"volume_type"`
}

type capacityProviderVolumeModel struct {
	EBSConfiguration fwtypes.ListNestedObjectValueOf[capacityProviderEBSModel] `tfsdk:"ebs_configuration"`
}

type capacityProviderEBSModel struct {
	Name       types.String                               `tfsdk:"name"`
	SizeGiB    types.Int32                                `tfsdk:"size_gib"`
	Encrypted  types.Bool                                 `tfsdk:"encrypted"`
	IOPS       types.Int32                                `tfsdk:"iops"`
	KMSKeyID   types.String                               `tfsdk:"kms_key_id"`
	SnapshotID types.String                               `tfsdk:"snapshot_id"`
	Throughput types.Int32                                `tfsdk:"throughput"`
	VolumeType fwtypes.StringEnum[awstypes.EbsVolumeType] `tfsdk:"volume_type"`
}

type capacityProviderReservationModel struct {
	CapacityReservationPreference fwtypes.StringEnum[awstypes.CapacityReservationPreference]              `tfsdk:"capacity_reservation_preference"`
	CapacityReservationTarget     fwtypes.ListNestedObjectValueOf[capacityProviderReservationTargetModel] `tfsdk:"capacity_reservation_target"`
}

type capacityProviderReservationTargetModel struct {
	CapacityReservationID               types.String `tfsdk:"capacity_reservation_id"`
	CapacityReservationResourceGroupARN fwtypes.ARN  `tfsdk:"capacity_reservation_resource_group_arn"`
}

type capacityProviderEphemeralVolumeModel struct {
	DeviceName  types.String `tfsdk:"device_name"`
	VirtualName types.String `tfsdk:"virtual_name"`
}

type capacityProviderLicenseModel struct {
	LicenseConfigurationARN fwtypes.ARN `tfsdk:"license_configuration_arn"`
}

var (
	_ flex.Expander  = capacityProviderComputeModel{}
	_ flex.Flattener = &capacityProviderComputeModel{}
)

func (m capacityProviderComputeModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	v, d := m.EC2Configuration.ToPtr(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return nil, diags
	}
	var out awstypes.ComputeConfigurationMemberEc2Configuration
	smerr.AddEnrich(ctx, &diags, flex.Expand(ctx, v, &out.Value))
	return &out, diags
}
func (m *capacityProviderComputeModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch v := v.(type) {
	case awstypes.ComputeConfigurationMemberEc2Configuration:
		var data capacityProviderEC2Model
		smerr.AddEnrich(ctx, &diags, flex.Flatten(ctx, v.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.EC2Configuration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("capacity provider configuration flatten: %T", v))
	}
	return diags
}

var (
	_ flex.Expander  = capacityProviderLaunchTemplateModel{}
	_ flex.Flattener = &capacityProviderLaunchTemplateModel{}
)

func (m capacityProviderLaunchTemplateModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	v, d := m.LaunchParameters.ToPtr(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return nil, diags
	}
	var out awstypes.LaunchTemplateSourceMemberLaunchParameters
	smerr.AddEnrich(ctx, &diags, flex.Expand(ctx, v, &out.Value))
	return &out, diags
}
func (m *capacityProviderLaunchTemplateModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch v := v.(type) {
	case awstypes.LaunchTemplateSourceMemberLaunchParameters:
		var data capacityProviderLaunchParametersModel
		smerr.AddEnrich(ctx, &diags, flex.Flatten(ctx, v.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.LaunchParameters = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("capacity provider configuration flatten: %T", v))
	}
	return diags
}

var (
	_ flex.Expander  = capacityProviderVolumeModel{}
	_ flex.Flattener = &capacityProviderVolumeModel{}
)

func (m capacityProviderVolumeModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	v, d := m.EBSConfiguration.ToPtr(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return nil, diags
	}
	var out awstypes.VolumeConfigurationMemberEbsConfiguration
	smerr.AddEnrich(ctx, &diags, flex.Expand(ctx, v, &out.Value))
	return &out, diags
}
func (m *capacityProviderVolumeModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch v := v.(type) {
	case awstypes.VolumeConfigurationMemberEbsConfiguration:
		var data capacityProviderEBSModel
		smerr.AddEnrich(ctx, &diags, flex.Flatten(ctx, v.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.EBSConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("capacity provider configuration flatten: %T", v))
	}
	return diags
}
