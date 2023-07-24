// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dms_replication_subnet_group", name="Replication Subnet Group")
// @Tags(identifierAttribute="replication_subnet_group_arn")
func ResourceReplicationSubnetGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceReplicationSubnetGroupCreate,
		ReadWithoutTimeout:   resourceReplicationSubnetGroupRead,
		UpdateWithoutTimeout: resourceReplicationSubnetGroupUpdate,
		DeleteWithoutTimeout: resourceReplicationSubnetGroupDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"replication_subnet_group_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replication_subnet_group_description": {
				Type:     schema.TypeString,
				Required: true,
			},
			"replication_subnet_group_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validReplicationSubnetGroupID,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				MinItems: 2,
				Required: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameReplicationSubnetGroup = "Replication Subnet Group"
)

func resourceReplicationSubnetGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	request := &dms.CreateReplicationSubnetGroupInput{
		ReplicationSubnetGroupIdentifier:  aws.String(d.Get("replication_subnet_group_id").(string)),
		ReplicationSubnetGroupDescription: aws.String(d.Get("replication_subnet_group_description").(string)),
		SubnetIds:                         flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
		Tags:                              getTagsIn(ctx),
	}

	log.Println("[DEBUG] DMS create replication subnet group:", request)

	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		_, err := conn.CreateReplicationSubnetGroupWithContext(ctx, request)

		if tfawserr.ErrCodeEquals(err, dms.ErrCodeAccessDeniedFault) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.CreateReplicationSubnetGroupWithContext(ctx, request)

		if err != nil {
			return create.DiagError(names.DMS, create.ErrActionCreating, ResNameReplicationSubnetGroup, d.Get("replication_subnet_group_id").(string), err)
		}
	}

	if err != nil {
		return create.DiagError(names.DMS, create.ErrActionCreating, ResNameReplicationSubnetGroup, d.Get("replication_subnet_group_id").(string), err)
	}

	d.SetId(d.Get("replication_subnet_group_id").(string))
	return append(diags, resourceReplicationSubnetGroupRead(ctx, d, meta)...)
}

func resourceReplicationSubnetGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	response, err := conn.DescribeReplicationSubnetGroupsWithContext(ctx, &dms.DescribeReplicationSubnetGroupsInput{
		Filters: []*dms.Filter{
			{
				Name:   aws.String("replication-subnet-group-id"),
				Values: []*string{aws.String(d.Id())}, // Must use d.Id() to work with import.
			},
		},
	})

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.DMS, create.ErrActionReading, ResNameReplicationSubnetGroup, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.DiagError(names.DMS, create.ErrActionReading, ResNameReplicationSubnetGroup, d.Id(), err)
	}

	if len(response.ReplicationSubnetGroups) == 0 {
		d.SetId("")
		return diags
	}

	// The AWS API for DMS subnet groups does not return the ARN which is required to
	// retrieve tags. This ARN can be built.
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "dms",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("subgrp:%s", d.Id()),
	}.String()
	d.Set("replication_subnet_group_arn", arn)

	group := response.ReplicationSubnetGroups[0]

	d.SetId(aws.StringValue(group.ReplicationSubnetGroupIdentifier))

	subnet_ids := []string{}
	for _, subnet := range group.Subnets {
		subnet_ids = append(subnet_ids, aws.StringValue(subnet.SubnetIdentifier))
	}

	d.Set("replication_subnet_group_description", group.ReplicationSubnetGroupDescription)
	d.Set("replication_subnet_group_id", group.ReplicationSubnetGroupIdentifier)
	d.Set("subnet_ids", subnet_ids)
	d.Set("vpc_id", group.VpcId)

	return diags
}

func resourceReplicationSubnetGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	// Updates to subnet groups are only valid when sending SubnetIds even if there are no
	// changes to SubnetIds.
	request := &dms.ModifyReplicationSubnetGroupInput{
		ReplicationSubnetGroupIdentifier: aws.String(d.Get("replication_subnet_group_id").(string)),
		SubnetIds:                        flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
	}

	if d.HasChange("replication_subnet_group_description") {
		request.ReplicationSubnetGroupDescription = aws.String(d.Get("replication_subnet_group_description").(string))
	}

	log.Println("[DEBUG] DMS update replication subnet group:", request)

	_, err := conn.ModifyReplicationSubnetGroupWithContext(ctx, request)
	if err != nil {
		return create.DiagError(names.DMS, create.ErrActionUpdating, ResNameReplicationSubnetGroup, d.Id(), err)
	}

	return append(diags, resourceReplicationSubnetGroupRead(ctx, d, meta)...)
}

func resourceReplicationSubnetGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	request := &dms.DeleteReplicationSubnetGroupInput{
		ReplicationSubnetGroupIdentifier: aws.String(d.Get("replication_subnet_group_id").(string)),
	}

	log.Printf("[DEBUG] DMS delete replication subnet group: %#v", request)

	_, err := conn.DeleteReplicationSubnetGroupWithContext(ctx, request)

	if err != nil {
		return create.DiagError(names.DMS, create.ErrActionDeleting, ResNameReplicationSubnetGroup, d.Id(), err)
	}

	return diags
}
