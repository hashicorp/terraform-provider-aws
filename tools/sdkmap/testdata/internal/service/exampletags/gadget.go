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

// @SDKResource("aws_exampletags_gadget", name="Gadget")
// @Tags(identifierAttribute="arn", resourceType="Gadget")
func resourceGadget() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGadgetCreate,
		ReadWithoutTimeout:   resourceGadgetRead,
		UpdateWithoutTimeout: resourceGadgetUpdate,
		DeleteWithoutTimeout: resourceGadgetDelete,

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

func resourceGadgetCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsClient(ctx)

	input := exampletags.CreateGadgetInput{
		Tags: getTagsIn(ctx),
	}

	output, err := conn.CreateGadget(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Gadget: %s", err)
	}

	d.SetId(aws.ToString(output.GadgetId))

	return append(diags, resourceGadgetRead(ctx, d, meta)...)
}

func resourceGadgetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsClient(ctx)

	gadget, err := findGadgetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Gadget (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Gadget (%s): %s", d.Id(), err)
	}

	d.Set("arn", gadget.GadgetArn)

	return diags
}

func resourceGadgetUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Tags only.
	return resourceGadgetRead(ctx, d, meta)
}

func resourceGadgetDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsClient(ctx)

	log.Printf("[DEBUG] Deleting Gadget: %s", d.Id())
	input := exampletags.DeleteGadgetInput{
		GadgetId: aws.String(d.Id()),
	}
	_, err := conn.DeleteGadget(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Gadget (%s): %s", d.Id(), err)
	}

	return diags
}

func findGadgetByID(ctx context.Context, conn *exampletags.Client, id string) (*awstypes.Gadget, error) {
	input := exampletags.DescribeGadgetInput{
		GadgetId: aws.String(id),
	}

	output, err := conn.DescribeGadget(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Gadget == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Gadget, nil
}
