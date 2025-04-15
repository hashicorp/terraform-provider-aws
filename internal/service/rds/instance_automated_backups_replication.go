// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// AWS flip-flop on the capitalization of status codes. Use uppercase.
const (
	instanceAutomatedBackupStatusPending     = "PENDING"
	instanceAutomatedBackupStatusReplicating = "REPLICATING"
	instanceAutomatedBackupStatusRetained    = "RETAINED"
)

// @SDKResource("aws_db_instance_automated_backups_replication", name="Instance Automated Backups Replication")
func resourceInstanceAutomatedBackupsReplication() *schema.Resource {
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

func resourceInstanceAutomatedBackupsReplicationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	sourceDBInstanceARN := d.Get("source_db_instance_arn").(string)
	input := &rds.StartDBInstanceAutomatedBackupsReplicationInput{
		BackupRetentionPeriod: aws.Int32(int32(d.Get(names.AttrRetentionPeriod).(int))),
		SourceDBInstanceArn:   aws.String(sourceDBInstanceARN),
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("pre_signed_url"); ok {
		input.PreSignedUrl = aws.String(v.(string))
	}

	output, err := conn.StartDBInstanceAutomatedBackupsReplication(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "starting RDS Instance Automated Backups Replication (%s): %s", sourceDBInstanceARN, err)
	}

	d.SetId(aws.ToString(output.DBInstanceAutomatedBackup.DBInstanceAutomatedBackupsArn))

	if _, err := waitDBInstanceAutomatedBackupCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Instance Automated Backup (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceInstanceAutomatedBackupsReplicationRead(ctx, d, meta)...)
}

func resourceInstanceAutomatedBackupsReplicationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	backup, err := findDBInstanceAutomatedBackupByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS DB Instance Automated Backup %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Instance Automated Backup (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrKMSKeyID, backup.KmsKeyId)
	d.Set(names.AttrRetentionPeriod, backup.BackupRetentionPeriod)
	d.Set("source_db_instance_arn", backup.DBInstanceArn)

	return diags
}

func resourceInstanceAutomatedBackupsReplicationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	backup, err := findDBInstanceAutomatedBackupByARN(ctx, conn, d.Id())

	switch {
	case tfresource.NotFound(err):
		return diags
	case err != nil:
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Instance Automated Backup (%s): %s", d.Id(), err)
	}

	dbInstanceID := aws.ToString(backup.DBInstanceIdentifier)
	sourceDatabaseARN, err := arn.Parse(aws.ToString(backup.DBInstanceArn))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Stopping RDS Instance Automated Backups Replication: %s", d.Id())
	sourceDBInstanceARN := d.Get("source_db_instance_arn").(string)
	_, err = conn.StopDBInstanceAutomatedBackupsReplication(ctx, &rds.StopDBInstanceAutomatedBackupsReplicationInput{
		SourceDBInstanceArn: aws.String(sourceDBInstanceARN),
	})

	if errs.IsA[*types.DBInstanceNotFoundFault](err) {
		return diags
	}

	if errs.IsAErrorMessageContains[*types.InvalidDBInstanceStateFault](err, "not replicating to the current region") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "stopping RDS Instance Automated Backups Replication (%s): %s", sourceDBInstanceARN, err)
	}

	// Make API calls in the source Region.
	optFn := func(o *rds.Options) {
		o.Region = sourceDatabaseARN.Region
	}

	if _, err := waitDBInstanceAutomatedBackupDeleted(ctx, conn, dbInstanceID, d.Id(), d.Timeout(schema.TimeoutCreate), optFn); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Instance Automated Backup (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findDBInstanceAutomatedBackupByARN(ctx context.Context, conn *rds.Client, arn string) (*types.DBInstanceAutomatedBackup, error) {
	input := &rds.DescribeDBInstanceAutomatedBackupsInput{
		DBInstanceAutomatedBackupsArn: aws.String(arn),
	}
	output, err := findDBInstanceAutomatedBackup(ctx, conn, input, tfslices.PredicateTrue[*types.DBInstanceAutomatedBackup]())

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.DBInstanceAutomatedBackupsArn) != arn {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	// AWS flip-flop on the capitalization of status codes. Case-insensitive comparison.
	if status := aws.ToString(output.Status); strings.EqualFold(status, instanceAutomatedBackupStatusRetained) {
		// If the automated backup is retained, the replication is stopped.
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output, nil
}

func findDBInstanceAutomatedBackup(ctx context.Context, conn *rds.Client, input *rds.DescribeDBInstanceAutomatedBackupsInput, filter tfslices.Predicate[*types.DBInstanceAutomatedBackup]) (*types.DBInstanceAutomatedBackup, error) {
	output, err := findDBInstanceAutomatedBackups(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDBInstanceAutomatedBackups(ctx context.Context, conn *rds.Client, input *rds.DescribeDBInstanceAutomatedBackupsInput, filter tfslices.Predicate[*types.DBInstanceAutomatedBackup]) ([]types.DBInstanceAutomatedBackup, error) {
	var output []types.DBInstanceAutomatedBackup

	pages := rds.NewDescribeDBInstanceAutomatedBackupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.DBInstanceAutomatedBackupNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.DBInstanceAutomatedBackups {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusDBInstanceAutomatedBackup(ctx context.Context, conn *rds.Client, arn string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findDBInstanceAutomatedBackupByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		// AWS flip-flop on the capitalization of status codes. Convert to uppercase.
		return output, strings.ToUpper(aws.ToString(output.Status)), nil
	}
}

func waitDBInstanceAutomatedBackupCreated(ctx context.Context, conn *rds.Client, arn string, timeout time.Duration) (*types.DBInstanceAutomatedBackup, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{instanceAutomatedBackupStatusPending},
		Target:  []string{instanceAutomatedBackupStatusReplicating},
		Refresh: statusDBInstanceAutomatedBackup(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DBInstanceAutomatedBackup); ok {
		return output, err
	}

	return nil, err
}

func waitDBInstanceAutomatedBackupDeleted(ctx context.Context, conn *rds.Client, dbInstanceID, dbInstanceAutomatedBackupsARN string, timeout time.Duration, optFns ...func(*rds.Options)) (*types.DBInstance, error) {
	var output *types.DBInstance

	_, err := tfresource.RetryUntilEqual(ctx, timeout, false, func() (bool, error) {
		dbInstance, err := findDBInstanceByID(ctx, conn, dbInstanceID, optFns...)

		if tfresource.NotFound(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		output = dbInstance

		return slices.ContainsFunc(dbInstance.DBInstanceAutomatedBackupsReplications, func(v types.DBInstanceAutomatedBackupsReplication) bool {
			return aws.ToString(v.DBInstanceAutomatedBackupsArn) == dbInstanceAutomatedBackupsARN
		}), nil
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
