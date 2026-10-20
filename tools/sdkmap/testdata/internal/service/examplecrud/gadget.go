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
	tfexamplecrossservice "github.com/hashicorp/terraform-provider-aws/internal/service/examplecrossservice"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	gadgetDeleteTimeout = 5 * time.Minute
)

// @SDKResource("aws_examplecrud_gadget", name="Gadget")
func resourceGadget() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGadgetCreate,
		ReadWithoutTimeout:   resourceGadgetRead,
		UpdateWithoutTimeout: resourceGadgetUpdate,
		DeleteWithoutTimeout: resourceGadgetDelete,

		Schema: map[string]*schema.Schema{
			"size": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"thing_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"widget_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceGadgetCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.ExampleCRUDClient(ctx)

	widgetID := d.Get("widget_id").(string)
	if _, err := findWidgetByID(ctx, conn, widgetID); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Widget (%s): %s", widgetID, err)
	}

	if v, ok := d.GetOk("thing_id"); ok {
		if _, err := tfexamplecrossservice.FindThingByID(ctx, c.ExampleCrossServiceClient(ctx), v.(string)); err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Thing (%s): %s", v, err)
		}
	}

	input := examplecrud.CreateGadgetInput{
		Size:     aws.Int32(int32(d.Get("size").(int))),
		WidgetId: aws.String(widgetID),
	}

	output, err := conn.CreateGadget(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Gadget: %s", err)
	}

	d.SetId(aws.ToString(output.GadgetId))

	const (
		inARow = 2
	)
	if _, err := retry.Op(func(ctx context.Context) (*awstypes.Gadget, error) {
		return findGadgetByID(ctx, conn, d.Id())
	}).UntilFoundN(inARow)(ctx, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Gadget (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceGadgetRead(ctx, d, meta)...)
}

func resourceGadgetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleCRUDClient(ctx)

	gadget, err := findGadgetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Gadget (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Gadget (%s): %s", d.Id(), err)
	}

	d.Set("size", gadget.Size)
	d.Set("widget_id", gadget.WidgetId)

	return diags
}

func resourceGadgetUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleCRUDClient(ctx)

	input := examplecrud.UpdateGadgetInput{
		GadgetId: aws.String(d.Id()),
		Size:     aws.Int32(int32(d.Get("size").(int))),
	}

	_, err := conn.UpdateGadget(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Gadget (%s): %s", d.Id(), err)
	}

	return append(diags, resourceGadgetRead(ctx, d, meta)...)
}

func resourceGadgetDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleCRUDClient(ctx)

	log.Printf("[DEBUG] Deleting Gadget: %s", d.Id())
	_, err := tfresource.RetryWhenIsA[any, *awstypes.ResourceInUseException](ctx, gadgetDeleteTimeout, func(ctx context.Context) (any, error) {
		input := examplecrud.DeleteGadgetInput{
			GadgetId: aws.String(d.Id()),
		}

		return conn.DeleteGadget(ctx, &input)
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Gadget (%s): %s", d.Id(), err)
	}

	return diags
}
