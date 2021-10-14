package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceLocalGateways() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceLocalGatewaysRead,
		Schema: map[string]*schema.Schema{
			"filter": CustomFiltersSchema(),

			"tags": tftags.TagsSchemaComputed(),

			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func dataSourceLocalGatewaysRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	req := &ec2.DescribeLocalGatewaysInput{}

	if tags, tagsOk := d.GetOk("tags"); tagsOk {
		req.Filters = append(req.Filters, BuildTagFilterList(
			tftags.New(tags.(map[string]interface{})).Ec2Tags(),
		)...)
	}

	if filters, filtersOk := d.GetOk("filter"); filtersOk {
		req.Filters = append(req.Filters, BuildCustomFilterList(
			filters.(*schema.Set),
		)...)
	}
	if len(req.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		req.Filters = nil
	}

	var localGateways []*ec2.LocalGateway

	err := conn.DescribeLocalGatewaysPages(req, func(page *ec2.DescribeLocalGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		localGateways = append(localGateways, page.LocalGateways...)

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error describing EC2 Local Gateways: %w", err)
	}

	if len(localGateways) == 0 {
		return fmt.Errorf("no matching EC2 Local Gateways found")
	}

	var ids []string

	for _, localGateway := range localGateways {
		if localGateway == nil {
			continue
		}

		ids = append(ids, aws.StringValue(localGateway.LocalGatewayId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	if err := d.Set("ids", ids); err != nil {
		return fmt.Errorf("error setting ids: %w", err)
	}

	return nil
}
