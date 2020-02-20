package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsEc2TrafficMirrorFilterRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2TrafficMirrorFilterRuleCreate,
		Read:   resourceAwsEc2TrafficMirrorFilterRuleRead,
		Update: resourceAwsEc2TrafficMirrorFilterRuleUpdate,
		Delete: resourceAwsEc2TrafficMirrorFilterRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"traffic_mirror_filter_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination_cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateCIDRNetworkAddress,
			},
			"destination_port_range": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
			},
			"protocol": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"rule_action": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"accept",
					"reject",
				}, false),
			},
			"rule_number": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"source_cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateCIDRNetworkAddress,
			},
			"source_port_range": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
			},
			"traffic_direction": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"ingress",
					"egress",
				}, false),
			},
		},
	}
}

func resourceAwsEc2TrafficMirrorFilterRuleCreate(d *schema.ResourceData, meta interface{}) error {
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

	if v, ok := d.GetOk("protocol"); ok {
		input.Protocol = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("destination_port_range"); ok {
		input.DestinationPortRange = buildTrafficMirrorPortRangeRequest(v.([]interface{}))
	}

	if v, ok := d.GetOk("source_port_range"); ok {
		input.SetSourcePortRange(buildTrafficMirrorPortRangeRequest(v.([]interface{})))
	}

	out, err := conn.CreateTrafficMirrorFilterRule(input)
	if err != nil {
		return fmt.Errorf("Error creating traffic mirror filter rule for %v", filterId)
	}

	d.SetId(*out.TrafficMirrorFilterRule.TrafficMirrorFilterRuleId)
	return resourceAwsEc2TrafficMirrorFilterRuleRead(d, meta)
}

func resourceAwsEc2TrafficMirrorFilterRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ruleId := d.Id()

	var rule *ec2.TrafficMirrorFilterRule
	filterId, filterIdSet := d.GetOk("traffic_mirror_filter_id")
	input := &ec2.DescribeTrafficMirrorFiltersInput{}
	if filterIdSet {
		input.TrafficMirrorFilterIds = []*string{aws.String(filterId.(string))}
		awsMutexKV.Lock(filterId.(string))
		defer awsMutexKV.Unlock(filterId.(string))
	} else {
		awsMutexKV.Lock("DescribeTrafficMirrorFilters")
		defer awsMutexKV.Unlock("DescribeTrafficMirrorFilters")
	}

	found := false
	for !found {
		out, err := conn.DescribeTrafficMirrorFilters(input)
		if err != nil {
			return fmt.Errorf("Error listing traffic mirror filters in the account")
		}

		if 0 == len(out.TrafficMirrorFilters) {
			return fmt.Errorf("No traffir mirror filters found")
		}

		_, rule, err = findRuleInFilters(ruleId, out.TrafficMirrorFilters)
		if nil == err {
			found = true
			break
		} else {
			if out.NextToken == nil {
				break
			} else {
				input.NextToken = out.NextToken
			}
		}
	}

	if !found {
		d.SetId("")
		return fmt.Errorf("Rule %s not found", ruleId)
	}

	return populateTrafficMirrorFilterRuleResource(d, rule)
}

