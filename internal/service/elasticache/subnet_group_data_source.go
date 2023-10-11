// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_elasticache_subnet_group")
func DataSourceSubnetGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSubnetGroupRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tftags.TagsSchema(),
		},
	}
}

func dataSourceSubnetGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)

	group, err := FindCacheSubnetGroupByName(ctx, conn, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ElastiCache Subnet Group (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(group.CacheSubnetGroupName))

	var subnetIds []*string
	for _, subnet := range group.Subnets {
		subnetIds = append(subnetIds, subnet.SubnetIdentifier)
	}

	d.Set("arn", group.ARN)
	d.Set("description", group.CacheSubnetGroupDescription)
	d.Set("subnet_ids", flex.FlattenStringSet(subnetIds))
	d.Set("name", group.CacheSubnetGroupName)

	tags, err := listTags(ctx, conn, d.Get("arn").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for ElastiCache Subnet Group (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
