package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsLbSSLPolicy() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLbSSLPolicyRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsLbSSLPolicyRead(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbv2conn

	input := &elbv2.DescribeSSLPoliciesInput{}

	if v, ok := d.GetOk("name"); ok {
		input.Names = []*string{aws.String(v.(string))}
	}

	resp, err := elbconn.DescribeSSLPolicies(input)
	if err != nil {
		return nil
	}

	if len(resp.SslPolicies) == 0 {
		return fmt.Errorf("SSL Policies not found")
	}

	policy := resp.SslPolicies[0]
	d.Set("name", policy.Name)
	d.SetId(*policy.Name)
	return nil
}
