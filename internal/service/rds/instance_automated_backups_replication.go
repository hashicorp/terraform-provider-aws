// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// AWS flip-flop on the capitalization of status codes. Use uppercase.
const (
	InstanceAutomatedBackupStatusPending     = "PENDING"
	InstanceAutomatedBackupStatusReplicating = "REPLICATING"
	InstanceAutomatedBackupStatusRetained    = "RETAINED"
)

// @SDKResource("aws_db_instance_automated_backups_replication")
func ResourceInstanceAutomatedBackupsReplication() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInstanceAutomatedBackupsReplicationCreate,
		ReadWithoutTimeout:   resourceInstanceAutomatedBackupsReplicationRead,
		DeleteWithoutTimeout: resourceInstanceAutomatedBackupsReplicationDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(75 * time.Minute),
			Delete: schema.DefaultTimeout(75 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrKMSKeyID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"pre_signed_url": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrRetentionPeriod: {
				Type:     schema.TypeInt,
				ForceNew: true,
				Optional: true,
				Default:  7,
			},
			"source_db_instance_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceInstanceAutomatedBackupsReplicationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	input := &rds.StartDBInstanceAutomatedBackupsReplicationInput{
		BackupRetentionPeriod: aws.Int64(int64(d.Get(names.AttrRetentionPeriod).(int))),
		SourceDBInstanceArn:   aws.String(d.Get("source_db_instance_arn").(string)),
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("pre_signed_url"); ok {
		input.PreSignedUrl = aws.String(v.(string))
	}

	output, err := conn.StartDBInstanceAutomatedBackupsReplicationWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "starting RDS instance automated backups replication: %s", err)
	}

	d.SetId(aws.StringValue(output.DBInstanceAutomatedBackup.DBInstanceAutomatedBackupsArn))

	if _, err := waitDBInstanceAutomatedBackupCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DB instance automated backup (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceInstanceAutomatedBackupsReplicationRead(ctx, d, meta)...)
}

func resourceInstanceAutomatedBackupsReplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	backup, err := FindDBInstanceAutomatedBackupByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS instance automated backup %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS instance automated backup (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrKMSKeyID, backup.KmsKeyId)
	d.Set(names.AttrRetentionPeriod, backup.BackupRetentionPeriod)
	d.Set("source_db_instance_arn", backup.DBInstanceArn)

	return diags
}

func resourceInstanceAutomatedBackupsReplicationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	backup, err := FindDBInstanceAutomatedBackupByARN(ctx, conn, d.Id())

	if tfresource.NotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS instance automated backup (%s): %s", d.Id(), err)
	}

	dbInstanceID := aws.StringValue(backup.DBInstanceIdentifier)
	sourceDatabaseARN, err := arn.Parse(aws.StringValue(backup.DBInstanceArn))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Stopping RDS Instance Automated Backups Replication: %s", d.Id())
	_, err = conn.StopDBInstanceAutomatedBackupsReplicationWithContext(ctx, &rds.StopDBInstanceAutomatedBackupsReplicationInput{
		SourceDBInstanceArn: aws.String(d.Get("source_db_instance_arn").(string)),
	})

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBInstanceNotFoundFault) {
		return diags
	}

	if tfawserr.ErrMessageContains(err, rds.ErrCodeInvalidDBInstanceStateFault, "not replicating to the current region") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS DB Instance Automated Backup (%s): %s", d.Id(), err)
	}

	// Create a new client to the source region.
	sourceDatabaseConn := meta.(*conns.AWSClient).RDSConnForRegion(ctx, sourceDatabaseARN.Region)

	if _, err := waitDBInstanceAutomatedBackupDeleted(ctx, sourceDatabaseConn, dbInstanceID, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DB instance automated backup (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindDBInstanceAutomatedBackupByARN(ctx context.Context, conn *rds.RDS, arn string) (*rds.DBInstanceAutomatedBackup, error) {
	input := &rds.DescribeDBInstanceAutomatedBackupsInput{
		DBInstanceAutomatedBackupsArn: aws.String(arn),
	}
	output, err := findDBInstanceAutomatedBackup(ctx, conn, input, tfslices.PredicateTrue[*rds.DBInstanceAutomatedBackup]())

	if err != nil {
		return nil, err
	}

	// AWS flip-flop on the capitalization of status codes. Case-insensitive comparison.
	if status := aws.StringValue(output.Status); strings.EqualFold(status, InstanceAutomatedBackupStatusRetained) {
		// If the automated backup is retained, the replication is stopped.
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.DBInstanceAutomatedBackupsArn) != arn {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findDBInstanceAutomatedBackup(ctx context.Context, conn *rds.RDS, input *rds.DescribeDBInstanceAutomatedBackupsInput, filter tfslices.Predicate[*rds.DBInstanceAutomatedBackup]) (*rds.DBInstanceAutomatedBackup, error) {
	output, err := findDBInstanceAutomatedBackups(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findDBInstanceAutomatedBackups(ctx context.Context, conn *rds.RDS, input *rds.DescribeDBInstanceAutomatedBackupsInput, filter tfslices.Predicate[*rds.DBInstanceAutomatedBackup]) ([]*rds.DBInstanceAutomatedBackup, error) {
	var output []*rds.DBInstanceAutomatedBackup

	err := conn.DescribeDBInstanceAutomatedBackupsPagesWithContext(ctx, input, func(page *rds.DescribeDBInstanceAutomatedBackupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBInstanceAutomatedBackups {
			if v != nil && filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBInstanceAutomatedBackupNotFoundFault) {
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

func statusDBInstanceAutomatedBackup(ctx context.Context, conn *rds.RDS, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDBInstanceAutomatedBackupByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		// AWS flip-flop on the capitalization of status codes. Convert to uppercase.
		return output, strings.ToUpper(aws.StringValue(output.Status)), nil
	}
}

func waitDBInstanceAutomatedBackupCreated(ctx context.Context, conn *rds.RDS, arn string, timeout time.Duration) (*rds.DBInstanceAutomatedBackup, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{InstanceAutomatedBackupStatusPending},
		Target:  []string{InstanceAutomatedBackupStatusReplicating},
		Refresh: statusDBInstanceAutomatedBackup(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*rds.DBInstanceAutomatedBackup); ok {
		return output, err
	}

	return nil, err
}

// statusDBInstanceHasAutomatedBackup returns whether or not a database instance has a specified automated backup.
// The connection must be valid for the database instance's Region.
func statusDBInstanceHasAutomatedBackup(ctx context.Context, conn *rds.RDS, dbInstanceID, dbInstanceAutomatedBackupsARN string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDBInstanceByIDSDKv1(ctx, conn, dbInstanceID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		for _, v := range output.DBInstanceAutomatedBackupsReplications {
			if aws.StringValue(v.DBInstanceAutomatedBackupsArn) == dbInstanceAutomatedBackupsARN {
				return output, strconv.FormatBool(true), nil
			}
		}

		return output, strconv.FormatBool(false), nil
	}
}

// waitDBInstanceAutomatedBackupDeleted waits for a specified automated backup to be deleted from a database instance.
// The connection must be valid for the database instance's Region.
func waitDBInstanceAutomatedBackupDeleted(ctx context.Context, conn *rds.RDS, dbInstanceID, dbInstanceAutomatedBackupsARN string, timeout time.Duration) (*rds.DBInstance, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{strconv.FormatBool(true)},
		Target:  []string{strconv.FormatBool(false)},
		Refresh: statusDBInstanceHasAutomatedBackup(ctx, conn, dbInstanceID, dbInstanceAutomatedBackupsARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*rds.DBInstance); ok {
		return output, err
	}

	return nil, err
}
