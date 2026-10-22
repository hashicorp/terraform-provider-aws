// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package exampletags

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/exampletags"
	awstypes "github.com/aws/aws-sdk-go-v2/service/exampletags/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_exampletags_gizmo", name="Gizmo")
// @Tags(identifierAttribute="arn", resourceType="Gizmo")
func resourceGizmo() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGizmoCreate,
		ReadWithoutTimeout:   resourceGizmoRead,
		UpdateWithoutTimeout: resourceGizmoUpdate,
		DeleteWithoutTimeout: resourceGizmoDelete,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceGizmoCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsClient(ctx)

	input := exampletags.CreateGizmoInput{
		Tags: getTagsIn(ctx),
	}

	output, err := conn.CreateGizmo(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Gizmo: %s", err)
	}

	d.SetId(aws.ToString(output.GizmoId))

	return append(diags, resourceGizmoRead(ctx, d, meta)...)
}

func resourceGizmoRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsClient(ctx)

	gizmo, err := findGizmoByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Gizmo (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Gizmo (%s): %s", d.Id(), err)
	}

	d.Set("arn", gizmo.GizmoArn)

	return diags
}

func resourceGizmoUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Tags only.
	return resourceGizmoRead(ctx, d, meta)
}

func resourceGizmoDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsClient(ctx)

	log.Printf("[DEBUG] Deleting Gizmo: %s", d.Id())
	input := exampletags.DeleteGizmoInput{
		GizmoId: aws.String(d.Id()),
	}
	_, err := conn.DeleteGizmo(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Gizmo (%s): %s", d.Id(), err)
	}

	return diags
}

func findGizmoByID(ctx context.Context, conn *exampletags.Client, id string) (*awstypes.Gizmo, error) {
	input := exampletags.DescribeGizmoInput{
		GizmoId: aws.String(id),
	}

	output, err := conn.DescribeGizmo(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Gizmo == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Gizmo, nil
}
