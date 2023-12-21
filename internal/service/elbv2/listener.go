// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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

		CustomizeDiff: customdiff.All(
			verify.SetTagsDiff,
			func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
				var errx []error

				path := cty.GetAttrPath("default_action")

				if v, ok := d.GetOk("default_action"); ok && len(v.([]any)) > 0 {
					for i, a := range v.([]any) {
						action := a.(map[string]interface{})

						path := path.IndexInt(i)
						tga, tgaOk := action["target_group_arn"].(string)
						f, fOk := action["forward"].([]any)

						if tgaOk && tga != "" && fOk && len(f) > 0 {
							errx = append(errx, sdkdiag.DiagnosticError(errs.NewAttributeErrorDiagnostic(path,
								"Invalid Attribute Combination",
								fmt.Sprintf("Only one of %q or %q can be specified",
									"target_group_arn",
									"forward",
								),
							)))
						}

						path = path.GetAttr("type")

						switch action["type"].(string) {
						case elbv2.ActionTypeEnumForward:
							if tga == "" && len(f) == 0 {
								errx = append(errx, sdkdiag.DiagnosticError(errs.NewAttributeWarningDiagnostic(path,
									"Invalid Attribute Combination",
									fmt.Sprintf("Either %q or %q must be specified when %q is %q.",
										"target_group_arn", "forward",
										"type",
										elbv2.ActionTypeEnumForward,
									),
								)))
							}

						case elbv2.ActionTypeEnumRedirect:
							if v, ok := action["redirect"].([]any); ok && len(v) == 0 {
								errx = append(errx, sdkdiag.DiagnosticError(errs.NewAttributeWarningDiagnostic(path,
									"Invalid Attribute Combination",
									fmt.Sprintf("Attribute %q must be specified when %q is %q.",
										"redirect",
										"type",
										elbv2.ActionTypeEnumRedirect,
									),
								)))
							}

						case elbv2.ActionTypeEnumFixedResponse:
							if v, ok := action["fixed_response"].([]any); ok && len(v) == 0 {
								errx = append(errx, sdkdiag.DiagnosticError(errs.NewAttributeWarningDiagnostic(path,
									"Invalid Attribute Combination",
									fmt.Sprintf("Attribute %q must be specified when %q is %q.",
										"fixed_response",
										"type",
										elbv2.ActionTypeEnumFixedResponse,
									),
								)))
							}

						case elbv2.ActionTypeEnumAuthenticateCognito:
							if v, ok := action["authenticate_cognito"].([]any); ok && len(v) == 0 {
								errx = append(errx, sdkdiag.DiagnosticError(errs.NewAttributeWarningDiagnostic(path,
									"Invalid Attribute Combination",
									fmt.Sprintf("Attribute %q must be specified when %q is %q.",
										"authenticate_cognito",
										"type",
										elbv2.ActionTypeEnumAuthenticateCognito,
									),
								)))
							}

						case elbv2.ActionTypeEnumAuthenticateOidc:
							if v, ok := action["authenticate_oidc"].([]any); ok && len(v) == 0 {
								errx = append(errx, sdkdiag.DiagnosticError(errs.NewAttributeWarningDiagnostic(path,
									"Invalid Attribute Combination",
									fmt.Sprintf("Attribute %q must be specified when %q is %q.",
										"authenticate_oidc",
										"type",
										elbv2.ActionTypeEnumAuthenticateOidc,
									),
								)))
							}
						}
					}
				}

				return errors.Join(errx...)
			},
		),

		Schema: map[string]*schema.Schema{
			"alpn_policy": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"HTTP1Only",
					"HTTP2Only",
					"HTTP2Optional",
					"HTTP2Preferred",
					"None",
				}, true),
			},
			"arn": {
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
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authentication_request_extra_params": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"on_unauthenticated_request": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ValidateFunc: validation.StringInSlice(
											elbv2.AuthenticateCognitoActionConditionalBehaviorEnum_Values(),
											true,
										),
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
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
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
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ValidateFunc: validation.StringInSlice(
											elbv2.AuthenticateOidcActionConditionalBehaviorEnum_Values(),
											true,
										),
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
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
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
										Type:     schema.TypeString,
										Optional: true,
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
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"target_group": {
										Type:     schema.TypeSet,
										MinItems: 1,
										MaxItems: 5,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"arn": {
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
												"enabled": {
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
							ValidateFunc: validation.IntBetween(1, 50000),
						},
						"redirect": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"host": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "#{host}",
									},
									"path": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "/#{path}",
									},
									"port": {
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
										Type:     schema.TypeString,
										Optional: true,
										Default:  "#{query}",
									},
									"status_code": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice(
											elbv2.RedirectActionStatusCodeEnum_Values(),
											false,
										),
									},
								},
							},
						},
						"target_group_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice(
								elbv2.ActionTypeEnum_Values(),
								true,
							),
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

			"port": {
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
				ValidateFunc: validation.StringInSlice(elbv2.ProtocolEnum_Values(), true),
			},
			"ssl_policy": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceListenerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	lbARN := d.Get("load_balancer_arn").(string)
	input := &elbv2.CreateListenerInput{
		LoadBalancerArn: aws.String(lbARN),
		Tags:            getTagsIn(ctx),
	}

	if v, ok := d.GetOk("alpn_policy"); ok {
		input.AlpnPolicy = aws.StringSlice([]string{v.(string)})
	}

	if v, ok := d.GetOk("certificate_arn"); ok {
		input.Certificates = []*elbv2.Certificate{{
			CertificateArn: aws.String(v.(string)),
		}}
	}

	if v, ok := d.GetOk("default_action"); ok && len(v.([]interface{})) > 0 {
		var err error
		input.DefaultActions, err = expandLbListenerActions(d, v.([]interface{}))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if v, ok := d.GetOk("mutual_authentication"); ok {
		input.MutualAuthentication = expandMutualAuthenticationAttributes(v.([]interface{}))
	}

	if v, ok := d.GetOk("port"); ok {
		input.Port = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("protocol"); ok {
		input.Protocol = aws.String(v.(string))
	} else if strings.Contains(lbARN, "loadbalancer/app/") {
		// Keep previous default of HTTP for Application Load Balancers.
		input.Protocol = aws.String(elbv2.ProtocolEnumHttp)
	}

	if v, ok := d.GetOk("ssl_policy"); ok {
		input.SslPolicy = aws.String(v.(string))
	}

	output, err := retryListenerCreate(ctx, conn, input, d.Timeout(schema.TimeoutCreate))

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
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

	d.SetId(aws.StringValue(output.Listeners[0].ListenerArn))

	_, err = tfresource.RetryWhenNotFound(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return FindListenerByARN(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ELBv2 Listener (%s) create: %s", d.Id(), err)
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
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
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	listener, err := FindListenerByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ELBv2 Listener (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELBv2 Listener (%s): %s", d.Id(), err)
	}

	if listener.AlpnPolicy != nil && len(listener.AlpnPolicy) == 1 && listener.AlpnPolicy[0] != nil {
		d.Set("alpn_policy", listener.AlpnPolicy[0])
	}
	d.Set("arn", listener.ListenerArn)
	if listener.Certificates != nil && len(listener.Certificates) == 1 && listener.Certificates[0] != nil {
		d.Set("certificate_arn", listener.Certificates[0].CertificateArn)
	}
	sort.Slice(listener.DefaultActions, func(i, j int) bool {
		return aws.Int64Value(listener.DefaultActions[i].Order) < aws.Int64Value(listener.DefaultActions[j].Order)
	})
	if err := d.Set("default_action", flattenLbListenerActions(d, listener.DefaultActions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting default_action: %s", err)
	}
	d.Set("load_balancer_arn", listener.LoadBalancerArn)
	if err := d.Set("mutual_authentication", flattenMutualAuthenticationAttributes(listener.MutualAuthentication)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting mutual_authentication: %s", err)
	}
	d.Set("port", listener.Port)
	d.Set("protocol", listener.Protocol)
	d.Set("ssl_policy", listener.SslPolicy)

	return diags
}

func resourceListenerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &elbv2.ModifyListenerInput{
			ListenerArn: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("alpn_policy"); ok {
			input.AlpnPolicy = aws.StringSlice([]string{v.(string)})
		}

		if v, ok := d.GetOk("certificate_arn"); ok {
			input.Certificates = []*elbv2.Certificate{{
				CertificateArn: aws.String(v.(string)),
			}}
		}

		if d.HasChange("default_action") {
			var err error
			input.DefaultActions, err = expandLbListenerActions(d, d.Get("default_action").([]interface{}))
			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}

		if d.HasChange("mutual_authentication") {
			input.MutualAuthentication = expandMutualAuthenticationAttributes(d.Get("mutual_authentication").([]interface{}))
		}

		if v, ok := d.GetOk("port"); ok {
			input.Port = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("protocol"); ok {
			input.Protocol = aws.String(v.(string))
		}

		if v, ok := d.GetOk("ssl_policy"); ok {
			input.SslPolicy = aws.String(v.(string))
		}

		_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (interface{}, error) {
			return conn.ModifyListenerWithContext(ctx, input)
		}, elbv2.ErrCodeCertificateNotFoundException)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying ELBv2 Listener (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceListenerRead(ctx, d, meta)...)
}

func resourceListenerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	log.Printf("[INFO] Deleting ELBv2 Listener: %s", d.Id())
	_, err := conn.DeleteListenerWithContext(ctx, &elbv2.DeleteListenerInput{
		ListenerArn: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ELBv2 Listener (%s): %s", d.Id(), err)
	}

	return diags
}

func retryListenerCreate(ctx context.Context, conn *elbv2.ELBV2, input *elbv2.CreateListenerInput, timeout time.Duration) (*elbv2.CreateListenerOutput, error) {
	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, timeout, func() (interface{}, error) {
		return conn.CreateListenerWithContext(ctx, input)
	}, elbv2.ErrCodeCertificateNotFoundException)

	if err != nil {
		return nil, err
	}

	return outputRaw.(*elbv2.CreateListenerOutput), nil
}

func FindListenerByARN(ctx context.Context, conn *elbv2.ELBV2, arn string) (*elbv2.Listener, error) {
	input := &elbv2.DescribeListenersInput{
		ListenerArns: aws.StringSlice([]string{arn}),
	}
	output, err := findListener(ctx, conn, input, tfslices.PredicateTrue[*elbv2.Listener]())

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.ListenerArn) != arn {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findListener(ctx context.Context, conn *elbv2.ELBV2, input *elbv2.DescribeListenersInput, filter tfslices.Predicate[*elbv2.Listener]) (*elbv2.Listener, error) {
	output, err := findListeners(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findListeners(ctx context.Context, conn *elbv2.ELBV2, input *elbv2.DescribeListenersInput, filter tfslices.Predicate[*elbv2.Listener]) ([]*elbv2.Listener, error) {
	var output []*elbv2.Listener

	err := conn.DescribeListenersPagesWithContext(ctx, input, func(page *elbv2.DescribeListenersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Listeners {
			if v != nil && filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeListenerNotFoundException) {
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

func expandLbListenerActions(d *schema.ResourceData, l []interface{}) ([]*elbv2.Action, error) {
	if len(l) == 0 {
		return nil, nil
	}

	var actions []*elbv2.Action
	var err error

	// TODO: this breaks `listener_rule`
	actionsPath := cty.GetAttrPath("default_action")

	for i, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		actionPath := actionsPath.IndexInt(i)

		action := &elbv2.Action{
			Order: aws.Int64(int64(i + 1)),
			Type:  aws.String(tfMap["type"].(string)),
		}

		if order, ok := tfMap["order"].(int); ok && order != 0 {
			action.Order = aws.Int64(int64(order))
		}

		switch tfMap["type"].(string) {
		case elbv2.ActionTypeEnumForward:
			rawConfig := d.GetRawConfig()
			if rawConfig.IsKnown() && !rawConfig.IsNull() {
				path := actionPath.GetAttr("target_group_arn")
				var value cty.Value
				value, err = path.Apply(rawConfig)
				if err != nil {
					return nil, err
				}
				if value.IsKnown() && !value.IsNull() && value.AsString() != "" {
					action.TargetGroupArn = aws.String(value.AsString())
				}
				path = actionPath.GetAttr("forward")
				value, err = path.Apply(rawConfig)
				if err != nil {
					return nil, err
				}
				if value.IsKnown() && !value.IsNull() && value.LengthInt() > 0 {
					action.ForwardConfig = expandLbListenerActionForwardConfig(tfMap["forward"].([]any))
				}
			}

		case elbv2.ActionTypeEnumRedirect:
			v := tfMap["redirect"].([]interface{})
			action.RedirectConfig = expandLbListenerRedirectActionConfig(v)

		case elbv2.ActionTypeEnumFixedResponse:
			v := tfMap["fixed_response"].([]interface{})
			action.FixedResponseConfig = expandLbListenerFixedResponseConfig(v)

		case elbv2.ActionTypeEnumAuthenticateCognito:
			v := tfMap["authenticate_cognito"].([]interface{})
			action.AuthenticateCognitoConfig = expandLbListenerAuthenticateCognitoConfig(v)

		case elbv2.ActionTypeEnumAuthenticateOidc:
			v := tfMap["authenticate_oidc"].([]interface{})
			action.AuthenticateOidcConfig = expandAuthenticateOIDCConfig(v)
		}

		actions = append(actions, action)
	}

	return actions, err
}

func expandLbListenerAuthenticateCognitoConfig(l []interface{}) *elbv2.AuthenticateCognitoActionConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	config := &elbv2.AuthenticateCognitoActionConfig{
		AuthenticationRequestExtraParams: flex.ExpandStringMap(tfMap["authentication_request_extra_params"].(map[string]interface{})),
		UserPoolArn:                      aws.String(tfMap["user_pool_arn"].(string)),
		UserPoolClientId:                 aws.String(tfMap["user_pool_client_id"].(string)),
		UserPoolDomain:                   aws.String(tfMap["user_pool_domain"].(string)),
	}

	if v, ok := tfMap["on_unauthenticated_request"].(string); ok && v != "" {
		config.OnUnauthenticatedRequest = aws.String(v)
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

func expandAuthenticateOIDCConfig(l []interface{}) *elbv2.AuthenticateOidcActionConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	config := &elbv2.AuthenticateOidcActionConfig{
		AuthenticationRequestExtraParams: flex.ExpandStringMap(tfMap["authentication_request_extra_params"].(map[string]interface{})),
		AuthorizationEndpoint:            aws.String(tfMap["authorization_endpoint"].(string)),
		ClientId:                         aws.String(tfMap["client_id"].(string)),
		ClientSecret:                     aws.String(tfMap["client_secret"].(string)),
		Issuer:                           aws.String(tfMap["issuer"].(string)),
		TokenEndpoint:                    aws.String(tfMap["token_endpoint"].(string)),
		UserInfoEndpoint:                 aws.String(tfMap["user_info_endpoint"].(string)),
	}

	if v, ok := tfMap["on_unauthenticated_request"].(string); ok && v != "" {
		config.OnUnauthenticatedRequest = aws.String(v)
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

func expandLbListenerFixedResponseConfig(l []interface{}) *elbv2.FixedResponseActionConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	return &elbv2.FixedResponseActionConfig{
		ContentType: aws.String(tfMap["content_type"].(string)),
		MessageBody: aws.String(tfMap["message_body"].(string)),
		StatusCode:  aws.String(tfMap["status_code"].(string)),
	}
}

func expandLbListenerRedirectActionConfig(l []interface{}) *elbv2.RedirectActionConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	return &elbv2.RedirectActionConfig{
		Host:       aws.String(tfMap["host"].(string)),
		Path:       aws.String(tfMap["path"].(string)),
		Port:       aws.String(tfMap["port"].(string)),
		Protocol:   aws.String(tfMap["protocol"].(string)),
		Query:      aws.String(tfMap["query"].(string)),
		StatusCode: aws.String(tfMap["status_code"].(string)),
	}
}

func expandLbListenerActionForwardConfig(l []interface{}) *elbv2.ForwardActionConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &elbv2.ForwardActionConfig{}

	if v, ok := tfMap["target_group"].(*schema.Set); ok && v.Len() > 0 {
		config.TargetGroups = expandLbListenerActionForwardConfigTargetGroups(v.List())
	}

	if v, ok := tfMap["stickiness"].([]interface{}); ok && len(v) > 0 {
		config.TargetGroupStickinessConfig = expandLbListenerActionForwardConfigTargetGroupStickinessConfig(v)
	}

	return config
}

func expandMutualAuthenticationAttributes(l []interface{}) *elbv2.MutualAuthenticationAttributes {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	mode := tfMap["mode"].(string)
	if mode == mutualAuthenticationOff {
		return &elbv2.MutualAuthenticationAttributes{
			Mode: aws.String(mode),
		}
	}

	return &elbv2.MutualAuthenticationAttributes{
		Mode:                          aws.String(mode),
		TrustStoreArn:                 aws.String(tfMap["trust_store_arn"].(string)),
		IgnoreClientCertificateExpiry: aws.Bool(tfMap["ignore_client_certificate_expiry"].(bool)),
	}
}

func expandLbListenerActionForwardConfigTargetGroups(l []interface{}) []*elbv2.TargetGroupTuple {
	if len(l) == 0 {
		return nil
	}

	var groups []*elbv2.TargetGroupTuple

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		group := &elbv2.TargetGroupTuple{
			TargetGroupArn: aws.String(tfMap["arn"].(string)),
			Weight:         aws.Int64(int64(tfMap["weight"].(int))),
		}

		groups = append(groups, group)
	}

	return groups
}

func expandLbListenerActionForwardConfigTargetGroupStickinessConfig(l []interface{}) *elbv2.TargetGroupStickinessConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return &elbv2.TargetGroupStickinessConfig{
		Enabled:         aws.Bool(tfMap["enabled"].(bool)),
		DurationSeconds: aws.Int64(int64(tfMap["duration"].(int))),
	}
}

func flattenLbListenerActions(d *schema.ResourceData, Actions []*elbv2.Action) []interface{} {
	if len(Actions) == 0 {
		return []interface{}{}
	}

	var vActions []interface{}

	for i, action := range Actions {
		m := map[string]interface{}{
			"type":  aws.StringValue(action.Type),
			"order": aws.Int64Value(action.Order),
		}

		switch aws.StringValue(action.Type) {
		case elbv2.ActionTypeEnumForward:
			if aws.StringValue(action.TargetGroupArn) != "" {
				m["target_group_arn"] = aws.StringValue(action.TargetGroupArn)
			} else {
				m["forward"] = flattenLbListenerActionForwardConfig(action.ForwardConfig)
			}

		case elbv2.ActionTypeEnumRedirect:
			m["redirect"] = flattenLbListenerActionRedirectConfig(action.RedirectConfig)

		case elbv2.ActionTypeEnumFixedResponse:
			m["fixed_response"] = flattenLbListenerActionFixedResponseConfig(action.FixedResponseConfig)

		case elbv2.ActionTypeEnumAuthenticateCognito:
			m["authenticate_cognito"] = flattenLbListenerActionAuthenticateCognitoConfig(action.AuthenticateCognitoConfig)

		case elbv2.ActionTypeEnumAuthenticateOidc:
			// The LB API currently provides no way to read the ClientSecret
			// Instead we passthrough the configuration value into the state
			var clientSecret string
			if v, ok := d.GetOk("default_action." + strconv.Itoa(i) + ".authenticate_oidc.0.client_secret"); ok {
				clientSecret = v.(string)
			}

			m["authenticate_oidc"] = flattenAuthenticateOIDCActionConfig(action.AuthenticateOidcConfig, clientSecret)
		}

		vActions = append(vActions, m)
	}

	return vActions
}

func flattenMutualAuthenticationAttributes(description *elbv2.MutualAuthenticationAttributes) []interface{} {
	if description == nil {
		return []interface{}{}
	}

	mode := aws.StringValue(description.Mode)
	if mode == mutualAuthenticationOff {
		return []interface{}{
			map[string]interface{}{
				"mode": mode,
			},
		}
	}

	m := map[string]interface{}{
		"mode":                             aws.StringValue(description.Mode),
		"trust_store_arn":                  aws.StringValue(description.TrustStoreArn),
		"ignore_client_certificate_expiry": aws.BoolValue(description.IgnoreClientCertificateExpiry),
	}

	return []interface{}{m}
}

func flattenAuthenticateOIDCActionConfig(config *elbv2.AuthenticateOidcActionConfig, clientSecret string) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"authentication_request_extra_params": aws.StringValueMap(config.AuthenticationRequestExtraParams),
		"authorization_endpoint":              aws.StringValue(config.AuthorizationEndpoint),
		"client_id":                           aws.StringValue(config.ClientId),
		"client_secret":                       clientSecret,
		"issuer":                              aws.StringValue(config.Issuer),
		"on_unauthenticated_request":          aws.StringValue(config.OnUnauthenticatedRequest),
		"scope":                               aws.StringValue(config.Scope),
		"session_cookie_name":                 aws.StringValue(config.SessionCookieName),
		"session_timeout":                     aws.Int64Value(config.SessionTimeout),
		"token_endpoint":                      aws.StringValue(config.TokenEndpoint),
		"user_info_endpoint":                  aws.StringValue(config.UserInfoEndpoint),
	}

	return []interface{}{m}
}

func flattenLbListenerActionAuthenticateCognitoConfig(config *elbv2.AuthenticateCognitoActionConfig) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"authentication_request_extra_params": aws.StringValueMap(config.AuthenticationRequestExtraParams),
		"on_unauthenticated_request":          aws.StringValue(config.OnUnauthenticatedRequest),
		"scope":                               aws.StringValue(config.Scope),
		"session_cookie_name":                 aws.StringValue(config.SessionCookieName),
		"session_timeout":                     aws.Int64Value(config.SessionTimeout),
		"user_pool_arn":                       aws.StringValue(config.UserPoolArn),
		"user_pool_client_id":                 aws.StringValue(config.UserPoolClientId),
		"user_pool_domain":                    aws.StringValue(config.UserPoolDomain),
	}

	return []interface{}{m}
}

func flattenLbListenerActionFixedResponseConfig(config *elbv2.FixedResponseActionConfig) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"content_type": aws.StringValue(config.ContentType),
		"message_body": aws.StringValue(config.MessageBody),
		"status_code":  aws.StringValue(config.StatusCode),
	}

	return []interface{}{m}
}

