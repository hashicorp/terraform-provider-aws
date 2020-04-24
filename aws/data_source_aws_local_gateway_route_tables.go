package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsLocalGatewayRouteTables() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLocalGatewayRouteTablesRead,
		Schema: map[string]*schema.Schema{
			"filter": ec2CustomFiltersSchema(),

			"tags": tagsSchemaComputed(),

			"local_gateway_route_table_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func dataSourceAwsLocalGatewayRouteTablesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.DescribeLocalGatewayRouteTablesInput{}

	if tags, tagsOk := d.GetOk("tags"); tagsOk {
		req.Filters = append(req.Filters, buildEC2TagFilterList(
			keyvaluetags.New(tags.(map[string]interface{})).Ec2Tags(),
		)...)
	}

	if filters, filtersOk := d.GetOk("filter"); filtersOk {
		req.Filters = append(req.Filters, buildEC2CustomFilterList(
			filters.(*schema.Set),
		)...)
	}
	if len(req.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		req.Filters = nil
	}

	log.Printf("[DEBUG] DescribeLocalGatewayRouteTables %s\n", req)
	resp, err := conn.DescribeLocalGatewayRouteTables(req)
	if err != nil {
		return err
	}

	if resp == nil || len(resp.LocalGatewayRouteTables) == 0 {
		return fmt.Errorf("no matching Local Gateway Route Table found")
	}

	localgatewayroutetables := make([]string, 0)

	for _, localgatewayroutetable := range resp.LocalGatewayRouteTables {
		localgatewayroutetables = append(localgatewayroutetables, aws.StringValue(localgatewayroutetable.LocalGatewayRouteTableId))
	}

	d.SetId(time.Now().UTC().String())
	if err := d.Set("local_gateway_route_table_ids", localgatewayroutetables); err != nil {
		return fmt.Errorf("Error setting local gateway route table ids: %s", err)
	}

	return nil
}
