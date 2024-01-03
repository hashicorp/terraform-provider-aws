// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_ec2_traffic_mirror_filter_rule")
func ResourceTrafficMirrorFilterRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTrafficMirrorFilterRuleCreate,
		ReadWithoutTimeout:   resourceTrafficMirrorFilterRuleRead,
		UpdateWithoutTimeout: resourceTrafficMirrorFilterRuleUpdate,
		DeleteWithoutTimeout: resourceTrafficMirrorFilterRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceTrafficMirrorFilterRuleImport,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"destination_cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidCIDRNetworkAddress,
			},
			"destination_port_range": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"from_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IsPortNumberOrZero,
						},
						"to_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IsPortNumberOrZero,
						},
					},
				},
			},
			"protocol": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"rule_action": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ec2.TrafficMirrorRuleAction_Values(), false),
			},
			"rule_number": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"source_cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidCIDRNetworkAddress,
			},
			"source_port_range": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"from_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IsPortNumberOrZero,
						},
						"to_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IsPortNumberOrZero,
						},
					},
				},
			},
			"traffic_direction": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ec2.TrafficDirection_Values(), false),
			},
			"traffic_mirror_filter_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTrafficMirrorFilterRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.CreateTrafficMirrorFilterRuleInput{
		DestinationCidrBlock:  aws.String(d.Get("destination_cidr_block").(string)),
		RuleAction:            aws.String(d.Get("rule_action").(string)),
		RuleNumber:            aws.Int64(int64(d.Get("rule_number").(int))),
		SourceCidrBlock:       aws.String(d.Get("source_cidr_block").(string)),
		TrafficDirection:      aws.String(d.Get("traffic_direction").(string)),
		TrafficMirrorFilterId: aws.String(d.Get("traffic_mirror_filter_id").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("destination_port_range"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DestinationPortRange = expandTrafficMirrorPortRangeRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("protocol"); ok {
		input.Protocol = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("source_port_range"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.SourcePortRange = expandTrafficMirrorPortRangeRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	output, err := conn.CreateTrafficMirrorFilterRuleWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Traffic Mirror Filter Rule: %s", err)
	}

	d.SetId(aws.StringValue(output.TrafficMirrorFilterRule.TrafficMirrorFilterRuleId))

	return append(diags, resourceTrafficMirrorFilterRuleRead(ctx, d, meta)...)
}

func resourceTrafficMirrorFilterRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	rule, err := FindTrafficMirrorFilterRuleByTwoPartKey(ctx, conn, d.Get("traffic_mirror_filter_id").(string), d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Traffic Mirror Filter Rule %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Traffic Mirror Filter Rule (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("traffic-mirror-filter-rule/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("description", rule.Description)
	d.Set("destination_cidr_block", rule.DestinationCidrBlock)
	if rule.DestinationPortRange != nil {
		if err := d.Set("destination_port_range", []interface{}{flattenTrafficMirrorPortRange(rule.DestinationPortRange)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting destination_port_range: %s", err)
		}
	} else {
		d.Set("destination_port_range", nil)
	}
	d.Set("protocol", rule.Protocol)
	d.Set("rule_action", rule.RuleAction)
	d.Set("rule_number", rule.RuleNumber)
	d.Set("source_cidr_block", rule.SourceCidrBlock)
	if rule.SourcePortRange != nil {
		if err := d.Set("source_port_range", []interface{}{flattenTrafficMirrorPortRange(rule.SourcePortRange)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting source_port_range: %s", err)
		}
	} else {
		d.Set("source_port_range", nil)
	}
	d.Set("traffic_direction", rule.TrafficDirection)
	d.Set("traffic_mirror_filter_id", rule.TrafficMirrorFilterId)

	return diags
}

func resourceTrafficMirrorFilterRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.ModifyTrafficMirrorFilterRuleInput{
		TrafficMirrorFilterRuleId: aws.String(d.Id()),
	}

	var removeFields []string

	if d.HasChange("description") {
		if v := d.Get("description").(string); v != "" {
			input.Description = aws.String(v)
		} else {
			removeFields = append(removeFields, ec2.TrafficMirrorFilterRuleFieldDescription)
		}
	}

	if d.HasChange("destination_cidr_block") {
		input.DestinationCidrBlock = aws.String(d.Get("destination_cidr_block").(string))
	}

	if d.HasChange("destination_port_range") {
		if v, ok := d.GetOk("destination_port_range"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.DestinationPortRange = expandTrafficMirrorPortRangeRequest(v.([]interface{})[0].(map[string]interface{}))
			// Modify request that adds port range seems to fail if protocol is not set in the request.
			input.Protocol = aws.Int64(int64(d.Get("protocol").(int)))
		} else {
			removeFields = append(removeFields, ec2.TrafficMirrorFilterRuleFieldDestinationPortRange)
		}
	}

	if d.HasChange("protocol") {
		if v := d.Get("protocol").(int); v != 0 {
			input.Protocol = aws.Int64(int64(v))
		} else {
			removeFields = append(removeFields, ec2.TrafficMirrorFilterRuleFieldProtocol)
		}
	}

	if d.HasChange("rule_action") {
		input.RuleAction = aws.String(d.Get("rule_action").(string))
	}

	if d.HasChange("rule_number") {
		input.RuleNumber = aws.Int64(int64(d.Get("rule_number").(int)))
	}

	if d.HasChange("source_cidr_block") {
		input.SourceCidrBlock = aws.String(d.Get("source_cidr_block").(string))
	}

	if d.HasChange("source_port_range") {
		if v, ok := d.GetOk("source_port_range"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.SourcePortRange = expandTrafficMirrorPortRangeRequest(v.([]interface{})[0].(map[string]interface{}))
			// Modify request that adds port range seems to fail if protocol is not set in the request.
			input.Protocol = aws.Int64(int64(d.Get("protocol").(int)))
		} else {
			removeFields = append(removeFields, ec2.TrafficMirrorFilterRuleFieldSourcePortRange)
		}
	}

	if d.HasChange("traffic_direction") {
		input.TrafficDirection = aws.String(d.Get("traffic_direction").(string))
	}

	if len(removeFields) > 0 {
		input.RemoveFields = aws.StringSlice(removeFields)
	}

	_, err := conn.ModifyTrafficMirrorFilterRuleWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EC2 Traffic Mirror Filter Rule (%s): %s", d.Id(), err)
	}

	return append(diags, resourceTrafficMirrorFilterRuleRead(ctx, d, meta)...)
}

func resourceTrafficMirrorFilterRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	log.Printf("[DEBUG] Deleting EC2 Traffic Mirror Filter Rule: %s", d.Id())
	_, err := conn.DeleteTrafficMirrorFilterRuleWithContext(ctx, &ec2.DeleteTrafficMirrorFilterRuleInput{
		TrafficMirrorFilterRuleId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTrafficMirrorFilterRuleIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Traffic Mirror Filter Rule (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceTrafficMirrorFilterRuleImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.SplitN(d.Id(), ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("unexpected format (%q), expected <filter-id>:<rule-id>", d.Id())
	}

	d.Set("traffic_mirror_filter_id", parts[0])
	d.SetId(parts[1])

	return []*schema.ResourceData{d}, nil
}

func expandTrafficMirrorPortRangeRequest(tfMap map[string]interface{}) *ec2.TrafficMirrorPortRangeRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.TrafficMirrorPortRangeRequest{}

	if v, ok := tfMap["from_port"].(int); ok {
		apiObject.FromPort = aws.Int64(int64(v))
	}

	if v, ok := tfMap["to_port"].(int); ok {
		apiObject.ToPort = aws.Int64(int64(v))
	}

	return apiObject
}

func flattenTrafficMirrorPortRange(apiObject *ec2.TrafficMirrorPortRange) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.FromPort; v != nil {
		tfMap["from_port"] = aws.Int64Value(v)
	}

	if v := apiObject.ToPort; v != nil {
		tfMap["to_port"] = aws.Int64Value(v)
	}

	return tfMap
}
