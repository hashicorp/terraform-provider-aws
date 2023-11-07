// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKDataSource("aws_security_group")
func DataSourceSecurityGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSecurityGroupRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"filter": CustomFiltersSchema(),
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceSecurityGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeSecurityGroupsInput{
		Filters: BuildAttributeFilterList(
			map[string]string{
				"group-name": d.Get("name").(string),
				"vpc-id":     d.Get("vpc_id").(string),
			},
		),
	}

	if v, ok := d.GetOk("id"); ok {
		input.GroupIds = aws.StringSlice([]string{v.(string)})
	}

	input.Filters = append(input.Filters, BuildTagFilterList(
		Tags(tftags.New(ctx, d.Get("tags").(map[string]interface{}))),
	)...)

	input.Filters = append(input.Filters, BuildCustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	sg, err := FindSecurityGroup(ctx, conn, input)

	if err != nil {
		return diag.FromErr(tfresource.SingularDataSourceFindError("EC2 Security Group", err))
	}

	d.SetId(aws.StringValue(sg.GroupId))

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: *sg.OwnerId,
		Resource:  fmt.Sprintf("security-group/%s", *sg.GroupId),
	}.String()
	d.Set("arn", arn)
	d.Set("description", sg.Description)
	d.Set("name", sg.GroupName)
	d.Set("vpc_id", sg.VpcId)

	if err := d.Set("tags", KeyValueTags(ctx, sg.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	return nil
}
