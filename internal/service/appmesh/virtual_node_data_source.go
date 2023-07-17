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

// @SDKDataSource("aws_appmesh_virtual_node")
func DataSourceVirtualNode() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceVirtualNodeRead,

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
				Optional: true,
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
			"spec":         dataSourcePropertyFromResourceProperty(resourceVirtualNodeSpecSchema()),
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceVirtualNodeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	virtualNodeName := d.Get("name").(string)
	vn, err := FindVirtualNodeByThreePartKey(ctx, conn, d.Get("mesh_name").(string), d.Get("mesh_owner").(string), virtualNodeName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading App Mesh Virtual Node (%s): %s", virtualNodeName, err)
	}

	d.SetId(aws.StringValue(vn.VirtualNodeName))
	arn := aws.StringValue(vn.Metadata.Arn)
	d.Set("arn", arn)
	d.Set("created_date", vn.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set("last_updated_date", vn.Metadata.LastUpdatedAt.Format(time.RFC3339))
	d.Set("mesh_name", vn.MeshName)
	meshOwner := aws.StringValue(vn.Metadata.MeshOwner)
	d.Set("mesh_owner", meshOwner)
	d.Set("name", vn.VirtualNodeName)
	d.Set("resource_owner", vn.Metadata.ResourceOwner)
	if err := d.Set("spec", flattenVirtualNodeSpec(vn.Spec)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting spec: %s", err)
	}

	// https://docs.aws.amazon.com/app-mesh/latest/userguide/sharing.html#sharing-permissions
	// Owners and consumers can list tags and can tag/untag resources in a mesh that the account created.
	// They can't list tags and tag/untag resources in a mesh that aren't created by the account.
	var tags tftags.KeyValueTags

	if meshOwner == meta.(*conns.AWSClient).AccountID {
		tags, err = listTags(ctx, conn, arn)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing tags for App Mesh Virtual Node (%s): %s", arn, err)
		}
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
