// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package mailmanager

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mailmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mailmanager/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_mailmanager_rule_set", name="Rule Set")
// @IdentityAttribute("id")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/mailmanager;mailmanager.GetRuleSetOutput")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccRuleSetPreCheck")
// @Testing(skipEmptyTags=true, skipNullTags=true)
func newRuleSetResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &ruleSetResource{}, nil
}

const ResNameRuleSet = "Rule Set"

type ruleSetResource struct {
	framework.ResourceWithModel[ruleSetResourceModel]
	framework.WithImportByIdentity
}

func (r *ruleSetResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreatedDate: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"last_modification_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrRule: ruleSetRuleBlock(ctx),
		},
	}
}

func ruleSetRuleBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[ruleSetRuleModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeBetween(1, 40),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrName: schema.StringAttribute{
					Optional: true,
				},
			},
			Blocks: map[string]schema.Block{
				names.AttrAction:    ruleSetActionBlock(ctx),
				names.AttrCondition: ruleSetConditionBlock(ctx),
				"unless":            ruleSetConditionBlock(ctx),
			},
		},
	}
}

func ruleSetConditionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[ruleSetConditionModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeBetween(1, 10),
		},
		NestedObject: schema.NestedBlockObject{Blocks: map[string]schema.Block{
			"boolean_expression": ruleBooleanExpressionBlock(ctx),
			"dmarc_expression":   ruleDMARCExpressionBlock(ctx),
			"ip_expression":      ruleIPExpressionBlock(ctx),
			"number_expression":  ruleNumberExpressionBlock(ctx),
			"string_expression":  ruleStringExpressionBlock(ctx),
			"verdict_expression": ruleVerdictExpressionBlock(ctx),
		}},
	}
}

func ruleBooleanExpressionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[ruleBooleanExpressionModel](ctx),
		Validators: conditionUnionValidators("dmarc_expression", "ip_expression", "number_expression", "string_expression", "verdict_expression"),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"operator": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.RuleBooleanOperator](),
					Required:   true,
				},
			},
			Blocks: map[string]schema.Block{
				"evaluate": ruleBooleanEvaluateBlock(ctx),
			},
		},
	}
}

func ruleBooleanEvaluateBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[ruleBooleanEvaluateModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeBetween(1, 1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"attribute": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.RuleBooleanEmailAttribute](),
					Optional:   true,
					Validators: []validator.String{stringvalidator.ExactlyOneOf(
						path.MatchRelative().AtParent().AtName("analysis"),
						path.MatchRelative().AtParent().AtName("is_in_address_list"),
					)},
				},
			},
			Blocks: map[string]schema.Block{
				"analysis": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[analysisModel](ctx),
					Validators: conditionUnionValidators("attribute", "is_in_address_list"),
					NestedObject: schema.NestedBlockObject{
						Attributes: analysisAttributes(),
					},
				},
				"is_in_address_list": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[ruleIsInAddressListModel](ctx),
					Validators: conditionUnionValidators("analysis", "attribute"),
					NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
						"address_lists": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							Required:    true,
							Validators: []validator.List{
								listvalidator.SizeBetween(1, 1),
							},
						},
						"attribute": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.RuleAddressListEmailAttribute](),
							Required:   true,
						},
					}},
				},
			},
		},
	}
}

func ruleDMARCExpressionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[ruleDMARCExpressionModel](ctx),
		Validators: conditionUnionValidators("boolean_expression", "ip_expression", "number_expression", "string_expression", "verdict_expression"),
		NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
			"operator": schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.RuleDmarcOperator](), Required: true},
			names.AttrValues: schema.ListAttribute{
				CustomType: fwtypes.ListOfStringEnumType[awstypes.RuleDmarcPolicy](),
				Required:   true,
				Validators: []validator.List{listvalidator.SizeAtLeast(1)},
			},
		}},
	}
}

func ruleIPExpressionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[ruleIPExpressionModel](ctx),
		Validators: conditionUnionValidators("boolean_expression", "dmarc_expression", "number_expression", "string_expression", "verdict_expression"),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"operator": schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.RuleIpOperator](), Required: true},
				names.AttrValues: schema.ListAttribute{
					CustomType:  fwtypes.ListOfStringType,
					ElementType: types.StringType,
					Required:    true,
					Validators: []validator.List{
						listvalidator.SizeBetween(1, 10),
						listvalidator.ValueStringsAre(fwvalidators.IPv4CIDRNetworkAddress()),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"evaluate": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[ruleIPEvaluateModel](ctx),
					Validators: []validator.List{listvalidator.SizeBetween(1, 1)},
					NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
						"attribute": schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.RuleIpEmailAttribute](), Required: true},
					}},
				},
			},
		},
	}
}

func ruleNumberExpressionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[ruleNumberExpressionModel](ctx),
		Validators: conditionUnionValidators("boolean_expression", "dmarc_expression", "ip_expression", "string_expression", "verdict_expression"),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"operator":      schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.RuleNumberOperator](), Required: true},
				names.AttrValue: schema.Float64Attribute{Required: true},
			},
			Blocks: map[string]schema.Block{
				"evaluate": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[ruleNumberEvaluateModel](ctx),
					Validators: []validator.List{listvalidator.SizeBetween(1, 1)},
					NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
						"attribute": schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.RuleNumberEmailAttribute](), Required: true},
					}},
				},
			},
		},
	}
}

func ruleStringExpressionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[ruleStringExpressionModel](ctx),
		Validators: conditionUnionValidators("boolean_expression", "dmarc_expression", "ip_expression", "number_expression", "verdict_expression"),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"operator": schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.RuleStringOperator](), Required: true},
				names.AttrValues: schema.ListAttribute{
					CustomType:  fwtypes.ListOfStringType,
					ElementType: types.StringType,
					Required:    true,
					Validators: []validator.List{
						listvalidator.SizeBetween(1, 10),
						listvalidator.ValueStringsAre(stringvalidator.LengthAtMost(4096)),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"evaluate": ruleStringEvaluateBlock(ctx),
			},
		},
	}
}

func ruleStringEvaluateBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[ruleStringEvaluateModel](ctx),
		Validators: []validator.List{listvalidator.SizeBetween(1, 1)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"attribute": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.RuleStringEmailAttribute](), Optional: true,
					Validators: []validator.String{stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("analysis"), path.MatchRelative().AtParent().AtName("client_certificate_attribute"), path.MatchRelative().AtParent().AtName("mime_header_attribute"))},
				},
				"client_certificate_attribute": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.RuleClientCertificateAttribute](), Optional: true,
					Validators: []validator.String{stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("analysis"), path.MatchRelative().AtParent().AtName("attribute"), path.MatchRelative().AtParent().AtName("mime_header_attribute"))},
				},
				"mime_header_attribute": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 256),
						stringvalidator.RegexMatches(regexache.MustCompile(`^[Xx]-.+`), "must begin with X- or x-"),
						stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("analysis"), path.MatchRelative().AtParent().AtName("attribute"), path.MatchRelative().AtParent().AtName("client_certificate_attribute")),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"analysis": schema.ListNestedBlock{
					CustomType:   fwtypes.NewListNestedObjectTypeOf[analysisModel](ctx),
					Validators:   conditionUnionValidators("attribute", "client_certificate_attribute", "mime_header_attribute"),
					NestedObject: schema.NestedBlockObject{Attributes: analysisAttributes()},
				},
			},
		},
	}
}

func ruleVerdictExpressionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[ruleVerdictExpressionModel](ctx),
		Validators: conditionUnionValidators("boolean_expression", "dmarc_expression", "ip_expression", "number_expression", "string_expression"),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"operator":       schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.RuleVerdictOperator](), Required: true},
				names.AttrValues: schema.ListAttribute{CustomType: fwtypes.ListOfStringEnumType[awstypes.RuleVerdict](), Required: true, Validators: []validator.List{listvalidator.SizeBetween(1, 10)}},
			},
			Blocks: map[string]schema.Block{
				"evaluate": ruleVerdictEvaluateBlock(ctx),
			},
		},
	}
}

func ruleVerdictEvaluateBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[ruleVerdictEvaluateModel](ctx),
		Validators: []validator.List{listvalidator.SizeBetween(1, 1)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"attribute": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.RuleVerdictAttribute](), Optional: true,
					Validators: []validator.String{stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("analysis"))},
				},
			},
			Blocks: map[string]schema.Block{
				"analysis": schema.ListNestedBlock{
					CustomType:   fwtypes.NewListNestedObjectTypeOf[analysisModel](ctx),
					Validators:   conditionUnionValidators("attribute"),
					NestedObject: schema.NestedBlockObject{Attributes: analysisAttributes()},
				},
			},
		},
	}
}

func ruleSetActionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[ruleSetActionModel](ctx),
		Validators: []validator.List{listvalidator.SizeBetween(1, 10)},
		NestedObject: schema.NestedBlockObject{Blocks: map[string]schema.Block{
			"add_header":            addHeaderActionBlock(ctx),
			"archive":               archiveActionBlock(ctx),
			"bounce":                bounceActionBlock(ctx),
			"deliver_to_mailbox":    deliverToMailboxActionBlock(ctx),
			"deliver_to_q_business": deliverToQBusinessActionBlock(ctx),
			"drop":                  dropActionBlock(ctx),
			"invoke_lambda":         invokeLambdaActionBlock(ctx),
			"publish_to_sns":        publishToSNSActionBlock(ctx),
			"relay":                 relayActionBlock(ctx),
			"replace_recipient":     replaceRecipientActionBlock(ctx),
			"send":                  sendActionBlock(ctx),
			"write_to_s3":           writeToS3ActionBlock(ctx),
		}},
	}
}

func actionUnionValidators(siblings ...string) []validator.List {
	paths := make([]path.Expression, len(siblings))
	for i, sibling := range siblings {
		paths[i] = path.MatchRelative().AtParent().AtName(sibling)
	}
	return []validator.List{listvalidator.SizeAtMost(1), listvalidator.ExactlyOneOf(paths...)}
}

func addHeaderActionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[addHeaderActionModel](ctx),
		Validators: actionUnionValidators("archive", "bounce", "deliver_to_mailbox", "deliver_to_q_business", "drop", "invoke_lambda", "publish_to_sns", "relay", "replace_recipient", "send", "write_to_s3"),
		NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
			"header_name": schema.StringAttribute{Required: true, Validators: []validator.String{
				stringvalidator.LengthBetween(1, 64),
				stringvalidator.RegexMatches(regexache.MustCompile(`^[Xx]-`), "must begin with X- or x-"),
			}},
			"header_value": schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.LengthBetween(1, 128)}},
		}},
	}
}

func archiveActionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[archiveActionModel](ctx),
		Validators: actionUnionValidators("add_header", "bounce", "deliver_to_mailbox", "deliver_to_q_business", "drop", "invoke_lambda", "publish_to_sns", "relay", "replace_recipient", "send", "write_to_s3"),
		NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
			"action_failure_policy": schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.ActionFailurePolicy](), Optional: true},
			"target_archive":        schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.LengthBetween(1, 2048)}},
		}},
	}
}

func bounceActionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[bounceActionModel](ctx),
		Validators: actionUnionValidators("add_header", "archive", "deliver_to_mailbox", "deliver_to_q_business", "drop", "invoke_lambda", "publish_to_sns", "relay", "replace_recipient", "send", "write_to_s3"),
		NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
			"action_failure_policy": schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.ActionFailurePolicy](), Optional: true},
			"diagnostic_message":    schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.LengthBetween(1, 256)}},
			names.AttrMessage:       schema.StringAttribute{Optional: true, Validators: []validator.String{stringvalidator.LengthBetween(1, 500)}},
			names.AttrRoleARN:       schema.StringAttribute{Required: true},
			"sender":                schema.StringAttribute{Required: true},
			"smtp_reply_code": schema.StringAttribute{Required: true, Validators: []validator.String{
				stringvalidator.RegexMatches(regexache.MustCompile(`^[45][0-9]{2}$`), "must be a 4xx or 5xx SMTP reply code"),
			}},
			names.AttrStatusCode: schema.StringAttribute{Required: true, Validators: []validator.String{
				stringvalidator.RegexMatches(regexache.MustCompile(`^[45]\.[0-9]{1,3}\.[0-9]{1,3}$`), "must be an enhanced SMTP status code beginning with 4 or 5"),
			}},
		}},
	}
}

func deliverToMailboxActionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[deliverToMailboxActionModel](ctx),
		Validators: actionUnionValidators("add_header", "archive", "bounce", "deliver_to_q_business", "drop", "invoke_lambda", "publish_to_sns", "relay", "replace_recipient", "send", "write_to_s3"),
		NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
			"action_failure_policy": schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.ActionFailurePolicy](), Optional: true},
			"mailbox_arn":           schema.StringAttribute{Required: true},
			names.AttrRoleARN:       schema.StringAttribute{Required: true},
		}},
	}
}

func deliverToQBusinessActionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[deliverToQBusinessActionModel](ctx),
		Validators: actionUnionValidators("add_header", "archive", "bounce", "deliver_to_mailbox", "drop", "invoke_lambda", "publish_to_sns", "relay", "replace_recipient", "send", "write_to_s3"),
		NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
			"action_failure_policy": schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.ActionFailurePolicy](), Optional: true},
			names.AttrApplicationID: schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.LengthBetween(36, 36)}},
			"index_id":              schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.LengthBetween(36, 36)}},
			names.AttrRoleARN:       schema.StringAttribute{Required: true},
		}},
	}
}

func dropActionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType:   fwtypes.NewListNestedObjectTypeOf[dropActionModel](ctx),
		Validators:   actionUnionValidators("add_header", "archive", "bounce", "deliver_to_mailbox", "deliver_to_q_business", "invoke_lambda", "publish_to_sns", "relay", "replace_recipient", "send", "write_to_s3"),
		NestedObject: schema.NestedBlockObject{},
	}
}

func invokeLambdaActionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[invokeLambdaActionModel](ctx),
		Validators: actionUnionValidators("add_header", "archive", "bounce", "deliver_to_mailbox", "deliver_to_q_business", "drop", "publish_to_sns", "relay", "replace_recipient", "send", "write_to_s3"),
		NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
			"action_failure_policy": schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.ActionFailurePolicy](), Optional: true},
			names.AttrFunctionARN:   schema.StringAttribute{Required: true},
			"invocation_type":       schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.LambdaInvocationType](), Required: true},
			"retry_time_minutes":    schema.Int32Attribute{Optional: true, Validators: []validator.Int32{int32validator.Between(0, 2160)}},
			names.AttrRoleARN:       schema.StringAttribute{Required: true},
		}},
	}
}

func publishToSNSActionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[publishToSNSActionModel](ctx),
		Validators: actionUnionValidators("add_header", "archive", "bounce", "deliver_to_mailbox", "deliver_to_q_business", "drop", "invoke_lambda", "relay", "replace_recipient", "send", "write_to_s3"),
		NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
			"action_failure_policy": schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.ActionFailurePolicy](), Optional: true},
			"encoding":              schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.SnsNotificationEncoding](), Optional: true},
			"payload_type":          schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.SnsNotificationPayloadType](), Optional: true},
			names.AttrRoleARN:       schema.StringAttribute{Required: true},
			names.AttrTopicARN:      schema.StringAttribute{Required: true},
		}},
	}
}

func relayActionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[relayActionModel](ctx),
		Validators: actionUnionValidators("add_header", "archive", "bounce", "deliver_to_mailbox", "deliver_to_q_business", "drop", "invoke_lambda", "publish_to_sns", "replace_recipient", "send", "write_to_s3"),
		NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
			"action_failure_policy": schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.ActionFailurePolicy](), Optional: true},
			"mail_from":             schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.MailFrom](), Optional: true},
			"relay":                 schema.StringAttribute{Required: true},
		}},
	}
}

func replaceRecipientActionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[replaceRecipientActionModel](ctx),
		Validators: actionUnionValidators("add_header", "archive", "bounce", "deliver_to_mailbox", "deliver_to_q_business", "drop", "invoke_lambda", "publish_to_sns", "relay", "send", "write_to_s3"),
		NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
			"replace_with": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 100),
					listvalidator.UniqueValues(),
					listvalidator.ValueStringsAre(stringvalidator.LengthAtMost(254)),
				},
			},
		}},
	}
}

func sendActionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[sendActionModel](ctx),
		Validators: actionUnionValidators("add_header", "archive", "bounce", "deliver_to_mailbox", "deliver_to_q_business", "drop", "invoke_lambda", "publish_to_sns", "relay", "replace_recipient", "write_to_s3"),
		NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
			"action_failure_policy": schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.ActionFailurePolicy](), Optional: true},
			names.AttrRoleARN:       schema.StringAttribute{Required: true},
		}},
	}
}

