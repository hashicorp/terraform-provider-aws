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

// @SDKResource("aws_exampletags_doohickey", name="Doohickey")
// @Tags(identifierAttribute="arn", resourceType="Doohickey")
func resourceDoohickey() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDoohickeyCreate,
		ReadWithoutTimeout:   resourceDoohickeyRead,
		UpdateWithoutTimeout: resourceDoohickeyUpdate,
		DeleteWithoutTimeout: resourceDoohickeyDelete,

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

func resourceDoohickeyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsClient(ctx)

	input := exampletags.CreateDoohickeyInput{
		Tags: getTagsIn(ctx),
	}

	output, err := conn.CreateDoohickey(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Doohickey: %s", err)
	}

	d.SetId(aws.ToString(output.DoohickeyId))

	return append(diags, resourceDoohickeyRead(ctx, d, meta)...)
}

func resourceDoohickeyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsClient(ctx)

	doohickey, err := findDoohickeyByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Doohickey (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Doohickey (%s): %s", d.Id(), err)
	}

	d.Set("arn", doohickey.DoohickeyArn)

	return diags
}

func resourceDoohickeyUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Tags only.
	return resourceDoohickeyRead(ctx, d, meta)
}

func resourceDoohickeyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsClient(ctx)

	log.Printf("[DEBUG] Deleting Doohickey: %s", d.Id())
	input := exampletags.DeleteDoohickeyInput{
		DoohickeyId: aws.String(d.Id()),
	}
	_, err := conn.DeleteDoohickey(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Doohickey (%s): %s", d.Id(), err)
	}

	return diags
}

func findDoohickeyByID(ctx context.Context, conn *exampletags.Client, id string) (*awstypes.Doohickey, error) {
	input := exampletags.DescribeDoohickeyInput{
		DoohickeyId: aws.String(id),
	}

	output, err := conn.DescribeDoohickey(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Doohickey == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Doohickey, nil
}
