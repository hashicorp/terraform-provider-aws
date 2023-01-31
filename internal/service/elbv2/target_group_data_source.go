package elbv2

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceTargetGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTargetGroupRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
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
				Type:     schema.TypeInt,
				Computed: true,
			},
			"health_check": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"healthy_threshold": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"interval": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"matcher": {
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
						"timeout": {
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
			"load_balancing_algorithm_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"preserve_client_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"protocol": {
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
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"target_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceTargetGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tagsToMatch := tftags.New(d.Get("tags").(map[string]interface{})).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	input := &elbv2.DescribeTargetGroupsInput{}

	if v, ok := d.GetOk("arn"); ok {
		input.TargetGroupArns = aws.StringSlice([]string{v.(string)})
	} else if v, ok := d.GetOk("name"); ok {
		input.Names = aws.StringSlice([]string{v.(string)})
	}

	results, err := FindTargetGroups(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELBv2 Target Groups: %s", err)
	}

	if len(tagsToMatch) > 0 {
		var targetGroups []*elbv2.TargetGroup

		for _, targetGroup := range results {
			arn := aws.StringValue(targetGroup.TargetGroupArn)
			tags, err := ListTags(ctx, conn, arn)

			if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeTargetGroupNotFoundException) {
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

	d.Set("arn", targetGroup.TargetGroupArn)
	d.Set("arn_suffix", TargetGroupSuffixFromARN(targetGroup.TargetGroupArn))
	d.Set("name", targetGroup.TargetGroupName)
	d.Set("target_type", targetGroup.TargetType)

	if err := d.Set("health_check", flattenLbTargetGroupHealthCheck(targetGroup)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting health_check: %s", err)
	}

	if v, _ := d.Get("target_type").(string); v != elbv2.TargetTypeEnumLambda {
		d.Set("vpc_id", targetGroup.VpcId)
		d.Set("port", targetGroup.Port)
		d.Set("protocol", targetGroup.Protocol)
	}
	switch d.Get("protocol").(string) {
	case elbv2.ProtocolEnumHttp, elbv2.ProtocolEnumHttps:
		d.Set("protocol_version", targetGroup.ProtocolVersion)
	}

	attrResp, err := conn.DescribeTargetGroupAttributesWithContext(ctx, &elbv2.DescribeTargetGroupAttributesInput{
		TargetGroupArn: aws.String(d.Id()),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "retrieving Target Group Attributes: %s", err)
	}

	for _, attr := range attrResp.Attributes {
		switch aws.StringValue(attr.Key) {
		case "deregistration_delay.connection_termination.enabled":
			enabled, err := strconv.ParseBool(aws.StringValue(attr.Value))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "converting deregistration_delay.connection_termination.enabled to bool: %s", aws.StringValue(attr.Value))
			}
			d.Set("connection_termination", enabled)
		case "deregistration_delay.timeout_seconds":
			timeout, err := strconv.Atoi(aws.StringValue(attr.Value))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "converting deregistration_delay.timeout_seconds to int: %s", aws.StringValue(attr.Value))
			}
			d.Set("deregistration_delay", timeout)
		case "lambda.multi_value_headers.enabled":
			enabled, err := strconv.ParseBool(aws.StringValue(attr.Value))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "converting lambda.multi_value_headers.enabled to bool: %s", aws.StringValue(attr.Value))
			}
			d.Set("lambda_multi_value_headers_enabled", enabled)
		case "proxy_protocol_v2.enabled":
			enabled, err := strconv.ParseBool(aws.StringValue(attr.Value))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "converting proxy_protocol_v2.enabled to bool: %s", aws.StringValue(attr.Value))
			}
			d.Set("proxy_protocol_v2", enabled)
		case "slow_start.duration_seconds":
			slowStart, err := strconv.Atoi(aws.StringValue(attr.Value))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "converting slow_start.duration_seconds to int: %s", aws.StringValue(attr.Value))
			}
			d.Set("slow_start", slowStart)
		case "load_balancing.algorithm.type":
			loadBalancingAlgorithm := aws.StringValue(attr.Value)
			d.Set("load_balancing_algorithm_type", loadBalancingAlgorithm)
		case "preserve_client_ip.enabled":
			_, err := strconv.ParseBool(aws.StringValue(attr.Value))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "converting preserve_client_ip.enabled to bool: %s", aws.StringValue(attr.Value))
			}
			d.Set("preserve_client_ip", attr.Value)
		}
	}

	stickinessAttr, err := flattenTargetGroupStickiness(attrResp.Attributes)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "flattening stickiness: %s", err)
	}

	if err := d.Set("stickiness", stickinessAttr); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting stickiness: %s", err)
	}

	tags, err := ListTags(ctx, conn, d.Id())

	if verify.ErrorISOUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] Unable to list tags for ELBv2 Target Group %s: %s", d.Id(), err)
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for ELBv2 Target Group (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
