package rds

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	dbInstanceAutomatedBackupReplicationRetained = "retained"
)

func ResourceInstanceAutomatedBackupReplication() *schema.Resource {
	return &schema.Resource{
		Create: resourceInstanceAutomatedBackupReplicationCreate,
		Read:   resourceInstanceAutomatedBackupReplicationRead,
		Delete: resourceInstanceAutomatedBackupReplicationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
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

func resourceInstanceAutomatedBackupReplicationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn

	input := &rds.StartDBInstanceAutomatedBackupsReplicationInput{
		SourceDBInstanceArn:   aws.String(d.Get("source_db_instance_arn").(string)),
		BackupRetentionPeriod: aws.Int64(int64(d.Get("retention_period").(int))),
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Starting RDS instance automated backup replication for: %s", *input.SourceDBInstanceArn)

	output, err := conn.StartDBInstanceAutomatedBackupsReplication(input)

	if err != nil {
		return fmt.Errorf("error creating RDS instance automated backup replication: %s", err)
	}

	d.SetId(aws.StringValue(output.DBInstanceAutomatedBackup.DBInstanceAutomatedBackupsArn))

	if _, err := waitDBInstanceAutomatedBackupAvailable(conn, d.Id(), d.Timeout(schema.TimeoutDefault)); err != nil {
		return fmt.Errorf("error waiting for DB instance automated backup (%s) creation: %w", d.Id(), err)
	}

	return resourceInstanceAutomatedBackupReplicationRead(d, meta)
}

func resourceInstanceAutomatedBackupReplicationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn

	input := &rds.DescribeDBInstanceAutomatedBackupsInput{
		DBInstanceAutomatedBackupsArn: aws.String(d.Id()),
	}

	output, err := conn.DescribeDBInstanceAutomatedBackups(input)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBInstanceAutomatedBackupNotFoundFault) {
		log.Printf("[WARN] RDS instance automated backup replication not found, removing from state: %s", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading RDS instance automated backup replication: %s", err)
	}

	for _, backup := range output.DBInstanceAutomatedBackups {
		if aws.StringValue(backup.DBInstanceAutomatedBackupsArn) == d.Id() {
			// Check if the automated backup is retained
			if aws.StringValue(backup.Status) == dbInstanceAutomatedBackupReplicationRetained {
				log.Printf("[WARN] RDS instance automated backup replication is retained, removing from state: %s", d.Id())
				d.SetId("")
				return nil // If the automated backup is retained, the replication is stopped.
			} else {
				d.Set("source_db_instance_arn", backup.DBInstanceArn)
				d.Set("kms_key_id", backup.KmsKeyId)
				d.Set("retention_period", backup.BackupRetentionPeriod)
			}

		} else {
			return fmt.Errorf("unable to find RDS instance automated backup replication: %s", d.Id())
		}
	}

	return nil
}

func resourceInstanceAutomatedBackupReplicationDelete(d *schema.ResourceData, meta interface{}) error {
	var sourceDatabaseRegion string
	var databaseIdentifier string

	conn := meta.(*conns.AWSClient).RDSConn

	describeInput := &rds.DescribeDBInstanceAutomatedBackupsInput{
		DBInstanceAutomatedBackupsArn: aws.String(d.Id()),
	}

	describeOutput, err := conn.DescribeDBInstanceAutomatedBackups(describeInput)

	// Get and set the region of the source database and database identifier
	for _, backup := range describeOutput.DBInstanceAutomatedBackups {
		if aws.StringValue(backup.DBInstanceAutomatedBackupsArn) == d.Id() {
			sourceDatabaseRegion = aws.StringValue(backup.Region)
			databaseIdentifier = aws.StringValue(backup.DBInstanceIdentifier)
		} else {
			return fmt.Errorf("unable to find RDS instance automated backup replication: %s", d.Id())
		}
	}

	if err != nil {
		return fmt.Errorf("error reading RDS instance automated backup replication: %s", err)
	}

	// Initiate a stop of the replication process
	input := &rds.StopDBInstanceAutomatedBackupsReplicationInput{
		SourceDBInstanceArn: aws.String(d.Get("source_db_instance_arn").(string)),
	}

	log.Printf("[DEBUG] Stopping RDS instance automated backup replication for: %s", *input.SourceDBInstanceArn)

	_, err = conn.StopDBInstanceAutomatedBackupsReplication(input)

	if err != nil {
		return fmt.Errorf("error stopping RDS instance automated backup replication: %s", err)
	}

	// Create a new client to the source region
	sourceDatabaseConn := conn
	if sourceDatabaseRegion != meta.(*conns.AWSClient).Region {
		sourceDatabaseConn = rds.New(meta.(*conns.AWSClient).Session, aws.NewConfig().WithRegion(sourceDatabaseRegion))
	}

	// Wait for the source database to be available after the replication is stopped
	if _, err := waitDBInstanceAvailable(sourceDatabaseConn, databaseIdentifier, d.Timeout(schema.TimeoutDefault)); err != nil {
		return fmt.Errorf("error waiting for DB Instance (%s) delete: %w", *input.SourceDBInstanceArn, err)
	}

	return nil
}
