// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"context"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func (r *resourceJobDefinition) jobDefinitionSchema0(ctx context.Context) schema.Schema {
	return schema.Schema{
		Version: 0,
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
			},
			// The ID includes the batch job definition version, and so it updates everytime
			// As a result we can't use framework.IDAttribute() do the plan modifier UseStateForUnknown
			names.AttrID: schema.StringAttribute{
				Computed: true,
			},
			"arn_prefix": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"deregister_on_new_revision": schema.BoolAttribute{
				Default:  booldefault.StaticBool(true),
				Optional: true,
				Computed: true,
			},

			"container_properties": schema.StringAttribute{
				Optional: true,
			},
			"ecs_properties": schema.StringAttribute{
				Optional: true,
			},
			"node_properties": schema.StringAttribute{
				Optional: true,
			},

			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[0-9A-Za-z]{1}[0-9A-Za-z_-]{0,127}$`),
						`must be up to 128 letters (uppercase and lowercase), numbers, underscores and dashes, and must start with an alphanumeric`,
					),
				},
			},

			names.AttrParameters: schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			"platform_capabilities": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						enum.FrameworkValidate[awstypes.PlatformCapability](),
					),
				},
			},
			names.AttrPropagateTags: schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"revision": schema.Int32Attribute{
				Computed: true,
			},
			"scheduling_priority": schema.Int32Attribute{
				Optional: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),

			names.AttrType: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.JobDefinitionType](),
					JobDefinitionTypeValidator{},
				},
			},
			names.AttrTimeout: schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[jobTimeoutModel](ctx),
				Optional:   true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"attempt_duration_seconds": types.Int64Type,
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"eks_properties": schema.ListNestedBlock{
				CustomType:   fwtypes.NewListNestedObjectTypeOf[eksPropertiesModel](ctx),
				NestedObject: r.SchemaEKSProperties(ctx),
			},

			"retry_strategy": schema.ListNestedBlock{ // https://docs.aws.amazon.com/batch/latest/APIReference/API_RetryStrategy.html
				CustomType: fwtypes.NewListNestedObjectTypeOf[retryStrategyModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"attempts": schema.Int32Attribute{
							Optional: true,
							Computed: true,
						},
					},
					Blocks: map[string]schema.Block{
						"evaluate_on_exit": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[evaluateOnExitModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrAction: schema.StringAttribute{
										// https://docs.aws.amazon.com/batch/latest/APIReference/API_EvaluateOnExit.html#Batch-Type-EvaluateOnExit-action
										// The only allowed values are "RETRY" and "EXIT".
										Validators: []validator.String{
											enum.FrameworkValidateIgnoreCase[awstypes.RetryAction](),
										},
										Optional: true,
										Computed: true,
									},
									"on_exit_code": schema.StringAttribute{
										Optional: true,
										Computed: true,
									},
									"on_reason": schema.StringAttribute{
										Optional: true,
										Computed: true,
									},
									"on_status_reason": schema.StringAttribute{
										Optional: true,
										Computed: true,
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

func upgradeJobDefinitionResourceStateV0toV1(ctx context.Context, request resource.UpgradeStateRequest, response *resource.UpgradeStateResponse) {
	type resourceJobDefinitionV0 struct {
		ARN                     types.String                                        `tfsdk:"arn"`
		ArnPrefix               types.String                                        `tfsdk:"arn_prefix" autoflex:"-"`
		ContainerProperties     types.String                                        `tfsdk:"container_properties"`
		DeregisterOnNewRevision types.Bool                                          `tfsdk:"deregister_on_new_revision"`
		ECSProperties           types.String                                        `tfsdk:"ecs_properties"`
		EKSProperties           fwtypes.ListNestedObjectValueOf[eksPropertiesModel] `tfsdk:"eks_properties"`
		ID                      types.String                                        `tfsdk:"id"`
		Name                    types.String                                        `tfsdk:"name"`
		NodeProperties          types.String                                        `tfsdk:"node_properties"`
		Parameters              fwtypes.MapOfString                                 `tfsdk:"parameters"`
		PlatformCapabilities    types.Set                                           `tfsdk:"platform_capabilities"`
		PropagateTags           types.Bool                                          `tfsdk:"propagate_tags" autoflex:",legacy"`
		Revision                types.Int32                                         `tfsdk:"revision"`
		RetryStrategy           fwtypes.ListNestedObjectValueOf[retryStrategyModel] `tfsdk:"retry_strategy"`
		SchedulingPriority      types.Int32                                         `tfsdk:"scheduling_priority"`
		Tags                    tftags.Map                                          `tfsdk:"tags"`
		TagsAll                 tftags.Map                                          `tfsdk:"tags_all"`
		Timeout                 fwtypes.ListNestedObjectValueOf[jobTimeoutModel]    `tfsdk:"timeout"`
		Type                    types.String                                        `tfsdk:"type"`
	}

	var jobDefV0 resourceJobDefinitionV0
	response.Diagnostics.Append(request.State.Get(ctx, &jobDefV0)...)
	if response.Diagnostics.HasError() {
		return
	}

	jobDefV1 := resourceJobDefinitionModel{
		ARN:                     jobDefV0.ARN,
		ArnPrefix:               jobDefV0.ArnPrefix,
		ContainerProperties:     fwtypes.NewListNestedObjectValueOfNull[containerPropertiesModel](ctx),
		DeregisterOnNewRevision: jobDefV0.DeregisterOnNewRevision,
		ECSProperties:           fwtypes.NewListNestedObjectValueOfNull[ecsPropertiesModel](ctx),
		EKSProperties:           jobDefV0.EKSProperties,
		ID:                      jobDefV0.ID,
		Name:                    jobDefV0.Name,
		NodeProperties:          fwtypes.NewListNestedObjectValueOfNull[nodePropertiesModel](ctx),
		Parameters:              jobDefV0.Parameters,
		PlatformCapabilities:    jobDefV0.PlatformCapabilities,
		PropagateTags:           jobDefV0.PropagateTags,
		Revision:                jobDefV0.Revision,
		RetryStrategy:           jobDefV0.RetryStrategy,
		SchedulingPriority:      jobDefV0.SchedulingPriority,
		Tags:                    jobDefV0.Tags,
		TagsAll:                 jobDefV0.TagsAll,
		Timeout:                 jobDefV0.Timeout,
		Type:                    jobDefV0.Type,
	}
	if jobDefV0.ContainerProperties.ValueString() != "" {
		props := &awstypes.ContainerProperties{}
		if err := tfjson.DecodeFromString(jobDefV0.ContainerProperties.ValueString(), props); err != nil {
			response.Diagnostics.AddError("Error marshalling container props", err.Error())
			return
		}
		response.Diagnostics.Append(fwflex.Flatten(ctx, props, &jobDefV1.ContainerProperties)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	if jobDefV0.ECSProperties.ValueString() != "" {
		props := &awstypes.EcsProperties{}
		if err := tfjson.DecodeFromString(jobDefV0.ECSProperties.ValueString(), props); err != nil {
			response.Diagnostics.AddError("Error marshalling ecs props", err.Error())
			return
		}
		response.Diagnostics.Append(fwflex.Flatten(ctx, props, &jobDefV1.ECSProperties)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	if jobDefV0.NodeProperties.ValueString() != "" {
		props := &awstypes.NodeProperties{}
		if err := tfjson.DecodeFromString(jobDefV0.NodeProperties.ValueString(), props); err != nil {
			response.Diagnostics.AddError("Error marshalling node props", err.Error())
			return
		}
		response.Diagnostics.Append(fwflex.Flatten(ctx, props, &jobDefV1.NodeProperties)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &jobDefV1)...)
}
