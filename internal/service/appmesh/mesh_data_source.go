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

// @SDKDataSource("aws_appmesh_mesh", name="Service Mesh")
// @Tags
// @Testing(serialize=true)
func dataSourceMesh() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceMeshRead,

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
				"spec":         sdkv2.DataSourcePropertyFromResourceProperty(resourceMeshSpecSchema()),
				names.AttrTags: tftags.TagsSchemaComputed(),
			}
		},
	}
}

func dataSourceMeshRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn(ctx)

	meshName := d.Get(names.AttrName).(string)
	mesh, err := findMeshByTwoPartKey(ctx, conn, meshName, d.Get("mesh_owner").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading App Mesh Service Mesh (%s): %s", meshName, err)
	}

	d.SetId(aws.StringValue(mesh.MeshName))
	arn := aws.StringValue(mesh.Metadata.Arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrCreatedDate, mesh.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set(names.AttrLastUpdatedDate, mesh.Metadata.LastUpdatedAt.Format(time.RFC3339))
	meshOwner := aws.StringValue(mesh.Metadata.MeshOwner)
	d.Set("mesh_owner", meshOwner)
	d.Set(names.AttrResourceOwner, mesh.Metadata.ResourceOwner)
	if err := d.Set("spec", flattenMeshSpec(mesh.Spec)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting spec: %s", err)
	}

	// https://docs.aws.amazon.com/app-mesh/latest/userguide/sharing.html#sharing-permissions
	// Owners and consumers can list tags and can tag/untag resources in a mesh that the account created.
	// They can't list tags and tag/untag resources in a mesh that aren't created by the account.
	var tags tftags.KeyValueTags

	if meshOwner == meta.(*conns.AWSClient).AccountID {
		tags, err = listTags(ctx, conn, arn)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing tags for App Mesh Service Mesh (%s): %s", arn, err)
		}
	}

	setKeyValueTagsOut(ctx, tags)

	return diags
}
