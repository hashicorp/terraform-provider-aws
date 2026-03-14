// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package wafv2

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	webACLRuleGroupAssociationResourceIDPartCount = 4
	overrideActionNone                            = "none"
	overrideActionCount                           = "count"
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
	usernameFieldLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[usernameFieldModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrIdentifier: schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.All(
							stringvalidator.LengthBetween(1, 512),
							stringvalidator.RegexMatches(
								regexache.MustCompile(`\S`),
								"can not be empty string",
							),
						),
					},
					Description: "Identifier of the username field",
				},
			},
		},
	}

	passwordFieldLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[passwordFieldModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrIdentifier: schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.All(
							stringvalidator.LengthBetween(1, 512),
							stringvalidator.RegexMatches(
								regexache.MustCompile(`\S`),
								"can not be empty string",
							),
						),
					},
					Description: "Identifier of the password field",
				},
			},
		},
	}

	managedRulegroupConfigACFPRequestInspectionLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[requestInspectionACFPModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"payload_type": schema.StringAttribute{
					CustomType:  fwtypes.StringEnumType[awstypes.PayloadType](),
					Required:    true,
					Description: "Payload type for inspection, either JSON or FORM_ENCODED.",
				},
			},
			Blocks: map[string]schema.Block{
				"username_field": usernameFieldLNB,
				"address_fields": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[addressFieldModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
						listvalidator.SizeAtLeast(0),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"identifiers": schema.ListAttribute{
								ElementType: types.StringType,
								Required:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
								Description: "Identifiers of the address fields",
							},
						},
					},
				},
				"email_field": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[emailFieldModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
						listvalidator.SizeAtLeast(0),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrIdentifier: schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									stringvalidator.All(
										stringvalidator.LengthBetween(1, 512),
										stringvalidator.RegexMatches(
											regexache.MustCompile(`\S`),
											"can not be empty string",
										),
									),
								},
								Description: "Identifier of the email field",
							},
						},
					},
				},
				"password_field": passwordFieldLNB,
				"phone_number_fields": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[phoneNumberFieldModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
						listvalidator.SizeAtLeast(0),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"identifiers": schema.ListAttribute{
								ElementType: types.StringType,
								Required:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
								Description: "Identifiers of the phone number fields",
							},
						},
					},
				},
			},
		},
	}
	managedRulegroupConfigATPRequestInspectionLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[requestInspectionModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"payload_type": schema.StringAttribute{
					CustomType:  fwtypes.StringEnumType[awstypes.PayloadType](),
					Required:    true,
					Description: "Payload type for inspection, either JSON or FORM_ENCODED.",
				},
			},
			Blocks: map[string]schema.Block{
				"password_field": passwordFieldLNB,
				"username_field": usernameFieldLNB,
			},
		},
	}
	managedRulegroupConfigResponseInspectionLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[responseInspectionModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"body_contains": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[responseInspectionBodyContainsModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"failure_strings": schema.SetAttribute{
								Required:    true,
								ElementType: types.StringType,
								Description: "Strings that indicate a failed login or account creation attempt",
							},
							"success_strings": schema.SetAttribute{
								Required:    true,
								ElementType: types.StringType,
								Description: "Strings that indicate a successful login or account creation attempt",
							},
						},
					},
				},
				names.AttrHeader: schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[responseInspectionHeaderModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrName: schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									stringvalidator.LengthBetween(1, 256),
								},
								Description: "Name of the HTTP header to inspect",
							},
							"failure_values": schema.SetAttribute{
								Required:    true,
								ElementType: types.StringType,
								Description: "Strings that indicate a failed login or account creation attempt",
							},
							"success_values": schema.SetAttribute{
								Required:    true,
								ElementType: types.StringType,
								Description: "Strings that indicate a successful login or account creation attempt",
							},
						},
					},
				},
				names.AttrJSON: schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[responseInspectionJsonModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrIdentifier: schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									stringvalidator.LengthBetween(1, 256),
								},
								Description: "Identifier of the JSON field to inspect",
							},
							"failure_values": schema.SetAttribute{
								Required:    true,
								ElementType: types.StringType,
								Description: "Strings that indicate a failed login or account creation attempt",
							},
							"success_values": schema.SetAttribute{
								Required:    true,
								ElementType: types.StringType,
								Description: "Strings that indicate a successful login or account creation attempt",
							},
						},
					},
				},
				names.AttrStatusCode: schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[responseInspectionStatusCodeModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"failure_codes": schema.SetAttribute{
								Required:    true,
								ElementType: types.Int32Type,
								Description: "Status codes that indicate a failed login or account creation attempt",
							},
							"success_codes": schema.SetAttribute{
								Required:    true,
								ElementType: types.Int32Type,
								Description: "Status codes that indicate a successful login or account creation attempt",
							},
						},
					},
				},
			},
		},
	}
	managedRulegroupConfigLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[managedRuleGroupConfigModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"aws_managed_rules_acfp_rule_set": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[awsManagedRulesACFPRuleSetModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"creation_path": schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									stringvalidator.All(
										stringvalidator.LengthBetween(1, 256),
										stringvalidator.RegexMatches(
											regexache.MustCompile(`\S`),
											"can not be empty string",
										),
									),
								},
								Description: "Path to the account creation endpoint on the protected website",
							},
							"enable_regex_in_path": schema.BoolAttribute{
								Optional: true,
								Computed: true,
								Default:  booldefault.StaticBool(false),
							},
							"registration_page_path": schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									stringvalidator.All(
										stringvalidator.LengthBetween(1, 256),
										stringvalidator.RegexMatches(
											regexache.MustCompile(`\S`),
											"can not be empty string",
										),
									),
								},
							},
						},
						Blocks: map[string]schema.Block{
							"request_inspection":  managedRulegroupConfigACFPRequestInspectionLNB,
							"response_inspection": managedRulegroupConfigResponseInspectionLNB,
						},
					},
				},
				"aws_managed_rules_anti_ddos_rule_set": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[awsManagedRulesAntiDDoSRuleSetModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"sensitivity_to_block": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.SensitivityToAct](),
								Optional:   true,
								Computed:   true,
								Default:    stringdefault.StaticString(string(awstypes.SensitivityToActLow)),
							},
						},
						Blocks: map[string]schema.Block{
							"client_side_action_config": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[clientSideActionConfigModel](ctx),
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
									listvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedBlockObject{
									Blocks: map[string]schema.Block{
										"challenge": schema.ListNestedBlock{
											CustomType: fwtypes.NewListNestedObjectTypeOf[clientSideActionModel](ctx),
											Validators: []validator.List{
												listvalidator.SizeAtMost(1),
												listvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedBlockObject{
												Blocks: map[string]schema.Block{
													"exempt_uri_regular_expression": schema.ListNestedBlock{
														//CustomType: fwtypes.NewListNestedObjectTypeOf[exemptUriRegularExpressionModel](ctx),
														CustomType: fwtypes.NewListNestedObjectTypeOf[regexModel](ctx),
														Validators: []validator.List{
															listvalidator.SizeAtMost(5),
														},
														NestedObject: schema.NestedBlockObject{
															Attributes: map[string]schema.Attribute{
																"regex_string": schema.StringAttribute{
																	Optional: true,
																	Validators: []validator.String{
																		stringvalidator.LengthBetween(1, 512),
																	},
																},
															},
														},
													},
												},
												Attributes: map[string]schema.Attribute{
													"sensitivity": schema.StringAttribute{
														CustomType: fwtypes.StringEnumType[awstypes.SensitivityToAct](),
														Optional:   true,
														Computed:   true,
														Default:    stringdefault.StaticString(string(awstypes.SensitivityToActHigh)),
													},
													"usage_of_action": schema.StringAttribute{
														CustomType: fwtypes.StringEnumType[awstypes.UsageOfAction](),
														Required:   true,
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
				"aws_managed_rules_atp_rule_set": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[awsManagedRulesATPRuleSetModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"enable_regex_in_path": schema.BoolAttribute{
								Optional: true,
								Computed: true,
								Default:  booldefault.StaticBool(false),
							},
							"login_path": schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									stringvalidator.All(
										stringvalidator.LengthBetween(1, 256),
										stringvalidator.RegexMatches(
											regexache.MustCompile(`\S`),
											"can not be empty string",
										),
									),
								},
							},
						},
						Blocks: map[string]schema.Block{
							"request_inspection":  managedRulegroupConfigATPRequestInspectionLNB,
							"response_inspection": managedRulegroupConfigResponseInspectionLNB,
						},
					},
				},
				"aws_managed_rules_bot_control_rule_set": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[awsManagedRulesBotControlRuleSetModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"enable_machine_learning": schema.BoolAttribute{
								Optional: true,
								Default:  booldefault.StaticBool(false),
								Computed: true,
							},
							"inspection_level": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.InspectionLevel](),
								Required:   true,
							},
						},
					},
				},
			},
		},
	}
	ruleActionOverrideLNB := schema.ListNestedBlock{
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
					CustomType: fwtypes.NewListNestedObjectTypeOf[ruleActionModel](ctx),
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
														CustomType: fwtypes.NewListNestedObjectTypeOf[customHTTPHeaderModel](ctx),
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
														CustomType: fwtypes.NewListNestedObjectTypeOf[customHTTPHeaderModel](ctx),
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
														CustomType: fwtypes.NewListNestedObjectTypeOf[customHTTPHeaderModel](ctx),
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
														CustomType: fwtypes.NewListNestedObjectTypeOf[customHTTPHeaderModel](ctx),
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
														CustomType: fwtypes.NewListNestedObjectTypeOf[customHTTPHeaderModel](ctx),
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
	}
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
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
					stringvalidator.OneOf(overrideActionNone, overrideActionCount),
				},
				Description: "Override action for the rule group. Valid values are 'none' and 'count'. Defaults to 'none'.",
			},
		},
		Blocks: map[string]schema.Block{
			"rule_group_reference": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[ruleGroupReferenceModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.SizeAtLeast(0),
					listvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("managed_rule_group")),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrARN: schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								fwvalidators.ARN(),
							},
							Description: "ARN of the Rule Group to associate with the Web ACL.",
						},
					},
					Blocks: map[string]schema.Block{
						"rule_action_override": ruleActionOverrideLNB,
					},
				},
				Description: "Rule Group reference configuration.",
			},
			"managed_rule_group": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[managedRuleGroupModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.SizeAtLeast(0),
					listvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("rule_group_reference")),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 128),
							},
							Description: "Name of the managed rule group.",
						},
						"vendor_name": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 128),
							},
							Description: "Name of the managed rule group vendor.",
						},
						names.AttrVersion: schema.StringAttribute{
							Optional: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 64),
							},
							Description: "Version of the managed rule group. Omit this to use the default version.",
						},
					},
					Blocks: map[string]schema.Block{
						"rule_action_override":       ruleActionOverrideLNB,
						"managed_rule_group_configs": managedRulegroupConfigLNB,
					},
				},
				Description: "Managed rule group configuration.",
			},
			"visibility_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[visibilityConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"sampled_requests_enabled": schema.BoolAttribute{
							Required:    true,
							Description: "Indicates whether to store a sampling of the web requests that match the rule.",
						},
						"cloudwatch_metrics_enabled": schema.BoolAttribute{
							Required:    true,
							Description: "Indicates whether the rule is available for use in the metrics for the web ACL.",
						},
						names.AttrMetricName: schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.All(
									stringvalidator.LengthBetween(1, 128),
									stringvalidator.RegexMatches(
										regexache.MustCompile(`^[0-9A-Za-z_-]+$`),
										"can only contain alphanumeric characters, hyphens, underscores, and colons",
									),
								),
							},
							Description: "A name for the metrics for this rule.",
						},
					},
				},
				Description: "Visibility configuration for the rule.",
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
		Description: "Associates a WAFv2 Rule Group (custom or managed) with a Web ACL by adding a rule that references the Rule Group.",
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

	// Get rule configuration from either custom or managed rule group
	var ruleGroupARN string
	var ruleGroupName string
	var ruleGroupVendorName string
	var ruleGroupVersion string
	var ruleActionOverrides []awstypes.RuleActionOverride
	var ruleStatement *awstypes.Statement

	// Check for custom rule group reference
	if !plan.RuleGroupReference.IsNull() && !plan.RuleGroupReference.IsUnknown() {
		ruleGroupRefs := plan.RuleGroupReference.Elements()
		if len(ruleGroupRefs) > 0 {
			var ruleGroupRefModel ruleGroupReferenceModel
			resp.Diagnostics.Append(ruleGroupRefs[0].(fwtypes.ObjectValueOf[ruleGroupReferenceModel]).As(ctx, &ruleGroupRefModel, basetypes.ObjectAsOptions{})...)
			if resp.Diagnostics.HasError() {
				return
			}
			ruleGroupARN = ruleGroupRefModel.ARN.ValueString()

			// Create rule group reference statement
			ruleGroupRefStatement := &awstypes.RuleGroupReferenceStatement{
				ARN: aws.String(ruleGroupARN),
			}

			// Add rule action overrides if specified
			if !ruleGroupRefModel.RuleActionOverride.IsNull() && !ruleGroupRefModel.RuleActionOverride.IsUnknown() {
				resp.Diagnostics.Append(fwflex.Expand(ctx, ruleGroupRefModel.RuleActionOverride, &ruleActionOverrides)...)
				if resp.Diagnostics.HasError() {
					return
				}
				ruleGroupRefStatement.RuleActionOverrides = ruleActionOverrides
			}

			ruleStatement = &awstypes.Statement{
				RuleGroupReferenceStatement: ruleGroupRefStatement,
			}
		}
	}

	// Check for managed rule group (mutually exclusive with custom)
	if !plan.ManagedRuleGroup.IsNull() && !plan.ManagedRuleGroup.IsUnknown() {
		managedRuleGroups := plan.ManagedRuleGroup.Elements()
		if len(managedRuleGroups) > 0 {
			var managedRuleGroupRef managedRuleGroupModel
			resp.Diagnostics.Append(managedRuleGroups[0].(fwtypes.ObjectValueOf[managedRuleGroupModel]).As(ctx, &managedRuleGroupRef, basetypes.ObjectAsOptions{})...)
			if resp.Diagnostics.HasError() {
				return
			}
			ruleGroupName = managedRuleGroupRef.Name.ValueString()
			ruleGroupVendorName = managedRuleGroupRef.VendorName.ValueString()
			if !managedRuleGroupRef.Version.IsNull() && !managedRuleGroupRef.Version.IsUnknown() {
				ruleGroupVersion = managedRuleGroupRef.Version.ValueString()
			}

			// Create managed rule group statement
			managedRuleGroupStatement := &awstypes.ManagedRuleGroupStatement{
				Name:       aws.String(ruleGroupName),
				VendorName: aws.String(ruleGroupVendorName),
			}
			if ruleGroupVersion != "" {
				managedRuleGroupStatement.Version = aws.String(ruleGroupVersion)
			}

			// Add managed rule group configurations if specified
			if !managedRuleGroupRef.ManagedRuleGroupConfig.IsNull() && !managedRuleGroupRef.ManagedRuleGroupConfig.IsUnknown() {
				var tfManagedRuleGroupConfigModel managedRuleGroupConfigModel
				var managedRuleGroupConfigs []awstypes.ManagedRuleGroupConfig

				resp.Diagnostics.Append(managedRuleGroupRef.ManagedRuleGroupConfig.Elements()[0].(fwtypes.ObjectValueOf[managedRuleGroupConfigModel]).As(ctx, &tfManagedRuleGroupConfigModel, basetypes.ObjectAsOptions{})...)
				if resp.Diagnostics.HasError() {
					return
				}

				if !tfManagedRuleGroupConfigModel.AWSManagedRulesACFPRuleSet.IsNull() && !tfManagedRuleGroupConfigModel.AWSManagedRulesACFPRuleSet.IsUnknown() {
					// custom expand for ACFP Managed Rule Group Config
					var tfACFPConfig awsManagedRulesACFPRuleSetModel
					resp.Diagnostics.Append(tfManagedRuleGroupConfigModel.AWSManagedRulesACFPRuleSet.Elements()[0].(fwtypes.ObjectValueOf[awsManagedRulesACFPRuleSetModel]).As(ctx, &tfACFPConfig, basetypes.ObjectAsOptions{})...)
					if resp.Diagnostics.HasError() {
						return
					}

					var acfpConfig awstypes.AWSManagedRulesACFPRuleSet
					resp.Diagnostics.Append(fwflex.Expand(ctx, tfACFPConfig, &acfpConfig)...)
					if resp.Diagnostics.HasError() {
						return
					}

					managedRuleGroupConfigs = append(managedRuleGroupConfigs, awstypes.ManagedRuleGroupConfig{
						AWSManagedRulesACFPRuleSet: &acfpConfig,
					})
				} else {
					// let auto flex handle expand
					resp.Diagnostics.Append(fwflex.Expand(ctx, managedRuleGroupRef.ManagedRuleGroupConfig, &managedRuleGroupConfigs)...)
					if resp.Diagnostics.HasError() {
						return
					}
				}

				managedRuleGroupStatement.ManagedRuleGroupConfigs = managedRuleGroupConfigs
			}

			// Add rule action overrides if specified
			if !managedRuleGroupRef.RuleActionOverride.IsNull() && !managedRuleGroupRef.RuleActionOverride.IsUnknown() {
				resp.Diagnostics.Append(fwflex.Expand(ctx, managedRuleGroupRef.RuleActionOverride, &ruleActionOverrides)...)
				if resp.Diagnostics.HasError() {
					return
				}
				managedRuleGroupStatement.RuleActionOverrides = ruleActionOverrides
			}

			ruleStatement = &awstypes.Statement{
				ManagedRuleGroupStatement: managedRuleGroupStatement,
			}
		}
	}

	if ruleStatement == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionCreating, ResNameWebACLRuleGroupAssociation, plan.RuleName.String(), nil),
			"Either rule_group_reference or managed_rule_group block is required",
		)
		return
	}

	// Build visibility config for the new rule
	var visibilityConfig awstypes.VisibilityConfig
	if !plan.VisibilityConfig.IsNull() && !plan.VisibilityConfig.IsUnknown() {
		// Expand user-provided visibility_config into AWS SDK type
		resp.Diagnostics.Append(fwflex.Expand(ctx, plan.VisibilityConfig, &visibilityConfig)...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		// Use defaults when not provided in plan, for backward compatibility of this provider
		visibilityConfig = awstypes.VisibilityConfig{
			SampledRequestsEnabled:   true,
			CloudWatchMetricsEnabled: true,
			MetricName:               plan.RuleName.ValueStringPointer(),
		}
	}

	// Create new rule with the appropriate statement type
	newRule := awstypes.Rule{
		Name:             plan.RuleName.ValueStringPointer(),
		Priority:         plan.Priority.ValueInt32(),
		Statement:        ruleStatement,
		VisibilityConfig: &visibilityConfig,
	}

	// Set override action
	overrideAction := plan.OverrideAction.ValueString()
	if overrideAction == "" {
		overrideAction = overrideActionNone
		plan.OverrideAction = types.StringValue(overrideActionNone) // Set the default in the plan
	}

	switch overrideAction {
	case overrideActionNone:
		newRule.OverrideAction = &awstypes.OverrideAction{
			None: &awstypes.NoneAction{},
		}
	case overrideActionCount:
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
	_, err = tfresource.RetryWhenIsA[any, *awstypes.WAFUnavailableEntityException](ctx, timeout, func(ctx context.Context) (any, error) {
		return conn.UpdateWebACL(ctx, updateInput)
	})

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionCreating, ResNameWebACLRuleGroupAssociation, plan.RuleName.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceWebACLRuleGroupAssociation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceWebACLRuleGroupAssociationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WAFV2Client(ctx)

	// Use attributes directly instead of parsing ID
	webACLARN := state.WebACLARN.ValueString()
	ruleName := state.RuleName.ValueString()

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
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Reading WAFv2 Web ACL Rule Group Association",
			fmt.Sprintf("Error reading Web ACL: %s", err),
		)
		return
	}

	// Find the rule group in the Web ACL rules
	found := false
	for _, rule := range webACL.WebACL.Rules {
		if aws.ToString(rule.Name) != ruleName {
			continue
		}

		// Check if this rule matches our rule group configuration from state
		if rule.Statement != nil {
			var matchesRuleGroup bool
			var ruleActionOverrides fwtypes.ListNestedObjectValueOf[ruleActionOverrideModel]

			// Check if we have a custom rule group in state
			if !state.RuleGroupReference.IsNull() && !state.RuleGroupReference.IsUnknown() && rule.Statement.RuleGroupReferenceStatement != nil {
				// Get the ARN from state for comparison
				ruleGroupRefs := state.RuleGroupReference.Elements()
				if len(ruleGroupRefs) > 0 {
					var ruleGroupRefModel ruleGroupReferenceModel
					resp.Diagnostics.Append(ruleGroupRefs[0].(fwtypes.ObjectValueOf[ruleGroupReferenceModel]).As(ctx, &ruleGroupRefModel, basetypes.ObjectAsOptions{})...)
					if resp.Diagnostics.HasError() {
						return
					}

					if aws.ToString(rule.Statement.RuleGroupReferenceStatement.ARN) == ruleGroupRefModel.ARN.ValueString() {
						matchesRuleGroup = true
						// Handle rule action overrides with autoflex
						if rule.Statement.RuleGroupReferenceStatement.RuleActionOverrides != nil {
							resp.Diagnostics.Append(fwflex.Flatten(ctx, rule.Statement.RuleGroupReferenceStatement.RuleActionOverrides, &ruleActionOverrides)...)
							if resp.Diagnostics.HasError() {
								return
							}
						} else {
							ruleActionOverrides = fwtypes.NewListNestedObjectValueOfNull[ruleActionOverrideModel](ctx)
						}

						// Update the rule group reference nested structure
						ruleGroupRefModel.RuleActionOverride = ruleActionOverrides
						listValue, diags := fwtypes.NewListNestedObjectValueOfSlice(ctx, []*ruleGroupReferenceModel{&ruleGroupRefModel}, nil)
						resp.Diagnostics.Append(diags...)
						if resp.Diagnostics.HasError() {
							return
						}
						state.RuleGroupReference = listValue
						state.ManagedRuleGroup = fwtypes.NewListNestedObjectValueOfNull[managedRuleGroupModel](ctx)
					}
				}
			} else if !state.ManagedRuleGroup.IsNull() && !state.ManagedRuleGroup.IsUnknown() && rule.Statement.ManagedRuleGroupStatement != nil {
				// Check if we have a managed rule group in state
				managedRuleGroups := state.ManagedRuleGroup.Elements()
				if len(managedRuleGroups) > 0 {
					var managedRuleGroupRef managedRuleGroupModel
					resp.Diagnostics.Append(managedRuleGroups[0].(fwtypes.ObjectValueOf[managedRuleGroupModel]).As(ctx, &managedRuleGroupRef, basetypes.ObjectAsOptions{})...)
					if resp.Diagnostics.HasError() {
						return
					}

					managedStmt := rule.Statement.ManagedRuleGroupStatement
					// Check if this matches our managed rule group from state
					if aws.ToString(managedStmt.Name) == managedRuleGroupRef.Name.ValueString() &&
						aws.ToString(managedStmt.VendorName) == managedRuleGroupRef.VendorName.ValueString() {
						// Check version match (both can be empty/null)
						stateVersion := managedRuleGroupRef.Version.ValueString()
						ruleVersion := aws.ToString(managedStmt.Version)
						if stateVersion == ruleVersion {
							matchesRuleGroup = true
							// Handle rule action overrides with autoflex
							if managedStmt.RuleActionOverrides != nil {
								resp.Diagnostics.Append(fwflex.Flatten(ctx, managedStmt.RuleActionOverrides, &ruleActionOverrides)...)
								if resp.Diagnostics.HasError() {
									return
								}
							} else {
								ruleActionOverrides = fwtypes.NewListNestedObjectValueOfNull[ruleActionOverrideModel](ctx)
							}

							var managedRuleGroupConfigs fwtypes.ListNestedObjectValueOf[managedRuleGroupConfigModel]
							if len(managedStmt.ManagedRuleGroupConfigs) > 0 {
								resp.Diagnostics.Append(fwflex.Flatten(ctx, managedStmt.ManagedRuleGroupConfigs, &managedRuleGroupConfigs)...)
								if resp.Diagnostics.HasError() {
									return
								}
							} else {
								managedRuleGroupConfigs = fwtypes.NewListNestedObjectValueOfNull[managedRuleGroupConfigModel](ctx)
							}
							// Update the managed rule group nested structure
							managedRuleGroupRef.RuleActionOverride = ruleActionOverrides
							managedRuleGroupRef.ManagedRuleGroupConfig = managedRuleGroupConfigs
							listValue, diags := fwtypes.NewListNestedObjectValueOfSlice(ctx, []*managedRuleGroupModel{&managedRuleGroupRef}, nil)
							resp.Diagnostics.Append(diags...)
							if resp.Diagnostics.HasError() {
								return
							}
							state.ManagedRuleGroup = listValue
							state.RuleGroupReference = fwtypes.NewListNestedObjectValueOfNull[ruleGroupReferenceModel](ctx)
						}
					}
				}
			}

			if matchesRuleGroup {
				found = true
				state.Priority = types.Int32Value(rule.Priority)

				var visibilityConfig fwtypes.ListNestedObjectValueOf[visibilityConfigModel]
				// for backward compatibility, visibility_config to be read only if it's not set to default
				if rule.VisibilityConfig != nil {
					isDefault := rule.VisibilityConfig.SampledRequestsEnabled &&
						rule.VisibilityConfig.CloudWatchMetricsEnabled &&
						aws.ToString(rule.VisibilityConfig.MetricName) == aws.ToString(rule.Name)

					if !isDefault {
						resp.Diagnostics.Append(fwflex.Flatten(ctx, rule.VisibilityConfig, &visibilityConfig)...)
						if resp.Diagnostics.HasError() {
							return
						}
						state.VisibilityConfig = visibilityConfig
					}
				}

				// Determine override action
				overrideAction := overrideActionNone
				if rule.OverrideAction != nil {
					if rule.OverrideAction.Count != nil {
						overrideAction = overrideActionCount
					} else if rule.OverrideAction.None != nil {
						overrideAction = overrideActionNone
					}
				}
				state.OverrideAction = types.StringValue(overrideAction)
				break
			}
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

	// Update state with current values (WebACLARN and RuleName should already be set from current state)
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
				overrideAction = overrideActionNone // Default value
			}

			switch overrideAction {
			case overrideActionNone:
				webACL.WebACL.Rules[i].OverrideAction = &awstypes.OverrideAction{
					None: &awstypes.NoneAction{},
				}
			case overrideActionCount:
				webACL.WebACL.Rules[i].OverrideAction = &awstypes.OverrideAction{
					Count: &awstypes.CountAction{},
				}
			}

			// update visibility config if changed
			var visibilityConfig awstypes.VisibilityConfig
			if !plan.VisibilityConfig.IsNull() && !plan.VisibilityConfig.IsUnknown() {
				resp.Diagnostics.Append(fwflex.Expand(ctx, plan.VisibilityConfig, &visibilityConfig)...)
				if resp.Diagnostics.HasError() {
					return
				}
				webACL.WebACL.Rules[i].VisibilityConfig = &visibilityConfig
			} else {
				visibilityConfig = awstypes.VisibilityConfig{
					SampledRequestsEnabled:   true,
					CloudWatchMetricsEnabled: true,
					MetricName:               plan.RuleName.ValueStringPointer(),
				}
			}

			// Update rule action overrides from nested structure (both custom and managed)
			var overrides []awstypes.RuleActionOverride
			var managedRuleGroupConfigs []awstypes.ManagedRuleGroupConfig

			if !plan.RuleGroupReference.IsNull() && !plan.RuleGroupReference.IsUnknown() {
				ruleGroupRefs := plan.RuleGroupReference.Elements()
				if len(ruleGroupRefs) > 0 {
					var ruleGroupRefModel ruleGroupReferenceModel
					resp.Diagnostics.Append(ruleGroupRefs[0].(fwtypes.ObjectValueOf[ruleGroupReferenceModel]).As(ctx, &ruleGroupRefModel, basetypes.ObjectAsOptions{})...)
					if resp.Diagnostics.HasError() {
						return
					}

					if !ruleGroupRefModel.RuleActionOverride.IsNull() && !ruleGroupRefModel.RuleActionOverride.IsUnknown() {
						resp.Diagnostics.Append(fwflex.Expand(ctx, ruleGroupRefModel.RuleActionOverride, &overrides)...)
						if resp.Diagnostics.HasError() {
							return
						}
					}
				}
			} else if !plan.ManagedRuleGroup.IsNull() && !plan.ManagedRuleGroup.IsUnknown() {
				managedRuleGroups := plan.ManagedRuleGroup.Elements()
				if len(managedRuleGroups) > 0 {
					var managedRuleGroupRef managedRuleGroupModel
					resp.Diagnostics.Append(managedRuleGroups[0].(fwtypes.ObjectValueOf[managedRuleGroupModel]).As(ctx, &managedRuleGroupRef, basetypes.ObjectAsOptions{})...)
					if resp.Diagnostics.HasError() {
						return
					}

					if !managedRuleGroupRef.RuleActionOverride.IsNull() && !managedRuleGroupRef.RuleActionOverride.IsUnknown() {
						resp.Diagnostics.Append(fwflex.Expand(ctx, managedRuleGroupRef.RuleActionOverride, &overrides)...)
						if resp.Diagnostics.HasError() {
							return
						}
					}

					if !managedRuleGroupRef.ManagedRuleGroupConfig.IsNull() && !managedRuleGroupRef.ManagedRuleGroupConfig.IsUnknown() {
						var tfManagedRuleGroupConfigModel managedRuleGroupConfigModel

						resp.Diagnostics.Append(managedRuleGroupRef.ManagedRuleGroupConfig.Elements()[0].(fwtypes.ObjectValueOf[managedRuleGroupConfigModel]).As(ctx, &tfManagedRuleGroupConfigModel, basetypes.ObjectAsOptions{})...)
						if resp.Diagnostics.HasError() {
							return
						}

						if !tfManagedRuleGroupConfigModel.AWSManagedRulesACFPRuleSet.IsNull() && !tfManagedRuleGroupConfigModel.AWSManagedRulesACFPRuleSet.IsUnknown() {
							// custom expand for ACFP Managed Rule Group Config
							var tfACFPConfig awsManagedRulesACFPRuleSetModel
							resp.Diagnostics.Append(tfManagedRuleGroupConfigModel.AWSManagedRulesACFPRuleSet.Elements()[0].(fwtypes.ObjectValueOf[awsManagedRulesACFPRuleSetModel]).As(ctx, &tfACFPConfig, basetypes.ObjectAsOptions{})...)
							if resp.Diagnostics.HasError() {
								return
							}

							var acfpConfig awstypes.AWSManagedRulesACFPRuleSet
							resp.Diagnostics.Append(fwflex.Expand(ctx, tfACFPConfig, &acfpConfig)...)
							if resp.Diagnostics.HasError() {
								return
							}

							managedRuleGroupConfigs = append(managedRuleGroupConfigs, awstypes.ManagedRuleGroupConfig{
								AWSManagedRulesACFPRuleSet: &acfpConfig,
							})
						} else {
							// let auto flex handle expand
							resp.Diagnostics.Append(fwflex.Expand(ctx, managedRuleGroupRef.ManagedRuleGroupConfig, &managedRuleGroupConfigs)...)
							if resp.Diagnostics.HasError() {
								return
							}
						}
					}
				}
			}

			// Update the appropriate statement type with new overrides
			if webACL.WebACL.Rules[i].Statement != nil {
				if webACL.WebACL.Rules[i].Statement.RuleGroupReferenceStatement != nil {
					webACL.WebACL.Rules[i].Statement.RuleGroupReferenceStatement.RuleActionOverrides = overrides
				} else if webACL.WebACL.Rules[i].Statement.ManagedRuleGroupStatement != nil {
					webACL.WebACL.Rules[i].Statement.ManagedRuleGroupStatement.RuleActionOverrides = overrides
					webACL.WebACL.Rules[i].Statement.ManagedRuleGroupStatement.ManagedRuleGroupConfigs = managedRuleGroupConfigs
				}
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
	_, err = tfresource.RetryWhenIsA[any, *awstypes.WAFUnavailableEntityException](ctx, updateTimeout, func(ctx context.Context) (any, error) {
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

	// Use attributes directly instead of parsing ID
	webACLARN := state.WebACLARN.ValueString()
	ruleName := state.RuleName.ValueString()

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
	if retry.NotFound(err) {
		// Web ACL is already gone, nothing to do
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionDeleting, ResNameWebACLRuleGroupAssociation, state.RuleName.String(), err),
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
	_, err = tfresource.RetryWhenIsA[any, *awstypes.WAFUnavailableEntityException](ctx, timeout, func(ctx context.Context) (any, error) {
		return conn.UpdateWebACL(ctx, updateInput)
	})

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionDeleting, ResNameWebACLRuleGroupAssociation, state.RuleName.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceWebACLRuleGroupAssociation) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := intflex.ExpandResourceId(req.ID, webACLRuleGroupAssociationResourceIDPartCount, true)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: web_acl_arn,rule_name,rule_group_type,rule_group_identifier. Got: %q", req.ID),
		)
		return
	}

	webACLARN := parts[0]
	ruleName := parts[1]
	ruleGroupType := parts[2]
	ruleGroupIdentifier := parts[3]

	// Parse Web ACL ARN to get ID, name, and scope
	_, _, _, err = parseWebACLARN(webACLARN)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Web ACL ARN",
			fmt.Sprintf("Error parsing Web ACL ARN: %s", err),
		)
		return
	}

	// Set basic attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("web_acl_arn"), webACLARN)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("rule_name"), ruleName)...)

	// Set the appropriate rule group nested structure based on type
	switch ruleGroupType {
	case "custom":
		// Custom rule group (ARN format)
		if !arn.IsARN(ruleGroupIdentifier) {
			resp.Diagnostics.AddError(
				"Invalid Custom Rule Group Identifier",
				"Custom rule group identifier should be an ARN",
			)
			return
		}

		ruleGroupRefModel := &ruleGroupReferenceModel{
			ARN:                types.StringValue(ruleGroupIdentifier),
			RuleActionOverride: fwtypes.NewListNestedObjectValueOfNull[ruleActionOverrideModel](ctx),
		}

		listValue, diags := fwtypes.NewListNestedObjectValueOfSlice(ctx, []*ruleGroupReferenceModel{ruleGroupRefModel}, nil)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("rule_group_reference"), listValue)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("managed_rule_group"), fwtypes.NewListNestedObjectValueOfNull[managedRuleGroupModel](ctx))...)
	case "managed":
		// Managed rule group (vendorName:ruleName[:version] format)
		identifierParts := strings.Split(ruleGroupIdentifier, ":")
		if len(identifierParts) < 2 {
			resp.Diagnostics.AddError(
				"Invalid Managed Rule Group Identifier",
				"Managed rule group identifier should be in format 'vendorName:ruleName[:version]'",
			)
			return
		}

		vendorName := identifierParts[0]
		ruleGroupName := identifierParts[1]
		var version string
		if len(identifierParts) > 2 {
			version = identifierParts[2]
		}

		managedRuleGroupRef := &managedRuleGroupModel{
			Name:               types.StringValue(ruleGroupName),
			VendorName:         types.StringValue(vendorName),
			RuleActionOverride: fwtypes.NewListNestedObjectValueOfNull[ruleActionOverrideModel](ctx),

			ManagedRuleGroupConfig: fwtypes.NewListNestedObjectValueOfNull[managedRuleGroupConfigModel](ctx),
		}
		if version != "" {
			managedRuleGroupRef.Version = types.StringValue(version)
		} else {
			managedRuleGroupRef.Version = types.StringNull()
		}

		listValue, diags := fwtypes.NewListNestedObjectValueOfSlice(ctx, []*managedRuleGroupModel{managedRuleGroupRef}, nil)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("managed_rule_group"), listValue)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("rule_group_reference"), fwtypes.NewListNestedObjectValueOfNull[ruleGroupReferenceModel](ctx))...)
	default:
		resp.Diagnostics.AddError(
			"Invalid Rule Group Type",
			fmt.Sprintf("Rule group type must be 'custom' or 'managed', got: %s", ruleGroupType),
		)
		return
	}
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
	RuleName           types.String                                             `tfsdk:"rule_name"`
	Priority           types.Int32                                              `tfsdk:"priority"`
	RuleGroupReference fwtypes.ListNestedObjectValueOf[ruleGroupReferenceModel] `tfsdk:"rule_group_reference"`
	ManagedRuleGroup   fwtypes.ListNestedObjectValueOf[managedRuleGroupModel]   `tfsdk:"managed_rule_group"`
	WebACLARN          types.String                                             `tfsdk:"web_acl_arn"`
	OverrideAction     types.String                                             `tfsdk:"override_action"`
	VisibilityConfig   fwtypes.ListNestedObjectValueOf[visibilityConfigModel]   `tfsdk:"visibility_config"`
	Timeouts           timeouts.Value                                           `tfsdk:"timeouts"`
}

