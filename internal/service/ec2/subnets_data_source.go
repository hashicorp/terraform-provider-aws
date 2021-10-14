package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceSubnets() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSubnetsRead,
		Schema: map[string]*schema.Schema{
			"filter": dataSourceFiltersSchema(),
			"tags":   tftags.TagsSchemaComputed(),

			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceSubnetsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeSubnetsInput{}

	if tags, tagsOk := d.GetOk("tags"); tagsOk {
		input.Filters = append(input.Filters, BuildTagFilterList(
			tftags.New(tags.(map[string]interface{})).Ec2Tags(),
		)...)
	}

	if filters, filtersOk := d.GetOk("filter"); filtersOk {
		input.Filters = append(input.Filters,
			buildAwsDataSourceFilters(filters.(*schema.Set))...)
	}

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	var subnetIDs []*string
	err := conn.DescribeSubnetsPages(input, func(page *ec2.DescribeSubnetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, subnet := range page.Subnets {
			subnetIDs = append(subnetIDs, subnet.SubnetId)
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error reading Subnets: %w", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("ids", aws.StringValueSlice(subnetIDs))

	return nil
}
