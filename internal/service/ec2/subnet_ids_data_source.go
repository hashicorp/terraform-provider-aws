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

func DataSourceSubnetIDs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSubnetIDsRead,
		Schema: map[string]*schema.Schema{
			"filter": CustomFiltersSchema(),
			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tftags.TagsSchemaComputed(),
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
		DeprecationMessage: `The aws_subnet_ids data source has been deprecated and will be removed in a future version. ` +
			`Use the aws_subnets data source instead.`,
	}
}

func dataSourceSubnetIDsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeSubnetsInput{}

	if vpc, vpcOk := d.GetOk("vpc_id"); vpcOk {
		input.Filters = BuildAttributeFilterList(
			map[string]string{
				"vpc-id": vpc.(string),
			},
		)
	}

	if tags, tagsOk := d.GetOk("tags"); tagsOk {
		input.Filters = append(input.Filters, BuildTagFilterList(
			Tags(tftags.New(tags.(map[string]interface{}))),
		)...)
	}

	if filters, filtersOk := d.GetOk("filter"); filtersOk {
		input.Filters = append(input.Filters, BuildCustomFilterList(
			filters.(*schema.Set),
		)...)
	}

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	output, err := FindSubnets(conn, input)

	if err != nil {
		return fmt.Errorf("error reading EC2 Subnets: %w", err)
	}

	if len(output) == 0 {
		return errors.New("no matching EC2 Subnets found")
	}

	var subnetIDs []string

	for _, v := range output {
		subnetIDs = append(subnetIDs, aws.StringValue(v.SubnetId))
	}

	d.SetId(d.Get("vpc_id").(string))
	d.Set("ids", subnetIDs)

	return nil
}
