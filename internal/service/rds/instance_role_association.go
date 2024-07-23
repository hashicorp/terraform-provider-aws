// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
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

// @SDKResource("aws_db_instance_role_association")
func ResourceInstanceRoleAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInstanceRoleAssociationCreate,
		ReadWithoutTimeout:   resourceInstanceRoleAssociationRead,
		DeleteWithoutTimeout: resourceInstanceRoleAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceInstanceRoleAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	dbInstanceIdentifier := d.Get("db_instance_identifier").(string)
	roleArn := d.Get(names.AttrRoleARN).(string)

	input := &rds.AddRoleToDBInstanceInput{
		DBInstanceIdentifier: aws.String(dbInstanceIdentifier),
		FeatureName:          aws.String(d.Get("feature_name").(string)),
		RoleArn:              aws.String(roleArn),
	}

	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		var err error
		_, err = conn.AddRoleToDBInstanceWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "IAM role ARN value is invalid or does not include the required permissions") {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.AddRoleToDBInstanceWithContext(ctx, input)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "associating RDS DB Instance (%s) IAM Role (%s): %s", dbInstanceIdentifier, roleArn, err)
	}

	d.SetId(fmt.Sprintf("%s,%s", dbInstanceIdentifier, roleArn))

	if err := waitForDBInstanceRoleAssociation(ctx, conn, dbInstanceIdentifier, roleArn); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Instance (%s) IAM Role (%s) association: %s", dbInstanceIdentifier, roleArn, err)
	}

	return append(diags, resourceInstanceRoleAssociationRead(ctx, d, meta)...)
}

func resourceInstanceRoleAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	dbInstanceIdentifier, roleArn, err := InstanceRoleAssociationDecodeID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading resource ID: %s", err)
	}

	dbInstanceRole, err := DescribeDBInstanceRole(ctx, conn, dbInstanceIdentifier, roleArn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS DB Instance (%s) not found, removing from state", dbInstanceIdentifier)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Instance (%s) IAM Role (%s) association: %s", dbInstanceIdentifier, roleArn, err)
	}

	d.Set("db_instance_identifier", dbInstanceIdentifier)
	d.Set("feature_name", dbInstanceRole.FeatureName)
	d.Set(names.AttrRoleARN, dbInstanceRole.RoleArn)

	return diags
}

func resourceInstanceRoleAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	dbInstanceIdentifier, roleArn, err := InstanceRoleAssociationDecodeID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading resource ID: %s", err)
	}

	input := &rds.RemoveRoleFromDBInstanceInput{
		DBInstanceIdentifier: aws.String(dbInstanceIdentifier),
		FeatureName:          aws.String(d.Get("feature_name").(string)),
		RoleArn:              aws.String(roleArn),
	}

	log.Printf("[DEBUG] RDS DB Instance (%s) IAM Role disassociating: %s", dbInstanceIdentifier, roleArn)
	_, err = conn.RemoveRoleFromDBInstanceWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBInstanceNotFoundFault) {
		return diags
	}

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBInstanceRoleNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disassociating RDS DB Instance (%s) IAM Role (%s): %s", dbInstanceIdentifier, roleArn, err)
	}

	if err := WaitForDBInstanceRoleDisassociation(ctx, conn, dbInstanceIdentifier, roleArn); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Instance (%s) IAM Role (%s) disassociation: %s", dbInstanceIdentifier, roleArn, err)
	}

	return diags
}

func InstanceRoleAssociationDecodeID(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected DB-INSTANCE-ID,ROLE-ARN", id)
	}

	return parts[0], parts[1], nil
}

func DescribeDBInstanceRole(ctx context.Context, conn *rds.RDS, dbInstanceIdentifier, roleArn string) (*rds.DBInstanceRole, error) {
	dbInstance, err := findDBInstanceByIDSDKv1(ctx, conn, dbInstanceIdentifier)
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

func waitForDBInstanceRoleAssociation(ctx context.Context, conn *rds.RDS, dbInstanceIdentifier, roleArn string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{dbInstanceRoleStatusPending},
		Target:  []string{dbInstanceRoleStatusActive},
		Refresh: statusDBInstanceRoleAssociation(ctx, conn, dbInstanceIdentifier, roleArn),
		Timeout: dbInstanceRoleAssociationCreatedTimeout,
	}

	log.Printf("[DEBUG] Waiting for RDS DB Instance (%s) IAM Role association: %s", dbInstanceIdentifier, roleArn)
	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitForDBInstanceRoleDisassociation(ctx context.Context, conn *rds.RDS, dbInstanceIdentifier, roleArn string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			dbInstanceRoleStatusActive,
			dbInstanceRoleStatusPending,
		},
		Target:  []string{},
		Refresh: statusDBInstanceRoleAssociation(ctx, conn, dbInstanceIdentifier, roleArn),
		Timeout: dbInstanceRoleAssociationDeletedTimeout,
	}

	log.Printf("[DEBUG] Waiting for RDS DB Instance (%s) IAM Role disassociation: %s", dbInstanceIdentifier, roleArn)
	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func statusDBInstanceRoleAssociation(ctx context.Context, conn *rds.RDS, dbInstanceIdentifier, roleArn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		dbInstanceRole, err := DescribeDBInstanceRole(ctx, conn, dbInstanceIdentifier, roleArn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return dbInstanceRole, aws.StringValue(dbInstanceRole.Status), nil
	}
}
