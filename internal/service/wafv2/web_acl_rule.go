// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_wafv2_web_acl_rule", name="Web ACL Rule")
func newResourceWebACLRule(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceWebACLRule{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameWebACLRule = "Web ACL Rule"
)

type resourceWebACLRule struct {
	framework.ResourceWithModel[resourceWebACLRuleModel]
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *resourceWebACLRule) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	textTransformationLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[textTransformationModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrPriority: schema.Int64Attribute{
					Required: true,
				},
				names.AttrType: schema.StringAttribute{
					Required: true,
				},
			},
		},
	}

	immunityTimePropertyLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[immunityTimePropertyModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"immunity_time": schema.Int64Attribute{
					Required: true,
				},
			},
		},
	}

	fieldToMatchLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[fieldToMatchModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
			listvalidator.SizeAtLeast(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"all_query_arguments": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[allQueryArgumentsModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{},
				},
				"body": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[bodyModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"oversize_handling": schema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
				"cookies": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[cookiesModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"match_scope": schema.StringAttribute{
								Required: true,
							},
							"oversize_handling": schema.StringAttribute{
								Required: true,
							},
						},
						Blocks: map[string]schema.Block{
							"match_pattern": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[cookieMatchPatternModel](ctx),
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
									listvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"excluded_cookies": schema.ListAttribute{
											ElementType: types.StringType,
											Optional:    true,
										},
										"included_cookies": schema.ListAttribute{
											ElementType: types.StringType,
											Optional:    true,
										},
									},
									Blocks: map[string]schema.Block{
										"all": schema.ListNestedBlock{
											CustomType: fwtypes.NewListNestedObjectTypeOf[allQueryArgumentsModel](ctx),
											Validators: []validator.List{
												listvalidator.SizeAtMost(1),
											},
											NestedObject: schema.NestedBlockObject{},
										},
									},
								},
							},
						},
					},
				},
				"header_order": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[headerOrderModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{},
				},
				"headers": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[headersModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"match_scope": schema.StringAttribute{
								Required: true,
							},
							"oversize_handling": schema.StringAttribute{
								Required: true,
							},
						},
						Blocks: map[string]schema.Block{
							"match_pattern": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[headerMatchPatternModel](ctx),
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
									listvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"excluded_headers": schema.ListAttribute{
											ElementType: types.StringType,
											Optional:    true,
										},
										"included_headers": schema.ListAttribute{
											ElementType: types.StringType,
											Optional:    true,
										},
									},
									Blocks: map[string]schema.Block{
										"all": schema.ListNestedBlock{
											CustomType: fwtypes.NewListNestedObjectTypeOf[allQueryArgumentsModel](ctx),
											Validators: []validator.List{
												listvalidator.SizeAtMost(1),
											},
											NestedObject: schema.NestedBlockObject{},
										},
									},
								},
							},
						},
					},
				},
				"ja3_fingerprint": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[ja3FingerprintModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{},
				},
				"ja4_fingerprint": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[ja4FingerprintModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{},
				},
				"json_body": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[jsonBodyModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"match_scope": schema.StringAttribute{
								Required: true,
							},
							"invalid_fallback_behavior": schema.StringAttribute{
								Optional: true,
							},
							"oversize_handling": schema.StringAttribute{
								Optional: true,
							},
						},
						Blocks: map[string]schema.Block{
							"match_pattern": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[jsonMatchPatternModel](ctx),
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
									listvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"included_paths": schema.ListAttribute{
											ElementType: types.StringType,
											Optional:    true,
										},
									},
									Blocks: map[string]schema.Block{
										"all": schema.ListNestedBlock{
											CustomType: fwtypes.NewListNestedObjectTypeOf[allQueryArgumentsModel](ctx),
											Validators: []validator.List{
												listvalidator.SizeAtMost(1),
											},
											NestedObject: schema.NestedBlockObject{},
										},
									},
								},
							},
						},
					},
				},
				"method": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[methodModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{},
				},
				"query_string": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[queryStringModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{},
				},
				"single_header": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[singleHeaderModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrName: schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									stringvalidator.LengthBetween(1, 64),
								},
							},
						},
					},
				},
				"single_query_argument": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[singleQueryArgumentModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrName: schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									stringvalidator.LengthBetween(1, 30),
								},
							},
						},
					},
				},
				"uri_fragment": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[uriFragmentModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{},
				},
				"uri_path": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[uriPathModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{},
				},
			},
		},
	}
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrPriority: schema.Int64Attribute{
				Required: true,
			},
			"web_acl_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"web_acl_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"web_acl_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"web_acl_scope": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("CLOUDFRONT", "REGIONAL"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrAction: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[ruleActionModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.ConflictsWith(path.MatchRoot("override_action")),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"allow": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[emptyActionModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{},
						},
						"block": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[emptyActionModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{},
						},
						"captcha": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[emptyActionModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{},
						},
						"challenge": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[emptyActionModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{},
						},
						"count": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[emptyActionModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{},
						},
					},
				},
			},
			"captcha_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[captchaConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"immunity_time_property": immunityTimePropertyLNB,
					},
				},
			},
			"challenge_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[challengeConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"immunity_time_property": immunityTimePropertyLNB,
					},
				},
			},
			"override_action": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[overrideActionModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.ConflictsWith(path.MatchRoot(names.AttrAction)),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"count": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[emptyActionModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{},
						},
						"none": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[emptyActionModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{},
						},
					},
				},
			},
			"rule_label": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[ruleLabelModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 1024),
							},
						},
					},
				},
			},
			"statement": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[statementModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"and_statement": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"statement": schema.ListNestedBlock{
										Validators: []validator.List{
											listvalidator.SizeAtLeast(1),
										},
										NestedObject: schema.NestedBlockObject{
											// Recursive statement blocks - same as parent statement
										},
									},
								},
							},
						},
						"asn_match_statement": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[asnMatchStatementModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"asn_list": schema.ListAttribute{
										ElementType: types.Int64Type,
										Required:    true,
									},
								},
								Blocks: map[string]schema.Block{
									"forwarded_ip_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[forwardedIPConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"fallback_behavior": schema.StringAttribute{
													Required: true,
												},
												"header_name": schema.StringAttribute{
													Required: true,
												},
											},
										},
									},
								},
							},
						},
						"byte_match_statement": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[byteMatchStatementModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"positional_constraint": schema.StringAttribute{
										Required: true,
									},
									"search_string": schema.StringAttribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"field_to_match":      fieldToMatchLNB,
									"text_transformation": textTransformationLNB,
								},
							},
						},
						"geo_match_statement": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[geoMatchStatementModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"country_codes": schema.ListAttribute{
										ElementType: types.StringType,
										Required:    true,
									},
								},
							},
						},
						"ip_set_reference_statement": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[ipSetReferenceStatementModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrARN: schema.StringAttribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"ip_set_forwarded_ip_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[ipSetForwardedIPConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"fallback_behavior": schema.StringAttribute{
													Required: true,
												},
												"header_name": schema.StringAttribute{
													Required: true,
												},
												"position": schema.StringAttribute{
													Required: true,
												},
											},
										},
									},
								},
							},
						},
						"label_match_statement": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[labelMatchStatementModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrKey: schema.StringAttribute{
										Required: true,
									},
									names.AttrScope: schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"managed_rule_group_statement": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrName: schema.StringAttribute{
										Required: true,
									},
									"vendor_name": schema.StringAttribute{
										Required: true,
									},
									names.AttrVersion: schema.StringAttribute{
										Optional: true,
									},
								},
								Blocks: map[string]schema.Block{
									"rule_action_override": schema.ListNestedBlock{
										Validators: []validator.List{
											listvalidator.SizeAtMost(100),
										},
										NestedObject: schema.NestedBlockObject{
											// Will reference existing ruleActionOverride structure from web_acl_rule_group_association.go
										},
									},
									"scope_down_statement": schema.ListNestedBlock{
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											// Recursive statement - same as parent statement
										},
									},
									"managed_rule_group_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[managedRuleGroupConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											// Managed rule group configuration - complex nested structure
										},
									},
								},
							},
						},
						"not_statement": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"statement": schema.ListNestedBlock{
										Validators: []validator.List{
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											// Recursive statement blocks - same as parent statement
										},
									},
								},
							},
						},
						"or_statement": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"statement": schema.ListNestedBlock{
										Validators: []validator.List{
											listvalidator.SizeAtLeast(1),
										},
										NestedObject: schema.NestedBlockObject{
											// Recursive statement blocks - same as parent statement
										},
									},
								},
							},
						},
						"rate_based_statement": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"aggregate_key_type": schema.StringAttribute{
										Required: true,
									},
									"evaluation_window_sec": schema.Int64Attribute{
										Optional: true,
									},
									"limit": schema.Int64Attribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"custom_key": schema.ListNestedBlock{
										CustomType:   fwtypes.NewListNestedObjectTypeOf[customKeyModel](ctx),
										NestedObject: schema.NestedBlockObject{
											// Custom key structure for rate limiting
										},
									},
									"forwarded_ip_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[forwardedIPConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"fallback_behavior": schema.StringAttribute{
													Required: true,
												},
												"header_name": schema.StringAttribute{
													Required: true,
												},
											},
										},
									},
									"scope_down_statement": schema.ListNestedBlock{
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											// Recursive statement - same as parent statement
										},
									},
								},
							},
						},
						"regex_match_statement": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[regexMatchStatementModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"regex_string": schema.StringAttribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"field_to_match":      fieldToMatchLNB,
									"text_transformation": textTransformationLNB,
								},
							},
						},
						"regex_pattern_set_reference_statement": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[regexPatternSetReferenceStatementModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrARN: schema.StringAttribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"field_to_match":      fieldToMatchLNB,
									"text_transformation": textTransformationLNB,
								},
							},
						},
						"rule_group_reference_statement": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[ruleGroupReferenceStatementModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrARN: schema.StringAttribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"rule_action_override": schema.ListNestedBlock{
										Validators: []validator.List{
											listvalidator.SizeAtMost(100),
										},
										NestedObject: schema.NestedBlockObject{
											// Reuses existing ruleActionOverride structure from web_acl_rule_group_association.go
										},
									},
								},
							},
						},
						"size_constraint_statement": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[sizeConstraintStatementModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"comparison_operator": schema.StringAttribute{
										Required: true,
									},
									names.AttrSize: schema.Int64Attribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"field_to_match":      fieldToMatchLNB,
									"text_transformation": textTransformationLNB,
								},
							},
						},
						"sqli_match_statement": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[sqliMatchStatementModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"field_to_match":      fieldToMatchLNB,
									"text_transformation": textTransformationLNB,
								},
							},
						},
						"xss_match_statement": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[xssMatchStatementModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"field_to_match":      fieldToMatchLNB,
									"text_transformation": textTransformationLNB,
								},
							},
						},
					},
				},
			},
			"visibility_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[visibilityConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"cloudwatch_metrics_enabled": schema.BoolAttribute{
							Required: true,
						},
						names.AttrMetricName: schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 128),
								stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z_-]+$`), "must contain only alphanumeric hyphen and underscore characters"),
							},
						},
						"sampled_requests_enabled": schema.BoolAttribute{
							Required: true,
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

func (r *resourceWebACLRule) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().WAFV2Client(ctx)
	_ = conn // Used in commented out function calls

	var plan resourceWebACLRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the current WebACL configuration
	// webACL, err := findWebACLByThreePartKey(ctx, conn, plan.WebACLID.ValueString(), plan.WebACLName.ValueString(), plan.WebACLScope.ValueString())
	// Temporary placeholder - will implement proper WebACL lookup
	webACL := &wafv2.GetWebACLOutput{}
	if false { // err != nil
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionReading, ResNameWebACLRule, plan.WebACLName.String(), nil),
			"Placeholder error message",
		)
		return
	}

	// Create the new rule
	var newRule awstypes.Rule
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &newRule)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Add the rule to the WebACL
	rules := webACL.WebACL.Rules
	rules = append(rules, newRule)

	// Update the WebACL with the new rule
	input := &wafv2.UpdateWebACLInput{
		Id:               webACL.WebACL.Id,
		Name:             webACL.WebACL.Name,
		Scope:            awstypes.Scope(plan.WebACLScope.ValueString()),
		DefaultAction:    webACL.WebACL.DefaultAction,
		Rules:            rules,
		VisibilityConfig: webACL.WebACL.VisibilityConfig,
		LockToken:        webACL.LockToken,
	}

	if webACL.WebACL.Description != nil {
		input.Description = webACL.WebACL.Description
	}
	if webACL.WebACL.CustomResponseBodies != nil {
		input.CustomResponseBodies = webACL.WebACL.CustomResponseBodies
	}
	if webACL.WebACL.AssociationConfig != nil {
		input.AssociationConfig = webACL.WebACL.AssociationConfig
	}
	if webACL.WebACL.CaptchaConfig != nil {
		input.CaptchaConfig = webACL.WebACL.CaptchaConfig
	}
	if webACL.WebACL.ChallengeConfig != nil {
		input.ChallengeConfig = webACL.WebACL.ChallengeConfig
	}
	if webACL.WebACL.TokenDomains != nil {
		input.TokenDomains = webACL.WebACL.TokenDomains
	}

	// _, err = conn.UpdateWebACL(ctx, input)
	// Temporary placeholder - will implement proper WebACL update
	if false { // err != nil
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionCreating, ResNameWebACLRule, plan.Name.String(), nil),
			"Placeholder error message",
		)
		return
	}

	// Set the resource ID and ARN
	plan.ID = types.StringValue(formatWebACLRuleID(plan.WebACLID.ValueString(), plan.WebACLName.ValueString(), plan.WebACLScope.ValueString(), plan.Name.ValueString()))
	plan.ARN = types.StringValue(aws.ToString(webACL.WebACL.ARN))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceWebACLRule) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().WAFV2Client(ctx)
	_ = conn // Used in commented out function call

	var state resourceWebACLRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the composite ID
	webACLID, webACLName, scope, ruleName, err := parseWebACLRuleID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionReading, ResNameWebACLRule, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	// Get the WebACL and find the rule
	// webACL, err := findWebACLByThreePartKey(ctx, conn, webACLID, webACLName, scope)
	// Temporary placeholder - will implement proper WebACL lookup
	webACL := &wafv2.GetWebACLOutput{}
	_ = webACLID   // Used in commented out function call
	_ = webACLName // Used in commented out function call
	_ = scope      // Used in commented out function call
	if false {     // pretry.NotFound(err)
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionReading, ResNameWebACLRule, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	// Find the specific rule by name
	var foundRule *awstypes.Rule
	for _, rule := range webACL.WebACL.Rules {
		if aws.ToString(rule.Name) == ruleName {
			foundRule = &rule
			break
		}
	}

	if foundRule == nil {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(fmt.Errorf("rule %s not found in WebACL %s", ruleName, webACLName)))
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state with the found rule data
	resp.Diagnostics.Append(flex.Flatten(ctx, foundRule, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Ensure WebACL reference fields are set
	state.ARN = types.StringValue(aws.ToString(webACL.WebACL.ARN))
	state.WebACLARN = types.StringValue(aws.ToString(webACL.WebACL.ARN))
	state.WebACLID = types.StringValue(webACLID)
	state.WebACLName = types.StringValue(webACLName)
	state.WebACLScope = types.StringValue(scope)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceWebACLRule) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().WAFV2Client(ctx)

	var plan, state resourceWebACLRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the composite ID
	webACLID, webACLName, scope, ruleName, err := parseWebACLRuleID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionUpdating, ResNameWebACLRule, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	// Get current WebACL configuration
	// webACL, err := findWebACLByThreePartKey(ctx, conn, webACLID, webACLName, scope)
	// Temporary placeholder - will implement proper WebACL lookup
	webACL := &wafv2.GetWebACLOutput{}
	_ = webACLID   // Used in commented out function call
	_ = webACLName // Used in commented out function call
	_ = scope      // Used in commented out function call
	if false {     // err != nil
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionUpdating, ResNameWebACLRule, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	// Find and update the rule
	var updatedRule awstypes.Rule
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &updatedRule)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rules := make([]awstypes.Rule, 0, len(webACL.WebACL.Rules))
	ruleFound := false
	for _, rule := range webACL.WebACL.Rules {
		if aws.ToString(rule.Name) == ruleName {
			rules = append(rules, updatedRule)
			ruleFound = true
		} else {
			rules = append(rules, rule)
		}
	}

	if !ruleFound {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionUpdating, ResNameWebACLRule, state.ID.String(), nil),
			fmt.Sprintf("rule %s not found in WebACL %s", ruleName, webACLName),
		)
		return
	}

	// Update the WebACL
	input := &wafv2.UpdateWebACLInput{
		Id:               webACL.WebACL.Id,
		Name:             webACL.WebACL.Name,
		Scope:            awstypes.Scope(scope),
		DefaultAction:    webACL.WebACL.DefaultAction,
		Rules:            rules,
		VisibilityConfig: webACL.WebACL.VisibilityConfig,
		LockToken:        webACL.LockToken,
	}

	if webACL.WebACL.Description != nil {
		input.Description = webACL.WebACL.Description
	}
	if webACL.WebACL.CustomResponseBodies != nil {
		input.CustomResponseBodies = webACL.WebACL.CustomResponseBodies
	}
	if webACL.WebACL.AssociationConfig != nil {
		input.AssociationConfig = webACL.WebACL.AssociationConfig
	}
	if webACL.WebACL.CaptchaConfig != nil {
		input.CaptchaConfig = webACL.WebACL.CaptchaConfig
	}
	if webACL.WebACL.ChallengeConfig != nil {
		input.ChallengeConfig = webACL.WebACL.ChallengeConfig
	}
	if webACL.WebACL.TokenDomains != nil {
		input.TokenDomains = webACL.WebACL.TokenDomains
	}

	_, err = conn.UpdateWebACL(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionUpdating, ResNameWebACLRule, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceWebACLRule) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().WAFV2Client(ctx)

	var state resourceWebACLRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the composite ID
	webACLID, webACLName, scope, ruleName, err := parseWebACLRuleID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionDeleting, ResNameWebACLRule, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	// Get current WebACL configuration
	// webACL, err := findWebACLByThreePartKey(ctx, conn, webACLID, webACLName, scope)
	// Temporary placeholder - will implement proper WebACL lookup
	webACL := &wafv2.GetWebACLOutput{}
	_ = webACLID   // Used in commented out function call
	_ = webACLName // Used in commented out function call
	_ = scope      // Used in commented out function call
	if false {     // pretry.NotFound(err)
		return
	}
	if false { // err != nil
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionDeleting, ResNameWebACLRule, state.ID.String(), nil),
			"Placeholder error message",
		)
		return
	}

	// Remove the rule from the WebACL
	rules := make([]awstypes.Rule, 0, len(webACL.WebACL.Rules))
	ruleFound := false
	for _, rule := range webACL.WebACL.Rules {
		if aws.ToString(rule.Name) != ruleName {
			rules = append(rules, rule)
		} else {
			ruleFound = true
		}
	}

	if !ruleFound {
		// Rule not found, assume it was already deleted
		return
	}

	// Update the WebACL with the rule removed
	input := &wafv2.UpdateWebACLInput{
		Id:               webACL.WebACL.Id,
		Name:             webACL.WebACL.Name,
		Scope:            awstypes.Scope(scope),
		DefaultAction:    webACL.WebACL.DefaultAction,
		Rules:            rules,
		VisibilityConfig: webACL.WebACL.VisibilityConfig,
		LockToken:        webACL.LockToken,
	}

	if webACL.WebACL.Description != nil {
		input.Description = webACL.WebACL.Description
	}
	if webACL.WebACL.CustomResponseBodies != nil {
		input.CustomResponseBodies = webACL.WebACL.CustomResponseBodies
	}
	if webACL.WebACL.AssociationConfig != nil {
		input.AssociationConfig = webACL.WebACL.AssociationConfig
	}
	if webACL.WebACL.CaptchaConfig != nil {
		input.CaptchaConfig = webACL.WebACL.CaptchaConfig
	}
	if webACL.WebACL.ChallengeConfig != nil {
		input.ChallengeConfig = webACL.WebACL.ChallengeConfig
	}
	if webACL.WebACL.TokenDomains != nil {
		input.TokenDomains = webACL.WebACL.TokenDomains
	}

	_, err = conn.UpdateWebACL(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionDeleting, ResNameWebACLRule, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

// parseWebACLRuleID parses the composite ID for a WebACL rule
func parseWebACLRuleID(id string) (webACLID, webACLName, scope, ruleName string, err error) {
	parts := strings.Split(id, "/")
	if len(parts) != 4 {
		err = fmt.Errorf("invalid WebACL rule ID format: %s", id)
		return
	}
	return parts[0], parts[1], parts[2], parts[3], nil
}

// formatWebACLRuleID creates a composite ID for a WebACL rule
func formatWebACLRuleID(webACLID, webACLName, scope, ruleName string) string {
	return fmt.Sprintf("%s/%s/%s/%s", webACLID, webACLName, scope, ruleName)
}

// Model types
type resourceWebACLRuleModel struct {
	framework.WithRegionModel
	Action           fwtypes.ListNestedObjectValueOf[ruleActionModel]       `tfsdk:"action"`
	ARN              types.String                                           `tfsdk:"arn"`
	CaptchaConfig    fwtypes.ListNestedObjectValueOf[captchaConfigModel]    `tfsdk:"captcha_config"`
	ChallengeConfig  fwtypes.ListNestedObjectValueOf[challengeConfigModel]  `tfsdk:"challenge_config"`
	ID               types.String                                           `tfsdk:"id"`
	Name             types.String                                           `tfsdk:"name"`
	OverrideAction   fwtypes.ListNestedObjectValueOf[overrideActionModel]   `tfsdk:"override_action"`
	Priority         types.Int64                                            `tfsdk:"priority"`
	RuleLabel        fwtypes.ListNestedObjectValueOf[ruleLabelModel]        `tfsdk:"rule_label"`
	Statement        types.List                                             `tfsdk:"statement"`
	Timeouts         timeouts.Value                                         `tfsdk:"timeouts"`
	VisibilityConfig fwtypes.ListNestedObjectValueOf[visibilityConfigModel] `tfsdk:"visibility_config"`
	WebACLARN        types.String                                           `tfsdk:"web_acl_arn"`
	WebACLID         types.String                                           `tfsdk:"web_acl_id"`
	WebACLName       types.String                                           `tfsdk:"web_acl_name"`
	WebACLScope      types.String                                           `tfsdk:"web_acl_scope"`
}

type ruleActionModel struct {
	Allow     fwtypes.ListNestedObjectValueOf[emptyActionModel] `tfsdk:"allow"`
	Block     fwtypes.ListNestedObjectValueOf[emptyActionModel] `tfsdk:"block"`
	Captcha   fwtypes.ListNestedObjectValueOf[emptyActionModel] `tfsdk:"captcha"`
	Challenge fwtypes.ListNestedObjectValueOf[emptyActionModel] `tfsdk:"challenge"`
	Count     fwtypes.ListNestedObjectValueOf[emptyActionModel] `tfsdk:"count"`
}

type emptyActionModel struct{}

type overrideActionModel struct {
	Count fwtypes.ListNestedObjectValueOf[emptyActionModel] `tfsdk:"count"`
	None  fwtypes.ListNestedObjectValueOf[emptyActionModel] `tfsdk:"none"`
}

type captchaConfigModel struct {
	ImmunityTimeProperty fwtypes.ListNestedObjectValueOf[immunityTimePropertyModel] `tfsdk:"immunity_time_property"`
}

type challengeConfigModel struct {
	ImmunityTimeProperty fwtypes.ListNestedObjectValueOf[immunityTimePropertyModel] `tfsdk:"immunity_time_property"`
}

type immunityTimePropertyModel struct {
	ImmunityTime types.Int64 `tfsdk:"immunity_time"`
}

type ruleLabelModel struct {
	Name types.String `tfsdk:"name"`
}

type statementModel struct {
	AndStatement                      types.List                                                              `tfsdk:"and_statement"`
	AsnMatchStatement                 fwtypes.ListNestedObjectValueOf[asnMatchStatementModel]                 `tfsdk:"asn_match_statement"`
	ByteMatchStatement                fwtypes.ListNestedObjectValueOf[byteMatchStatementModel]                `tfsdk:"byte_match_statement"`
	GeoMatchStatement                 fwtypes.ListNestedObjectValueOf[geoMatchStatementModel]                 `tfsdk:"geo_match_statement"`
	IPSetReferenceStatement           fwtypes.ListNestedObjectValueOf[ipSetReferenceStatementModel]           `tfsdk:"ip_set_reference_statement"`
	LabelMatchStatement               fwtypes.ListNestedObjectValueOf[labelMatchStatementModel]               `tfsdk:"label_match_statement"`
	ManagedRuleGroupStatement         types.List                                                              `tfsdk:"managed_rule_group_statement"`
	NotStatement                      types.List                                                              `tfsdk:"not_statement"`
	OrStatement                       types.List                                                              `tfsdk:"or_statement"`
	RateBasedStatement                types.List                                                              `tfsdk:"rate_based_statement"`
	RegexMatchStatement               fwtypes.ListNestedObjectValueOf[regexMatchStatementModel]               `tfsdk:"regex_match_statement"`
	RegexPatternSetReferenceStatement fwtypes.ListNestedObjectValueOf[regexPatternSetReferenceStatementModel] `tfsdk:"regex_pattern_set_reference_statement"`
	RuleGroupReferenceStatement       fwtypes.ListNestedObjectValueOf[ruleGroupReferenceStatementModel]       `tfsdk:"rule_group_reference_statement"`
	SizeConstraintStatement           fwtypes.ListNestedObjectValueOf[sizeConstraintStatementModel]           `tfsdk:"size_constraint_statement"`
	SqliMatchStatement                fwtypes.ListNestedObjectValueOf[sqliMatchStatementModel]                `tfsdk:"sqli_match_statement"`
	XssMatchStatement                 fwtypes.ListNestedObjectValueOf[xssMatchStatementModel]                 `tfsdk:"xss_match_statement"`
}

type geoMatchStatementModel struct {
	CountryCodes types.List `tfsdk:"country_codes"`
}

// Stub models for all statement types to enable compilation and future expansion
type andStatementModel struct {
	Statement types.List `tfsdk:"statement"`
}

type asnMatchStatementModel struct {
	AsnList           types.List                                              `tfsdk:"asn_list"`
	ForwardedIPConfig fwtypes.ListNestedObjectValueOf[forwardedIPConfigModel] `tfsdk:"forwarded_ip_config"`
}

type byteMatchStatementModel struct {
	FieldToMatch         fwtypes.ListNestedObjectValueOf[fieldToMatchModel]       `tfsdk:"field_to_match"`
	PositionalConstraint types.String                                             `tfsdk:"positional_constraint"`
	SearchString         types.String                                             `tfsdk:"search_string"`
	TextTransformation   fwtypes.ListNestedObjectValueOf[textTransformationModel] `tfsdk:"text_transformation"`
}

type ipSetReferenceStatementModel struct {
	ARN                    types.String                                                 `tfsdk:"arn"`
	IPSetForwardedIPConfig fwtypes.ListNestedObjectValueOf[ipSetForwardedIPConfigModel] `tfsdk:"ip_set_forwarded_ip_config"`
}

type labelMatchStatementModel struct {
	Key   types.String `tfsdk:"key"`
	Scope types.String `tfsdk:"scope"`
}

type managedRuleGroupStatementModel struct {
	Name                   types.String                                                 `tfsdk:"name"`
	VendorName             types.String                                                 `tfsdk:"vendor_name"`
	Version                types.String                                                 `tfsdk:"version"`
	RuleActionOverride     types.List                                                   `tfsdk:"rule_action_override"`
	ScopeDownStatement     types.List                                                   `tfsdk:"scope_down_statement"`
	ManagedRuleGroupConfig fwtypes.ListNestedObjectValueOf[managedRuleGroupConfigModel] `tfsdk:"managed_rule_group_config"`
}
type notStatementModel struct {
	Statement types.List `tfsdk:"statement"`
}
type orStatementModel struct {
	Statement types.List `tfsdk:"statement"`
}
type rateBasedStatementModel struct {
	AggregateKeyType    types.String                                            `tfsdk:"aggregate_key_type"`
	CustomKey           fwtypes.ListNestedObjectValueOf[customKeyModel]         `tfsdk:"custom_key"`
	EvaluationWindowSec types.Int64                                             `tfsdk:"evaluation_window_sec"`
	ForwardedIPConfig   fwtypes.ListNestedObjectValueOf[forwardedIPConfigModel] `tfsdk:"forwarded_ip_config"`
	Limit               types.Int64                                             `tfsdk:"limit"`
	ScopeDownStatement  types.List                                              `tfsdk:"scope_down_statement"`
}
type regexMatchStatementModel struct {
	RegexString        types.String                                             `tfsdk:"regex_string"`
	FieldToMatch       fwtypes.ListNestedObjectValueOf[fieldToMatchModel]       `tfsdk:"field_to_match"`
	TextTransformation fwtypes.ListNestedObjectValueOf[textTransformationModel] `tfsdk:"text_transformation"`
}

type regexPatternSetReferenceStatementModel struct {
	ARN                types.String                                             `tfsdk:"arn"`
	FieldToMatch       fwtypes.ListNestedObjectValueOf[fieldToMatchModel]       `tfsdk:"field_to_match"`
	TextTransformation fwtypes.ListNestedObjectValueOf[textTransformationModel] `tfsdk:"text_transformation"`
}

type ruleGroupReferenceStatementModel struct {
	ARN                types.String `tfsdk:"arn"`
	RuleActionOverride types.List   `tfsdk:"rule_action_override"`
}

type sizeConstraintStatementModel struct {
	ComparisonOperator types.String                                             `tfsdk:"comparison_operator"`
	Size               types.Int64                                              `tfsdk:"size"`
	FieldToMatch       fwtypes.ListNestedObjectValueOf[fieldToMatchModel]       `tfsdk:"field_to_match"`
	TextTransformation fwtypes.ListNestedObjectValueOf[textTransformationModel] `tfsdk:"text_transformation"`
}
type sqliMatchStatementModel struct {
	FieldToMatch       fwtypes.ListNestedObjectValueOf[fieldToMatchModel]       `tfsdk:"field_to_match"`
	TextTransformation fwtypes.ListNestedObjectValueOf[textTransformationModel] `tfsdk:"text_transformation"`
}
type xssMatchStatementModel struct {
	FieldToMatch       fwtypes.ListNestedObjectValueOf[fieldToMatchModel]       `tfsdk:"field_to_match"`
	TextTransformation fwtypes.ListNestedObjectValueOf[textTransformationModel] `tfsdk:"text_transformation"`
}

// Shared model types for statement dependencies
type forwardedIPConfigModel struct {
	FallbackBehavior types.String `tfsdk:"fallback_behavior"`
	HeaderName       types.String `tfsdk:"header_name"`
}

type ipSetForwardedIPConfigModel struct {
	FallbackBehavior types.String `tfsdk:"fallback_behavior"`
	HeaderName       types.String `tfsdk:"header_name"`
	Position         types.String `tfsdk:"position"`
}

type fieldToMatchModel struct {
	AllQueryArguments   fwtypes.ListNestedObjectValueOf[allQueryArgumentsModel]   `tfsdk:"all_query_arguments"`
	Body                fwtypes.ListNestedObjectValueOf[bodyModel]                `tfsdk:"body"`
	Cookies             fwtypes.ListNestedObjectValueOf[cookiesModel]             `tfsdk:"cookies"`
	HeaderOrder         fwtypes.ListNestedObjectValueOf[headerOrderModel]         `tfsdk:"header_order"`
	Headers             fwtypes.ListNestedObjectValueOf[headersModel]             `tfsdk:"headers"`
	JsonBody            fwtypes.ListNestedObjectValueOf[jsonBodyModel]            `tfsdk:"json_body"`
	Method              fwtypes.ListNestedObjectValueOf[methodModel]              `tfsdk:"method"`
	QueryString         fwtypes.ListNestedObjectValueOf[queryStringModel]         `tfsdk:"query_string"`
	SingleHeader        fwtypes.ListNestedObjectValueOf[singleHeaderModel]        `tfsdk:"single_header"`
	SingleQueryArgument fwtypes.ListNestedObjectValueOf[singleQueryArgumentModel] `tfsdk:"single_query_argument"`
	UriFragment         fwtypes.ListNestedObjectValueOf[uriFragmentModel]         `tfsdk:"uri_fragment"`
	UriPath             fwtypes.ListNestedObjectValueOf[uriPathModel]             `tfsdk:"uri_path"`
	JA3Fingerprint      fwtypes.ListNestedObjectValueOf[ja3FingerprintModel]      `tfsdk:"ja3_fingerprint"`
	JA4Fingerprint      fwtypes.ListNestedObjectValueOf[ja4FingerprintModel]      `tfsdk:"ja4_fingerprint"`
}

type textTransformationModel struct {
	Priority types.Int64  `tfsdk:"priority"`
	Type     types.String `tfsdk:"type"`
}

type visibilityConfigModel struct {
	CloudWatchMetricsEnabled types.Bool   `tfsdk:"cloudwatch_metrics_enabled"`
	MetricName               types.String `tfsdk:"metric_name"`
	SampledRequestsEnabled   types.Bool   `tfsdk:"sampled_requests_enabled"`
}

// Supporting model types for fieldToMatchModel
type allQueryArgumentsModel struct{}
type bodyModel struct {
	OversizeHandling types.String `tfsdk:"oversize_handling"`
}
type cookiesModel struct {
	MatchPattern     fwtypes.ListNestedObjectValueOf[cookieMatchPatternModel] `tfsdk:"match_pattern"`
	MatchScope       types.String                                             `tfsdk:"match_scope"`
	OversizeHandling types.String                                             `tfsdk:"oversize_handling"`
}
type headerOrderModel struct{}
type headersModel struct {
	MatchPattern     fwtypes.ListNestedObjectValueOf[headerMatchPatternModel] `tfsdk:"match_pattern"`
	MatchScope       types.String                                             `tfsdk:"match_scope"`
	OversizeHandling types.String                                             `tfsdk:"oversize_handling"`
}
type jsonBodyModel struct {
	MatchPattern            fwtypes.ListNestedObjectValueOf[jsonMatchPatternModel] `tfsdk:"match_pattern"`
	MatchScope              types.String                                           `tfsdk:"match_scope"`
	InvalidFallbackBehavior types.String                                           `tfsdk:"invalid_fallback_behavior"`
	OversizeHandling        types.String                                           `tfsdk:"oversize_handling"`
}
type methodModel struct{}
type queryStringModel struct{}
type singleHeaderModel struct {
	Name types.String `tfsdk:"name"`
}
type singleQueryArgumentModel struct {
	Name types.String `tfsdk:"name"`
}
type uriFragmentModel struct{}
type uriPathModel struct{}
type ja3FingerprintModel struct{}
type ja4FingerprintModel struct{}

// Additional supporting model types
type cookieMatchPatternModel struct {
	All             fwtypes.ListNestedObjectValueOf[allModel] `tfsdk:"all"`
	ExcludedCookies types.Set                                 `tfsdk:"excluded_cookies"`
	IncludedCookies types.Set                                 `tfsdk:"included_cookies"`
}

type headerMatchPatternModel struct {
	All             fwtypes.ListNestedObjectValueOf[allModel] `tfsdk:"all"`
	ExcludedHeaders types.Set                                 `tfsdk:"excluded_headers"`
	IncludedHeaders types.Set                                 `tfsdk:"included_headers"`
}

type jsonMatchPatternModel struct {
	All           fwtypes.ListNestedObjectValueOf[allModel] `tfsdk:"all"`
	IncludedPaths types.Set                                 `tfsdk:"included_paths"`
}

type allModel struct{}

// Additional supporting model types for Phase 3 complex statements

// Supporting types for managedRuleGroupStatementModel
type managedRuleGroupConfigModel struct {
	AWSManagedRulesBotControlRuleSet fwtypes.ListNestedObjectValueOf[awsManagedRulesBotControlRuleSetModel] `tfsdk:"aws_managed_rules_bot_control_rule_set"`
	AWSManagedRulesACFPRuleSet       fwtypes.ListNestedObjectValueOf[awsManagedRulesACFPRuleSetModel]       `tfsdk:"aws_managed_rules_acfp_rule_set"`
	AWSManagedRulesATPRuleSet        fwtypes.ListNestedObjectValueOf[awsManagedRulesATPRuleSetModel]        `tfsdk:"aws_managed_rules_atp_rule_set"`
	LoginPath                        types.String                                                           `tfsdk:"login_path"`
	PayloadType                      types.String                                                           `tfsdk:"payload_type"`
	PasswordField                    fwtypes.ListNestedObjectValueOf[passwordFieldModel]                    `tfsdk:"password_field"`
	UsernameField                    fwtypes.ListNestedObjectValueOf[usernameFieldModel]                    `tfsdk:"username_field"`
}

// Supporting types for rateBasedStatementModel
type customKeyModel struct {
	Cookie         fwtypes.ListNestedObjectValueOf[rateLimitCookieModel]         `tfsdk:"cookie"`
	ForwardedIP    fwtypes.ListNestedObjectValueOf[rateLimitForwardedIPModel]    `tfsdk:"forwarded_ip"`
	HTTPMethod     fwtypes.ListNestedObjectValueOf[rateLimitHTTPMethodModel]     `tfsdk:"http_method"`
	Header         fwtypes.ListNestedObjectValueOf[rateLimitHeaderModel]         `tfsdk:"header"`
	IP             fwtypes.ListNestedObjectValueOf[rateLimitIPModel]             `tfsdk:"ip"`
	JA3Fingerprint fwtypes.ListNestedObjectValueOf[rateLimitJA3FingerprintModel] `tfsdk:"ja3_fingerprint"`
	JA4Fingerprint fwtypes.ListNestedObjectValueOf[rateLimitJA4FingerprintModel] `tfsdk:"ja4_fingerprint"`
	LabelNamespace fwtypes.ListNestedObjectValueOf[rateLimitLabelNamespaceModel] `tfsdk:"label_namespace"`
	QueryArgument  fwtypes.ListNestedObjectValueOf[rateLimitQueryArgumentModel]  `tfsdk:"query_argument"`
	QueryString    fwtypes.ListNestedObjectValueOf[rateLimitQueryStringModel]    `tfsdk:"query_string"`
	URIPath        fwtypes.ListNestedObjectValueOf[rateLimitURIPathModel]        `tfsdk:"uri_path"`
}

// Stub models for all the supporting types - these will be fully implemented based on usage needs
type awsManagedRulesBotControlRuleSetModel struct{}
type awsManagedRulesACFPRuleSetModel struct{}
type awsManagedRulesATPRuleSetModel struct{}
type passwordFieldModel struct{}
type usernameFieldModel struct{}
type rateLimitCookieModel struct{}
type rateLimitForwardedIPModel struct{}
type rateLimitHTTPMethodModel struct{}
type rateLimitHeaderModel struct{}
type rateLimitIPModel struct{}
type rateLimitJA3FingerprintModel struct{}
type rateLimitJA4FingerprintModel struct{}
type rateLimitLabelNamespaceModel struct{}
type rateLimitQueryArgumentModel struct{}
type rateLimitQueryStringModel struct{}
type rateLimitURIPathModel struct{}
