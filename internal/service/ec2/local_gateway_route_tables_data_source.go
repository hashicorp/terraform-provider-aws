package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceLocalGatewayRouteTables() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceLocalGatewayRouteTablesRead,
		Schema: map[string]*schema.Schema{
			"filter": CustomFiltersSchema(),

			"tags": tftags.TagsSchemaComputed(),

			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func dataSourceLocalGatewayRouteTablesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	req := &ec2.DescribeLocalGatewayRouteTablesInput{}

	req.Filters = append(req.Filters, BuildTagFilterList(
		Tags(tftags.New(d.Get("tags").(map[string]interface{}))),
	)...)

	req.Filters = append(req.Filters, BuildCustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)
	if len(req.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		req.Filters = nil
	}

	var localGatewayRouteTables []*ec2.LocalGatewayRouteTable

	err := conn.DescribeLocalGatewayRouteTablesPages(req, func(page *ec2.DescribeLocalGatewayRouteTablesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		localGatewayRouteTables = append(localGatewayRouteTables, page.LocalGatewayRouteTables...)

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error describing EC2 Local Gateway Route Tables: %w", err)
	}

	if len(localGatewayRouteTables) == 0 {
		return fmt.Errorf("no matching EC2 Local Gateway Route Tables found")
	}

	var ids []string

	for _, localGatewayRouteTable := range localGatewayRouteTables {
		if localGatewayRouteTable == nil {
			continue
		}

		ids = append(ids, aws.StringValue(localGatewayRouteTable.LocalGatewayRouteTableId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	if err := d.Set("ids", ids); err != nil {
		return fmt.Errorf("error setting ids: %w", err)
	}

	return nil
}
