// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package examplecrud

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/examplecrud"
	awstypes "github.com/aws/aws-sdk-go-v2/service/examplecrud/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_examplecrud_gizmo", name="Gizmo")
func resourceGizmo() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGizmoCreate,
		ReadWithoutTimeout:   resourceGizmoRead,
		DeleteWithoutTimeout: resourceGizmoDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceGizmoCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleCRUDClient(ctx)

	name := d.Get("name").(string)
	input := examplecrud.CreateGizmoInput{
		Name: aws.String(name),
	}

	output, err := conn.CreateGizmo(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Gizmo (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.GizmoId))

	return append(diags, resourceGizmoRead(ctx, d, meta)...)
}

func resourceGizmoRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleCRUDClient(ctx)

	input := examplecrud.GetGizmoInput{
		GizmoId: aws.String(d.Id()),
	}

	output, err := conn.GetGizmo(ctx, &input)

	if !d.IsNewResource() && errs.IsA[*awstypes.ResourceNotFoundException](err) {
		log.Printf("[WARN] Gizmo (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Gizmo (%s): %s", d.Id(), err)
	}

	d.Set("name", output.Name)

	return diags
}

func resourceGizmoDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleCRUDClient(ctx)

	log.Printf("[DEBUG] Deleting Gizmo: %s", d.Id())
	input := examplecrud.DeleteGizmoInput{
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
