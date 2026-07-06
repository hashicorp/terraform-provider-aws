// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_gateway_rule", name="Gateway Rule")
func newGatewayRuleResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &gatewayRuleResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type gatewayRuleResource struct {
	framework.ResourceWithModel[gatewayRuleResourceModel]
	framework.WithTimeouts
}

func (r *gatewayRuleResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"gateway_arn": framework.ARNAttributeComputedOnly(),
			"gateway_identifier": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^([0-9a-z][-]?){1,100}-[0-9a-z]{10}$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrPriority: schema.Int64Attribute{
				Required: true,
				Validators: []validator.Int64{
					int64validator.Between(1, 1000000),
				},
			},
			"rule_id": framework.IDAttribute(),
			names.AttrStatus: schema.StringAttribute{
				Computed:   true,
				CustomType: fwtypes.StringEnumType[awstypes.GatewayRuleStatus](),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"system": framework.ResourceComputedListOfObjectsAttribute[systemManagedBlockModel](ctx),
		},
		Blocks: map[string]schema.Block{
			names.AttrAction: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[actionModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 2),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"configuration_bundle": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[configurationBundleActionModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("route_to_target"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"static_override": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[staticOverrideModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
											listvalidator.ExactlyOneOf(
												path.MatchRelative().AtParent().AtName("weighted_override"),
											),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"bundle_arn": schema.StringAttribute{
													Required:   true,
													CustomType: fwtypes.ARNType,
													Validators: []validator.String{
														stringvalidator.RegexMatches(
															regexache.MustCompile(`^arn:aws[a-zA-Z-]*:bedrock-agentcore:[a-z0-9-]+:[0-9]{12}:configuration-bundle/[a-zA-Z][a-zA-Z0-9-_]{0,99}-[a-zA-Z0-9]{10}$`),
															"",
														),
													},
												},
												"bundle_version": schema.StringAttribute{
													Required: true,
												},
											},
										},
									},
									"weighted_override": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[weightedOverrideModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"traffic_split": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[trafficSplitEntryModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeBetween(2, 2),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrDescription: schema.StringAttribute{
																Optional: true,
															},
															"metadata": schema.MapAttribute{
																CustomType:  fwtypes.MapOfStringType,
																ElementType: types.StringType,
																Optional:    true,
															},
															names.AttrName: schema.StringAttribute{
																Required: true,
															},
															names.AttrWeight: schema.Int64Attribute{
																Required: true,
															},
														},
														Blocks: map[string]schema.Block{
															"configuration_bundle": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[configurationBundleReferenceModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeBetween(1, 1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"bundle_arn": schema.StringAttribute{
																			Required:   true,
																			CustomType: fwtypes.ARNType,
																		},
																		"bundle_version": schema.StringAttribute{
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
						"route_to_target": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[routeToTargetActionModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"static_route": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[staticRouteModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
											listvalidator.ExactlyOneOf(
												path.MatchRelative().AtParent().AtName("weighted_route"),
											),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"target_name": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.RegexMatches(regexache.MustCompile(`^([0-9a-zA-Z][-]?){1,100}$`), ""),
													},
												},
											},
										},
									},
									"weighted_route": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[weightedRouteModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"traffic_split": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[targetTrafficSplitEntryModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeBetween(2, 2),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrDescription: schema.StringAttribute{
																Optional: true,
															},
															"metadata": schema.MapAttribute{
																CustomType:  fwtypes.MapOfStringType,
																ElementType: types.StringType,
																Optional:    true,
															},
															names.AttrName: schema.StringAttribute{
																Required: true,
															},
															"target_name": schema.StringAttribute{
																Required: true,
																Validators: []validator.String{
																	stringvalidator.RegexMatches(regexache.MustCompile(`^([0-9a-zA-Z][-]?){1,100}$`), ""),
																},
															},
															names.AttrWeight: schema.Int64Attribute{
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
			names.AttrCondition: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[conditionModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(2),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"match_paths": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[matchPathsModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("match_principals"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"any_of": schema.ListAttribute{
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Required:    true,
									},
								},
							},
						},
						"match_principals": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[matchPrincipalsModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"any_of": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[matchPrincipalEntryModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtLeast(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"iam_principal": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[iamPrincipalModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrARN: schema.StringAttribute{
																Required:   true,
																CustomType: fwtypes.ARNType,
															},
															"operator": schema.StringAttribute{
																Optional:   true,
																Computed:   true,
																CustomType: fwtypes.StringEnumType[awstypes.PrincipalMatchOperator](),
																PlanModifiers: []planmodifier.String{
																	stringplanmodifier.UseStateForUnknown(),
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
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *gatewayRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data gatewayRuleResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	gatewayIdentifier := fwflex.StringValueFromFramework(ctx, data.GatewayIdentifier)
	var input bedrockagentcorecontrol.CreateGatewayRuleInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(create.UniqueId(ctx))

	out, err := conn.CreateGatewayRule(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, gatewayIdentifier)
		return
	}

	ruleID := aws.ToString(out.RuleId)

	rule, err := waitGatewayRuleCreated(ctx, conn, gatewayIdentifier, ruleID, r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		// Taint the resource so a follow-up plan can reconcile it.
		response.State.SetAttribute(ctx, path.Root("gateway_identifier"), gatewayIdentifier)
		response.State.SetAttribute(ctx, path.Root("rule_id"), ruleID)
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, rule, &data, fwflex.WithIgnoredFieldNames([]string{"GatewayArn"})))
	if response.Diagnostics.HasError() {
		return
	}
	data.GatewayArn = fwflex.StringToFramework(ctx, rule.GatewayArn)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *gatewayRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data gatewayRuleResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	gatewayIdentifier, ruleID := fwflex.StringValueFromFramework(ctx, data.GatewayIdentifier), fwflex.StringValueFromFramework(ctx, data.RuleID)
	out, err := findGatewayRuleByTwoPartKey(ctx, conn, gatewayIdentifier, ruleID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data, fwflex.WithIgnoredFieldNames([]string{"GatewayArn"})))
	if response.Diagnostics.HasError() {
		return
	}
	data.GatewayArn = fwflex.StringToFramework(ctx, out.GatewayArn)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *gatewayRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state gatewayRuleResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		gatewayIdentifier, ruleID := fwflex.StringValueFromFramework(ctx, plan.GatewayIdentifier), fwflex.StringValueFromFramework(ctx, plan.RuleID)
		var input bedrockagentcorecontrol.UpdateGatewayRuleInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if response.Diagnostics.HasError() {
			return
		}

		if _, err := conn.UpdateGatewayRule(ctx, &input); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleID)
			return
		}

		rule, err := waitGatewayRuleUpdated(ctx, conn, gatewayIdentifier, ruleID, r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleID)
			return
		}

		// Re-hydrate computed fields (status, system, gateway_arn) from the
		// authoritative Get output so nothing is left Unknown after apply.
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, rule, &plan, fwflex.WithIgnoredFieldNames([]string{"GatewayArn"})))
		if response.Diagnostics.HasError() {
			return
		}
		plan.GatewayArn = fwflex.StringToFramework(ctx, rule.GatewayArn)
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *gatewayRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data gatewayRuleResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	gatewayIdentifier, ruleID := fwflex.StringValueFromFramework(ctx, data.GatewayIdentifier), fwflex.StringValueFromFramework(ctx, data.RuleID)
	input := bedrockagentcorecontrol.DeleteGatewayRuleInput{
		GatewayIdentifier: aws.String(gatewayIdentifier),
		RuleId:            aws.String(ruleID),
	}
	if _, err := conn.DeleteGatewayRule(ctx, &input); err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleID)
		return
	}

	if _, err := waitGatewayRuleDeleted(ctx, conn, gatewayIdentifier, ruleID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleID)
		return
	}
}

func (r *gatewayRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	parts := strings.Split(request.ID, ",")

	if len(parts) != 2 {
		smerr.AddError(ctx, &response.Diagnostics, fmt.Errorf(`unexpected format for import ID (%s), use: "GatewayIdentifier,RuleId"`, request.ID))
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root("gateway_identifier"), parts[0]))
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root("rule_id"), parts[1]))
}

func waitGatewayRuleCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, gatewayIdentifier, ruleID string, timeout time.Duration) (*bedrockagentcorecontrol.GetGatewayRuleOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.GatewayRuleStatusCreating),
		Target:                    enum.Slice(awstypes.GatewayRuleStatusActive),
		Refresh:                   statusGatewayRule(conn, gatewayIdentifier, ruleID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetGatewayRuleOutput); ok {
		return out, smarterr.NewError(err)
	}
	return nil, smarterr.NewError(err)
}

func waitGatewayRuleUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, gatewayIdentifier, ruleID string, timeout time.Duration) (*bedrockagentcorecontrol.GetGatewayRuleOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.GatewayRuleStatusUpdating),
		Target:                    enum.Slice(awstypes.GatewayRuleStatusActive),
		Refresh:                   statusGatewayRule(conn, gatewayIdentifier, ruleID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetGatewayRuleOutput); ok {
		return out, smarterr.NewError(err)
	}
	return nil, smarterr.NewError(err)
}

func waitGatewayRuleDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, gatewayIdentifier, ruleID string, timeout time.Duration) (*bedrockagentcorecontrol.GetGatewayRuleOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.GatewayRuleStatusDeleting, awstypes.GatewayRuleStatusActive),
		Target:  []string{},
		Refresh: statusGatewayRule(conn, gatewayIdentifier, ruleID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetGatewayRuleOutput); ok {
		return out, smarterr.NewError(err)
	}
	return nil, smarterr.NewError(err)
}

