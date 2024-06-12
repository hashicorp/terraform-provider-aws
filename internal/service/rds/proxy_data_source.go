// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_db_proxy", name="DB Proxy")
func dataSourceProxy() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceProxyRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auth": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auth_scheme": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"client_password_auth_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrDescription: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"iam_auth": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"secret_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrUsername: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"debug_logging": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrEndpoint: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_family": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"idle_client_timeout": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"require_tls": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrRoleARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVPCSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vpc_subnet_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceProxyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	name := d.Get(names.AttrName).(string)
	dbProxy, err := findDBProxyByName(ctx, conn, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Proxy (%s): %s", name, err)
	}

	d.SetId(name)
	d.Set(names.AttrARN, dbProxy.DBProxyArn)
	d.Set("auth", flattenUserAuthConfigInfos(dbProxy.Auth))
	d.Set("debug_logging", dbProxy.DebugLogging)
	d.Set(names.AttrEndpoint, dbProxy.Endpoint)
	d.Set("engine_family", dbProxy.EngineFamily)
	d.Set("idle_client_timeout", dbProxy.IdleClientTimeout)
	d.Set("require_tls", dbProxy.RequireTLS)
	d.Set(names.AttrRoleARN, dbProxy.RoleArn)
	d.Set(names.AttrVPCID, dbProxy.VpcId)
	d.Set(names.AttrVPCSecurityGroupIDs, dbProxy.VpcSecurityGroupIds)
	d.Set("vpc_subnet_ids", dbProxy.VpcSubnetIds)

	return diags
}
