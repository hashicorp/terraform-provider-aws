// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appmesh

import (
	"context"
	"fmt"
	"log"
	"strings"
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

// @SDKResource("aws_appmesh_virtual_service", name="Virtual Service")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go/service/appmesh;appmesh.VirtualServiceData")
// @Testing(serialize=true)
// @Testing(importStateIdFunc=testAccVirtualServiceImportStateIdFunc)
func resourceVirtualService() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVirtualServiceCreate,
		ReadWithoutTimeout:   resourceVirtualServiceRead,
		UpdateWithoutTimeout: resourceVirtualServiceUpdate,
		DeleteWithoutTimeout: resourceVirtualServiceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceVirtualServiceImport,
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
				"mesh_name": {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringLenBetween(1, 255),
				},
				"mesh_owner": {
					Type:         schema.TypeString,
					Optional:     true,
					Computed:     true,
					ForceNew:     true,
					ValidateFunc: verify.ValidAccountID,
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
				"spec":            resourceVirtualServiceSpecSchema(),
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
			}
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVirtualServiceSpecSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"provider": {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 0,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"virtual_node": {
								Type:          schema.TypeList,
								Optional:      true,
								MinItems:      0,
								MaxItems:      1,
								ConflictsWith: []string{"spec.0.provider.0.virtual_router"},
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"virtual_node_name": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringLenBetween(1, 255),
										},
									},
								},
							},
							"virtual_router": {
								Type:          schema.TypeList,
								Optional:      true,
								MinItems:      0,
								MaxItems:      1,
								ConflictsWith: []string{"spec.0.provider.0.virtual_node"},
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"virtual_router_name": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringLenBetween(1, 255),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceVirtualServiceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn(ctx)

	name := d.Get(names.AttrName).(string)
	input := &appmesh.CreateVirtualServiceInput{
		MeshName:           aws.String(d.Get("mesh_name").(string)),
		Spec:               expandVirtualServiceSpec(d.Get("spec").([]interface{})),
		Tags:               getTagsIn(ctx),
		VirtualServiceName: aws.String(name),
	}

	if v, ok := d.GetOk("mesh_owner"); ok {
		input.MeshOwner = aws.String(v.(string))
	}

	output, err := conn.CreateVirtualServiceWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating App Mesh Virtual Service (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.VirtualService.Metadata.Uid))

	return append(diags, resourceVirtualServiceRead(ctx, d, meta)...)
}

func resourceVirtualServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return findVirtualServiceByThreePartKey(ctx, conn, d.Get("mesh_name").(string), d.Get("mesh_owner").(string), d.Get(names.AttrName).(string))
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] App Mesh Virtual Service (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading App Mesh Virtual Service (%s): %s", d.Id(), err)
	}

	vs := outputRaw.(*appmesh.VirtualServiceData)

	arn := aws.StringValue(vs.Metadata.Arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrCreatedDate, vs.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set(names.AttrLastUpdatedDate, vs.Metadata.LastUpdatedAt.Format(time.RFC3339))
	d.Set("mesh_name", vs.MeshName)
	d.Set("mesh_owner", vs.Metadata.MeshOwner)
	d.Set(names.AttrName, vs.VirtualServiceName)
	d.Set(names.AttrResourceOwner, vs.Metadata.ResourceOwner)
	if err := d.Set("spec", flattenVirtualServiceSpec(vs.Spec)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting spec: %s", err)
	}

	return diags
}

func resourceVirtualServiceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn(ctx)

	if d.HasChange("spec") {
		input := &appmesh.UpdateVirtualServiceInput{
			MeshName:           aws.String(d.Get("mesh_name").(string)),
			Spec:               expandVirtualServiceSpec(d.Get("spec").([]interface{})),
			VirtualServiceName: aws.String(d.Get(names.AttrName).(string)),
		}

		if v, ok := d.GetOk("mesh_owner"); ok {
			input.MeshOwner = aws.String(v.(string))
		}

		_, err := conn.UpdateVirtualServiceWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating App Mesh Virtual Service (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceVirtualServiceRead(ctx, d, meta)...)
}

func resourceVirtualServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn(ctx)

	log.Printf("[DEBUG] Deleting App Mesh Virtual Service: %s", d.Id())
	input := &appmesh.DeleteVirtualServiceInput{
		MeshName:           aws.String(d.Get("mesh_name").(string)),
		VirtualServiceName: aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk("mesh_owner"); ok {
		input.MeshOwner = aws.String(v.(string))
	}

	_, err := conn.DeleteVirtualServiceWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting App Mesh Virtual Service (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceVirtualServiceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of import ID (%s), use: 'mesh-name/virtual-service-name'", d.Id())
	}

	conn := meta.(*conns.AWSClient).AppMeshConn(ctx)
	meshName := parts[0]
	name := parts[1]

	vs, err := findVirtualServiceByThreePartKey(ctx, conn, meshName, "", name)

	if err != nil {
		return nil, err
	}

	d.SetId(aws.StringValue(vs.Metadata.Uid))
	d.Set("mesh_name", vs.MeshName)
	d.Set(names.AttrName, vs.VirtualServiceName)

	return []*schema.ResourceData{d}, nil
}

func findVirtualServiceByThreePartKey(ctx context.Context, conn *appmesh.AppMesh, meshName, meshOwner, name string) (*appmesh.VirtualServiceData, error) {
	input := &appmesh.DescribeVirtualServiceInput{
		MeshName:           aws.String(meshName),
		VirtualServiceName: aws.String(name),
	}
	if meshOwner != "" {
		input.MeshOwner = aws.String(meshOwner)
	}

	output, err := findVirtualService(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := aws.StringValue(output.Status.Status); status == appmesh.VirtualServiceStatusCodeDeleted {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output, nil
}

func findVirtualService(ctx context.Context, conn *appmesh.AppMesh, input *appmesh.DescribeVirtualServiceInput) (*appmesh.VirtualServiceData, error) {
	output, err := conn.DescribeVirtualServiceWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.VirtualService == nil || output.VirtualService.Metadata == nil || output.VirtualService.Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.VirtualService, nil
}
