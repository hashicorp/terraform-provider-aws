// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_dms_replication_subnet_group")
func DataSourceReplicationSubnetGroup() *schema.Resource {
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
			"subnet_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNameReplicationSubnetGroup = "Replication Subnet Group Data Source"
)

func dataSourceReplicationSubnetGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DMSConn(ctx)
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	subnetID := d.Get("replication_subnet_group_id").(string)

	out, err := FindReplicationSubnetGroupByID(ctx, conn, subnetID)
	if err != nil {
		return create.DiagError(names.DMS, create.ErrActionReading, DSNameReplicationSubnetGroup, d.Id(), err)
	}

	d.SetId(aws.StringValue(out.ReplicationSubnetGroupIdentifier))

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "dms",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("subgrp:%s", d.Id()),
	}.String()
	d.Set("replication_subnet_group_arn", arn)

	d.Set("replication_subnet_group_description", out.ReplicationSubnetGroupDescription)
	d.Set("replication_subnet_group_id", out.ReplicationSubnetGroupIdentifier)
	d.Set("vpc_id", out.VpcId)

	subnetIDs := []string{}
	for _, subnet := range out.Subnets {
		subnetIDs = append(subnetIDs, aws.StringValue(subnet.SubnetIdentifier))
	}
	d.Set("subnet_ids", subnetIDs)

	tags, err := listTags(ctx, conn, arn)
	if err != nil {
		return create.DiagError(names.DMS, create.ErrActionReading, DSNameReplicationSubnetGroup, d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.DMS, create.ErrActionSetting, DSNameReplicationSubnetGroup, d.Id(), err)
	}
	return nil
}

func FindReplicationSubnetGroupByID(ctx context.Context, conn *dms.DatabaseMigrationService, id string) (*dms.ReplicationSubnetGroup, error) {
	input := &dms.DescribeReplicationSubnetGroupsInput{
		Filters: []*dms.Filter{
			{
				Name:   aws.String("replication-subnet-group-id"),
				Values: []*string{aws.String(id)},
			},
		},
	}
	response, err := conn.DescribeReplicationSubnetGroupsWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if response == nil || len(response.ReplicationSubnetGroups) == 0 || response.ReplicationSubnetGroups[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return response.ReplicationSubnetGroups[0], nil
}
