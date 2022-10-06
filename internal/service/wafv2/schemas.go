package wafv2

import (
	"math"
	"regexp"

	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func emptySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
}

func emptySchemaDeprecated() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
		Deprecated: "Not supported by WAFv2 API",
	}
}

func ruleLabelsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 1024),
						validation.StringMatch(regexp.MustCompile(`^[0-9A-Za-z_\-:]+$`), "must contain only alphanumeric, underscore, hyphen, and colon characters"),
					),
				},
			},
		},
	}
}

func rootStatementSchema(level int) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"and_statement":                         statementSchema(level - 1),
				"byte_match_statement":                  byteMatchStatementSchema(),
				"geo_match_statement":                   geoMatchStatementSchema(),
				"ip_set_reference_statement":            ipSetReferenceStatementSchema(),
				"label_match_statement":                 labelMatchStatementSchema(),
				"not_statement":                         statementSchema(level - 1),
				"or_statement":                          statementSchema(level - 1),
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
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.StringInSlice([]string{
						wafv2.PositionalConstraintContains,
						wafv2.PositionalConstraintContainsWord,
						wafv2.PositionalConstraintEndsWith,
						wafv2.PositionalConstraintExactly,
						wafv2.PositionalConstraintStartsWith,
					}, false),
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
				"forwarded_ip_config": forwardedIPConfig(),
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
				"arn": {
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
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringInSlice(wafv2.FallbackBehavior_Values(), false),
							},
							"header_name": {
								Type:     schema.TypeString,
								Required: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 255),
									validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-]+$`), "must contain only alphanumeric and hyphen characters"),
								),
							},
							"position": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringInSlice(wafv2.ForwardedIPPosition_Values(), false),
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
				"key": {
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 1024),
						validation.StringMatch(regexp.MustCompile(`^[0-9A-Za-z_\-:]+$`), "must contain only alphanumeric, underscore, hyphen, and colon characters"),
					),
				},
				"scope": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice(wafv2.LabelMatchScope_Values(), false),
				},
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
				"arn": {
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
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.StringInSlice([]string{
						wafv2.ComparisonOperatorEq,
						wafv2.ComparisonOperatorGe,
						wafv2.ComparisonOperatorGt,
						wafv2.ComparisonOperatorLe,
						wafv2.ComparisonOperatorLt,
						wafv2.ComparisonOperatorNe,
					}, false),
				},
				"field_to_match": fieldToMatchSchema(),
				"size": {
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
				"field_to_match":      fieldToMatchSchema(),
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
			"body":                emptySchema(),
			"cookies": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"match_scope": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(wafv2.MapMatchScope_Values(), false),
						},
						"oversize_handling": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(wafv2.OversizeHandling_Values(), false),
						},
						"match_pattern": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"all": emptySchema(),
									"included_cookies": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"excluded_cookies": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
			"json_body":    jsonBodySchema(),
			"method":       emptySchema(),
			"query_string": emptySchema(),
			"single_header": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 40),
								// The value is returned in lower case by the API.
								// Trying to solve it with StateFunc and/or DiffSuppressFunc resulted in hash problem of the rule field or didn't work.
								validation.StringMatch(regexp.MustCompile(`^[a-z0-9-_]+$`), "must contain only lowercase alphanumeric characters, underscores, and hyphens"),
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
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 30),
								// The value is returned in lower case by the API.
								// Trying to solve it with StateFunc and/or DiffSuppressFunc resulted in hash problem of the rule field or didn't work.
								validation.StringMatch(regexp.MustCompile(`^[a-z0-9-_]+$`), "must contain only lowercase alphanumeric characters, underscores, and hyphens"),
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
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringInSlice(wafv2.BodyParsingFallbackBehavior_Values(), false),
				},
				"match_pattern": jsonMatchPattern(),
				"match_scope": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice(wafv2.JsonMatchScope_Values(), false),
				},
				"oversize_handling": {
					Type:         schema.TypeString,
					Optional:     true,
					Default:      wafv2.OversizeHandlingContinue,
					ValidateFunc: validation.StringInSlice(wafv2.OversizeHandling_Values(), false),
				},
			},
		},
	}
}

func jsonMatchPattern() *schema.Schema {
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
							validation.StringMatch(regexp.MustCompile(`(/)|(/(([^~])|(~[01]))+)`), "must be a valid JSON pointer")),
					},
				},
			},
		},
	}
}

func forwardedIPConfig() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"fallback_behavior": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice(wafv2.FallbackBehavior_Values(), false),
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
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"priority": {
					Type:     schema.TypeInt,
					Required: true,
				},
				"type": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice(wafv2.TextTransformationType_Values(), false),
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
				"metric_name": {
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 128),
						validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-_]+$`), "must contain only alphanumeric hyphen and underscore characters"),
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
							"name": {
								Type:     schema.TypeString,
								Required: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 64),
									validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9._$-]+$`), "must contain only alphanumeric, hyphen, underscore, dot and $ characters"),
								),
							},
							"value": {
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
						validation.StringMatch(regexp.MustCompile(`^[\w\-]+$`), "must contain only alphanumeric, hyphen, and underscore characters"),
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
							"name": {
								Type:     schema.TypeString,
								Required: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 64),
									validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9._$-]+$`), "must contain only alphanumeric, hyphen, underscore, dot and $ characters"),
								),
							},
							"value": {
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
				"key": {
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 128),
						validation.StringMatch(regexp.MustCompile(`^[\w\-]+$`), "must contain only alphanumeric, hyphen, and underscore characters"),
					),
				},
				"content": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 10240),
				},
				"content_type": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice(wafv2.ResponseContentType_Values(), false),
				},
			},
		},
	}
}
