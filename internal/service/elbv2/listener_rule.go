// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	listenerRulePriorityMin     = 1
	listenerRulePriorityMax     = 50_000
	listenerRulePriorityDefault = 99_999

	listenerActionOrderMin = 1
	listenerActionOrderMax = 50_000
)

// @SDKResource("aws_alb_listener_rule", name="Listener Rule")
// @SDKResource("aws_lb_listener_rule", name="Listener Rule")
// @Tags(identifierAttribute="id")
func ResourceListenerRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceListenerRuleCreate,
		ReadWithoutTimeout:   resourceListenerRuleRead,
		UpdateWithoutTimeout: resourceListenerRuleUpdate,
		DeleteWithoutTimeout: resourceListenerRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"listener_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"priority": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     false,
				ValidateFunc: validListenerRulePriority,
			},
			"action": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.ValidateIgnoreCase[awstypes.ActionTypeEnum](),
						},
						"order": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntBetween(listenerActionOrderMin, listenerActionOrderMax),
						},

						"target_group_arn": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressIfActionTypeNot(awstypes.ActionTypeEnumForward),
							ValidateFunc:     verify.ValidARN,
						},

						"forward": {
							Type:                  schema.TypeList,
							Optional:              true,
							DiffSuppressOnRefresh: true,
							DiffSuppressFunc:      diffSuppressMissingForward("action"),
							MaxItems:              1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"target_group": {
										Type:     schema.TypeSet,
										MinItems: 1,
										MaxItems: 5,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrARN: {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: verify.ValidARN,
												},
												"weight": {
													Type:         schema.TypeInt,
													ValidateFunc: validation.IntBetween(0, 999),
													Default:      1,
													Optional:     true,
												},
											},
										},
									},
									"stickiness": {
										Type:             schema.TypeList,
										Optional:         true,
										DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
										MaxItems:         1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrEnabled: {
													Type:     schema.TypeBool,
													Optional: true,
													Default:  false,
												},
												"duration": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntBetween(1, 604800),
												},
											},
										},
									},
								},
							},
						},

						"redirect": {
							Type:             schema.TypeList,
							Optional:         true,
							DiffSuppressFunc: suppressIfActionTypeNot(awstypes.ActionTypeEnumRedirect),
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"host": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      "#{host}",
										ValidateFunc: validation.StringLenBetween(1, 128),
									},

									"path": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      "/#{path}",
										ValidateFunc: validation.StringLenBetween(1, 128),
									},

									names.AttrPort: {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "#{port}",
									},

									"protocol": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "#{protocol}",
										ValidateFunc: validation.StringInSlice([]string{
											"#{protocol}",
											"HTTP",
											"HTTPS",
										}, false),
									},

									"query": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      "#{query}",
										ValidateFunc: validation.StringLenBetween(0, 128),
									},

									"status_code": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.RedirectActionStatusCodeEnum](),
									},
								},
							},
						},

						"fixed_response": {
							Type:             schema.TypeList,
							Optional:         true,
							DiffSuppressFunc: suppressIfActionTypeNot(awstypes.ActionTypeEnumFixedResponse),
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"content_type": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											"text/plain",
											"text/css",
											"text/html",
											"application/javascript",
											"application/json",
										}, false),
									},

									"message_body": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 1024),
									},

									"status_code": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[245]\d\d$`), ""),
									},
								},
							},
						},

						"authenticate_cognito": {
							Type:             schema.TypeList,
							Optional:         true,
							DiffSuppressFunc: suppressIfActionTypeNot(awstypes.ActionTypeEnumAuthenticateCognito),
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authentication_request_extra_params": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"on_unauthenticated_request": {
										Type:             schema.TypeString,
										Optional:         true,
										Computed:         true,
										ValidateDiagFunc: enum.ValidateIgnoreCase[awstypes.AuthenticateCognitoActionConditionalBehaviorEnum](),
									},
									"scope": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "openid",
									},
									"session_cookie_name": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "AWSELBAuthSessionCookie",
									},
									"session_timeout": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  604800,
									},
									"user_pool_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"user_pool_client_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"user_pool_domain": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},

						"authenticate_oidc": {
							Type:             schema.TypeList,
							Optional:         true,
							DiffSuppressFunc: suppressIfActionTypeNot(awstypes.ActionTypeEnumAuthenticateOidc),
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authentication_request_extra_params": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"authorization_endpoint": {
										Type:     schema.TypeString,
										Required: true,
									},
									"client_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"client_secret": {
										Type:      schema.TypeString,
										Required:  true,
										Sensitive: true,
									},
									"issuer": {
										Type:     schema.TypeString,
										Required: true,
									},
									"on_unauthenticated_request": {
										Type:             schema.TypeString,
										Optional:         true,
										Computed:         true,
										ValidateDiagFunc: enum.ValidateIgnoreCase[awstypes.AuthenticateOidcActionConditionalBehaviorEnum](),
									},
									"scope": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "openid",
									},
									"session_cookie_name": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "AWSELBAuthSessionCookie",
									},
									"session_timeout": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  604800,
									},
									"token_endpoint": {
										Type:     schema.TypeString,
										Required: true,
									},
									"user_info_endpoint": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"condition": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host_header": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"values": {
										Type:     schema.TypeSet,
										Required: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 128),
										},
										Set: schema.HashString,
									},
								},
							},
						},
						"http_header": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"http_header_name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringMatch(regexache.MustCompile("^[0-9A-Za-z_!#$%&'*+,.^`|~-]{1,40}$"), ""), // was "," meant to be included? +-. creates a range including: +,-.
									},
									"values": {
										Type: schema.TypeSet,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 128),
										},
										Required: true,
										Set:      schema.HashString,
									},
								},
							},
						},
						"http_request_method": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"values": {
										Type: schema.TypeSet,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[A-Za-z-_]{1,40}$`), ""),
										},
										Required: true,
										Set:      schema.HashString,
									},
								},
							},
						},
						"path_pattern": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"values": {
										Type:     schema.TypeSet,
										Required: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 128),
										},
										Set: schema.HashString,
									},
								},
							},
						},
						"query_string": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrKey: {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrValue: {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"source_ip": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"values": {
										Type: schema.TypeSet,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: verify.ValidCIDRNetworkAddress,
										},
										Required: true,
										Set:      schema.HashString,
									},
								},
							},
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: customdiff.All(
			verify.SetTagsDiff,
			validateListenerActionsCustomDiff("action"),
		),
	}
}

func suppressIfActionTypeNot(t awstypes.ActionTypeEnum) schema.SchemaDiffSuppressFunc {
	return func(k, old, new string, d *schema.ResourceData) bool {
		take := 2
		i := strings.IndexFunc(k, func(r rune) bool {
			if r == '.' {
				take -= 1
				return take == 0
			}
			return false
		})
		at := k[:i+1] + names.AttrType
		return awstypes.ActionTypeEnum(d.Get(at).(string)) != t
	}
}

func resourceListenerRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)

	listenerARN := d.Get("listener_arn").(string)
	input := &elasticloadbalancingv2.CreateRuleInput{
		ListenerArn: aws.String(listenerARN),
		Tags:        getTagsInV2(ctx),
	}

	input.Actions = expandLbListenerActions(cty.GetAttrPath("action"), d.Get("action").([]any), &diags)
	if diags.HasError() {
		return diags
	}

	var err error

	input.Conditions, err = lbListenerRuleConditions(d.Get("condition").(*schema.Set).List())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := retryListenerRuleCreate(ctx, conn, d, input, listenerARN)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(meta.(*conns.AWSClient).Partition, err) {
		input.Tags = nil

		output, err = retryListenerRuleCreate(ctx, conn, d, input, listenerARN)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ELBv2 Listener Rule: %s", err)
	}

	d.SetId(aws.ToString(output.Rules[0].RuleArn))

	// Post-create tagging supported in some partitions
	if tags := getTagsInV2(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTagsV2(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(meta.(*conns.AWSClient).Partition, err) {
			return append(diags, resourceListenerRuleRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ELBv2 Listener Rule (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceListenerRuleRead(ctx, d, meta)...)
}

func resourceListenerRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)

	var resp *elasticloadbalancingv2.DescribeRulesOutput
	var req = &elasticloadbalancingv2.DescribeRulesInput{
		RuleArns: []string{d.Id()},
	}

	err := retry.RetryContext(ctx, 1*time.Minute, func() *retry.RetryError {
		var err error
		resp, err = conn.DescribeRules(ctx, req)
		if err != nil {
			if d.IsNewResource() && errs.IsA[*awstypes.RuleNotFoundException](err) {
				return retry.RetryableError(err)
			} else {
				return retry.NonRetryableError(err)
			}
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		resp, err = conn.DescribeRules(ctx, req)
	}
	if err != nil {
		if errs.IsA[*awstypes.RuleNotFoundException](err) {
			log.Printf("[WARN] DescribeRules - removing %s from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "retrieving Rules for listener %q: %s", d.Id(), err)
	}

	if len(resp.Rules) != 1 {
		return sdkdiag.AppendErrorf(diags, "retrieving Rule %q", d.Id())
	}

	rule := resp.Rules[0]

	d.Set(names.AttrARN, rule.RuleArn)

	// The listener arn isn't in the response but can be derived from the rule arn
	d.Set("listener_arn", ListenerARNFromRuleARN(aws.ToString(rule.RuleArn)))

	// Rules are evaluated in priority order, from the lowest value to the highest value. The default rule has the lowest priority.
	if aws.ToString(rule.Priority) == "default" {
		d.Set("priority", listenerRulePriorityDefault)
	} else {
		if priority, err := strconv.Atoi(aws.ToString(rule.Priority)); err != nil {
			return sdkdiag.AppendErrorf(diags, "Cannot convert rule priority %q to int: %s", aws.ToString(rule.Priority), err)
		} else {
			d.Set("priority", priority)
		}
	}

	sort.Slice(rule.Actions, func(i, j int) bool {
		return aws.ToInt32(rule.Actions[i].Order) < aws.ToInt32(rule.Actions[j].Order)
	})
	if err := d.Set("action", flattenLbListenerActions(d, "action", rule.Actions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting action: %s", err)
	}

	conditions := make([]interface{}, len(rule.Conditions))
	for i, condition := range rule.Conditions {
		conditionMap := make(map[string]interface{})

		switch aws.ToString(condition.Field) {
		case "host-header":
			conditionMap["host_header"] = []interface{}{
				map[string]interface{}{
					"values": flex.FlattenStringValueSet(condition.HostHeaderConfig.Values),
				},
			}

		case "http-header":
			conditionMap["http_header"] = []interface{}{
				map[string]interface{}{
					"http_header_name": aws.ToString(condition.HttpHeaderConfig.HttpHeaderName),
					"values":           flex.FlattenStringValueSet(condition.HttpHeaderConfig.Values),
				},
			}

		case "http-request-method":
			conditionMap["http_request_method"] = []interface{}{
				map[string]interface{}{
					"values": flex.FlattenStringValueSet(condition.HttpRequestMethodConfig.Values),
				},
			}

		case "path-pattern":
			conditionMap["path_pattern"] = []interface{}{
				map[string]interface{}{
					"values": flex.FlattenStringValueSet(condition.PathPatternConfig.Values),
				},
			}

		case "query-string":
			values := make([]interface{}, len(condition.QueryStringConfig.Values))
			for k, value := range condition.QueryStringConfig.Values {
				values[k] = map[string]interface{}{
					names.AttrKey:   aws.ToString(value.Key),
					names.AttrValue: aws.ToString(value.Value),
				}
			}
			conditionMap["query_string"] = values

		case "source-ip":
			conditionMap["source_ip"] = []interface{}{
				map[string]interface{}{
					"values": flex.FlattenStringValueSet(condition.SourceIpConfig.Values),
				},
			}
		}

		conditions[i] = conditionMap
	}
	if err := d.Set("condition", conditions); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting condition: %s", err)
	}

	return diags
}

func resourceListenerRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		if d.HasChange("priority") {
			params := &elasticloadbalancingv2.SetRulePrioritiesInput{
				RulePriorities: []awstypes.RulePriorityPair{
					{
						RuleArn:  aws.String(d.Id()),
						Priority: aws.Int32(int32(d.Get("priority").(int))),
					},
				},
			}

			_, err := conn.SetRulePriorities(ctx, params)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating ELB v2 Listener Rule (%s): setting priority: %s", d.Id(), err)
			}
		}

		requestUpdate := false
		input := &elasticloadbalancingv2.ModifyRuleInput{
			RuleArn: aws.String(d.Id()),
		}

		if d.HasChange("action") {
			input.Actions = expandLbListenerActions(cty.GetAttrPath("action"), d.Get("action").([]any), &diags)
			if diags.HasError() {
				return diags
			}
			requestUpdate = true
		}

		if d.HasChange("condition") {
			var err error
			input.Conditions, err = lbListenerRuleConditions(d.Get("condition").(*schema.Set).List())
			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
			requestUpdate = true
		}

		if requestUpdate {
			resp, err := conn.ModifyRule(ctx, input)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "modifying LB Listener Rule: %s", err)
			}

			if len(resp.Rules) == 0 {
				return sdkdiag.AppendErrorf(diags, "modifying creating LB Listener Rule: no rules returned in response")
			}
		}
	}

	return append(diags, resourceListenerRuleRead(ctx, d, meta)...)
}

func resourceListenerRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)

	_, err := conn.DeleteRule(ctx, &elasticloadbalancingv2.DeleteRuleInput{
		RuleArn: aws.String(d.Id()),
	})
	if err != nil && !errs.IsA[*awstypes.RuleNotFoundException](err) {
		return sdkdiag.AppendErrorf(diags, "deleting LB Listener Rule: %s", err)
	}
	return diags
}

func retryListenerRuleCreate(ctx context.Context, conn *elasticloadbalancingv2.Client, d *schema.ResourceData, params *elasticloadbalancingv2.CreateRuleInput, listenerARN string) (*elasticloadbalancingv2.CreateRuleOutput, error) {
	var resp *elasticloadbalancingv2.CreateRuleOutput
	if v, ok := d.GetOk("priority"); ok {
		var err error
		params.Priority = aws.Int32(int32(v.(int)))
		resp, err = conn.CreateRule(ctx, params)

		if err != nil {
			return nil, err
		}
	} else {
		var priority int32

		err := retry.RetryContext(ctx, 5*time.Minute, func() *retry.RetryError {
			var err error
			priority, err = highestListenerRulePriority(ctx, conn, listenerARN)
			if err != nil {
				return retry.NonRetryableError(err)
			}
			params.Priority = aws.Int32(priority + 1)
			resp, err = conn.CreateRule(ctx, params)
			if err != nil {
				if errs.IsA[*awstypes.PriorityInUseException](err) {
					return retry.RetryableError(err)
				}
				return retry.NonRetryableError(err)
			}
			return nil
		})

		if tfresource.TimedOut(err) {
			priority, err = highestListenerRulePriority(ctx, conn, listenerARN)
			if err != nil {
				return nil, fmt.Errorf("getting highest listener rule (%s) priority: %w", listenerARN, err)
			}
			params.Priority = aws.Int32(priority + 1)
			resp, err = conn.CreateRule(ctx, params)
		}

		if err != nil {
			return nil, err
		}
	}

	if resp == nil || len(resp.Rules) == 0 {
		return nil, fmt.Errorf("creating LB Listener Rule (%s): no rules returned in response", listenerARN)
	}

	return resp, nil
}

func validListenerRulePriority(v interface{}, k string) (ws []string, errors []error) {
	value := v.(int)
	if value < listenerRulePriorityMin || (value > listenerRulePriorityMax && value != listenerRulePriorityDefault) {
		errors = append(errors, fmt.Errorf("%q must be in the range %d-%d for normal rule or %d for the default rule", k, listenerRulePriorityMin, listenerRulePriorityMax, listenerRulePriorityDefault))
	}
	return
}

// from arn:
// arn:aws:elasticloadbalancing:us-east-1:012345678912:listener-rule/app/name/0123456789abcdef/abcdef0123456789/456789abcedf1234
// select submatches:
// (arn:aws:elasticloadbalancing:us-east-1:012345678912:listener)-rule(/app/name/0123456789abcdef/abcdef0123456789)/456789abcedf1234
// concat to become:
// arn:aws:elasticloadbalancing:us-east-1:012345678912:listener/app/name/0123456789abcdef/abcdef0123456789
var lbListenerARNFromRuleARNRegexp = regexache.MustCompile(`^(arn:.+:listener)-rule(/.+)/[^/]+$`)

func ListenerARNFromRuleARN(ruleArn string) string {
	if arnComponents := lbListenerARNFromRuleARNRegexp.FindStringSubmatch(ruleArn); len(arnComponents) > 1 {
		return arnComponents[1] + arnComponents[2]
	}

	return ""
}

func highestListenerRulePriority(ctx context.Context, conn *elasticloadbalancingv2.Client, arn string) (priority int32, err error) {
	var priorities []int32
	var nextMarker *string

	for {
		out, aerr := conn.DescribeRules(ctx, &elasticloadbalancingv2.DescribeRulesInput{
			ListenerArn: aws.String(arn),
			Marker:      nextMarker,
		})
		if aerr != nil {
			return 0, aerr
		}
		for _, rule := range out.Rules {
			if priority := aws.ToString(rule.Priority); priority != "default" {
				p, _ := strconv.ParseInt(priority, 0, 32)
				priorities = append(priorities, int32(p))
			}
		}
		if out.NextMarker == nil {
			break
		}
		nextMarker = out.NextMarker
	}

	if len(priorities) == 0 {
		return 0, nil
	}

	slices.Sort(priorities)

	return priorities[len(priorities)-1], nil
}

// lbListenerRuleConditions converts data source generated by Terraform into
// an elasticloadbalancingv2/types.RuleCondition object suitable for submitting to AWS API.
func lbListenerRuleConditions(conditions []interface{}) ([]awstypes.RuleCondition, error) {
	elbConditions := make([]awstypes.RuleCondition, len(conditions))
	for i, condition := range conditions {
		elbConditions[i] = awstypes.RuleCondition{}
		conditionMap := condition.(map[string]interface{})
		var field string
		var attrs int

		if hostHeader, ok := conditionMap["host_header"].([]interface{}); ok && len(hostHeader) > 0 {
			field = "host-header"
			attrs += 1
			values := hostHeader[0].(map[string]interface{})["values"].(*schema.Set)

			elbConditions[i].HostHeaderConfig = &awstypes.HostHeaderConditionConfig{
				Values: flex.ExpandStringValueSet(values),
			}
		}

		if httpHeader, ok := conditionMap["http_header"].([]interface{}); ok && len(httpHeader) > 0 {
			field = "http-header"
			attrs += 1
			httpHeaderMap := httpHeader[0].(map[string]interface{})
			values := httpHeaderMap["values"].(*schema.Set)

			elbConditions[i].HttpHeaderConfig = &awstypes.HttpHeaderConditionConfig{
				HttpHeaderName: aws.String(httpHeaderMap["http_header_name"].(string)),
				Values:         flex.ExpandStringValueSet(values),
			}
		}

		if httpRequestMethod, ok := conditionMap["http_request_method"].([]interface{}); ok && len(httpRequestMethod) > 0 {
			field = "http-request-method"
			attrs += 1
			values := httpRequestMethod[0].(map[string]interface{})["values"].(*schema.Set)

			elbConditions[i].HttpRequestMethodConfig = &awstypes.HttpRequestMethodConditionConfig{
				Values: flex.ExpandStringValueSet(values),
			}
		}

		if pathPattern, ok := conditionMap["path_pattern"].([]interface{}); ok && len(pathPattern) > 0 {
			field = "path-pattern"
			attrs += 1
			values := pathPattern[0].(map[string]interface{})["values"].(*schema.Set)

			elbConditions[i].PathPatternConfig = &awstypes.PathPatternConditionConfig{
				Values: flex.ExpandStringValueSet(values),
			}
		}

		if queryString, ok := conditionMap["query_string"].(*schema.Set); ok && queryString.Len() > 0 {
			field = "query-string"
			attrs += 1
			values := queryString.List()

			elbConditions[i].QueryStringConfig = &awstypes.QueryStringConditionConfig{
				Values: make([]awstypes.QueryStringKeyValuePair, len(values)),
			}
			for j, p := range values {
				valuePair := p.(map[string]interface{})
				elbValuePair := awstypes.QueryStringKeyValuePair{
					Value: aws.String(valuePair[names.AttrValue].(string)),
				}
				if valuePair[names.AttrKey].(string) != "" {
					elbValuePair.Key = aws.String(valuePair[names.AttrKey].(string))
				}
				elbConditions[i].QueryStringConfig.Values[j] = elbValuePair
			}
		}

		if sourceIp, ok := conditionMap["source_ip"].([]interface{}); ok && len(sourceIp) > 0 {
			field = "source-ip"
			attrs += 1
			values := sourceIp[0].(map[string]interface{})["values"].(*schema.Set)

			elbConditions[i].SourceIpConfig = &awstypes.SourceIpConditionConfig{
				Values: flex.ExpandStringValueSet(values),
			}
		}

		// FIXME Rework this and use ConflictsWith when it finally works with collections:
		// https://github.com/hashicorp/terraform/issues/13016
		// Still need to ensure that one of the condition attributes is set.
		if attrs == 0 {
			return nil, errors.New("One of host_header, http_header, http_request_method, path_pattern, query_string or source_ip must be set in a condition block")
		} else if attrs > 1 {
			return nil, errors.New("Only one of host_header, http_header, http_request_method, path_pattern, query_string or source_ip can be set in a condition block")
		}

		elbConditions[i].Field = aws.String(field)
	}
	return elbConditions, nil
}
