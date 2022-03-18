package rds

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func resourceInstanceBackupReplication() *schema.Resource {
	return &schema.Resource{
		Create: resourceInstanceBackupReplicationCreate,
		Read:   resourceInstanceBackupReplicationRead,
		Delete: resourceInstanceBackupReplicationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"source_db_instance_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"destination_region": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"retention_period": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceInstanceBackupReplicationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn

	input := &rds.StartDBInstanceAutomatedBackupsReplicationInput{
		BackupRetentionPeriod: aws.Int64(int64(d.Get("retention_period").(int))),
		DestinationRegion:     aws.String(d.Get("destination_region").(string)),
		SourceDBInstanceArn:   aws.String(d.Get("source_db_instance_arn").(string)),
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Starting automated backup replication for: %s", *input.SourceDBInstanceArn)

	output, err := conn.StartDBInstanceAutomatedBackupsReplication(input)

	if err != nil {
		return fmt.Errorf("error creating RDS instance backup replication: %s", err)
	}

	d.SetId(aws.StringValue(output.DBInstanceAutomatedBackup.DBInstanceAutomatedBackupsArn))

	return resourceInstanceBackupReplicationRead(d, meta)
}

func resourceInstanceBackupReplicationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn

	return nil
}

func resourceInstanceBackupReplicationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn

	return nil
}
