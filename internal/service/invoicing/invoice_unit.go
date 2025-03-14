// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package invoicing

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/invoicing"
	"github.com/aws/aws-sdk-go-v2/service/invoicing/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"invoice_receiver": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidAccountID,
				ForceNew:     true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tax_inheritance_disabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"linked_accounts": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidAccountID,
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceInvoiceUnitCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InvoicingClient(ctx)

	input := invoicing.CreateInvoiceUnitInput{}

	input.InvoiceReceiver = aws.String(d.Get("invoice_receiver").(string))
	input.Name = aws.String(d.Get("name").(string))
	input.Rule = &types.InvoiceUnitRule{
		LinkedAccounts: flex.ExpandStringValueSet(d.Get("linked_accounts").(*schema.Set)),
	}

	if v, ok := d.GetOk("tax_inheritance_disabled"); ok {
		input.TaxInheritanceDisabled = v.(bool)
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	input.ResourceTags = getTagsIn(ctx)

	res, err := conn.CreateInvoiceUnit(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Creating Invoicing Unit: %s", err)
	}

	d.SetId(aws.ToString(res.InvoiceUnitArn))

	return append(diags, resourceInvoiceUnitRead(ctx, d, meta)...)
}

func resourceInvoiceUnitUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InvoicingClient(ctx)

	input := invoicing.UpdateInvoiceUnitInput{}

	input.InvoiceUnitArn = aws.String(d.Id())
	input.Rule = &types.InvoiceUnitRule{
		LinkedAccounts: flex.ExpandStringValueSet(d.Get("linked_accounts").(*schema.Set)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tax_inheritance_disabled"); ok {
		input.TaxInheritanceDisabled = aws.Bool(v.(bool))
	}

	if _, err := conn.UpdateInvoiceUnit(ctx, &input); err != nil {
		return sdkdiag.AppendErrorf(diags, "Updating Invoice Unit (%s): %s", d.Id(), err)
	}

	return append(diags, resourceInvoiceUnitRead(ctx, d, meta)...)
}

func resourceInvoiceUnitRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InvoicingClient(ctx)

	ctx = tflog.SetField(ctx, logging.KeyResourceId, d.Id())

	input := invoicing.GetInvoiceUnitInput{
		InvoiceUnitArn: aws.String(d.Id()),
	}

	res, err := conn.GetInvoiceUnit(ctx, &input)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Invoice Unit (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Reading Invoice Unit (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, res.InvoiceUnitArn)
	d.Set(names.AttrDescription, res.Description)
	d.Set("name", res.Name)
	d.Set("invoice_receiver", res.InvoiceReceiver)
	d.Set("linked_accounts", res.Rule.LinkedAccounts)
	d.Set("tax_inheritance_disabled", res.TaxInheritanceDisabled)

	return diags
}

func resourceInvoiceUnitDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).InvoicingClient(ctx)

	input := invoicing.DeleteInvoiceUnitInput{}

	input.InvoiceUnitArn = aws.String(d.Id())

	if _, err := conn.DeleteInvoiceUnit(ctx, &input); err != nil {
		return sdkdiag.AppendErrorf(diags, "Deleting Invoice Unit (%s): %s", d.Id(), err)
	}

	return diags
}
