// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshift_snapshot_schedule_association", name="Snapshot Schedule Association")
func resourceSnapshotScheduleAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSnapshotScheduleAssociationCreate,
		ReadWithoutTimeout:   resourceSnapshotScheduleAssociationRead,
		DeleteWithoutTimeout: resourceSnapshotScheduleAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrClusterIdentifier: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"schedule_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSnapshotScheduleAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	clusterIdentifier := d.Get(names.AttrClusterIdentifier).(string)
	scheduleIdentifier := d.Get("schedule_identifier").(string)
	id := SnapshotScheduleAssociationCreateResourceID(clusterIdentifier, scheduleIdentifier)
	input := &redshift.ModifyClusterSnapshotScheduleInput{
		ClusterIdentifier:    aws.String(clusterIdentifier),
		ScheduleIdentifier:   aws.String(scheduleIdentifier),
		DisassociateSchedule: aws.Bool(false),
	}

	_, err := conn.ModifyClusterSnapshotSchedule(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Snapshot Schedule Association (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitSnapshotScheduleAssociationCreated(ctx, conn, clusterIdentifier, scheduleIdentifier); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Snapshot Schedule Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceSnapshotScheduleAssociationRead(ctx, d, meta)...)
}

func resourceSnapshotScheduleAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	clusterIdentifier, scheduleIdentifier, err := SnapshotScheduleAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	association, err := findSnapshotScheduleAssociationByTwoPartKey(ctx, conn, clusterIdentifier, scheduleIdentifier)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Snapshot Schedule Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Snapshot Schedule Association (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrClusterIdentifier, association.ClusterIdentifier)
	d.Set("schedule_identifier", scheduleIdentifier)

	return diags
}

func resourceSnapshotScheduleAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	clusterIdentifier, scheduleIdentifier, err := SnapshotScheduleAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Redshift Snapshot Schedule Association: %s", d.Id())
	_, err = conn.ModifyClusterSnapshotSchedule(ctx, &redshift.ModifyClusterSnapshotScheduleInput{
		ClusterIdentifier:    aws.String(clusterIdentifier),
		ScheduleIdentifier:   aws.String(scheduleIdentifier),
		DisassociateSchedule: aws.Bool(true),
	})
	if errs.IsA[*awstypes.ClusterNotFoundFault](err) || errs.IsA[*awstypes.SnapshotScheduleNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Snapshot Schedule Association (%s): %s", d.Id(), err)
	}

	if _, err := waitSnapshotScheduleAssociationDeleted(ctx, conn, clusterIdentifier, scheduleIdentifier); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Snapshot Schedule Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const snapshotScheduleAssociationIDSeparator = "/"

func SnapshotScheduleAssociationCreateResourceID(clusterIdentifier, scheduleIdentifier string) string {
	parts := []string{clusterIdentifier, scheduleIdentifier}
	id := strings.Join(parts, snapshotScheduleAssociationIDSeparator)

	return id
}

func SnapshotScheduleAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, snapshotScheduleAssociationIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected <ClusterIdentifier>%[2]s<ScheduleIdentifier>", id, snapshotScheduleAssociationIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findSnapshotScheduleAssociationByTwoPartKey(ctx context.Context, conn *redshift.Client, clusterIdentifier, scheduleIdentifier string) (*awstypes.ClusterAssociatedToSchedule, error) {
	input := &redshift.DescribeSnapshotSchedulesInput{
		ClusterIdentifier:  aws.String(clusterIdentifier),
		ScheduleIdentifier: aws.String(scheduleIdentifier),
	}

	output, err := conn.DescribeSnapshotSchedules(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	schedule, err := tfresource.AssertSingleValueResult(output.SnapshotSchedules)

	if err != nil {
		return nil, err
	}

	for _, v := range schedule.AssociatedClusters {
		if aws.ToString(v.ClusterIdentifier) == clusterIdentifier {
			return &v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func statusSnapshotScheduleAssociation(ctx context.Context, conn *redshift.Client, clusterIdentifier, scheduleIdentifier string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findSnapshotScheduleAssociationByTwoPartKey(ctx, conn, clusterIdentifier, scheduleIdentifier)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ScheduleAssociationState), nil
	}
}

func waitSnapshotScheduleAssociationCreated(ctx context.Context, conn *redshift.Client, clusterIdentifier, scheduleIdentifier string) (*awstypes.ClusterAssociatedToSchedule, error) {
	const (
		timeout = 75 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ScheduleStateModifying),
		Target:     enum.Slice(awstypes.ScheduleStateActive),
		Refresh:    statusSnapshotScheduleAssociation(ctx, conn, clusterIdentifier, scheduleIdentifier),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ClusterAssociatedToSchedule); ok {
		return output, err
	}

	return nil, err
}

func waitSnapshotScheduleAssociationDeleted(ctx context.Context, conn *redshift.Client, clusterIdentifier, scheduleIdentifier string) (*awstypes.ClusterAssociatedToSchedule, error) { //nolint:unparam
	const (
		timeout = 75 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ScheduleStateModifying, awstypes.ScheduleStateActive),
		Target:     []string{},
		Refresh:    statusSnapshotScheduleAssociation(ctx, conn, clusterIdentifier, scheduleIdentifier),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ClusterAssociatedToSchedule); ok {
		return output, err
	}

	return nil, err
}
