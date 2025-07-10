// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalog"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicecatalog/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_servicecatalog_principal_portfolio_association", name="Principal Portfolio Association")
func resourcePrincipalPortfolioAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePrincipalPortfolioAssociationCreate,
		ReadWithoutTimeout:   resourcePrincipalPortfolioAssociationRead,
		DeleteWithoutTimeout: resourcePrincipalPortfolioAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Read:   schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourcePrincipalPortfolioAssociationV0().CoreConfigSchema().ImpliedType(),
				Upgrade: principalPortfolioAssociationUpgradeV0,
				Version: 0,
			},
		},

		Schema: map[string]*schema.Schema{
			"accept_language": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      acceptLanguageEnglish,
				ValidateFunc: validation.StringInSlice(acceptLanguage_Values(), false),
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
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.PrincipalTypeIam,
				ValidateDiagFunc: enum.Validate[awstypes.PrincipalType](),
			},
		},
	}
}

func resourcePrincipalPortfolioAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogClient(ctx)

	acceptLanguage, principalARN, portfolioID, principalType := d.Get("accept_language").(string), d.Get("principal_arn").(string), d.Get("portfolio_id").(string), d.Get("principal_type").(string)
	id := principalPortfolioAssociationCreateResourceID(acceptLanguage, principalARN, portfolioID, principalType)
	input := &servicecatalog.AssociatePrincipalWithPortfolioInput{
		AcceptLanguage: aws.String(acceptLanguage),
		PortfolioId:    aws.String(portfolioID),
		PrincipalARN:   aws.String(principalARN),
		PrincipalType:  awstypes.PrincipalType(principalType),
	}

	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidParametersException](ctx, d.Timeout(schema.TimeoutCreate), func() (any, error) {
		return conn.AssociatePrincipalWithPortfolio(ctx, input)
	}, "profile does not exist")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Service Catalog Principal Portfolio Association (%s): %s", id, err)
	}

	d.SetId(id)

	_, err = tfresource.RetryWhenNotFound(ctx, d.Timeout(schema.TimeoutRead), func() (any, error) {
		return findPrincipalPortfolioAssociation(ctx, conn, acceptLanguage, principalARN, portfolioID, principalType)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Service Catalog Principal Portfolio Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourcePrincipalPortfolioAssociationRead(ctx, d, meta)...)
}

func resourcePrincipalPortfolioAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogClient(ctx)

	acceptLanguage, principalARN, portfolioID, principalType, err := principalPortfolioAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findPrincipalPortfolioAssociation(ctx, conn, acceptLanguage, principalARN, portfolioID, principalType)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Service Catalog Principal Portfolio Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Service Catalog Principal Portfolio Association (%s): %s", d.Id(), err)
	}

	d.Set("accept_language", acceptLanguage)
	d.Set("portfolio_id", portfolioID)
	d.Set("principal_arn", output.PrincipalARN)
	d.Set("principal_type", output.PrincipalType)

	return diags
}

func resourcePrincipalPortfolioAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogClient(ctx)

	acceptLanguage, principalARN, portfolioID, principalType, err := principalPortfolioAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &servicecatalog.DisassociatePrincipalFromPortfolioInput{
		PortfolioId:    aws.String(portfolioID),
		PrincipalARN:   aws.String(principalARN),
		AcceptLanguage: aws.String(acceptLanguage),
		PrincipalType:  awstypes.PrincipalType(principalType),
	}

	log.Printf("[WARN] Deleting Service Catalog Principal Portfolio Association: %s", d.Id())
	_, err = conn.DisassociatePrincipalFromPortfolio(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Service Catalog Principal Portfolio Association (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, d.Timeout(schema.TimeoutDelete), func() (any, error) {
		return findPrincipalPortfolioAssociation(ctx, conn, acceptLanguage, principalARN, portfolioID, principalType)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Service Catalog Principal Portfolio Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const principalPortfolioAssociationResourceIDSeparator = ","

func principalPortfolioAssociationParseResourceID(id string) (string, string, string, string, error) {
	parts := strings.SplitN(id, principalPortfolioAssociationResourceIDSeparator, 4)

	if len(parts) != 4 || parts[0] == "" || parts[1] == "" || parts[2] == "" || parts[3] == "" {
		return "", "", "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected acceptLanguage%[2]sprincipalARN%[2]sportfolioID%[2]sprincipalType", id, principalPortfolioAssociationResourceIDSeparator)
	}

	return parts[0], parts[1], parts[2], parts[3], nil
}

func principalPortfolioAssociationCreateResourceID(acceptLanguage, principalARN, portfolioID, principalType string) string {
	return strings.Join([]string{acceptLanguage, principalARN, portfolioID, principalType}, principalPortfolioAssociationResourceIDSeparator)
}

func findPrincipalForPortfolio(ctx context.Context, conn *servicecatalog.Client, input *servicecatalog.ListPrincipalsForPortfolioInput, filter tfslices.Predicate[awstypes.Principal]) (*awstypes.Principal, error) {
	output, err := findPrincipalsForPortfolio(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findPrincipalsForPortfolio(ctx context.Context, conn *servicecatalog.Client, input *servicecatalog.ListPrincipalsForPortfolioInput, filter tfslices.Predicate[awstypes.Principal]) ([]awstypes.Principal, error) {
	var output []awstypes.Principal

	pages := servicecatalog.NewListPrincipalsForPortfolioPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Principals {
			if v.PrincipalARN != nil && filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func findPrincipalPortfolioAssociation(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, principalARN, portfolioID, principalType string) (*awstypes.Principal, error) {
	input := &servicecatalog.ListPrincipalsForPortfolioInput{
		AcceptLanguage: aws.String(acceptLanguage),
		PortfolioId:    aws.String(portfolioID),
	}
	filter := func(v awstypes.Principal) bool {
		return aws.ToString(v.PrincipalARN) == principalARN && string(v.PrincipalType) == principalType
	}

	return findPrincipalForPortfolio(ctx, conn, input, filter)
}

// aws_autoscaling_group aws_servicecatalog_principal_portfolio_association's Schema @v5.15.0 minus validators.
func resourcePrincipalPortfolioAssociationV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"accept_language": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  acceptLanguageEnglish,
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
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.PrincipalTypeIam,
				ValidateDiagFunc: enum.Validate[awstypes.PrincipalType](),
			},
		},
	}
}

func principalPortfolioAssociationUpgradeV0(_ context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	if rawState == nil {
		rawState = map[string]any{}
	}

	// Is resource ID in the correct format?
	if _, _, _, _, err := principalPortfolioAssociationParseResourceID(rawState[names.AttrID].(string)); err != nil {
		acceptLanguage, principalARN, portfolioID, principalType := rawState["accept_language"].(string), rawState["principal_arn"].(string), rawState["portfolio_id"].(string), rawState["principal_type"].(string)
		rawState[names.AttrID] = principalPortfolioAssociationCreateResourceID(acceptLanguage, principalARN, portfolioID, principalType)
	}

	return rawState, nil
}
