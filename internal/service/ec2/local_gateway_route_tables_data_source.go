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
			"filter": DataSourceFiltersSchema(),
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceLocalGatewayRouteTablesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeLocalGatewayRouteTablesInput{}

	input.Filters = append(input.Filters, BuildTagFilterList(
		Tags(tftags.New(d.Get("tags").(map[string]interface{}))),
	)...)

	input.Filters = append(input.Filters, BuildFiltersDataSource(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	output, err := FindLocalGatewayRouteTables(conn, input)

	if err != nil {
		return fmt.Errorf("error reading EC2 Local Gateway Route Tables: %w", err)
	}

	var routeTableIDs []string

	for _, v := range output {
		routeTableIDs = append(routeTableIDs, aws.StringValue(v.LocalGatewayRouteTableId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("ids", routeTableIDs)

	return nil
}
