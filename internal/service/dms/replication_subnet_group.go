// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
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
				MinItems: 2,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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

func resourceReplicationSubnetGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	replicationSubnetGroupID := d.Get("replication_subnet_group_id").(string)
	input := &dms.CreateReplicationSubnetGroupInput{
		ReplicationSubnetGroupDescription: aws.String(d.Get("replication_subnet_group_description").(string)),
		ReplicationSubnetGroupIdentifier:  aws.String(replicationSubnetGroupID),
		SubnetIds:                         flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
		Tags:                              getTagsIn(ctx),
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreateReplicationSubnetGroupWithContext(ctx, input)
	}, dms.ErrCodeAccessDeniedFault)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DMS Replication Subnet Group (%s): %s", replicationSubnetGroupID, err)
	}

	d.SetId(replicationSubnetGroupID)

	return append(diags, resourceReplicationSubnetGroupRead(ctx, d, meta)...)
}

func resourceReplicationSubnetGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	group, err := FindReplicationSubnetGroupByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DMS Replication Subnet Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DMS Replication Subnet Group (%s): %s", d.Id(), err)
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
	d.Set("replication_subnet_group_description", group.ReplicationSubnetGroupDescription)
	d.Set("replication_subnet_group_id", group.ReplicationSubnetGroupIdentifier)
	subnetIDs := tfslices.ApplyToAll(group.Subnets, func(sn *dms.Subnet) string {
		return aws.StringValue(sn.SubnetIdentifier)
	})
	d.Set("subnet_ids", subnetIDs)
	d.Set("vpc_id", group.VpcId)

	return diags
}

func resourceReplicationSubnetGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		// Updates to subnet groups are only valid when sending SubnetIds even if there are no
		// changes to SubnetIds.
		input := &dms.ModifyReplicationSubnetGroupInput{
			ReplicationSubnetGroupIdentifier: aws.String(d.Get("replication_subnet_group_id").(string)),
			SubnetIds:                        flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
		}

		if d.HasChange("replication_subnet_group_description") {
			input.ReplicationSubnetGroupDescription = aws.String(d.Get("replication_subnet_group_description").(string))
		}

		_, err := conn.ModifyReplicationSubnetGroupWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DMS Replication Subnet Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceReplicationSubnetGroupRead(ctx, d, meta)...)
}

func resourceReplicationSubnetGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	log.Printf("[DEBUG] Deleting DMS Replication Subnet Group: %s", d.Id())
	_, err := conn.DeleteReplicationSubnetGroupWithContext(ctx, &dms.DeleteReplicationSubnetGroupInput{
		ReplicationSubnetGroupIdentifier: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DMS Replication Subnet Group (%s): %s", d.Id(), err)
	}

	return diags
}

func FindReplicationSubnetGroupByID(ctx context.Context, conn *dms.DatabaseMigrationService, id string) (*dms.ReplicationSubnetGroup, error) {
	input := &dms.DescribeReplicationSubnetGroupsInput{
		Filters: []*dms.Filter{
			{
				Name:   aws.String("replication-subnet-group-id"),
				Values: aws.StringSlice([]string{id}),
			},
		},
	}

	return findReplicationSubnetGroup(ctx, conn, input)
}

func findReplicationSubnetGroup(ctx context.Context, conn *dms.DatabaseMigrationService, input *dms.DescribeReplicationSubnetGroupsInput) (*dms.ReplicationSubnetGroup, error) {
	output, err := findReplicationSubnetGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findReplicationSubnetGroups(ctx context.Context, conn *dms.DatabaseMigrationService, input *dms.DescribeReplicationSubnetGroupsInput) ([]*dms.ReplicationSubnetGroup, error) {
	var output []*dms.ReplicationSubnetGroup

	err := conn.DescribeReplicationSubnetGroupsPagesWithContext(ctx, input, func(page *dms.DescribeReplicationSubnetGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ReplicationSubnetGroups {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