func statusGatewayRule(conn *bedrockagentcorecontrol.Client, gatewayIdentifier, ruleID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findGatewayRuleByTwoPartKey(ctx, conn, gatewayIdentifier, ruleID)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}
		if out == nil {
			return nil, "", nil
		}
		return out, string(out.Status), nil
	}
}

func findGatewayRuleByTwoPartKey(ctx context.Context, conn *bedrockagentcorecontrol.Client, gatewayIdentifier, ruleID string) (*bedrockagentcorecontrol.GetGatewayRuleOutput, error) {
	input := bedrockagentcorecontrol.GetGatewayRuleInput{
		GatewayIdentifier: aws.String(gatewayIdentifier),
		RuleId:            aws.String(ruleID),
	}

	out, err := conn.GetGatewayRule(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}
	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}
	return out, nil
}

// Models.

type gatewayRuleResourceModel struct {
	framework.WithRegionModel
	Actions           fwtypes.ListNestedObjectValueOf[actionModel]             `tfsdk:"action"`
	Conditions        fwtypes.ListNestedObjectValueOf[conditionModel]          `tfsdk:"condition"`
	Description       types.String                                             `tfsdk:"description"`
	GatewayArn        types.String                                             `tfsdk:"gateway_arn"`
	GatewayIdentifier types.String                                             `tfsdk:"gateway_identifier"`
	Priority          types.Int64                                              `tfsdk:"priority"`
	RuleID            types.String                                             `tfsdk:"rule_id"`
	Status            fwtypes.StringEnum[awstypes.GatewayRuleStatus]           `tfsdk:"status"`
	System            fwtypes.ListNestedObjectValueOf[systemManagedBlockModel] `tfsdk:"system"`
	Timeouts          timeouts.Value                                           `tfsdk:"timeouts"`
}

// Action union.

type actionModel struct {
	ConfigurationBundle fwtypes.ListNestedObjectValueOf[configurationBundleActionModel] `tfsdk:"configuration_bundle"`
	RouteToTarget       fwtypes.ListNestedObjectValueOf[routeToTargetActionModel]       `tfsdk:"route_to_target"`
}

var (
	_ fwflex.Expander  = actionModel{}
	_ fwflex.Flattener = &actionModel{}
)

