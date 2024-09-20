// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
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

// @SDKResource("aws_alb_listener", name="Listener")
// @SDKResource("aws_lb_listener", name="Listener")
// @Tags(identifierAttribute="id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types;awstypes;awstypes.Listener")
// @Testing(importIgnore="default_action.0.forward")
func resourceListener() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceListenerCreate,
		ReadWithoutTimeout:   resourceListenerRead,
		UpdateWithoutTimeout: resourceListenerUpdate,
		DeleteWithoutTimeout: resourceListenerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"alpn_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(alpnPolicyEnum_Values(), true),
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCertificateARN: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrDefaultAction: {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authenticate_cognito": {
							Type:             schema.TypeList,
							Optional:         true,
							DiffSuppressFunc: suppressIfDefaultActionTypeNot(awstypes.ActionTypeEnumAuthenticateCognito),
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
										Computed: true,
									},
									"session_cookie_name": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"session_timeout": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
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
							DiffSuppressFunc: suppressIfDefaultActionTypeNot(awstypes.ActionTypeEnumAuthenticateOidc),
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
										Computed: true,
									},
									"session_cookie_name": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"session_timeout": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
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
							DiffSuppressFunc: suppressIfDefaultActionTypeNot(awstypes.ActionTypeEnumFixedResponse),
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
							Type:             schema.TypeList,
							Optional:         true,
							DiffSuppressFunc: suppressIfDefaultActionTypeNot(awstypes.ActionTypeEnumForward),
							MaxItems:         1,
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
												names.AttrWeight: {
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
							DiffSuppressFunc: suppressIfDefaultActionTypeNot(awstypes.ActionTypeEnumRedirect),
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
							DiffSuppressFunc: suppressIfDefaultActionTypeNot(awstypes.ActionTypeEnumForward),
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
			"load_balancer_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"mutual_authentication": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ignore_client_certificate_expiry": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						names.AttrMode: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(mutualAuthenticationModeEnum_Values(), true),
						},
						"trust_store_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			names.AttrPort: {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IsPortNumber,
			},
			names.AttrProtocol: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				StateFunc: func(v interface{}) string {
					return strings.ToUpper(v.(string))
				},
				ValidateDiagFunc: enum.ValidateIgnoreCase[awstypes.ProtocolEnum](),
			},
			"ssl_policy": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: customdiff.All(
			verify.SetTagsDiff,
			validateListenerActionsCustomDiff(names.AttrDefaultAction),
		),
	}
}

