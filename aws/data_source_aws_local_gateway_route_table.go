package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsLocalGatewayRouteTable() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLocalGatewayRouteTableRead,

		Schema: map[string]*schema.Schema{
			"local_gateway_route_table_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"local_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"outpost_arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"state": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"tags": tagsSchemaComputed(),

			"filter": ec2CustomFiltersSchema(),
		},
	}
}

func dataSourceAwsLocalGatewayRouteTableRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.DescribeLocalGatewayRouteTablesInput{}

	var id string
	if cid, ok := d.GetOk("local_gateway_route_table_id"); ok {
		id = cid.(string)
	}

	if id != "" {
		req.LocalGatewayRouteTableIds = []*string{aws.String(id)}
	}

	req.Filters = buildEC2AttributeFilterList(
		map[string]string{
			"local-gateway-id": d.Get("local_gateway_id").(string),
			"outpost-arn":      d.Get("outpost_arn").(string),
			"state":            d.Get("state").(string),
		},
	)

	req.Filters = append(req.Filters, buildEC2TagFilterList(
		keyvaluetags.New(d.Get("tags").(map[string]interface{})).Ec2Tags(),
	)...)

	req.Filters = append(req.Filters, buildEC2CustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)
	if len(req.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		req.Filters = nil
	}

	log.Printf("[DEBUG] Reading AWS Local Gateway Route Table: %s", req)
	resp, err := conn.DescribeLocalGatewayRouteTables(req)
	if err != nil {
		return err
	}
	if resp == nil || len(resp.LocalGatewayRouteTables) == 0 {
		return fmt.Errorf("no matching Local Gateway Route Table found")
	}
	if len(resp.LocalGatewayRouteTables) > 1 {
		return fmt.Errorf("multiple Local Gateway Route Tables matched; use additional constraints to reduce matches to a single Local Gateway Route Table")
	}

	localgatewayroutetable := resp.LocalGatewayRouteTables[0]

	d.SetId(aws.StringValue(localgatewayroutetable.LocalGatewayRouteTableId))
	d.Set("local_gateway_id", localgatewayroutetable.LocalGatewayId)
	d.Set("outpost_arn", localgatewayroutetable.OutpostArn)
	d.Set("state", localgatewayroutetable.State)

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(localgatewayroutetable.Tags).IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}
