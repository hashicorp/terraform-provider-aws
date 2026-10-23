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

// @SDKResource("aws_exampletags_baz", name="Baz")
// @Tags(identifierAttribute="arn", resourceType="Baz")
func resourceBaz() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBazCreate,
		ReadWithoutTimeout:   resourceBazRead,
		UpdateWithoutTimeout: resourceBazUpdate,
		DeleteWithoutTimeout: resourceBazDelete,

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

func resourceBazCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsClient(ctx)

	input := exampletags.CreateBazInput{
		Tags: getTagsIn(ctx),
	}

	output, err := conn.CreateBaz(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Baz: %s", err)
	}

	d.SetId(aws.ToString(output.BazId))

	return append(diags, resourceBazRead(ctx, d, meta)...)
}

func resourceBazRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsClient(ctx)

	baz, err := findBazByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Baz (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Baz (%s): %s", d.Id(), err)
	}

	d.Set("arn", baz.BazArn)

	return diags
}

func resourceBazUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Tags only.
	return resourceBazRead(ctx, d, meta)
}

func resourceBazDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsClient(ctx)

	log.Printf("[DEBUG] Deleting Baz: %s", d.Id())
	input := exampletags.DeleteBazInput{
		BazId: aws.String(d.Id()),
	}
	_, err := conn.DeleteBaz(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Baz (%s): %s", d.Id(), err)
	}

	return diags
}

func findBazByID(ctx context.Context, conn *exampletags.Client, id string) (*awstypes.Baz, error) {
	input := exampletags.DescribeBazInput{
		BazId: aws.String(id),
	}

	output, err := conn.DescribeBaz(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Baz == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Baz, nil
}
