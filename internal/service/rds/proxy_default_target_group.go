// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_db_proxy_default_target_group", name="DB Proxy Default Target Group")
func resourceProxyDefaultTargetGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProxyDefaultTargetGroupPut,
		ReadWithoutTimeout:   resourceProxyDefaultTargetGroupRead,
		UpdateWithoutTimeout: resourceProxyDefaultTargetGroupPut,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_pool_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connection_borrow_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      120,
							ValidateFunc: validation.IntBetween(0, 3600),
						},
						"init_query": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"max_connections_percent": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      100,
							ValidateFunc: validation.IntBetween(1, 100),
						},
						"max_idle_connections_percent": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      50,
							ValidateFunc: validation.IntBetween(0, 100),
						},
						"session_pinning_filters": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								// This isn't available as a constant
								ValidateFunc: validation.StringInSlice([]string{
									"EXCLUDE_VARIABLE_SETS",
								}, false),
							},
						},
					},
				},
			},
			"db_proxy_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validIdentifier,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceProxyDefaultTargetGroupPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	dbProxyName := d.Get("db_proxy_name").(string)
	input := &rds.ModifyDBProxyTargetGroupInput{
		DBProxyName:     aws.String(dbProxyName),
		TargetGroupName: aws.String("default"),
	}

	if v, ok := d.GetOk("connection_pool_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ConnectionPoolConfig = expandConnectionPoolConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	_, err := conn.ModifyDBProxyTargetGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating RDS DB Proxy Default Target Group (%s): %s", dbProxyName, err)
	}

	timeout := d.Timeout(schema.TimeoutUpdate)
	if d.IsNewResource() {
		timeout = d.Timeout(schema.TimeoutCreate)

		d.SetId(dbProxyName)
	}

	if _, err := waitDefaultDBProxyTargetGroupAvailable(ctx, conn, dbProxyName, timeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Proxy Default Target Group (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceProxyDefaultTargetGroupRead(ctx, d, meta)...)
}

func resourceProxyDefaultTargetGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	tg, err := findDefaultDBProxyTargetGroupByDBProxyName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS DB Proxy Default Target Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Proxy Default Target Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, tg.TargetGroupArn)
	if tg.ConnectionPoolConfig != nil {
		if err := d.Set("connection_pool_config", []interface{}{flattenConnectionPoolConfigurationInfo(tg.ConnectionPoolConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting connection_pool_config: %s", err)
		}
	} else {
		d.Set("connection_pool_config", nil)
	}
	d.Set("db_proxy_name", tg.DBProxyName)
	d.Set(names.AttrName, tg.TargetGroupName)

	return diags
}

func findDefaultDBProxyTargetGroupByDBProxyName(ctx context.Context, conn *rds.Client, dbProxyName string) (*types.DBProxyTargetGroup, error) {
	input := &rds.DescribeDBProxyTargetGroupsInput{
		DBProxyName: aws.String(dbProxyName),
	}

	return findDBProxyTargetGroup(ctx, conn, input, func(v *types.DBProxyTargetGroup) bool {
		return aws.ToBool(v.IsDefault)
	})
}

func findDBProxyTargetGroup(ctx context.Context, conn *rds.Client, input *rds.DescribeDBProxyTargetGroupsInput, filter tfslices.Predicate[*types.DBProxyTargetGroup]) (*types.DBProxyTargetGroup, error) {
	output, err := findDBProxyTargetGroups(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDBProxyTargetGroups(ctx context.Context, conn *rds.Client, input *rds.DescribeDBProxyTargetGroupsInput, filter tfslices.Predicate[*types.DBProxyTargetGroup]) ([]types.DBProxyTargetGroup, error) {
	var output []types.DBProxyTargetGroup

	pages := rds.NewDescribeDBProxyTargetGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.DBProxyNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.TargetGroups {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusDefaultDBProxyTargetGroup(ctx context.Context, conn *rds.Client, dbProxyName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDefaultDBProxyTargetGroupByDBProxyName(ctx, conn, dbProxyName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

func waitDefaultDBProxyTargetGroupAvailable(ctx context.Context, conn *rds.Client, dbProxyName string, timeout time.Duration) (*types.DBProxyTargetGroup, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.DBProxyStatusModifying),
		Target:  enum.Slice(types.DBProxyStatusAvailable),
		Refresh: statusDefaultDBProxyTargetGroup(ctx, conn, dbProxyName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DBProxyTargetGroup); ok {
		return output, err
	}

	return nil, err
}

func expandConnectionPoolConfiguration(tfMap map[string]interface{}) *types.ConnectionPoolConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ConnectionPoolConfiguration{
		ConnectionBorrowTimeout:   aws.Int32(int32(tfMap["connection_borrow_timeout"].(int))),
		MaxConnectionsPercent:     aws.Int32(int32(tfMap["max_connections_percent"].(int))),
		MaxIdleConnectionsPercent: aws.Int32(int32(tfMap["max_idle_connections_percent"].(int))),
	}

	if v, ok := tfMap["init_query"].(string); ok && v != "" {
		apiObject.InitQuery = aws.String(v)
	}

	if v, ok := tfMap["session_pinning_filters"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SessionPinningFilters = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func flattenConnectionPoolConfigurationInfo(apiObject *types.ConnectionPoolConfigurationInfo) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["connection_borrow_timeout"] = aws.ToInt32(apiObject.ConnectionBorrowTimeout)
	tfMap["init_query"] = aws.ToString(apiObject.InitQuery)
	tfMap["max_connections_percent"] = aws.ToInt32(apiObject.MaxConnectionsPercent)
	tfMap["max_idle_connections_percent"] = aws.ToInt32(apiObject.MaxIdleConnectionsPercent)
	tfMap["session_pinning_filters"] = apiObject.SessionPinningFilters

	return tfMap
}
