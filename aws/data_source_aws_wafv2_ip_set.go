package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAwsWafv2IPSet() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsWafv2IPSetRead,

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

func dataSourceAwsWafv2IPSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn
	name := d.Get("name").(string)

	var foundIpSet *wafv2.IPSetSummary
	input := &wafv2.ListIPSetsInput{
		Scope: aws.String(d.Get("scope").(string)),
	}
	for {
		output, err := conn.ListIPSets(input)
		if err != nil {
			return fmt.Errorf("Error reading WAFV2 IPSets: %s", err)
		}

		for _, ipSet := range output.IPSets {
			if aws.StringValue(ipSet.Name) == name {
				foundIpSet = ipSet
				break
			}
		}

		if output.NextMarker == nil || foundIpSet != nil {
			break
		}
		input.NextMarker = output.NextMarker
	}

	if foundIpSet == nil {
		return fmt.Errorf("WAFV2 IPSet not found for name: %s", name)
	}

	d.SetId(aws.StringValue(foundIpSet.Id))
	d.Set("arn", aws.StringValue(foundIpSet.ARN))
	d.Set("description", aws.StringValue(foundIpSet.Description))

	return nil
}
