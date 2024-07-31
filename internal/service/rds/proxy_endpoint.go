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
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_db_proxy_endpoint", name="DB Proxy Endpoint")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func resourceProxyEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProxyEndpointCreate,
		ReadWithoutTimeout:   resourceProxyEndpointRead,
		DeleteWithoutTimeout: resourceProxyEndpointDelete,
		UpdateWithoutTimeout: resourceProxyEndpointUpdate,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_proxy_endpoint_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validIdentifier,
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
			"is_default": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"target_role": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          types.DBProxyEndpointTargetRoleReadWrite,
				ValidateDiagFunc: enum.Validate[types.DBProxyEndpointTargetRole](),
			},
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVPCSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vpc_subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceProxyEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	dbProxyName, dbProxyEndpointName := d.Get("db_proxy_name").(string), d.Get("db_proxy_endpoint_name").(string)
	id := proxyEndpointCreateResourceID(dbProxyName, dbProxyEndpointName)
	input := &rds.CreateDBProxyEndpointInput{
		DBProxyName:         aws.String(dbProxyName),
		DBProxyEndpointName: aws.String(dbProxyEndpointName),
		Tags:                getTagsInV2(ctx),
		TargetRole:          types.DBProxyEndpointTargetRole(d.Get("target_role").(string)),
		VpcSubnetIds:        flex.ExpandStringValueSet(d.Get("vpc_subnet_ids").(*schema.Set)),
	}

	if v, ok := d.GetOk(names.AttrVPCSecurityGroupIDs); ok && v.(*schema.Set).Len() > 0 {
		input.VpcSecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	_, err := conn.CreateDBProxyEndpoint(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS DB Proxy Endpoint (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitDBProxyEndpointAvailable(ctx, conn, dbProxyName, dbProxyEndpointName, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Proxy Endpoint (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceProxyEndpointRead(ctx, d, meta)...)
}

func resourceProxyEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	dbProxyName, dbProxyEndpointName, err := proxyEndpointParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	dbProxyEndpoint, err := findDBProxyEndpointByTwoPartKey(ctx, conn, dbProxyName, dbProxyEndpointName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS DB Proxy Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Proxy Endpoint (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, dbProxyEndpoint.DBProxyEndpointArn)
	d.Set("db_proxy_endpoint_name", dbProxyEndpoint.DBProxyEndpointName)
	d.Set("db_proxy_name", dbProxyEndpoint.DBProxyName)
	d.Set(names.AttrEndpoint, dbProxyEndpoint.Endpoint)
	d.Set("is_default", dbProxyEndpoint.IsDefault)
	d.Set("target_role", dbProxyEndpoint.TargetRole)
	d.Set(names.AttrVPCID, dbProxyEndpoint.VpcId)
	d.Set(names.AttrVPCSecurityGroupIDs, dbProxyEndpoint.VpcSecurityGroupIds)
	d.Set("vpc_subnet_ids", dbProxyEndpoint.VpcSubnetIds)

	return diags
}

func resourceProxyEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	dbProxyName, dbProxyEndpointName, err := proxyEndpointParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChange(names.AttrVPCSecurityGroupIDs) {
		input := &rds.ModifyDBProxyEndpointInput{
			DBProxyEndpointName: aws.String(dbProxyEndpointName),
			VpcSecurityGroupIds: flex.ExpandStringValueSet(d.Get(names.AttrVPCSecurityGroupIDs).(*schema.Set)),
		}

		_, err := conn.ModifyDBProxyEndpoint(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RDS DB Proxy Endpoint (%s): %s", d.Id(), err)
		}

		if _, err := waitDBProxyEndpointAvailable(ctx, conn, dbProxyName, dbProxyEndpointName, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Proxy Endpoint (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceProxyEndpointRead(ctx, d, meta)...)
}

func resourceProxyEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	dbProxyName, dbProxyEndpointName, err := proxyEndpointParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting RDS DB Proxy Endpoint: %s", d.Id())
	_, err = conn.DeleteDBProxyEndpoint(ctx, &rds.DeleteDBProxyEndpointInput{
		DBProxyEndpointName: aws.String(dbProxyEndpointName),
	})

	if errs.IsA[*types.DBProxyNotFoundFault](err) || errs.IsA[*types.DBProxyEndpointNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS DB Proxy Endpoint (%s): %s", d.Id(), err)
	}

	if _, err := waitDBProxyEndpointDeleted(ctx, conn, dbProxyName, dbProxyEndpointName, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Proxy Endpoint (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const proxyEndpointResourceIDSeparator = "/"

func proxyEndpointCreateResourceID(dbProxyName, dbProxyEndpointName string) string {
	parts := []string{dbProxyName, dbProxyEndpointName}
	id := strings.Join(parts, proxyEndpointResourceIDSeparator)

	return id
}

func proxyEndpointParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, proxyEndpointResourceIDSeparator, 2)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected DBPROXYNAME%[2]sDBPROXYENDPOINTNAME", id, proxyEndpointResourceIDSeparator)
}

func findDBProxyEndpointByTwoPartKey(ctx context.Context, conn *rds.Client, dbProxyName, dbProxyEndpointName string) (*types.DBProxyEndpoint, error) {
	input := &rds.DescribeDBProxyEndpointsInput{
		DBProxyName:         aws.String(dbProxyName),
		DBProxyEndpointName: aws.String(dbProxyEndpointName),
	}
	output, err := findDBProxyEndpoint(ctx, conn, input, tfslices.PredicateTrue[*types.DBProxyEndpoint]())

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.DBProxyName) != dbProxyName || aws.ToString(output.DBProxyEndpointName) != dbProxyEndpointName {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findDBProxyEndpoint(ctx context.Context, conn *rds.Client, input *rds.DescribeDBProxyEndpointsInput, filter tfslices.Predicate[*types.DBProxyEndpoint]) (*types.DBProxyEndpoint, error) {
	output, err := findDBProxyEndpoints(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDBProxyEndpoints(ctx context.Context, conn *rds.Client, input *rds.DescribeDBProxyEndpointsInput, filter tfslices.Predicate[*types.DBProxyEndpoint]) ([]types.DBProxyEndpoint, error) {
	var output []types.DBProxyEndpoint

	pages := rds.NewDescribeDBProxyEndpointsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.DBProxyNotFoundFault](err) || errs.IsA[*types.DBProxyEndpointNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.DBProxyEndpoints {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusDBProxyEndpoint(ctx context.Context, conn *rds.Client, dbProxyName, dbProxyEndpointName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDBProxyEndpointByTwoPartKey(ctx, conn, dbProxyName, dbProxyEndpointName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitDBProxyEndpointAvailable(ctx context.Context, conn *rds.Client, dbProxyName, dbProxyEndpointName string, timeout time.Duration) (*types.DBProxyEndpoint, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.DBProxyEndpointStatusCreating, types.DBProxyEndpointStatusModifying),
		Target:  enum.Slice(types.DBProxyEndpointStatusAvailable),
		Refresh: statusDBProxyEndpoint(ctx, conn, dbProxyName, dbProxyEndpointName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DBProxyEndpoint); ok {
		return output, err
	}

	return nil, err
}

func waitDBProxyEndpointDeleted(ctx context.Context, conn *rds.Client, dbProxyName, dbProxyEndpointName string, timeout time.Duration) (*types.DBProxyEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.DBProxyEndpointStatusDeleting),
		Target:  []string{},
		Refresh: statusDBProxyEndpoint(ctx, conn, dbProxyName, dbProxyEndpointName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DBProxyEndpoint); ok {
		return output, err
	}

	return nil, err
}
