package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceLocalGatewayVirtualInterfaceGroups() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceLocalGatewayVirtualInterfaceGroupsRead,

		Schema: map[string]*schema.Schema{
			"filter": DataSourceFiltersSchema(),
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"local_gateway_virtual_interface_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceLocalGatewayVirtualInterfaceGroupsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeLocalGatewayVirtualInterfaceGroupsInput{}

	input.Filters = append(input.Filters, BuildTagFilterList(
		Tags(tftags.New(d.Get("tags").(map[string]interface{}))),
	)...)

	input.Filters = append(input.Filters, BuildFiltersDataSource(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	output, err := FindLocalGatewayVirtualInterfaceGroups(conn, input)

	if err != nil {
		return fmt.Errorf("error reading EC2 Local Gateway Virtual Interface Groups: %w", err)
	}

	var groupIDs, interfaceIDs []string

	for _, v := range output {
		groupIDs = append(groupIDs, aws.StringValue(v.LocalGatewayVirtualInterfaceGroupId))
		interfaceIDs = append(interfaceIDs, aws.StringValueSlice(v.LocalGatewayVirtualInterfaceIds)...)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("ids", groupIDs)
	d.Set("local_gateway_virtual_interface_ids", interfaceIDs)

	return nil
}
