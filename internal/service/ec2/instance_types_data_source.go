package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceInstanceTypes() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceInstanceTypesRead,

		Schema: map[string]*schema.Schema{
			"filter": DataSourceFiltersSchema(),
			"instance_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceInstanceTypesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeInstanceTypesInput{}

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = BuildFiltersDataSource(v.(*schema.Set))
	}

	var instanceTypes []string

	err := conn.DescribeInstanceTypesPages(input, func(page *ec2.DescribeInstanceTypesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, instanceType := range page.InstanceTypes {
			if instanceType == nil {
				continue
			}

			instanceTypes = append(instanceTypes, aws.StringValue(instanceType.InstanceType))
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error listing EC2 Instance Types: %w", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("instance_types", instanceTypes)

	return nil
}
