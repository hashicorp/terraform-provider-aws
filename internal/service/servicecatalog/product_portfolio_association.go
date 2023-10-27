// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_servicecatalog_product_portfolio_association")
func ResourceProductPortfolioAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProductPortfolioAssociationCreate,
		ReadWithoutTimeout:   resourceProductPortfolioAssociationRead,
		DeleteWithoutTimeout: resourceProductPortfolioAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(ProductPortfolioAssociationReadyTimeout),
			Read:   schema.DefaultTimeout(ProductPortfolioAssociationReadTimeout),
			Delete: schema.DefaultTimeout(ProductPortfolioAssociationDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			"accept_language": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      AcceptLanguageEnglish,
				ValidateFunc: validation.StringInSlice(AcceptLanguage_Values(), false),
			},
			"portfolio_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"product_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source_portfolio_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceProductPortfolioAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	input := &servicecatalog.AssociateProductWithPortfolioInput{
		PortfolioId: aws.String(d.Get("portfolio_id").(string)),
		ProductId:   aws.String(d.Get("product_id").(string)),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("source_portfolio_id"); ok {
		input.SourcePortfolioId = aws.String(v.(string))
	}

	var output *servicecatalog.AssociateProductWithPortfolioOutput
	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		var err error

		output, err = conn.AssociateProductWithPortfolioWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.AssociateProductWithPortfolioWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "associating Service Catalog Product with Portfolio: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating Service Catalog Product Portfolio Association: empty response")
	}

	d.SetId(ProductPortfolioAssociationCreateID(d.Get("accept_language").(string), d.Get("portfolio_id").(string), d.Get("product_id").(string)))

	return append(diags, resourceProductPortfolioAssociationRead(ctx, d, meta)...)
}

func resourceProductPortfolioAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	acceptLanguage, portfolioID, productID, err := ProductPortfolioAssociationParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "could not parse ID (%s): %s", d.Id(), err)
	}

	output, err := WaitProductPortfolioAssociationReady(ctx, conn, acceptLanguage, portfolioID, productID, d.Timeout(schema.TimeoutRead))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Service Catalog Product Portfolio Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Service Catalog Product Portfolio Association (%s): %s", d.Id(), err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "getting Service Catalog Product Portfolio Association (%s): empty response", d.Id())
	}

	d.Set("accept_language", acceptLanguage)
	d.Set("portfolio_id", output.Id)
	d.Set("product_id", productID)
	d.Set("source_portfolio_id", d.Get("source_portfolio_id").(string))

	return diags
}

func resourceProductPortfolioAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	acceptLanguage, portfolioID, productID, err := ProductPortfolioAssociationParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "could not parse ID (%s): %s", d.Id(), err)
	}

	input := &servicecatalog.DisassociateProductFromPortfolioInput{
		PortfolioId: aws.String(portfolioID),
		ProductId:   aws.String(productID),
	}

	if acceptLanguage != "" {
		input.AcceptLanguage = aws.String(acceptLanguage)
	}

	_, err = conn.DisassociateProductFromPortfolioWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disassociating Service Catalog Product from Portfolio (%s): %s", d.Id(), err)
	}

	err = WaitProductPortfolioAssociationDeleted(ctx, conn, acceptLanguage, portfolioID, productID, d.Timeout(schema.TimeoutDelete))

	if err != nil && !tfresource.NotFound(err) {
		return sdkdiag.AppendErrorf(diags, "waiting for Service Catalog Product Portfolio Disassociation (%s): %s", d.Id(), err)
	}

	return diags
}
