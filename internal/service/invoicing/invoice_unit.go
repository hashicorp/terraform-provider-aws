// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package invoicing

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/invoicing"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_invoicing_invoice_unit", name="Invoice Unit")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceInvoiceUnit() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInvoiceUnitCreate,
		ReadWithoutTimeout:   resourceInvoiceUnitRead,
		UpdateWithoutTimeout: resourceInvoiceUnitUpdate,
		DeleteWithoutTimeout: resourceInvoiceUnitDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"invoice_receiver": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tax_inheritance_disabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceInvoiceUnitCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InvoicingClient(ctx)

	input := invoicing.CreateInvoiceUnitInput{
		InvoiceReceiver: aws.String(d.Get("invoice_receiver").(string)),
	}

	res, err := conn.CreateInvoiceUnit(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IPAM Scope (%s): %s", d.Id(), err)
	}

	d.SetId(aws.ToString(res.InvoiceUnitArn))

	return append(diags, resourceInvoiceUnitRead(ctx, d, meta)...)
}

func resourceInvoiceUnitUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// conn := meta.(*conns.AWSClient).InvoicingClient(ctx)

	log.Printf("[WARN] TODO", d.Id())

	return append(diags, resourceInvoiceUnitRead(ctx, d, meta)...)
}

func resourceInvoiceUnitRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InvoicingClient(ctx)

	ctx = tflog.SetField(ctx, logging.KeyResourceId, d.Id())

	res, err := conn.GetInvoiceUnit(ctx, &invoicing.GetInvoiceUnitInput{
		InvoiceUnitArn: aws.String(d.Id()),
	})

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Invoice Unit (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Invoice Unit (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, res.InvoiceUnitArn)
	d.Set(names.AttrDescription, res.Description)
	d.Set("name", res.Name)

	return diags
}

func resourceInvoiceUnitDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Printf("[WARN] TODO", d.Id())
	d.SetId("")

	return diags
}