func suppressIfDefaultActionTypeNot(t awstypes.ActionTypeEnum) schema.SchemaDiffSuppressFunc {
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

func resourceListenerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)

	lbARN := d.Get("load_balancer_arn").(string)
	input := &elasticloadbalancingv2.CreateListenerInput{
		LoadBalancerArn: aws.String(lbARN),
		Tags:            getTagsIn(ctx),
	}

	if v, ok := d.GetOk("alpn_policy"); ok {
		input.AlpnPolicy = []string{v.(string)}
	}

	if v, ok := d.GetOk(names.AttrCertificateARN); ok {
		input.Certificates = []awstypes.Certificate{{
			CertificateArn: aws.String(v.(string)),
		}}
	}

	if v, ok := d.GetOk(names.AttrDefaultAction); ok && len(v.([]interface{})) > 0 {
		input.DefaultActions = expandLbListenerActions(cty.GetAttrPath(names.AttrDefaultAction), v.([]any), &diags)
		if diags.HasError() {
			return diags
		}
	}

	if v, ok := d.GetOk("mutual_authentication"); ok {
		input.MutualAuthentication = expandMutualAuthenticationAttributes(v.([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrPort); ok {
		input.Port = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk(names.AttrProtocol); ok {
		input.Protocol = awstypes.ProtocolEnum(v.(string))
	} else if strings.Contains(lbARN, "loadbalancer/app/") {
		// Keep previous default of HTTP for Application Load Balancers.
		input.Protocol = awstypes.ProtocolEnumHttp
	}

	if v, ok := d.GetOk("ssl_policy"); ok {
		input.SslPolicy = aws.String(v.(string))
	}

	output, err := retryListenerCreate(ctx, conn, input, d.Timeout(schema.TimeoutCreate))

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(meta.(*conns.AWSClient).Partition, err) {
		input.Tags = nil

		output, err = retryListenerCreate(ctx, conn, input, d.Timeout(schema.TimeoutCreate))
	}

	// Tags are not supported on creation with some load balancer types (i.e. Gateway)
	// Retry creation without tags
	if input.Tags != nil && tfawserr.ErrMessageContains(err, errCodeValidationError, tagsOnCreationErrMessage) {
		input.Tags = nil

		output, err = retryListenerCreate(ctx, conn, input, d.Timeout(schema.TimeoutCreate))
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ELBv2 Listener (%s): %s", lbARN, err)
	}

	d.SetId(aws.ToString(output.Listeners[0].ListenerArn))

	_, err = tfresource.RetryWhenNotFound(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return findListenerByARN(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ELBv2 Listener (%s) create: %s", d.Id(), err)
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(meta.(*conns.AWSClient).Partition, err) {
			return append(diags, resourceListenerRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ELBv2 Listener (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceListenerRead(ctx, d, meta)...)
}

func resourceListenerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)

	listener, err := findListenerByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ELBv2 Listener (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELBv2 Listener (%s): %s", d.Id(), err)
	}

	if len(listener.AlpnPolicy) == 1 {
		d.Set("alpn_policy", listener.AlpnPolicy[0])
	}
	d.Set(names.AttrARN, listener.ListenerArn)
	if len(listener.Certificates) == 1 {
		d.Set(names.AttrCertificateARN, listener.Certificates[0].CertificateArn)
	}
	sort.Slice(listener.DefaultActions, func(i, j int) bool {
		return aws.ToInt32(listener.DefaultActions[i].Order) < aws.ToInt32(listener.DefaultActions[j].Order)
	})
	if err := d.Set(names.AttrDefaultAction, flattenLbListenerActions(d, names.AttrDefaultAction, listener.DefaultActions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting default_action: %s", err)
	}
	d.Set("load_balancer_arn", listener.LoadBalancerArn)
	if err := d.Set("mutual_authentication", flattenMutualAuthenticationAttributes(listener.MutualAuthentication)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting mutual_authentication: %s", err)
	}
	d.Set(names.AttrPort, listener.Port)
	d.Set(names.AttrProtocol, listener.Protocol)
	d.Set("ssl_policy", listener.SslPolicy)

	return diags
}

func resourceListenerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &elasticloadbalancingv2.ModifyListenerInput{
			ListenerArn: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("alpn_policy"); ok {
			input.AlpnPolicy = []string{v.(string)}
		}

		if v, ok := d.GetOk(names.AttrCertificateARN); ok {
			input.Certificates = []awstypes.Certificate{{
				CertificateArn: aws.String(v.(string)),
			}}
		}

		if d.HasChange(names.AttrDefaultAction) {
			input.DefaultActions = expandLbListenerActions(cty.GetAttrPath(names.AttrDefaultAction), d.Get(names.AttrDefaultAction).([]any), &diags)
			if diags.HasError() {
				return diags
			}
		}

		if d.HasChange("mutual_authentication") {
			input.MutualAuthentication = expandMutualAuthenticationAttributes(d.Get("mutual_authentication").([]interface{}))
		}

		if v, ok := d.GetOk(names.AttrPort); ok {
			input.Port = aws.Int32(int32(v.(int)))
		}

		if v, ok := d.GetOk(names.AttrProtocol); ok {
			input.Protocol = awstypes.ProtocolEnum(v.(string))
		}

		if v, ok := d.GetOk("ssl_policy"); ok {
			input.SslPolicy = aws.String(v.(string))
		}

		_, err := tfresource.RetryWhenIsA[*awstypes.CertificateNotFoundException](ctx, d.Timeout(schema.TimeoutUpdate), func() (interface{}, error) {
			return conn.ModifyListener(ctx, input)
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying ELBv2 Listener (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceListenerRead(ctx, d, meta)...)
}

func resourceListenerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)

	log.Printf("[INFO] Deleting ELBv2 Listener: %s", d.Id())
	_, err := conn.DeleteListener(ctx, &elasticloadbalancingv2.DeleteListenerInput{
		ListenerArn: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ELBv2 Listener (%s): %s", d.Id(), err)
	}

	return diags
}

func retryListenerCreate(ctx context.Context, conn *elasticloadbalancingv2.Client, input *elasticloadbalancingv2.CreateListenerInput, timeout time.Duration) (*elasticloadbalancingv2.CreateListenerOutput, error) {
	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.CertificateNotFoundException](ctx, timeout, func() (interface{}, error) {
		return conn.CreateListener(ctx, input)
	})

	if err != nil {
		return nil, err
	}

	return outputRaw.(*elasticloadbalancingv2.CreateListenerOutput), nil
}

func findListenerByARN(ctx context.Context, conn *elasticloadbalancingv2.Client, arn string) (*awstypes.Listener, error) {
	input := &elasticloadbalancingv2.DescribeListenersInput{
		ListenerArns: []string{arn},
	}
	output, err := findListener(ctx, conn, input, tfslices.PredicateTrue[*awstypes.Listener]())

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.ListenerArn) != arn {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findListener(ctx context.Context, conn *elasticloadbalancingv2.Client, input *elasticloadbalancingv2.DescribeListenersInput, filter tfslices.Predicate[*awstypes.Listener]) (*awstypes.Listener, error) {
	output, err := findListeners(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findListeners(ctx context.Context, conn *elasticloadbalancingv2.Client, input *elasticloadbalancingv2.DescribeListenersInput, filter tfslices.Predicate[*awstypes.Listener]) ([]awstypes.Listener, error) {
	var output []awstypes.Listener

	pages := elasticloadbalancingv2.NewDescribeListenersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ListenerNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Listeners {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func expandLbListenerActions(actionsPath cty.Path, l []any, diags *diag.Diagnostics) []awstypes.Action {
	if len(l) == 0 {
		return nil
	}

	var actions []awstypes.Action

	for i, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		actions = append(actions, expandLbListenerAction(actionsPath.IndexInt(i), i, tfMap, diags))
	}

	return actions
}

func expandLbListenerAction(actionPath cty.Path, i int, tfMap map[string]any, diags *diag.Diagnostics) awstypes.Action {
	action := awstypes.Action{
		Order: aws.Int32(int32(i + 1)),
		Type:  awstypes.ActionTypeEnum(tfMap[names.AttrType].(string)),
	}

	if order, ok := tfMap["order"].(int); ok && order != 0 {
		action.Order = aws.Int32(int32(order))
	}

	switch awstypes.ActionTypeEnum(tfMap[names.AttrType].(string)) {
	case awstypes.ActionTypeEnumForward:
		if v, ok := tfMap["forward"].([]interface{}); ok && len(v) > 0 {
			action.ForwardConfig = expandLbListenerActionForwardConfig(v)
		}
		if v, ok := tfMap["target_group_arn"].(string); ok && v != "" {
			action.TargetGroupArn = aws.String(v)
		}

	case awstypes.ActionTypeEnumRedirect:
		if v, ok := tfMap["redirect"].([]interface{}); ok {
			action.RedirectConfig = expandLbListenerRedirectActionConfig(v)
		}

	case awstypes.ActionTypeEnumFixedResponse:
		if v, ok := tfMap["fixed_response"].([]interface{}); ok {
			action.FixedResponseConfig = expandLbListenerFixedResponseConfig(v)
		}

	case awstypes.ActionTypeEnumAuthenticateCognito:
		if v, ok := tfMap["authenticate_cognito"].([]interface{}); ok {
			action.AuthenticateCognitoConfig = expandLbListenerAuthenticateCognitoConfig(v)
		}

	case awstypes.ActionTypeEnumAuthenticateOidc:
		if v, ok := tfMap["authenticate_oidc"].([]interface{}); ok {
			action.AuthenticateOidcConfig = expandAuthenticateOIDCConfig(v)
		}
	}

	listenerActionRuntimeValidate(actionPath, tfMap, diags)

	return action
}

func expandLbListenerAuthenticateCognitoConfig(l []interface{}) *awstypes.AuthenticateCognitoActionConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	config := &awstypes.AuthenticateCognitoActionConfig{
		AuthenticationRequestExtraParams: flex.ExpandStringValueMap(tfMap["authentication_request_extra_params"].(map[string]interface{})),
		UserPoolArn:                      aws.String(tfMap["user_pool_arn"].(string)),
		UserPoolClientId:                 aws.String(tfMap["user_pool_client_id"].(string)),
		UserPoolDomain:                   aws.String(tfMap["user_pool_domain"].(string)),
	}

	if v, ok := tfMap["on_unauthenticated_request"].(string); ok && v != "" {
		config.OnUnauthenticatedRequest = awstypes.AuthenticateCognitoActionConditionalBehaviorEnum(v)
	}

	if v, ok := tfMap[names.AttrScope].(string); ok && v != "" {
		config.Scope = aws.String(v)
	}

	if v, ok := tfMap["session_cookie_name"].(string); ok && v != "" {
		config.SessionCookieName = aws.String(v)
	}

	if v, ok := tfMap["session_timeout"].(int); ok && v != 0 {
		config.SessionTimeout = aws.Int64(int64(v))
	}

	return config
}

func expandAuthenticateOIDCConfig(l []interface{}) *awstypes.AuthenticateOidcActionConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	config := &awstypes.AuthenticateOidcActionConfig{
		AuthenticationRequestExtraParams: flex.ExpandStringValueMap(tfMap["authentication_request_extra_params"].(map[string]interface{})),
		AuthorizationEndpoint:            aws.String(tfMap["authorization_endpoint"].(string)),
		ClientId:                         aws.String(tfMap[names.AttrClientID].(string)),
		ClientSecret:                     aws.String(tfMap[names.AttrClientSecret].(string)),
		Issuer:                           aws.String(tfMap[names.AttrIssuer].(string)),
		TokenEndpoint:                    aws.String(tfMap["token_endpoint"].(string)),
		UserInfoEndpoint:                 aws.String(tfMap["user_info_endpoint"].(string)),
	}

	if v, ok := tfMap["on_unauthenticated_request"].(string); ok && v != "" {
		config.OnUnauthenticatedRequest = awstypes.AuthenticateOidcActionConditionalBehaviorEnum(v)
	}

	if v, ok := tfMap[names.AttrScope].(string); ok && v != "" {
		config.Scope = aws.String(v)
	}

	if v, ok := tfMap["session_cookie_name"].(string); ok && v != "" {
		config.SessionCookieName = aws.String(v)
	}

	if v, ok := tfMap["session_timeout"].(int); ok && v != 0 {
		config.SessionTimeout = aws.Int64(int64(v))
	}

	return config
}

func expandLbListenerFixedResponseConfig(l []interface{}) *awstypes.FixedResponseActionConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	fr := &awstypes.FixedResponseActionConfig{
		StatusCode: aws.String(tfMap[names.AttrStatusCode].(string)),
	}

	if v, ok := tfMap[names.AttrContentType]; ok {
		fr.ContentType = aws.String(v.(string))
	}
	if v, ok := tfMap["message_body"]; ok {
		fr.MessageBody = aws.String(v.(string))
	}

	return fr
}

func expandLbListenerRedirectActionConfig(l []interface{}) *awstypes.RedirectActionConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	rac := &awstypes.RedirectActionConfig{
		StatusCode: awstypes.RedirectActionStatusCodeEnum(tfMap[names.AttrStatusCode].(string)),
	}

	if v, ok := tfMap["host"]; ok {
		rac.Host = aws.String(v.(string))
	}
	if v, ok := tfMap[names.AttrPath]; ok {
		rac.Path = aws.String(v.(string))
	}
	if v, ok := tfMap[names.AttrPort]; ok {
		rac.Port = aws.String(v.(string))
	}
	if v, ok := tfMap[names.AttrProtocol]; ok {
		rac.Protocol = aws.String(v.(string))
	}
	if v, ok := tfMap["query"]; ok {
		rac.Query = aws.String(v.(string))
	}

	return rac
}

func expandLbListenerActionForwardConfig(l []interface{}) *awstypes.ForwardActionConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &awstypes.ForwardActionConfig{}

	if v, ok := tfMap["target_group"].(*schema.Set); ok && v.Len() > 0 {
		config.TargetGroups = expandLbListenerActionForwardConfigTargetGroups(v.List())
	}

	if v, ok := tfMap["stickiness"].([]interface{}); ok && len(v) > 0 {
		config.TargetGroupStickinessConfig = expandLbListenerActionForwardConfigTargetGroupStickinessConfig(v)
	}

	return config
}

func expandMutualAuthenticationAttributes(l []interface{}) *awstypes.MutualAuthenticationAttributes {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	switch mode := tfMap[names.AttrMode].(string); mode {
	case mutualAuthenticationOff:
		return &awstypes.MutualAuthenticationAttributes{
			Mode: aws.String(mode),
		}
	case mutualAuthenticationPassthrough:
		return &awstypes.MutualAuthenticationAttributes{
			Mode:          aws.String(mode),
			TrustStoreArn: aws.String(tfMap["trust_store_arn"].(string)),
		}
	default:
		return &awstypes.MutualAuthenticationAttributes{
			Mode:                          aws.String(mode),
			TrustStoreArn:                 aws.String(tfMap["trust_store_arn"].(string)),
			IgnoreClientCertificateExpiry: aws.Bool(tfMap["ignore_client_certificate_expiry"].(bool)),
		}
	}
}

func expandLbListenerActionForwardConfigTargetGroups(l []interface{}) []awstypes.TargetGroupTuple {
	if len(l) == 0 {
		return nil
	}

	var groups []awstypes.TargetGroupTuple

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		group := awstypes.TargetGroupTuple{}
		if v, ok := tfMap[names.AttrARN]; ok && v.(string) != "" {
			group.TargetGroupArn = aws.String(v.(string))
		}
		if v, ok := tfMap[names.AttrWeight]; ok {
			group.Weight = aws.Int32(int32(v.(int)))
		}

		groups = append(groups, group)
	}

	return groups
}

func expandLbListenerActionForwardConfigTargetGroupStickinessConfig(l []interface{}) *awstypes.TargetGroupStickinessConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	tgs := &awstypes.TargetGroupStickinessConfig{}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		tgs.Enabled = aws.Bool(v)
	}

	if v, ok := tfMap[names.AttrDuration].(int); ok && v > 0 {
		tgs.DurationSeconds = aws.Int32(int32(v))
	}

	return tgs
}

func flattenLbListenerActions(d *schema.ResourceData, attrName string, apiObjects []awstypes.Action) []interface{} {
	if len(apiObjects) == 0 {
		return []interface{}{}
	}

	var tfList []interface{}

	for i, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			names.AttrType: apiObject.Type,
			"order":        aws.ToInt32(apiObject.Order),
		}

		switch apiObject.Type {
		case awstypes.ActionTypeEnumForward:
			flattenLbForwardAction(d, attrName, i, apiObject, tfMap)

		case awstypes.ActionTypeEnumRedirect:
			tfMap["redirect"] = flattenLbListenerActionRedirectConfig(apiObject.RedirectConfig)

		case awstypes.ActionTypeEnumFixedResponse:
			tfMap["fixed_response"] = flattenLbListenerActionFixedResponseConfig(apiObject.FixedResponseConfig)

		case awstypes.ActionTypeEnumAuthenticateCognito:
			tfMap["authenticate_cognito"] = flattenLbListenerActionAuthenticateCognitoConfig(apiObject.AuthenticateCognitoConfig)

		case awstypes.ActionTypeEnumAuthenticateOidc:
			// The LB API currently provides no way to read the ClientSecret
			// Instead we passthrough the configuration value into the state
			var clientSecret string
			if v, ok := d.GetOk(attrName + "." + strconv.Itoa(i) + ".authenticate_oidc.0.client_secret"); ok {
				clientSecret = v.(string)
			}

			tfMap["authenticate_oidc"] = flattenAuthenticateOIDCActionConfig(apiObject.AuthenticateOidcConfig, clientSecret)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenLbForwardAction(d *schema.ResourceData, attrName string, i int, awsAction awstypes.Action, actionMap map[string]any) {
	// On create and update, we have a Config
	// On refresh, we have a populated State
	// On import, we have nothing:
	//  - Config is known but null.
	//  - State is known, not null, but empty.
	//  - Plan is known but null.

	// During import, it's impossible to determine from AWS's response, the config, the state, or the plan
	// whether the target group ARN was defined at the top level or within a forward action. AWS returns
	// ARNs in both the default action (top-level) and in at least one forward action, regardless of
	// whether a forward is actually defined.

	// You can specify both a target group list and a top-level target group ARN only if the ARNs match
	if rawConfig := d.GetRawConfig(); rawConfig.IsKnown() && !rawConfig.IsNull() {
		if v, ok := d.GetOk("default_action.0.target_group_arn"); ok && v.(string) != "" {
			actionMap["target_group_arn"] = aws.ToString(awsAction.TargetGroupArn)
		}
		if actions := rawConfig.GetAttr(attrName); actions.IsKnown() && !actions.IsNull() {
			flattenLbForwardActionOneOf(actions, i, awsAction, actionMap)
			return
		}
	}

	if rawState := d.GetRawState(); rawState.IsKnown() && !rawState.IsNull() {
		if v, ok := d.GetOk("default_action.0.target_group_arn"); ok && v.(string) != "" {
			actionMap["target_group_arn"] = aws.ToString(awsAction.TargetGroupArn)
		}
		if actions := rawState.GetAttr(attrName); actions.IsKnown() && !actions.IsNull() && actions.LengthInt() > 0 {
			flattenLbForwardActionOneOf(actions, i, awsAction, actionMap)
			return
		}
	}

	flattenLbForwardActionBoth(awsAction, actionMap)
}

func emptyForwardConfig(f *awstypes.ForwardActionConfig) bool {
	if f == nil {
		return true
	}
	if len(f.TargetGroups) == 0 && f.TargetGroupStickinessConfig == nil {
		return true
	}
	for _, tg := range f.TargetGroups {
		if tg.TargetGroupArn != nil || tg.Weight != nil {
			return false
		}
	}
	if f.TargetGroupStickinessConfig == nil {
		return true
	}
	if f.TargetGroupStickinessConfig.Enabled == nil && f.TargetGroupStickinessConfig.DurationSeconds == nil {
		return true
	}
	return false
}

func flattenLbForwardActionOneOf(actions cty.Value, i int, awsAction awstypes.Action, actionMap map[string]any) {
	if actions.IsKnown() && !actions.IsNull() {
		index := cty.NumberIntVal(int64(i))
		if actions.HasIndex(index).True() {
			action := actions.Index(index)
			if action.IsKnown() && !action.IsNull() {
				forward := action.GetAttr("forward")
				if forward.IsKnown() && forward.LengthInt() > 0 {
					actionMap["forward"] = flattenLbListenerActionForwardConfig(awsAction.ForwardConfig)
				} else {
					actionMap["target_group_arn"] = aws.ToString(awsAction.TargetGroupArn)
				}
			}
		}
	}
}

func flattenLbForwardActionBoth(awsAction awstypes.Action, actionMap map[string]any) {
	actionMap["target_group_arn"] = aws.ToString(awsAction.TargetGroupArn)
	actionMap["forward"] = flattenLbListenerActionForwardConfig(awsAction.ForwardConfig)
}

func flattenMutualAuthenticationAttributes(description *awstypes.MutualAuthenticationAttributes) []interface{} {
	if description == nil {
		return []interface{}{}
	}

	mode := aws.ToString(description.Mode)
	if mode == mutualAuthenticationOff {
		return []interface{}{
			map[string]interface{}{
				names.AttrMode: mode,
			},
		}
	}

	m := map[string]interface{}{}
	if description.Mode != nil {
		m[names.AttrMode] = aws.ToString(description.Mode)
	}
	if description.TrustStoreArn != nil {
		m["trust_store_arn"] = aws.ToString(description.TrustStoreArn)
	}
	if description.IgnoreClientCertificateExpiry != nil {
		m["ignore_client_certificate_expiry"] = aws.ToBool(description.IgnoreClientCertificateExpiry)
	}

	return []interface{}{m}
}

func flattenAuthenticateOIDCActionConfig(apiObject *awstypes.AuthenticateOidcActionConfig, clientSecret string) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{}

	if apiObject.AuthenticationRequestExtraParams != nil {
		tfMap["authentication_request_extra_params"] = apiObject.AuthenticationRequestExtraParams
	}
	if apiObject.AuthorizationEndpoint != nil {
		tfMap["authorization_endpoint"] = aws.ToString(apiObject.AuthorizationEndpoint)
	}
	if apiObject.ClientId != nil {
		tfMap[names.AttrClientID] = aws.ToString(apiObject.ClientId)
	}
	if clientSecret != "" {
		tfMap[names.AttrClientSecret] = clientSecret
	}
	if apiObject.Issuer != nil {
		tfMap[names.AttrIssuer] = aws.ToString(apiObject.Issuer)
	}
	if string(apiObject.OnUnauthenticatedRequest) != "" {
		tfMap["on_unauthenticated_request"] = apiObject.OnUnauthenticatedRequest
	}
	if apiObject.Scope != nil {
		tfMap[names.AttrScope] = aws.ToString(apiObject.Scope)
	}
	if apiObject.SessionCookieName != nil {
		tfMap["session_cookie_name"] = aws.ToString(apiObject.SessionCookieName)
	}
	if apiObject.SessionTimeout != nil {
		tfMap["session_timeout"] = aws.ToInt64(apiObject.SessionTimeout)
	}
	if apiObject.TokenEndpoint != nil {
		tfMap["token_endpoint"] = aws.ToString(apiObject.TokenEndpoint)
	}
	if apiObject.UserInfoEndpoint != nil {
		tfMap["user_info_endpoint"] = aws.ToString(apiObject.UserInfoEndpoint)
	}

	return []interface{}{tfMap}
}

func flattenLbListenerActionAuthenticateCognitoConfig(apiObject *awstypes.AuthenticateCognitoActionConfig) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{}

	if apiObject.AuthenticationRequestExtraParams != nil {
		tfMap["authentication_request_extra_params"] = apiObject.AuthenticationRequestExtraParams
	}
	if string(apiObject.OnUnauthenticatedRequest) != "" {
		tfMap["on_unauthenticated_request"] = apiObject.OnUnauthenticatedRequest
	}
	if apiObject.Scope != nil {
		tfMap[names.AttrScope] = aws.ToString(apiObject.Scope)
	}
	if apiObject.SessionCookieName != nil {
		tfMap["session_cookie_name"] = aws.ToString(apiObject.SessionCookieName)
	}
	if apiObject.SessionTimeout != nil {
		tfMap["session_timeout"] = aws.ToInt64(apiObject.SessionTimeout)
	}
	if apiObject.UserPoolArn != nil {
		tfMap["user_pool_arn"] = aws.ToString(apiObject.UserPoolArn)
	}
	if apiObject.UserPoolClientId != nil {
		tfMap["user_pool_client_id"] = aws.ToString(apiObject.UserPoolClientId)
	}
	if apiObject.UserPoolDomain != nil {
		tfMap["user_pool_domain"] = aws.ToString(apiObject.UserPoolDomain)
	}

	return []interface{}{tfMap}
}

func flattenLbListenerActionFixedResponseConfig(config *awstypes.FixedResponseActionConfig) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}
	if config.ContentType != nil {
		m[names.AttrContentType] = aws.ToString(config.ContentType)
	}
	if config.MessageBody != nil {
		m["message_body"] = aws.ToString(config.MessageBody)
	}
	if config.StatusCode != nil {
		m[names.AttrStatusCode] = aws.ToString(config.StatusCode)
	}

	return []interface{}{m}
}

