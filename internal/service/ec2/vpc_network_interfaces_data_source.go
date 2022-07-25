package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceNetworkInterfaces() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNetworkInterfacesRead,

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

func dataSourceNetworkInterfacesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeNetworkInterfacesInput{}

	input.Filters = append(input.Filters, BuildTagFilterList(
		Tags(tftags.New(d.Get("tags").(map[string]interface{}))),
	)...)

	input.Filters = append(input.Filters, BuildFiltersDataSource(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	networkInterfaceIDs := []string{}

	output, err := FindNetworkInterfaces(conn, input)

	if err != nil {
		return fmt.Errorf("error reading EC2 Network Interfaces: %w", err)
	}

	for _, v := range output {
		networkInterfaceIDs = append(networkInterfaceIDs, aws.StringValue(v.NetworkInterfaceId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("ids", networkInterfaceIDs)

	return nil
}
