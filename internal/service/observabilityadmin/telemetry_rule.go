// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin

import (
	"context"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/observabilityadmin/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	tfobjectvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/objectvalidator"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_observabilityadmin_telemetry_rule", name="Telemetry Rule")
// @Tags(identifierAttribute="rule_arn")
// @IdentityAttribute("rule_name")
// @Testing(tagsTest=false)
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttribute="rule_name")
// @Testing(serialize=true)
func newTelemetryRuleResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &telemetryRuleResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type telemetryRuleResource struct {
	framework.ResourceWithModel[telemetryRuleResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *telemetryRuleResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"rule_arn": framework.ARNAttributeComputedOnly(),
			"rule_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z\-_.#/]+$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrRule: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[telemetryRuleModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"all_regions": schema.BoolAttribute{
							Optional: true,
						},
						"allow_field_updates": schema.BoolAttribute{
							Optional: true,
						},
						// regions is treated as an unordered set because the AWS API does
						// not preserve the input order on GET (returns sorted). Using a
						// SetAttribute avoids "provider produced inconsistent result"
						// errors when the planned order differs from the API-returned
						// order. Marked Computed with UseStateForUnknown so that when the
						// field is omitted from config, state retains the API-returned
						// value across refreshes.
						"regions": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Optional:    true,
							Computed:    true,
							Validators: []validator.Set{
								setvalidator.ValueStringsAre(fwvalidators.AWSRegion()),
							},
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.UseStateForUnknown(),
							},
						},
						names.AttrResourceType: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ResourceType](),
							Optional:   true,
						},
						names.AttrScope: schema.StringAttribute{
							Optional: true,
						},
						"selection_criteria": schema.StringAttribute{
							Optional: true,
						},
						"telemetry_source_types": schema.ListAttribute{
							CustomType: fwtypes.ListOfStringEnumType[awstypes.TelemetrySourceType](),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
						},
						"telemetry_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.TelemetryType](),
							Required:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"destination_configuration": telemetryRuleDestinationConfigurationBlock(ctx),
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *telemetryRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data telemetryRuleResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	ruleName := fwflex.StringValueFromFramework(ctx, data.RuleName)
	var input observabilityadmin.CreateTelemetryRuleInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateTelemetryRule(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleName)
		return
	}

	// Set values for unknowns.
	data.RuleARN = fwflex.StringToFramework(ctx, output.RuleArn)

	// Read back to populate computed fields (e.g. API-derived telemetry_source_types,
	// regions when all_regions is true).
	out, err := findTelemetryRuleByName(ctx, conn, ruleName)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleName)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *telemetryRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data telemetryRuleResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	ruleName := fwflex.StringValueFromFramework(ctx, data.RuleName)
	out, err := findTelemetryRuleByName(ctx, conn, ruleName)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleName)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *telemetryRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old telemetryRuleResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		ruleName := fwflex.StringValueFromFramework(ctx, new.RuleName)
		var input observabilityadmin.UpdateTelemetryRuleInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input))
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.RuleIdentifier = aws.String(ruleName)

		_, err := conn.UpdateTelemetryRule(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleName)
			return
		}

		// Read back to refresh computed fields after the update.
		out, err := findTelemetryRuleByName(ctx, conn, ruleName)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleName)
			return
		}

		smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, out, &new))
		if response.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *telemetryRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data telemetryRuleResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	ruleName := fwflex.StringValueFromFramework(ctx, data.RuleName)
	input := observabilityadmin.DeleteTelemetryRuleInput{
		RuleIdentifier: aws.String(ruleName),
	}

	_, err := conn.DeleteTelemetryRule(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleName)
		return
	}
}

func (r *telemetryRuleResource) flatten(ctx context.Context, telemetryRule *observabilityadmin.GetTelemetryRuleOutput, data *telemetryRuleResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, telemetryRule, data, fwflex.WithFieldNamePrefix("Telemetry"))...)
	return diags
}

func findTelemetryRuleByName(ctx context.Context, conn *observabilityadmin.Client, name string) (*observabilityadmin.GetTelemetryRuleOutput, error) {
	input := observabilityadmin.GetTelemetryRuleInput{
		RuleIdentifier: aws.String(name),
	}

	return findTelemetryRule(ctx, conn, &input)
}

