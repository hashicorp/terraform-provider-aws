// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"math"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var listOfEmptyObjectSchema *schema.Schema = &schema.Schema{
	Type:     schema.TypeList,
	Optional: true,
	MaxItems: 1,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{},
	},
}

func emptySchema() *schema.Schema {
	return listOfEmptyObjectSchema
}

func ruleLabelsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 1024),
						validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_\-:]+$`), "must contain only alphanumeric, underscore, hyphen, and colon characters"),
					),
				},
			},
		},
	}
}

func ruleGroupRootStatementSchema(level int) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"and_statement":                         statementSchema(level),
				"byte_match_statement":                  byteMatchStatementSchema(),
				"geo_match_statement":                   geoMatchStatementSchema(),
				"ip_set_reference_statement":            ipSetReferenceStatementSchema(),
				"label_match_statement":                 labelMatchStatementSchema(),
				"not_statement":                         statementSchema(level),
				"or_statement":                          statementSchema(level),
				"rate_based_statement":                  rateBasedStatementSchema(level),
				"regex_match_statement":                 regexMatchStatementSchema(),
				"regex_pattern_set_reference_statement": regexPatternSetReferenceStatementSchema(),
				"size_constraint_statement":             sizeConstraintSchema(),
				"sqli_match_statement":                  sqliMatchStatementSchema(),
				"xss_match_statement":                   xssMatchStatementSchema(),
			},
		},
	}
}

func statementSchema(level int) *schema.Schema {
	if level > 1 {
		return &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"statement": {
						Type:     schema.TypeList,
						Required: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"and_statement":                         statementSchema(level - 1),
								"byte_match_statement":                  byteMatchStatementSchema(),
								"geo_match_statement":                   geoMatchStatementSchema(),
								"ip_set_reference_statement":            ipSetReferenceStatementSchema(),
								"label_match_statement":                 labelMatchStatementSchema(),
								"not_statement":                         statementSchema(level - 1),
								"or_statement":                          statementSchema(level - 1),
								"regex_match_statement":                 regexMatchStatementSchema(),
								"regex_pattern_set_reference_statement": regexPatternSetReferenceStatementSchema(),
								"size_constraint_statement":             sizeConstraintSchema(),
								"sqli_match_statement":                  sqliMatchStatementSchema(),
								"xss_match_statement":                   xssMatchStatementSchema(),
							},
						},
					},
				},
			},
		}
	}

	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"statement": {
					Type:     schema.TypeList,
					Required: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"byte_match_statement":                  byteMatchStatementSchema(),
							"geo_match_statement":                   geoMatchStatementSchema(),
							"ip_set_reference_statement":            ipSetReferenceStatementSchema(),
							"label_match_statement":                 labelMatchStatementSchema(),
							"regex_match_statement":                 regexMatchStatementSchema(),
							"regex_pattern_set_reference_statement": regexPatternSetReferenceStatementSchema(),
							"size_constraint_statement":             sizeConstraintSchema(),
							"sqli_match_statement":                  sqliMatchStatementSchema(),
							"xss_match_statement":                   xssMatchStatementSchema(),
						},
					},
				},
			},
		},
	}
}

func byteMatchStatementSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"field_to_match": fieldToMatchSchema(),
				"positional_constraint": {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[awstypes.PositionalConstraint](),
				},
				"search_string": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 200),
				},
				"text_transformation": textTransformationSchema(),
			},
		},
	}
}

func geoMatchStatementSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"country_codes": {
					Type:     schema.TypeList,
					Required: true,
					MinItems: 1,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"forwarded_ip_config": forwardedIPConfigSchema(),
			},
		},
	}
}

func ipSetReferenceStatementSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrARN: {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: verify.ValidARN,
				},
				"ip_set_forwarded_ip_config": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"fallback_behavior": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[awstypes.FallbackBehavior](),
							},
							"header_name": {
								Type:     schema.TypeString,
								Required: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 255),
									validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z-]+$`), "must contain only alphanumeric and hyphen characters"),
								),
							},
							"position": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[awstypes.ForwardedIPPosition](),
							},
						},
					},
				},
			},
		},
	}
}

func labelMatchStatementSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrKey: {
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 1024),
						validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_\-:]+$`), "must contain only alphanumeric, underscore, hyphen, and colon characters"),
					),
				},
				names.AttrScope: {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[awstypes.LabelMatchScope](),
				},
			},
		},
	}
}

func regexMatchStatementSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"regex_string": {
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 512),
						validation.StringIsValidRegExp,
					),
				},
				"field_to_match":      fieldToMatchSchema(),
				"text_transformation": textTransformationSchema(),
			},
		},
	}
}

func regexPatternSetReferenceStatementSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrARN: {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: verify.ValidARN,
				},
				"field_to_match":      fieldToMatchSchema(),
				"text_transformation": textTransformationSchema(),
			},
		},
	}
}

func sizeConstraintSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"comparison_operator": {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[awstypes.ComparisonOperator](),
				},
				"field_to_match": fieldToMatchSchema(),
				names.AttrSize: {
					Type:         schema.TypeInt,
					Required:     true,
					ValidateFunc: validation.IntBetween(0, math.MaxInt32),
				},
				"text_transformation": textTransformationSchema(),
			},
		},
	}
}

func sqliMatchStatementSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"field_to_match": fieldToMatchSchema(),
				"sensitivity_level": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[awstypes.SensitivityLevel](),
				},
				"text_transformation": textTransformationSchema(),
			},
		},
	}
}

func xssMatchStatementSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"field_to_match":      fieldToMatchSchema(),
				"text_transformation": textTransformationSchema(),
			},
		},
	}
}

func fieldToMatchBaseSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"all_query_arguments": emptySchema(),
			"body":                bodySchema(),
			"cookies":             cookiesSchema(),
			"header_order":        headerOrderSchema(),
			"headers":             headersSchema(),
			"ja3_fingerprint":     ja3fingerprintSchema(),
			"json_body":           jsonBodySchema(),
			"method":              emptySchema(),
			"query_string":        emptySchema(),
			"single_header": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 40),
								// The value is returned in lower case by the API.
								// Trying to solve it with StateFunc and/or DiffSuppressFunc resulted in hash problem of the rule field or didn't work.
								validation.StringMatch(regexache.MustCompile(`^[0-9a-z_-]+$`), "must contain only lowercase alphanumeric characters, underscores, and hyphens"),
							),
						},
					},
				},
			},
			"single_query_argument": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 30),
								// The value is returned in lower case by the API.
								// Trying to solve it with StateFunc and/or DiffSuppressFunc resulted in hash problem of the rule field or didn't work.
								validation.StringMatch(regexache.MustCompile(`^[0-9a-z_-]+$`), "must contain only lowercase alphanumeric characters, underscores, and hyphens"),
							),
						},
					},
				},
			},
			"uri_path": emptySchema(),
		},
	}
}

func fieldToMatchSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem:     fieldToMatchBaseSchema(),
	}
}

func jsonBodySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"invalid_fallback_behavior": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[awstypes.BodyParsingFallbackBehavior](),
				},
				"match_pattern": jsonBodyMatchPatternSchema(),
				"match_scope": {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[awstypes.JsonMatchScope](),
				},
				"oversize_handling": oversizeHandlingOptionalSchema(string(awstypes.OversizeHandlingContinue)),
			},
		},
	}
}

func jsonBodyMatchPatternSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"all": emptySchema(),
				"included_paths": {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					Elem: &schema.Schema{
						Type: schema.TypeString,
						ValidateFunc: validation.All(
							validation.StringLenBetween(1, 512),
							validation.StringMatch(regexache.MustCompile(`(/)|(/(([^~])|(~[01]))+)`), "must be a valid JSON pointer")),
					},
				},
			},
		},
	}
}

func forwardedIPConfigSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"fallback_behavior": {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[awstypes.FallbackBehavior](),
				},
				"header_name": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func textTransformationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Required: true,
		MinItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrPriority: {
					Type:     schema.TypeInt,
					Required: true,
				},
				names.AttrType: {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[awstypes.TextTransformationType](),
				},
			},
		},
	}
}

func visibilityConfigSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cloudwatch_metrics_enabled": {
					Type:     schema.TypeBool,
					Required: true,
				},
				names.AttrMetricName: {
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 128),
						validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_-]+$`), "must contain only alphanumeric hyphen and underscore characters"),
					),
				},
				"sampled_requests_enabled": {
					Type:     schema.TypeBool,
					Required: true,
				},
			},
		},
	}
}

func associationConfigSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"request_body": requestBodySchema(),
			},
		},
	}
}

func requestBodySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"api_gateway": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"default_size_inspection_limit": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[awstypes.SizeInspectionLimit](),
							},
						},
					},
				},
				"app_runner_service": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"default_size_inspection_limit": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[awstypes.SizeInspectionLimit](),
							},
						},
					},
				},
				"cloudfront": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"default_size_inspection_limit": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[awstypes.SizeInspectionLimit](),
							},
						},
					},
				},
				"cognito_user_pool": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"default_size_inspection_limit": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[awstypes.SizeInspectionLimit](),
							},
						},
					},
				},
				"verified_access_instance": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"default_size_inspection_limit": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[awstypes.SizeInspectionLimit](),
							},
						},
					},
				},
			},
		},
	}
}

func allowConfigSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"custom_request_handling": customRequestHandlingSchema(),
			},
		},
	}
}

func captchaConfigSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"custom_request_handling": customRequestHandlingSchema(),
			},
		},
	}
}

func outerCaptchaConfigSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"immunity_time_property": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"immunity_time": {
								Type:     schema.TypeInt,
								Optional: true,
							},
						},
					},
				},
			},
		},
	}
}

func challengeConfigSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"custom_request_handling": customRequestHandlingSchema(),
			},
		},
	}
}

func outerChallengeConfigSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"immunity_time_property": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"immunity_time": {
								Type:     schema.TypeInt,
								Optional: true,
							},
						},
					},
				},
			},
		},
	}
}

func countConfigSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"custom_request_handling": customRequestHandlingSchema(),
			},
		},
	}
}

func blockConfigSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"custom_response": customResponseSchema(),
			},
		},
	}
}

func customRequestHandlingSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"insert_header": {
					Type:     schema.TypeSet,
					Required: true,
					MinItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrName: {
								Type:     schema.TypeString,
								Required: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 64),
									validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_$.-]+$`), "must contain only alphanumeric, hyphen, underscore, dot and $ characters"),
								),
							},
							names.AttrValue: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringLenBetween(1, 255),
							},
						},
					},
				},
			},
		},
	}
}

func customResponseSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"custom_response_body_key": {
					Type:     schema.TypeString,
					Optional: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 128),
						validation.StringMatch(regexache.MustCompile(`^[\w\-]+$`), "must contain only alphanumeric, hyphen, and underscore characters"),
					),
				},
				"response_code": {
					Type:         schema.TypeInt,
					Required:     true,
					ValidateFunc: validation.IntBetween(200, 600),
				},
				"response_header": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrName: {
								Type:     schema.TypeString,
								Required: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 64),
									validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_$.-]+$`), "must contain only alphanumeric, hyphen, underscore, dot and $ characters"),
								),
							},
							names.AttrValue: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringLenBetween(1, 255),
							},
						},
					},
				},
			},
		},
	}
}

func customResponseBodySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrKey: {
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 128),
						validation.StringMatch(regexache.MustCompile(`^[\w\-]+$`), "must contain only alphanumeric, hyphen, and underscore characters"),
					),
				},
				names.AttrContent: {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 10240),
				},
				names.AttrContentType: {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[awstypes.ResponseContentType](),
				},
			},
		},
	}
}

func cookiesSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"match_scope":       matchScopeSchema(),
				"oversize_handling": oversizeHandlingRequiredSchema(),
				"match_pattern":     cookiesMatchPatternSchema(),
			},
		},
	}
}

func cookiesMatchPatternSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"all": emptySchema(),
				"excluded_cookies": {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"included_cookies": {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
			},
		},
	}
}

func ja3fingerprintSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"fallback_behavior": {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[awstypes.FallbackBehavior](),
				},
			},
		},
	}
}

func bodySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"oversize_handling": oversizeHandlingOptionalSchema(string(awstypes.OversizeHandlingContinue)),
			},
		},
	}
}

func oversizeHandlingOptionalSchema(defaultValue string) *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Default:          defaultValue,
		ValidateDiagFunc: enum.Validate[awstypes.OversizeHandling](),
	}
}

func oversizeHandlingRequiredSchema() *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeString,
		Required:         true,
		ValidateDiagFunc: enum.Validate[awstypes.OversizeHandling](),
	}
}

func matchScopeSchema() *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeString,
		Required:         true,
		ValidateDiagFunc: enum.Validate[awstypes.MapMatchScope](),
	}
}

func headerOrderSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"oversize_handling": oversizeHandlingRequiredSchema(),
			},
		},
	}
}

func headersSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"match_pattern": {
					Type:     schema.TypeList,
					Required: true,
					MaxItems: 1,
					MinItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"all":              emptySchema(),
							"excluded_headers": headersMatchPatternBaseSchema(),
							"included_headers": headersMatchPatternBaseSchema(),
						},
					},
				},
				"match_scope":       matchScopeSchema(),
				"oversize_handling": oversizeHandlingRequiredSchema(),
			},
		},
	}
}

func headersMatchPatternBaseSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 199,
		Elem: &schema.Schema{
			Type: schema.TypeString,
			ValidateFunc: validation.All(
				validation.StringLenBetween(1, 64),
				validation.StringMatch(regexache.MustCompile(`.*\S.*`), ""),
			),
		},
	}
}

func webACLRootStatementSchema(level int) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"and_statement":                         statementSchema(level),
				"byte_match_statement":                  byteMatchStatementSchema(),
				"geo_match_statement":                   geoMatchStatementSchema(),
				"ip_set_reference_statement":            ipSetReferenceStatementSchema(),
				"label_match_statement":                 labelMatchStatementSchema(),
				"managed_rule_group_statement":          managedRuleGroupStatementSchema(level),
				"not_statement":                         statementSchema(level),
				"or_statement":                          statementSchema(level),
				"rate_based_statement":                  rateBasedStatementSchema(level),
				"regex_match_statement":                 regexMatchStatementSchema(),
				"regex_pattern_set_reference_statement": regexPatternSetReferenceStatementSchema(),
				"rule_group_reference_statement":        ruleGroupReferenceStatementSchema(),
				"size_constraint_statement":             sizeConstraintSchema(),
				"sqli_match_statement":                  sqliMatchStatementSchema(),
				"xss_match_statement":                   xssMatchStatementSchema(),
			},
		},
	}
}

func managedRuleGroupStatementSchema(level int) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"managed_rule_group_configs": managedRuleGroupConfigSchema(),
				names.AttrName: {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 128),
				},
				"rule_action_override": ruleActionOverrideSchema(),
				"scope_down_statement": scopeDownStatementSchema(level - 1),
				"vendor_name": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 128),
				},
				names.AttrVersion: {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(1, 128),
				},
			},
		},
	}
}

