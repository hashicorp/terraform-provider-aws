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
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
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
func ResourceListener() *schema.Resource {
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
			"certificate_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"default_action": {
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
									"scope": {
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
												"duration": {
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
						"mode": {
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
			"protocol": {
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
			validateListenerActionsCustomDiff("default_action"),
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
		Tags:            getTagsInV2(ctx),
	}

	if v, ok := d.GetOk("alpn_policy"); ok {
		input.AlpnPolicy = []string{v.(string)}
	}

	if v, ok := d.GetOk("certificate_arn"); ok {
		input.Certificates = []awstypes.Certificate{{
			CertificateArn: aws.String(v.(string)),
		}}
	}

	if v, ok := d.GetOk("default_action"); ok && len(v.([]interface{})) > 0 {
		input.DefaultActions = expandLbListenerActions(cty.GetAttrPath("default_action"), v.([]any), &diags)
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

	if v, ok := d.GetOk("protocol"); ok {
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
	if tags := getTagsInV2(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTagsV2(ctx, conn, d.Id(), tags)

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

	if listener.AlpnPolicy != nil && len(listener.AlpnPolicy) == 1 {
		d.Set("alpn_policy", listener.AlpnPolicy[0])
	}
	d.Set(names.AttrARN, listener.ListenerArn)
	if listener.Certificates != nil && len(listener.Certificates) == 1 {
		d.Set("certificate_arn", listener.Certificates[0].CertificateArn)
	}
	sort.Slice(listener.DefaultActions, func(i, j int) bool {
		return aws.ToInt32(listener.DefaultActions[i].Order) < aws.ToInt32(listener.DefaultActions[j].Order)
	})
	if err := d.Set("default_action", flattenLbListenerActions(d, "default_action", listener.DefaultActions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting default_action: %s", err)
	}
	d.Set("load_balancer_arn", listener.LoadBalancerArn)
	if err := d.Set("mutual_authentication", flattenMutualAuthenticationAttributes(listener.MutualAuthentication)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting mutual_authentication: %s", err)
	}
	d.Set(names.AttrPort, listener.Port)
	d.Set("protocol", listener.Protocol)
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

		if v, ok := d.GetOk("certificate_arn"); ok {
			input.Certificates = []awstypes.Certificate{{
				CertificateArn: aws.String(v.(string)),
			}}
		}

		if d.HasChange("default_action") {
			input.DefaultActions = expandLbListenerActions(cty.GetAttrPath("default_action"), d.Get("default_action").([]any), &diags)
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

		if v, ok := d.GetOk("protocol"); ok {
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

	paginator := elasticloadbalancingv2.NewDescribeListenersPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
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
		if v, ok := tfMap["target_group_arn"].(string); ok && v != "" {
			action.TargetGroupArn = aws.String(v)
		} else if v, ok := tfMap["forward"].([]interface{}); ok {
			action.ForwardConfig = expandLbListenerActionForwardConfig(v)
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

	if v, ok := tfMap["scope"].(string); ok && v != "" {
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
		ClientId:                         aws.String(tfMap["client_id"].(string)),
		ClientSecret:                     aws.String(tfMap["client_secret"].(string)),
		Issuer:                           aws.String(tfMap["issuer"].(string)),
		TokenEndpoint:                    aws.String(tfMap["token_endpoint"].(string)),
		UserInfoEndpoint:                 aws.String(tfMap["user_info_endpoint"].(string)),
	}

	if v, ok := tfMap["on_unauthenticated_request"].(string); ok && v != "" {
		config.OnUnauthenticatedRequest = awstypes.AuthenticateOidcActionConditionalBehaviorEnum(v)
	}

	if v, ok := tfMap["scope"].(string); ok && v != "" {
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

	return &awstypes.FixedResponseActionConfig{
		ContentType: aws.String(tfMap["content_type"].(string)),
		MessageBody: aws.String(tfMap["message_body"].(string)),
		StatusCode:  aws.String(tfMap["status_code"].(string)),
	}
}

func expandLbListenerRedirectActionConfig(l []interface{}) *awstypes.RedirectActionConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	return &awstypes.RedirectActionConfig{
		Host:       aws.String(tfMap["host"].(string)),
		Path:       aws.String(tfMap["path"].(string)),
		Port:       aws.String(tfMap[names.AttrPort].(string)),
		Protocol:   aws.String(tfMap["protocol"].(string)),
		Query:      aws.String(tfMap["query"].(string)),
		StatusCode: awstypes.RedirectActionStatusCodeEnum(tfMap["status_code"].(string)),
	}
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

	switch mode := tfMap["mode"].(string); mode {
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

		group := awstypes.TargetGroupTuple{
			TargetGroupArn: aws.String(tfMap[names.AttrARN].(string)),
			Weight:         aws.Int32(int32(tfMap["weight"].(int))),
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

	// The Plugin SDK stores a `nil` returned by the API as a `0` in the state. This is a invalid value.
	var duration *int32
	if v := tfMap["duration"].(int); v > 0 {
		duration = aws.Int32(int32(v))
	}

	return &awstypes.TargetGroupStickinessConfig{
		Enabled:         aws.Bool(tfMap[names.AttrEnabled].(bool)),
		DurationSeconds: duration,
	}
}

func flattenLbListenerActions(d *schema.ResourceData, attrName string, actions []awstypes.Action) []interface{} {
	if len(actions) == 0 {
		return []interface{}{}
	}

	var vActions []interface{}

	for i, action := range actions {
		m := map[string]interface{}{
			names.AttrType: string(action.Type),
			"order":        aws.ToInt32(action.Order),
		}

		switch action.Type {
		case awstypes.ActionTypeEnumForward:
			flattenLbForwardAction(d, attrName, i, action, m)

		case awstypes.ActionTypeEnumRedirect:
			m["redirect"] = flattenLbListenerActionRedirectConfig(action.RedirectConfig)

		case awstypes.ActionTypeEnumFixedResponse:
			m["fixed_response"] = flattenLbListenerActionFixedResponseConfig(action.FixedResponseConfig)

		case awstypes.ActionTypeEnumAuthenticateCognito:
			m["authenticate_cognito"] = flattenLbListenerActionAuthenticateCognitoConfig(action.AuthenticateCognitoConfig)

		case awstypes.ActionTypeEnumAuthenticateOidc:
			// The LB API currently provides no way to read the ClientSecret
			// Instead we passthrough the configuration value into the state
			var clientSecret string
			if v, ok := d.GetOk(attrName + "." + strconv.Itoa(i) + ".authenticate_oidc.0.client_secret"); ok {
				clientSecret = v.(string)
			}

			m["authenticate_oidc"] = flattenAuthenticateOIDCActionConfig(action.AuthenticateOidcConfig, clientSecret)
		}

		vActions = append(vActions, m)
	}

	return vActions
}

func flattenLbForwardAction(d *schema.ResourceData, attrName string, i int, awsAction awstypes.Action, actionMap map[string]any) {
	// On create and update, we have a Config
	// On refresh, we have a populated State
	// On import, we have an empty State and empty Config

	if rawConfig := d.GetRawConfig(); rawConfig.IsKnown() && !rawConfig.IsNull() {
		actions := rawConfig.GetAttr(attrName)
		flattenLbForwardActionOneOf(actions, i, awsAction, actionMap)
		return
	}

	rawState := d.GetRawState()
	defaultActions := rawState.GetAttr(attrName)

	if defaultActions.LengthInt() > 0 {
		flattenLbForwardActionOneOf(defaultActions, i, awsAction, actionMap)
		return
	}

	flattenLbForwardActionBoth(awsAction, actionMap)
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
				"mode": mode,
			},
		}
	}

	m := map[string]interface{}{
		"mode":                             aws.ToString(description.Mode),
		"trust_store_arn":                  aws.ToString(description.TrustStoreArn),
		"ignore_client_certificate_expiry": aws.ToBool(description.IgnoreClientCertificateExpiry),
	}

	return []interface{}{m}
}

func flattenAuthenticateOIDCActionConfig(config *awstypes.AuthenticateOidcActionConfig, clientSecret string) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"authentication_request_extra_params": config.AuthenticationRequestExtraParams,
		"authorization_endpoint":              aws.ToString(config.AuthorizationEndpoint),
		"client_id":                           aws.ToString(config.ClientId),
		"client_secret":                       clientSecret,
		"issuer":                              aws.ToString(config.Issuer),
		"on_unauthenticated_request":          string(config.OnUnauthenticatedRequest),
		"scope":                               aws.ToString(config.Scope),
		"session_cookie_name":                 aws.ToString(config.SessionCookieName),
		"session_timeout":                     aws.ToInt64(config.SessionTimeout),
		"token_endpoint":                      aws.ToString(config.TokenEndpoint),
		"user_info_endpoint":                  aws.ToString(config.UserInfoEndpoint),
	}

	return []interface{}{m}
}

func flattenLbListenerActionAuthenticateCognitoConfig(config *awstypes.AuthenticateCognitoActionConfig) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"authentication_request_extra_params": config.AuthenticationRequestExtraParams,
		"on_unauthenticated_request":          string(config.OnUnauthenticatedRequest),
		"scope":                               aws.ToString(config.Scope),
		"session_cookie_name":                 aws.ToString(config.SessionCookieName),
		"session_timeout":                     aws.ToInt64(config.SessionTimeout),
		"user_pool_arn":                       aws.ToString(config.UserPoolArn),
		"user_pool_client_id":                 aws.ToString(config.UserPoolClientId),
		"user_pool_domain":                    aws.ToString(config.UserPoolDomain),
	}

	return []interface{}{m}
}

func flattenLbListenerActionFixedResponseConfig(config *awstypes.FixedResponseActionConfig) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"content_type": aws.ToString(config.ContentType),
		"message_body": aws.ToString(config.MessageBody),
		"status_code":  aws.ToString(config.StatusCode),
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
		m := map[string]interface{}{
			names.AttrARN: aws.ToString(group.TargetGroupArn),
			"weight":      aws.ToInt32(group.Weight),
		}

		vGroups = append(vGroups, m)
	}

	return vGroups
}

func flattenLbListenerActionForwardConfigTargetGroupStickinessConfig(config *awstypes.TargetGroupStickinessConfig) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrEnabled: aws.ToBool(config.Enabled),
		"duration":        aws.ToInt32(config.DurationSeconds),
	}

	return []interface{}{m}
}

func flattenLbListenerActionRedirectConfig(config *awstypes.RedirectActionConfig) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"host":         aws.ToString(config.Host),
		"path":         aws.ToString(config.Path),
		names.AttrPort: aws.ToString(config.Port),
		"protocol":     aws.ToString(config.Protocol),
		"query":        aws.ToString(config.Query),
		"status_code":  string(config.StatusCode),
	}

	return []interface{}{m}
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
		if (tga.IsKnown() && !tga.IsNull() && tga.AsString() != "") && (f.IsKnown() && !f.IsNull() && f.LengthInt() > 0) {
			*diags = append(*diags, errs.NewAttributeErrorDiagnostic(actionPath,
				"Invalid Attribute Combination",
				fmt.Sprintf("Only one of %q or %q can be specified.",
					errs.PathString(actionPath.GetAttr("target_group_arn")),
					errs.PathString(actionPath.GetAttr("forward")),
				),
			))
		}

		switch awstypes.ActionTypeEnum(actionType.AsString()) {
		case awstypes.ActionTypeEnumForward:
			if tga.IsNull() && (f.IsNull() || f.LengthInt() == 0) {
				typePath := actionPath.GetAttr(names.AttrType)
				*diags = append(*diags, errs.NewAttributeErrorDiagnostic(typePath,
					"Invalid Attribute Combination",
					fmt.Sprintf("Either %q or %q must be specified when %q is %q.",
						errs.PathString(actionPath.GetAttr("target_group_arn")), errs.PathString(actionPath.GetAttr("forward")),
						errs.PathString(typePath),
						awstypes.ActionTypeEnumForward,
					),
				))
			}

		case awstypes.ActionTypeEnumRedirect:
			if r := action.GetAttr("redirect"); r.IsNull() || r.LengthInt() == 0 {
				*diags = append(*diags, errs.NewAttributeRequiredWhenError(
					actionPath.GetAttr("redirect"),
					actionPath.GetAttr(names.AttrType),
					string(awstypes.ActionTypeEnumRedirect),
				))
			}

		case awstypes.ActionTypeEnumFixedResponse:
			if fr := action.GetAttr("fixed_response"); fr.IsNull() || fr.LengthInt() == 0 {
				*diags = append(*diags, errs.NewAttributeRequiredWhenError(
					actionPath.GetAttr("fixed_response"),
					actionPath.GetAttr(names.AttrType),
					string(awstypes.ActionTypeEnumFixedResponse),
				))
			}

		case awstypes.ActionTypeEnumAuthenticateCognito:
			if ac := action.GetAttr("authenticate_cognito"); ac.IsNull() || ac.LengthInt() == 0 {
				*diags = append(*diags, errs.NewAttributeRequiredWhenError(
					actionPath.GetAttr("authenticate_cognito"),
					actionPath.GetAttr(names.AttrType),
					string(awstypes.ActionTypeEnumAuthenticateCognito),
				))
			}

		case awstypes.ActionTypeEnumAuthenticateOidc:
			if ao := action.GetAttr("authenticate_oidc"); ao.IsNull() || ao.LengthInt() == 0 {
				*diags = append(*diags, errs.NewAttributeRequiredWhenError(
					actionPath.GetAttr("authenticate_oidc"),
					actionPath.GetAttr(names.AttrType),
					string(awstypes.ActionTypeEnumAuthenticateOidc),
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
