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
			"addresses": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_address_version": {
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
			return fmt.Errorf("Error reading WAFv2 IPSets: %s", err)
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
		return fmt.Errorf("WAFv2 IPSet not found for name: %s", name)
	}

	resp, err := conn.GetIPSet(&wafv2.GetIPSetInput{
		Id:    aws.String(*foundIpSet.Id),
		Name:  aws.String(*foundIpSet.Name),
		Scope: aws.String(d.Get("scope").(string)),
	})

	if err != nil {
		return fmt.Errorf("WAFv2 IPSet not found for name: %s", name)
	}

	d.SetId(aws.StringValue(resp.IPSet.Id))
	d.Set("addresses", flattenStringList(resp.IPSet.Addresses))
	d.Set("arn", aws.StringValue(resp.IPSet.ARN))
	d.Set("description", aws.StringValue(resp.IPSet.Description))
	d.Set("ip_address_version", aws.StringValue(resp.IPSet.IPAddressVersion))

	return nil
}
