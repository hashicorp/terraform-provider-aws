// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"
	awstypes "github.com/aws/aws-sdk-go-v2/service/imagebuilder/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Lifecycle Policy")
// @Tags(identifierAttribute="id")
func newResourceLifecyclePolicy(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceLifecyclePolicy{}, nil
}

const (
	ResNameLifecyclePolicy = "Lifecycle Policy"
)

type resourceLifecyclePolicy struct {
	framework.ResourceWithConfigure
}

func (r *resourceLifecyclePolicy) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_imagebuilder_lifecycle_policy"
}

func (r *resourceLifecyclePolicy) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"description": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"execution_role": schema.StringAttribute{
				Required: true,
			},
			"id": framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.LifecyclePolicyResourceType](),
				},
			},
			"status": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.LifecyclePolicyStatus](),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"policy_details": schema.SetNestedBlock{
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.SizeAtMost(3),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"action": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.IsRequired(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											enum.FrameworkValidate[awstypes.LifecyclePolicyDetailActionType](),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"include_resources": schema.ListNestedBlock{
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"amis": schema.BoolAttribute{
													Optional: true,
													Computed: true,
													PlanModifiers: []planmodifier.Bool{
														boolplanmodifier.UseStateForUnknown(),
													},
												},
												"containers": schema.BoolAttribute{
													Optional: true,
													Computed: true,
													PlanModifiers: []planmodifier.Bool{
														boolplanmodifier.UseStateForUnknown(),
													},
												},
												"snapshots": schema.BoolAttribute{
													Optional: true,
													Computed: true,
													PlanModifiers: []planmodifier.Bool{
														boolplanmodifier.UseStateForUnknown(),
													},
												},
											},
										},
									},
								},
							},
						},
						"filter": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.IsRequired(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											enum.FrameworkValidate[awstypes.LifecyclePolicyDetailFilterType](),
										},
									},
									"value": schema.Int64Attribute{
										Required: true,
									},
									"retain_at_least": schema.Int64Attribute{
										Optional: true,
									},
									"unit": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											enum.FrameworkValidate[awstypes.LifecyclePolicyTimeUnit](),
										},
									},
								},
							},
						},
						"exclusion_rules": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"tag_map": schema.MapAttribute{
										ElementType: types.StringType,
										Optional:    true,
									},
								},
								Blocks: map[string]schema.Block{
									"amis": schema.ListNestedBlock{
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"is_public": schema.BoolAttribute{
													Optional: true,
												},
												"regions": schema.ListAttribute{
													ElementType: types.StringType,
													Optional:    true,
												},
												"shared_accounts": schema.ListAttribute{
													ElementType: types.StringType,
													Optional:    true,
												},
												"tag_map": schema.MapAttribute{
													ElementType: types.StringType,
													Optional:    true,
												},
											},
											Blocks: map[string]schema.Block{
												"last_launched": schema.ListNestedBlock{
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"unit": schema.StringAttribute{
																Required: true,
																Validators: []validator.String{
																	enum.FrameworkValidate[awstypes.LifecyclePolicyTimeUnit](),
																},
															},
															"value": schema.Int64Attribute{
																Required: true,
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
			"resource_selection": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"tag_map": schema.MapAttribute{
							ElementType: types.StringType,
							Optional:    true,
						},
					},
					Blocks: map[string]schema.Block{
						"recipes": schema.SetNestedBlock{
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
								setvalidator.SizeAtMost(50),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Required: true,
									},
									"semantic_version": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *resourceLifecyclePolicy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceLifecyclePolicyData

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ImageBuilderClient(ctx)

	in := &imagebuilder.CreateLifecyclePolicyInput{
		ClientToken:   aws.String(id.UniqueId()),
		ExecutionRole: aws.String(plan.ExecutionRole.ValueString()),
		Name:          aws.String(plan.Name.ValueString()),
		ResourceType:  awstypes.LifecyclePolicyResourceType(plan.ResourceType.ValueString()),
		Tags:          getTagsIn(ctx),
	}

	if !plan.Description.IsNull() {
		in.Description = aws.String(plan.Description.ValueString())
	}

	if !plan.PolicyDetails.IsNull() {
		var tfList []resourcePolicyDetailsData
		resp.Diagnostics.Append(plan.PolicyDetails.ElementsAs(ctx, &tfList, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		policyDetails, d := expandPolicyDetails(ctx, tfList)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		in.PolicyDetails = policyDetails
	}

	if !plan.ResourceSelection.IsNull() {
		var tfList []resourceResourceSelectionData
		resp.Diagnostics.Append(plan.ResourceSelection.ElementsAs(ctx, &tfList, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		resourceSelection, d := expandResourceSelection(ctx, tfList)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		in.ResourceSelection = resourceSelection
	}

	if !plan.Status.IsNull() {
		in.Status = awstypes.LifecyclePolicyStatus(plan.Status.ValueString())
	}

	// Include retry handling to allow for IAM propagation
	var out *imagebuilder.CreateLifecyclePolicyOutput
	err := tfresource.Retry(ctx, propagationTimeout, func() *retry.RetryError {
		var err error
		out, err = conn.CreateLifecyclePolicy(ctx, in)
		if err != nil {
			if errs.MessageContains(err, InvalidParameterValueException, "The provided role does not exist or does not have sufficient permissions") {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ImageBuilder, create.ErrActionCreating, ResNameLifecyclePolicy, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ImageBuilder, create.ErrActionCreating, ResNameLifecyclePolicy, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ID = flex.StringToFramework(ctx, out.LifecyclePolicyArn)
	plan.ARN = flex.StringToFramework(ctx, out.LifecyclePolicyArn)

	// Read to retrieve computed arguments not part of the create response
	readOut, err := findLifecyclePolicyByARN(ctx, conn, plan.ARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ImageBuilder, create.ErrActionCreating, ResNameLifecyclePolicy, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	plan.Status = flex.StringValueToFramework(ctx, readOut.Status)
	plan.ResourceType = flex.StringValueToFramework(ctx, readOut.ResourceType)

	policyDetails, d := flattenPolicyDetails(ctx, readOut.PolicyDetails)
	resp.Diagnostics.Append(d...)
	plan.PolicyDetails = policyDetails

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceLifecyclePolicy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ImageBuilderClient(ctx)

	var state resourceLifecyclePolicyData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findLifecyclePolicyByARN(ctx, conn, state.ID.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ImageBuilder, create.ErrActionReading, ResNameLifecyclePolicy, state.Name.String(), nil),
			err.Error(),
		)
		return
	}

	state.ARN = flex.StringToFramework(ctx, out.Arn)
	state.ID = flex.StringToFramework(ctx, out.Arn)
	state.Description = flex.StringToFramework(ctx, out.Description)
	state.ExecutionRole = flex.StringToFramework(ctx, out.ExecutionRole)
	state.Name = flex.StringToFramework(ctx, out.Name)

	policyDetails, d := flattenPolicyDetails(ctx, out.PolicyDetails)
	resp.Diagnostics.Append(d...)
	state.PolicyDetails = policyDetails

	resourceSelection, d := flattenResourceSelection(ctx, out.ResourceSelection)
	resp.Diagnostics.Append(d...)
	state.ResourceSelection = resourceSelection

	state.ResourceType = flex.StringValueToFramework(ctx, out.ResourceType)
	state.Status = flex.StringValueToFramework(ctx, out.Status)

	setTagsOut(ctx, out.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceLifecyclePolicy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().ImageBuilderClient(ctx)

	var plan, state resourceLifecyclePolicyData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) ||
		!plan.ExecutionRole.Equal(state.ExecutionRole) ||
		!plan.PolicyDetails.Equal(state.PolicyDetails) ||
		!plan.ResourceSelection.Equal(state.ResourceSelection) ||
		!plan.ResourceType.Equal(state.ResourceType) ||
		!plan.Status.Equal(state.Status) {
		in := &imagebuilder.UpdateLifecyclePolicyInput{
			LifecyclePolicyArn: aws.String(plan.ID.ValueString()),
			ExecutionRole:      aws.String(plan.ExecutionRole.ValueString()),
			ResourceType:       awstypes.LifecyclePolicyResourceType(plan.ResourceType.ValueString()),
			Status:             awstypes.LifecyclePolicyStatus(plan.Status.ValueString()),
		}

		if !plan.Description.IsNull() {
			in.Description = aws.String(plan.Description.ValueString())
		}

		if !plan.PolicyDetails.IsNull() {
			var tfList []resourcePolicyDetailsData
			resp.Diagnostics.Append(plan.PolicyDetails.ElementsAs(ctx, &tfList, false)...)
			if resp.Diagnostics.HasError() {
				return
			}

			policyDetails, d := expandPolicyDetails(ctx, tfList)
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}
			in.PolicyDetails = policyDetails
		}

		if !plan.ResourceSelection.IsNull() {
			var tfList []resourceResourceSelectionData
			resp.Diagnostics.Append(plan.ResourceSelection.ElementsAs(ctx, &tfList, false)...)
			if resp.Diagnostics.HasError() {
				return
			}

			resourceSelection, d := expandResourceSelection(ctx, tfList)
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}
			in.ResourceSelection = resourceSelection
		}

		// Include retry handling to allow for IAM propagation
		var out *imagebuilder.UpdateLifecyclePolicyOutput
		err := tfresource.Retry(ctx, propagationTimeout, func() *retry.RetryError {
			var err error
			out, err = conn.UpdateLifecyclePolicy(ctx, in)
			if err != nil {
				if errs.MessageContains(err, InvalidParameterValueException, "The provided role does not exist or does not have sufficient permissions") {
					return retry.RetryableError(err)
				}
				return retry.NonRetryableError(err)
			}
			return nil
		})

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ImageBuilder, create.ErrActionUpdating, ResNameLifecyclePolicy, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ImageBuilder, create.ErrActionUpdating, ResNameLifecyclePolicy, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceLifecyclePolicy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ImageBuilderClient(ctx)

	var state resourceLifecyclePolicyData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteLifecyclePolicy(ctx, &imagebuilder.DeleteLifecyclePolicyInput{
		LifecyclePolicyArn: aws.String(state.ID.ValueString()),
	})

	if err != nil {
		if errs.MessageContains(err, ResourceNotFoundException, "cannot be found") {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ImageBuilder, create.ErrActionDeleting, ResNameLifecyclePolicy, state.ID.String(), nil),
			err.Error(),
		)
		return
	}
}

func (r *resourceLifecyclePolicy) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *resourceLifecyclePolicy) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

func findLifecyclePolicyByARN(ctx context.Context, conn *imagebuilder.Client, arn string) (*awstypes.LifecyclePolicy, error) {
	in := &imagebuilder.GetLifecyclePolicyInput{
		LifecyclePolicyArn: aws.String(arn),
	}

	out, err := conn.GetLifecyclePolicy(ctx, in)
	if err != nil {
		if errs.MessageContains(err, ResourceNotFoundException, "cannot be found") {
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

	return out.LifecyclePolicy, nil
}

func expandPolicyDetails(ctx context.Context, tfList []resourcePolicyDetailsData) ([]awstypes.LifecyclePolicyDetail, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil, diags
	}

	apiResult := []awstypes.LifecyclePolicyDetail{}

	for _, policyDetail := range tfList {
		apiObject := awstypes.LifecyclePolicyDetail{}

		if !policyDetail.Action.IsNull() {
			var tfList []resourceActionData
			diags.Append(policyDetail.Action.ElementsAs(ctx, &tfList, false)...)
			if diags.HasError() {
				return nil, diags
			}

			action, d := expandPolicyDetailAction(ctx, tfList)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
			apiObject.Action = action
		}

		if !policyDetail.Filter.IsNull() {
			var tfList []resourceFilterData
			diags.Append(policyDetail.Filter.ElementsAs(ctx, &tfList, false)...)
			if diags.HasError() {
				return nil, diags
			}

			apiObject.Filter = expandPolicyDetailFilter(tfList)
		}

		if !policyDetail.ExclusionRules.IsNull() {
			var tfList []resourceExclusionRulesData
			diags.Append(policyDetail.ExclusionRules.ElementsAs(ctx, &tfList, false)...)
			if diags.HasError() {
				return nil, diags
			}

			exclusionRules, d := expandPolicyDetailExclusionRules(ctx, tfList)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
			apiObject.ExclusionRules = exclusionRules
		}

		apiResult = append(apiResult, apiObject)
	}

	return apiResult, diags
}

func expandPolicyDetailAction(ctx context.Context, tfList []resourceActionData) (*awstypes.LifecyclePolicyDetailAction, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil, diags
	}

	tfObj := tfList[0]

	apiObject := awstypes.LifecyclePolicyDetailAction{
		Type: awstypes.LifecyclePolicyDetailActionType(tfObj.Type.ValueString()),
	}

	if !tfObj.IncludeResources.IsNull() {
		var tfList []resourceIncludeResourcesData
		diags.Append(tfObj.IncludeResources.ElementsAs(ctx, &tfList, false)...)
		if diags.HasError() {
			return nil, diags
		}

		apiObject.IncludeResources = expandPolicyDetailActionIncludeResources(tfList)
	}

	return &apiObject, diags
}

func expandPolicyDetailActionIncludeResources(tfList []resourceIncludeResourcesData) *awstypes.LifecyclePolicyDetailActionIncludeResources {
	tfObj := tfList[0]

	apiObject := awstypes.LifecyclePolicyDetailActionIncludeResources{
		Amis:       tfObj.Amis.ValueBool(),
		Containers: tfObj.Containers.ValueBool(),
		Snapshots:  tfObj.Snapshots.ValueBool(),
	}

	return &apiObject
}

func expandPolicyDetailFilter(tfList []resourceFilterData) *awstypes.LifecyclePolicyDetailFilter {
	if len(tfList) == 0 {
		return nil
	}

	tfObj := tfList[0]

	apiObject := awstypes.LifecyclePolicyDetailFilter{}

	if !tfObj.Type.IsNull() {
		apiObject.Type = awstypes.LifecyclePolicyDetailFilterType(tfObj.Type.ValueString())
	}

	if !tfObj.Value.IsNull() {
		apiObject.Value = aws.Int32(int32(tfObj.Value.ValueInt64()))
	}

	if !tfObj.RetainAtLeast.IsNull() {
		apiObject.RetainAtLeast = aws.Int32(int32(tfObj.RetainAtLeast.ValueInt64()))
	}

	if !tfObj.Unit.IsNull() {
		apiObject.Unit = awstypes.LifecyclePolicyTimeUnit(tfObj.Unit.ValueString())
	}

	return &apiObject
}

func expandPolicyDetailExclusionRules(ctx context.Context, tfList []resourceExclusionRulesData) (*awstypes.LifecyclePolicyDetailExclusionRules, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil, diags
	}

	tfObj := tfList[0]

	apiObject := awstypes.LifecyclePolicyDetailExclusionRules{}

	if !tfObj.AMIs.IsNull() {
		var tfList []resourceAMISData
		diags.Append(tfObj.AMIs.ElementsAs(ctx, &tfList, false)...)
		if diags.HasError() {
			return nil, diags
		}

		Amis, d := expandPolicyDetailExclusionRulesAMIS(ctx, tfList)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		apiObject.Amis = Amis
	}

	if !tfObj.TagMap.IsNull() {
		apiObject.TagMap = flex.ExpandFrameworkStringValueMap(ctx, tfObj.TagMap)
	}

	return &apiObject, diags
}

func expandPolicyDetailExclusionRulesAMIS(ctx context.Context, tfList []resourceAMISData) (*awstypes.LifecyclePolicyDetailExclusionRulesAmis, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil, diags
	}

	tfObj := tfList[0]

	apiObject := awstypes.LifecyclePolicyDetailExclusionRulesAmis{}

	if !tfObj.IsPublic.IsNull() {
		apiObject.IsPublic = aws.ToBool(tfObj.IsPublic.ValueBoolPointer())
	}

	if !tfObj.LastLaunched.IsNull() {
		var tfList []resourceLastLaunchedData
		diags.Append(tfObj.LastLaunched.ElementsAs(ctx, &tfList, false)...)
		if diags.HasError() {
			return nil, diags
		}

		apiObject.LastLaunched = expandPolicyDetailExclusionRulesAMISLastLaunched(tfList)
	}

	if !tfObj.Regions.IsNull() {
		apiObject.Regions = flex.ExpandFrameworkStringValueList(ctx, tfObj.Regions)
	}

	if !tfObj.SharedAccounts.IsNull() {
		apiObject.SharedAccounts = flex.ExpandFrameworkStringValueList(ctx, tfObj.SharedAccounts)
	}

	if !tfObj.TagMap.IsNull() {
		apiObject.TagMap = flex.ExpandFrameworkStringValueMap(ctx, tfObj.TagMap)
	}

	return &apiObject, diags
}

func expandPolicyDetailExclusionRulesAMISLastLaunched(tfList []resourceLastLaunchedData) *awstypes.LifecyclePolicyDetailExclusionRulesAmisLastLaunched {
	tfObj := tfList[0]

	apiObject := awstypes.LifecyclePolicyDetailExclusionRulesAmisLastLaunched{}

	if !tfObj.Unit.IsNull() {
		apiObject.Unit = awstypes.LifecyclePolicyTimeUnit(tfObj.Unit.ValueString())
	}

	if !tfObj.Value.IsNull() {
		apiObject.Value = aws.Int32(int32(tfObj.Value.ValueInt64()))
	}

	return &apiObject
}

func expandResourceSelection(ctx context.Context, tfList []resourceResourceSelectionData) (*awstypes.LifecyclePolicyResourceSelection, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil, diags
	}

	tfObj := tfList[0]

	apiObject := awstypes.LifecyclePolicyResourceSelection{}

	if !tfObj.Recipes.IsNull() {
		var tfList []resourceRecipesData
		diags.Append(tfObj.Recipes.ElementsAs(ctx, &tfList, false)...)
		if diags.HasError() {
			return nil, diags
		}

		apiObject.Recipes = expandResourceSelectionRecipes(tfList)
	}

	if !tfObj.TagMap.IsNull() {
		apiObject.TagMap = flex.ExpandFrameworkStringValueMap(ctx, tfObj.TagMap)
	}

	return &apiObject, diags
}

func expandResourceSelectionRecipes(tfList []resourceRecipesData) []awstypes.LifecyclePolicyResourceSelectionRecipe {
	apiResult := []awstypes.LifecyclePolicyResourceSelectionRecipe{}

	for _, tfObj := range tfList {
		apiObject := awstypes.LifecyclePolicyResourceSelectionRecipe{}

		if !tfObj.Name.IsNull() {
			apiObject.Name = aws.String(tfObj.Name.ValueString())
		}
		if !tfObj.SemanticVersion.IsNull() {
			apiObject.SemanticVersion = aws.String(tfObj.SemanticVersion.ValueString())
		}

		apiResult = append(apiResult, apiObject)
	}
	return apiResult
}

func flattenPolicyDetails(ctx context.Context, apiObject []awstypes.LifecyclePolicyDetail) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: resourcePolicyDetailAttrTypes}

	if apiObject == nil {
		return types.SetNull(elemType), diags
	}

	result := []attr.Value{}

	for _, policyDetail := range apiObject {
		action, d := flattenDetailAction(ctx, policyDetail.Action)
		diags.Append(d...)

		filter, d := flattenDetailFilter(ctx, policyDetail.Filter)
		diags.Append(d...)

		exclusionRules, d := flattenDetailExclusionRules(ctx, policyDetail.ExclusionRules)
		diags.Append(d...)

		obj := map[string]attr.Value{
			"action":          action,
			"filter":          filter,
			"exclusion_rules": exclusionRules,
		}

		objVal, d := types.ObjectValue(resourcePolicyDetailAttrTypes, obj)
		diags.Append(d...)

		result = append(result, objVal)
	}

	setVal, d := types.SetValue(elemType, result)
	diags.Append(d...)

	return setVal, diags
}

func flattenDetailAction(ctx context.Context, apiObject *awstypes.LifecyclePolicyDetailAction) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: resourceActionAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	includeResources, d := flattenIncludeResources(ctx, apiObject.IncludeResources)
	diags.Append(d...)

	obj := map[string]attr.Value{
		"include_resources": includeResources,
		"type":              flex.StringValueToFramework(ctx, apiObject.Type),
	}

	objVal, d := types.ObjectValue(resourceActionAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenIncludeResources(ctx context.Context, apiObject *awstypes.LifecyclePolicyDetailActionIncludeResources) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: resourceIncludeResourcesAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"amis":       flex.BoolToFramework(ctx, aws.Bool(apiObject.Amis)),
		"containers": flex.BoolToFramework(ctx, aws.Bool(apiObject.Containers)),
		"snapshots":  flex.BoolToFramework(ctx, aws.Bool(apiObject.Snapshots)),
	}

	objVal, d := types.ObjectValue(resourceIncludeResourcesAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenDetailFilter(ctx context.Context, apiObject *awstypes.LifecyclePolicyDetailFilter) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: resourceFilterAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"type":            flex.StringValueToFramework(ctx, apiObject.Type),
		"value":           flex.Int32ToFramework(ctx, apiObject.Value),
		"retain_at_least": flex.Int32ToFramework(ctx, apiObject.RetainAtLeast),
		"unit":            flex.StringValueToFramework(ctx, apiObject.Unit),
	}

	objVal, d := types.ObjectValue(resourceFilterAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenDetailExclusionRules(ctx context.Context, apiObject *awstypes.LifecyclePolicyDetailExclusionRules) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: resourceExclusionRulesAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	amis, d := flattenExclusionRulesAMIS(ctx, apiObject.Amis)
	diags.Append(d...)

	obj := map[string]attr.Value{
		"amis":    amis,
		"tag_map": flex.FlattenFrameworkStringValueMap(ctx, apiObject.TagMap),
	}

	objVal, d := types.ObjectValue(resourceExclusionRulesAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenExclusionRulesAMIS(ctx context.Context, apiObject *awstypes.LifecyclePolicyDetailExclusionRulesAmis) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: resourceAMISAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	lastLaunched, d := flattenExclusionRulesAMISLastLaunched(ctx, apiObject.LastLaunched)
	diags.Append(d...)

	obj := map[string]attr.Value{
		"is_public":       flex.BoolToFramework(ctx, aws.Bool(apiObject.IsPublic)),
		"regions":         flex.FlattenFrameworkStringValueList(ctx, apiObject.Regions),
		"shared_accounts": flex.FlattenFrameworkStringValueList(ctx, apiObject.SharedAccounts),
		"tag_map":         flex.FlattenFrameworkStringValueMap(ctx, apiObject.TagMap),
		"last_launched":   lastLaunched,
	}

	objVal, d := types.ObjectValue(resourceAMISAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenExclusionRulesAMISLastLaunched(ctx context.Context, apiObject *awstypes.LifecyclePolicyDetailExclusionRulesAmisLastLaunched) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: resourceLastLaunchedAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"unit":  flex.StringValueToFramework(ctx, apiObject.Unit),
		"value": flex.Int32ToFramework(ctx, apiObject.Value),
	}

	objVal, d := types.ObjectValue(resourceLastLaunchedAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenResourceSelection(ctx context.Context, apiObject *awstypes.LifecyclePolicyResourceSelection) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: resourceResourceSelectionAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	recipes, d := flattenResourceSelectionRecipes(ctx, apiObject.Recipes)
	diags.Append(d...)

	obj := map[string]attr.Value{
		"recipes": recipes,
		"tag_map": flex.FlattenFrameworkStringValueMap(ctx, apiObject.TagMap),
	}

	objVal, d := types.ObjectValue(resourceResourceSelectionAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenResourceSelectionRecipes(ctx context.Context, apiObject []awstypes.LifecyclePolicyResourceSelectionRecipe) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: resourceRecipeAttrTypes}

	if apiObject == nil {
		return types.SetNull(elemType), diags
	}

	result := []attr.Value{}

	for _, recipe := range apiObject {
		obj := map[string]attr.Value{
			"name":             flex.StringToFramework(ctx, recipe.Name),
			"semantic_version": flex.StringToFramework(ctx, recipe.SemanticVersion),
		}

		objVal, d := types.ObjectValue(resourceRecipeAttrTypes, obj)
		diags.Append(d...)

		result = append(result, objVal)
	}

	setVal, d := types.SetValue(elemType, result)
	diags.Append(d...)

	return setVal, diags
}

type resourceLifecyclePolicyData struct {
	ID                types.String `tfsdk:"id"`
	ARN               types.String `tfsdk:"arn"`
	Description       types.String `tfsdk:"description"`
	Name              types.String `tfsdk:"name"`
	ExecutionRole     types.String `tfsdk:"execution_role"`
	ResourceType      types.String `tfsdk:"resource_type"`
	Status            types.String `tfsdk:"status"`
	PolicyDetails     types.Set    `tfsdk:"policy_details"`
	ResourceSelection types.List   `tfsdk:"resource_selection"`
	Tags              types.Map    `tfsdk:"tags"`
	TagsAll           types.Map    `tfsdk:"tags_all"`
}

type resourcePolicyDetailsData struct {
	Action         types.List `tfsdk:"action"`
	Filter         types.List `tfsdk:"filter"`
	ExclusionRules types.List `tfsdk:"exclusion_rules"`
}

type resourceResourceSelectionData struct {
	TagMap  types.Map `tfsdk:"tag_map"`
	Recipes types.Set `tfsdk:"recipes"`
}

type resourceRecipesData struct {
	Name            types.String `tfsdk:"name"`
	SemanticVersion types.String `tfsdk:"semantic_version"`
}

type resourceActionData struct {
	Type             types.String `tfsdk:"type"`
	IncludeResources types.List   `tfsdk:"include_resources"`
}

type resourceIncludeResourcesData struct {
	Amis       types.Bool `tfsdk:"amis"`
	Containers types.Bool `tfsdk:"containers"`
	Snapshots  types.Bool `tfsdk:"snapshots"`
}

type resourceFilterData struct {
	Type          types.String `tfsdk:"type"`
	Value         types.Int64  `tfsdk:"value"`
	RetainAtLeast types.Int64  `tfsdk:"retain_at_least"`
	Unit          types.String `tfsdk:"unit"`
}

type resourceExclusionRulesData struct {
	AMIs   types.List `tfsdk:"amis"`
	TagMap types.Map  `tfsdk:"tag_map"`
}

type resourceAMISData struct {
	IsPublic       types.Bool `tfsdk:"is_public"`
	LastLaunched   types.List `tfsdk:"last_launched"`
	Regions        types.List `tfsdk:"regions"`
	SharedAccounts types.List `tfsdk:"shared_accounts"`
	TagMap         types.Map  `tfsdk:"tag_map"`
}

type resourceLastLaunchedData struct {
	Unit  types.String `tfsdk:"unit"`
	Value types.Int64  `tfsdk:"value"`
}

var resourcePolicyDetailAttrTypes = map[string]attr.Type{
	"action":          types.ListType{ElemType: types.ObjectType{AttrTypes: resourceActionAttrTypes}},
	"filter":          types.ListType{ElemType: types.ObjectType{AttrTypes: resourceFilterAttrTypes}},
	"exclusion_rules": types.ListType{ElemType: types.ObjectType{AttrTypes: resourceExclusionRulesAttrTypes}},
}

var resourceActionAttrTypes = map[string]attr.Type{
	"type":              types.StringType,
	"include_resources": types.ListType{ElemType: types.ObjectType{AttrTypes: resourceIncludeResourcesAttrTypes}},
}

var resourceIncludeResourcesAttrTypes = map[string]attr.Type{
	"amis":       types.BoolType,
	"containers": types.BoolType,
	"snapshots":  types.BoolType,
}

var resourceFilterAttrTypes = map[string]attr.Type{
	"type":            types.StringType,
	"value":           types.Int64Type,
	"retain_at_least": types.Int64Type,
	"unit":            types.StringType,
}

var resourceExclusionRulesAttrTypes = map[string]attr.Type{
	"amis":    types.ListType{ElemType: types.ObjectType{AttrTypes: resourceAMISAttrTypes}},
	"tag_map": types.MapType{ElemType: types.StringType},
}

var resourceAMISAttrTypes = map[string]attr.Type{
	"is_public":       types.BoolType,
	"last_launched":   types.ListType{ElemType: types.ObjectType{AttrTypes: resourceLastLaunchedAttrTypes}},
	"regions":         types.ListType{ElemType: types.StringType},
	"shared_accounts": types.ListType{ElemType: types.StringType},
	"tag_map":         types.MapType{ElemType: types.StringType},
}

var resourceLastLaunchedAttrTypes = map[string]attr.Type{
	"unit":  types.StringType,
	"value": types.Int64Type,
}

var resourceResourceSelectionAttrTypes = map[string]attr.Type{
	"recipes": types.SetType{ElemType: types.ObjectType{AttrTypes: resourceRecipeAttrTypes}},
	"tag_map": types.MapType{ElemType: types.StringType},
}

var resourceRecipeAttrTypes = map[string]attr.Type{
	"name":             types.StringType,
	"semantic_version": types.StringType,
}
