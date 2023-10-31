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
func ResourceMesh() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMeshCreate,
		ReadWithoutTimeout:   resourceMeshRead,
		UpdateWithoutTimeout: resourceMeshUpdate,
		DeleteWithoutTimeout: resourceMeshDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

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
								Type:         schema.TypeString,
								Optional:     true,
								Default:      appmesh.EgressFilterTypeDropAll,
								ValidateFunc: validation.StringInSlice(appmesh.EgressFilterType_Values(), false),
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

	name := d.Get("name").(string)
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
		return FindMeshByTwoPartKey(ctx, conn, d.Id(), d.Get("mesh_owner").(string))
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

func FindMeshByTwoPartKey(ctx context.Context, conn *appmesh.AppMesh, name, owner string) (*appmesh.MeshData, error) {
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

// Adapted from https://github.com/hashicorp/terraform-provider-google/google/datasource_helpers.go. Thanks!
// TODO Move to a shared package.

// dataSourceSchemaFromResourceSchema is a recursive func that
// converts an existing Resource schema to a Datasource schema.
// All schema elements are copied, but certain attributes are ignored or changed:
// - all attributes have Computed = true
// - all attributes have ForceNew, Required = false
// - Validation funcs and attributes (e.g. MaxItems) are not copied
func dataSourceSchemaFromResourceSchema(rs map[string]*schema.Schema) map[string]*schema.Schema {
	ds := make(map[string]*schema.Schema, len(rs))

	for k, v := range rs {
		ds[k] = dataSourcePropertyFromResourceProperty(v)
	}

	return ds
}

func dataSourcePropertyFromResourceProperty(rs *schema.Schema) *schema.Schema {
	ds := &schema.Schema{
		Computed:    true,
		Description: rs.Description,
		Type:        rs.Type,
	}

	switch rs.Type {
	case schema.TypeSet:
		ds.Set = rs.Set
		fallthrough
	case schema.TypeList, schema.TypeMap:
		// List & Set types are generally used for 2 cases:
		// - a list/set of simple primitive values (e.g. list of strings)
		// - a sub resource
		// Maps are usually used for maps of simple primitives
		switch elem := rs.Elem.(type) {
		case *schema.Resource:
			// handle the case where the Element is a sub-resource
			ds.Elem = &schema.Resource{
				Schema: dataSourceSchemaFromResourceSchema(elem.Schema),
			}
		case *schema.Schema:
			// handle simple primitive case
			ds.Elem = &schema.Schema{Type: elem.Type}
		}
	}

	return ds
}
