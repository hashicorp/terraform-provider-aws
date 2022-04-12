package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceNATGateways() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNATGatewaysRead,

		Schema: map[string]*schema.Schema{
			"filter": DataSourceFiltersSchema(),
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tftags.TagsSchemaComputed(),
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceNATGatewaysRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeNatGatewaysInput{}

	if v, ok := d.GetOk("vpc_id"); ok {
		input.Filter = append(input.Filter, BuildAttributeFilterList(
			map[string]string{
				"vpc-id": v.(string),
			},
		)...)
	}

	if tags, ok := d.GetOk("tags"); ok {
		input.Filter = append(input.Filter, BuildTagFilterList(
			Tags(tftags.New(tags.(map[string]interface{}))),
		)...)
	}

	input.Filter = append(input.Filter, BuildFiltersDataSource(
		d.Get("filter").(*schema.Set),
	)...)
	if len(input.Filter) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filter = nil
	}

	ngws, err := FindNATGateways(conn, input)

	if err != nil {
		return fmt.Errorf("error reading EC2 NAT Gateways: %w", err)
	}

	var natGatewayIds []string

	for _, v := range ngws {
		natGatewayIds = append(natGatewayIds, aws.StringValue(v.NatGatewayId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("ids", natGatewayIds)

	return nil
}