func flattenLbListenerActionForwardConfig(config *elbv2.ForwardActionConfig) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"target_group": flattenLbListenerActionForwardConfigTargetGroups(config.TargetGroups),
		"stickiness":   flattenLbListenerActionForwardConfigTargetGroupStickinessConfig(config.TargetGroupStickinessConfig),
	}

	return []interface{}{m}
}

func flattenLbListenerActionForwardConfigTargetGroups(groups []*elbv2.TargetGroupTuple) []interface{} {
	if len(groups) == 0 {
		return []interface{}{}
	}

	var vGroups []interface{}

	for _, group := range groups {
		if group == nil {
			continue
		}

		m := map[string]interface{}{
			"arn":    aws.StringValue(group.TargetGroupArn),
			"weight": aws.Int64Value(group.Weight),
		}

		vGroups = append(vGroups, m)
	}

	return vGroups
}

func flattenLbListenerActionForwardConfigTargetGroupStickinessConfig(config *elbv2.TargetGroupStickinessConfig) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"enabled":  aws.BoolValue(config.Enabled),
		"duration": aws.Int64Value(config.DurationSeconds),
	}

	return []interface{}{m}
}

func flattenLbListenerActionRedirectConfig(config *elbv2.RedirectActionConfig) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"host":        aws.StringValue(config.Host),
		"path":        aws.StringValue(config.Path),
		"port":        aws.StringValue(config.Port),
		"protocol":    aws.StringValue(config.Protocol),
		"query":       aws.StringValue(config.Query),
		"status_code": aws.StringValue(config.StatusCode),
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
