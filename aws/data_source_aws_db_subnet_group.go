package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsDbSubnetGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsDbSubnetGroupRead,

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
				ForceNew: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"subnets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsDbSubnetGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	opts := &rds.DescribeDBSubnetGroupsInput{
		DBSubnetGroupName: aws.String(d.Get("name").(string)),
	}

	log.Printf("[DEBUG] Reading DB SubnetGroup: %s", opts)

	resp, err := conn.DescribeDBSubnetGroups(opts)
	if err != nil {
		return err
	}

	if len(resp.DBSubnetGroups) < 1 {
		return fmt.Errorf("Your query returned no results. Please change your search criteria and try again.")
	}
	if len(resp.DBSubnetGroups) > 1 {
		return fmt.Errorf("Your query returned more than one result. Please try a more specific search criteria.")
	}

	dbSubnetGroup := *resp.DBSubnetGroups[0]

	d.SetId(d.Get("name").(string))

	d.Set("arn", dbSubnetGroup.DBSubnetGroupArn)
	d.Set("description", dbSubnetGroup.DBSubnetGroupDescription)
	d.Set("name", dbSubnetGroup.DBSubnetGroupName)
	d.Set("status", dbSubnetGroup.SubnetGroupStatus)

	var subnets []string
	for _, v := range dbSubnetGroup.Subnets {
		subnets = append(subnets, *v.SubnetIdentifier)
	}
	if err := d.Set("subnets", subnets); err != nil {
		return fmt.Errorf("Error setting subnets attribute: %#v, error: %#v", subnets, err)
	}

	d.Set("vpc_id", dbSubnetGroup.VpcId)

	return nil
}
