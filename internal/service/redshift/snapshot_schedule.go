// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshift_snapshot_schedule", name="Snapshot Schedule")
// @Tags(identifierAttribute="arn")
func ResourceSnapshotSchedule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSnapshotScheduleCreate,
		ReadWithoutTimeout:   resourceSnapshotScheduleRead,
		UpdateWithoutTimeout: resourceSnapshotScheduleUpdate,
		DeleteWithoutTimeout: resourceSnapshotScheduleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identifier": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"identifier_prefix"},
			},
			"identifier_prefix": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"definitions": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceSnapshotScheduleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	var identifier string
	if v, ok := d.GetOk("identifier"); ok {
		identifier = v.(string)
	} else {
		if v, ok := d.GetOk("identifier_prefix"); ok {
			identifier = id.PrefixedUniqueId(v.(string))
		} else {
			identifier = id.UniqueId()
		}
	}
	createOpts := &redshift.CreateSnapshotScheduleInput{
		ScheduleIdentifier:  aws.String(identifier),
		ScheduleDefinitions: flex.ExpandStringSet(d.Get("definitions").(*schema.Set)),
		Tags:                getTagsIn(ctx),
	}
	if attr, ok := d.GetOk("description"); ok {
		createOpts.ScheduleDescription = aws.String(attr.(string))
	}

	resp, err := conn.CreateSnapshotScheduleWithContext(ctx, createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Snapshot Schedule: %s", err)
	}

	d.SetId(aws.StringValue(resp.ScheduleIdentifier))

	return append(diags, resourceSnapshotScheduleRead(ctx, d, meta)...)
}

func resourceSnapshotScheduleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	descOpts := &redshift.DescribeSnapshotSchedulesInput{
		ScheduleIdentifier: aws.String(d.Id()),
	}

	resp, err := conn.DescribeSnapshotSchedulesWithContext(ctx, descOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Redshift Cluster Snapshot Schedule %s: %s", d.Id(), err)
	}

	if !d.IsNewResource() && (resp.SnapshotSchedules == nil || len(resp.SnapshotSchedules) != 1) {
		log.Printf("[WARN] Redshift Cluster Snapshot Schedule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	snapshotSchedule := resp.SnapshotSchedules[0]

	d.Set("identifier", snapshotSchedule.ScheduleIdentifier)
	d.Set("description", snapshotSchedule.ScheduleDescription)
	if err := d.Set("definitions", flex.FlattenStringList(snapshotSchedule.ScheduleDefinitions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting definitions: %s", err)
	}

	setTagsOut(ctx, snapshotSchedule.Tags)

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "redshift",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("snapshotschedule:%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	return diags
}

func resourceSnapshotScheduleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	if d.HasChange("definitions") {
		modifyOpts := &redshift.ModifySnapshotScheduleInput{
			ScheduleIdentifier:  aws.String(d.Id()),
			ScheduleDefinitions: flex.ExpandStringSet(d.Get("definitions").(*schema.Set)),
		}
		_, err := conn.ModifySnapshotScheduleWithContext(ctx, modifyOpts)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying Redshift Snapshot Schedule %s: %s", d.Id(), err)
		}
	}

	return append(diags, resourceSnapshotScheduleRead(ctx, d, meta)...)
}

func resourceSnapshotScheduleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	if d.Get("force_destroy").(bool) {
		if err := resourceSnapshotScheduleDeleteAllAssociatedClusters(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting Redshift Snapshot Schedule (%s): %s", d.Id(), err)
		}
	}

	_, err := conn.DeleteSnapshotScheduleWithContext(ctx, &redshift.DeleteSnapshotScheduleInput{
		ScheduleIdentifier: aws.String(d.Id()),
	})
	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeSnapshotScheduleNotFoundFault) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Snapshot Schedule (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceSnapshotScheduleDeleteAllAssociatedClusters(ctx context.Context, conn *redshift.Redshift, scheduleIdentifier string) error {
	resp, err := conn.DescribeSnapshotSchedulesWithContext(ctx, &redshift.DescribeSnapshotSchedulesInput{
		ScheduleIdentifier: aws.String(scheduleIdentifier),
	})
	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeSnapshotScheduleNotFoundFault) {
		return nil
	}
	if err != nil {
		return err
	}
	if resp.SnapshotSchedules == nil || len(resp.SnapshotSchedules) != 1 {
		log.Printf("[WARN] Unable to find Redshift Cluster Snapshot Schedule (%s)", scheduleIdentifier)
		return nil
	}

	snapshotSchedule := resp.SnapshotSchedules[0]

	for _, associatedCluster := range snapshotSchedule.AssociatedClusters {
		_, err = conn.ModifyClusterSnapshotScheduleWithContext(ctx, &redshift.ModifyClusterSnapshotScheduleInput{
			ClusterIdentifier:    associatedCluster.ClusterIdentifier,
			ScheduleIdentifier:   aws.String(scheduleIdentifier),
			DisassociateSchedule: aws.Bool(true),
		})

		clusterId := aws.StringValue(associatedCluster.ClusterIdentifier)

		if tfawserr.ErrCodeEquals(err, redshift.ErrCodeClusterNotFoundFault) {
			log.Printf("[WARN] Redshift Snapshot Cluster (%s) not found, removing from state", clusterId)
			continue
		}
		if tfawserr.ErrCodeEquals(err, redshift.ErrCodeSnapshotScheduleNotFoundFault) {
			log.Printf("[WARN] Redshift Snapshot Schedule (%s) not found, removing from state", scheduleIdentifier)
			continue
		}
		if err != nil {
			return fmt.Errorf("disassociating Redshift Cluster (%s): %s", clusterId, err)
		}
	}

	for _, associatedCluster := range snapshotSchedule.AssociatedClusters {
		id := fmt.Sprintf("%s/%s", aws.StringValue(associatedCluster.ClusterIdentifier), scheduleIdentifier)
		if _, err := waitScheduleAssociationDeleted(ctx, conn, id); err != nil {
			return err
		}
	}

	return nil
}
