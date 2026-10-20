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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKResource("aws_exampletags_without_identifier", name="Without Identifier")
// @Tags
func resourceWithoutIdentifier() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWithoutIdentifierCreate,
		ReadWithoutTimeout:   resourceWithoutIdentifierRead,
		DeleteWithoutTimeout: resourceWithoutIdentifierDelete,

		Schema: map[string]*schema.Schema{
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceWithoutIdentifierCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsClient(ctx)

	input := exampletags.CreateSprocketInput{
		Tags: getTagsIn(ctx),
	}

	output, err := conn.CreateSprocket(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Sprocket: %s", err)
	}

	d.SetId(aws.ToString(output.SprocketId))

	return append(diags, resourceWithoutIdentifierRead(ctx, d, meta)...)
}

func resourceWithoutIdentifierRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsClient(ctx)

	input := exampletags.GetSprocketInput{
		SprocketId: aws.String(d.Id()),
	}

	output, err := conn.GetSprocket(ctx, &input)

	if !d.IsNewResource() && errs.IsA[*awstypes.ResourceNotFoundException](err) {
		log.Printf("[WARN] Sprocket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Sprocket (%s): %s", d.Id(), err)
	}

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceWithoutIdentifierDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsClient(ctx)

	log.Printf("[DEBUG] Deleting Sprocket: %s", d.Id())
	input := exampletags.DeleteSprocketInput{
		SprocketId: aws.String(d.Id()),
	}
	_, err := conn.DeleteSprocket(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Sprocket (%s): %s", d.Id(), err)
	}

	return diags
}
