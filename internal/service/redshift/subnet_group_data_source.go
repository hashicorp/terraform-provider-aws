// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_redshift_subnet_group", name="Subnet Group")
// @Tags
func dataSourceSubnetGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSubnetGroupRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrSubnetIDs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceSubnetGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	subnetgroup, err := findSubnetGroupByName(ctx, conn, d.Get(names.AttrName).(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Subnet Group (%s): %s", d.Id(), err)
	}

	d.SetId(aws.StringValue(subnetgroup.ClusterSubnetGroupName))
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   redshift.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("subnetgroup:%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, subnetgroup.Description)
	d.Set(names.AttrName, subnetgroup.ClusterSubnetGroupName)
	d.Set(names.AttrSubnetIDs, subnetIdsToSlice(subnetgroup.Subnets))

	setTagsOut(ctx, subnetgroup.Tags)

	return diags
}
