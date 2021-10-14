package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfrds "github.com/hashicorp/terraform-provider-aws/aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/rds/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/rds/waiter"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
	conn := meta.(*conns.AWSClient).RDSConn

	dbClusterID := d.Get("db_cluster_identifier").(string)
	roleARN := d.Get("role_arn").(string)
	input := &rds.AddRoleToDBClusterInput{
		DBClusterIdentifier: aws.String(dbClusterID),
		FeatureName:         aws.String(d.Get("feature_name").(string)),
		RoleArn:             aws.String(roleARN),
	}

	log.Printf("[DEBUG] Creating RDS DB Cluster IAM Role Association: %s", input)
	_, err := conn.AddRoleToDBCluster(input)

	if err != nil {
		return fmt.Errorf("error creating RDS DB Cluster (%s) IAM Role (%s) Association: %w", dbClusterID, roleARN, err)
	}

	d.SetId(tfrds.ClusterRoleAssociationCreateResourceID(dbClusterID, roleARN))

	_, err = waiter.DBClusterRoleAssociationCreated(conn, dbClusterID, roleARN)

	if err != nil {
		return fmt.Errorf("error waiting for RDS DB Cluster (%s) IAM Role (%s) Association to create: %w", dbClusterID, roleARN, err)
	}

	return resourceAwsRDSClusterRoleAssociationRead(d, meta)
}

func resourceAwsRDSClusterRoleAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn

	dbClusterID, roleARN, err := tfrds.ClusterRoleAssociationParseResourceID(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing RDS DB Cluster IAM Role Association ID: %s", err)
	}

	output, err := finder.DBClusterRoleByDBClusterIDAndRoleARN(conn, dbClusterID, roleARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS DB Cluster (%s) IAM Role (%s) Association not found, removing from state", dbClusterID, roleARN)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading RDS DB Cluster (%s) IAM Role (%s) Association: %w", dbClusterID, roleARN, err)
	}

	d.Set("db_cluster_identifier", dbClusterID)
	d.Set("feature_name", output.FeatureName)
	d.Set("role_arn", output.RoleArn)

	return nil
}

func resourceAwsRDSClusterRoleAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn

	dbClusterID, roleARN, err := tfrds.ClusterRoleAssociationParseResourceID(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing RDS DB Cluster IAM Role Association ID: %s", err)
	}

	input := &rds.RemoveRoleFromDBClusterInput{
		DBClusterIdentifier: aws.String(dbClusterID),
		FeatureName:         aws.String(d.Get("feature_name").(string)),
		RoleArn:             aws.String(roleARN),
	}

	log.Printf("[DEBUG] Deleting RDS DB Cluster IAM Role Association: %s", d.Id())
	_, err = conn.RemoveRoleFromDBCluster(input)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBClusterNotFoundFault) || tfawserr.ErrCodeEquals(err, rds.ErrCodeDBClusterRoleNotFoundFault) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting RDS DB Cluster (%s) IAM Role (%s) Association: %w", dbClusterID, roleARN, err)
	}

	_, err = waiter.DBClusterRoleAssociationDeleted(conn, dbClusterID, roleARN)

	if err != nil {
		return fmt.Errorf("error waiting for RDS DB Cluster (%s) IAM Role (%s) Association to delete: %w", dbClusterID, roleARN, err)
	}

	return nil
}
