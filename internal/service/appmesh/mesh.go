// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appmesh

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appmesh"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appmesh/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appmesh_mesh", name="Service Mesh")
// @Tags(identifierAttribute="arn")
func resourceMesh() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMeshCreate,
		ReadWithoutTimeout:   resourceMeshRead,
		UpdateWithoutTimeout: resourceMeshUpdate,
		DeleteWithoutTimeout: resourceMeshDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
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
				"mesh_owner": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"name": {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringLenBetween(1, 255),
				},
				"resource_owner": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"spec":            resourceMeshSpecSchema(),
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
			}
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceMeshSpecSchema() *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeList,
		Optional:         true,
		MinItems:         0,
		MaxItems:         1,
		DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"egress_filter": {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 0,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"type": {
								Type:             schema.TypeString,
								Optional:         true,
								Default:          awstypes.EgressFilterTypeDropAll,
								ValidateDiagFunc: enum.Validate[awstypes.EgressFilterType](),
							},
						},
					},
				},
				"service_discovery": {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 0,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"ip_preference": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[awstypes.IpPreference](),
							},
						},
					},
				},
			},
		},
	}
}

func resourceMeshCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshClient(ctx)

	name := d.Get("name").(string)
	input := &appmesh.CreateMeshInput{
		MeshName: aws.String(name),
		Spec:     expandMeshSpec(d.Get("spec").([]interface{})),
		Tags:     getTagsIn(ctx),
	}

	_, err := conn.CreateMesh(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating App Mesh Service Mesh (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceMeshRead(ctx, d, meta)...)
}

func resourceMeshRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshClient(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return findMeshByTwoPartKey(ctx, conn, d.Id(), d.Get("mesh_owner").(string))
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] App Mesh Service Mesh (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading App Mesh Service Mesh (%s): %s", d.Id(), err)
	}

	mesh := outputRaw.(*awstypes.MeshData)
	arn := aws.ToString(mesh.Metadata.Arn)
	d.Set("arn", arn)
	d.Set("created_date", mesh.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set("last_updated_date", mesh.Metadata.LastUpdatedAt.Format(time.RFC3339))
	d.Set("mesh_owner", mesh.Metadata.MeshOwner)
	d.Set("name", mesh.MeshName)
	d.Set("resource_owner", mesh.Metadata.ResourceOwner)
	if err := d.Set("spec", flattenMeshSpec(mesh.Spec)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting spec: %s", err)
	}

	return diags
}

func resourceMeshUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshClient(ctx)

	if d.HasChange("spec") {
		input := &appmesh.UpdateMeshInput{
			MeshName: aws.String(d.Id()),
			Spec:     expandMeshSpec(d.Get("spec").([]interface{})),
		}

		_, err := conn.UpdateMesh(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating App Mesh Service Mesh (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceMeshRead(ctx, d, meta)...)
}

func resourceMeshDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshClient(ctx)

	log.Printf("[DEBUG] Deleting App Mesh Service Mesh: %s", d.Id())
	_, err := conn.DeleteMesh(ctx, &appmesh.DeleteMeshInput{
		MeshName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting App Mesh Service Mesh (%s): %s", d.Id(), err)
	}

	return diags
}

func findMeshByTwoPartKey(ctx context.Context, conn *appmesh.Client, name, owner string) (*awstypes.MeshData, error) {
	input := &appmesh.DescribeMeshInput{
		MeshName: aws.String(name),
	}
	if owner != "" {
		input.MeshOwner = aws.String(owner)
	}

	output, err := findMesh(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if output.Status.Status == awstypes.MeshStatusCodeDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(output.Status.Status),
			LastRequest: input,
		}
	}

	return output, nil
}

func findMesh(ctx context.Context, conn *appmesh.Client, input *appmesh.DescribeMeshInput) (*awstypes.MeshData, error) {
	output, err := conn.DescribeMesh(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Mesh == nil || output.Mesh.Metadata == nil || output.Mesh.Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Mesh, nil
}
