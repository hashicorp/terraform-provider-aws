// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	webACLRuleGroupAssociationResourceIDPartCount = 3
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_wafv2_web_acl_rule_group_association", name="Web ACL Rule Group Association")
func newResourceWebACLRuleGroupAssociation(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceWebACLRuleGroupAssociation{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameWebACLRuleGroupAssociation = "Web ACL Rule Group Association"
)

type resourceWebACLRuleGroupAssociation struct {
	framework.ResourceWithModel[resourceWebACLRuleGroupAssociationModel]
	framework.WithTimeouts
}

func (r *resourceWebACLRuleGroupAssociation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"rule_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
				},
				Description: "Name of the rule to create in the Web ACL that references the rule group.",
			},
			names.AttrPriority: schema.Int32Attribute{
				Required: true,
				Validators: []validator.Int32{
					int32validator.AtLeast(0),
				},
				Description: "Priority of the rule within the Web ACL.",
			},
			"rule_group_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					fwvalidators.ARN(),
				},
				Description: "ARN of the Rule Group to associate with the Web ACL.",
			},
			"web_acl_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					fwvalidators.ARN(),
				},
				Description: "ARN of the Web ACL to associate the Rule Group with.",
			},
			"override_action": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("none", "count"),
				},
				Description: "Override action for the rule group. Valid values are 'none' and 'count'. Defaults to 'none'.",
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"rule_action_override": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[ruleActionOverrideModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(100),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 128),
							},
							Description: "Name of the rule to override.",
						},
					},
					Blocks: map[string]schema.Block{
						"action_to_use": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[actionToUseModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"allow": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[allowActionModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"custom_request_handling": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[customRequestHandlingModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Blocks: map[string]schema.Block{
															"insert_header": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[insertHeaderModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtLeast(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		names.AttrName: schema.StringAttribute{
																			Required: true,
																			Validators: []validator.String{
																				stringvalidator.LengthBetween(1, 64),
																			},
																		},
																		names.AttrValue: schema.StringAttribute{
																			Required: true,
																			Validators: []validator.String{
																				stringvalidator.LengthBetween(1, 255),
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
									"block": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[blockActionModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"custom_response": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[customResponseModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"custom_response_body_key": schema.StringAttribute{
																Optional: true,
																Validators: []validator.String{
																	stringvalidator.LengthBetween(1, 128),
																},
															},
															"response_code": schema.Int32Attribute{
																Required: true,
																Validators: []validator.Int32{
																	int32validator.Between(200, 600),
																},
															},
														},
														Blocks: map[string]schema.Block{
															"response_header": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[responseHeaderModel](ctx),
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		names.AttrName: schema.StringAttribute{
																			Required: true,
																			Validators: []validator.String{
																				stringvalidator.LengthBetween(1, 64),
																			},
																		},
																		names.AttrValue: schema.StringAttribute{
																			Required: true,
																			Validators: []validator.String{
																				stringvalidator.LengthBetween(1, 255),
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
									"captcha": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[captchaActionModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"custom_request_handling": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[customRequestHandlingModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Blocks: map[string]schema.Block{
															"insert_header": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[insertHeaderModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtLeast(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		names.AttrName: schema.StringAttribute{
																			Required: true,
																			Validators: []validator.String{
																				stringvalidator.LengthBetween(1, 64),
																			},
																		},
																		names.AttrValue: schema.StringAttribute{
																			Required: true,
																			Validators: []validator.String{
																				stringvalidator.LengthBetween(1, 255),
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
									"challenge": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[challengeActionModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"custom_request_handling": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[customRequestHandlingModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Blocks: map[string]schema.Block{
															"insert_header": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[insertHeaderModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtLeast(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		names.AttrName: schema.StringAttribute{
																			Required: true,
																			Validators: []validator.String{
																				stringvalidator.LengthBetween(1, 64),
																			},
																		},
																		names.AttrValue: schema.StringAttribute{
																			Required: true,
																			Validators: []validator.String{
																				stringvalidator.LengthBetween(1, 255),
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
									"count": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[countActionModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"custom_request_handling": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[customRequestHandlingModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Blocks: map[string]schema.Block{
															"insert_header": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[insertHeaderModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtLeast(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		names.AttrName: schema.StringAttribute{
																			Required: true,
																			Validators: []validator.String{
																				stringvalidator.LengthBetween(1, 64),
																			},
																		},
																		names.AttrValue: schema.StringAttribute{
																			Required: true,
																			Validators: []validator.String{
																				stringvalidator.LengthBetween(1, 255),
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
							Description: "Action to use in place of the rule action.",
						},
					},
				},
				Description: "Action settings to use in place of rule actions configured inside the rule group. You can specify up to 100 overrides.",
			},
		},
		Description: "Associates a WAFv2 Rule Group with a Web ACL by adding a rule that references the Rule Group.",
	}
}

func (r *resourceWebACLRuleGroupAssociation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().WAFV2Client(ctx)

	var plan resourceWebACLRuleGroupAssociationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse Web ACL ARN to get ID, name, and scope
	webACLID, webACLName, webACLScope, err := parseWebACLARN(plan.WebACLARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionCreating, ResNameWebACLRuleGroupAssociation, plan.RuleName.String(), err),
			err.Error(),
		)
		return
	}

	// Get current Web ACL configuration
	webACL, err := findWebACLByThreePartKey(ctx, conn, webACLID, webACLName, webACLScope)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionCreating, ResNameWebACLRuleGroupAssociation, plan.RuleName.String(), err),
			err.Error(),
		)
		return
	}

	// Check if rule with same priority or name already exists
	for _, rule := range webACL.WebACL.Rules {
		if rule.Priority == plan.Priority.ValueInt32() {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.WAFV2, create.ErrActionCreating, ResNameWebACLRuleGroupAssociation, plan.RuleName.String(), nil),
				fmt.Sprintf("Rule with priority %d already exists in Web ACL", plan.Priority.ValueInt32()),
			)
			return
		}
		if aws.ToString(rule.Name) == plan.RuleName.ValueString() {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.WAFV2, create.ErrActionCreating, ResNameWebACLRuleGroupAssociation, plan.RuleName.String(), nil),
				fmt.Sprintf("Rule with name %s already exists in Web ACL", plan.RuleName.ValueString()),
			)
			return
		}
	}

	// Create new rule with rule group reference statement
	ruleGroupRefStatement := &awstypes.RuleGroupReferenceStatement{
		ARN: plan.RuleGroupARN.ValueStringPointer(),
	}

	// Add rule action overrides if specified
	var ruleActionOverrides []awstypes.RuleActionOverride
	if !plan.RuleActionOverride.IsNull() && !plan.RuleActionOverride.IsUnknown() {
		resp.Diagnostics.Append(fwflex.Expand(ctx, plan.RuleActionOverride, &ruleActionOverrides)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	ruleGroupRefStatement.RuleActionOverrides = ruleActionOverrides

	newRule := awstypes.Rule{
		Name:     plan.RuleName.ValueStringPointer(),
		Priority: plan.Priority.ValueInt32(),
		Statement: &awstypes.Statement{
			RuleGroupReferenceStatement: ruleGroupRefStatement,
		},
		VisibilityConfig: &awstypes.VisibilityConfig{
			SampledRequestsEnabled:   true,
			CloudWatchMetricsEnabled: true,
			MetricName:               plan.RuleName.ValueStringPointer(),
		},
	}

	// Set override action
	overrideAction := plan.OverrideAction.ValueString()
	if overrideAction == "" {
		overrideAction = "none"
		plan.OverrideAction = types.StringValue("none") // Set the default in the plan
	}

	switch overrideAction {
	case "none":
		newRule.OverrideAction = &awstypes.OverrideAction{
			None: &awstypes.NoneAction{},
		}
	case "count":
		newRule.OverrideAction = &awstypes.OverrideAction{
			Count: &awstypes.CountAction{},
		}
	}

	// Add the new rule to existing rules
	webACL.WebACL.Rules = append(webACL.WebACL.Rules, newRule)

	// Update the Web ACL
	updateInput := &wafv2.UpdateWebACLInput{
		Id:                   aws.String(webACLID),
		Name:                 aws.String(webACLName),
		Scope:                awstypes.Scope(webACLScope),
		DefaultAction:        webACL.WebACL.DefaultAction,
		Rules:                webACL.WebACL.Rules,
		VisibilityConfig:     webACL.WebACL.VisibilityConfig,
		LockToken:            webACL.LockToken,
		AssociationConfig:    webACL.WebACL.AssociationConfig,
		CaptchaConfig:        webACL.WebACL.CaptchaConfig,
		ChallengeConfig:      webACL.WebACL.ChallengeConfig,
		CustomResponseBodies: webACL.WebACL.CustomResponseBodies,
		TokenDomains:         webACL.WebACL.TokenDomains,
	}

	// Only set description if it's not empty
	if webACL.WebACL.Description != nil && aws.ToString(webACL.WebACL.Description) != "" {
		updateInput.Description = webACL.WebACL.Description
	}

	const timeout = 5 * time.Minute
	_, err = tfresource.RetryWhenIsA[*awstypes.WAFUnavailableEntityException](ctx, timeout, func() (any, error) {
		return conn.UpdateWebACL(ctx, updateInput)
	})

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionCreating, ResNameWebACLRuleGroupAssociation, plan.RuleName.String(), err),
			err.Error(),
		)
		return
	}

	// Set the ID using the standard flex utility with comma separators
	// Format: webACLARN,ruleName,ruleGroupARN
	id, err := flex.FlattenResourceId([]string{
		plan.WebACLARN.ValueString(),
		plan.RuleName.ValueString(),
		plan.RuleGroupARN.ValueString(),
	}, webACLRuleGroupAssociationResourceIDPartCount, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating resource ID",
			fmt.Sprintf("Could not create resource ID: %s", err),
		)
		return
	}
	plan.ID = types.StringValue(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceWebACLRuleGroupAssociation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceWebACLRuleGroupAssociationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WAFV2Client(ctx)

	// Parse the ID using the standard flex utility
	// Format: webACLARN,ruleName,ruleGroupARN
	parts, err := flex.ExpandResourceId(state.ID.ValueString(), webACLRuleGroupAssociationResourceIDPartCount, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Resource ID",
			fmt.Sprintf("Could not parse resource ID: %s", err),
		)
		return
	}

	webACLARN := parts[0]
	ruleName := parts[1]
	ruleGroupARN := parts[2]

	// Parse Web ACL ARN to get ID, name, and scope
	webACLID, webACLName, webACLScope, err := parseWebACLARN(webACLARN)
	if err != nil {
		resp.Diagnostics.AddError(
			"Reading WAFv2 Web ACL Rule Group Association",
			fmt.Sprintf("Error parsing Web ACL ARN: %s", err),
		)
		return
	}

	// Get the Web ACL and check if the rule group is associated
	webACL, err := findWebACLByThreePartKey(ctx, conn, webACLID, webACLName, webACLScope)
	if err != nil {
		if retry.NotFound(err) {
			resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Reading WAFv2 Web ACL Rule Group Association",
			fmt.Sprintf("Error reading Web ACL: %s", err),
		)
		return
	}

	// Find the rule group in the Web ACL rules
	found := false
	for _, rule := range webACL.WebACL.Rules {
		if aws.ToString(rule.Name) == ruleName &&
			rule.Statement != nil &&
			rule.Statement.RuleGroupReferenceStatement != nil &&
			aws.ToString(rule.Statement.RuleGroupReferenceStatement.ARN) == ruleGroupARN {
			found = true
			state.Priority = types.Int32Value(rule.Priority)

			// Determine override action
			overrideAction := "none"
			if rule.OverrideAction != nil {
				if rule.OverrideAction.Count != nil {
					overrideAction = "count"
				} else if rule.OverrideAction.None != nil {
					overrideAction = "none"
				}
			}
			state.OverrideAction = types.StringValue(overrideAction)

			// Handle rule action overrides
			if rule.Statement.RuleGroupReferenceStatement.RuleActionOverrides != nil {
				var ruleActionOverrides []*ruleActionOverrideModel
				for _, override := range rule.Statement.RuleGroupReferenceStatement.RuleActionOverrides {
					overrideModel := &ruleActionOverrideModel{
						Name: types.StringValue(aws.ToString(override.Name)),
					}

					// Convert the action to use
					if override.ActionToUse != nil {
						actionToUse := &actionToUseModel{
							Allow:     fwtypes.NewListNestedObjectValueOfNull[allowActionModel](ctx),
							Block:     fwtypes.NewListNestedObjectValueOfNull[blockActionModel](ctx),
							Captcha:   fwtypes.NewListNestedObjectValueOfNull[captchaActionModel](ctx),
							Challenge: fwtypes.NewListNestedObjectValueOfNull[challengeActionModel](ctx),
							Count:     fwtypes.NewListNestedObjectValueOfNull[countActionModel](ctx),
						}
						if override.ActionToUse.Allow != nil {
							allowModel := &allowActionModel{}
							if override.ActionToUse.Allow.CustomRequestHandling != nil {
								customHandlingModel := &customRequestHandlingModel{}
								var insertHeaders []*insertHeaderModel
								for _, header := range override.ActionToUse.Allow.CustomRequestHandling.InsertHeaders {
									insertHeaders = append(insertHeaders, &insertHeaderModel{
										Name:  types.StringValue(aws.ToString(header.Name)),
										Value: types.StringValue(aws.ToString(header.Value)),
									})
								}
								customHandlingModel.InsertHeader = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, insertHeaders)
								allowModel.CustomRequestHandling = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*customRequestHandlingModel{customHandlingModel})
							}
							actionToUse.Allow = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*allowActionModel{allowModel})
						} else if override.ActionToUse.Block != nil {
							blockModel := &blockActionModel{}
							if override.ActionToUse.Block.CustomResponse != nil {
								customResponse := &customResponseModel{
									ResponseCode: types.Int32Value(aws.ToInt32(override.ActionToUse.Block.CustomResponse.ResponseCode)),
								}
								if override.ActionToUse.Block.CustomResponse.CustomResponseBodyKey != nil {
									customResponse.CustomResponseBodyKey = types.StringValue(aws.ToString(override.ActionToUse.Block.CustomResponse.CustomResponseBodyKey))
								}
								var responseHeaders []*responseHeaderModel
								for _, header := range override.ActionToUse.Block.CustomResponse.ResponseHeaders {
									responseHeaders = append(responseHeaders, &responseHeaderModel{
										Name:  types.StringValue(aws.ToString(header.Name)),
										Value: types.StringValue(aws.ToString(header.Value)),
									})
								}
								customResponse.ResponseHeader = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, responseHeaders)
								blockModel.CustomResponse = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*customResponseModel{customResponse})
							}
							actionToUse.Block = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*blockActionModel{blockModel})
						} else if override.ActionToUse.Captcha != nil {
							captchaModel := &captchaActionModel{}
							if override.ActionToUse.Captcha.CustomRequestHandling != nil {
								customHandlingModel := &customRequestHandlingModel{}
								var insertHeaders []*insertHeaderModel
								for _, header := range override.ActionToUse.Captcha.CustomRequestHandling.InsertHeaders {
									insertHeaders = append(insertHeaders, &insertHeaderModel{
										Name:  types.StringValue(aws.ToString(header.Name)),
										Value: types.StringValue(aws.ToString(header.Value)),
									})
								}
								customHandlingModel.InsertHeader = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, insertHeaders)
								captchaModel.CustomRequestHandling = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*customRequestHandlingModel{customHandlingModel})
							}
							actionToUse.Captcha = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*captchaActionModel{captchaModel})
						} else if override.ActionToUse.Challenge != nil {
							challengeModel := &challengeActionModel{}
							if override.ActionToUse.Challenge.CustomRequestHandling != nil {
								customHandlingModel := &customRequestHandlingModel{}
								var insertHeaders []*insertHeaderModel
								for _, header := range override.ActionToUse.Challenge.CustomRequestHandling.InsertHeaders {
									insertHeaders = append(insertHeaders, &insertHeaderModel{
										Name:  types.StringValue(aws.ToString(header.Name)),
										Value: types.StringValue(aws.ToString(header.Value)),
									})
								}
								customHandlingModel.InsertHeader = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, insertHeaders)
								challengeModel.CustomRequestHandling = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*customRequestHandlingModel{customHandlingModel})
							}
							actionToUse.Challenge = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*challengeActionModel{challengeModel})
						} else if override.ActionToUse.Count != nil {
							countModel := &countActionModel{}
							if override.ActionToUse.Count.CustomRequestHandling != nil {
								customHandlingModel := &customRequestHandlingModel{}
								var insertHeaders []*insertHeaderModel
								for _, header := range override.ActionToUse.Count.CustomRequestHandling.InsertHeaders {
									insertHeaders = append(insertHeaders, &insertHeaderModel{
										Name:  types.StringValue(aws.ToString(header.Name)),
										Value: types.StringValue(aws.ToString(header.Value)),
									})
								}
								customHandlingModel.InsertHeader = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, insertHeaders)
								countModel.CustomRequestHandling = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*customRequestHandlingModel{customHandlingModel})
							}
							actionToUse.Count = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*countActionModel{countModel})
						}

						overrideModel.ActionToUse = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*actionToUseModel{actionToUse})
					} else {
						// If ActionToUse is nil, set it to null
						overrideModel.ActionToUse = fwtypes.NewListNestedObjectValueOfNull[actionToUseModel](ctx)
					}
					ruleActionOverrides = append(ruleActionOverrides, overrideModel)
				}
				state.RuleActionOverride = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, ruleActionOverrides)
			} else {
				state.RuleActionOverride = fwtypes.NewListNestedObjectValueOfNull[ruleActionOverrideModel](ctx)
			}
			break
		}
	}

	if !found {
		resp.Diagnostics.AddWarning(
			"Rule Group Association Not Found",
			"Rule group association was not found in Web ACL, removing from state",
		)
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state with current values
	state.WebACLARN = types.StringValue(webACLARN)
	state.RuleGroupARN = types.StringValue(ruleGroupARN)
	state.RuleName = types.StringValue(ruleName)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceWebACLRuleGroupAssociation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceWebACLRuleGroupAssociationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WAFV2Client(ctx)

	// Parse Web ACL ARN to get ID, name, and scope
	webACLARN := plan.WebACLARN.ValueString()
	webACLID, webACLName, webACLScope, err := parseWebACLARN(webACLARN)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionUpdating, ResNameWebACLRuleGroupAssociation, plan.RuleName.String(), err),
			err.Error(),
		)
		return
	}

	// Get current Web ACL configuration
	webACL, err := findWebACLByThreePartKey(ctx, conn, webACLID, webACLName, webACLScope)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionUpdating, ResNameWebACLRuleGroupAssociation, plan.RuleName.String(), err),
			err.Error(),
		)
		return
	}

	// Find the rule to update
	ruleName := plan.RuleName.ValueString()
	ruleFound := false
	for i, rule := range webACL.WebACL.Rules {
		if aws.ToString(rule.Name) == ruleName {
			ruleFound = true

			// Update the rule's priority
			webACL.WebACL.Rules[i].Priority = plan.Priority.ValueInt32()

			// Update override action
			overrideAction := plan.OverrideAction.ValueString()
			if overrideAction == "" {
				overrideAction = "none" // Default value
			}

			switch overrideAction {
			case "none":
				webACL.WebACL.Rules[i].OverrideAction = &awstypes.OverrideAction{
					None: &awstypes.NoneAction{},
				}
			case "count":
				webACL.WebACL.Rules[i].OverrideAction = &awstypes.OverrideAction{
					Count: &awstypes.CountAction{},
				}
			}

			// Update rule action overrides
			var overrides []awstypes.RuleActionOverride
			if !plan.RuleActionOverride.IsNull() && !plan.RuleActionOverride.IsUnknown() {
				resp.Diagnostics.Append(fwflex.Expand(ctx, plan.RuleActionOverride, &overrides)...)
				if resp.Diagnostics.HasError() {
					return
				}
			}

			// Update the rule group reference statement with new overrides
			if webACL.WebACL.Rules[i].Statement != nil && webACL.WebACL.Rules[i].Statement.RuleGroupReferenceStatement != nil {
				webACL.WebACL.Rules[i].Statement.RuleGroupReferenceStatement.RuleActionOverrides = overrides
			}

			break
		}
	}

	if !ruleFound {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionUpdating, ResNameWebACLRuleGroupAssociation, plan.RuleName.String(), nil),
			fmt.Sprintf("Rule %s not found in Web ACL", ruleName),
		)
		return
	}

	// Check for priority conflicts with other rules
	for _, rule := range webACL.WebACL.Rules {
		if aws.ToString(rule.Name) != ruleName && rule.Priority == plan.Priority.ValueInt32() {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.WAFV2, create.ErrActionUpdating, ResNameWebACLRuleGroupAssociation, plan.RuleName.String(), nil),
				fmt.Sprintf("Rule with priority %d already exists in Web ACL", plan.Priority.ValueInt32()),
			)
			return
		}
	}

	// Update the Web ACL with the modified rule
	updateInput := &wafv2.UpdateWebACLInput{
		Id:                   aws.String(webACLID),
		Name:                 aws.String(webACLName),
		Scope:                awstypes.Scope(webACLScope),
		DefaultAction:        webACL.WebACL.DefaultAction,
		Rules:                webACL.WebACL.Rules,
		VisibilityConfig:     webACL.WebACL.VisibilityConfig,
		LockToken:            webACL.LockToken,
		AssociationConfig:    webACL.WebACL.AssociationConfig,
		CaptchaConfig:        webACL.WebACL.CaptchaConfig,
		ChallengeConfig:      webACL.WebACL.ChallengeConfig,
		CustomResponseBodies: webACL.WebACL.CustomResponseBodies,
		TokenDomains:         webACL.WebACL.TokenDomains,
	}

	// Only set description if it's not empty
	if webACL.WebACL.Description != nil && aws.ToString(webACL.WebACL.Description) != "" {
		updateInput.Description = webACL.WebACL.Description
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err = tfresource.RetryWhenIsA[*awstypes.WAFUnavailableEntityException](ctx, updateTimeout, func() (any, error) {
		return conn.UpdateWebACL(ctx, updateInput)
	})

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionUpdating, ResNameWebACLRuleGroupAssociation, plan.RuleName.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceWebACLRuleGroupAssociation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceWebACLRuleGroupAssociationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WAFV2Client(ctx)

	// Parse the ID using the standard flex utility
	// Format: webACLARN,ruleName,ruleGroupARN
	parts, err := flex.ExpandResourceId(state.ID.ValueString(), webACLRuleGroupAssociationResourceIDPartCount, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Resource ID",
			fmt.Sprintf("Could not parse resource ID: %s", err),
		)
		return
	}

	webACLARN := parts[0]
	ruleName := parts[1]

	// Parse Web ACL ARN to get ID, name, and scope
	webACLID, webACLName, webACLScope, err := parseWebACLARN(webACLARN)
	if err != nil {
		resp.Diagnostics.AddError(
			"Deleting WAFv2 Web ACL Rule Group Association",
			fmt.Sprintf("Error parsing Web ACL ARN: %s", err),
		)
		return
	}

	// Get the Web ACL
	webACL, err := findWebACLByThreePartKey(ctx, conn, webACLID, webACLName, webACLScope)
	if err != nil {
		if retry.NotFound(err) {
			// Web ACL is already gone, nothing to do
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionDeleting, ResNameWebACLRuleGroupAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	// Filter out the rule we want to remove
	var updatedRules []awstypes.Rule
	ruleFound := false
	for _, rule := range webACL.WebACL.Rules {
		if aws.ToString(rule.Name) != ruleName {
			updatedRules = append(updatedRules, rule)
		} else {
			ruleFound = true
		}
	}

	if !ruleFound {
		// Rule is already gone, nothing to do
		return
	}

	// Update the Web ACL without the rule
	updateInput := &wafv2.UpdateWebACLInput{
		Id:                   aws.String(webACLID),
		Name:                 aws.String(webACLName),
		Scope:                awstypes.Scope(webACLScope),
		DefaultAction:        webACL.WebACL.DefaultAction,
		Rules:                updatedRules,
		VisibilityConfig:     webACL.WebACL.VisibilityConfig,
		LockToken:            webACL.LockToken,
		AssociationConfig:    webACL.WebACL.AssociationConfig,
		CaptchaConfig:        webACL.WebACL.CaptchaConfig,
		ChallengeConfig:      webACL.WebACL.ChallengeConfig,
		CustomResponseBodies: webACL.WebACL.CustomResponseBodies,
		TokenDomains:         webACL.WebACL.TokenDomains,
	}

	// Only set description if it's not empty
	if webACL.WebACL.Description != nil && aws.ToString(webACL.WebACL.Description) != "" {
		updateInput.Description = webACL.WebACL.Description
	}

	const timeout = 5 * time.Minute
	_, err = tfresource.RetryWhenIsA[*awstypes.WAFUnavailableEntityException](ctx, timeout, func() (any, error) {
		return conn.UpdateWebACL(ctx, updateInput)
	})

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionDeleting, ResNameWebACLRuleGroupAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceWebACLRuleGroupAssociation) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: webACLARN,ruleGroupARN,ruleName
	parts := strings.Split(req.ID, ",")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID should be in format 'webACLARN,ruleGroupARN,ruleName'",
		)
		return
	}

	webACLARN := parts[0]
	ruleGroupARN := parts[1]
	ruleName := parts[2]

	// Parse Web ACL ARN to get ID, name, and scope
	_, _, _, err := parseWebACLARN(webACLARN)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Web ACL ARN",
			fmt.Sprintf("Error parsing Web ACL ARN: %s", err),
		)
		return
	}

	// Set the ID using the standard flex utility with comma separators
	id, err := flex.FlattenResourceId([]string{webACLARN, ruleName, ruleGroupARN}, webACLRuleGroupAssociationResourceIDPartCount, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating resource ID",
			fmt.Sprintf("Could not create resource ID: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("web_acl_arn"), webACLARN)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("rule_group_arn"), ruleGroupARN)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("rule_name"), ruleName)...)
}

// parseWebACLARN extracts the Web ACL ID, name, and scope from the ARN
func parseWebACLARN(arn string) (id, name, scope string, err error) {
	// ARN format: arn:aws:wafv2:region:account-id:scope/webacl/name/id
	// or for CloudFront: arn:aws:wafv2:global:account-id:global/webacl/name/id
	parts := strings.Split(arn, ":")
	if len(parts) < 6 {
		return "", "", "", fmt.Errorf("invalid Web ACL ARN format: %s", arn)
	}

	resourceParts := strings.Split(parts[5], "/")
	if len(resourceParts) < 4 {
		return "", "", "", fmt.Errorf("invalid Web ACL ARN resource format: %s", parts[5])
	}

	// Validate that this is a webacl ARN
	if resourceParts[1] != "webacl" {
		return "", "", "", fmt.Errorf("invalid Web ACL ARN: expected webacl resource type, got %s", resourceParts[1])
	}

	// Determine scope
	scopeValue := "REGIONAL"
	if parts[3] == "global" || resourceParts[0] == "global" {
		scopeValue = "CLOUDFRONT"
	}

	// Extract name and ID
	nameIndex := len(resourceParts) - 2
	idIndex := len(resourceParts) - 1

	return resourceParts[idIndex], resourceParts[nameIndex], scopeValue, nil
}

type resourceWebACLRuleGroupAssociationModel struct {
	framework.WithRegionModel
	ID                 types.String                                             `tfsdk:"id"`
	RuleName           types.String                                             `tfsdk:"rule_name"`
	Priority           types.Int32                                              `tfsdk:"priority"`
	RuleGroupARN       types.String                                             `tfsdk:"rule_group_arn"`
	WebACLARN          types.String                                             `tfsdk:"web_acl_arn"`
	OverrideAction     types.String                                             `tfsdk:"override_action"`
	RuleActionOverride fwtypes.ListNestedObjectValueOf[ruleActionOverrideModel] `tfsdk:"rule_action_override"`
	Timeouts           timeouts.Value                                           `tfsdk:"timeouts"`
}

type ruleActionOverrideModel struct {
	Name        types.String                                      `tfsdk:"name"`
	ActionToUse fwtypes.ListNestedObjectValueOf[actionToUseModel] `tfsdk:"action_to_use"`
}

type actionToUseModel struct {
	Allow     fwtypes.ListNestedObjectValueOf[allowActionModel]     `tfsdk:"allow"`
	Block     fwtypes.ListNestedObjectValueOf[blockActionModel]     `tfsdk:"block"`
	Captcha   fwtypes.ListNestedObjectValueOf[captchaActionModel]   `tfsdk:"captcha"`
	Challenge fwtypes.ListNestedObjectValueOf[challengeActionModel] `tfsdk:"challenge"`
	Count     fwtypes.ListNestedObjectValueOf[countActionModel]     `tfsdk:"count"`
}

type allowActionModel struct {
	CustomRequestHandling fwtypes.ListNestedObjectValueOf[customRequestHandlingModel] `tfsdk:"custom_request_handling"`
}

type blockActionModel struct {
	CustomResponse fwtypes.ListNestedObjectValueOf[customResponseModel] `tfsdk:"custom_response"`
}

type captchaActionModel struct {
	CustomRequestHandling fwtypes.ListNestedObjectValueOf[customRequestHandlingModel] `tfsdk:"custom_request_handling"`
}

type challengeActionModel struct {
	CustomRequestHandling fwtypes.ListNestedObjectValueOf[customRequestHandlingModel] `tfsdk:"custom_request_handling"`
}

type countActionModel struct {
	CustomRequestHandling fwtypes.ListNestedObjectValueOf[customRequestHandlingModel] `tfsdk:"custom_request_handling"`
}

type customRequestHandlingModel struct {
	InsertHeader fwtypes.ListNestedObjectValueOf[insertHeaderModel] `tfsdk:"insert_header"`
}

type customResponseModel struct {
	CustomResponseBodyKey types.String                                         `tfsdk:"custom_response_body_key"`
	ResponseCode          types.Int32                                          `tfsdk:"response_code"`
	ResponseHeader        fwtypes.ListNestedObjectValueOf[responseHeaderModel] `tfsdk:"response_header"`
}

type insertHeaderModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type responseHeaderModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}
