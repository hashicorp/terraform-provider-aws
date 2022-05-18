package rds

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// Constants not currently provided by the AWS Go SDK
const (
	instanceRoleStatusActive  = "ACTIVE"
	instanceRoleStatusDeleted = "DELETED"
	instanceRoleStatusPending = "PENDING"
)

func ResourceInstanceRoleAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceInstanceRoleAssociationCreate,
		Read:   resourceInstanceRoleAssociationRead,
		Delete: resourceInstanceRoleAssociationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"db_instance_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"feature_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceInstanceRoleAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn

	dbInstanceIdentifier := d.Get("db_instance_identifier").(string)
	roleArn := d.Get("role_arn").(string)

	input := &rds.AddRoleToDBInstanceInput{
		DBInstanceIdentifier: aws.String(dbInstanceIdentifier),
		FeatureName:          aws.String(d.Get("feature_name").(string)),
		RoleArn:              aws.String(roleArn),
	}

	log.Printf("[DEBUG] RDS DB Instance (%s) IAM Role associating: %s", dbInstanceIdentifier, roleArn)
	_, err := conn.AddRoleToDBInstance(input)

	if err != nil {
		return fmt.Errorf("error associating RDS DB Instance (%s) IAM Role (%s): %s", dbInstanceIdentifier, roleArn, err)
	}

	d.SetId(fmt.Sprintf("%s,%s", dbInstanceIdentifier, roleArn))

	if err := waitForDBInstanceRoleAssociation(conn, dbInstanceIdentifier, roleArn); err != nil {
		return fmt.Errorf("error waiting for RDS DB Instance (%s) IAM Role (%s) association: %s", dbInstanceIdentifier, roleArn, err)
	}

	return resourceInstanceRoleAssociationRead(d, meta)
}

func resourceInstanceRoleAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn

	dbInstanceIdentifier, roleArn, err := InstanceRoleAssociationDecodeID(d.Id())

	if err != nil {
		return fmt.Errorf("error reading resource ID: %s", err)
	}

	dbInstanceRole, err := DescribeInstanceRole(conn, dbInstanceIdentifier, roleArn)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBInstanceNotFoundFault) {
		log.Printf("[WARN] RDS DB Instance (%s) not found, removing from state", dbInstanceIdentifier)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading RDS DB Instance (%s) IAM Role (%s) association: %s", dbInstanceIdentifier, roleArn, err)
	}

	if dbInstanceRole == nil {
		log.Printf("[WARN] RDS DB Instance (%s) IAM Role (%s) association not found, removing from state", dbInstanceIdentifier, roleArn)
		d.SetId("")
		return nil
	}

	d.Set("db_instance_identifier", dbInstanceIdentifier)
	d.Set("feature_name", dbInstanceRole.FeatureName)
	d.Set("role_arn", dbInstanceRole.RoleArn)

	return nil
}

func resourceInstanceRoleAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn

	dbInstanceIdentifier, roleArn, err := InstanceRoleAssociationDecodeID(d.Id())

	if err != nil {
		return fmt.Errorf("error reading resource ID: %s", err)
	}

	input := &rds.RemoveRoleFromDBInstanceInput{
		DBInstanceIdentifier: aws.String(dbInstanceIdentifier),
		FeatureName:          aws.String(d.Get("feature_name").(string)),
		RoleArn:              aws.String(roleArn),
	}

	log.Printf("[DEBUG] RDS DB Instance (%s) IAM Role disassociating: %s", dbInstanceIdentifier, roleArn)
	_, err = conn.RemoveRoleFromDBInstance(input)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBInstanceNotFoundFault) {
		return nil
	}

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBInstanceRoleNotFoundFault) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error disassociating RDS DB Instance (%s) IAM Role (%s): %s", dbInstanceIdentifier, roleArn, err)
	}

	if err := WaitForInstanceRoleDisassociation(conn, dbInstanceIdentifier, roleArn); err != nil {
		return fmt.Errorf("error waiting for RDS DB Instance (%s) IAM Role (%s) disassociation: %s", dbInstanceIdentifier, roleArn, err)
	}

	return nil
}

func InstanceRoleAssociationDecodeID(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected DB-INSTANCE-ID,ROLE-ARN", id)
	}

	return parts[0], parts[1], nil
}

func DescribeInstanceRole(conn *rds.RDS, dbInstanceIdentifier, roleArn string) (*rds.DBInstanceRole, error) {
	input := &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(dbInstanceIdentifier),
	}

	log.Printf("[DEBUG] Describing RDS DB Instance: %s", input)
	output, err := conn.DescribeDBInstances(input)

	if err != nil {
		return nil, err
	}

	var dbInstance *rds.DBInstance

	for _, outputDbInstance := range output.DBInstances {
		if aws.StringValue(outputDbInstance.DBInstanceIdentifier) == dbInstanceIdentifier {
			dbInstance = outputDbInstance
			break
		}
	}

	if dbInstance == nil {
		return nil, nil
	}

	var dbInstanceRole *rds.DBInstanceRole

	for _, associatedRole := range dbInstance.AssociatedRoles {
		if aws.StringValue(associatedRole.RoleArn) == roleArn {
			dbInstanceRole = associatedRole
			break
		}
	}

	return dbInstanceRole, nil
}

func waitForDBInstanceRoleAssociation(conn *rds.RDS, dbInstanceIdentifier, roleArn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{instanceRoleStatusPending},
		Target:  []string{instanceRoleStatusActive},
		Refresh: func() (interface{}, string, error) {
			dbInstanceRole, err := DescribeInstanceRole(conn, dbInstanceIdentifier, roleArn)

			if err != nil {
				return nil, "", err
			}

			if dbInstanceRole == nil {
				return nil, instanceRoleStatusPending, nil
			}

			return dbInstanceRole, aws.StringValue(dbInstanceRole.Status), nil
		},
		Timeout: 5 * time.Minute,
		Delay:   5 * time.Second,
	}

	log.Printf("[DEBUG] Waiting for RDS DB Instance (%s) IAM Role association: %s", dbInstanceIdentifier, roleArn)
	_, err := stateConf.WaitForState()

	return err
}

func WaitForInstanceRoleDisassociation(conn *rds.RDS, dbInstanceIdentifier, roleArn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			instanceRoleStatusActive,
			instanceRoleStatusPending,
		},
		Target: []string{instanceRoleStatusDeleted},
		Refresh: func() (interface{}, string, error) {
			dbInstanceRole, err := DescribeInstanceRole(conn, dbInstanceIdentifier, roleArn)

			if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBInstanceNotFoundFault) {
				return &rds.DBInstanceRole{}, instanceRoleStatusDeleted, nil
			}

			if err != nil {
				return nil, "", err
			}

			if dbInstanceRole != nil {
				return dbInstanceRole, aws.StringValue(dbInstanceRole.Status), nil
			}

			return &rds.DBInstanceRole{}, instanceRoleStatusDeleted, nil
		},
		Timeout: 5 * time.Minute,
		Delay:   5 * time.Second,
	}

	log.Printf("[DEBUG] Waiting for RDS DB Instance (%s) IAM Role disassociation: %s", dbInstanceIdentifier, roleArn)
	_, err := stateConf.WaitForState()

	return err
}
