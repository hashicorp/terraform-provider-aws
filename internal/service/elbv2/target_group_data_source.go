// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_alb_target_group", name="Target Group")
// @SDKDataSource("aws_lb_target_group", name="Target Group")
// @Testing(tagsTest=true)
func dataSourceTargetGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTargetGroupRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"arn_suffix": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_termination": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"deregistration_delay": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrHealthCheck: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"healthy_threshold": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrInterval: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"matcher": {
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
						names.AttrTimeout: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"unhealthy_threshold": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"lambda_multi_value_headers_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"load_balancer_arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"load_balancing_algorithm_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"load_balancing_anomaly_mitigation": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"load_balancing_cross_zone_enabled": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrPort: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"preserve_client_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrProtocol: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"protocol_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"proxy_protocol_v2": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"slow_start": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"stickiness": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cookie_duration": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"cookie_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Computed: true,
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"target_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceTargetGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)
	partition := meta.(*conns.AWSClient).Partition
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tagsToMatch := tftags.New(ctx, d.Get(names.AttrTags).(map[string]interface{})).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	input := &elasticloadbalancingv2.DescribeTargetGroupsInput{}

	if v, ok := d.GetOk(names.AttrARN); ok {
		input.TargetGroupArns = []string{v.(string)}
	} else if v, ok := d.GetOk(names.AttrName); ok {
		input.Names = []string{v.(string)}
	}

	results, err := findTargetGroups(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELBv2 Target Groups: %s", err)
	}

	if len(tagsToMatch) > 0 {
		var targetGroups []awstypes.TargetGroup

		for _, targetGroup := range results {
			arn := aws.StringValue(targetGroup.TargetGroupArn)
			tags, err := listTags(ctx, conn, arn)

			if errs.IsA[*awstypes.TargetGroupNotFoundException](err) {
				continue
			}

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "listing tags for ELBv2 Target Group (%s): %s", arn, err)
			}

			if !tags.ContainsAll(tagsToMatch) {
				continue
			}

			targetGroups = append(targetGroups, targetGroup)
		}

		results = targetGroups
	}

	if len(results) != 1 {
		return sdkdiag.AppendErrorf(diags, "Search returned %d results, please revise so only one is returned", len(results))
	}

	targetGroup := results[0]
	d.SetId(aws.StringValue(targetGroup.TargetGroupArn))
	d.Set(names.AttrARN, targetGroup.TargetGroupArn)
	d.Set("arn_suffix", TargetGroupSuffixFromARN(targetGroup.TargetGroupArn))
	d.Set("load_balancer_arns", flex.FlattenStringValueSet(targetGroup.LoadBalancerArns))
	d.Set(names.AttrName, targetGroup.TargetGroupName)
	d.Set("target_type", targetGroup.TargetType)

	if err := d.Set(names.AttrHealthCheck, flattenTargetGroupHealthCheck(&targetGroup)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting health_check: %s", err)
	}
	d.Set(names.AttrName, targetGroup.TargetGroupName)
	targetType := targetGroup.TargetType
	d.Set("target_type", targetType)

	var protocol awstypes.ProtocolEnum
	if targetType != awstypes.TargetTypeEnumLambda {
		d.Set(names.AttrPort, targetGroup.Port)
		protocol = targetGroup.Protocol
		d.Set(names.AttrProtocol, protocol)
		d.Set(names.AttrVPCID, targetGroup.VpcId)
	}
	switch targetGroup.Protocol {
	case awstypes.ProtocolEnumHttp, awstypes.ProtocolEnumHttps:
		d.Set("protocol_version", targetGroup.ProtocolVersion)
	}

	attributes, err := findTargetGroupAttributesByARN(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELBv2 Target Group (%s) attributes: %s", d.Id(), err)
	}

	if err := d.Set("stickiness", []interface{}{flattenTargetGroupStickinessAttributes(attributes, protocol)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting stickiness: %s", err)
	}

	targetGroupAttributes.flatten(d, targetType, attributes)

	tags, err := listTags(ctx, conn, d.Id())

	if errs.IsUnsupportedOperationInPartitionError(partition, err) {
		log.Printf("[WARN] Unable to list tags for ELBv2 Target Group %s: %s", d.Id(), err)
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for ELBv2 Target Group (%s): %s", d.Id(), err)
	}

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
