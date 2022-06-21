package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceSecurityGroups() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSecurityGroupsRead,

		Schema: map[string]*schema.Schema{
			"arns": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"filter": DataSourceFiltersSchema(),
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tftags.TagsSchemaComputed(),
			"vpc_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceSecurityGroupsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeSecurityGroupsInput{}

	input.Filters = append(input.Filters, BuildTagFilterList(
		Tags(tftags.New(d.Get("tags").(map[string]interface{}))),
	)...)

	input.Filters = append(input.Filters, BuildFiltersDataSource(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	output, err := FindSecurityGroups(conn, input)

	if err != nil {
		return fmt.Errorf("error reading EC2 Security Groups: %w", err)
	}

	var arns, securityGroupIDs, vpcIDs []string

	for _, v := range output {
		arn := arn.ARN{
			Partition: meta.(*conns.AWSClient).Partition,
			Service:   ec2.ServiceName,
			Region:    meta.(*conns.AWSClient).Region,
			AccountID: aws.StringValue(v.OwnerId),
			Resource:  fmt.Sprintf("security-group/%s", aws.StringValue(v.GroupId)),
		}.String()
		arns = append(arns, arn)
		securityGroupIDs = append(securityGroupIDs, aws.StringValue(v.GroupId))
		vpcIDs = append(vpcIDs, aws.StringValue(v.VpcId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("arns", arns)
	d.Set("ids", securityGroupIDs)
	d.Set("vpc_ids", vpcIDs)

	return nil
}
