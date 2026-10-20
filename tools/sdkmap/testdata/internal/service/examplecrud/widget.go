// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package examplecrud

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/examplecrud"
	awstypes "github.com/aws/aws-sdk-go-v2/service/examplecrud/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

// @SDKResource("aws_examplecrud_widget", name="Widget")
func resourceWidget() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWidgetCreate,
		ReadWithoutTimeout:   resourceWidgetRead,
		UpdateWithoutTimeout: resourceWidgetUpdate,
		DeleteWithoutTimeout: resourceWidgetDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceWidgetCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleCRUDClient(ctx)

	name := d.Get("name").(string)
	input := examplecrud.CreateWidgetInput{
		Description: aws.String(d.Get("description").(string)),
		Name:        aws.String(name),
	}

	output, err := conn.CreateWidget(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Widget (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.WidgetId))

	if _, err := waitWidgetCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Widget (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceWidgetRead(ctx, d, meta)...)
}

func resourceWidgetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleCRUDClient(ctx)

	widget, err := findWidgetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Widget (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Widget (%s): %s", d.Id(), err)
	}

	d.Set("description", widget.Description)
	d.Set("name", widget.Name)
	d.Set("region", conn.Options().Region)

	return diags
}

func resourceWidgetUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleCRUDClient(ctx)

	input := examplecrud.UpdateWidgetInput{
		Description: aws.String(d.Get("description").(string)),
		WidgetId:    aws.String(d.Id()),
	}

	_, err := conn.UpdateWidget(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Widget (%s): %s", d.Id(), err)
	}

	return append(diags, resourceWidgetRead(ctx, d, meta)...)
}

func resourceWidgetDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleCRUDClient(ctx)

	if err := deleteWidgetParts(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Widget (%s) parts: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting Widget: %s", d.Id())
	input := examplecrud.DeleteWidgetInput{
		WidgetId: aws.String(d.Id()),
	}
	_, err := conn.DeleteWidget(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Widget (%s): %s", d.Id(), err)
	}

	if _, err := waitWidgetDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Widget (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func deleteWidgetParts(ctx context.Context, conn *examplecrud.Client, widgetID string) error {
	input := examplecrud.ListWidgetPartsInput{
		WidgetId: aws.String(widgetID),
	}

	pages := examplecrud.NewListWidgetPartsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return err
		}

		for _, part := range page.Parts {
			input := examplecrud.DeleteWidgetPartInput{
				PartId:   part.PartId,
				WidgetId: aws.String(widgetID),
			}

			if _, err := conn.DeleteWidgetPart(ctx, &input); err != nil {
				return err
			}
		}
	}

	return nil
}