func findTelemetryRule(ctx context.Context, conn *observabilityadmin.Client, input *observabilityadmin.GetTelemetryRuleInput) (*observabilityadmin.GetTelemetryRuleOutput, error) {
	output, err := conn.GetTelemetryRule(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "Telemetry evaluation is not enabled") {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.TelemetryRule == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func telemetryRuleDestinationConfigurationBlock(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[destinationConfigurationModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"destination_pattern": schema.StringAttribute{
					Optional: true,
				},
				"destination_type": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.DestinationType](),
					Optional:   true,
				},
				"retention_in_days": schema.Int32Attribute{
					Optional: true,
				},
			},
			Blocks: map[string]schema.Block{
				"cloudtrail_parameters": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[cloudtrailParametersModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Blocks: map[string]schema.Block{
							"advanced_event_selectors": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[advancedEventSelectorModel](ctx),
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										names.AttrName: schema.StringAttribute{
											Optional: true,
										},
									},
									Blocks: map[string]schema.Block{
										"field_selectors": schema.ListNestedBlock{
											CustomType: fwtypes.NewListNestedObjectTypeOf[advancedFieldSelectorModel](ctx),
											NestedObject: schema.NestedBlockObject{
												Attributes: map[string]schema.Attribute{
													names.AttrField: schema.StringAttribute{
														Required: true,
													},
													"ends_with": schema.ListAttribute{
														CustomType:  fwtypes.ListOfStringType,
														ElementType: types.StringType,
														Optional:    true,
													},
													"equals": schema.ListAttribute{
														CustomType:  fwtypes.ListOfStringType,
														ElementType: types.StringType,
														Optional:    true,
													},
													"not_ends_with": schema.ListAttribute{
														CustomType:  fwtypes.ListOfStringType,
														ElementType: types.StringType,
														Optional:    true,
													},
													"not_equals": schema.ListAttribute{
														CustomType:  fwtypes.ListOfStringType,
														ElementType: types.StringType,
														Optional:    true,
													},
													"not_starts_with": schema.ListAttribute{
														CustomType:  fwtypes.ListOfStringType,
														ElementType: types.StringType,
														Optional:    true,
													},
													"starts_with": schema.ListAttribute{
														CustomType:  fwtypes.ListOfStringType,
														ElementType: types.StringType,
														Optional:    true,
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
				"elb_load_balancer_logging_parameters": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[elbLoadBalancerLoggingParametersModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"field_delimiter": schema.StringAttribute{
								Optional: true,
							},
							"output_format": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.OutputFormat](),
								Optional:   true,
							},
						},
					},
				},
				"log_delivery_parameters": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[logDeliveryParametersModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"log_types": schema.ListAttribute{
								CustomType: fwtypes.ListOfStringEnumType[awstypes.LogType](),
								Optional:   true,
							},
						},
					},
				},
				"msk_monitoring_parameters": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[mskMonitoringParametersModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"enhanced_monitoring": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.MskEnhancedMonitoringLevel](),
								Optional:   true,
							},
						},
					},
				},
				"vpc_flow_log_parameters": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[vpcFlowLogParametersModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"log_format": schema.StringAttribute{
								Optional: true,
							},
							"max_aggregation_interval": schema.Int32Attribute{
								Optional: true,
							},
							"traffic_type": schema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
				"waf_logging_parameters": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[wafLoggingParametersModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"log_type": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.WAFLogType](),
								Optional:   true,
							},
						},
						Blocks: map[string]schema.Block{
							"logging_filter": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[loggingFilterModel](ctx),
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"default_behavior": schema.StringAttribute{
											CustomType: fwtypes.StringEnumType[awstypes.FilterBehavior](),
											Optional:   true,
										},
									},
									Blocks: map[string]schema.Block{
										"filters": schema.ListNestedBlock{
											CustomType: fwtypes.NewListNestedObjectTypeOf[filterModel](ctx),
											NestedObject: schema.NestedBlockObject{
												Attributes: map[string]schema.Attribute{
													"behavior": schema.StringAttribute{
														CustomType: fwtypes.StringEnumType[awstypes.FilterBehavior](),
														Optional:   true,
													},
													"requirement": schema.StringAttribute{
														CustomType: fwtypes.StringEnumType[awstypes.FilterRequirement](),
														Optional:   true,
													},
												},
												Blocks: map[string]schema.Block{
													"conditions": schema.ListNestedBlock{
														CustomType: fwtypes.NewListNestedObjectTypeOf[conditionModel](ctx),
														NestedObject: schema.NestedBlockObject{
															Validators: []validator.Object{
																tfobjectvalidator.ExactlyOneOfChildren(
																	path.MatchRelative().AtName("action_condition"),
																	path.MatchRelative().AtName("label_name_condition"),
																),
															},
															Blocks: map[string]schema.Block{
																"action_condition": schema.ListNestedBlock{
																	CustomType: fwtypes.NewListNestedObjectTypeOf[actionConditionModel](ctx),
																	Validators: []validator.List{
																		listvalidator.SizeAtMost(1),
																	},
																	NestedObject: schema.NestedBlockObject{
																		Attributes: map[string]schema.Attribute{
																			names.AttrAction: schema.StringAttribute{
																				CustomType: fwtypes.StringEnumType[awstypes.Action](),
																				Required:   true,
																			},
																		},
																	},
																},
																"label_name_condition": schema.ListNestedBlock{
																	CustomType: fwtypes.NewListNestedObjectTypeOf[labelNameConditionModel](ctx),
																	Validators: []validator.List{
																		listvalidator.SizeAtMost(1),
																	},
																	NestedObject: schema.NestedBlockObject{
																		Attributes: map[string]schema.Attribute{
																			"label_name": schema.StringAttribute{
																				Optional: true,
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
							"redacted_fields": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[fieldToMatchModel](ctx),
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"method": schema.StringAttribute{
											Optional: true,
										},
										"query_string": schema.StringAttribute{
											Optional: true,
										},
										"uri_path": schema.StringAttribute{
											Optional: true,
										},
									},
									Blocks: map[string]schema.Block{
										"single_header": schema.ListNestedBlock{
											CustomType: fwtypes.NewListNestedObjectTypeOf[singleHeaderModel](ctx),
											Validators: []validator.List{
												listvalidator.SizeAtMost(1),
											},
											NestedObject: schema.NestedBlockObject{
												Attributes: map[string]schema.Attribute{
													names.AttrName: schema.StringAttribute{
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
	}
}

type telemetryRuleResourceModel struct {
	framework.WithRegionModel
	Rule     fwtypes.ListNestedObjectValueOf[telemetryRuleModel] `tfsdk:"rule"`
	RuleARN  types.String                                        `tfsdk:"rule_arn"`
	RuleName types.String                                        `tfsdk:"rule_name"`
	Tags     tftags.Map                                          `tfsdk:"tags"`
	TagsAll  tftags.Map                                          `tfsdk:"tags_all"`
	Timeouts timeouts.Value                                      `tfsdk:"timeouts"`
}

type telemetryRuleModel struct {
	AllRegions               types.Bool                                                            `tfsdk:"all_regions"`
	AllowFieldUpdates        types.Bool                                                            `tfsdk:"allow_field_updates"`
	DestinationConfiguration fwtypes.ListNestedObjectValueOf[destinationConfigurationModel]        `tfsdk:"destination_configuration"`
	Regions                  fwtypes.SetValueOf[types.String]                                      `tfsdk:"regions"`
	ResourceType             fwtypes.StringEnum[awstypes.ResourceType]                             `tfsdk:"resource_type"`
	Scope                    types.String                                                          `tfsdk:"scope"`
	SelectionCriteria        types.String                                                          `tfsdk:"selection_criteria"`
	TelemetrySourceTypes     fwtypes.ListValueOf[fwtypes.StringEnum[awstypes.TelemetrySourceType]] `tfsdk:"telemetry_source_types"`
	TelemetryType            fwtypes.StringEnum[awstypes.TelemetryType]                            `tfsdk:"telemetry_type"`
}

type destinationConfigurationModel struct {
	CloudtrailParameters             fwtypes.ListNestedObjectValueOf[cloudtrailParametersModel]             `tfsdk:"cloudtrail_parameters"`
	DestinationPattern               types.String                                                           `tfsdk:"destination_pattern"`
	DestinationType                  fwtypes.StringEnum[awstypes.DestinationType]                           `tfsdk:"destination_type"`
	ELBLoadBalancerLoggingParameters fwtypes.ListNestedObjectValueOf[elbLoadBalancerLoggingParametersModel] `tfsdk:"elb_load_balancer_logging_parameters"`
	LogDeliveryParameters            fwtypes.ListNestedObjectValueOf[logDeliveryParametersModel]            `tfsdk:"log_delivery_parameters"`
	MskMonitoringParameters          fwtypes.ListNestedObjectValueOf[mskMonitoringParametersModel]          `tfsdk:"msk_monitoring_parameters"`
	RetentionInDays                  types.Int32                                                            `tfsdk:"retention_in_days"`
	VPCFlowLogParameters             fwtypes.ListNestedObjectValueOf[vpcFlowLogParametersModel]             `tfsdk:"vpc_flow_log_parameters"`
	WAFLoggingParameters             fwtypes.ListNestedObjectValueOf[wafLoggingParametersModel]             `tfsdk:"waf_logging_parameters"`
}

type mskMonitoringParametersModel struct {
	EnhancedMonitoring fwtypes.StringEnum[awstypes.MskEnhancedMonitoringLevel] `tfsdk:"enhanced_monitoring"`
}

type cloudtrailParametersModel struct {
	AdvancedEventSelectors fwtypes.ListNestedObjectValueOf[advancedEventSelectorModel] `tfsdk:"advanced_event_selectors"`
}

type advancedEventSelectorModel struct {
	FieldSelectors fwtypes.ListNestedObjectValueOf[advancedFieldSelectorModel] `tfsdk:"field_selectors"`
	Name           types.String                                                `tfsdk:"name"`
}

type advancedFieldSelectorModel struct {
	Field         types.String                      `tfsdk:"field"`
	EndsWith      fwtypes.ListValueOf[types.String] `tfsdk:"ends_with"`
	Equals        fwtypes.ListValueOf[types.String] `tfsdk:"equals"`
	NotEndsWith   fwtypes.ListValueOf[types.String] `tfsdk:"not_ends_with"`
	NotEquals     fwtypes.ListValueOf[types.String] `tfsdk:"not_equals"`
	NotStartsWith fwtypes.ListValueOf[types.String] `tfsdk:"not_starts_with"`
	StartsWith    fwtypes.ListValueOf[types.String] `tfsdk:"starts_with"`
}

type elbLoadBalancerLoggingParametersModel struct {
	FieldDelimiter types.String                              `tfsdk:"field_delimiter"`
	OutputFormat   fwtypes.StringEnum[awstypes.OutputFormat] `tfsdk:"output_format"`
}

type logDeliveryParametersModel struct {
	LogTypes fwtypes.ListValueOf[fwtypes.StringEnum[awstypes.LogType]] `tfsdk:"log_types"`
}

type vpcFlowLogParametersModel struct {
	LogFormat              types.String `tfsdk:"log_format"`
	MaxAggregationInterval types.Int32  `tfsdk:"max_aggregation_interval"`
	TrafficType            types.String `tfsdk:"traffic_type"`
}

type wafLoggingParametersModel struct {
	LogType        fwtypes.StringEnum[awstypes.WAFLogType]             `tfsdk:"log_type"`
	LoggingFilter  fwtypes.ListNestedObjectValueOf[loggingFilterModel] `tfsdk:"logging_filter"`
	RedactedFields fwtypes.ListNestedObjectValueOf[fieldToMatchModel]  `tfsdk:"redacted_fields"`
}

type loggingFilterModel struct {
	DefaultBehavior fwtypes.StringEnum[awstypes.FilterBehavior]  `tfsdk:"default_behavior"`
	Filters         fwtypes.ListNestedObjectValueOf[filterModel] `tfsdk:"filters"`
}

type filterModel struct {
	Behavior    fwtypes.StringEnum[awstypes.FilterBehavior]     `tfsdk:"behavior"`
	Conditions  fwtypes.ListNestedObjectValueOf[conditionModel] `tfsdk:"conditions"`
	Requirement fwtypes.StringEnum[awstypes.FilterRequirement]  `tfsdk:"requirement"`
}

type conditionModel struct {
	ActionCondition    fwtypes.ListNestedObjectValueOf[actionConditionModel]    `tfsdk:"action_condition"`
	LabelNameCondition fwtypes.ListNestedObjectValueOf[labelNameConditionModel] `tfsdk:"label_name_condition"`
}

type actionConditionModel struct {
	Action fwtypes.StringEnum[awstypes.Action] `tfsdk:"action"`
}

type labelNameConditionModel struct {
	LabelName types.String `tfsdk:"label_name"`
}

type fieldToMatchModel struct {
	Method       types.String                                       `tfsdk:"method"`
	QueryString  types.String                                       `tfsdk:"query_string"`
	SingleHeader fwtypes.ListNestedObjectValueOf[singleHeaderModel] `tfsdk:"single_header"`
	UriPath      types.String                                       `tfsdk:"uri_path"`
}

type singleHeaderModel struct {
	Name types.String `tfsdk:"name"`
}
