// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

//go:build !generate

package examplebuildconstraints

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/examplebuildconstraints"
	awstypes "github.com/aws/aws-sdk-go-v2/service/examplebuildconstraints/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_examplebuildconstraints_included", name="Included")
func resourceIncluded() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIncludedCreate,
		ReadWithoutTimeout:   resourceIncludedRead,
		DeleteWithoutTimeout: resourceIncludedDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceIncludedCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleBuildConstraintsClient(ctx)

	name := d.Get("name").(string)
	input := examplebuildconstraints.CreateThingInput{
		Name: aws.String(name),
	}

	output, err := conn.CreateThing(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Thing (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ThingId))

	return append(diags, resourceIncludedRead(ctx, d, meta)...)
}

func resourceIncludedRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleBuildConstraintsClient(ctx)

	input := examplebuildconstraints.GetThingInput{
		ThingId: aws.String(d.Id()),
	}

	output, err := conn.GetThing(ctx, &input)

	if !d.IsNewResource() && errs.IsA[*awstypes.ResourceNotFoundException](err) {
		log.Printf("[WARN] Thing (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Thing (%s): %s", d.Id(), err)
	}

	d.Set("name", output.Name)

	return diags
}

func resourceIncludedDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleBuildConstraintsClient(ctx)

	log.Printf("[DEBUG] Deleting Thing: %s", d.Id())
	input := examplebuildconstraints.DeleteThingInput{
		ThingId: aws.String(d.Id()),
	}
	_, err := conn.DeleteThing(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Thing (%s): %s", d.Id(), err)
	}

	return diags
}
