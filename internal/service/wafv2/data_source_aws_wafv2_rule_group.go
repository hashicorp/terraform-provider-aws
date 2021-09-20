package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceRuleGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceRuleGroupRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"scope": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					wafv2.ScopeCloudfront,
					wafv2.ScopeRegional,
				}, false),
			},
		},
	}
}

func dataSourceRuleGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFV2Conn
	name := d.Get("name").(string)

	var foundRuleGroup *wafv2.RuleGroupSummary
	input := &wafv2.ListRuleGroupsInput{
		Scope: aws.String(d.Get("scope").(string)),
		Limit: aws.Int64(100),
	}

	for {
		resp, err := conn.ListRuleGroups(input)
		if err != nil {
			return fmt.Errorf("Error reading WAFv2 RuleGroups: %w", err)
		}

		if resp == nil || resp.RuleGroups == nil {
			return fmt.Errorf("Error reading WAFv2 RuleGroups")
		}

		for _, ruleGroup := range resp.RuleGroups {
			if aws.StringValue(ruleGroup.Name) == name {
				foundRuleGroup = ruleGroup
				break
			}
		}

		if resp.NextMarker == nil || foundRuleGroup != nil {
			break
		}
		input.NextMarker = resp.NextMarker
	}

	if foundRuleGroup == nil {
		return fmt.Errorf("WAFv2 RuleGroup not found for name: %s", name)
	}

	d.SetId(aws.StringValue(foundRuleGroup.Id))
	d.Set("arn", foundRuleGroup.ARN)
	d.Set("description", foundRuleGroup.Description)

	return nil
}