func writeToS3ActionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[writeToS3ActionModel](ctx),
		Validators: actionUnionValidators("add_header", "archive", "bounce", "deliver_to_mailbox", "deliver_to_q_business", "drop", "invoke_lambda", "publish_to_sns", "relay", "replace_recipient", "send"),
		NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
			"action_failure_policy": schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.ActionFailurePolicy](), Optional: true},
			names.AttrRoleARN:       schema.StringAttribute{Required: true},
			names.AttrS3Bucket:      schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.LengthBetween(1, 62)}},
			"s3_prefix":             schema.StringAttribute{Optional: true, Validators: []validator.String{stringvalidator.LengthBetween(1, 62)}},
			"s3_sse_kms_key_id":     schema.StringAttribute{Optional: true, Validators: []validator.String{stringvalidator.LengthBetween(20, 2048)}},
		}},
	}
}

func (r *ruleSetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().MailManagerClient(ctx)
	var plan ruleSetResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}
	var input mailmanager.CreateRuleSetInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("RuleSet")))
	if resp.Diagnostics.HasError() {
		return
	}
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)
	out, err := conn.CreateRuleSet(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	created, err := findRuleSetByID(ctx, conn, aws.ToString(out.RuleSetId))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, aws.ToString(out.RuleSetId))
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, created, &plan, flex.WithFieldNamePrefix("RuleSet")))
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *ruleSetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().MailManagerClient(ctx)
	var state ruleSetResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()
	out, err := findRuleSetByID(ctx, conn, id)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state, flex.WithFieldNamePrefix("RuleSet")))
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *ruleSetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().MailManagerClient(ctx)
	var plan, state ruleSetResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}
	if diff.HasChanges() {
		var input mailmanager.UpdateRuleSetInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("RuleSet")))
		if resp.Diagnostics.HasError() {
			return
		}
		if _, err := conn.UpdateRuleSet(ctx, &input); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}
		out, err := findRuleSetByID(ctx, conn, plan.ID.ValueString())
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan, flex.WithFieldNamePrefix("RuleSet")))
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		// Tag updates are handled by the provider's tagging interceptor.
		plan.ARN, plan.CreatedDate, plan.LastModificationDate = state.ARN, state.CreatedDate, state.LastModificationDate
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *ruleSetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().MailManagerClient(ctx)
	var state ruleSetResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()
	input := mailmanager.DeleteRuleSetInput{
		RuleSetId: aws.String(id),
	}
	_, err := conn.DeleteRuleSet(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
	}
}

func findRuleSetByID(ctx context.Context, conn *mailmanager.Client, id string) (*mailmanager.GetRuleSetOutput, error) {
	input := mailmanager.GetRuleSetInput{
		RuleSetId: aws.String(id),
	}
	out, err := conn.GetRuleSet(ctx, &input)
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

var (
	_ flex.Expander  = ruleSetActionModel{}
	_ flex.Flattener = &ruleSetActionModel{}
	_ flex.Expander  = ruleSetConditionModel{}
	_ flex.Flattener = &ruleSetConditionModel{}
	_ flex.Expander  = ruleBooleanEvaluateModel{}
	_ flex.Flattener = &ruleBooleanEvaluateModel{}
	_ flex.Expander  = ruleIPEvaluateModel{}
	_ flex.Flattener = &ruleIPEvaluateModel{}
	_ flex.Expander  = ruleNumberEvaluateModel{}
	_ flex.Flattener = &ruleNumberEvaluateModel{}
	_ flex.Expander  = ruleStringEvaluateModel{}
	_ flex.Flattener = &ruleStringEvaluateModel{}
	_ flex.Expander  = ruleVerdictEvaluateModel{}
	_ flex.Flattener = &ruleVerdictEvaluateModel{}
)

func (m ruleSetConditionModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.BooleanExpression.IsNull():
		var x awstypes.RuleConditionMemberBooleanExpression
		diags.Append(flex.Expand(ctx, m.BooleanExpression, &x.Value)...)
		return &x, diags
	case !m.DMARCExpression.IsNull():
		var x awstypes.RuleConditionMemberDmarcExpression
		diags.Append(flex.Expand(ctx, m.DMARCExpression, &x.Value)...)
		return &x, diags
	case !m.IPExpression.IsNull():
		var x awstypes.RuleConditionMemberIpExpression
		diags.Append(flex.Expand(ctx, m.IPExpression, &x.Value)...)
		return &x, diags
	case !m.NumberExpression.IsNull():
		var x awstypes.RuleConditionMemberNumberExpression
		diags.Append(flex.Expand(ctx, m.NumberExpression, &x.Value)...)
		return &x, diags
	case !m.StringExpression.IsNull():
		var x awstypes.RuleConditionMemberStringExpression
		diags.Append(flex.Expand(ctx, m.StringExpression, &x.Value)...)
		return &x, diags
	case !m.VerdictExpression.IsNull():
		var x awstypes.RuleConditionMemberVerdictExpression
		diags.Append(flex.Expand(ctx, m.VerdictExpression, &x.Value)...)
		return &x, diags
	}
	return nil, diags
}

func (m *ruleSetConditionModel) Flatten(ctx context.Context, value any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch x := value.(type) {
	case awstypes.RuleConditionMemberBooleanExpression:
		var model ruleBooleanExpressionModel
		diags.Append(flex.Flatten(ctx, x.Value, &model)...)
		m.BooleanExpression = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	case awstypes.RuleConditionMemberDmarcExpression:
		var model ruleDMARCExpressionModel
		diags.Append(flex.Flatten(ctx, x.Value, &model)...)
		m.DMARCExpression = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	case awstypes.RuleConditionMemberIpExpression:
		var model ruleIPExpressionModel
		diags.Append(flex.Flatten(ctx, x.Value, &model)...)
		m.IPExpression = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	case awstypes.RuleConditionMemberNumberExpression:
		var model ruleNumberExpressionModel
		diags.Append(flex.Flatten(ctx, x.Value, &model)...)
		m.NumberExpression = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	case awstypes.RuleConditionMemberStringExpression:
		var model ruleStringExpressionModel
		diags.Append(flex.Flatten(ctx, x.Value, &model)...)
		m.StringExpression = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	case awstypes.RuleConditionMemberVerdictExpression:
		var model ruleVerdictExpressionModel
		diags.Append(flex.Flatten(ctx, x.Value, &model)...)
		m.VerdictExpression = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	default:
		diags.AddError("Unexpected Type", fmt.Sprintf("rule condition flatten: %T", value))
	}
	return diags
}

func (m ruleBooleanEvaluateModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.Analysis.IsNull():
		var x awstypes.RuleBooleanToEvaluateMemberAnalysis
		diags.Append(flex.Expand(ctx, m.Analysis, &x.Value)...)
		return &x, diags
	case !m.Attribute.IsNull():
		return &awstypes.RuleBooleanToEvaluateMemberAttribute{Value: m.Attribute.ValueEnum()}, diags
	case !m.IsInAddressList.IsNull():
		var x awstypes.RuleBooleanToEvaluateMemberIsInAddressList
		diags.Append(flex.Expand(ctx, m.IsInAddressList, &x.Value)...)
		return &x, diags
	}
	return nil, diags
}

func (m *ruleBooleanEvaluateModel) Flatten(ctx context.Context, value any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch x := value.(type) {
	case awstypes.RuleBooleanToEvaluateMemberAnalysis:
		var model analysisModel
		diags.Append(flex.Flatten(ctx, x.Value, &model)...)
		m.Analysis = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	case awstypes.RuleBooleanToEvaluateMemberAttribute:
		m.Attribute = fwtypes.StringEnumValue(x.Value)
	case awstypes.RuleBooleanToEvaluateMemberIsInAddressList:
		var model ruleIsInAddressListModel
		diags.Append(flex.Flatten(ctx, x.Value, &model)...)
		m.IsInAddressList = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	default:
		diags.AddError("Unexpected Type", fmt.Sprintf("rule boolean evaluate flatten: %T", value))
	}
	return diags
}

func (m ruleIPEvaluateModel) Expand(context.Context) (any, diag.Diagnostics) {
	return &awstypes.RuleIpToEvaluateMemberAttribute{Value: m.Attribute.ValueEnum()}, nil
}

func (m *ruleIPEvaluateModel) Flatten(_ context.Context, value any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch x := value.(type) {
	case awstypes.RuleIpToEvaluateMemberAttribute:
		m.Attribute = fwtypes.StringEnumValue(x.Value)
	default:
		diags.AddError("Unexpected Type", fmt.Sprintf("rule IP evaluate flatten: %T", value))
	}
	return diags
}

func (m ruleNumberEvaluateModel) Expand(context.Context) (any, diag.Diagnostics) {
	return &awstypes.RuleNumberToEvaluateMemberAttribute{Value: m.Attribute.ValueEnum()}, nil
}

func (m *ruleNumberEvaluateModel) Flatten(_ context.Context, value any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch x := value.(type) {
	case awstypes.RuleNumberToEvaluateMemberAttribute:
		m.Attribute = fwtypes.StringEnumValue(x.Value)
	default:
		diags.AddError("Unexpected Type", fmt.Sprintf("rule number evaluate flatten: %T", value))
	}
	return diags
}

func (m ruleStringEvaluateModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.Analysis.IsNull():
		var x awstypes.RuleStringToEvaluateMemberAnalysis
		diags.Append(flex.Expand(ctx, m.Analysis, &x.Value)...)
		return &x, diags
	case !m.Attribute.IsNull():
		return &awstypes.RuleStringToEvaluateMemberAttribute{Value: m.Attribute.ValueEnum()}, diags
	case !m.ClientCertificateAttribute.IsNull():
		return &awstypes.RuleStringToEvaluateMemberClientCertificateAttribute{Value: m.ClientCertificateAttribute.ValueEnum()}, diags
	case !m.MIMEHeaderAttribute.IsNull():
		return &awstypes.RuleStringToEvaluateMemberMimeHeaderAttribute{Value: m.MIMEHeaderAttribute.ValueString()}, diags
	}
	return nil, diags
}

func (m *ruleStringEvaluateModel) Flatten(ctx context.Context, value any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch x := value.(type) {
	case awstypes.RuleStringToEvaluateMemberAnalysis:
		var model analysisModel
		diags.Append(flex.Flatten(ctx, x.Value, &model)...)
		m.Analysis = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	case awstypes.RuleStringToEvaluateMemberAttribute:
		m.Attribute = fwtypes.StringEnumValue(x.Value)
	case awstypes.RuleStringToEvaluateMemberClientCertificateAttribute:
		m.ClientCertificateAttribute = fwtypes.StringEnumValue(x.Value)
	case awstypes.RuleStringToEvaluateMemberMimeHeaderAttribute:
		m.MIMEHeaderAttribute = types.StringValue(x.Value)
	default:
		diags.AddError("Unexpected Type", fmt.Sprintf("rule string evaluate flatten: %T", value))
	}
	return diags
}

func (m ruleVerdictEvaluateModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.Analysis.IsNull():
		var x awstypes.RuleVerdictToEvaluateMemberAnalysis
		diags.Append(flex.Expand(ctx, m.Analysis, &x.Value)...)
		return &x, diags
	case !m.Attribute.IsNull():
		return &awstypes.RuleVerdictToEvaluateMemberAttribute{Value: m.Attribute.ValueEnum()}, diags
	}
	return nil, diags
}