type visibilityConfigModel struct {
	CloudWatchMetricsEnabled types.Bool   `tfsdk:"cloudwatch_metrics_enabled"`
	MetricName               types.String `tfsdk:"metric_name"`
	SampledRequestsEnabled   types.Bool   `tfsdk:"sampled_requests_enabled"`
}

type ruleGroupReferenceModel struct {
	ARN                types.String                                             `tfsdk:"arn"`
	RuleActionOverride fwtypes.ListNestedObjectValueOf[ruleActionOverrideModel] `tfsdk:"rule_action_override"`
}

type managedRuleGroupModel struct {
	Name               types.String                                             `tfsdk:"name"`
	VendorName         types.String                                             `tfsdk:"vendor_name"`
	Version            types.String                                             `tfsdk:"version"`
	RuleActionOverride fwtypes.ListNestedObjectValueOf[ruleActionOverrideModel] `tfsdk:"rule_action_override"`

	ManagedRuleGroupConfig fwtypes.ListNestedObjectValueOf[managedRuleGroupConfigModel] `tfsdk:"managed_rule_group_configs"`
}

type managedRuleGroupConfigModel struct {
	AWSManagedRulesACFPRuleSet       fwtypes.ListNestedObjectValueOf[awsManagedRulesACFPRuleSetModel]       `tfsdk:"aws_managed_rules_acfp_rule_set"`
	AWSManagedRulesATPRuleSet        fwtypes.ListNestedObjectValueOf[awsManagedRulesATPRuleSetModel]        `tfsdk:"aws_managed_rules_atp_rule_set"`
	AWSManagedRulesAntiDDoSRuleSet   fwtypes.ListNestedObjectValueOf[awsManagedRulesAntiDDoSRuleSetModel]   `tfsdk:"aws_managed_rules_anti_ddos_rule_set"`
	AWSManagedRulesBotControlRuleSet fwtypes.ListNestedObjectValueOf[awsManagedRulesBotControlRuleSetModel] `tfsdk:"aws_managed_rules_bot_control_rule_set"`
}