func rateBasedStatementSchema(level int) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"aggregate_key_type": {
					Type:             schema.TypeString,
					Optional:         true,
					Default:          awstypes.RateBasedStatementAggregateKeyTypeIp,
					ValidateDiagFunc: enum.Validate[awstypes.RateBasedStatementAggregateKeyType](),
				},
				"custom_key": {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 5,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"cookie": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrName: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringLenBetween(1, 64),
										},
										"text_transformation": textTransformationSchema(),
									},
								},
							},
							"forwarded_ip": emptySchema(),
							"http_method":  emptySchema(),
							names.AttrHeader: {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrName: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringLenBetween(1, 64),
										},
										"text_transformation": textTransformationSchema(),
									},
								},
							},
							"ip": emptySchema(),
							"label_namespace": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrNamespace: {
											Type:     schema.TypeString,
											Required: true,
											ValidateFunc: validation.All(
												validation.StringLenBetween(1, 1024),
												validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_\-:]+$`), "must contain only alphanumeric, underscore, hyphen, and colon characters"),
											),
										},
									},
								},
							},
							"query_argument": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrName: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringLenBetween(1, 64),
										},
										"text_transformation": textTransformationSchema(),
									},
								},
							},
							"query_string": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"text_transformation": textTransformationSchema(),
									},
								},
							},
							"uri_path": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"text_transformation": textTransformationSchema(),
									},
								},
							},
						},
					},
				},
				"evaluation_window_sec": {
					Type:         schema.TypeInt,
					Optional:     true,
					Default:      300,
					ValidateFunc: validation.IntInSlice([]int{60, 120, 300, 600}),
				},
				"forwarded_ip_config": forwardedIPConfigSchema(),
				"limit": {
					Type:         schema.TypeInt,
					Required:     true,
					ValidateFunc: validation.IntBetween(100, 2000000000),
				},
				"scope_down_statement": scopeDownStatementSchema(level - 1),
			},
		},
	}
}

func scopeDownStatementSchema(level int) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"and_statement":                         statementSchema(level),
				"byte_match_statement":                  byteMatchStatementSchema(),
				"geo_match_statement":                   geoMatchStatementSchema(),
				"label_match_statement":                 labelMatchStatementSchema(),
				"ip_set_reference_statement":            ipSetReferenceStatementSchema(),
				"not_statement":                         statementSchema(level),
				"or_statement":                          statementSchema(level),
				"regex_match_statement":                 regexMatchStatementSchema(),
				"regex_pattern_set_reference_statement": regexPatternSetReferenceStatementSchema(),
				"size_constraint_statement":             sizeConstraintSchema(),
				"sqli_match_statement":                  sqliMatchStatementSchema(),
				"xss_match_statement":                   xssMatchStatementSchema(),
			},
		},
	}
}

func ruleActionOverrideSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 100,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"action_to_use": actionToUseSchema(),
				names.AttrName: {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 128),
				},
			},
		},
	}
}

func managedRuleGroupConfigSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"aws_managed_rules_acfp_rule_set": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"enable_regex_in_path": {
								Type:     schema.TypeBool,
								Optional: true,
								Computed: true,
							},
							"creation_path": {
								Type:     schema.TypeString,
								Required: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 256),
									validation.StringMatch(regexache.MustCompile(`.*\S.*`), `must conform to pattern .*\S.* `),
								),
							},
							"registration_page_path": {
								Type:     schema.TypeString,
								Required: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 256),
									validation.StringMatch(regexache.MustCompile(`.*\S.*`), `must conform to pattern .*\S.* `),
								),
							},
							"request_inspection":  managedRuleGroupConfigACFPRequestInspectionSchema(),
							"response_inspection": managedRuleGroupConfigATPResponseInspectionSchema(),
						},
					},
				},
				"aws_managed_rules_atp_rule_set": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"enable_regex_in_path": {
								Type:     schema.TypeBool,
								Optional: true,
								Computed: true,
							},
							"login_path": {
								Type:     schema.TypeString,
								Required: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 256),
									validation.StringMatch(regexache.MustCompile(`.*\S.*`), `must conform to pattern .*\S.* `),
								),
							},
							"request_inspection":  managedRuleGroupConfigATPRequestInspectionSchema(),
							"response_inspection": managedRuleGroupConfigATPResponseInspectionSchema(),
						},
					},
				},
				"aws_managed_rules_bot_control_rule_set": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"enable_machine_learning": {
								Type:     schema.TypeBool,
								Optional: true,
								Default:  true,
							},
							"inspection_level": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[awstypes.InspectionLevel](),
							},
						},
					},
				},
				"login_path": {
					Type:     schema.TypeString,
					Optional: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 256),
						validation.StringMatch(regexache.MustCompile(`.*\S.*`), `must conform to pattern .*\S.* `),
					),
				},
				"password_field": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrIdentifier: {
								Type:     schema.TypeString,
								Required: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 512),
									validation.StringMatch(regexache.MustCompile(`.*\S.*`), `must conform to pattern .*\S.* `),
								),
							},
						},
					},
				},
				"payload_type": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[awstypes.PayloadType](),
				},
				"username_field": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrIdentifier: {
								Type:     schema.TypeString,
								Required: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 512),
									validation.StringMatch(regexache.MustCompile(`.*\S.*`), `must conform to pattern .*\S.* `),
								),
							},
						},
					},
				},
			},
		},
	}
}

func actionToUseSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"allow":     allowConfigSchema(),
				"block":     blockConfigSchema(),
				"captcha":   captchaConfigSchema(),
				"challenge": challengeConfigSchema(),
				"count":     countConfigSchema(),
			},
		},
	}
}

func ruleGroupReferenceStatementSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrARN: {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: verify.ValidARN,
				},
				"rule_action_override": ruleActionOverrideSchema(),
			},
		},
	}
}

func managedRuleGroupConfigACFPRequestInspectionSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"address_fields": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"identifiers": {
								Type:     schema.TypeList,
								Required: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
								MinItems: 1,
							},
						},
					},
				},
				"email_field": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrIdentifier: {
								Type:     schema.TypeString,
								Required: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 512),
									validation.StringMatch(regexache.MustCompile(`.*\S.*`), `must conform to pattern .*\S.* `),
								),
							},
						},
					},
				},
				"password_field": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrIdentifier: {
								Type:     schema.TypeString,
								Required: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 512),
									validation.StringMatch(regexache.MustCompile(`.*\S.*`), `must conform to pattern .*\S.* `),
								),
							},
						},
					},
				},
				"phone_number_fields": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"identifiers": {
								Type:     schema.TypeList,
								Required: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
								MinItems: 1,
							},
						},
					},
				},
				"payload_type": {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[awstypes.PayloadType](),
				},
				"username_field": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrIdentifier: {
								Type:     schema.TypeString,
								Required: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 512),
									validation.StringMatch(regexache.MustCompile(`.*\S.*`), `must conform to pattern .*\S.* `),
								),
							},
						},
					},
				},
			},
		},
	}
}

func managedRuleGroupConfigATPRequestInspectionSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"password_field": {
					Type:     schema.TypeList,
					Required: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrIdentifier: {
								Type:     schema.TypeString,
								Required: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 512),
									validation.StringMatch(regexache.MustCompile(`.*\S.*`), `must conform to pattern .*\S.* `),
								),
							},
						},
					},
				},
				"payload_type": {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[awstypes.PayloadType](),
				},
				"username_field": {
					Type:     schema.TypeList,
					Required: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrIdentifier: {
								Type:     schema.TypeString,
								Required: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 512),
									validation.StringMatch(regexache.MustCompile(`.*\S.*`), `must conform to pattern .*\S.* `),
								),
							},
						},
					},
				},
			},
		},
	}
}

func managedRuleGroupConfigATPResponseInspectionSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"body_contains": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"failure_strings": {
								Type:     schema.TypeSet,
								Required: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"success_strings": {
								Type:     schema.TypeSet,
								Required: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
						},
					},
				},
				names.AttrHeader: {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"failure_values": {
								Type:     schema.TypeSet,
								Required: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
								// TODO: ValidateFunc: length > 0
							},
							names.AttrName: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringLenBetween(1, 256),
							},
							"success_values": {
								Type:     schema.TypeSet,
								Required: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
								// TODO: ValidateFunc: length > 0
							},
						},
					},
				},
				names.AttrJSON: {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"failure_values": {
								Type:     schema.TypeSet,
								Required: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
								// TODO: ValidateFunc: length > 0
							},
							names.AttrIdentifier: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringLenBetween(1, 256),
							},
							"success_values": {
								Type:     schema.TypeSet,
								Required: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
								// TODO: ValidateFunc: length > 0
							},
						},
					},
				},
				names.AttrStatusCode: {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"failure_codes": {
								Type:     schema.TypeSet,
								Required: true,
								Elem:     &schema.Schema{Type: schema.TypeInt},
								// TODO: ValidateFunc: length > 0
							},
							"success_codes": {
								Type:     schema.TypeSet,
								Required: true,
								Elem:     &schema.Schema{Type: schema.TypeInt},
								// TODO: ValidateFunc: length > 0
							},
						},
					},
				},
			},
		},
	}
}
