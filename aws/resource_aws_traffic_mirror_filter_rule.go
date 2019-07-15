package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/validation"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsTrafficMirrorFilterRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsTrafficMirrorFilterRuleCreate,
		Read:   resourceAwsTrafficMirrorFilterRuleRead,
		Update: resourceAwsTrafficMirrorFilterRuleUpdate,
		Delete: resourceAwsTrafficMirrorFilterRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"traffic_mirror_filter_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination_cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateCIDRNetworkAddress,
			},
			"destination_port_range": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: map[string]*schema.Schema{
					"from_port": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"to_port": {
						Type:     schema.TypeInt,
						Optional: true,
					},
				},
			},
			"protocol": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"rule_action": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"accept",
					"reject",
				}, false),
			},
			"rule_number": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"source_cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateCIDRNetworkAddress,
			},
			"source_port_range": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: map[string]*schema.Schema{
					"from_port": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"to_port": {
						Type:     schema.TypeInt,
						Optional: true,
					},
				},
			},
			"traffic_direction": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"ingress",
					"egress",
				}, false),
			},
		},
	}
}

func resourceAwsTrafficMirrorFilterRuleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	filterId := d.Get("traffic_mirror_filter_id")

	input := &ec2.CreateTrafficMirrorFilterRuleInput{
		TrafficMirrorFilterId: aws.String(filterId.(string)),
		DestinationCidrBlock:  aws.String(d.Get("destination_cidr_block").(string)),
		SourceCidrBlock:       aws.String(d.Get("source_cidr_block").(string)),
		RuleAction:            aws.String(d.Get("rule_action").(string)),
		RuleNumber:            aws.Int64(int64(d.Get("rule_number").(int))),
		TrafficDirection:      aws.String(d.Get("traffic_direction").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("destination_port_range"); ok {
		input.SetDestinationPortRange(buildTrafficMirrorPortRangeRequest(v.([]interface{})))
	}

	if v, ok := d.GetOk("source_port_range"); ok {
		input.SetSourcePortRange(buildTrafficMirrorPortRangeRequest(v.([]interface{})))
	}

	_, err := conn.CreateTrafficMirrorFilterRule(input)
	if err != nil {
		return fmt.Errorf("Error creating traffic mirror filter rule for %v", filterId)
	}

	return resourceAwsTrafficMirrorFilterRuleRead(d, meta)
}

func resourceAwsTrafficMirrorFilterRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	ruleId := d.Id()
	filterId := d.Get("traffic_mirror_filter_id").(string)

	awsMutexKV.Lock(filterId)
	defer awsMutexKV.Unlock(filterId)

	var filterIds []*string
	filterIds = append(filterIds, &filterId)
	input := &ec2.DescribeTrafficMirrorFiltersInput{
		TrafficMirrorFilterIds: filterIds,
	}

	out, err := conn.DescribeTrafficMirrorFilters(input)
	if err != nil || len(out.TrafficMirrorFilters) == 0 {
		d.SetId("")
		return fmt.Errorf("Error finding traffic mirror filter %v", err)
	}

	filter := out.TrafficMirrorFilters[0]
	direction := d.Get("traffic_direction")
	var ruleList []*ec2.TrafficMirrorFilterRule
	var rule *ec2.TrafficMirrorFilterRule
	switch direction {
	case "egress":
		ruleList = filter.EgressFilterRules
	case "ingres":
		ruleList = filter.IngressFilterRules
	}

	for _, v := range ruleList {
		if *v.TrafficMirrorFilterRuleId == ruleId {
			rule = v
			break
		}
	}

	if rule != nil {
		d.SetId("")
		return fmt.Errorf("Filter rule id %v not found in filter %v (%v)", ruleId, filterId, direction)
	}

	return populateTrafficMirrorFilterRuleResource(d, rule)
}

func resourceAwsTrafficMirrorFilterRuleUpdate(d *schema.ResourceData, meta interface{}) error {

	return nil
}

func resourceAwsTrafficMirrorFilterRuleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	ruleId := d.Id()
	filterId := d.Get("traffic_mirror_filter_id")
	input := &ec2.DeleteTrafficMirrorFilterRuleInput{
		TrafficMirrorFilterRuleId: &ruleId,
	}

	_, err := conn.DeleteTrafficMirrorFilterRule(input)
	if err != nil {
		return fmt.Errorf("Error deleting traffic mirror filter rule %v (%v)", ruleId, filterId)
	}

	return nil
}

func buildTrafficMirrorPortRangeRequest(p []interface{}) (out *ec2.TrafficMirrorPortRangeRequest) {
	portSchema := p[0].(map[string]interface{})

	portRange := ec2.TrafficMirrorPortRangeRequest{}
	if v, ok := portSchema["from_port"]; ok {
		portRange.FromPort = aws.Int64(int64(v.(int)))
		out = &portRange
	}

	if v, ok := portSchema["to_port"]; ok {
		portRange.ToPort = aws.Int64(int64(v.(int)))
		out = &portRange
	}

	return out
}

func populateTrafficMirrorFilterRuleResource(d *schema.ResourceData, rule *ec2.TrafficMirrorFilterRule) error {
	d.SetId(*rule.TrafficMirrorFilterRuleId)
	d.Set("traffic_mirror_filter_id", rule.TrafficMirrorFilterId)
	d.Set("destination_cidr_range", rule.DestinationCidrBlock)
	d.Set("source_cidr_block", rule.SourceCidrBlock)
	d.Set("rule_action", rule.RuleAction)
	d.Set("rule_number", rule.RuleNumber)
	d.Set("traffic_direction", rule.TrafficDirection)

	if rule.Description != nil {
		d.Set("description", rule.Description)
	}

	if rule.Protocol != nil {
		d.Set("protocol", rule.Protocol)
	}

	if rule.DestinationPortRange != nil {
		d.Set("destination_port_range", buildTrafficMirrorFilterRulePortRangeSchema(rule.DestinationPortRange))
	}

	if rule.SourcePortRange != nil {
		d.Set("source_port_range", buildTrafficMirrorFilterRulePortRangeSchema(rule.SourcePortRange))
	}

	return nil
}

func buildTrafficMirrorFilterRulePortRangeSchema(portRange *ec2.TrafficMirrorPortRange) interface{} {
	var out [1]interface{}
	elem := make(map[string]interface{})
	elem["from_port"] = portRange.FromPort
	elem["to_port"] = portRange.ToPort
	out[0] = elem

	return out
}
