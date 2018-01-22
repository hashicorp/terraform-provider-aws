package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func resourceAwsDynamoDbBackup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDynamoDbBackupCreate,
		Read:   resourceAwsDynamoDbBackupRead,
		Delete: resourceAwsDynamoDbBackupDelete,

		SchemaVersion: 1,
		MigrateState:  resourceAwsDynamoDbTableMigrateState,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"table_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"backup_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsDynamoDbBackupCreate(d *schema.ResourceData, meta interface{}) error {
	dynamodbconn := meta.(*AWSClient).dynamodbconn

	tableName := d.Get("table_name").(string)
	backupName := d.Get("backup_name").(string)

	log.Printf("[DEBUG] DynamoDB backup create: %s", tableName)

	input := &dynamodb.CreateBackupInput{
		TableName:  aws.String(tableName),
		BackupName: aws.String(backupName),
	}

	out, err := dynamodbconn.CreateBackup(input)
	if err != nil {
		return fmt.Errorf("Error creating DynamoDB backup: %s", err)
	}
	stateConf := &resource.StateChangeConf{
		Pending: []string{"CREATING"},
		Target:  []string{"AVAILABLE"},
		Refresh: resourceAwsDynamoDbBackupRefreshFunc(dynamodbconn, out.BackupDetails.BackupArn),

		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		log.Printf("[ERR] Error waiting for backup (%s) to become ready: %s", d.Id(), err)
	}

	d.SetId(*out.BackupDetails.BackupArn)
	d.Set("arn", out.BackupDetails.BackupArn)

	return resourceAwsDynamoDbBackupRead(d, meta)
}

func resourceAwsDynamoDbBackupRead(d *schema.ResourceData, meta interface{}) error {
	dynamodbconn := meta.(*AWSClient).dynamodbconn
	log.Printf("[DEBUG] Checking for DynamoDB Backup '%s'", d.Id())
	req := &dynamodb.DescribeBackupInput{
		BackupArn: aws.String(d.Id()),
	}

	_, err := dynamodbconn.DescribeBackup(req)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
			log.Printf("[WARN] Dynamodb Backup (%s) not found, error code (404)", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	return nil
}

func resourceAwsDynamoDbBackupDelete(d *schema.ResourceData, meta interface{}) error {
	dynamodbconn := meta.(*AWSClient).dynamodbconn

	req := &dynamodb.DeleteBackupInput{
		BackupArn: aws.String(d.Id()),
	}

	_, err := dynamodbconn.DeleteBackup(req)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
			log.Printf("[WARN] Dynamodb Backup (%s) not found, error code (404)", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	return nil
}

func resourceAwsDynamoDbBackupRefreshFunc(conn *dynamodb.DynamoDB, arn *string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		log.Printf("[DEBUG] Checking if DynamoDb Backup Operation (%s) is Completed", *arn)

		req, output := conn.DescribeBackupRequest(&dynamodb.DescribeBackupInput{
			BackupArn: arn,
		})

		if req.Error != nil {
			return output, "FAILED", req.Error
		}

		if req.Operation == nil {
			return nil, "Failed", fmt.Errorf("[ERR] Error retrieving Operation info for operation (%v)", *req.Operation)
		}

		log.Printf("[DEBUG] Backup Operation (%v) is currently %q", *req.Operation, *output.BackupDescription.BackupDetails.BackupStatus)
		return req, *output.BackupDescription.BackupDetails.BackupStatus, nil
	}
}
