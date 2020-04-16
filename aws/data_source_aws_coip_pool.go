package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsCoipPool() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCoipPoolRead,

		Schema: map[string]*schema.Schema{
			"local_gateway_route_table_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"pool_cidrs": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Set:      schema.HashString,
			},

			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"tags": tagsSchemaComputed(),

			"filter": ec2CustomFiltersSchema(),
		},
	}
}

func dataSourceAwsCoipPoolRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.DescribeCoipPoolsInput{}

	var id string
	if cid, ok := d.GetOk("id"); ok {
		id = cid.(string)
	}

	if id != "" {
		req.PoolIds = []*string{aws.String(id)}
	}

	filters := map[string]string{}

	if v, ok := d.GetOk("local_gateway_route_table_id"); ok {
		filters["coip-pool.local-gateway-route-table-id"] = v.(string)
	}

	req.Filters = buildEC2AttributeFilterList(filters)

	if tags, tagsOk := d.GetOk("tags"); tagsOk {
		req.Filters = append(req.Filters, buildEC2TagFilterList(
			keyvaluetags.New(tags.(map[string]interface{})).Ec2Tags(),
		)...)
	}

	req.Filters = append(req.Filters, buildEC2CustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)
	if len(req.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		req.Filters = nil
	}

	log.Printf("[DEBUG] Reading AWS COIP Pool: %s", req)
	resp, err := conn.DescribeCoipPools(req)
	if err != nil {
		return err
	}
	if resp == nil || len(resp.CoipPools) == 0 {
		return fmt.Errorf("no matching COIP Pool found")
	}
	if len(resp.CoipPools) > 1 {
		return fmt.Errorf("multiple Coip Pools matched; use additional constraints to reduce matches to a single COIP Pool")
	}

	coip := resp.CoipPools[0]

	d.SetId(aws.StringValue(coip.PoolId))
	d.Set("pool_cidrs", coip.PoolCidrs)
	d.Set("local_gateway_route_table_id", coip.LocalGatewayRouteTableId)

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(coip.Tags).IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}
