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
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
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

// @SDKResource("aws_rds_cluster_role_association", name="Cluster IAM Role Association")
func resourceClusterRoleAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterRoleAssociationCreate,
		ReadWithoutTimeout:   resourceClusterRoleAssociationRead,
		DeleteWithoutTimeout: resourceClusterRoleAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
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
			names.AttrRoleARN: {
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
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	dbClusterID := d.Get("db_cluster_identifier").(string)
	roleARN := d.Get(names.AttrRoleARN).(string)
	id := clusterRoleAssociationCreateResourceID(dbClusterID, roleARN)
	input := &rds.AddRoleToDBClusterInput{
		DBClusterIdentifier: aws.String(dbClusterID),
		FeatureName:         aws.String(d.Get("feature_name").(string)),
		RoleArn:             aws.String(roleARN),
	}

	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.AddRoleToDBCluster(ctx, input)
	}, errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS Cluster IAM Role Association (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitDBClusterRoleAssociationCreated(ctx, conn, dbClusterID, roleARN, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Cluster IAM Role Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceClusterRoleAssociationRead(ctx, d, meta)...)
}

func resourceClusterRoleAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	dbClusterID, roleARN, err := clusterRoleAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findDBClusterRoleByTwoPartKey(ctx, conn, dbClusterID, roleARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS DB Cluster (%s) IAM Role (%s) Association not found, removing from state", dbClusterID, roleARN)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Cluster IAM Role Association (%s): %s", d.Id(), err)
	}

	d.Set("db_cluster_identifier", dbClusterID)
	d.Set("feature_name", output.FeatureName)
	d.Set(names.AttrRoleARN, output.RoleArn)

	return diags
}

func resourceClusterRoleAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	dbClusterID, roleARN, err := clusterRoleAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting RDS Cluster IAM Role Association: %s", d.Id())
	_, err = conn.RemoveRoleFromDBCluster(ctx, &rds.RemoveRoleFromDBClusterInput{
		DBClusterIdentifier: aws.String(dbClusterID),
		FeatureName:         aws.String(d.Get("feature_name").(string)),
		RoleArn:             aws.String(roleARN),
	})

	if errs.IsA[*types.DBClusterNotFoundFault](err) || errs.IsA[*types.DBClusterRoleNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS Cluster IAM Role Association (%s): %s", d.Id(), err)
	}

	if _, err := waitDBClusterRoleAssociationDeleted(ctx, conn, dbClusterID, roleARN, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Cluster IAM Role Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const clusterRoleAssociationResourceIDSeparator = ","

func clusterRoleAssociationCreateResourceID(dbClusterID, roleARN string) string {
	parts := []string{dbClusterID, roleARN}
	id := strings.Join(parts, clusterRoleAssociationResourceIDSeparator)

	return id
}

func clusterRoleAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, clusterRoleAssociationResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected DBCLUSTERID%[2]sROLEARN", id, clusterRoleAssociationResourceIDSeparator)
}

func findDBClusterRoleByTwoPartKey(ctx context.Context, conn *rds.Client, dbClusterID, roleARN string) (*types.DBClusterRole, error) {
	dbCluster, err := findDBClusterByIDV2(ctx, conn, dbClusterID)

	if err != nil {
		return nil, err
	}

	output, err := tfresource.AssertSingleValueResult(tfslices.Filter(dbCluster.AssociatedRoles, func(v types.DBClusterRole) bool {
		return aws.ToString(v.RoleArn) == roleARN
	}))

	if err != nil {
		return nil, err
	}

	if status := aws.ToString(output.Status); status == clusterRoleStatusDeleted {
		return nil, &retry.NotFoundError{
			Message: status,
		}
	}

	return output, nil
}

func statusDBClusterRole(ctx context.Context, conn *rds.Client, dbClusterID, roleARN string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDBClusterRoleByTwoPartKey(ctx, conn, dbClusterID, roleARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

func waitDBClusterRoleAssociationCreated(ctx context.Context, conn *rds.Client, dbClusterID, roleARN string, timeout time.Duration) (*types.DBClusterRole, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{clusterRoleStatusPending},
		Target:     []string{clusterRoleStatusActive},
		Refresh:    statusDBClusterRole(ctx, conn, dbClusterID, roleARN),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DBClusterRole); ok {
		return output, err
	}

	return nil, err
}

func waitDBClusterRoleAssociationDeleted(ctx context.Context, conn *rds.Client, dbClusterID, roleARN string, timeout time.Duration) (*types.DBClusterRole, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{clusterRoleStatusActive, clusterRoleStatusPending},
		Target:     []string{},
		Refresh:    statusDBClusterRole(ctx, conn, dbClusterID, roleARN),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DBClusterRole); ok {
		return output, err
	}

	return nil, err
}

// TODO Remove once aws_rds_cluster is migrated.
func findDBClusterByIDV2(ctx context.Context, conn *rds.Client, id string) (*types.DBCluster, error) {
	input := &rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(id),
	}
	output, err := findDBClusterV2(ctx, conn, input, tfslices.PredicateTrue[*types.DBCluster]())

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if arn.IsARN(id) {
		if aws.ToString(output.DBClusterArn) != id {
			return nil, &retry.NotFoundError{
				LastRequest: input,
			}
		}
	} else if aws.ToString(output.DBClusterIdentifier) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findDBClusterV2(ctx context.Context, conn *rds.Client, input *rds.DescribeDBClustersInput, filter tfslices.Predicate[*types.DBCluster]) (*types.DBCluster, error) {
	output, err := findDBClustersV2(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDBClustersV2(ctx context.Context, conn *rds.Client, input *rds.DescribeDBClustersInput, filter tfslices.Predicate[*types.DBCluster]) ([]types.DBCluster, error) {
	var output []types.DBCluster

	pages := rds.NewDescribeDBClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.DBClusterNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.DBClusters {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
