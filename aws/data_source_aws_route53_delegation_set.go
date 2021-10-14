package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceDelegationSet() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDelegationSetRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"caller_reference": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name_servers": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
		},
	}
}

func dataSourceDelegationSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53Conn

	dSetID := d.Get("id").(string)

	input := &route53.GetReusableDelegationSetInput{
		Id: aws.String(dSetID),
	}

	log.Printf("[DEBUG] Reading Route53 delegation set: %s", input)

	resp, err := conn.GetReusableDelegationSet(input)
	if err != nil {
		return fmt.Errorf("Failed getting Route53 delegation set (%s): %w", dSetID, err)
	}

	d.SetId(dSetID)
	d.Set("caller_reference", resp.DelegationSet.CallerReference)

	if err := d.Set("name_servers", aws.StringValueSlice(resp.DelegationSet.NameServers)); err != nil {
		return fmt.Errorf("error setting name_servers: %w", err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "route53",
		Resource:  fmt.Sprintf("delegationset/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	return nil
}