func (m *actionModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	// Every unset sibling variant must be explicitly nulled — a zero-value
	// ListNestedObjectValueOf serializes with a nil element type and panics
	// downstream. Do it once here and overwrite the active variant below.
	m.ConfigurationBundle = fwtypes.NewListNestedObjectValueOfNull[configurationBundleActionModel](ctx)
	m.RouteToTarget = fwtypes.NewListNestedObjectValueOfNull[routeToTargetActionModel](ctx)

	// AutoFlex dispatches either the pointer form (when iterating a slice of
	// interfaces like []Action) or the value form (when a caller passes the
	// wrapped .Value of another union member) — accept both by unwrapping.
	switch t := v.(type) {
	case *awstypes.ActionMemberConfigurationBundle:
		return m.Flatten(ctx, *t)
	case awstypes.ActionMemberConfigurationBundle:
		var data configurationBundleActionModel
		smerr.AddEnrich(ctx, &diags, data.Flatten(ctx, t.Value))
		if diags.HasError() {
			return diags
		}
		m.ConfigurationBundle = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	case *awstypes.ActionMemberRouteToTarget:
		return m.Flatten(ctx, *t)
	case awstypes.ActionMemberRouteToTarget:
		var data routeToTargetActionModel
		smerr.AddEnrich(ctx, &diags, data.Flatten(ctx, t.Value))
		if diags.HasError() {
			return diags
		}
		m.RouteToTarget = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("action flatten: %T", v))
	}
	return diags
}

func (m actionModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.ConfigurationBundle.IsNull():
		data, d := m.ConfigurationBundle.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		inner, d := data.Expand(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		if inner == nil {
			return nil, diags
		}
		return &awstypes.ActionMemberConfigurationBundle{Value: castConfigurationBundleAction(inner)}, diags

	case !m.RouteToTarget.IsNull():
		data, d := m.RouteToTarget.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		inner, d := data.Expand(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		if inner == nil {
			return nil, diags
		}
		return &awstypes.ActionMemberRouteToTarget{Value: castRouteToTargetAction(inner)}, diags
	}
	return nil, diags
}

// castConfigurationBundleAction unwraps the union member returned by
// configurationBundleActionModel.Expand back into the ConfigurationBundleAction
// interface value expected by the parent Action union member.
func castConfigurationBundleAction(v any) awstypes.ConfigurationBundleAction {
	switch t := v.(type) {
	case *awstypes.ConfigurationBundleActionMemberStaticOverride:
		return t
	case *awstypes.ConfigurationBundleActionMemberWeightedOverride:
		return t
	}
	return nil
}

func castRouteToTargetAction(v any) awstypes.RouteToTargetAction {
	switch t := v.(type) {
	case *awstypes.RouteToTargetActionMemberStaticRoute:
		return t
	case *awstypes.RouteToTargetActionMemberWeightedRoute:
		return t
	}
	return nil
}

// ConfigurationBundleAction union.

type configurationBundleActionModel struct {
	StaticOverride   fwtypes.ListNestedObjectValueOf[staticOverrideModel]   `tfsdk:"static_override"`
	WeightedOverride fwtypes.ListNestedObjectValueOf[weightedOverrideModel] `tfsdk:"weighted_override"`
}

var (
	_ fwflex.Expander  = configurationBundleActionModel{}
	_ fwflex.Flattener = &configurationBundleActionModel{}
)

func (m *configurationBundleActionModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	m.StaticOverride = fwtypes.NewListNestedObjectValueOfNull[staticOverrideModel](ctx)
	m.WeightedOverride = fwtypes.NewListNestedObjectValueOfNull[weightedOverrideModel](ctx)
	switch t := v.(type) {
	case *awstypes.ConfigurationBundleActionMemberStaticOverride:
		return m.Flatten(ctx, *t)
	case awstypes.ConfigurationBundleActionMemberStaticOverride:
		var data staticOverrideModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.StaticOverride = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	case *awstypes.ConfigurationBundleActionMemberWeightedOverride:
		return m.Flatten(ctx, *t)
	case awstypes.ConfigurationBundleActionMemberWeightedOverride:
		var data weightedOverrideModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.WeightedOverride = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("configuration_bundle flatten: %T", v))
	}
	return diags
}

