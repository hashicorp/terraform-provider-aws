// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_alb", name="Load Balancer")
// @SDKDataSource("aws_lb", name="Load Balancer")
// @Testing(tagsTest=true)
func dataSourceLoadBalancer() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLoadBalancerRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"access_logs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrBucket: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Computed: true,
						},
						names.AttrPrefix: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrARN: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
			},
			"arn_suffix": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_keep_alive": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"connection_logs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrBucket: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Computed: true,
						},
						names.AttrPrefix: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"customer_owned_ipv4_pool": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"desync_mitigation_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDNSName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_record_client_routing_policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"drop_invalid_header_fields": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"enable_cross_zone_load_balancing": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"enable_deletion_protection": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"enable_http2": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"enable_tls_version_and_cipher_suite_headers": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"enable_waf_fail_open": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"enable_xff_client_port": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"enforce_security_group_inbound_rules_on_private_link_traffic": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"idle_timeout": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"internal": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrIPAddressType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"load_balancer_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"preserve_host_header": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrSecurityGroups: {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"subnet_mapping": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allocation_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ipv6_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"outpost_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"private_ipv4_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrSubnetID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrSubnets: {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"xff_header_processing_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceLoadBalancerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)
	partition := meta.(*conns.AWSClient).Partition
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	tagsToMatch := tftags.New(ctx, d.Get(names.AttrTags).(map[string]interface{})).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	input := &elasticloadbalancingv2.DescribeLoadBalancersInput{}

	if v, ok := d.GetOk(names.AttrARN); ok {
		input.LoadBalancerArns = []string{v.(string)}
	} else if v, ok := d.GetOk(names.AttrName); ok {
		input.Names = []string{v.(string)}
	}

	results, err := findLoadBalancers(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELBv2 Load Balancers: %s", err)
	}

	if len(tagsToMatch) > 0 {
		var loadBalancers []awstypes.LoadBalancer

		for _, loadBalancer := range results {
			arn := aws.ToString(loadBalancer.LoadBalancerArn)
			tags, err := listTags(ctx, conn, arn)

			if errs.IsA[*awstypes.LoadBalancerNotFoundException](err) {
				continue
			}

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "listing tags for (%s): %s", arn, err)
			}

			if !tags.ContainsAll(tagsToMatch) {
				continue
			}

			loadBalancers = append(loadBalancers, loadBalancer)
		}

		results = loadBalancers
	}

	if len(results) != 1 {
		return sdkdiag.AppendErrorf(diags, "Search returned %d results, please revise so only one is returned", len(results))
	}

	lb := results[0]
	d.SetId(aws.ToString(lb.LoadBalancerArn))
	d.Set(names.AttrARN, lb.LoadBalancerArn)
	d.Set("arn_suffix", suffixFromARN(lb.LoadBalancerArn))
	d.Set("customer_owned_ipv4_pool", lb.CustomerOwnedIpv4Pool)
	d.Set(names.AttrDNSName, lb.DNSName)
	d.Set("enforce_security_group_inbound_rules_on_private_link_traffic", lb.EnforceSecurityGroupInboundRulesOnPrivateLinkTraffic)
	d.Set(names.AttrIPAddressType, lb.IpAddressType)
	d.Set(names.AttrName, lb.LoadBalancerName)
	d.Set("internal", string(lb.Scheme) == "internal")
	d.Set("load_balancer_type", lb.Type)
	d.Set(names.AttrSecurityGroups, lb.SecurityGroups)
	if err := d.Set("subnet_mapping", flattenSubnetMappingsFromAvailabilityZones(lb.AvailabilityZones)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting subnet_mapping: %s", err)
	}
	if err := d.Set(names.AttrSubnets, flattenSubnetsFromAvailabilityZones(lb.AvailabilityZones)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting subnets: %s", err)
	}
	d.Set(names.AttrVPCID, lb.VpcId)
	d.Set("zone_id", lb.CanonicalHostedZoneId)

	attributes, err := findLoadBalancerAttributesByARN(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELBv2 Load Balancer (%s) attributes: %s", d.Id(), err)
	}

	if err := d.Set("access_logs", []interface{}{flattenLoadBalancerAccessLogsAttributes(attributes)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting access_logs: %s", err)
	}

	if err := d.Set("connection_logs", []interface{}{flattenLoadBalancerConnectionLogsAttributes(attributes)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting connection_logs: %s", err)
	}

	loadBalancerAttributes.flatten(d, attributes)

	tags, err := listTags(ctx, conn, d.Id())

	if errs.IsUnsupportedOperationInPartitionError(partition, err) {
		log.Printf("[WARN] Unable to list tags for ELBv2 Load Balancer (%s): %s", d.Id(), err)
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
