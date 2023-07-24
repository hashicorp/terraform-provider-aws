// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

// @SDKResource("aws_db_proxy_default_target_group")
func ResourceProxyDefaultTargetGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProxyDefaultTargetGroupCreate,
		ReadWithoutTimeout:   resourceProxyDefaultTargetGroupRead,
		UpdateWithoutTimeout: resourceProxyDefaultTargetGroupUpdate,
		DeleteWithoutTimeout: schema.NoopContext,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_proxy_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validIdentifier,
			},
			"name": {
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
							Set: schema.HashString,
						},
					},
				},
			},
		},
	}
}

func resourceProxyDefaultTargetGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	tg, err := resourceProxyDefaultTargetGroupGet(ctx, conn, d.Id())
	if err != nil {
		if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyNotFoundFault) {
			log.Printf("[WARN] DB Proxy (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Proxy (%s) Default Target Group: %s", d.Id(), err)
	}

	if tg == nil {
		log.Printf("[WARN] DB Proxy default target group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("arn", tg.TargetGroupArn)
	d.Set("db_proxy_name", tg.DBProxyName)
	d.Set("name", tg.TargetGroupName)

	cpc := tg.ConnectionPoolConfig
	d.Set("connection_pool_config", flattenProxyTargetGroupConnectionPoolConfig(cpc))

	return diags
}

func resourceProxyDefaultTargetGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId(d.Get("db_proxy_name").(string))
	return resourceProxyDefaultTargetGroupCreateUpdate(ctx, d, schema.TimeoutCreate, meta)
}

func resourceProxyDefaultTargetGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceProxyDefaultTargetGroupCreateUpdate(ctx, d, schema.TimeoutUpdate, meta)
}

func resourceProxyDefaultTargetGroupCreateUpdate(ctx context.Context, d *schema.ResourceData, timeout string, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	params := rds.ModifyDBProxyTargetGroupInput{
		DBProxyName:     aws.String(d.Get("db_proxy_name").(string)),
		TargetGroupName: aws.String("default"),
	}

	if v, ok := d.GetOk("connection_pool_config"); ok {
		params.ConnectionPoolConfig = expandProxyConnectionPoolConfig(v.([]interface{}))
	}

	if _, err := conn.ModifyDBProxyTargetGroupWithContext(ctx, &params); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating RDS DB Proxy (%s) default target group: %s", d.Id(), err)
	}

	stateChangeConf := &retry.StateChangeConf{
		Pending: []string{rds.DBProxyStatusModifying},
		Target:  []string{rds.DBProxyStatusAvailable},
		Refresh: resourceProxyDefaultTargetGroupRefreshFunc(ctx, conn, d.Id()),
		Timeout: d.Timeout(timeout),
	}

	if _, err := stateChangeConf.WaitForStateContext(ctx); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating RDS DB Proxy (%s) default target group: waiting for completion: %s", d.Id(), err)
	}

	return append(diags, resourceProxyDefaultTargetGroupRead(ctx, d, meta)...)
}

func expandProxyConnectionPoolConfig(configs []interface{}) *rds.ConnectionPoolConfiguration {
	if len(configs) < 1 {
		return nil
	}

	config := configs[0].(map[string]interface{})

	result := &rds.ConnectionPoolConfiguration{
		ConnectionBorrowTimeout:   aws.Int64(int64(config["connection_borrow_timeout"].(int))),
		InitQuery:                 aws.String(config["init_query"].(string)),
		MaxConnectionsPercent:     aws.Int64(int64(config["max_connections_percent"].(int))),
		MaxIdleConnectionsPercent: aws.Int64(int64(config["max_idle_connections_percent"].(int))),
		SessionPinningFilters:     flex.ExpandStringSet(config["session_pinning_filters"].(*schema.Set)),
	}

	return result
}

func flattenProxyTargetGroupConnectionPoolConfig(cpc *rds.ConnectionPoolConfigurationInfo) []interface{} {
	if cpc == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	m["connection_borrow_timeout"] = aws.Int64Value(cpc.ConnectionBorrowTimeout)
	m["init_query"] = aws.StringValue(cpc.InitQuery)
	m["max_connections_percent"] = aws.Int64Value(cpc.MaxConnectionsPercent)
	m["max_idle_connections_percent"] = aws.Int64Value(cpc.MaxIdleConnectionsPercent)
	m["session_pinning_filters"] = flex.FlattenStringSet(cpc.SessionPinningFilters)

	return []interface{}{m}
}

func resourceProxyDefaultTargetGroupGet(ctx context.Context, conn *rds.RDS, proxyName string) (*rds.DBProxyTargetGroup, error) {
	params := &rds.DescribeDBProxyTargetGroupsInput{
		DBProxyName: aws.String(proxyName),
	}

	var defaultTargetGroup *rds.DBProxyTargetGroup
	err := conn.DescribeDBProxyTargetGroupsPagesWithContext(ctx, params, func(page *rds.DescribeDBProxyTargetGroupsOutput, lastPage bool) bool {
		for _, targetGroup := range page.TargetGroups {
			if *targetGroup.IsDefault {
				defaultTargetGroup = targetGroup
				return false
			}
		}
		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	// Return default target group
	return defaultTargetGroup, nil
}

func resourceProxyDefaultTargetGroupRefreshFunc(ctx context.Context, conn *rds.RDS, proxyName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		tg, err := resourceProxyDefaultTargetGroupGet(ctx, conn, proxyName)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyNotFoundFault) {
				return 42, "", nil
			}
			return 42, "", err
		}

		return tg, *tg.Status, nil
	}
}
