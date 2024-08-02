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
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
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
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types;awstypes;awstypes.Rule")
// @Testing(importIgnore="action.0.forward")
func resourceListenerRule() *schema.Resource {
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
			names.AttrAction: {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
									names.AttrScope: {
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
									names.AttrClientID: {
										Type:     schema.TypeString,
										Required: true,
									},
									names.AttrClientSecret: {
										Type:      schema.TypeString,
										Required:  true,
										Sensitive: true,
									},
									names.AttrIssuer: {
										Type:     schema.TypeString,
										Required: true,
									},
									"on_unauthenticated_request": {
										Type:             schema.TypeString,
										Optional:         true,
										Computed:         true,
										ValidateDiagFunc: enum.ValidateIgnoreCase[awstypes.AuthenticateOidcActionConditionalBehaviorEnum](),
									},
									names.AttrScope: {
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
						"fixed_response": {
							Type:             schema.TypeList,
							Optional:         true,
							DiffSuppressFunc: suppressIfActionTypeNot(awstypes.ActionTypeEnumFixedResponse),
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrContentType: {
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
									names.AttrStatusCode: {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[245]\d\d$`), ""),
									},
								},
							},
						},
						"forward": {
							Type:                  schema.TypeList,
							Optional:              true,
							DiffSuppressOnRefresh: true,
							DiffSuppressFunc:      diffSuppressMissingForward(names.AttrAction),
							MaxItems:              1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"stickiness": {
										Type:             schema.TypeList,
										Optional:         true,
										DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
										MaxItems:         1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrDuration: {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntBetween(1, 604800),
												},
												names.AttrEnabled: {
													Type:     schema.TypeBool,
													Optional: true,
													Default:  false,
												},
											},
										},
									},
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
												names.AttrWeight: {
													Type:         schema.TypeInt,
													ValidateFunc: validation.IntBetween(0, 999),
													Default:      1,
													Optional:     true,
												},
											},
										},
									},
								},
							},
						},
						"order": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntBetween(listenerActionOrderMin, listenerActionOrderMax),
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
									names.AttrPath: {
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
									names.AttrProtocol: {
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
									names.AttrStatusCode: {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.RedirectActionStatusCodeEnum](),
									},
								},
							},
						},
						"target_group_arn": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressIfActionTypeNot(awstypes.ActionTypeEnumForward),
							ValidateFunc:     verify.ValidARN,
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.ValidateIgnoreCase[awstypes.ActionTypeEnum](),
						},
					},
				},
			},
			names.AttrCondition: {
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
									names.AttrValues: {
										Type:     schema.TypeSet,
										Required: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 128),
										},
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
									names.AttrValues: {
										Type: schema.TypeSet,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 128),
										},
										Required: true,
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
									names.AttrValues: {
										Type: schema.TypeSet,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[A-Za-z-_]{1,40}$`), ""),
										},
										Required: true,
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
									names.AttrValues: {
										Type:     schema.TypeSet,
										Required: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 128),
										},
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
									names.AttrValues: {
										Type: schema.TypeSet,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: verify.ValidCIDRNetworkAddress,
										},
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"listener_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrPriority: {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     false,
				ValidateFunc: validListenerRulePriority,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: customdiff.All(
			verify.SetTagsDiff,
			validateListenerActionsCustomDiff(names.AttrAction),
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
		Tags:        getTagsIn(ctx),
	}

	input.Actions = expandLbListenerActions(cty.GetAttrPath(names.AttrAction), d.Get(names.AttrAction).([]any), &diags)
	if diags.HasError() {
		return diags
	}

	var err error

	input.Conditions, err = expandRuleConditions(d.Get(names.AttrCondition).(*schema.Set).List())
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
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, d.Id(), tags)

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

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, elbv2PropagationTimeout, func() (interface{}, error) {
		return findListenerRuleByARN(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ELBv2 Listener Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELBv2 Listener Rule (%s): %s", d.Id(), err)
	}

	rule := outputRaw.(*awstypes.Rule)

	d.Set(names.AttrARN, rule.RuleArn)

	// The listener arn isn't in the response but can be derived from the rule arn
	d.Set("listener_arn", listenerARNFromRuleARN(aws.ToString(rule.RuleArn)))

	// Rules are evaluated in priority order, from the lowest value to the highest value. The default rule has the lowest priority.
	if v := aws.ToString(rule.Priority); v == "default" {
		d.Set(names.AttrPriority, listenerRulePriorityDefault)
	} else {
		if v, err := strconv.Atoi(v); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		} else {
			d.Set(names.AttrPriority, v)
		}
	}

	sort.Slice(rule.Actions, func(i, j int) bool {
		return aws.ToInt32(rule.Actions[i].Order) < aws.ToInt32(rule.Actions[j].Order)
	})
	if err := d.Set(names.AttrAction, flattenLbListenerActions(d, names.AttrAction, rule.Actions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting action: %s", err)
	}

	conditions := make([]interface{}, len(rule.Conditions))
	for i, condition := range rule.Conditions {
		conditionMap := make(map[string]interface{})

		switch aws.ToString(condition.Field) {
		case "host-header":
			conditionMap["host_header"] = []interface{}{
				map[string]interface{}{
					names.AttrValues: flex.FlattenStringValueSet(condition.HostHeaderConfig.Values),
				},
			}

		case "http-header":
			conditionMap["http_header"] = []interface{}{
				map[string]interface{}{
					"http_header_name": aws.ToString(condition.HttpHeaderConfig.HttpHeaderName),
					names.AttrValues:   flex.FlattenStringValueSet(condition.HttpHeaderConfig.Values),
				},
			}

		case "http-request-method":
			conditionMap["http_request_method"] = []interface{}{
				map[string]interface{}{
					names.AttrValues: flex.FlattenStringValueSet(condition.HttpRequestMethodConfig.Values),
				},
			}

		case "path-pattern":
			conditionMap["path_pattern"] = []interface{}{
				map[string]interface{}{
					names.AttrValues: flex.FlattenStringValueSet(condition.PathPatternConfig.Values),
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
					names.AttrValues: flex.FlattenStringValueSet(condition.SourceIpConfig.Values),
				},
			}
		}

		conditions[i] = conditionMap
	}
	if err := d.Set(names.AttrCondition, conditions); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting condition: %s", err)
	}

	return diags
}

func resourceListenerRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		if d.HasChange(names.AttrPriority) {
			input := &elasticloadbalancingv2.SetRulePrioritiesInput{
				RulePriorities: []awstypes.RulePriorityPair{
					{
						RuleArn:  aws.String(d.Id()),
						Priority: aws.Int32(int32(d.Get(names.AttrPriority).(int))),
					},
				},
			}

			_, err := conn.SetRulePriorities(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating ELB v2 Listener Rule (%s): setting priority: %s", d.Id(), err)
			}
		}

		requestUpdate := false
		input := &elasticloadbalancingv2.ModifyRuleInput{
			RuleArn: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrAction) {
			input.Actions = expandLbListenerActions(cty.GetAttrPath(names.AttrAction), d.Get(names.AttrAction).([]any), &diags)
			if diags.HasError() {
				return diags
			}
			requestUpdate = true
		}

		if d.HasChange(names.AttrCondition) {
			var err error
			input.Conditions, err = expandRuleConditions(d.Get(names.AttrCondition).(*schema.Set).List())
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

	log.Printf("[INFO] Deleting ELBv2 Listener Rule: %s", d.Id())
	_, err := conn.DeleteRule(ctx, &elasticloadbalancingv2.DeleteRuleInput{
		RuleArn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.RuleNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ELBv2 Listener Rule (%s): %s", d.Id(), err)
	}

	return diags
}

func retryListenerRuleCreate(ctx context.Context, conn *elasticloadbalancingv2.Client, d *schema.ResourceData, input *elasticloadbalancingv2.CreateRuleInput, listenerARN string) (*elasticloadbalancingv2.CreateRuleOutput, error) {
	if v, ok := d.GetOk(names.AttrPriority); ok {
		input.Priority = aws.Int32(int32(v.(int)))

		return conn.CreateRule(ctx, input)
	}

	const (
		timeout = 5 * time.Minute
	)
	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.PriorityInUseException](ctx, timeout, func() (interface{}, error) {
		priority, err := highestListenerRulePriority(ctx, conn, listenerARN)
		if err != nil {
			return nil, err
		}

		input.Priority = aws.Int32(priority + 1)
		return conn.CreateRule(ctx, input)
	})

	if err != nil {
		return nil, err
	}

	return outputRaw.(*elasticloadbalancingv2.CreateRuleOutput), nil
}

func findListenerRule(ctx context.Context, conn *elasticloadbalancingv2.Client, input *elasticloadbalancingv2.DescribeRulesInput, filter tfslices.Predicate[*awstypes.Rule]) (*awstypes.Rule, error) {
	output, err := findListenerRules(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findListenerRules(ctx context.Context, conn *elasticloadbalancingv2.Client, input *elasticloadbalancingv2.DescribeRulesInput, filter tfslices.Predicate[*awstypes.Rule]) ([]awstypes.Rule, error) {
	var output []awstypes.Rule

	err := describeRulesPages(ctx, conn, input, func(page *elasticloadbalancingv2.DescribeRulesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Rules {
			if filter(&v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if errs.IsA[*awstypes.RuleNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findListenerRuleByARN(ctx context.Context, conn *elasticloadbalancingv2.Client, arn string) (*awstypes.Rule, error) {
	input := &elasticloadbalancingv2.DescribeRulesInput{
		RuleArns: []string{arn},
	}

	return findListenerRule(ctx, conn, input, tfslices.PredicateTrue[*awstypes.Rule]())
}

func highestListenerRulePriority(ctx context.Context, conn *elasticloadbalancingv2.Client, arn string) (int32, error) {
	input := &elasticloadbalancingv2.DescribeRulesInput{
		ListenerArn: aws.String(arn),
	}
	rules, err := findListenerRules(ctx, conn, input, func(v *awstypes.Rule) bool {
		return aws.ToString(v.Priority) != "default"
	})

	if err != nil {
		return 0, err
	}

	priorities := tfslices.ApplyToAll(rules, func(v awstypes.Rule) int32 {
		return flex.StringToInt32Value(v.Priority)
	})

	if len(priorities) == 0 {
		return 0, nil
	}

	return slices.Max(priorities), nil
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

func listenerARNFromRuleARN(ruleARN string) string {
	if arnComponents := lbListenerARNFromRuleARNRegexp.FindStringSubmatch(ruleARN); len(arnComponents) > 1 {
		return arnComponents[1] + arnComponents[2]
	}

	return ""
}

func expandRuleConditions(tfList []interface{}) ([]awstypes.RuleCondition, error) {
	apiObjects := make([]awstypes.RuleCondition, len(tfList))

	for i, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]interface{})
		apiObjects[i] = awstypes.RuleCondition{}

		var field string
		var attrs int

		if hostHeader, ok := tfMap["host_header"].([]interface{}); ok && len(hostHeader) > 0 {
			field = "host-header"
			attrs += 1
			values := hostHeader[0].(map[string]interface{})[names.AttrValues].(*schema.Set)

			apiObjects[i].HostHeaderConfig = &awstypes.HostHeaderConditionConfig{
				Values: flex.ExpandStringValueSet(values),
			}
		}

		if httpHeader, ok := tfMap["http_header"].([]interface{}); ok && len(httpHeader) > 0 {
			field = "http-header"
			attrs += 1
			httpHeaderMap := httpHeader[0].(map[string]interface{})
			values := httpHeaderMap[names.AttrValues].(*schema.Set)

			apiObjects[i].HttpHeaderConfig = &awstypes.HttpHeaderConditionConfig{
				HttpHeaderName: aws.String(httpHeaderMap["http_header_name"].(string)),
				Values:         flex.ExpandStringValueSet(values),
			}
		}

		if httpRequestMethod, ok := tfMap["http_request_method"].([]interface{}); ok && len(httpRequestMethod) > 0 {
			field = "http-request-method"
			attrs += 1
			values := httpRequestMethod[0].(map[string]interface{})[names.AttrValues].(*schema.Set)

			apiObjects[i].HttpRequestMethodConfig = &awstypes.HttpRequestMethodConditionConfig{
				Values: flex.ExpandStringValueSet(values),
			}
		}

		if pathPattern, ok := tfMap["path_pattern"].([]interface{}); ok && len(pathPattern) > 0 {
			field = "path-pattern"
			attrs += 1
			values := pathPattern[0].(map[string]interface{})[names.AttrValues].(*schema.Set)

			apiObjects[i].PathPatternConfig = &awstypes.PathPatternConditionConfig{
				Values: flex.ExpandStringValueSet(values),
			}
		}

		if queryString, ok := tfMap["query_string"].(*schema.Set); ok && queryString.Len() > 0 {
			field = "query-string"
			attrs += 1
			values := queryString.List()

			apiObjects[i].QueryStringConfig = &awstypes.QueryStringConditionConfig{
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
				apiObjects[i].QueryStringConfig.Values[j] = elbValuePair
			}
		}

		if sourceIp, ok := tfMap["source_ip"].([]interface{}); ok && len(sourceIp) > 0 {
			field = "source-ip"
			attrs += 1
			values := sourceIp[0].(map[string]interface{})[names.AttrValues].(*schema.Set)

			apiObjects[i].SourceIpConfig = &awstypes.SourceIpConditionConfig{
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

		apiObjects[i].Field = aws.String(field)
	}

	return apiObjects, nil
}