func flattenLbListenerActionForwardConfig(config *awstypes.ForwardActionConfig) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"target_group": flattenLbListenerActionForwardConfigTargetGroups(config.TargetGroups),
		"stickiness":   flattenLbListenerActionForwardConfigTargetGroupStickinessConfig(config.TargetGroupStickinessConfig),
	}

	return []interface{}{m}
}

func flattenLbListenerActionForwardConfigTargetGroups(groups []awstypes.TargetGroupTuple) []interface{} {
	if len(groups) == 0 {
		return []interface{}{}
	}

	var vGroups []interface{}

	for _, group := range groups {
		m := map[string]interface{}{}
		if group.TargetGroupArn != nil {
			m[names.AttrARN] = aws.ToString(group.TargetGroupArn)
		}
		if group.Weight != nil {
			m[names.AttrWeight] = aws.ToInt32(group.Weight)
		}

		vGroups = append(vGroups, m)
	}

	return vGroups
}

func flattenLbListenerActionForwardConfigTargetGroupStickinessConfig(config *awstypes.TargetGroupStickinessConfig) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}
	if config.Enabled != nil {
		m[names.AttrEnabled] = aws.ToBool(config.Enabled)
	}
	if config.DurationSeconds != nil {
		m[names.AttrDuration] = aws.ToInt32(config.DurationSeconds)
	}

	return []interface{}{m}
}

