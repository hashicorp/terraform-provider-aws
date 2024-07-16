// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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

func resourceSnapshotScheduleAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	clusterIdentifier := d.Get(names.AttrClusterIdentifier).(string)
	scheduleIdentifier := d.Get("schedule_identifier").(string)
	id := SnapshotScheduleAssociationCreateResourceID(clusterIdentifier, scheduleIdentifier)
	input := &redshift.ModifyClusterSnapshotScheduleInput{
		ClusterIdentifier:    aws.String(clusterIdentifier),
		ScheduleIdentifier:   aws.String(scheduleIdentifier),
		DisassociateSchedule: aws.Bool(false),
	}

	_, err := conn.ModifyClusterSnapshotScheduleWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Snapshot Schedule Association (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitSnapshotScheduleAssociationCreated(ctx, conn, clusterIdentifier, scheduleIdentifier); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Snapshot Schedule Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceSnapshotScheduleAssociationRead(ctx, d, meta)...)
}

func resourceSnapshotScheduleAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

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

func resourceSnapshotScheduleAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	clusterIdentifier, scheduleIdentifier, err := SnapshotScheduleAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Redshift Snapshot Schedule Association: %s", d.Id())
	_, err = conn.ModifyClusterSnapshotScheduleWithContext(ctx, &redshift.ModifyClusterSnapshotScheduleInput{
		ClusterIdentifier:    aws.String(clusterIdentifier),
		ScheduleIdentifier:   aws.String(scheduleIdentifier),
		DisassociateSchedule: aws.Bool(true),
	})

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeClusterNotFoundFault, redshift.ErrCodeSnapshotScheduleNotFoundFault) {
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

func findSnapshotScheduleAssociationByTwoPartKey(ctx context.Context, conn *redshift.Redshift, clusterIdentifier, scheduleIdentifier string) (*redshift.ClusterAssociatedToSchedule, error) {
	input := &redshift.DescribeSnapshotSchedulesInput{
		ClusterIdentifier:  aws.String(clusterIdentifier),
		ScheduleIdentifier: aws.String(scheduleIdentifier),
	}

	output, err := conn.DescribeSnapshotSchedulesWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	schedule, err := tfresource.AssertSinglePtrResult(output.SnapshotSchedules)

	if err != nil {
		return nil, err
	}

	for _, v := range schedule.AssociatedClusters {
		if aws.StringValue(v.ClusterIdentifier) == clusterIdentifier {
			return v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func statusSnapshotScheduleAssociation(ctx context.Context, conn *redshift.Redshift, clusterIdentifier, scheduleIdentifier string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findSnapshotScheduleAssociationByTwoPartKey(ctx, conn, clusterIdentifier, scheduleIdentifier)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ScheduleAssociationState), nil
	}
}

func waitSnapshotScheduleAssociationCreated(ctx context.Context, conn *redshift.Redshift, clusterIdentifier, scheduleIdentifier string) (*redshift.ClusterAssociatedToSchedule, error) {
	const (
		timeout = 75 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:    []string{redshift.ScheduleStateModifying},
		Target:     []string{redshift.ScheduleStateActive},
		Refresh:    statusSnapshotScheduleAssociation(ctx, conn, clusterIdentifier, scheduleIdentifier),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*redshift.ClusterAssociatedToSchedule); ok {
		return output, err
	}

	return nil, err
}

func waitSnapshotScheduleAssociationDeleted(ctx context.Context, conn *redshift.Redshift, clusterIdentifier, scheduleIdentifier string) (*redshift.ClusterAssociatedToSchedule, error) { //nolint:unparam
	const (
		timeout = 75 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:    []string{redshift.ScheduleStateModifying, redshift.ScheduleStateActive},
		Target:     []string{},
		Refresh:    statusSnapshotScheduleAssociation(ctx, conn, clusterIdentifier, scheduleIdentifier),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*redshift.ClusterAssociatedToSchedule); ok {
		return output, err
	}

	return nil, err
}
