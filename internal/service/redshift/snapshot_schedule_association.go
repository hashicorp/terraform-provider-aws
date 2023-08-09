// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"errors"
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
)

// @SDKResource("aws_redshift_snapshot_schedule_association")
func ResourceSnapshotScheduleAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSnapshotScheduleAssociationCreate,
		ReadWithoutTimeout:   resourceSnapshotScheduleAssociationRead,
		DeleteWithoutTimeout: resourceSnapshotScheduleAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"cluster_identifier": {
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
	clusterIdentifier := d.Get("cluster_identifier").(string)
	scheduleIdentifier := d.Get("schedule_identifier").(string)

	_, err := conn.ModifyClusterSnapshotScheduleWithContext(ctx, &redshift.ModifyClusterSnapshotScheduleInput{
		ClusterIdentifier:    aws.String(clusterIdentifier),
		ScheduleIdentifier:   aws.String(scheduleIdentifier),
		DisassociateSchedule: aws.Bool(false),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Cluster Snapshot Schedule (%s/%s): %s", clusterIdentifier, scheduleIdentifier, err)
	}

	d.SetId(fmt.Sprintf("%s/%s", clusterIdentifier, scheduleIdentifier))

	if _, err := WaitScheduleAssociationActive(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Cluster Snapshot Schedule (%s): waiting for completion: %s", d.Id(), err)
	}

	return append(diags, resourceSnapshotScheduleAssociationRead(ctx, d, meta)...)
}

func resourceSnapshotScheduleAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	scheduleIdentifier, assoicatedCluster, err := FindScheduleAssociationById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Schedule Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Schedule Association (%s): %s", d.Id(), err)
	}
	d.Set("cluster_identifier", assoicatedCluster.ClusterIdentifier)
	d.Set("schedule_identifier", scheduleIdentifier)

	return diags
}

func resourceSnapshotScheduleAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)
	clusterIdentifier, scheduleIdentifier, err := SnapshotScheduleAssociationParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Cluster Snapshot Schedule (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting Redshift Cluster Snapshot Schedule Association: %s", d.Id())
	_, err = conn.ModifyClusterSnapshotScheduleWithContext(ctx, &redshift.ModifyClusterSnapshotScheduleInput{
		ClusterIdentifier:    aws.String(clusterIdentifier),
		ScheduleIdentifier:   aws.String(scheduleIdentifier),
		DisassociateSchedule: aws.Bool(true),
	})

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeClusterNotFoundFault, redshift.ErrCodeSnapshotScheduleNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Cluster Snapshot Schedule (%s): %s", d.Id(), err)
	}

	if _, err := waitScheduleAssociationDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Cluster Snapshot Schedule (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}

func SnapshotScheduleAssociationParseID(id string) (clusterIdentifier, scheduleIdentifier string, err error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		err = fmt.Errorf("aws_redshift_snapshot_schedule_association id must be of the form <ClusterIdentifier>/<ScheduleIdentifier>")
		return
	}

	clusterIdentifier = parts[0]
	scheduleIdentifier = parts[1]
	return
}

func FindScheduleAssociationById(ctx context.Context, conn *redshift.Redshift, id string) (string, *redshift.ClusterAssociatedToSchedule, error) {
	clusterIdentifier, scheduleIdentifier, err := SnapshotScheduleAssociationParseID(id)
	if err != nil {
		return "", nil, fmt.Errorf("parsing Redshift Cluster Snapshot Schedule Association ID %s: %s", id, err)
	}

	input := &redshift.DescribeSnapshotSchedulesInput{
		ClusterIdentifier:  aws.String(clusterIdentifier),
		ScheduleIdentifier: aws.String(scheduleIdentifier),
	}
	resp, err := conn.DescribeSnapshotSchedulesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeSnapshotScheduleNotFoundFault) {
		return "", nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return "", nil, err
	}

	if resp.SnapshotSchedules == nil || len(resp.SnapshotSchedules) == 0 {
		return "", nil, tfresource.NewEmptyResultError(input)
	}

	snapshotSchedule := resp.SnapshotSchedules[0]

	if snapshotSchedule == nil {
		return "", nil, tfresource.NewEmptyResultError(input)
	}

	var associatedCluster *redshift.ClusterAssociatedToSchedule
	for _, cluster := range snapshotSchedule.AssociatedClusters {
		if aws.StringValue(cluster.ClusterIdentifier) == clusterIdentifier {
			associatedCluster = cluster
			break
		}
	}

	if associatedCluster == nil {
		return "", nil, tfresource.NewEmptyResultError(input)
	}

	return aws.StringValue(snapshotSchedule.ScheduleIdentifier), associatedCluster, nil
}

func statusScheduleAssociation(ctx context.Context, conn *redshift.Redshift, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		_, output, err := FindScheduleAssociationById(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ScheduleAssociationState), nil
	}
}

const (
	snapshotScheduleAssociationActivatedTimeout = 75 * time.Minute
	snapshotScheduleAssociationDestroyedTimeout = 75 * time.Minute
)

func WaitScheduleAssociationActive(ctx context.Context, conn *redshift.Redshift, id string) (*redshift.ClusterAssociatedToSchedule, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{redshift.ScheduleStateModifying},
		Target:     []string{redshift.ScheduleStateActive},
		Refresh:    statusScheduleAssociation(ctx, conn, id),
		Timeout:    snapshotScheduleAssociationActivatedTimeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*redshift.ClusterAssociatedToSchedule); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.ScheduleAssociationState)))

		return output, err
	}

	return nil, err
}

func waitScheduleAssociationDeleted(ctx context.Context, conn *redshift.Redshift, id string) (*redshift.ClusterAssociatedToSchedule, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    []string{redshift.ScheduleStateModifying, redshift.ScheduleStateActive},
		Target:     []string{},
		Refresh:    statusScheduleAssociation(ctx, conn, id),
		Timeout:    snapshotScheduleAssociationDestroyedTimeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*redshift.ClusterAssociatedToSchedule); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.ScheduleAssociationState)))
		return output, err
	}

	return nil, err
}
