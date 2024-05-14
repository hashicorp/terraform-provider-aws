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
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_db_proxy_target", name="DB Proxy Target")
func resourceProxyTarget() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProxyTargetCreate,
		ReadWithoutTimeout:   resourceProxyTargetRead,
		DeleteWithoutTimeout: resourceProxyTargetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"db_cluster_identifier": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validIdentifier,
				ExactlyOneOf: []string{
					"db_instance_identifier",
					"db_cluster_identifier",
				},
			},
			"db_instance_identifier": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validIdentifier,
				ExactlyOneOf: []string{
					"db_instance_identifier",
					"db_cluster_identifier",
				},
			},
			"db_proxy_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validIdentifier,
			},
			names.AttrEndpoint: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPort: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"rds_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTargetARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"target_group_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validIdentifier,
			},
			"tracked_cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrType: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceProxyTargetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	dbProxyName := d.Get("db_proxy_name").(string)
	targetGroupName := d.Get("target_group_name").(string)
	input := &rds.RegisterDBProxyTargetsInput{
		DBProxyName:     aws.String(dbProxyName),
		TargetGroupName: aws.String(targetGroupName),
	}

	if v, ok := d.GetOk("db_cluster_identifier"); ok {
		input.DBClusterIdentifiers = []string{v.(string)}
	}

	if v, ok := d.GetOk("db_instance_identifier"); ok {
		input.DBInstanceIdentifiers = []string{v.(string)}
	}

	const (
		timeout = 5 * time.Minute
	)
	outputRaw, err := tfresource.RetryWhenIsAErrorMessageContains[*types.InvalidDBInstanceStateFault](ctx, timeout,
		func() (interface{}, error) {
			return conn.RegisterDBProxyTargets(ctx, input)
		},
		"CREATING")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "registering RDS DB Proxy Target (%s/%s): %s", dbProxyName, targetGroupName, err)
	}

	dbProxyTarget := outputRaw.(*rds.RegisterDBProxyTargetsOutput).DBProxyTargets[0]
	d.SetId(proxyTargetCreateResourceID(dbProxyName, targetGroupName, string(dbProxyTarget.Type), aws.ToString(dbProxyTarget.RdsResourceId)))

	return append(diags, resourceProxyTargetRead(ctx, d, meta)...)
}

func resourceProxyTargetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	dbProxyName, targetGroupName, targetType, rdsResourceID, err := proxyTargetParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	dbProxyTarget, err := findDBProxyTargetByFourPartKey(ctx, conn, dbProxyName, targetGroupName, targetType, rdsResourceID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS DB Proxy Target (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Proxy Target (%s): %s", d.Id(), err)
	}

	switch dbProxyTarget.Type {
	case types.TargetTypeRdsInstance:
		d.Set("db_instance_identifier", dbProxyTarget.RdsResourceId)
	default:
		d.Set("db_cluster_identifier", dbProxyTarget.RdsResourceId)
	}
	d.Set("db_proxy_name", dbProxyName)
	d.Set(names.AttrEndpoint, dbProxyTarget.Endpoint)
	d.Set(names.AttrPort, dbProxyTarget.Port)
	d.Set("rds_resource_id", dbProxyTarget.RdsResourceId)
	d.Set(names.AttrTargetARN, dbProxyTarget.TargetArn)
	d.Set("target_group_name", targetGroupName)
	d.Set("tracked_cluster_id", dbProxyTarget.TrackedClusterId)
	d.Set(names.AttrType, dbProxyTarget.Type)

	return diags
}

func resourceProxyTargetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	dbProxyName, targetGroupName, targetType, rdsResourceID, err := proxyTargetParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &rds.DeregisterDBProxyTargetsInput{
		DBProxyName:     aws.String(dbProxyName),
		TargetGroupName: aws.String(targetGroupName),
	}

	switch targetType {
	case string(types.TargetTypeRdsInstance):
		input.DBInstanceIdentifiers = []string{rdsResourceID}
	default:
		input.DBClusterIdentifiers = []string{rdsResourceID}
	}

	log.Printf("[DEBUG] Deleting RDS DB Proxy Target: %s", d.Id())
	_, err = conn.DeregisterDBProxyTargets(ctx, input)

	if errs.IsA[*types.DBProxyNotFoundFault](err) || errs.IsA[*types.DBProxyTargetGroupNotFoundFault](err) || errs.IsA[*types.DBProxyTargetNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deregistering RDS DB Proxy Target (%s): %s", d.Id(), err)
	}

	return diags
}

const proxyTargetResourceIDSeparator = "/"

func proxyTargetCreateResourceID(dbProxyName, targetGroupName, targetType, rdsResourceID string) string {
	parts := []string{dbProxyName, targetGroupName, targetType, rdsResourceID}
	id := strings.Join(parts, proxyTargetResourceIDSeparator)

	return id
}

func proxyTargetParseResourceID(id string) (string, string, string, string, error) {
	parts := strings.SplitN(id, proxyTargetResourceIDSeparator, 4)

	if len(parts) == 4 && parts[0] != "" && parts[1] != "" && parts[2] != "" && parts[3] != "" {
		return parts[0], parts[1], parts[2], parts[3], nil
	}

	return "", "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected DBPROXYNAME%[2]sTARGETGROUPNAME%[2]sTARGETTYPE%[2]sRDSRESOURCEID", id, proxyTargetResourceIDSeparator)
}

func findDBProxyTargetByFourPartKey(ctx context.Context, conn *rds.Client, dbProxyName, targetGroupName, targetType, rdsResourceID string) (*types.DBProxyTarget, error) {
	input := &rds.DescribeDBProxyTargetsInput{
		DBProxyName:     aws.String(dbProxyName),
		TargetGroupName: aws.String(targetGroupName),
	}

	return findDBProxyTarget(ctx, conn, input, func(v *types.DBProxyTarget) bool {
		return string(v.Type) == targetType && aws.ToString(v.RdsResourceId) == rdsResourceID
	})
}

func findDBProxyTarget(ctx context.Context, conn *rds.Client, input *rds.DescribeDBProxyTargetsInput, filter tfslices.Predicate[*types.DBProxyTarget]) (*types.DBProxyTarget, error) {
	output, err := findDBProxyTargets(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDBProxyTargets(ctx context.Context, conn *rds.Client, input *rds.DescribeDBProxyTargetsInput, filter tfslices.Predicate[*types.DBProxyTarget]) ([]types.DBProxyTarget, error) {
	var output []types.DBProxyTarget

	pages := rds.NewDescribeDBProxyTargetsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.DBProxyNotFoundFault](err) || errs.IsA[*types.DBProxyTargetGroupNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Targets {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
