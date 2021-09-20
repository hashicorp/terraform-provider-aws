package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceRouteTables() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceRouteTablesRead,
		Schema: map[string]*schema.Schema{

			"filter": ec2CustomFiltersSchema(),

			"tags": tftags.TagsSchemaComputed(),

			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
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

func dataSourceRouteTablesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	req := &ec2.DescribeRouteTablesInput{}

	if v, ok := d.GetOk("vpc_id"); ok {
		req.Filters = BuildAttributeFilterList(
			map[string]string{
				"vpc-id": v.(string),
			},
		)
	}

	req.Filters = append(req.Filters, buildEC2TagFilterList(
		tftags.New(d.Get("tags").(map[string]interface{})).Ec2Tags(),
	)...)

	req.Filters = append(req.Filters, buildEC2CustomFilterList(
		d.Get("filter").(*schema.Set),
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
		routeTables = append(routeTables, aws.StringValue(routeTable.RouteTableId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	if err = d.Set("ids", routeTables); err != nil {
		return fmt.Errorf("error setting ids: %w", err)
	}

	return nil
}