func flattenLbListenerActionRedirectConfig(apiObject *awstypes.RedirectActionConfig) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{}
	if apiObject.Host != nil {
		tfMap["host"] = aws.ToString(apiObject.Host)
	}
	if apiObject.Path != nil {
		tfMap[names.AttrPath] = aws.ToString(apiObject.Path)
	}
	if apiObject.Port != nil {
		tfMap[names.AttrPort] = aws.ToString(apiObject.Port)
	}
	if apiObject.Protocol != nil {
		tfMap[names.AttrProtocol] = aws.ToString(apiObject.Protocol)
	}
	if apiObject.Query != nil {
		tfMap["query"] = aws.ToString(apiObject.Query)
	}
	if string(apiObject.StatusCode) != "" {
		tfMap[names.AttrStatusCode] = apiObject.StatusCode
	}

	return []interface{}{tfMap}
}

const (
	mutualAuthenticationOff         = "off"
	mutualAuthenticationVerify      = "verify"
	mutualAuthenticationPassthrough = "passthrough"
)

func mutualAuthenticationModeEnum_Values() []string {
	return []string{
		mutualAuthenticationOff,
		mutualAuthenticationVerify,
		mutualAuthenticationPassthrough,
	}
}

const (
	alpnPolicyHTTP1Only      = "HTTP1Only"
	alpnPolicyHTTP2Only      = "HTTP2Only"
	alpnPolicyHTTP2Optional  = "HTTP2Optional"
	alpnPolicyHTTP2Preferred = "HTTP2Preferred"
	alpnPolicyNone           = "None"
)

