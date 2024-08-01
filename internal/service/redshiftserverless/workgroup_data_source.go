// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_redshiftserverless_workgroup", name="Workgroup")
func dataSourceWorkgroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceWorkgroupRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEndpoint: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAddress: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrPort: {
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
												names.AttrAvailabilityZone: {
													Type:     schema.TypeString,
													Computed: true,
												},
												names.AttrNetworkInterfaceID: {
													Type:     schema.TypeString,
													Computed: true,
												},
												"private_ip_address": {
													Type:     schema.TypeString,
													Computed: true,
												},
												names.AttrSubnetID: {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
									names.AttrVPCEndpointID: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrVPCID: {
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
			names.AttrPubliclyAccessible: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			names.AttrSubnetIDs: {
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
	d.Set(names.AttrARN, resource.WorkgroupArn)
	if err := d.Set(names.AttrEndpoint, []interface{}{flattenEndpoint(resource.Endpoint)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting endpoint: %s", err)
	}
	d.Set("enhanced_vpc_routing", resource.EnhancedVpcRouting)
	d.Set("namespace_name", resource.NamespaceName)
	d.Set(names.AttrPubliclyAccessible, resource.PubliclyAccessible)
	d.Set(names.AttrSecurityGroupIDs, resource.SecurityGroupIds)
	d.Set(names.AttrSubnetIDs, resource.SubnetIds)
	d.Set("workgroup_id", resource.WorkgroupId)

	return diags
}
