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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_appmesh_virtual_gateway")
func DataSourceVirtualGateway() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceVirtualGatewayRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mesh_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"mesh_owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"resource_owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"spec":         dataSourcePropertyFromResourceProperty(resourceVirtualGatewaySpecSchema()),
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceVirtualGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	virtualGatewayName := d.Get("name").(string)
	virtualGateway, err := FindVirtualGatewayByThreePartKey(ctx, conn, d.Get("mesh_name").(string), d.Get("mesh_owner").(string), virtualGatewayName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading App Mesh Virtual Gateway (%s): %s", virtualGatewayName, err)
	}

	d.SetId(aws.StringValue(virtualGateway.VirtualGatewayName))
	arn := aws.StringValue(virtualGateway.Metadata.Arn)
	d.Set("arn", arn)
	d.Set("created_date", virtualGateway.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set("last_updated_date", virtualGateway.Metadata.LastUpdatedAt.Format(time.RFC3339))
	d.Set("mesh_name", virtualGateway.MeshName)
	meshOwner := aws.StringValue(virtualGateway.Metadata.MeshOwner)
	d.Set("mesh_owner", meshOwner)
	d.Set("name", virtualGateway.VirtualGatewayName)
	d.Set("resource_owner", virtualGateway.Metadata.ResourceOwner)
	if err := d.Set("spec", flattenVirtualGatewaySpec(virtualGateway.Spec)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting spec: %s", err)
	}

	// https://docs.aws.amazon.com/app-mesh/latest/userguide/sharing.html#sharing-permissions
	// Owners and consumers can list tags and can tag/untag resources in a mesh that the account created.
	// They can't list tags and tag/untag resources in a mesh that aren't created by the account.
	var tags tftags.KeyValueTags

	if meshOwner == meta.(*conns.AWSClient).AccountID {
		tags, err = listTags(ctx, conn, arn)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing tags for App Mesh Virtual Gateway (%s): %s", arn, err)
		}
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
