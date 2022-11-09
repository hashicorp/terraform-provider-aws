package ec2

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceVPCPeeringConnections() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVPCPeeringConnectionsRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

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

func dataSourceVPCPeeringConnectionsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeVpcPeeringConnectionsInput{}

	input.Filters = append(input.Filters, BuildTagFilterList(
		Tags(tftags.New(d.Get("tags").(map[string]interface{}))),
	)...)
	input.Filters = append(input.Filters, BuildFiltersDataSource(
		d.Get("filter").(*schema.Set),
	)...)
	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	output, err := FindVPCPeeringConnections(conn, input)

	if err != nil {
		return fmt.Errorf("error reading EC2 VPC Peering Connections: %w", err)
	}

	var vpcPeeringConnectionIDs []string

	for _, v := range output {
		vpcPeeringConnectionIDs = append(vpcPeeringConnectionIDs, aws.StringValue(v.VpcPeeringConnectionId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("ids", vpcPeeringConnectionIDs)

	return nil
}
