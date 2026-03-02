// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func statementBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleStatementModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
			listvalidator.SizeAtLeast(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"ip_set_reference_statement":            ipSetReferenceStatementBlock(ctx),
				"geo_match_statement":                   geoMatchStatementBlock(ctx),
				"rule_group_reference_statement":        ruleGroupReferenceStatementBlock(ctx),
				"managed_rule_group_statement":          managedRuleGroupStatementBlock(ctx),
				"regex_pattern_set_reference_statement": regexPatternSetReferenceStatementBlock(ctx),
				"rate_based_statement":                  rateBasedStatementBlock(ctx),
				"byte_match_statement":                  byteMatchStatementBlock(ctx),
				"sqli_match_statement":                  sqliMatchStatementBlock(ctx),
				"xss_match_statement":                   xssMatchStatementBlock(ctx),
				"size_constraint_statement":             sizeConstraintStatementBlock(ctx),
				"regex_match_statement":                 regexMatchStatementBlock(ctx),
				"label_match_statement":                 labelMatchStatementBlock(ctx),
				"asn_match_statement":                   asnMatchStatementBlock(ctx),
			},
		},
		Description: "Rule statement. Exactly one statement type must be specified.",
	}
}

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
					Description: "ARN of the rule group to reference.",
				},
			},
		},
		Description: "Rule group reference statement.",
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
					Description: "ARN of the regex pattern set to reference.",
				},
			},
		},
		Description: "Regex pattern set reference statement.",
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
					Description: "Rate limit threshold.",
				},
				"aggregate_key_type": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Setting that indicates how to aggregate the request counts.",
				},
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
				},
			},
		},
		Description: "Byte match statement.",
	}
}

func sqliMatchStatementBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType:   fwtypes.NewListNestedObjectTypeOf[webACLRuleSqliMatchStatementModel](ctx),
		Validators:   []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{},
		Description:  "SQL injection match statement.",
	}
}

func xssMatchStatementBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType:   fwtypes.NewListNestedObjectTypeOf[webACLRuleXssMatchStatementModel](ctx),
		Validators:   []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{},
		Description:  "Cross-site scripting match statement.",
	}
}

func sizeConstraintStatementBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType:   fwtypes.NewListNestedObjectTypeOf[webACLRuleSizeConstraintStatementModel](ctx),
		Validators:   []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{},
		Description:  "Size constraint statement.",
	}
}

func regexMatchStatementBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType:   fwtypes.NewListNestedObjectTypeOf[webACLRuleRegexMatchStatementModel](ctx),
		Validators:   []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{},
		Description:  "Regex match statement.",
	}
}

func labelMatchStatementBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType:   fwtypes.NewListNestedObjectTypeOf[webACLRuleLabelMatchStatementModel](ctx),
		Validators:   []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{},
		Description:  "Label match statement.",
	}
}

func asnMatchStatementBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType:   fwtypes.NewListNestedObjectTypeOf[webACLRuleAsnMatchStatementModel](ctx),
		Validators:   []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{},
		Description:  "ASN match statement.",
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
		CustomType:   fwtypes.NewListNestedObjectTypeOf[webACLRuleRuleActionOverrideModel](ctx),
		NestedObject: schema.NestedBlockObject{},
		Description:  "Rule action overrides.",
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
