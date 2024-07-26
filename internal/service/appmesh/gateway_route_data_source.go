// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appmesh

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_appmesh_gateway_route", name="Gateway Route")
// @Tags
// @Testing(serialize=true)
func dataSourceGatewayRoute() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceGatewayRouteRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrCreatedDate: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrLastUpdatedDate: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"mesh_name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"mesh_owner": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
				},
				names.AttrResourceOwner: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"spec":         sdkv2.DataSourcePropertyFromResourceProperty(resourceGatewayRouteSpecSchema()),
				names.AttrTags: tftags.TagsSchemaComputed(),
				"virtual_gateway_name": {
					Type:     schema.TypeString,
					Required: true,
				},
			}
		},
	}
}

func dataSourceGatewayRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn(ctx)

	gatewayRouteName := d.Get(names.AttrName).(string)
	gatewayRoute, err := findGatewayRouteByFourPartKey(ctx, conn, d.Get("mesh_name").(string), d.Get("mesh_owner").(string), d.Get("virtual_gateway_name").(string), gatewayRouteName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading App Mesh Gateway Route (%s): %s", gatewayRouteName, err)
	}

	d.SetId(aws.StringValue(gatewayRoute.GatewayRouteName))
	arn := aws.StringValue(gatewayRoute.Metadata.Arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrCreatedDate, gatewayRoute.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set(names.AttrLastUpdatedDate, gatewayRoute.Metadata.LastUpdatedAt.Format(time.RFC3339))
	d.Set("mesh_name", gatewayRoute.MeshName)
	meshOwner := aws.StringValue(gatewayRoute.Metadata.MeshOwner)
	d.Set("mesh_owner", meshOwner)
	d.Set(names.AttrName, gatewayRoute.GatewayRouteName)
	d.Set(names.AttrResourceOwner, gatewayRoute.Metadata.ResourceOwner)
	if err := d.Set("spec", flattenGatewayRouteSpec(gatewayRoute.Spec)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting spec: %s", err)
	}
	d.Set("virtual_gateway_name", gatewayRoute.VirtualGatewayName)

	// https://docs.aws.amazon.com/app-mesh/latest/userguide/sharing.html#sharing-permissions
	// Owners and consumers can list tags and can tag/untag resources in a mesh that the account created.
	// They can't list tags and tag/untag resources in a mesh that aren't created by the account.
	var tags tftags.KeyValueTags

	if meshOwner == meta.(*conns.AWSClient).AccountID {
		tags, err = listTags(ctx, conn, arn)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing tags for App Mesh Gateway Route (%s): %s", arn, err)
		}
	}

	setKeyValueTagsOut(ctx, tags)

	return diags
}