func alpnPolicyEnum_Values() []string {
	return []string{
		alpnPolicyHTTP1Only,
		alpnPolicyHTTP2Only,
		alpnPolicyHTTP2Optional,
		alpnPolicyHTTP2Preferred,
		alpnPolicyNone,
	}
}

func validateListenerActionsCustomDiff(attrName string) schema.CustomizeDiffFunc {
	return func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
		var diags diag.Diagnostics

		configRaw := d.GetRawConfig()
		if !configRaw.IsKnown() || configRaw.IsNull() {
			return nil
		}

		actionsPath := cty.GetAttrPath(attrName)
		actions := configRaw.GetAttr(attrName)
		if actions.IsKnown() && !actions.IsNull() {
			listenerActionsPlantimeValidate(actionsPath, actions, &diags)
		}

		return sdkdiag.DiagnosticsError(diags)
	}
}

func listenerActionsPlantimeValidate(actionsPath cty.Path, actions cty.Value, diags *diag.Diagnostics) {
	it := actions.ElementIterator()
	for it.Next() {
		i, action := it.Element()
		actionPath := actionsPath.Index(i)

		listenerActionPlantimeValidate(actionPath, action, diags)
	}
}

func listenerActionPlantimeValidate(actionPath cty.Path, action cty.Value, diags *diag.Diagnostics) {
	actionType := action.GetAttr(names.AttrType)
	if !actionType.IsKnown() {
		return
	}
	if actionType.IsNull() {
		return
	}

	if action.IsKnown() && !action.IsNull() {
		tga := action.GetAttr("target_group_arn")
		f := action.GetAttr("forward")

		// If `ignore_changes` is set, even if there is no value in the configuration, the value in RawConfig is "" on refresh.

		tgKnown := tga.IsKnown() && !tga.IsNull() && tga.AsString() != ""
		fKnown := f.IsKnown() && !f.IsNull() && f.LengthInt() > 0

		var tgArn string
		if tgKnown && tga.AsString() != "" {
			tgArn = tga.AsString()
		}

		if fKnown && tgArn != "" {
			firstForward := f.Index(cty.NumberIntVal(0))
			tgSet := firstForward.GetAttr("target_group")
			if tgSet.IsKnown() && !tgSet.IsNull() && tgSet.LengthInt() > 0 {
				tgSetIt := tgSet.ElementIterator()
				for tgSetIt.Next() {
					_, ftg := tgSetIt.Element()
					ftgARN := ftg.GetAttr("arn")
					if ftgARN.IsKnown() && !ftgARN.IsNull() && ftgARN.AsString() != "" && tgArn != ftgARN.AsString() {
						*diags = append(*diags, errs.NewAttributeErrorDiagnostic(actionPath,
							"Invalid Attribute Combination",
							fmt.Sprintf("You can specify both a top-level target group ARN (%q) and, with %q, a target group list with ARNs, only if the ARNs match.",
								errs.PathString(actionPath.GetAttr("target_group_arn")),
								errs.PathString(actionPath.GetAttr("forward")),
							),
						))
					}
				}
			}
		}

		switch actionType := awstypes.ActionTypeEnum(actionType.AsString()); actionType {
		case awstypes.ActionTypeEnumForward:
			if tga.IsNull() && (f.IsNull() || f.LengthInt() == 0) {
				typePath := actionPath.GetAttr(names.AttrType)
				*diags = append(*diags, errs.NewAttributeErrorDiagnostic(typePath,
					"Invalid Attribute Combination",
					fmt.Sprintf("Either %q or %q must be specified when %q is %q.",
						errs.PathString(actionPath.GetAttr("target_group_arn")), errs.PathString(actionPath.GetAttr("forward")),
						errs.PathString(typePath),
						actionType,
					),
				))
			}

		case awstypes.ActionTypeEnumRedirect:
			if r := action.GetAttr("redirect"); r.IsNull() || r.LengthInt() == 0 {
				*diags = append(*diags, errs.NewAttributeRequiredWhenError(
					actionPath.GetAttr("redirect"),
					actionPath.GetAttr(names.AttrType),
					string(actionType),
				))
			}

		case awstypes.ActionTypeEnumFixedResponse:
			if fr := action.GetAttr("fixed_response"); fr.IsNull() || fr.LengthInt() == 0 {
				*diags = append(*diags, errs.NewAttributeRequiredWhenError(
					actionPath.GetAttr("fixed_response"),
					actionPath.GetAttr(names.AttrType),
					string(actionType),
				))
			}

		case awstypes.ActionTypeEnumAuthenticateCognito:
			if ac := action.GetAttr("authenticate_cognito"); ac.IsNull() || ac.LengthInt() == 0 {
				*diags = append(*diags, errs.NewAttributeRequiredWhenError(
					actionPath.GetAttr("authenticate_cognito"),
					actionPath.GetAttr(names.AttrType),
					string(actionType),
				))
			}

		case awstypes.ActionTypeEnumAuthenticateOidc:
			if ao := action.GetAttr("authenticate_oidc"); ao.IsNull() || ao.LengthInt() == 0 {
				*diags = append(*diags, errs.NewAttributeRequiredWhenError(
					actionPath.GetAttr("authenticate_oidc"),
					actionPath.GetAttr(names.AttrType),
					string(actionType),
				))
			}
		}
	}
}