type awsManagedRulesBotControlRuleSetModel struct {
	InspectionLevel       types.String `tfsdk:"inspection_level"`
	EnableMachineLearning types.Bool   `tfsdk:"enable_machine_learning"`
}

var _ fwflex.Flattener = &awsManagedRulesBotControlRuleSetModel{}

func (m *awsManagedRulesBotControlRuleSetModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch val := v.(type) {
	case *awstypes.AWSManagedRulesBotControlRuleSet:
		m.InspectionLevel = types.StringValue(string(val.InspectionLevel))
		if val.InspectionLevel == awstypes.InspectionLevelCommon && (val.EnableMachineLearning == nil || !*val.EnableMachineLearning) {
			m.EnableMachineLearning = types.BoolValue(false)
		} else {
			m.EnableMachineLearning = fwflex.BoolToFramework(ctx, val.EnableMachineLearning)
		}
	case awstypes.AWSManagedRulesBotControlRuleSet:
		m.InspectionLevel = types.StringValue(string(val.InspectionLevel))
		if val.InspectionLevel == awstypes.InspectionLevelCommon && (val.EnableMachineLearning == nil || !*val.EnableMachineLearning) {
			m.EnableMachineLearning = types.BoolValue(false)
		} else {
			m.EnableMachineLearning = fwflex.BoolToFramework(ctx, val.EnableMachineLearning)
		}
	default:
		diags.AddError("Unexpected type", fmt.Sprintf("expected awstypes.AWSManagedRulesBotControlRuleSet, got %T", v))
	}
	return diags
}

