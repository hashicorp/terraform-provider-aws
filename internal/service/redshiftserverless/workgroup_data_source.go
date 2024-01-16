// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_redshiftserverless_workgroup")
func DataSourceWorkgroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceWorkgroupRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"vpc_endpoint": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"network_interface": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"availability_zone": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"network_interface_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"private_ip_address": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"subnet_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
									"vpc_endpoint_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"vpc_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"enhanced_vpc_routing": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"namespace_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"publicly_accessible": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"workgroup_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workgroup_name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceWorkgroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn(ctx)

	workgroupName := d.Get("workgroup_name").(string)

	resource, err := FindWorkgroupByName(ctx, conn, workgroupName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Serverless Workgroup (%s): %s", workgroupName, err)
	}

	d.SetId(workgroupName)
	d.Set("arn", resource.WorkgroupArn)
	if err := d.Set("endpoint", []interface{}{flattenEndpoint(resource.Endpoint)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting endpoint: %s", err)
	}
	d.Set("enhanced_vpc_routing", resource.EnhancedVpcRouting)
	d.Set("namespace_name", resource.NamespaceName)
	d.Set("publicly_accessible", resource.PubliclyAccessible)
	d.Set("security_group_ids", resource.SecurityGroupIds)
	d.Set("subnet_ids", resource.SubnetIds)
	d.Set("workgroup_id", resource.WorkgroupId)

	return diags
}
