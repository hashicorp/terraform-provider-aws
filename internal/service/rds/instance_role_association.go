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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// Constants not currently provided by the AWS Go SDK
const (
	dbInstanceRoleStatusActive  = "ACTIVE"
	dbInstanceRoleStatusPending = "PENDING"
)

const (
	dbInstanceRoleAssociationCreatedTimeout = 10 * time.Minute
	dbInstanceRoleAssociationDeletedTimeout = 10 * time.Minute
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

	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error
		_, err = conn.AddRoleToDBInstance(input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "IAM role ARN value is invalid or does not include the required permissions") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.AddRoleToDBInstance(input)
	}
	if err != nil {
		return fmt.Errorf("error associating RDS DB Instance (%s) IAM Role (%s): %w", dbInstanceIdentifier, roleArn, err)
	}

	d.SetId(fmt.Sprintf("%s,%s", dbInstanceIdentifier, roleArn))

	if err := waitForDBInstanceRoleAssociation(conn, dbInstanceIdentifier, roleArn); err != nil {
		return fmt.Errorf("error waiting for RDS DB Instance (%s) IAM Role (%s) association: %w", dbInstanceIdentifier, roleArn, err)
	}

	return resourceInstanceRoleAssociationRead(d, meta)
}

func resourceInstanceRoleAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn

	dbInstanceIdentifier, roleArn, err := InstanceRoleAssociationDecodeID(d.Id())

	if err != nil {
		return fmt.Errorf("error reading resource ID: %w", err)
	}

	dbInstanceRole, err := DescribeDBInstanceRole(conn, dbInstanceIdentifier, roleArn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS DB Instance (%s) not found, removing from state", dbInstanceIdentifier)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading RDS DB Instance (%s) IAM Role (%s) association: %w", dbInstanceIdentifier, roleArn, err)
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
		return fmt.Errorf("error reading resource ID: %w", err)
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
		return fmt.Errorf("error disassociating RDS DB Instance (%s) IAM Role (%s): %w", dbInstanceIdentifier, roleArn, err)
	}

	if err := WaitForDBInstanceRoleDisassociation(conn, dbInstanceIdentifier, roleArn); err != nil {
		return fmt.Errorf("error waiting for RDS DB Instance (%s) IAM Role (%s) disassociation: %w", dbInstanceIdentifier, roleArn, err)
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

func DescribeDBInstanceRole(conn *rds.RDS, dbInstanceIdentifier, roleArn string) (*rds.DBInstanceRole, error) {
	dbInstance, err := FindDBInstanceByID(conn, dbInstanceIdentifier)
	if err != nil {
		return nil, err
	}

	for _, associatedRole := range dbInstance.AssociatedRoles {
		if aws.StringValue(associatedRole.RoleArn) == roleArn {
			return associatedRole, nil
		}
	}

	return nil, &tfresource.EmptyResultError{}
}

func waitForDBInstanceRoleAssociation(conn *rds.RDS, dbInstanceIdentifier, roleArn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{dbInstanceRoleStatusPending},
		Target:  []string{dbInstanceRoleStatusActive},
		Refresh: statusDBInstanceRoleAssociation(conn, dbInstanceIdentifier, roleArn),
		Timeout: dbInstanceRoleAssociationCreatedTimeout,
	}

	log.Printf("[DEBUG] Waiting for RDS DB Instance (%s) IAM Role association: %s", dbInstanceIdentifier, roleArn)
	_, err := stateConf.WaitForState()

	return err
}

func WaitForDBInstanceRoleDisassociation(conn *rds.RDS, dbInstanceIdentifier, roleArn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dbInstanceRoleStatusActive,
			dbInstanceRoleStatusPending,
		},
		Target:  []string{},
		Refresh: statusDBInstanceRoleAssociation(conn, dbInstanceIdentifier, roleArn),
		Timeout: dbInstanceRoleAssociationDeletedTimeout,
	}

	log.Printf("[DEBUG] Waiting for RDS DB Instance (%s) IAM Role disassociation: %s", dbInstanceIdentifier, roleArn)
	_, err := stateConf.WaitForState()

	return err
}

func statusDBInstanceRoleAssociation(conn *rds.RDS, dbInstanceIdentifier, roleArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		dbInstanceRole, err := DescribeDBInstanceRole(conn, dbInstanceIdentifier, roleArn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return dbInstanceRole, aws.StringValue(dbInstanceRole.Status), nil
	}
}