func listenerActionRuntimeValidate(actionPath cty.Path, action map[string]any, diags *diag.Diagnostics) {
	actionType := awstypes.ActionTypeEnum(action[names.AttrType].(string))

	if v, ok := action["target_group_arn"].(string); ok && v != "" {
		if actionType != awstypes.ActionTypeEnumForward {
			*diags = append(*diags, errs.NewAttributeConflictsWhenWillBeError(
				actionPath.GetAttr("target_group_arn"),
				actionPath.GetAttr(names.AttrType),
				string(actionType),
			))
		}
	}

	if v, ok := action["forward"].([]interface{}); ok && len(v) > 0 {
		if actionType != awstypes.ActionTypeEnumForward {
			*diags = append(*diags, errs.NewAttributeConflictsWhenWillBeError(
				actionPath.GetAttr("forward"),
				actionPath.GetAttr(names.AttrType),
				string(actionType),
			))
		}
	}

	if v, ok := action["authenticate_cognito"].([]interface{}); ok && len(v) > 0 {
		if actionType != awstypes.ActionTypeEnumAuthenticateCognito {
			*diags = append(*diags, errs.NewAttributeConflictsWhenWillBeError(
				actionPath.GetAttr("authenticate_cognito"),
				actionPath.GetAttr(names.AttrType),
				string(actionType),
			))
		}
	}

	if v, ok := action["authenticate_oidc"].([]interface{}); ok && len(v) > 0 {
		if actionType != awstypes.ActionTypeEnumAuthenticateOidc {
			*diags = append(*diags, errs.NewAttributeConflictsWhenWillBeError(
				actionPath.GetAttr("authenticate_oidc"),
				actionPath.GetAttr(names.AttrType),
				string(actionType),
			))
		}
	}

	if v, ok := action["fixed_response"].([]interface{}); ok && len(v) > 0 {
		if actionType != awstypes.ActionTypeEnumFixedResponse {
			*diags = append(*diags, errs.NewAttributeConflictsWhenWillBeError(
				actionPath.GetAttr("fixed_response"),
				actionPath.GetAttr(names.AttrType),
				string(actionType),
			))
		}
	}

	if v, ok := action["redirect"].([]interface{}); ok && len(v) > 0 {
		if actionType != awstypes.ActionTypeEnumRedirect {
			*diags = append(*diags, errs.NewAttributeConflictsWhenWillBeError(
				actionPath.GetAttr("redirect"),
				actionPath.GetAttr(names.AttrType),
				string(actionType),
			))
		}
	}
}

func diffSuppressMissingForward(attrName string) schema.SchemaDiffSuppressFunc {
	return func(k, old, new string, d *schema.ResourceData) bool {
		if regexache.MustCompile(fmt.Sprintf(`^%s\.\d+\.forward\.#$`, attrName)).MatchString(k) {
			return old == "1" && new == "0"
		}
		if regexache.MustCompile(fmt.Sprintf(`^%s\.\d+\.forward\.\d+\.target_group\.#$`, attrName)).MatchString(k) {
			return old == "1" && new == "0"
		}
		return false
	}
}
