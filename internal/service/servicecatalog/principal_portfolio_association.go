package servicecatalog

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourcePrincipalPortfolioAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePrincipalPortfolioAssociationCreate,
		ReadWithoutTimeout:   resourcePrincipalPortfolioAssociationRead,
		DeleteWithoutTimeout: resourcePrincipalPortfolioAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(PrincipalPortfolioAssociationReadyTimeout),
			Read:   schema.DefaultTimeout(PrincipalPortfolioAssociationReadTimeout),
			Delete: schema.DefaultTimeout(PrincipalPortfolioAssociationDeleteTimeout),
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
			"principal_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"principal_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      servicecatalog.PrincipalTypeIam,
				ValidateFunc: validation.StringInSlice(servicecatalog.PrincipalType_Values(), false),
			},
		},
	}
}

func resourcePrincipalPortfolioAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn()

	input := &servicecatalog.AssociatePrincipalWithPortfolioInput{
		PortfolioId:  aws.String(d.Get("portfolio_id").(string)),
		PrincipalARN: aws.String(d.Get("principal_arn").(string)),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("principal_type"); ok {
		input.PrincipalType = aws.String(v.(string))
	}

	var output *servicecatalog.AssociatePrincipalWithPortfolioOutput
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		var err error

		output, err = conn.AssociatePrincipalWithPortfolioWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.AssociatePrincipalWithPortfolioWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "associating Service Catalog Principal with Portfolio: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating Service Catalog Principal Portfolio Association: empty response")
	}

	d.SetId(PrincipalPortfolioAssociationID(d.Get("accept_language").(string), d.Get("principal_arn").(string), d.Get("portfolio_id").(string)))

	return append(diags, resourcePrincipalPortfolioAssociationRead(ctx, d, meta)...)
}

func resourcePrincipalPortfolioAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn()

	acceptLanguage, principalARN, portfolioID, err := PrincipalPortfolioAssociationParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "could not parse ID (%s): %s", d.Id(), err)
	}

	if acceptLanguage == "" {
		acceptLanguage = AcceptLanguageEnglish
	}

	output, err := WaitPrincipalPortfolioAssociationReady(ctx, conn, acceptLanguage, principalARN, portfolioID, d.Timeout(schema.TimeoutRead))

	if !d.IsNewResource() && (tfresource.NotFound(err) || tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException)) {
		log.Printf("[WARN] Service Catalog Principal Portfolio Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Service Catalog Principal Portfolio Association (%s): %s", d.Id(), err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "getting Service Catalog Principal Portfolio Association (%s): empty response", d.Id())
	}

	d.Set("accept_language", acceptLanguage)
	d.Set("portfolio_id", portfolioID)
	d.Set("principal_arn", output.PrincipalARN)
	d.Set("principal_type", output.PrincipalType)

	return diags
}

func resourcePrincipalPortfolioAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn()

	acceptLanguage, principalARN, portfolioID, err := PrincipalPortfolioAssociationParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "could not parse ID (%s): %s", d.Id(), err)
	}

	if acceptLanguage == "" {
		acceptLanguage = AcceptLanguageEnglish
	}

	input := &servicecatalog.DisassociatePrincipalFromPortfolioInput{
		PortfolioId:    aws.String(portfolioID),
		PrincipalARN:   aws.String(principalARN),
		AcceptLanguage: aws.String(acceptLanguage),
	}

	_, err = conn.DisassociatePrincipalFromPortfolioWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disassociating Service Catalog Principal from Portfolio (%s): %s", d.Id(), err)
	}

	err = WaitPrincipalPortfolioAssociationDeleted(ctx, conn, acceptLanguage, principalARN, portfolioID, d.Timeout(schema.TimeoutDelete))

	if tfresource.NotFound(err) || tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Service Catalog Principal Portfolio Disassociation (%s): %s", d.Id(), err)
	}

	return diags
}
