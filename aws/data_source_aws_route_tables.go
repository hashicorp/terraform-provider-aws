package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsRouteTables() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsRouteTableIDsRead,
		Schema: map[string]*schema.Schema{

			"tags": tagsSchemaComputed(),

			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func dataSourceAwsRouteTableIDsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.DescribeRouteTablesInput{}

	req.Filters = buildEC2AttributeFilterList(
		map[string]string{
			"vpc-id": d.Get("vpc_id").(string),
		},
	)

	req.Filters = append(req.Filters, buildEC2TagFilterList(
		tagsFromMap(d.Get("tags").(map[string]interface{})),
	)...)

	log.Printf("[DEBUG] DescribeRouteTables %s\n", req)
	resp, err := conn.DescribeRouteTables(req)
	if err != nil {
		return err
	}

	if resp == nil || len(resp.RouteTables) == 0 {
		return fmt.Errorf("no matching route tables found for vpc with id %s", d.Get("vpc_id").(string))
	}

	routeTables := make([]string, 0)

	for _, routeTable := range resp.RouteTables {
		routeTables = append(routeTables, *routeTable.RouteTableId)
	}

	d.SetId(d.Get("vpc_id").(string))
	d.Set("ids", routeTables)

	return nil
}
