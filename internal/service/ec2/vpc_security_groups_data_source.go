package ec2

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceSecurityGroups() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSecurityGroupsRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

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

func dataSourceSecurityGroupsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()

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

	output, err := FindSecurityGroups(ctx, conn, input)

	if err != nil {
		return diag.Errorf("reading EC2 Security Groups: %s", err)
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