func findRuleInFilters(ruleId string, filters []*ec2.TrafficMirrorFilter) (filter *ec2.TrafficMirrorFilter, rule *ec2.TrafficMirrorFilterRule, err error) {
	log.Printf("[DEBUG] searching %s in %d filters", ruleId, len(filters))
	err = nil
	found := false
	for _, v := range filters {
		log.Printf("[DEBUG]: searching filter %s, ingress rule count = %d, egress rule count = %d", *v.TrafficMirrorFilterId, len(v.IngressFilterRules), len(v.EgressFilterRules))
		for _, r := range v.IngressFilterRules {
			if *r.TrafficMirrorFilterRuleId == ruleId {
				rule = r
				filter = v
				found = true
				break
			}
		}
		for _, r := range v.EgressFilterRules {
			if *r.TrafficMirrorFilterRuleId == ruleId {
				rule = r
				filter = v
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		err = fmt.Errorf("Rule %s not found", ruleId)
	}

	log.Printf("[DEBUG]: Found %s in %s %s", ruleId, *filter.TrafficMirrorFilterId, *rule.TrafficDirection)
	return filter, rule, err
}

func resourceAwsEc2TrafficMirrorFilterRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ruleId := d.Id()

	input := &ec2.ModifyTrafficMirrorFilterRuleInput{
		TrafficMirrorFilterRuleId: &ruleId,
	}

	var removeFields []*string
	if d.HasChange("protocol") {
		_, n := d.GetChange("protocol")
		if n == "0" {
			removeFields = append(removeFields, aws.String("protocol"))
		} else {
			input.Protocol = aws.Int64(int64(d.Get("protocol").(int)))
		}
	}

	if d.HasChange("description") {
		_, n := d.GetChange("description")
		if n != "" {
			input.Description = aws.String(n.(string))
		} else {
			removeFields = append(removeFields, aws.String("description"))
		}
	}

	if d.HasChange("destination_cidr_block") {
		input.DestinationCidrBlock = aws.String(d.Get("destination_cidr_block").(string))
	}

	if d.HasChange("source_cidr_block") {
		input.SourceCidrBlock = aws.String(d.Get("source_cidr_block").(string))
	}

	if d.HasChange("destination_port_range") {
		_, v := d.GetChange("destination_port_range")
		n := v.([]interface{})
		if 0 == len(n) {
			removeFields = append(removeFields, aws.String("destination-port-range"))
		} else {
			//Modify request that adds port range seems to fail if protocol is not set in the request
			input.Protocol = aws.Int64(int64(d.Get("protocol").(int)))
			input.SetDestinationPortRange(buildTrafficMirrorPortRangeRequest(n))
		}
	}

	if d.HasChange("source_port_range") {
		_, v := d.GetChange("source_port_range")
		n := v.([]interface{})
		if 0 == len(n) {
			removeFields = append(removeFields, aws.String("source-port-range"))
		} else {
			//Modify request that adds port range seems to fail if protocol is not set in the request
			input.Protocol = aws.Int64(int64(d.Get("protocol").(int)))
			input.SetSourcePortRange(buildTrafficMirrorPortRangeRequest(n))
		}
	}

	if d.HasChange("rule_action") {
		input.RuleAction = aws.String(d.Get("rule_action").(string))
	}

	if d.HasChange("rule_number") {
		input.RuleNumber = aws.Int64(int64(d.Get("rule_action").(int)))
	}

	if d.HasChange("traffic_direction") {
		input.TrafficDirection = aws.String(d.Get("traffic_direction").(string))
	}

	if len(removeFields) > 0 {
		input.SetRemoveFields(removeFields)
	}

	_, err := conn.ModifyTrafficMirrorFilterRule(input)
	if err != nil {
		return fmt.Errorf("Error modifying rule %v", ruleId)
	}

	return resourceAwsEc2TrafficMirrorFilterRuleRead(d, meta)
}

func resourceAwsEc2TrafficMirrorFilterRuleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	ruleId := d.Id()
	input := &ec2.DeleteTrafficMirrorFilterRuleInput{
		TrafficMirrorFilterRuleId: &ruleId,
	}

	_, err := conn.DeleteTrafficMirrorFilterRule(input)
	if err != nil {
		return fmt.Errorf("Error deleting traffic mirror filter rule %v", ruleId)
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
	d.Set("destination_cidr_block", rule.DestinationCidrBlock)
	d.Set("source_cidr_block", rule.SourceCidrBlock)
	d.Set("rule_action", rule.RuleAction)
	d.Set("rule_number", rule.RuleNumber)
	d.Set("traffic_direction", rule.TrafficDirection)
	d.Set("description", rule.Description)
	d.Set("protocol", rule.Protocol)
	d.Set("destination_port_range", buildTrafficMirrorFilterRulePortRangeSchema(rule.DestinationPortRange))
	d.Set("source_port_range", buildTrafficMirrorFilterRulePortRangeSchema(rule.SourcePortRange))

	return nil
}

func buildTrafficMirrorFilterRulePortRangeSchema(portRange *ec2.TrafficMirrorPortRange) interface{} {
	if nil == portRange {
		return nil
	}

	var out [1]interface{}
	elem := make(map[string]interface{})
	elem["from_port"] = portRange.FromPort
	elem["to_port"] = portRange.ToPort
	out[0] = elem

	return out
}
