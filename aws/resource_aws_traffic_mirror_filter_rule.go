package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/validation"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsTrafficMirroringFilterRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsTrafficMirroringFilterRuleCreate,
		Read: resourceAwsTrafficMirroringFilterRuleRead,
		Delete: resourceAwsTrafficMirroringFilterRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type: schema.TypeString,
				Required: false,
				ForceNew: true,
			},
			"traffic_mirror_filter_id": {
				Type: schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination_cidr_block": {
				Type:schema.TypeString,
				Required:true,
				ForceNew:true,
				ValidateFunc: validateCIDRNetworkAddress,
			},
			"destination_port_range": {
				Type:schema.TypeList,
				Required: false,
				ForceNew: true,
				MaxItems: 1,
				Elem: map[string]*schema.Schema{
					"from_port": {
						Type:schema.TypeInt,
						Required:false,
						ForceNew:true,
					},
					"to_port": {
						Type:schema.TypeInt,
						Required:false,
						ForceNew:true,
					},
				},
			},
			"protocol": {
				Type: schema.TypeInt,
				Required: false,
				ForceNew: true,
			},
			"rule_action": {
				Type:schema.TypeString,
				Required:true,
				ForceNew:true,
				ValidateFunc: validation.StringInSlice([]string{
					"accept",
					"reject",
				}, false),
			},
			"rule_number": {
				Type:schema.TypeInt,
				Required:true,
				ForceNew:true,
			},
			"source_cidr_block": {
				Type:schema.TypeString,
				Required:true,
				ForceNew:true,
				ValidateFunc: validateCIDRNetworkAddress,
			},
			"source_port_range": {
				Type:schema.TypeList,
				Required:false,
				ForceNew:true,
				MaxItems: 1,
				Elem: map[string]*schema.Schema{
					"from_port": {
						Type:schema.TypeInt,
						Required:false,
						ForceNew:true,
					},
					"to_port": {
						Type:schema.TypeInt,
						Required:false,
						ForceNew:true,
					},
				},
			},
			"traffic_direction": {
				Type:schema.TypeString,
				Required:true,
				ForceNew:true,
				ValidateFunc:validation.StringInSlice([]string{
					"ingress",
					"egress",
				}, false),
			},
		},
	}
}

func resourceAwsTrafficMirroringFilterRuleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := d.(*AWSClient).ec2conn

	filterId := d.Get("traffic_mirror_filter_id")

	input := &ec2.CreateTrafficMirrorFilterRuleInput{
		TrafficMirrorFilterId: aws.String(filterId.(string)),
		DestinationCidrBlock: aws.String(d.Get("destination_cidr_block").(string)),
		SourceCidrBlock: aws.String(d.Get("source_cidr_block").(string)),
		RuleAction:aws.String(d.Get("rule_action").(string)),
		RuleNumber:aws.Int64(int64(d.Get("rule_number").(int))),
		TrafficDirection:aws.String(d.Get("traffic_direction").(string)),
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

	return resourceAwsTrafficMirroringFilterRuleRead(d, meta)
}

func resourceAwsTrafficMirroringFilterRuleRead(d *schema.ResourceData, meta interface{}) error  {
	conn := meta.(*AWSClient).ec2conn

	ruleId := d.Id()
	filterId := d.Get("traffic_mirror_filter_id").(string)

	awsMutexKV.Lock(filterId);
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

func resourceAwsTrafficMirroringFilterRuleDelete(d *schema.ResourceData, meta interface{}) error  {

	return nil
}

func buildTrafficMirrorPortRangeRequest(p []interface{}) (out *ec2.TrafficMirrorPortRangeRequest)  {
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

func populateTrafficMirrorFilterRuleResource(d *schema.ResourceData, filter *ec2.TrafficMirrorFilterRule) error {
	d.SetId(rule.S)
	return nil
}