package ec2

import (
	"errors"
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
			"filter": CustomFiltersSchema(),
			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceNetworkInterfacesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeNetworkInterfacesInput{}

	if v, ok := d.GetOk("tags"); ok {
		input.Filters = BuildTagFilterList(
			Tags(tftags.New(v.(map[string]interface{}))),
		)
	}

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = append(input.Filters, BuildCustomFilterList(
			v.(*schema.Set),
		)...)
	}

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	networkInterfaceIDs := []string{}

	output, err := FindNetworkInterfaces(conn, input)

	if err != nil {
		return fmt.Errorf("error reading EC2 Network Interfaces: %w", err)
	}

	if len(output) == 0 {
		return errors.New("no matching network interfaces found")
	}

	for _, v := range output {
		networkInterfaceIDs = append(networkInterfaceIDs, aws.StringValue(v.NetworkInterfaceId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	if err := d.Set("ids", networkInterfaceIDs); err != nil {
		return fmt.Errorf("error setting ids: %w", err)
	}

	return nil
}
