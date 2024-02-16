// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/imagebuilder/types"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/aws/aws-sdk-go/service/opensearchserverless"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource
func newResourceLifecyclePolicy(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceLifecyclePolicy{}, nil
}

type resourceLifecyclePolicyData struct {
	ID                types.String                  `tfsdk:"id"`
	Arn               types.String                  `tfsdk:"arn"`
	Description       types.String                  `tfsdk:"description"`
	Name              types.String                  `tfsdk:"name"`
	ExecutionRole     types.String                  `tfsdk:"execution_role"`
	ResourceType      types.String                  `tfsdk:"resource_type"`
	Status            types.String                  `tfsdk:"status"`
	PolicyDetails     resourcePolicyDetailsData     `tfsdk:"policy_details"`
	ResourceSelection resourceResourceSelectionData `tfsdk:"resource_selection"`
}

type resourcePolicyDetailsData struct {
	Action         resourceActionData         `tfsdk:"action"`
	Filter         resourceFilterData         `tfsdk:"filter"`
	ExclusionRules resourceExclusionRulesData `tfsdk:"exclusion_rules"`
}

type resourceResourceSelectionData struct {
	TagMap  types.String        `tfsdk:"tag_map"`
	Recipes resourceRecipesData `tfsdk:"recipes"`
}

type resourceRecipesData struct {
	Name            types.String `tfsdk:"name"`
	SemanticVersion types.String `tfsdk:"semantic_version"`
}

type resourceActionData struct {
	Type             types.String                 `tfsdk:"type"`
	IncludeResources resourceIncludeResourcesData `tfsdk:"include_resources"`
}

type resourceIncludeResourcesData struct {
	Amis       types.String `tfsdk:"amis"`
	Containers types.String `tfsdk:"containers"`
	Snapshots  types.String `tfsdk:"snapshots"`
}

type resourceFilterData struct {
	Type          types.String `tfsdk:"type"`
	Value         types.String `tfsdk:"value"`
	RetainAtLeast types.String `tfsdk:"retain_at_least"`
	Unit          types.String `tfsdk:"unit"`
}

type resourceExclusionRulesData struct {
	AMIs   resourceAMIsData `tfsdk:"ami"`
	TagMap types.String     `tfsdk:"tag_map"`
}

type resourceAMIsData struct {
	IsPublic       types.String             `tfsdk:"is_public"`
	Regions        types.String             `tfsdk:"regions"`
	SharedAccounts types.String             `tfsdk:"shared_accounts"`
	TagMap         types.String             `tfsdk:"tag_map"`
	LastLaunched   resourceLastLaunchedData `tfsdk:"last_launched"`
}

type resourceLastLaunchedData struct {
	Unit  types.String `tfsdk:"unit"`
	Value types.String `tfsdk:"value"`
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
			"id": framework.IDAttribute(),
			"name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"execution_role": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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
			"policy_details": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
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
									},
								},
								Blocks: map[string]schema.Block{
									"include_resources": schema.ListNestedBlock{
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
											listvalidator.IsRequired(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"amis": schema.StringAttribute{
													Required: true,
												},
												"containers": schema.StringAttribute{
													Required: true,
												},
												"snapshots": schema.StringAttribute{
													Required: true,
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
									},
									"value": schema.StringAttribute{
										Required: true,
									},
									"retain_at_least": schema.StringAttribute{
										Required: true,
									},
									"unit": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"exclusion_rules": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.IsRequired(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"tag_map": schema.StringAttribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"amis": schema.ListNestedBlock{
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
											listvalidator.IsRequired(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"is_public": schema.StringAttribute{
													Required: true,
												},
												"regions": schema.StringAttribute{
													Required: true,
												},
												"shared_accounts": schema.StringAttribute{
													Required: true,
												},
												"tag_map": schema.StringAttribute{
													Required: true,
												},
											},
											Blocks: map[string]schema.Block{
												"last_launched": schema.ListNestedBlock{
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
														listvalidator.IsRequired(),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"unit": schema.StringAttribute{
																Required: true,
															},
															"value": schema.StringAttribute{
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
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"tag_map": schema.StringAttribute{
							Required: true,
						},
					},
					Blocks: map[string]schema.Block{
						"recipes": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.IsRequired(),
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
		ClientToken: aws.String(id.UniqueId()),
		Name:        aws.String(plan.Name.ValueString()),
	}

	if !plan.Description.IsNull() {
		in.Description = aws.String(plan.Description.ValueString())
	}

	out, err := conn.CreateLifecyclePolicy(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ImageBuilder, create.ErrActionCreating, ResNameLifecyclePolicy, plan.Name.String(), nil),
			err.Error(),
		)
		return
	}

	state := plan
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceLifecyclePolicy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ImageBuilderClient(ctx)

	var state resourceLifecyclePolicyData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.GetLifecyclePolicy(ctx, &imagebuilder.GetLifecyclePolicyInput{
		LifecyclePolicyArn: aws.String(state.ID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ImageBuilder, create.ErrActionReading, ResNameLifecyclePolicy, state.Name.String(), nil),
			err.Error(),
		)
		return
	}
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
		!plan.Policy.Equal(state.Policy) {
		input := &opensearchserverless.UpdateAccessPolicyInput{
			ClientToken:   aws.String(id.UniqueId()),
			Name:          flex.StringFromFramework(ctx, plan.Name),
			PolicyVersion: flex.StringFromFramework(ctx, state.PolicyVersion),
			Type:          awstypes.AccessPolicyType(plan.Type.ValueString()),
		}

		if !plan.Description.Equal(state.Description) {
			input.Description = aws.String(plan.Description.ValueString())
		}

		if !plan.Policy.Equal(state.Policy) {
			input.Policy = aws.String(plan.Policy.ValueString())
		}

		out, err := conn.UpdateAccessPolicy(ctx, input)

		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("updating Security Policy (%s)", plan.Name.ValueString()), err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
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
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ImageBuilder, create.ErrActionDeleting, ResNameLifecyclePolicy, state.Name.String(), nil),
			err.Error(),
		)
	}
}
