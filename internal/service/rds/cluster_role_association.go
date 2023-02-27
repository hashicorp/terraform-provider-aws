package rds

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceClusterRoleAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterRoleAssociationCreate,
		ReadWithoutTimeout:   resourceClusterRoleAssociationRead,
		DeleteWithoutTimeout: resourceClusterRoleAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceClusterRoleAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn()

	dbClusterID := d.Get("db_cluster_identifier").(string)
	roleARN := d.Get("role_arn").(string)
	input := &rds.AddRoleToDBClusterInput{
		DBClusterIdentifier: aws.String(dbClusterID),
		FeatureName:         aws.String(d.Get("feature_name").(string)),
		RoleArn:             aws.String(roleARN),
	}

	err := resource.RetryContext(ctx, propagationTimeout, func() *resource.RetryError {
		var err error
		_, err = conn.AddRoleToDBClusterWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "IAM role ARN value is invalid or does not include the required permissions") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.AddRoleToDBClusterWithContext(ctx, input)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS DB Cluster (%s) IAM Role (%s) Association: %s", dbClusterID, roleARN, err)
	}

	d.SetId(ClusterRoleAssociationCreateResourceID(dbClusterID, roleARN))

	_, err = waitDBClusterRoleAssociationCreated(ctx, conn, dbClusterID, roleARN)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Cluster (%s) IAM Role (%s) Association to create: %s", dbClusterID, roleARN, err)
	}

	return append(diags, resourceClusterRoleAssociationRead(ctx, d, meta)...)
}

func resourceClusterRoleAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn()

	dbClusterID, roleARN, err := ClusterRoleAssociationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing RDS DB Cluster IAM Role Association ID: %s", err)
	}

	output, err := FindDBClusterRoleByDBClusterIDAndRoleARN(ctx, conn, dbClusterID, roleARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS DB Cluster (%s) IAM Role (%s) Association not found, removing from state", dbClusterID, roleARN)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Cluster (%s) IAM Role (%s) Association: %s", dbClusterID, roleARN, err)
	}

	d.Set("db_cluster_identifier", dbClusterID)
	d.Set("feature_name", output.FeatureName)
	d.Set("role_arn", output.RoleArn)

	return diags
}

func resourceClusterRoleAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn()

	dbClusterID, roleARN, err := ClusterRoleAssociationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing RDS DB Cluster IAM Role Association ID: %s", err)
	}

	input := &rds.RemoveRoleFromDBClusterInput{
		DBClusterIdentifier: aws.String(dbClusterID),
		FeatureName:         aws.String(d.Get("feature_name").(string)),
		RoleArn:             aws.String(roleARN),
	}

	log.Printf("[DEBUG] Deleting RDS DB Cluster IAM Role Association: %s", d.Id())
	_, err = conn.RemoveRoleFromDBClusterWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBClusterNotFoundFault) || tfawserr.ErrCodeEquals(err, rds.ErrCodeDBClusterRoleNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS DB Cluster (%s) IAM Role (%s) Association: %s", dbClusterID, roleARN, err)
	}

	_, err = waitDBClusterRoleAssociationDeleted(ctx, conn, dbClusterID, roleARN)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Cluster (%s) IAM Role (%s) Association to delete: %s", dbClusterID, roleARN, err)
	}

	return diags
}