func (m configurationBundleActionModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.StaticOverride.IsNull():
		data, d := m.StaticOverride.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ConfigurationBundleActionMemberStaticOverride
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.WeightedOverride.IsNull():
		data, d := m.WeightedOverride.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ConfigurationBundleActionMemberWeightedOverride
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

// RouteToTargetAction union.

type routeToTargetActionModel struct {
	StaticRoute   fwtypes.ListNestedObjectValueOf[staticRouteModel]   `tfsdk:"static_route"`
	WeightedRoute fwtypes.ListNestedObjectValueOf[weightedRouteModel] `tfsdk:"weighted_route"`
}

var (
	_ fwflex.Expander  = routeToTargetActionModel{}
	_ fwflex.Flattener = &routeToTargetActionModel{}
)

func (m *routeToTargetActionModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	m.StaticRoute = fwtypes.NewListNestedObjectValueOfNull[staticRouteModel](ctx)
	m.WeightedRoute = fwtypes.NewListNestedObjectValueOfNull[weightedRouteModel](ctx)
	switch t := v.(type) {
	case *awstypes.RouteToTargetActionMemberStaticRoute:
		return m.Flatten(ctx, *t)
	case awstypes.RouteToTargetActionMemberStaticRoute:
		var data staticRouteModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.StaticRoute = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	case *awstypes.RouteToTargetActionMemberWeightedRoute:
		return m.Flatten(ctx, *t)
	case awstypes.RouteToTargetActionMemberWeightedRoute:
		var data weightedRouteModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.WeightedRoute = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("route_to_target flatten: %T", v))
	}
	return diags
}

func (m routeToTargetActionModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.StaticRoute.IsNull():
		data, d := m.StaticRoute.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.RouteToTargetActionMemberStaticRoute
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.WeightedRoute.IsNull():
		data, d := m.WeightedRoute.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.RouteToTargetActionMemberWeightedRoute
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

// Condition union.

type conditionModel struct {
	MatchPaths      fwtypes.ListNestedObjectValueOf[matchPathsModel]      `tfsdk:"match_paths"`
	MatchPrincipals fwtypes.ListNestedObjectValueOf[matchPrincipalsModel] `tfsdk:"match_principals"`
}

var (
	_ fwflex.Expander  = conditionModel{}
	_ fwflex.Flattener = &conditionModel{}
)

func (m *conditionModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	m.MatchPaths = fwtypes.NewListNestedObjectValueOfNull[matchPathsModel](ctx)
	m.MatchPrincipals = fwtypes.NewListNestedObjectValueOfNull[matchPrincipalsModel](ctx)
	switch t := v.(type) {
	case *awstypes.ConditionMemberMatchPaths:
		return m.Flatten(ctx, *t)
	case awstypes.ConditionMemberMatchPaths:
		var data matchPathsModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.MatchPaths = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	case *awstypes.ConditionMemberMatchPrincipals:
		return m.Flatten(ctx, *t)
	case awstypes.ConditionMemberMatchPrincipals:
		var data matchPrincipalsModel
		smerr.AddEnrich(ctx, &diags, data.Flatten(ctx, t.Value))
		if diags.HasError() {
			return diags
		}
		m.MatchPrincipals = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("condition flatten: %T", v))
	}
	return diags
}

func (m conditionModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.MatchPaths.IsNull():
		data, d := m.MatchPaths.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ConditionMemberMatchPaths
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.MatchPrincipals.IsNull():
		data, d := m.MatchPrincipals.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		inner, d := data.Expand(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		mp, ok := inner.(*awstypes.MatchPrincipals)
		if !ok {
			return nil, diags
		}
		return &awstypes.ConditionMemberMatchPrincipals{Value: *mp}, diags
	}
	return nil, diags
}

// MatchPrincipals — needs a custom Expand to translate the anyOf list of
// MatchPrincipalEntry union values, since AutoFlex can't build interface slices.

type matchPrincipalsModel struct {
	AnyOf fwtypes.ListNestedObjectValueOf[matchPrincipalEntryModel] `tfsdk:"any_of"`
}

var (
	_ fwflex.Expander  = matchPrincipalsModel{}
	_ fwflex.Flattener = &matchPrincipalsModel{}
)