func (m *ruleVerdictEvaluateModel) Flatten(ctx context.Context, value any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch x := value.(type) {
	case awstypes.RuleVerdictToEvaluateMemberAnalysis:
		var model analysisModel
		diags.Append(flex.Flatten(ctx, x.Value, &model)...)
		m.Analysis = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	case awstypes.RuleVerdictToEvaluateMemberAttribute:
		m.Attribute = fwtypes.StringEnumValue(x.Value)
	default:
		diags.AddError("Unexpected Type", fmt.Sprintf("rule verdict evaluate flatten: %T", value))
	}
	return diags
}

func (m ruleSetActionModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.AddHeader.IsNull():
		var x awstypes.RuleActionMemberAddHeader
		diags.Append(flex.Expand(ctx, m.AddHeader, &x.Value)...)
		return &x, diags
	case !m.Archive.IsNull():
		var x awstypes.RuleActionMemberArchive
		diags.Append(flex.Expand(ctx, m.Archive, &x.Value)...)
		return &x, diags
	case !m.Bounce.IsNull():
		var x awstypes.RuleActionMemberBounce
		diags.Append(flex.Expand(ctx, m.Bounce, &x.Value)...)
		return &x, diags
	case !m.DeliverToMailbox.IsNull():
		var x awstypes.RuleActionMemberDeliverToMailbox
		diags.Append(flex.Expand(ctx, m.DeliverToMailbox, &x.Value)...)
		return &x, diags
	case !m.DeliverToQBusiness.IsNull():
		var x awstypes.RuleActionMemberDeliverToQBusiness
		diags.Append(flex.Expand(ctx, m.DeliverToQBusiness, &x.Value)...)
		return &x, diags
	case !m.Drop.IsNull():
		return &awstypes.RuleActionMemberDrop{Value: awstypes.DropAction{}}, diags
	case !m.InvokeLambda.IsNull():
		var x awstypes.RuleActionMemberInvokeLambda
		diags.Append(flex.Expand(ctx, m.InvokeLambda, &x.Value)...)
		return &x, diags
	case !m.PublishToSNS.IsNull():
		var x awstypes.RuleActionMemberPublishToSns
		diags.Append(flex.Expand(ctx, m.PublishToSNS, &x.Value)...)
		return &x, diags
	case !m.Relay.IsNull():
		var x awstypes.RuleActionMemberRelay
		diags.Append(flex.Expand(ctx, m.Relay, &x.Value)...)
		return &x, diags
	case !m.ReplaceRecipient.IsNull():
		var x awstypes.RuleActionMemberReplaceRecipient
		diags.Append(flex.Expand(ctx, m.ReplaceRecipient, &x.Value)...)
		return &x, diags
	case !m.Send.IsNull():
		var x awstypes.RuleActionMemberSend
		diags.Append(flex.Expand(ctx, m.Send, &x.Value)...)
		return &x, diags
	case !m.WriteToS3.IsNull():
		var x awstypes.RuleActionMemberWriteToS3
		diags.Append(flex.Expand(ctx, m.WriteToS3, &x.Value)...)
		return &x, diags
	}
	return nil, diags
}

func (m *ruleSetActionModel) Flatten(ctx context.Context, value any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch x := value.(type) {
	case awstypes.RuleActionMemberAddHeader:
		var model addHeaderActionModel
		diags.Append(flex.Flatten(ctx, x.Value, &model)...)
		m.AddHeader = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	case awstypes.RuleActionMemberArchive:
		var model archiveActionModel
		diags.Append(flex.Flatten(ctx, x.Value, &model)...)
		m.Archive = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	case awstypes.RuleActionMemberBounce:
		var model bounceActionModel
		diags.Append(flex.Flatten(ctx, x.Value, &model)...)
		m.Bounce = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	case awstypes.RuleActionMemberDeliverToMailbox:
		var model deliverToMailboxActionModel
		diags.Append(flex.Flatten(ctx, x.Value, &model)...)
		m.DeliverToMailbox = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	case awstypes.RuleActionMemberDeliverToQBusiness:
		var model deliverToQBusinessActionModel
		diags.Append(flex.Flatten(ctx, x.Value, &model)...)
		m.DeliverToQBusiness = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	case awstypes.RuleActionMemberDrop:
		m.Drop = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &dropActionModel{})
	case awstypes.RuleActionMemberInvokeLambda:
		var model invokeLambdaActionModel
		diags.Append(flex.Flatten(ctx, x.Value, &model)...)
		m.InvokeLambda = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	case awstypes.RuleActionMemberPublishToSns:
		var model publishToSNSActionModel
		diags.Append(flex.Flatten(ctx, x.Value, &model)...)
		m.PublishToSNS = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	case awstypes.RuleActionMemberRelay:
		var model relayActionModel
		diags.Append(flex.Flatten(ctx, x.Value, &model)...)
		m.Relay = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	case awstypes.RuleActionMemberReplaceRecipient:
		var model replaceRecipientActionModel
		diags.Append(flex.Flatten(ctx, x.Value, &model)...)
		m.ReplaceRecipient = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	case awstypes.RuleActionMemberSend:
		var model sendActionModel
		diags.Append(flex.Flatten(ctx, x.Value, &model)...)
		m.Send = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	case awstypes.RuleActionMemberWriteToS3:
		var model writeToS3ActionModel
		diags.Append(flex.Flatten(ctx, x.Value, &model)...)
		m.WriteToS3 = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	default:
		diags.AddError("Unexpected Type", fmt.Sprintf("rule action flatten: %T", value))
	}
	return diags
}

type ruleSetResourceModel struct {
	framework.WithRegionModel
	ARN                  types.String                                      `tfsdk:"arn"`
	CreatedDate          timetypes.RFC3339                                 `tfsdk:"created_date"`
	ID                   types.String                                      `tfsdk:"id"`
	LastModificationDate timetypes.RFC3339                                 `tfsdk:"last_modification_date"`
	Name                 types.String                                      `tfsdk:"name"`
	Rules                fwtypes.ListNestedObjectValueOf[ruleSetRuleModel] `tfsdk:"rule"`
	Tags                 tftags.Map                                        `tfsdk:"tags"`
	TagsAll              tftags.Map                                        `tfsdk:"tags_all"`
}

