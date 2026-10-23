// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package eventbridgev2

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridgev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/eventbridgev2/types"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_eventbridgev2_subscriber", name="Subscriber")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/eventbridgev2;eventbridgev2.DescribeSubscriberOutput")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
// The service returns resolved defaults for these blocks, which are not tracked unless configured.
// @Testing(importIgnore="batch_configuration;log_configuration;resume_position;retry_policy;transformer")
func newSubscriberResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &subscriberResource{}, nil
}

type subscriberResource struct {
	framework.ResourceWithModel[subscriberResourceModel]
	framework.WithImportByIdentity
}

func (r *subscriberResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	jsonataStringAttribute := schema.StringAttribute{
		Optional: true,
	}

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreationTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"event_bus_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"last_modified_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z][0-9A-Za-z._-]*$`), "must start with an alphanumeric character and contain only alphanumeric characters, periods, hyphens, and underscores"),
				},
			},
			// Only honored when the subscriber transitions from STOPPED to RUNNING. Not returned by the API.
			"resume_position": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ResumePosition](),
				Optional:   true,
			},
			"revoked": schema.BoolAttribute{
				Computed: true,
			},
			"starting_position": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.StartingPosition](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrState: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.SubscriberState](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.OrderingType](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"batch_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[batchConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"max_batch_size": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.UseStateForUnknown(),
							},
							Validators: []validator.Int32{
								int32validator.Between(1, 500),
							},
						},
						"max_batch_window_in_seconds": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.UseStateForUnknown(),
							},
							Validators: []validator.Int32{
								int32validator.Between(0, 300),
							},
						},
					},
				},
			},
			"filter_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[filterConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"language": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.FilterLanguage](),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						names.AttrFilter: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[filterModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"pattern": schema.StringAttribute{
										CustomType: jsontypes.NormalizedType{},
										Required:   true,
									},
									names.AttrScope: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.FilterScope](),
										Required:   true,
									},
								},
							},
						},
					},
				},
			},
			"invoke_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[invokeConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrRoleARN: schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
						},
						names.AttrTargetARN: schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"event_bus_v2_parameters": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[eventBusV2ParametersModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"metadata": schema.MapAttribute{
										CustomType:  fwtypes.MapOfStringType,
										ElementType: types.StringType,
										Optional:    true,
									},
								},
								Blocks: map[string]schema.Block{
									"deduplication_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[deduplicationConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"deduplication_type": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.DeduplicationType](),
													Required:   true,
												},
											},
										},
									},
									"system_metadata": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[eventBusV2SystemMetadataModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"deduplication_id": jsonataStringAttribute,
												"event_group_id":   jsonataStringAttribute,
											},
										},
									},
								},
							},
						},
						"http_parameters": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[httpParametersModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"header_parameters": schema.MapAttribute{
										CustomType:  fwtypes.MapOfStringType,
										ElementType: types.StringType,
										Optional:    true,
									},
									"invocation_timeout_seconds": jsonataStringAttribute,
									"path_parameter_values": schema.ListAttribute{
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Optional:    true,
									},
									"query_string_parameters": schema.MapAttribute{
										CustomType:  fwtypes.MapOfStringType,
										ElementType: types.StringType,
										Optional:    true,
									},
								},
							},
						},
						"kinesis_parameters": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[kinesisParametersModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"explicit_hash_key": jsonataStringAttribute,
									"partition_key":     jsonataStringAttribute,
								},
							},
						},
						"lambda_parameters": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[lambdaParametersModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"durable_execution_name":     jsonataStringAttribute,
									"invocation_timeout_seconds": jsonataStringAttribute,
									"invocation_type": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.InvocationType](),
										Optional:   true,
									},
									"qualifier": jsonataStringAttribute,
									"tenant_id": jsonataStringAttribute,
								},
							},
						},
						"sns_parameters": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[snsParametersModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"message_deduplication_id": jsonataStringAttribute,
									"message_group_id":         jsonataStringAttribute,
									"message_structure":        jsonataStringAttribute,
									"subject":                  jsonataStringAttribute,
								},
								Blocks: map[string]schema.Block{
									"message_attributes": messageAttributeValueBlock[snsMessageAttributeValueModel](ctx),
								},
							},
						},
						"sqs_parameters": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[sqsParametersModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"delay_seconds":            jsonataStringAttribute,
									"message_deduplication_id": jsonataStringAttribute,
									"message_group_id":         jsonataStringAttribute,
								},
								Blocks: map[string]schema.Block{
									"message_attributes":        messageAttributeValueBlock[sqsMessageAttributeValueModel](ctx),
									"message_system_attributes": messageAttributeValueBlock[sqsMessageAttributeValueModel](ctx),
								},
							},
						},
						"step_functions_parameters": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[stepFunctionsParametersModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"invocation_timeout_seconds": jsonataStringAttribute,
									"invocation_type": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.InvocationType](),
										Optional:   true,
									},
									names.AttrName: jsonataStringAttribute,
									"trace_header": jsonataStringAttribute,
								},
							},
						},
						"universal_target_parameters": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[universalTargetParametersModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"input": schema.StringAttribute{
										Required: true,
									},
									"invocation_timeout_seconds": jsonataStringAttribute,
								},
							},
						},
					},
				},
			},
			"log_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[logConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"include_payload": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.IncludePayload](),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"level": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.LogLevel](),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			"on_failure_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[onFailureConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrARN: schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
						},
					},
				},
			},
			"point_in_time_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[pointInTimeConfigurationModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"end_point": schema.StringAttribute{
							CustomType: timetypes.RFC3339Type{},
							Optional:   true,
						},
						"point_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.PointType](),
							Required:   true,
						},
						"starting_point": schema.StringAttribute{
							CustomType: timetypes.RFC3339Type{},
							Optional:   true,
						},
					},
				},
			},
			"retry_policy": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[retryPolicyModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"max_event_age_in_seconds": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.UseStateForUnknown(),
							},
							Validators: []validator.Int32{
								int32validator.Between(60, 86400),
							},
						},
						"max_retry_attempts": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.UseStateForUnknown(),
							},
							Validators: []validator.Int32{
								int32validator.Between(0, 185),
							},
						},
						"retry_strategy": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.RetryStrategy](),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			"transformer": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[transformerModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrType: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.TransformerType](),
							Optional:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"jsonata_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[jsonataConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrExpression: schema.StringAttribute{
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

func messageAttributeValueBlock[T any](ctx context.Context) schema.SetNestedBlock {
	return schema.SetNestedBlock{
		CustomType: fwtypes.NewSetNestedObjectTypeOf[T](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"binary_value": schema.StringAttribute{
					Optional: true,
				},
				"data_type": schema.StringAttribute{
					Optional: true,
				},
				names.AttrName: schema.StringAttribute{
					Required: true,
				},
				"string_value": schema.StringAttribute{
					Optional: true,
				},
			},
		},
	}
}

func (r *subscriberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().EventBridgeV2Client(ctx)

	var plan subscriberResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input eventbridgev2.CreateSubscriberInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Subscriber")))
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	out, err := tfresource.RetryWhenIsA[*eventbridgev2.CreateSubscriberOutput, *awstypes.ConcurrentModificationException](ctx, concurrentModificationTimeout, func(ctx context.Context) (*eventbridgev2.CreateSubscriberOutput, error) {
		return conn.CreateSubscriber(ctx, &input)
	})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.Name, name)
		return
	}

	arn := aws.ToString(out.SubscriberArn)

	// Subscribers are provisioned synchronously, so read back the full configuration immediately.
	output, err := findSubscriberByARN(ctx, conn, arn)
	if err != nil {
		// Taint the resource.
		resp.State.SetAttribute(ctx, path.Root(names.AttrARN), arn)
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flattenConfigured(ctx, output, plan, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *subscriberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().EventBridgeV2Client(ctx)

	var state subscriberResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	out, err := findSubscriberByARN(ctx, conn, arn)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flattenConfigured(ctx, out, state, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *subscriberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().EventBridgeV2Client(ctx)

	var plan, state subscriberResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		arn := fwflex.StringValueFromFramework(ctx, plan.ARN)
		var input eventbridgev2.UpdateSubscriberInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Subscriber")))
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := tfresource.RetryWhenIsA[*eventbridgev2.UpdateSubscriberOutput, *awstypes.ConcurrentModificationException](ctx, concurrentModificationTimeout, func(ctx context.Context) (*eventbridgev2.UpdateSubscriberOutput, error) {
			return conn.UpdateSubscriber(ctx, &input)
		})
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		output, err := findSubscriberByARN(ctx, conn, arn)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, r.flattenConfigured(ctx, output, plan, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		plan.LastModifiedTime = state.LastModifiedTime
		plan.Revoked = state.Revoked
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *subscriberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().EventBridgeV2Client(ctx)

	var state subscriberResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	input := eventbridgev2.DeleteSubscriberInput{
		SubscriberArn: aws.String(arn),
	}

	_, err := tfresource.RetryWhenIsA[*eventbridgev2.DeleteSubscriberOutput, *awstypes.ConcurrentModificationException](ctx, concurrentModificationTimeout, func(ctx context.Context) (*eventbridgev2.DeleteSubscriberOutput, error) {
		return conn.DeleteSubscriber(ctx, &input)
	})
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}
}

func (r *subscriberResource) flatten(ctx context.Context, out *eventbridgev2.DescribeSubscriberOutput, data *subscriberResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, out, data, fwflex.WithFieldNamePrefix("Subscriber")))
	if diags.HasError() {
		return diags
	}

	// The service returns an empty description when none is set.
	if aws.ToString(out.Description) == "" {
		data.Description = types.StringNull()
	}

	return diags
}

// flattenConfigured flattens the API response, but keeps the batch, log, retry, and transformer configurations empty when they
// are not configured. The service always returns resolved defaults for these, which would otherwise cause a perpetual
// diff because blocks cannot be Computed. On import (no prior name), the service values are kept.
func (r *subscriberResource) flattenConfigured(ctx context.Context, out *eventbridgev2.DescribeSubscriberOutput, prior subscriberResourceModel, data *subscriberResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	smerr.AddEnrich(ctx, &diags, r.flatten(ctx, out, data))
	if diags.HasError() {
		return diags
	}

	if prior.Name.IsNull() {
		return diags
	}

	if prior.BatchConfiguration.Length(fwtypes.CollectionLengthUnhandledAsZero) == 0 {
		data.BatchConfiguration = prior.BatchConfiguration
	}
	if prior.LogConfiguration.Length(fwtypes.CollectionLengthUnhandledAsZero) == 0 {
		data.LogConfiguration = prior.LogConfiguration
	}
	if prior.RetryPolicy.Length(fwtypes.CollectionLengthUnhandledAsZero) == 0 {
		data.RetryPolicy = prior.RetryPolicy
	}
	if prior.Transformer.Length(fwtypes.CollectionLengthUnhandledAsZero) == 0 {
		data.Transformer = prior.Transformer
	}

	return diags
}

func findSubscriberByARN(ctx context.Context, conn *eventbridgev2.Client, arn string) (*eventbridgev2.DescribeSubscriberOutput, error) {
	input := eventbridgev2.DescribeSubscriberInput{
		SubscriberArn: aws.String(arn),
	}

	out, err := conn.DescribeSubscriber(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type subscriberResourceModel struct {
	framework.WithRegionModel
	ARN                      types.String                                                   `tfsdk:"arn"`
	BatchConfiguration       fwtypes.ListNestedObjectValueOf[batchConfigurationModel]       `tfsdk:"batch_configuration"`
	CreationTime             timetypes.RFC3339                                              `tfsdk:"creation_time"`
	Description              types.String                                                   `tfsdk:"description"`
	EventBusARN              fwtypes.ARN                                                    `tfsdk:"event_bus_arn"`
	FilterConfiguration      fwtypes.ListNestedObjectValueOf[filterConfigurationModel]      `tfsdk:"filter_configuration"`
	InvokeConfiguration      fwtypes.ListNestedObjectValueOf[invokeConfigurationModel]      `tfsdk:"invoke_configuration"`
	LastModifiedTime         timetypes.RFC3339                                              `tfsdk:"last_modified_time"`
	LogConfiguration         fwtypes.ListNestedObjectValueOf[logConfigurationModel]         `tfsdk:"log_configuration"`
	Name                     types.String                                                   `tfsdk:"name"`
	OnFailureConfiguration   fwtypes.ListNestedObjectValueOf[onFailureConfigurationModel]   `tfsdk:"on_failure_configuration"`
	PointInTimeConfiguration fwtypes.ListNestedObjectValueOf[pointInTimeConfigurationModel] `tfsdk:"point_in_time_configuration"`
	ResumePosition           fwtypes.StringEnum[awstypes.ResumePosition]                    `tfsdk:"resume_position"`
	RetryPolicy              fwtypes.ListNestedObjectValueOf[retryPolicyModel]              `tfsdk:"retry_policy"`
	Revoked                  types.Bool                                                     `tfsdk:"revoked"`
	StartingPosition         fwtypes.StringEnum[awstypes.StartingPosition]                  `tfsdk:"starting_position"`
	State                    fwtypes.StringEnum[awstypes.SubscriberState]                   `tfsdk:"state"`
	Tags                     tftags.Map                                                     `tfsdk:"tags"`
	TagsAll                  tftags.Map                                                     `tfsdk:"tags_all"`
	Transformer              fwtypes.ListNestedObjectValueOf[transformerModel]              `tfsdk:"transformer"`
	Type                     fwtypes.StringEnum[awstypes.OrderingType]                      `tfsdk:"type"`
}

type batchConfigurationModel struct {
	MaxBatchSize            types.Int32 `tfsdk:"max_batch_size"`
	MaxBatchWindowInSeconds types.Int32 `tfsdk:"max_batch_window_in_seconds"`
}

type filterConfigurationModel struct {
	Filters  fwtypes.ListNestedObjectValueOf[filterModel] `tfsdk:"filter"`
	Language fwtypes.StringEnum[awstypes.FilterLanguage]  `tfsdk:"language"`
}

type filterModel struct {
	Pattern jsontypes.Normalized                     `tfsdk:"pattern"`
	Scope   fwtypes.StringEnum[awstypes.FilterScope] `tfsdk:"scope"`
}

type invokeConfigurationModel struct {
	EventBusV2Parameters      fwtypes.ListNestedObjectValueOf[eventBusV2ParametersModel]      `tfsdk:"event_bus_v2_parameters"`
	HTTPParameters            fwtypes.ListNestedObjectValueOf[httpParametersModel]            `tfsdk:"http_parameters"`
	KinesisParameters         fwtypes.ListNestedObjectValueOf[kinesisParametersModel]         `tfsdk:"kinesis_parameters"`
	LambdaParameters          fwtypes.ListNestedObjectValueOf[lambdaParametersModel]          `tfsdk:"lambda_parameters"`
	RoleARN                   fwtypes.ARN                                                     `tfsdk:"role_arn"`
	SNSParameters             fwtypes.ListNestedObjectValueOf[snsParametersModel]             `tfsdk:"sns_parameters"`
	SQSParameters             fwtypes.ListNestedObjectValueOf[sqsParametersModel]             `tfsdk:"sqs_parameters"`
	StepFunctionsParameters   fwtypes.ListNestedObjectValueOf[stepFunctionsParametersModel]   `tfsdk:"step_functions_parameters"`
	TargetARN                 types.String                                                    `tfsdk:"target_arn"`
	UniversalTargetParameters fwtypes.ListNestedObjectValueOf[universalTargetParametersModel] `tfsdk:"universal_target_parameters"`
}

type eventBusV2ParametersModel struct {
	DeduplicationConfiguration fwtypes.ListNestedObjectValueOf[deduplicationConfigurationModel] `tfsdk:"deduplication_configuration"`
	Metadata                   fwtypes.MapOfString                                              `tfsdk:"metadata"`
	SystemMetadata             fwtypes.ListNestedObjectValueOf[eventBusV2SystemMetadataModel]   `tfsdk:"system_metadata"`
}

type deduplicationConfigurationModel struct {
	DeduplicationType fwtypes.StringEnum[awstypes.DeduplicationType] `tfsdk:"deduplication_type"`
}

type eventBusV2SystemMetadataModel struct {
	DeduplicationID types.String `tfsdk:"deduplication_id"`
	EventGroupID    types.String `tfsdk:"event_group_id"`
}

type httpParametersModel struct {
	HeaderParameters         fwtypes.MapOfString  `tfsdk:"header_parameters"`
	InvocationTimeoutSeconds types.String         `tfsdk:"invocation_timeout_seconds"`
	PathParameterValues      fwtypes.ListOfString `tfsdk:"path_parameter_values"`
	QueryStringParameters    fwtypes.MapOfString  `tfsdk:"query_string_parameters"`
}

type kinesisParametersModel struct {
	ExplicitHashKey types.String `tfsdk:"explicit_hash_key"`
	PartitionKey    types.String `tfsdk:"partition_key"`
}

type lambdaParametersModel struct {
	DurableExecutionName     types.String                                `tfsdk:"durable_execution_name"`
	InvocationTimeoutSeconds types.String                                `tfsdk:"invocation_timeout_seconds"`
	InvocationType           fwtypes.StringEnum[awstypes.InvocationType] `tfsdk:"invocation_type"`
	Qualifier                types.String                                `tfsdk:"qualifier"`
	TenantID                 types.String                                `tfsdk:"tenant_id"`
}

type snsParametersModel struct {
	MessageAttributes      fwtypes.SetNestedObjectValueOf[snsMessageAttributeValueModel] `tfsdk:"message_attributes"`
	MessageDeduplicationID types.String                                                  `tfsdk:"message_deduplication_id"`
	MessageGroupID         types.String                                                  `tfsdk:"message_group_id"`
	MessageStructure       types.String                                                  `tfsdk:"message_structure"`
	Subject                types.String                                                  `tfsdk:"subject"`
}

type snsMessageAttributeValueModel struct {
	BinaryValue types.String `tfsdk:"binary_value"`
	DataType    types.String `tfsdk:"data_type"`
	MapBlockKey types.String `tfsdk:"name"`
	StringValue types.String `tfsdk:"string_value"`
}

type sqsParametersModel struct {
	DelaySeconds            types.String                                                  `tfsdk:"delay_seconds"`
	MessageAttributes       fwtypes.SetNestedObjectValueOf[sqsMessageAttributeValueModel] `tfsdk:"message_attributes"`
	MessageDeduplicationID  types.String                                                  `tfsdk:"message_deduplication_id"`
	MessageGroupID          types.String                                                  `tfsdk:"message_group_id"`
	MessageSystemAttributes fwtypes.SetNestedObjectValueOf[sqsMessageAttributeValueModel] `tfsdk:"message_system_attributes"`
}

type sqsMessageAttributeValueModel struct {
	BinaryValue types.String `tfsdk:"binary_value"`
	DataType    types.String `tfsdk:"data_type"`
	MapBlockKey types.String `tfsdk:"name"`
	StringValue types.String `tfsdk:"string_value"`
}

type stepFunctionsParametersModel struct {
	InvocationTimeoutSeconds types.String                                `tfsdk:"invocation_timeout_seconds"`
	InvocationType           fwtypes.StringEnum[awstypes.InvocationType] `tfsdk:"invocation_type"`
	Name                     types.String                                `tfsdk:"name"`
	TraceHeader              types.String                                `tfsdk:"trace_header"`
}

type universalTargetParametersModel struct {
	Input                    types.String `tfsdk:"input"`
	InvocationTimeoutSeconds types.String `tfsdk:"invocation_timeout_seconds"`
}

type logConfigurationModel struct {
	IncludePayload fwtypes.StringEnum[awstypes.IncludePayload] `tfsdk:"include_payload"`
	Level          fwtypes.StringEnum[awstypes.LogLevel]       `tfsdk:"level"`
}

type pointInTimeConfigurationModel struct {
	EndPoint      timetypes.RFC3339                      `tfsdk:"end_point"`
	PointType     fwtypes.StringEnum[awstypes.PointType] `tfsdk:"point_type"`
	StartingPoint timetypes.RFC3339                      `tfsdk:"starting_point"`
}

type retryPolicyModel struct {
	MaxEventAgeInSeconds types.Int32                                `tfsdk:"max_event_age_in_seconds"`
	MaxRetryAttempts     types.Int32                                `tfsdk:"max_retry_attempts"`
	RetryStrategy        fwtypes.StringEnum[awstypes.RetryStrategy] `tfsdk:"retry_strategy"`
}

type transformerModel struct {
	JSONataConfiguration fwtypes.ListNestedObjectValueOf[jsonataConfigurationModel] `tfsdk:"jsonata_configuration"`
	Type                 fwtypes.StringEnum[awstypes.TransformerType]               `tfsdk:"type"`
}

type jsonataConfigurationModel struct {
	Expression types.String `tfsdk:"expression"`
}