type awsManagedRulesAntiDDoSRuleSetModel struct {
	ClientSideActionConfig fwtypes.ListNestedObjectValueOf[clientSideActionConfigModel] `tfsdk:"client_side_action_config"`
	SensitivityToBlock     types.String                                                 `tfsdk:"sensitivity_to_block"`
}

type clientSideActionConfigModel struct {
	Challenge fwtypes.ListNestedObjectValueOf[clientSideActionModel] `tfsdk:"challenge"`
}

type clientSideActionModel struct {
	UsageOfAction               types.String                                `tfsdk:"usage_of_action"`
	ExemptUriRegularExpressions fwtypes.ListNestedObjectValueOf[regexModel] `tfsdk:"exempt_uri_regular_expression"`
	Sensitivity                 types.String                                `tfsdk:"sensitivity"`
}

type regexModel struct {
	RegexString types.String `tfsdk:"regex_string"`
}

type awsManagedRulesATPRuleSetModel struct {
	LoginPath          types.String                                             `tfsdk:"login_path"`
	EnableRegexInPath  types.Bool                                               `tfsdk:"enable_regex_in_path"`
	RequestInspection  fwtypes.ListNestedObjectValueOf[requestInspectionModel]  `tfsdk:"request_inspection"`
	ResponseInspection fwtypes.ListNestedObjectValueOf[responseInspectionModel] `tfsdk:"response_inspection"`
}

