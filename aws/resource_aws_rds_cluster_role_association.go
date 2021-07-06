package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Constants not currently provided by the AWS Go SDK
const (
	rdsDbClusterRoleStatusActive  = "ACTIVE"
	rdsDbClusterRoleStatusDeleted = "DELETED"
	rdsDbClusterRoleStatusPending = "PENDING"
)

func resourceAwsRDSClusterRoleAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRDSClusterRoleAssociationCreate,
		Read:   resourceAwsRDSClusterRoleAssociationRead,
		Delete: resourceAwsRDSClusterRoleAssociationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"db_cluster_identifier": {
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
				ValidateFunc: validateArn,
			},
		},
	}
}

func resourceAwsRDSClusterRoleAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	dbClusterIdentifier := d.Get("db_cluster_identifier").(string)
	roleArn := d.Get("role_arn").(string)

	input := &rds.AddRoleToDBClusterInput{
		DBClusterIdentifier: aws.String(dbClusterIdentifier),
		FeatureName:         aws.String(d.Get("feature_name").(string)),
		RoleArn:             aws.String(roleArn),
	}

	log.Printf("[DEBUG] RDS DB Cluster (%s) IAM Role associating: %s", dbClusterIdentifier, roleArn)
	_, err := conn.AddRoleToDBCluster(input)

	if err != nil {
		return fmt.Errorf("error associating RDS DB Cluster (%s) IAM Role (%s): %s", dbClusterIdentifier, roleArn, err)
	}

	d.SetId(fmt.Sprintf("%s,%s", dbClusterIdentifier, roleArn))

	if err := waitForRdsDbClusterRoleAssociation(conn, dbClusterIdentifier, roleArn); err != nil {
		return fmt.Errorf("error waiting for RDS DB Cluster (%s) IAM Role (%s) association: %s", dbClusterIdentifier, roleArn, err)
	}

	return resourceAwsRDSClusterRoleAssociationRead(d, meta)
}

func resourceAwsRDSClusterRoleAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	dbClusterIdentifier, roleArn, err := resourceAwsDbClusterRoleAssociationDecodeID(d.Id())

	if err != nil {
		return fmt.Errorf("error reading resource ID: %s", err)
	}

	dbClusterRole, err := rdsDescribeDbClusterRole(conn, dbClusterIdentifier, roleArn)

	if isAWSErr(err, rds.ErrCodeDBClusterNotFoundFault, "") {
		log.Printf("[WARN] RDS DB Cluster (%s) not found, removing from state", dbClusterIdentifier)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading RDS DB Cluster (%s) IAM Role (%s) association: %s", dbClusterIdentifier, roleArn, err)
	}

	if dbClusterRole == nil {
		log.Printf("[WARN] RDS DB Cluster (%s) IAM Role (%s) association not found, removing from state", dbClusterIdentifier, roleArn)
		d.SetId("")
		return nil
	}

	d.Set("db_cluster_identifier", dbClusterIdentifier)
	d.Set("feature_name", dbClusterRole.FeatureName)
	d.Set("role_arn", dbClusterRole.RoleArn)

	return nil
}

func resourceAwsRDSClusterRoleAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	dbClusterIdentifier, roleArn, err := resourceAwsDbClusterRoleAssociationDecodeID(d.Id())

	if err != nil {
		return fmt.Errorf("error reading resource ID: %s", err)
	}

	input := &rds.RemoveRoleFromDBClusterInput{
		DBClusterIdentifier: aws.String(dbClusterIdentifier),
		FeatureName:         aws.String(d.Get("feature_name").(string)),
		RoleArn:             aws.String(roleArn),
	}

	log.Printf("[DEBUG] RDS DB Cluster (%s) IAM Role disassociating: %s", dbClusterIdentifier, roleArn)
	_, err = conn.RemoveRoleFromDBCluster(input)

	if isAWSErr(err, rds.ErrCodeDBClusterNotFoundFault, "") {
		return nil
	}

	if isAWSErr(err, rds.ErrCodeDBClusterRoleNotFoundFault, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error disassociating RDS DB Cluster (%s) IAM Role (%s): %s", dbClusterIdentifier, roleArn, err)
	}

	if err := waitForRdsDbClusterRoleDisassociation(conn, dbClusterIdentifier, roleArn); err != nil {
		return fmt.Errorf("error waiting for RDS DB Cluster (%s) IAM Role (%s) disassociation: %s", dbClusterIdentifier, roleArn, err)
	}

	return nil
}

func resourceAwsDbClusterRoleAssociationDecodeID(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected DB-CLUSTER-ID,ROLE-ARN", id)
	}

	return parts[0], parts[1], nil
}

func rdsDescribeDbClusterRole(conn *rds.RDS, dbClusterIdentifier, roleArn string) (*rds.DBClusterRole, error) {
	input := &rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(dbClusterIdentifier),
	}

	log.Printf("[DEBUG] Describing RDS DB Cluster: %s", input)
	output, err := conn.DescribeDBClusters(input)

	if err != nil {
		return nil, err
	}

	var dbCluster *rds.DBCluster

	for _, outputDbCluster := range output.DBClusters {
		if aws.StringValue(outputDbCluster.DBClusterIdentifier) == dbClusterIdentifier {
			dbCluster = outputDbCluster
			break
		}
	}

	if dbCluster == nil {
		return nil, nil
	}

	var dbClusterRole *rds.DBClusterRole

	for _, associatedRole := range dbCluster.AssociatedRoles {
		if aws.StringValue(associatedRole.RoleArn) == roleArn {
			dbClusterRole = associatedRole
			break
		}
	}

	return dbClusterRole, nil
}

func waitForRdsDbClusterRoleAssociation(conn *rds.RDS, dbClusterIdentifier, roleArn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{rdsDbClusterRoleStatusPending},
		Target:  []string{rdsDbClusterRoleStatusActive},
		Refresh: func() (interface{}, string, error) {
			dbClusterRole, err := rdsDescribeDbClusterRole(conn, dbClusterIdentifier, roleArn)

			if err != nil {
				return nil, "", err
			}

			return dbClusterRole, aws.StringValue(dbClusterRole.Status), nil
		},
		Timeout: 5 * time.Minute,
		Delay:   5 * time.Second,
	}

	log.Printf("[DEBUG] Waiting for RDS DB Cluster (%s) IAM Role association: %s", dbClusterIdentifier, roleArn)
	_, err := stateConf.WaitForState()

	return err
}

func waitForRdsDbClusterRoleDisassociation(conn *rds.RDS, dbClusterIdentifier, roleArn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			rdsDbClusterRoleStatusActive,
			rdsDbClusterRoleStatusPending,
		},
		Target: []string{rdsDbClusterRoleStatusDeleted},
		Refresh: func() (interface{}, string, error) {
			dbClusterRole, err := rdsDescribeDbClusterRole(conn, dbClusterIdentifier, roleArn)

			if isAWSErr(err, rds.ErrCodeDBClusterNotFoundFault, "") {
				return &rds.DBClusterRole{}, rdsDbClusterRoleStatusDeleted, nil
			}

			if err != nil {
				return nil, "", err
			}

			if dbClusterRole != nil {
				return dbClusterRole, aws.StringValue(dbClusterRole.Status), nil
			}

			return &rds.DBClusterRole{}, rdsDbClusterRoleStatusDeleted, nil
		},
		Timeout: 5 * time.Minute,
		Delay:   5 * time.Second,
	}

	log.Printf("[DEBUG] Waiting for RDS DB Cluster (%s) IAM Role disassociation: %s", dbClusterIdentifier, roleArn)
	_, err := stateConf.WaitForState()

	return err
}
