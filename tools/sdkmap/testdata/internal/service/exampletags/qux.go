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

// @SDKResource("aws_exampletags_qux", name="Qux")
// @Tags(identifierAttribute="arn", resourceType="Qux")
func resourceQux() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceQuxCreate,
		ReadWithoutTimeout:   resourceQuxRead,
		UpdateWithoutTimeout: resourceQuxUpdate,
		DeleteWithoutTimeout: resourceQuxDelete,

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

func resourceQuxCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsClient(ctx)

	input := exampletags.CreateQuxInput{
		Tags: getTagsIn(ctx),
	}

	output, err := conn.CreateQux(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Qux: %s", err)
	}

	d.SetId(aws.ToString(output.QuxId))

	return append(diags, resourceQuxRead(ctx, d, meta)...)
}

func resourceQuxRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsClient(ctx)

	qux, err := findQuxByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Qux (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Qux (%s): %s", d.Id(), err)
	}

	d.Set("arn", qux.QuxArn)

	return diags
}

func resourceQuxUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Tags only.
	return resourceQuxRead(ctx, d, meta)
}

func resourceQuxDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsClient(ctx)

	log.Printf("[DEBUG] Deleting Qux: %s", d.Id())
	input := exampletags.DeleteQuxInput{
		QuxId: aws.String(d.Id()),
	}
	_, err := conn.DeleteQux(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Qux (%s): %s", d.Id(), err)
	}

	return diags
}

func findQuxByID(ctx context.Context, conn *exampletags.Client, id string) (*awstypes.Qux, error) {
	input := exampletags.DescribeQuxInput{
		QuxId: aws.String(id),
	}

	output, err := conn.DescribeQux(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Qux == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Qux, nil
}