type requestInspectionModel struct {
	PasswordField fwtypes.ListNestedObjectValueOf[passwordFieldModel] `tfsdk:"password_field"`
	PayloadType   types.String                                        `tfsdk:"payload_type"`
	UsernameField fwtypes.ListNestedObjectValueOf[usernameFieldModel] `tfsdk:"username_field"`
}

type usernameFieldModel struct {
	Identifier types.String `tfsdk:"identifier"`
}

type passwordFieldModel struct {
	Identifier types.String `tfsdk:"identifier"`
}

var (
	_ fwflex.Expander = awsManagedRulesACFPRuleSetModel{}
)

type awsManagedRulesACFPRuleSetModel struct {
	CreationPath         types.String                                                `tfsdk:"creation_path"`
	RegistrationPagePath types.String                                                `tfsdk:"registration_page_path"`
	RequestInspection    fwtypes.ListNestedObjectValueOf[requestInspectionACFPModel] `tfsdk:"request_inspection"`
	EnableRegexInPath    types.Bool                                                  `tfsdk:"enable_regex_in_path"`
	ResponseInspection   fwtypes.ListNestedObjectValueOf[responseInspectionModel]    `tfsdk:"response_inspection"`
}

func (m awsManagedRulesACFPRuleSetModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	var r awstypes.AWSManagedRulesACFPRuleSet

	r.CreationPath = m.CreationPath.ValueStringPointer()
	r.RegistrationPagePath = m.RegistrationPagePath.ValueStringPointer()
	r.EnableRegexInPath = m.EnableRegexInPath.ValueBool()

	if !m.RequestInspection.IsNull() && len(m.RequestInspection.Elements()) > 0 {
		var reqInspection requestInspectionACFPModel
		diags.Append(m.RequestInspection.Elements()[0].(fwtypes.ObjectValueOf[requestInspectionACFPModel]).As(ctx, &reqInspection, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		expanded, d := reqInspection.Expand(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		r.RequestInspection = expanded.(*awstypes.RequestInspectionACFP)
	}

	diags.Append(fwflex.Expand(ctx, m.ResponseInspection, &r.ResponseInspection)...)

	return &r, diags
}

var (
	_ fwflex.Expander  = awsManagedRulesACFPRuleSetModel{}
	_ fwflex.Flattener = &awsManagedRulesACFPRuleSetModel{}
	_ fwflex.Expander  = requestInspectionACFPModel{}
	_ fwflex.Flattener = &requestInspectionACFPModel{}
)

func (m *requestInspectionACFPModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch val := v.(type) {
	case *awstypes.RequestInspectionACFP:
		m.PayloadType = types.StringValue(string(val.PayloadType))
		diags.Append(fwflex.Flatten(ctx, val.EmailField, &m.EmailField)...)
		diags.Append(fwflex.Flatten(ctx, val.PasswordField, &m.PasswordField)...)
		diags.Append(fwflex.Flatten(ctx, val.UsernameField, &m.UsernameField)...)

		if len(val.AddressFields) > 0 {
			var identifiers []attr.Value
			for _, af := range val.AddressFields {
				identifiers = append(identifiers, types.StringValue(aws.ToString(af.Identifier)))
			}
			m.AddressFields = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*addressFieldModel{{
				Identifiers: fwtypes.NewListValueOfMust[types.String](ctx, identifiers),
			}})
		}

		if len(val.PhoneNumberFields) > 0 {
			var identifiers []attr.Value
			for _, pf := range val.PhoneNumberFields {
				identifiers = append(identifiers, types.StringValue(aws.ToString(pf.Identifier)))
			}
			m.PhoneNumberFields = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*phoneNumberFieldModel{{
				Identifiers: fwtypes.NewListValueOfMust[types.String](ctx, identifiers),
			}})
		}
	case awstypes.RequestInspectionACFP:
		m.PayloadType = types.StringValue(string(val.PayloadType))
		diags.Append(fwflex.Flatten(ctx, val.EmailField, &m.EmailField)...)
		diags.Append(fwflex.Flatten(ctx, val.PasswordField, &m.PasswordField)...)
		diags.Append(fwflex.Flatten(ctx, val.UsernameField, &m.UsernameField)...)

		if len(val.AddressFields) > 0 {
			var identifiers []attr.Value
			for _, af := range val.AddressFields {
				identifiers = append(identifiers, types.StringValue(aws.ToString(af.Identifier)))
			}
			m.AddressFields = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*addressFieldModel{{
				Identifiers: fwtypes.NewListValueOfMust[types.String](ctx, identifiers),
			}})
		}

		if len(val.PhoneNumberFields) > 0 {
			var identifiers []attr.Value
			for _, pf := range val.PhoneNumberFields {
				identifiers = append(identifiers, types.StringValue(aws.ToString(pf.Identifier)))
			}
			m.PhoneNumberFields = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*phoneNumberFieldModel{{
				Identifiers: fwtypes.NewListValueOfMust[types.String](ctx, identifiers),
			}})
		}
	default:
		diags.AddError("Unexpected type", fmt.Sprintf("expected *awstypes.RequestInspectionACFP, got %T", v))
	}
	return diags
}

func (m *awsManagedRulesACFPRuleSetModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch val := v.(type) {
	case *awstypes.AWSManagedRulesACFPRuleSet:
		m.CreationPath = fwflex.StringToFramework(ctx, val.CreationPath)
		m.RegistrationPagePath = fwflex.StringToFramework(ctx, val.RegistrationPagePath)
		m.EnableRegexInPath = types.BoolValue(val.EnableRegexInPath)
		m.ResponseInspection = fwtypes.NewListNestedObjectValueOfNull[responseInspectionModel](ctx)
		diags.Append(fwflex.Flatten(ctx, val.RequestInspection, &m.RequestInspection)...)
	case awstypes.AWSManagedRulesACFPRuleSet:
		m.CreationPath = fwflex.StringToFramework(ctx, val.CreationPath)
		m.RegistrationPagePath = fwflex.StringToFramework(ctx, val.RegistrationPagePath)
		m.EnableRegexInPath = types.BoolValue(val.EnableRegexInPath)
		m.ResponseInspection = fwtypes.NewListNestedObjectValueOfNull[responseInspectionModel](ctx)
		diags.Append(fwflex.Flatten(ctx, val.RequestInspection, &m.RequestInspection)...)
	default:
		diags.AddError("Unexpected type", fmt.Sprintf("expected *awstypes.AWSManagedRulesACFPRuleSet, got %T", v))
	}
	return diags
}

