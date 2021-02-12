package aws

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsEc2TransitGatewayRouteTables() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEc2TransitGatewayRouteTablesRead,

		Schema: map[string]*schema.Schema{
			"transit_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"filter": ec2CustomFiltersSchema(),
			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func dataSourceAwsEc2TransitGatewayRouteTablesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.DescribeTransitGatewayRouteTablesInput{}

	if v, ok := d.GetOk("transit_gateway_id"); ok {
		input.Filters = buildEC2AttributeFilterList(
			map[string]string{
				"transit-gateway-id": v.(string),
			},
		)
	}

	input.Filters = append(input.Filters, buildEC2CustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)

	log.Printf("[DEBUG] Reading EC2 Transit Gateway Route Tables: %s", input)
	output, err := conn.DescribeTransitGatewayRouteTables(input)

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Route Table: %s", err)
	}

	if output == nil || len(output.TransitGatewayRouteTables) == 0 {
		return errors.New("error reading EC2 Transit Gateway Route Table: no results found")
	}

	d.SetId(meta.(*AWSClient).region)

	routeTables := make([]string, 0)

	for _, routeTable := range output.TransitGatewayRouteTables {
		routeTables = append(routeTables, aws.StringValue(routeTable.TransitGatewayRouteTableId))
	}

	if err = d.Set("ids", routeTables); err != nil {
		return fmt.Errorf("error setting ids: %s", err)
	}

	return nil
}
