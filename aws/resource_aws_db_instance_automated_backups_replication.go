package aws

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/rds/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/rds/waiter"
)

func resourceAwsDbInstanceAutomatedBackupsReplication() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsDbInstanceAutomatedBackupsReplicationCreate,
		ReadContext:   resourceAwsDbInstanceAutomatedBackupsReplicationRead,
		DeleteContext: resourceAwsDbInstanceAutomatedBackupsReplicationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"backup_retention_period": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  1,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"pre_signed_url": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"source_db_instance_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
		},
	}
}

func resourceAwsDbInstanceAutomatedBackupsReplicationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).rdsconn

	input := &rds.StartDBInstanceAutomatedBackupsReplicationInput{
		SourceDBInstanceArn: aws.String(d.Get("source_db_instance_arn").(string)),
	}

	if attr, ok := d.GetOk("backup_retention_period"); ok {
		input.BackupRetentionPeriod = aws.Int64(int64(attr.(int)))
	}

	if attr, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(attr.(string))
	}

	if attr, ok := d.GetOk("pre_signed_url"); ok {
		input.PreSignedUrl = aws.String(attr.(string))
	}

	log.Printf("[DEBUG] RDS DB Instance Start Automated Backups Replication: (%s)", input)
	output, err := conn.StartDBInstanceAutomatedBackupsReplicationWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to Start Automated Backup Replication: %s", err))
	}

	if output == nil || output.DBInstanceAutomatedBackup == nil {
		return diag.FromErr(fmt.Errorf("error starting RDS DB Instance Start Automated Backups Replication: empty output"))
	}

	dbInstanceAutomatedBackupsArn := aws.StringValue(output.DBInstanceAutomatedBackup.DBInstanceAutomatedBackupsArn)
	d.SetId(dbInstanceAutomatedBackupsArn)

	if _, err := waiter.DBInstanceAutomatedBackupsReplicationStarted(ctx, conn, dbInstanceAutomatedBackupsArn); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting to RDS DB Instance Start Automated Backups Replication: %s", err))
	}

	return resourceAwsDbInstanceAutomatedBackupsReplicationRead(ctx, d, meta)
}

func resourceAwsDbInstanceAutomatedBackupsReplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).rdsconn

	dbInstanceAutomatedBackup, err := finder.DBInstanceAutomatedBackup(ctx, conn, d.Id())
	if isAWSErr(err, rds.ErrCodeDBInstanceAutomatedBackupNotFoundFault, "") {
		log.Printf("[WARN] RDS DB Instance Automated Backups Replication not found (%s), removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error describe RDS DB Instance Automated Backups Replication: %s", err))
	}

	if err := d.Set("backup_retention_period", dbInstanceAutomatedBackup.BackupRetentionPeriod); err != nil {
		return diag.FromErr(fmt.Errorf("error setting backup retention period for RDS DB Instance: %s", err))
	}
	if err := d.Set("kms_key_id", dbInstanceAutomatedBackup.KmsKeyId); err != nil {
		return diag.FromErr(fmt.Errorf("error setting kms key id for RDS DB Instance: %s", err))
	}
	if err := d.Set("source_db_instance_arn", dbInstanceAutomatedBackup.DBInstanceArn); err != nil {
		return diag.FromErr(fmt.Errorf("error setting source db instance arn for RDS DB Instance: (%s)", err))
	}

	return nil
}

func resourceAwsDbInstanceAutomatedBackupsReplicationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).rdsconn

	input := &rds.DeleteDBInstanceAutomatedBackupInput{
		DBInstanceAutomatedBackupsArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Delete RDS DB Instance Automated Backups Replication: %s", d.Id())
	_, err := conn.DeleteDBInstanceAutomatedBackupWithContext(ctx, input)

	if isAWSErr(err, rds.ErrCodeDBInstanceAutomatedBackupNotFoundFault, "") {
		return nil
	}

	if isAWSErr(err, rds.ErrCodeInvalidDBInstanceAutomatedBackupStateFault, "") {
		return nil
	}

	if isAWSErr(err, rds.ErrCodeInvalidDBInstanceStateFault, "") {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error delete RDS DB Instance Automated Backups Replication:  %s", err))
	}

	if _, err = waiter.DBInstanceAutomatedBackupsReplicationDeleted(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting to delete RDS DB Instance Automated Backups Replication: %s", err))
	}

	return nil
}
