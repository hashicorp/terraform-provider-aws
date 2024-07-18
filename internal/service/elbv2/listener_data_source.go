// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"
	"log"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_alb_listener", name="Listener")
// @SDKDataSource("aws_lb_listener", name="Listener")
// @Testing(tagsTest=true)
func dataSourceListener() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceListenerRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"alpn_policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"load_balancer_arn", names.AttrPort},
			},
			names.AttrCertificateARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDefaultAction: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authenticate_cognito": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authentication_request_extra_params": {
										Type:     schema.TypeMap,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"on_unauthenticated_request": {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrScope: {
										Type:     schema.TypeString,
										Computed: true,
									},
									"session_cookie_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"session_timeout": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"user_pool_arn": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"user_pool_client_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"user_pool_domain": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"authenticate_oidc": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authentication_request_extra_params": {
										Type:     schema.TypeMap,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"authorization_endpoint": {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrClientID: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrClientSecret: {
										Type:      schema.TypeString,
										Computed:  true,
										Sensitive: true,
									},
									names.AttrIssuer: {
										Type:     schema.TypeString,
										Computed: true,
									},
									"on_unauthenticated_request": {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrScope: {
										Type:     schema.TypeString,
										Computed: true,
									},
									"session_cookie_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"session_timeout": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"token_endpoint": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"user_info_endpoint": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"fixed_response": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrContentType: {
										Type:     schema.TypeString,
										Computed: true,
									},
									"message_body": {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrStatusCode: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"forward": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"stickiness": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrDuration: {
													Type:     schema.TypeInt,
													Computed: true,
												},
												names.AttrEnabled: {
													Type:     schema.TypeBool,
													Computed: true,
												},
											},
										},
									},
									"target_group": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrARN: {
													Type:     schema.TypeString,
													Computed: true,
												},
												names.AttrWeight: {
													Type:     schema.TypeInt,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"order": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"redirect": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"host": {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrPath: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrPort: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrProtocol: {
										Type:     schema.TypeString,
										Computed: true,
									},
									"query": {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrStatusCode: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"target_group_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"load_balancer_arn": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{names.AttrARN},
				RequiredWith:  []string{names.AttrPort},
			},
			"mutual_authentication": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrMode: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"trust_store_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ignore_client_certificate_expiry": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			names.AttrPort: {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{names.AttrARN},
				RequiredWith:  []string{"load_balancer_arn"},
			},
			names.AttrProtocol: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ssl_policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceListenerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &elasticloadbalancingv2.DescribeListenersInput{}

	if v, ok := d.GetOk(names.AttrARN); ok {
		input.ListenerArns = []string{v.(string)}
	} else if v, ok := d.GetOk("load_balancer_arn"); ok {
		input.LoadBalancerArn = aws.String(v.(string))
	}

	filter := tfslices.PredicateTrue[*awstypes.Listener]()
	if v, ok := d.GetOk(names.AttrPort); ok {
		port := v.(int)
		filter = func(v *awstypes.Listener) bool {
			return int(aws.ToInt32(v.Port)) == port
		}
	}
	listener, err := findListener(ctx, conn, input, filter)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("ELBv2 Listener", err))
	}

	d.SetId(aws.ToString(listener.ListenerArn))
	if listener.AlpnPolicy != nil && len(listener.AlpnPolicy) == 1 {
		d.Set("alpn_policy", listener.AlpnPolicy[0])
	}
	d.Set(names.AttrARN, listener.ListenerArn)
	if listener.Certificates != nil && len(listener.Certificates) == 1 {
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

	tags, err := listTags(ctx, conn, d.Id())

	if errs.IsUnsupportedOperationInPartitionError(meta.(*conns.AWSClient).Partition, err) {
		log.Printf("[WARN] Unable to list tags for ELBv2 Listener %s: %s", d.Id(), err)
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for (%s): %s", d.Id(), err)
	}

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
