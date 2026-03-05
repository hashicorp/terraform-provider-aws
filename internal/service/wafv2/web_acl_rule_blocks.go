// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// statementBlock and related functions are generated in web_acl_rule_statement_models_gen.go

func ipSetReferenceStatementBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleIPSetReferenceStatementModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
			listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("geo_match_statement")),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrARN: schema.StringAttribute{
					Required:    true,
					Description: "ARN of the IP set to reference.",
				},
			},
			Blocks: map[string]schema.Block{
				"ip_set_forwarded_ip_config": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleIPSetForwardedIPConfigModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"fallback_behavior": schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									stringvalidator.OneOf("MATCH", "NO_MATCH"),
								},
							},
							"header_name": schema.StringAttribute{
								Required: true,
							},
							"position": schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									stringvalidator.OneOf("FIRST", "LAST", "ANY"),
								},
							},
						},
					},
				},
			},
		},
		Description: "IP set reference statement.",
	}
}

func geoMatchStatementBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleGeoMatchStatementModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
			listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("ip_set_reference_statement")),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"country_codes": schema.ListAttribute{
					ElementType: types.StringType,
					Required:    true,
					Description: "List of two-character country codes (e.g., US, CA).",
				},
			},
			Blocks: map[string]schema.Block{
				"forwarded_ip_config": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleForwardedIPConfigModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"fallback_behavior": schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									stringvalidator.OneOf("MATCH", "NO_MATCH"),
								},
							},
							"header_name": schema.StringAttribute{
								Required: true,
							},
						},
					},
				},
			},
		},
		Description: "Geo match statement.",
	}
}

func ruleGroupReferenceStatementBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleRuleGroupReferenceStatementModel](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrARN: schema.StringAttribute{
					Required:    true,
					Description: "ARN of the RuleGroup (20-2048 characters).",
					Validators: []validator.String{
						stringvalidator.LengthBetween(20, 2048),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"excluded_rule":        excludedRuleBlock(ctx),
				"rule_action_override": ruleActionOverrideBlock(ctx),
			},
		},
		Description: "Rule statement used to run the rules that are defined in a RuleGroup.",
	}
}

func excludedRuleBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleExcludedRuleModel](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(100)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrName: schema.StringAttribute{
					Required:    true,
					Description: "Name of the rule to exclude (1-128 characters).",
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 128),
						stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z_-]{1,128}$`), "must contain only alphanumeric characters, underscores, and hyphens"),
					},
				},
			},
		},
		Description: "Rules in the referenced rule group whose actions are set to Count. Deprecated: use rule_action_override instead.",
	}
}

func managedRuleGroupStatementBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleManagedRuleGroupStatementModel](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrName: schema.StringAttribute{
					Required:    true,
					Description: "Name of the managed rule group.",
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 128),
					},
				},
				"vendor_name": schema.StringAttribute{
					Required:    true,
					Description: "Name of the managed rule group vendor.",
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 128),
					},
				},
				names.AttrVersion: schema.StringAttribute{
					Optional:    true,
					Description: "Version of the managed rule group.",
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 128),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"managed_rule_group_configs": managedRuleGroupConfigsBlock(ctx),
				"rule_action_override":       ruleActionOverrideBlock(ctx),
				"scope_down_statement":       scopeDownStatementBlock(ctx),
			},
		},
		Description: "Managed rule group statement.",
	}
}

func regexPatternSetReferenceStatementBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleRegexPatternSetReferenceStatementModel](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrARN: schema.StringAttribute{
					Required:    true,
					Description: "ARN of the RegexPatternSet (20-2048 characters).",
					Validators: []validator.String{
						stringvalidator.LengthBetween(20, 2048),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"field_to_match":      fieldToMatchBlock(ctx),
				"text_transformation": textTransformationBlock(ctx),
			},
		},
		Description: "Rule statement used to search web request components for matches with regular expressions from a RegexPatternSet.",
	}
}

func rateBasedStatementBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleRateBasedStatementModel](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"limit": schema.Int64Attribute{
					Required:    true,
					Description: "Rate limit threshold (10-2000000000).",
					Validators: []validator.Int64{
						int64validator.Between(10, 2000000000),
					},
				},
				"aggregate_key_type": schema.StringAttribute{
					CustomType:  fwtypes.StringEnumType[awstypes.RateBasedStatementAggregateKeyType](),
					Required:    true,
					Description: "Setting that indicates how to aggregate the request counts. Valid values: IP, FORWARDED_IP, CUSTOM_KEYS, CONSTANT.",
				},
				"evaluation_window_sec": schema.Int64Attribute{
					Optional:    true,
					Computed:    true,
					Description: "Time window for AWS WAF to use to check the rate (60, 120, 300, 600). Default: 300.",
					Validators: []validator.Int64{
						int64validator.OneOf(60, 120, 300, 600),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"custom_keys":          rateBasedStatementCustomKeysBlock(ctx),
				"forwarded_ip_config":  forwardedIPConfigBlock(ctx),
				"scope_down_statement": scopeDownStatementBlock(ctx),
			},
		},
		Description: "Rate-based statement.",
	}
}

func byteMatchStatementBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleByteMatchStatementModel](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"search_string": schema.StringAttribute{
					Required:    true,
					Description: "String value to search for within the request.",
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 200),
					},
				},
				"positional_constraint": schema.StringAttribute{
					Required:    true,
					Description: "Area within the portion of a web request that you want AWS WAF to search for SearchString.",
					Validators: []validator.String{
						stringvalidator.OneOf("EXACTLY", "STARTS_WITH", "ENDS_WITH", "CONTAINS", "CONTAINS_WORD"),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"field_to_match":      fieldToMatchBlock(ctx),
				"text_transformation": textTransformationBlock(ctx),
			},
		},
		Description: "Byte match statement.",
	}
}

func sqliMatchStatementBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleSqliMatchStatementModel](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"sensitivity_level": schema.StringAttribute{
					CustomType:  fwtypes.StringEnumType[awstypes.SensitivityLevel](),
					Optional:    true,
					Computed:    true,
					Default:     stringdefault.StaticString(string(awstypes.SensitivityLevelLow)),
					Description: "Sensitivity level for detecting SQL injection attacks. Valid values: `HIGH`, `LOW`. Defaults to `LOW`.",
				},
			},
			Blocks: map[string]schema.Block{
				"field_to_match":      fieldToMatchBlock(ctx),
				"text_transformation": textTransformationBlock(ctx),
			},
		},
		Description: "SQL injection match statement.",
	}
}

func xssMatchStatementBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleXssMatchStatementModel](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"field_to_match":      fieldToMatchBlock(ctx),
				"text_transformation": textTransformationBlock(ctx),
			},
		},
		Description: "Cross-site scripting match statement.",
	}
}

func sizeConstraintStatementBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleSizeConstraintStatementModel](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"comparison_operator": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.ComparisonOperator](),
					Required:   true,
				},
				names.AttrSize: schema.Int64Attribute{
					Required: true,
					Validators: []validator.Int64{
						int64validator.AtLeast(0),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"field_to_match":      fieldToMatchBlock(ctx),
				"text_transformation": textTransformationBlock(ctx),
			},
		},
		Description: "Size constraint statement.",
	}
}

func regexMatchStatementBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleRegexMatchStatementModel](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"regex_string": schema.StringAttribute{
					Required:    true,
					Description: "Regular expression string (1-512 characters).",
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 512),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"field_to_match":      fieldToMatchBlock(ctx),
				"text_transformation": textTransformationBlock(ctx),
			},
		},
		Description: "Rule statement used to search web request components for a match against a single regular expression.",
	}
}

func labelMatchStatementBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleLabelMatchStatementModel](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrKey: schema.StringAttribute{
					Required:    true,
					Description: "String to match against. Must be 1-1024 characters and match pattern ^[0-9A-Za-z_\\-:]+$.",
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 1024),
						stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z_\-:]+$`), "must contain only alphanumeric characters, underscores, hyphens, and colons"),
					},
				},
				names.AttrScope: schema.StringAttribute{
					CustomType:  fwtypes.StringEnumType[awstypes.LabelMatchScope](),
					Required:    true,
					Description: "Specify whether to match using the label name or just the namespace. Valid values: LABEL, NAMESPACE.",
				},
			},
		},
		Description: "Label match statement.",
	}
}

func asnMatchStatementBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleAsnMatchStatementModel](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"asn_list": schema.ListAttribute{
					ElementType: types.Int64Type,
					Required:    true,
					Validators: []validator.List{
						listvalidator.SizeAtLeast(1),
						listvalidator.SizeAtMost(100),
					},
					Description: "List of ASN numbers (0-4294967295).",
				},
			},
			Blocks: map[string]schema.Block{
				"forwarded_ip_config": forwardedIPConfigBlock(ctx),
			},
		},
		Description: "ASN match statement.",
	}
}

func managedRuleGroupConfigsBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType:   fwtypes.NewListNestedObjectTypeOf[webACLRuleManagedRuleGroupConfigModel](ctx),
		NestedObject: schema.NestedBlockObject{},
		Description:  "Managed rule group configurations.",
	}
}

func ruleActionOverrideBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleRuleActionOverrideModel](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(100)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrName: schema.StringAttribute{
					Required:    true,
					Description: "Name of the rule to override (1-128 characters).",
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 128),
						stringvalidator.RegexMatches(regexache.MustCompile(`^[\w\-]+$`), "must contain only word characters and hyphens"),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"action_to_use": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleActionModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{
						Blocks: map[string]schema.Block{
							"allow": schema.ListNestedBlock{
								CustomType:   fwtypes.NewListNestedObjectTypeOf[webACLRuleEmptyModel](ctx),
								Validators:   []validator.List{listvalidator.SizeAtMost(1)},
								NestedObject: schema.NestedBlockObject{},
							},
							"block": schema.ListNestedBlock{
								CustomType:   fwtypes.NewListNestedObjectTypeOf[webACLRuleBlockActionModel](ctx),
								Validators:   []validator.List{listvalidator.SizeAtMost(1)},
								NestedObject: schema.NestedBlockObject{},
							},
							"count": schema.ListNestedBlock{
								CustomType:   fwtypes.NewListNestedObjectTypeOf[webACLRuleEmptyModel](ctx),
								Validators:   []validator.List{listvalidator.SizeAtMost(1)},
								NestedObject: schema.NestedBlockObject{},
							},
						},
					},
					Description: "Override action to use in place of the configured action of the rule in the rule group.",
				},
			},
		},
		Description: "Action settings to use in place of rule actions configured inside the rule group.",
	}
}

func scopeDownStatementBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleScopeDownStatementModel](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"ip_set_reference_statement": ipSetReferenceStatementBlock(ctx),
				"geo_match_statement":        geoMatchStatementBlock(ctx),
				"byte_match_statement":       byteMatchStatementBlock(ctx),
				"sqli_match_statement":       sqliMatchStatementBlock(ctx),
				"xss_match_statement":        xssMatchStatementBlock(ctx),
				"size_constraint_statement":  sizeConstraintStatementBlock(ctx),
				"regex_match_statement":      regexMatchStatementBlock(ctx),
				"label_match_statement":      labelMatchStatementBlock(ctx),
				"asn_match_statement":        asnMatchStatementBlock(ctx),
			},
		},
		Description: "Scope down statement for managed rule groups.",
	}
}

func customRequestHandlingBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleCustomRequestHandlingModel](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"insert_header": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleCustomHeaderModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtLeast(1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrName:  schema.StringAttribute{Required: true},
							names.AttrValue: schema.StringAttribute{Required: true},
						},
					},
				},
			},
		},
	}
}

func customResponseBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleCustomResponseModel](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"response_code": schema.Int32Attribute{
					Required: true,
				},
				"custom_response_body_key": schema.StringAttribute{
					Optional: true,
				},
			},
			Blocks: map[string]schema.Block{
				"response_header": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleCustomHeaderModel](ctx),
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrName:  schema.StringAttribute{Required: true},
							names.AttrValue: schema.StringAttribute{Required: true},
						},
					},
				},
			},
		},
	}
}
func forwardedIPConfigBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleForwardedIPConfigModel](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"fallback_behavior": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.OneOf("MATCH", "NO_MATCH"),
					},
				},
				"header_name": schema.StringAttribute{
					Required: true,
				},
			},
		},
	}
}
func fieldToMatchBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleFieldToMatchModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
			listvalidator.SizeAtLeast(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"all_query_arguments": schema.ListNestedBlock{
					CustomType:   fwtypes.NewListNestedObjectTypeOf[webACLRuleTrulyEmptyModel](ctx),
					Validators:   []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{},
				},
				"body": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleBodyModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"oversize_handling": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.OversizeHandling](),
								Optional:   true,
								Computed:   true,
								Default:    stringdefault.StaticString(string(awstypes.OversizeHandlingContinue)),
							},
						},
					},
				},
				"cookies": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleCookiesModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"match_scope": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.MapMatchScope](),
								Required:   true,
							},
							"oversize_handling": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.OversizeHandling](),
								Required:   true,
							},
						},
						Blocks: map[string]schema.Block{
							"match_pattern": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleCookiesMatchPatternModel](ctx),
								Validators: []validator.List{listvalidator.SizeAtLeast(1)},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"excluded_cookies": schema.ListAttribute{
											CustomType:  fwtypes.ListOfStringType,
											ElementType: types.StringType,
											Optional:    true,
										},
										"included_cookies": schema.ListAttribute{
											CustomType:  fwtypes.ListOfStringType,
											ElementType: types.StringType,
											Optional:    true,
										},
									},
									Blocks: map[string]schema.Block{
										"all": schema.ListNestedBlock{
											CustomType:   fwtypes.NewListNestedObjectTypeOf[webACLRuleTrulyEmptyModel](ctx),
											Validators:   []validator.List{listvalidator.SizeAtMost(1)},
											NestedObject: schema.NestedBlockObject{},
										},
									},
								},
							},
						},
					},
				},
				"header_order": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleHeaderOrderModel](ctx),
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"oversize_handling": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.OversizeHandling](),
								Required:   true,
							},
						},
					},
				},
				"headers": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleHeadersModel](ctx),
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"match_scope": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.MapMatchScope](),
								Required:   true,
							},
							"oversize_handling": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.OversizeHandling](),
								Required:   true,
							},
						},
						Blocks: map[string]schema.Block{
							"match_pattern": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleHeadersMatchPatternModel](ctx),
								Validators: []validator.List{listvalidator.SizeAtMost(1), listvalidator.SizeAtLeast(1)},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"excluded_headers": schema.ListAttribute{
											CustomType:  fwtypes.ListOfStringType,
											ElementType: types.StringType,
											Optional:    true,
										},
										"included_headers": schema.ListAttribute{
											CustomType:  fwtypes.ListOfStringType,
											ElementType: types.StringType,
											Optional:    true,
										},
									},
									Blocks: map[string]schema.Block{
										"all": schema.ListNestedBlock{
											CustomType:   fwtypes.NewListNestedObjectTypeOf[webACLRuleTrulyEmptyModel](ctx),
											Validators:   []validator.List{listvalidator.SizeAtMost(1)},
											NestedObject: schema.NestedBlockObject{},
										},
									},
								},
							},
						},
					},
				},
				"ja3_fingerprint": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleJAFingerprintModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"fallback_behavior": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.FallbackBehavior](),
								Required:   true,
							},
						},
					},
				},
				"ja4_fingerprint": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleJAFingerprintModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"fallback_behavior": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.FallbackBehavior](),
								Required:   true,
							},
						},
					},
				},
				"json_body": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleJsonBodyModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"invalid_fallback_behavior": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.BodyParsingFallbackBehavior](),
								Optional:   true,
							},
							"match_scope": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.JsonMatchScope](),
								Required:   true,
							},
							"oversize_handling": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.OversizeHandling](),
								Optional:   true,
								Computed:   true,
								Default:    stringdefault.StaticString(string(awstypes.OversizeHandlingContinue)),
							},
						},
						Blocks: map[string]schema.Block{
							"match_pattern": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleJsonBodyMatchPatternModel](ctx),
								Validators: []validator.List{listvalidator.SizeAtMost(1), listvalidator.SizeAtLeast(1)},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"included_paths": schema.ListAttribute{
											CustomType:  fwtypes.ListOfStringType,
											ElementType: types.StringType,
											Optional:    true,
										},
									},
									Blocks: map[string]schema.Block{
										"all": schema.ListNestedBlock{
											CustomType:   fwtypes.NewListNestedObjectTypeOf[webACLRuleTrulyEmptyModel](ctx),
											Validators:   []validator.List{listvalidator.SizeAtMost(1)},
											NestedObject: schema.NestedBlockObject{},
										},
									},
								},
							},
						},
					},
				},
				"method": schema.ListNestedBlock{
					CustomType:   fwtypes.NewListNestedObjectTypeOf[webACLRuleTrulyEmptyModel](ctx),
					Validators:   []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{},
				},
				"query_string": schema.ListNestedBlock{
					CustomType:   fwtypes.NewListNestedObjectTypeOf[webACLRuleTrulyEmptyModel](ctx),
					Validators:   []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{},
				},
				"single_header": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleSingleHeaderModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrName: schema.StringAttribute{
								Required: true,
							},
						},
					},
				},
				"single_query_argument": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleSingleQueryArgumentModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrName: schema.StringAttribute{
								Required: true,
							},
						},
					},
				},
				"uri_fragment": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleUriFragmentModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"fallback_behavior": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.FallbackBehavior](),
								Optional:   true,
								Computed:   true,
							},
						},
					},
				},
				"uri_path": schema.ListNestedBlock{
					CustomType:   fwtypes.NewListNestedObjectTypeOf[webACLRuleTrulyEmptyModel](ctx),
					Validators:   []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{},
				},
			},
		},
	}
}

func textTransformationBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleTextTransformModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
			listvalidator.SizeAtMost(10),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrPriority: schema.Int32Attribute{
					Required: true,
				},
				names.AttrType: schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.OneOf("NONE", "COMPRESS_WHITE_SPACE", "HTML_ENTITY_DECODE", "LOWERCASE", "CMD_LINE", "URL_DECODE", "BASE64_DECODE", "HEX_DECODE", "MD5", "REPLACE_COMMENTS", "ESCAPE_SEQ_DECODE", "SQL_HEX_DECODE", "CSS_DECODE", "JS_DECODE", "NORMALIZE_PATH", "NORMALIZE_PATH_WIN", "REMOVE_NULLS", "REPLACE_NULLS", "BASE64_DECODE_EXT", "URL_DECODE_UNI", "UTF8_TO_UNICODE"),
					},
				},
			},
		},
	}
}
func rateBasedStatementCustomKeysBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleRateBasedStatementCustomKeyModel](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(5)},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"asn": schema.ListNestedBlock{
					CustomType:   fwtypes.NewListNestedObjectTypeOf[webACLRuleTrulyEmptyModel](ctx),
					Validators:   []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{},
				},
				"cookie": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleRateBasedStatementCustomKeyCookieModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrName: schema.StringAttribute{
								Required: true,
							},
						},
						Blocks: map[string]schema.Block{
							"text_transformation": textTransformationBlock(ctx),
						},
					},
				},
				"forwarded_ip": schema.ListNestedBlock{
					CustomType:   fwtypes.NewListNestedObjectTypeOf[webACLRuleTrulyEmptyModel](ctx),
					Validators:   []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{},
				},
				names.AttrHeader: schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleRateBasedStatementCustomKeyHeaderModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrName: schema.StringAttribute{
								Required: true,
							},
						},
						Blocks: map[string]schema.Block{
							"text_transformation": textTransformationBlock(ctx),
						},
					},
				},
				"http_method": schema.ListNestedBlock{
					CustomType:   fwtypes.NewListNestedObjectTypeOf[webACLRuleTrulyEmptyModel](ctx),
					Validators:   []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{},
				},
				"ip": schema.ListNestedBlock{
					CustomType:   fwtypes.NewListNestedObjectTypeOf[webACLRuleTrulyEmptyModel](ctx),
					Validators:   []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{},
				},
				"ja3_fingerprint": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleRateBasedStatementCustomKeyJAFingerprintModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"fallback_behavior": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.FallbackBehavior](),
								Required:   true,
							},
						},
					},
				},
				"ja4_fingerprint": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleRateBasedStatementCustomKeyJAFingerprintModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"fallback_behavior": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.FallbackBehavior](),
								Required:   true,
							},
						},
					},
				},
				"label_namespace": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleRateBasedStatementCustomKeyLabelNamespaceModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrNamespace: schema.StringAttribute{
								Required: true,
							},
						},
					},
				},
				"query_argument": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleRateBasedStatementCustomKeyQueryArgumentModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrName: schema.StringAttribute{
								Required: true,
							},
						},
						Blocks: map[string]schema.Block{
							"text_transformation": textTransformationBlock(ctx),
						},
					},
				},
				"query_string": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleRateBasedStatementCustomKeyQueryStringModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{
						Blocks: map[string]schema.Block{
							"text_transformation": textTransformationBlock(ctx),
						},
					},
				},
				"uri_path": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleRateBasedStatementCustomKeyUriPathModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{
						Blocks: map[string]schema.Block{
							"text_transformation": textTransformationBlock(ctx),
						},
					},
				},
			},
		},
	}
}

// andStatementBlock and related functions are generated in web_acl_rule_statement_models_gen.go
