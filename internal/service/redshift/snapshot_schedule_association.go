// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package redshift

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
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
	id := snapshotScheduleAssociationCreateResourceID(clusterIdentifier, scheduleIdentifier)
	input := redshift.ModifyClusterSnapshotScheduleInput{
		ClusterIdentifier:    aws.String(clusterIdentifier),
		DisassociateSchedule: aws.Bool(false),
		ScheduleIdentifier:   aws.String(scheduleIdentifier),
	}

	_, err := conn.ModifyClusterSnapshotSchedule(ctx, &input)

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

	clusterIdentifier, scheduleIdentifier, err := snapshotScheduleAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	association, err := findSnapshotScheduleAssociationByTwoPartKey(ctx, conn, clusterIdentifier, scheduleIdentifier)

	if !d.IsNewResource() && retry.NotFound(err) {
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

	clusterIdentifier, scheduleIdentifier, err := snapshotScheduleAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Redshift Snapshot Schedule Association: %s", d.Id())
	input := redshift.ModifyClusterSnapshotScheduleInput{
		ClusterIdentifier:    aws.String(clusterIdentifier),
		DisassociateSchedule: aws.Bool(true),
		ScheduleIdentifier:   aws.String(scheduleIdentifier),
	}
	_, err = conn.ModifyClusterSnapshotSchedule(ctx, &input)

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

func snapshotScheduleAssociationCreateResourceID(clusterIdentifier, scheduleIdentifier string) string {
	parts := []string{clusterIdentifier, scheduleIdentifier}
	id := strings.Join(parts, snapshotScheduleAssociationIDSeparator)

	return id
}

func snapshotScheduleAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, snapshotScheduleAssociationIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected <ClusterIdentifier>%[2]s<ScheduleIdentifier>", id, snapshotScheduleAssociationIDSeparator)
	}

	return parts[0], parts[1], nil
}