type ruleSetRuleModel struct {
	Actions    fwtypes.ListNestedObjectValueOf[ruleSetActionModel]    `tfsdk:"action"`
	Conditions fwtypes.ListNestedObjectValueOf[ruleSetConditionModel] `tfsdk:"condition"`
	Name       types.String                                           `tfsdk:"name"`
	Unless     fwtypes.ListNestedObjectValueOf[ruleSetConditionModel] `tfsdk:"unless"`
}

type ruleSetConditionModel struct {
	BooleanExpression fwtypes.ListNestedObjectValueOf[ruleBooleanExpressionModel] `tfsdk:"boolean_expression"`
	DMARCExpression   fwtypes.ListNestedObjectValueOf[ruleDMARCExpressionModel]   `tfsdk:"dmarc_expression"`
	IPExpression      fwtypes.ListNestedObjectValueOf[ruleIPExpressionModel]      `tfsdk:"ip_expression"`
	NumberExpression  fwtypes.ListNestedObjectValueOf[ruleNumberExpressionModel]  `tfsdk:"number_expression"`
	StringExpression  fwtypes.ListNestedObjectValueOf[ruleStringExpressionModel]  `tfsdk:"string_expression"`
	VerdictExpression fwtypes.ListNestedObjectValueOf[ruleVerdictExpressionModel] `tfsdk:"verdict_expression"`
}

type ruleBooleanExpressionModel struct {
	Evaluate fwtypes.ListNestedObjectValueOf[ruleBooleanEvaluateModel] `tfsdk:"evaluate"`
	Operator fwtypes.StringEnum[awstypes.RuleBooleanOperator]          `tfsdk:"operator"`
}

type ruleBooleanEvaluateModel struct {
	Analysis        fwtypes.ListNestedObjectValueOf[analysisModel]            `tfsdk:"analysis"`
	Attribute       fwtypes.StringEnum[awstypes.RuleBooleanEmailAttribute]    `tfsdk:"attribute"`
	IsInAddressList fwtypes.ListNestedObjectValueOf[ruleIsInAddressListModel] `tfsdk:"is_in_address_list"`
}

type ruleIsInAddressListModel struct {
	AddressLists fwtypes.ListValueOf[types.String]                          `tfsdk:"address_lists"`
	Attribute    fwtypes.StringEnum[awstypes.RuleAddressListEmailAttribute] `tfsdk:"attribute"`
}

type ruleDMARCExpressionModel struct {
	Operator fwtypes.StringEnum[awstypes.RuleDmarcOperator]                    `tfsdk:"operator"`
	Values   fwtypes.ListValueOf[fwtypes.StringEnum[awstypes.RuleDmarcPolicy]] `tfsdk:"values"`
}

type ruleIPExpressionModel struct {
	Evaluate fwtypes.ListNestedObjectValueOf[ruleIPEvaluateModel] `tfsdk:"evaluate"`
	Operator fwtypes.StringEnum[awstypes.RuleIpOperator]          `tfsdk:"operator"`
	Values   fwtypes.ListValueOf[types.String]                    `tfsdk:"values"`
}

type ruleIPEvaluateModel struct {
	Attribute fwtypes.StringEnum[awstypes.RuleIpEmailAttribute] `tfsdk:"attribute"`
}

type ruleNumberExpressionModel struct {
	Evaluate fwtypes.ListNestedObjectValueOf[ruleNumberEvaluateModel] `tfsdk:"evaluate"`
	Operator fwtypes.StringEnum[awstypes.RuleNumberOperator]          `tfsdk:"operator"`
	Value    types.Float64                                            `tfsdk:"value"`
}

type ruleNumberEvaluateModel struct {
	Attribute fwtypes.StringEnum[awstypes.RuleNumberEmailAttribute] `tfsdk:"attribute"`
}

type ruleStringExpressionModel struct {
	Evaluate fwtypes.ListNestedObjectValueOf[ruleStringEvaluateModel] `tfsdk:"evaluate"`
	Operator fwtypes.StringEnum[awstypes.RuleStringOperator]          `tfsdk:"operator"`
	Values   fwtypes.ListValueOf[types.String]                        `tfsdk:"values"`
}

type ruleStringEvaluateModel struct {
	Analysis                   fwtypes.ListNestedObjectValueOf[analysisModel]              `tfsdk:"analysis"`
	Attribute                  fwtypes.StringEnum[awstypes.RuleStringEmailAttribute]       `tfsdk:"attribute"`
	ClientCertificateAttribute fwtypes.StringEnum[awstypes.RuleClientCertificateAttribute] `tfsdk:"client_certificate_attribute"`
	MIMEHeaderAttribute        types.String                                                `tfsdk:"mime_header_attribute"`
}

type ruleVerdictExpressionModel struct {
	Evaluate fwtypes.ListNestedObjectValueOf[ruleVerdictEvaluateModel]     `tfsdk:"evaluate"`
	Operator fwtypes.StringEnum[awstypes.RuleVerdictOperator]              `tfsdk:"operator"`
	Values   fwtypes.ListValueOf[fwtypes.StringEnum[awstypes.RuleVerdict]] `tfsdk:"values"`
}

type ruleVerdictEvaluateModel struct {
	Analysis  fwtypes.ListNestedObjectValueOf[analysisModel]    `tfsdk:"analysis"`
	Attribute fwtypes.StringEnum[awstypes.RuleVerdictAttribute] `tfsdk:"attribute"`
}

