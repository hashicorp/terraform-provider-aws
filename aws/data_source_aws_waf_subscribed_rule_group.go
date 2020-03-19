package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsWafSubscribedRuleGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsWafSubscribedRuleGroupRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"metric_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsWafSubscribedRuleGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafconn
	name := d.Get("name").(string)
	rules := make([]*waf.SubscribedRuleGroupSummary, 0)

	input := &waf.ListSubscribedRuleGroupsInput{}
	for {
		output, err := conn.ListSubscribedRuleGroups(input)
		if err != nil {
			return fmt.Errorf("error reading WAF Subscribed Rule Group: %s", err)
		}
		for _, rule := range output.RuleGroups {
			if aws.StringValue(rule.Name) == name {
				rules = append(rules, rule)
			}
		}

		if output.NextMarker == nil {
			break
		}
		input.NextMarker = output.NextMarker
	}

	if len(rules) == 0 {
		return fmt.Errorf("WAF Subscribed Rule Group not found for the given name: %s", name)
	}

	if len(rules) > 1 {
		return fmt.Errorf("multiple WAF Subscribed Rule Group found for the given name: %s", name)
	}

	rule := rules[0]

	d.SetId(aws.StringValue(rule.RuleGroupId))
	d.Set("metric_name", rule.MetricName)

	return nil
}
