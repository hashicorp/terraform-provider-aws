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

// @SDKResource("aws_exampletags_widget", name="Widget")
// @SDKResource("aws_exampletags_legacy_widget", name="Widget")
// @Tags(identifierAttribute="arn", resourceType="Widget")
func resourceWidget() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWidgetCreate,
		ReadWithoutTimeout:   resourceWidgetRead,
		UpdateWithoutTimeout: resourceWidgetUpdate,
		DeleteWithoutTimeout: resourceWidgetDelete,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceWidgetCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsClient(ctx)

	name := d.Get("name").(string)
	input := exampletags.CreateWidgetInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
	}

	output, err := conn.CreateWidget(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Widget (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.WidgetId))

	return append(diags, resourceWidgetRead(ctx, d, meta)...)
}

func resourceWidgetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsClient(ctx)

	widget, err := findWidgetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Widget (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Widget (%s): %s", d.Id(), err)
	}

	d.Set("arn", widget.WidgetArn)
	d.Set("name", widget.Name)

	return diags
}

func resourceWidgetUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Tags only.
	return resourceWidgetRead(ctx, d, meta)
}

func resourceWidgetDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsClient(ctx)

	log.Printf("[DEBUG] Deleting Widget: %s", d.Id())
	input := exampletags.DeleteWidgetInput{
		WidgetId: aws.String(d.Id()),
	}
	_, err := conn.DeleteWidget(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Widget (%s): %s", d.Id(), err)
	}

	return diags
}

func findWidgetByID(ctx context.Context, conn *exampletags.Client, id string) (*awstypes.Widget, error) {
	input := exampletags.DescribeWidgetInput{
		WidgetId: aws.String(id),
	}

	output, err := conn.DescribeWidget(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Widget == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Widget, nil
}
