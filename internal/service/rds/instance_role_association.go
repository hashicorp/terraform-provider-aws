// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Constants not currently provided by the AWS Go SDK
const (
	dbInstanceRoleStatusActive  = "ACTIVE"
	dbInstanceRoleStatusPending = "PENDING"
)

// @SDKResource("aws_db_instance_role_association", name="DB Instance IAM Role Association")
func resourceInstanceRoleAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInstanceRoleAssociationCreate,
		ReadWithoutTimeout:   resourceInstanceRoleAssociationRead,
		DeleteWithoutTimeout: resourceInstanceRoleAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
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

func resourceInstanceRoleAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	dbInstanceIdentifier := d.Get("db_instance_identifier").(string)
	roleARN := d.Get(names.AttrRoleARN).(string)
	id := instanceRoleAssociationCreateResourceID(dbInstanceIdentifier, roleARN)
	input := &rds.AddRoleToDBInstanceInput{
		DBInstanceIdentifier: aws.String(dbInstanceIdentifier),
		FeatureName:          aws.String(d.Get("feature_name").(string)),
		RoleArn:              aws.String(roleARN),
	}

	_, err := conn.AddRoleToDBInstance(ctx, input)

	// check if the instance is in a valid state to add the role association
	if errs.IsA[*types.InvalidDBInstanceStateFault](err) {
		if _, err := waitDBInstanceAvailable(ctx, conn, dbInstanceIdentifier, d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Instance (%s) available: %s", dbInstanceIdentifier, err)
		}

		_, err = conn.AddRoleToDBInstance(ctx, input)
	}

	if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, errIAMRolePropagationMessage) {
		_, err = tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (any, error) {
			return conn.AddRoleToDBInstance(ctx, input)
		}, errCodeInvalidParameterValue, errIAMRolePropagationMessage)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS DB Instance IAM Role Association (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitDBInstanceRoleAssociationCreated(ctx, conn, dbInstanceIdentifier, roleARN, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Instance IAM Role Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceInstanceRoleAssociationRead(ctx, d, meta)...)
}

func resourceInstanceRoleAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	dbInstanceIdentifier, roleARN, err := instanceRoleAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	dbInstanceRole, err := findDBInstanceRoleByTwoPartKey(ctx, conn, dbInstanceIdentifier, roleARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS DB Instance (%s) IAM Role (%s) Association not found, removing from state", dbInstanceIdentifier, roleARN)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Instance IAM Role Association (%s): %s", d.Id(), err)
	}

	d.Set("db_instance_identifier", dbInstanceIdentifier)
	d.Set("feature_name", dbInstanceRole.FeatureName)
	d.Set(names.AttrRoleARN, dbInstanceRole.RoleArn)

	return diags
}

func resourceInstanceRoleAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	dbInstanceIdentifier, roleARN, err := instanceRoleAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting RDS DB Instance IAM Role Association: %s", d.Id())
	_, err = conn.RemoveRoleFromDBInstance(ctx, &rds.RemoveRoleFromDBInstanceInput{
		DBInstanceIdentifier: aws.String(dbInstanceIdentifier),
		FeatureName:          aws.String(d.Get("feature_name").(string)),
		RoleArn:              aws.String(roleARN),
	})

	if errs.IsA[*types.DBInstanceNotFoundFault](err) || errs.IsA[*types.DBInstanceRoleNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS DB Instance IAM Role Association (%s): %s", d.Id(), err)
	}

	if _, err := waitDBInstanceRoleAssociationDeleted(ctx, conn, dbInstanceIdentifier, roleARN, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Instance IAM Role Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const instanceRoleAssociationResourceIDSeparator = ","

func instanceRoleAssociationCreateResourceID(dbInstanceID, roleARN string) string {
	parts := []string{dbInstanceID, roleARN}
	id := strings.Join(parts, instanceRoleAssociationResourceIDSeparator)

	return id
}

func instanceRoleAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, instanceRoleAssociationResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected DB-INSTANCE-ID%[2]sROLE-ARN", id, instanceRoleAssociationResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findDBInstanceRoleByTwoPartKey(ctx context.Context, conn *rds.Client, dbInstanceIdentifier, roleARN string) (*types.DBInstanceRole, error) {
	dbInstance, err := findDBInstanceByID(ctx, conn, dbInstanceIdentifier)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(tfslices.Filter(dbInstance.AssociatedRoles, func(v types.DBInstanceRole) bool {
		return aws.ToString(v.RoleArn) == roleARN
	}))
}

func statusDBInstanceRoleAssociation(ctx context.Context, conn *rds.Client, dbInstanceIdentifier, roleARN string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findDBInstanceRoleByTwoPartKey(ctx, conn, dbInstanceIdentifier, roleARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

func waitDBInstanceRoleAssociationCreated(ctx context.Context, conn *rds.Client, dbInstanceIdentifier, roleARN string, timeout time.Duration) (*types.DBInstanceRole, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{dbInstanceRoleStatusPending},
		Target:  []string{dbInstanceRoleStatusActive},
		Refresh: statusDBInstanceRoleAssociation(ctx, conn, dbInstanceIdentifier, roleARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DBInstanceRole); ok {
		return output, err
	}

	return nil, err
}

func waitDBInstanceRoleAssociationDeleted(ctx context.Context, conn *rds.Client, dbInstanceIdentifier, roleARN string, timeout time.Duration) (*types.DBInstanceRole, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{dbInstanceRoleStatusActive, dbInstanceRoleStatusPending},
		Target:  []string{},
		Refresh: statusDBInstanceRoleAssociation(ctx, conn, dbInstanceIdentifier, roleARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DBInstanceRole); ok {
		return output, err
	}

	return nil, err
}
