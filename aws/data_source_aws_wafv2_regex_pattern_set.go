package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAwsWafv2RegexPatternSet() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsWafv2RegexPatternSetRead,

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

func dataSourceAwsWafv2RegexPatternSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn
	name := d.Get("name").(string)

	var foundRegexPatternSet *wafv2.RegexPatternSetSummary
	input := &wafv2.ListRegexPatternSetsInput{
		Scope: aws.String(d.Get("scope").(string)),
	}
	for {
		output, err := conn.ListRegexPatternSets(input)
		if err != nil {
			return fmt.Errorf("Error reading WAFV2 RegexPatternSets: %s", err)
		}

		for _, regexPatternSet := range output.RegexPatternSets {
			if aws.StringValue(regexPatternSet.Name) == name {
				foundRegexPatternSet = regexPatternSet
				break
			}
		}

		if output.NextMarker == nil || foundRegexPatternSet != nil {
			break
		}
		input.NextMarker = output.NextMarker
	}

	if foundRegexPatternSet == nil {
		return fmt.Errorf("WAFV2 RegexPatternSet not found for name: %s", name)
	}

	d.SetId(aws.StringValue(foundRegexPatternSet.Id))
	d.Set("arn", foundRegexPatternSet.ARN)
	d.Set("description", aws.StringValue(foundRegexPatternSet.Description))

	return nil
}