type ruleSetActionModel struct {
	AddHeader          fwtypes.ListNestedObjectValueOf[addHeaderActionModel]          `tfsdk:"add_header"`
	Archive            fwtypes.ListNestedObjectValueOf[archiveActionModel]            `tfsdk:"archive"`
	Bounce             fwtypes.ListNestedObjectValueOf[bounceActionModel]             `tfsdk:"bounce"`
	DeliverToMailbox   fwtypes.ListNestedObjectValueOf[deliverToMailboxActionModel]   `tfsdk:"deliver_to_mailbox"`
	DeliverToQBusiness fwtypes.ListNestedObjectValueOf[deliverToQBusinessActionModel] `tfsdk:"deliver_to_q_business"`
	Drop               fwtypes.ListNestedObjectValueOf[dropActionModel]               `tfsdk:"drop"`
	InvokeLambda       fwtypes.ListNestedObjectValueOf[invokeLambdaActionModel]       `tfsdk:"invoke_lambda"`
	PublishToSNS       fwtypes.ListNestedObjectValueOf[publishToSNSActionModel]       `tfsdk:"publish_to_sns"`
	Relay              fwtypes.ListNestedObjectValueOf[relayActionModel]              `tfsdk:"relay"`
	ReplaceRecipient   fwtypes.ListNestedObjectValueOf[replaceRecipientActionModel]   `tfsdk:"replace_recipient"`
	Send               fwtypes.ListNestedObjectValueOf[sendActionModel]               `tfsdk:"send"`
	WriteToS3          fwtypes.ListNestedObjectValueOf[writeToS3ActionModel]          `tfsdk:"write_to_s3"`
}
type addHeaderActionModel struct {
	HeaderName  types.String `tfsdk:"header_name"`
	HeaderValue types.String `tfsdk:"header_value"`
}
type archiveActionModel struct {
	ActionFailurePolicy fwtypes.StringEnum[awstypes.ActionFailurePolicy] `tfsdk:"action_failure_policy"`
	TargetArchive       types.String                                     `tfsdk:"target_archive"`
}
type bounceActionModel struct {
	ActionFailurePolicy fwtypes.StringEnum[awstypes.ActionFailurePolicy] `tfsdk:"action_failure_policy"`
	DiagnosticMessage   types.String                                     `tfsdk:"diagnostic_message"`
	Message             types.String                                     `tfsdk:"message"`
	RoleARN             types.String                                     `tfsdk:"role_arn"`
	Sender              types.String                                     `tfsdk:"sender"`
	SMTPReplyCode       types.String                                     `tfsdk:"smtp_reply_code"`
	StatusCode          types.String                                     `tfsdk:"status_code"`
}
type deliverToMailboxActionModel struct {
	ActionFailurePolicy fwtypes.StringEnum[awstypes.ActionFailurePolicy] `tfsdk:"action_failure_policy"`
	MailboxARN          types.String                                     `tfsdk:"mailbox_arn"`
	RoleARN             types.String                                     `tfsdk:"role_arn"`
}
type deliverToQBusinessActionModel struct {
	ActionFailurePolicy fwtypes.StringEnum[awstypes.ActionFailurePolicy] `tfsdk:"action_failure_policy"`
	ApplicationID       types.String                                     `tfsdk:"application_id"`
	IndexID             types.String                                     `tfsdk:"index_id"`
	RoleARN             types.String                                     `tfsdk:"role_arn"`
}
type dropActionModel struct{}
type invokeLambdaActionModel struct {
	ActionFailurePolicy fwtypes.StringEnum[awstypes.ActionFailurePolicy]  `tfsdk:"action_failure_policy"`
	FunctionARN         types.String                                      `tfsdk:"function_arn"`
	InvocationType      fwtypes.StringEnum[awstypes.LambdaInvocationType] `tfsdk:"invocation_type"`
	RetryTimeMinutes    types.Int32                                       `tfsdk:"retry_time_minutes"`
	RoleARN             types.String                                      `tfsdk:"role_arn"`
}
type publishToSNSActionModel struct {
	ActionFailurePolicy fwtypes.StringEnum[awstypes.ActionFailurePolicy]        `tfsdk:"action_failure_policy"`
	Encoding            fwtypes.StringEnum[awstypes.SnsNotificationEncoding]    `tfsdk:"encoding"`
	PayloadType         fwtypes.StringEnum[awstypes.SnsNotificationPayloadType] `tfsdk:"payload_type"`
	RoleARN             types.String                                            `tfsdk:"role_arn"`
	TopicARN            types.String                                            `tfsdk:"topic_arn"`
}
type relayActionModel struct {
	ActionFailurePolicy fwtypes.StringEnum[awstypes.ActionFailurePolicy] `tfsdk:"action_failure_policy"`
	MailFrom            fwtypes.StringEnum[awstypes.MailFrom]            `tfsdk:"mail_from"`
	Relay               types.String                                     `tfsdk:"relay"`
}
type replaceRecipientActionModel struct {
	ReplaceWith fwtypes.ListValueOf[types.String] `tfsdk:"replace_with"`
}
type sendActionModel struct {
	ActionFailurePolicy fwtypes.StringEnum[awstypes.ActionFailurePolicy] `tfsdk:"action_failure_policy"`
	RoleARN             types.String                                     `tfsdk:"role_arn"`
}
type writeToS3ActionModel struct {
	ActionFailurePolicy fwtypes.StringEnum[awstypes.ActionFailurePolicy] `tfsdk:"action_failure_policy"`
	RoleARN             types.String                                     `tfsdk:"role_arn"`
	S3Bucket            types.String                                     `tfsdk:"s3_bucket"`
	S3Prefix            types.String                                     `tfsdk:"s3_prefix"`
	S3SSEKMSKeyID       types.String                                     `tfsdk:"s3_sse_kms_key_id"`
}