type requestInspectionACFPModel struct {
	PayloadType       types.String                                           `tfsdk:"payload_type"`
	AddressFields     fwtypes.ListNestedObjectValueOf[addressFieldModel]     `tfsdk:"address_fields"`
	EmailField        fwtypes.ListNestedObjectValueOf[emailFieldModel]       `tfsdk:"email_field"`
	PasswordField     fwtypes.ListNestedObjectValueOf[passwordFieldModel]    `tfsdk:"password_field"`
	UsernameField     fwtypes.ListNestedObjectValueOf[usernameFieldModel]    `tfsdk:"username_field"`
	PhoneNumberFields fwtypes.ListNestedObjectValueOf[phoneNumberFieldModel] `tfsdk:"phone_number_fields"`
}

func (m requestInspectionACFPModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	var r awstypes.RequestInspectionACFP

	r.PayloadType = awstypes.PayloadType(m.PayloadType.ValueString())

	for _, af := range m.AddressFields.Elements() {
		var addressField addressFieldModel
		diags.Append(af.(fwtypes.ObjectValueOf[addressFieldModel]).As(ctx, &addressField, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		expanded, d := addressField.Expand(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		r.AddressFields = append(r.AddressFields, expanded.([]awstypes.AddressField)...)
	}

	for _, pf := range m.PhoneNumberFields.Elements() {
		var phoneField phoneNumberFieldModel
		diags.Append(pf.(fwtypes.ObjectValueOf[phoneNumberFieldModel]).As(ctx, &phoneField, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		expanded, d := phoneField.Expand(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		r.PhoneNumberFields = append(r.PhoneNumberFields, expanded.([]awstypes.PhoneNumberField)...)
	}

	diags.Append(fwflex.Expand(ctx, m.EmailField, &r.EmailField)...)
	diags.Append(fwflex.Expand(ctx, m.PasswordField, &r.PasswordField)...)
	diags.Append(fwflex.Expand(ctx, m.UsernameField, &r.UsernameField)...)

	return &r, diags
}

type phoneNumberFieldModel struct {
	Identifiers fwtypes.ListValueOf[types.String] `tfsdk:"identifiers"`
}

var _ fwflex.Expander = phoneNumberFieldModel{}

func (m phoneNumberFieldModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	var phoneNumberFields []awstypes.PhoneNumberField

	for _, identifier := range m.Identifiers.Elements() {
		phoneNumberFields = append(phoneNumberFields, awstypes.PhoneNumberField{
			Identifier: identifier.(types.String).ValueStringPointer(),
		})
	}

	return phoneNumberFields, diags
}

type emailFieldModel struct {
	Identifier types.String `tfsdk:"identifier"`
}

type addressFieldModel struct {
	Identifiers fwtypes.ListValueOf[types.String] `tfsdk:"identifiers"`
}

var _ fwflex.Expander = addressFieldModel{}

func (m addressFieldModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	var addressFields []awstypes.AddressField

	for _, identifier := range m.Identifiers.Elements() {
		addressFields = append(addressFields, awstypes.AddressField{
			Identifier: identifier.(types.String).ValueStringPointer(),
		})
	}

	return addressFields, diags
}

type responseInspectionModel struct {
	BodyContains fwtypes.ListNestedObjectValueOf[responseInspectionBodyContainsModel] `tfsdk:"body_contains"`
	Header       fwtypes.ListNestedObjectValueOf[responseInspectionHeaderModel]       `tfsdk:"header"`
	Json         fwtypes.ListNestedObjectValueOf[responseInspectionJsonModel]         `tfsdk:"json"`
	StatusCode   fwtypes.ListNestedObjectValueOf[responseInspectionStatusCodeModel]   `tfsdk:"status_code"`
}

type responseInspectionBodyContainsModel struct {
	FailureStrings fwtypes.SetValueOf[types.String] `tfsdk:"failure_strings"`
	SuccessStrings fwtypes.SetValueOf[types.String] `tfsdk:"success_strings"`
}

type responseInspectionHeaderModel struct {
	FailureValues fwtypes.SetValueOf[types.String] `tfsdk:"failure_values"`
	Name          types.String                     `tfsdk:"name"`
	SuccessValues fwtypes.SetValueOf[types.String] `tfsdk:"success_values"`
}

type responseInspectionJsonModel struct {
	FailureValues fwtypes.SetValueOf[types.String] `tfsdk:"failure_values"`
	Identifier    types.String                     `tfsdk:"identifier"`
	SuccessValues fwtypes.SetValueOf[types.String] `tfsdk:"success_values"`
}

type responseInspectionStatusCodeModel struct {
	FailureCodes fwtypes.SetValueOf[types.Int32] `tfsdk:"failure_codes"`
	SuccessCodes fwtypes.SetValueOf[types.Int32] `tfsdk:"success_codes"`
}

type ruleActionOverrideModel struct {
	Name        types.String                                     `tfsdk:"name"`
	ActionToUse fwtypes.ListNestedObjectValueOf[ruleActionModel] `tfsdk:"action_to_use"`
}

type ruleActionModel struct {
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
	InsertHeader fwtypes.ListNestedObjectValueOf[customHTTPHeaderModel] `tfsdk:"insert_header"`
}

type customResponseModel struct {
	CustomResponseBodyKey types.String                                           `tfsdk:"custom_response_body_key"`
	ResponseCode          types.Int32                                            `tfsdk:"response_code"`
	ResponseHeader        fwtypes.ListNestedObjectValueOf[customHTTPHeaderModel] `tfsdk:"response_header"`
}

type customHTTPHeaderModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}
