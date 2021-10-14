package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceSubnetIDs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSubnetIDsRead,
		Schema: map[string]*schema.Schema{
			"filter": CustomFiltersSchema(),

			"tags": tftags.TagsSchemaComputed(),

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

func dataSourceSubnetIDsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	req := &ec2.DescribeSubnetsInput{}

	if vpc, vpcOk := d.GetOk("vpc_id"); vpcOk {
		req.Filters = BuildAttributeFilterList(
			map[string]string{
				"vpc-id": vpc.(string),
			},
		)
	}

	if tags, tagsOk := d.GetOk("tags"); tagsOk {
		req.Filters = append(req.Filters, BuildTagFilterList(
			Tags(tftags.New(tags.(map[string]interface{}))),
		)...)
	}

	if filters, filtersOk := d.GetOk("filter"); filtersOk {
		req.Filters = append(req.Filters, BuildCustomFilterList(
			filters.(*schema.Set),
		)...)
	}

	if len(req.Filters) == 0 {
		req.Filters = nil
	}

	log.Printf("[DEBUG] DescribeSubnets %s\n", req)
	resp, err := conn.DescribeSubnets(req)
	if err != nil {
		return err
	}

	if resp == nil || len(resp.Subnets) == 0 {
		return fmt.Errorf("no matching subnet found for vpc with id %s", d.Get("vpc_id").(string))
	}

	subnets := make([]string, 0)

	for _, subnet := range resp.Subnets {
		subnets = append(subnets, *subnet.SubnetId)
	}

	d.SetId(d.Get("vpc_id").(string))
	d.Set("ids", subnets)

	return nil
}
