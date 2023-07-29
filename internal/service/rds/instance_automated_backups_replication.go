// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	InstanceAutomatedBackupsReplicationCreateTimeout = 75 * time.Minute
	InstanceAutomatedBackupsReplicationDeleteTimeout = 75 * time.Minute
)

// @SDKResource("aws_db_instance_automated_backups_replication")
func ResourceInstanceAutomatedBackupsReplication() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInstanceAutomatedBackupsReplicationCreate,
		ReadWithoutTimeout:   resourceInstanceAutomatedBackupsReplicationRead,
		DeleteWithoutTimeout: resourceInstanceAutomatedBackupsReplicationDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(InstanceAutomatedBackupsReplicationCreateTimeout),
			Delete: schema.DefaultTimeout(InstanceAutomatedBackupsReplicationDeleteTimeout),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"kms_key_id": {
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
			"retention_period": {
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
		BackupRetentionPeriod: aws.Int64(int64(d.Get("retention_period").(int))),
		SourceDBInstanceArn:   aws.String(d.Get("source_db_instance_arn").(string)),
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("pre_signed_url"); ok {
		input.PreSignedUrl = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Starting RDS instance automated backups replication: %s", input)
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

	d.Set("kms_key_id", backup.KmsKeyId)
	d.Set("retention_period", backup.BackupRetentionPeriod)
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
		return sdkdiag.AppendErrorf(diags, "deleting RDS DB Instance Automated Backup (%s): %s", d.Id(), err)
	}

	dbInstanceID := aws.StringValue(backup.DBInstanceIdentifier)
	sourceDatabaseARN, err := arn.Parse(aws.StringValue(backup.DBInstanceArn))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS DB Instance Automated Backup (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Stopping RDS Instance Automated Backups Replication: %s", d.Id())
	_, err = conn.StopDBInstanceAutomatedBackupsReplicationWithContext(ctx, &rds.StopDBInstanceAutomatedBackupsReplicationInput{
		SourceDBInstanceArn: aws.String(d.Get("source_db_instance_arn").(string)),
	})

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBInstanceNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS DB Instance Automated Backup (%s): %s", d.Id(), err)
	}

	// Create a new client to the source region.
	sourceDatabaseConn := conn
	if sourceDatabaseARN.Region != meta.(*conns.AWSClient).Region {
		sourceDatabaseConn = rds.New(meta.(*conns.AWSClient).Session, aws.NewConfig().WithRegion(sourceDatabaseARN.Region))
	}

	if _, err := waitDBInstanceAutomatedBackupDeleted(ctx, sourceDatabaseConn, dbInstanceID, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS DB Instance Automated Backup (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}
