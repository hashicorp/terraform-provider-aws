// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"
	"log"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKDataSource("aws_alb_listener")
// @SDKDataSource("aws_lb_listener")
func DataSourceListener() *schema.Resource {
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
			"arn": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"load_balancer_arn", "port"},
			},
			"certificate_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_action": {
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
									"scope": {
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
									"client_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"client_secret": {
										Type:      schema.TypeString,
										Computed:  true,
										Sensitive: true,
									},
									"issuer": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"on_unauthenticated_request": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"scope": {
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
									"content_type": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"message_body": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"status_code": {
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
												"duration": {
													Type:     schema.TypeInt,
													Computed: true,
												},
												"enabled": {
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
												"arn": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"weight": {
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
									"path": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"port": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"protocol": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"query": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"status_code": {
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
						"type": {
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
				ConflictsWith: []string{"arn"},
				RequiredWith:  []string{"port"},
			},
			"mutual_authentication": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mode": {
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
			"port": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"arn"},
				RequiredWith:  []string{"load_balancer_arn"},
			},
			"protocol": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ssl_policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceListenerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &elbv2.DescribeListenersInput{}

	if v, ok := d.GetOk("arn"); ok {
		input.ListenerArns = aws.StringSlice([]string{v.(string)})
	} else if v, ok := d.GetOk("load_balancer_arn"); ok {
		input.LoadBalancerArn = aws.String(v.(string))
	}

	filter := tfslices.PredicateTrue[*elbv2.Listener]()
	if v, ok := d.GetOk("port"); ok {
		port := v.(int)
		filter = func(v *elbv2.Listener) bool {
			return int(aws.Int64Value(v.Port)) == port
		}
	}
	listener, err := findListener(ctx, conn, input, filter)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("ELBv2 Listener", err))
	}

	d.SetId(aws.StringValue(listener.ListenerArn))
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

	tags, err := listTags(ctx, conn, d.Id())

	if errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		log.Printf("[WARN] Unable to list tags for ELBv2 Listener %s: %s", d.Id(), err)
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
