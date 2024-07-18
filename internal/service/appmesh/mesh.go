// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appmesh

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appmesh_mesh", name="Service Mesh")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go/service/appmesh;appmesh.MeshData")
// @Testing(serialize=true)
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
					Computed: true,
				},
				names.AttrName: {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringLenBetween(1, 255),
				},
				names.AttrResourceOwner: {
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
							names.AttrType: {
								Type:         schema.TypeString,
								Optional:     true,
								Default:      appmesh.EgressFilterTypeDropAll,
								ValidateFunc: validation.StringInSlice(appmesh.EgressFilterType_Values(), false),
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
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringInSlice(appmesh.IpPreference_Values(), false),
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
	conn := meta.(*conns.AWSClient).AppMeshConn(ctx)

	name := d.Get(names.AttrName).(string)
	input := &appmesh.CreateMeshInput{
		MeshName: aws.String(name),
		Spec:     expandMeshSpec(d.Get("spec").([]interface{})),
		Tags:     getTagsIn(ctx),
	}

	_, err := conn.CreateMeshWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating App Mesh Service Mesh (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceMeshRead(ctx, d, meta)...)
}

func resourceMeshRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn(ctx)

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

	mesh := outputRaw.(*appmesh.MeshData)
	arn := aws.StringValue(mesh.Metadata.Arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrCreatedDate, mesh.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set(names.AttrLastUpdatedDate, mesh.Metadata.LastUpdatedAt.Format(time.RFC3339))
	d.Set("mesh_owner", mesh.Metadata.MeshOwner)
	d.Set(names.AttrName, mesh.MeshName)
	d.Set(names.AttrResourceOwner, mesh.Metadata.ResourceOwner)
	if err := d.Set("spec", flattenMeshSpec(mesh.Spec)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting spec: %s", err)
	}

	return diags
}

func resourceMeshUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn(ctx)

	if d.HasChange("spec") {
		input := &appmesh.UpdateMeshInput{
			MeshName: aws.String(d.Id()),
			Spec:     expandMeshSpec(d.Get("spec").([]interface{})),
		}

		_, err := conn.UpdateMeshWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating App Mesh Service Mesh (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceMeshRead(ctx, d, meta)...)
}

func resourceMeshDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn(ctx)

	log.Printf("[DEBUG] Deleting App Mesh Service Mesh: %s", d.Id())
	_, err := conn.DeleteMeshWithContext(ctx, &appmesh.DeleteMeshInput{
		MeshName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting App Mesh Service Mesh (%s): %s", d.Id(), err)
	}

	return diags
}

func findMeshByTwoPartKey(ctx context.Context, conn *appmesh.AppMesh, name, owner string) (*appmesh.MeshData, error) {
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

	if status := aws.StringValue(output.Status.Status); status == appmesh.MeshStatusCodeDeleted {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output, nil
}

func findMesh(ctx context.Context, conn *appmesh.AppMesh, input *appmesh.DescribeMeshInput) (*appmesh.MeshData, error) {
	output, err := conn.DescribeMeshWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
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
