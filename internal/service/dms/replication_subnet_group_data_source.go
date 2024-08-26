// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	awstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_dms_replication_subnet_group", name="Replication Subnet Group")
func dataSourceReplicationSubnetGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceReplicationSubnetGroupRead,

		Schema: map[string]*schema.Schema{
			"replication_subnet_group_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replication_subnet_group_description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replication_subnet_group_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"subnet_group_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrSubnetIDs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceReplicationSubnetGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSClient(ctx)
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	replicationSubnetGroupID := d.Get("replication_subnet_group_id").(string)
	group, err := findReplicationSubnetGroupByID(ctx, conn, replicationSubnetGroupID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DMS Replication Subnet Group (%s): %s", replicationSubnetGroupID, err)
	}

	d.SetId(aws.ToString(group.ReplicationSubnetGroupIdentifier))
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "dms",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("subgrp:%s", d.Id()),
	}.String()
	d.Set("replication_subnet_group_arn", arn)
	d.Set("replication_subnet_group_description", group.ReplicationSubnetGroupDescription)
	d.Set("replication_subnet_group_id", group.ReplicationSubnetGroupIdentifier)
	subnetIDs := tfslices.ApplyToAll(group.Subnets, func(sn awstypes.Subnet) string {
		return aws.ToString(sn.SubnetIdentifier)
	})
	d.Set(names.AttrSubnetIDs, subnetIDs)
	d.Set(names.AttrVPCID, group.VpcId)

	tags, err := listTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for DMS Replication Subnet Group (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set(names.AttrTags, tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