func (m *matchPrincipalsModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	var mp awstypes.MatchPrincipals
	switch t := v.(type) {
	case awstypes.MatchPrincipals:
		mp = t
	case *awstypes.MatchPrincipals:
		if t == nil {
			return diags
		}
		mp = *t
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("match_principals flatten: %T", v))
		return diags
	}
	entries := make([]*matchPrincipalEntryModel, 0, len(mp.AnyOf))
	for _, e := range mp.AnyOf {
		var em matchPrincipalEntryModel
		smerr.AddEnrich(ctx, &diags, em.Flatten(ctx, e))
		if diags.HasError() {
			return diags
		}
		entries = append(entries, &em)
	}
	m.AnyOf = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, entries)
	return diags
}

func (m matchPrincipalsModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	entries, d := m.AnyOf.ToSlice(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return nil, diags
	}
	anyOf := make([]awstypes.MatchPrincipalEntry, 0, len(entries))
	for _, e := range entries {
		inner, d := e.Expand(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		if mp, ok := inner.(awstypes.MatchPrincipalEntry); ok {
			anyOf = append(anyOf, mp)
		}
	}
	return &awstypes.MatchPrincipals{AnyOf: anyOf}, diags
}

// MatchPrincipalEntry union.

type matchPrincipalEntryModel struct {
	IamPrincipal fwtypes.ListNestedObjectValueOf[iamPrincipalModel] `tfsdk:"iam_principal"`
}

var (
	_ fwflex.Expander  = matchPrincipalEntryModel{}
	_ fwflex.Flattener = &matchPrincipalEntryModel{}
)

func (m *matchPrincipalEntryModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	m.IamPrincipal = fwtypes.NewListNestedObjectValueOfNull[iamPrincipalModel](ctx)
	switch t := v.(type) {
	case *awstypes.MatchPrincipalEntryMemberIamPrincipal:
		return m.Flatten(ctx, *t)
	case awstypes.MatchPrincipalEntryMemberIamPrincipal:
		var data iamPrincipalModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.IamPrincipal = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("match_principal_entry flatten: %T", v))
	}
	return diags
}

func (m matchPrincipalEntryModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.IamPrincipal.IsNull():
		data, d := m.IamPrincipal.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.MatchPrincipalEntryMemberIamPrincipal
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

// Leaf models. Field names line up with the SDK Go struct fields so that AutoFlex
// handles them without further customization.

type matchPathsModel struct {
	AnyOf fwtypes.ListOfString `tfsdk:"any_of"`
}

type iamPrincipalModel struct {
	ARN      fwtypes.ARN                                         `tfsdk:"arn"`
	Operator fwtypes.StringEnum[awstypes.PrincipalMatchOperator] `tfsdk:"operator"`
}

type staticOverrideModel struct {
	BundleArn     fwtypes.ARN  `tfsdk:"bundle_arn"`
	BundleVersion types.String `tfsdk:"bundle_version"`
}

type weightedOverrideModel struct {
	TrafficSplit fwtypes.ListNestedObjectValueOf[trafficSplitEntryModel] `tfsdk:"traffic_split"`
}

type trafficSplitEntryModel struct {
	ConfigurationBundle fwtypes.ListNestedObjectValueOf[configurationBundleReferenceModel] `tfsdk:"configuration_bundle"`
	Description         types.String                                                       `tfsdk:"description"`
	Metadata            fwtypes.MapOfString                                                `tfsdk:"metadata"`
	Name                types.String                                                       `tfsdk:"name"`
	Weight              types.Int64                                                        `tfsdk:"weight"`
}

type configurationBundleReferenceModel struct {
	BundleArn     fwtypes.ARN  `tfsdk:"bundle_arn"`
	BundleVersion types.String `tfsdk:"bundle_version"`
}

type staticRouteModel struct {
	TargetName types.String `tfsdk:"target_name"`
}

type weightedRouteModel struct {
	TrafficSplit fwtypes.ListNestedObjectValueOf[targetTrafficSplitEntryModel] `tfsdk:"traffic_split"`
}

type targetTrafficSplitEntryModel struct {
	Description types.String        `tfsdk:"description"`
	Metadata    fwtypes.MapOfString `tfsdk:"metadata"`
	Name        types.String        `tfsdk:"name"`
	TargetName  types.String        `tfsdk:"target_name"`
	Weight      types.Int64         `tfsdk:"weight"`
}

type systemManagedBlockModel struct {
	ManagedBy types.String `tfsdk:"managed_by"`
}
