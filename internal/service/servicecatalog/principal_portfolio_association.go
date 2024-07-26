// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_servicecatalog_principal_portfolio_association")
func ResourcePrincipalPortfolioAssociation() *schema.Resource {
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
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	acceptLanguage, principalARN, portfolioID, principalType := d.Get("accept_language").(string), d.Get("principal_arn").(string), d.Get("portfolio_id").(string), d.Get("principal_type").(string)
	id := PrincipalPortfolioAssociationCreateResourceID(acceptLanguage, principalARN, portfolioID, principalType)
	input := &servicecatalog.AssociatePrincipalWithPortfolioInput{
		AcceptLanguage: aws.String(acceptLanguage),
		PortfolioId:    aws.String(portfolioID),
		PrincipalARN:   aws.String(principalARN),
		PrincipalType:  aws.String(principalType),
	}

	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return conn.AssociatePrincipalWithPortfolioWithContext(ctx, input)
	}, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Service Catalog Principal Portfolio Association (%s): %s", id, err)
	}

	d.SetId(id)

	_, err = tfresource.RetryWhenNotFound(ctx, d.Timeout(schema.TimeoutRead), func() (interface{}, error) {
		return FindPrincipalPortfolioAssociation(ctx, conn, acceptLanguage, principalARN, portfolioID, principalType)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Service Catalog Principal Portfolio Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourcePrincipalPortfolioAssociationRead(ctx, d, meta)...)
}

func resourcePrincipalPortfolioAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	acceptLanguage, principalARN, portfolioID, principalType, err := PrincipalPortfolioAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := FindPrincipalPortfolioAssociation(ctx, conn, acceptLanguage, principalARN, portfolioID, principalType)

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

func resourcePrincipalPortfolioAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	acceptLanguage, principalARN, portfolioID, principalType, err := PrincipalPortfolioAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &servicecatalog.DisassociatePrincipalFromPortfolioInput{
		PortfolioId:    aws.String(portfolioID),
		PrincipalARN:   aws.String(principalARN),
		AcceptLanguage: aws.String(acceptLanguage),
		PrincipalType:  aws.String(principalType),
	}

	log.Printf("[WARN] Deleting Service Catalog Principal Portfolio Association: %s", d.Id())
	_, err = conn.DisassociatePrincipalFromPortfolioWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Service Catalog Principal Portfolio Association (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return FindPrincipalPortfolioAssociation(ctx, conn, acceptLanguage, principalARN, portfolioID, principalType)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Service Catalog Principal Portfolio Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const principalPortfolioAssociationResourceIDSeparator = ","

func PrincipalPortfolioAssociationParseResourceID(id string) (string, string, string, string, error) {
	parts := strings.SplitN(id, principalPortfolioAssociationResourceIDSeparator, 4)

	if len(parts) != 4 || parts[0] == "" || parts[1] == "" || parts[2] == "" || parts[3] == "" {
		return "", "", "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected acceptLanguage%[2]sprincipalARN%[2]sportfolioID%[2]sprincipalType", id, principalPortfolioAssociationResourceIDSeparator)
	}

	return parts[0], parts[1], parts[2], parts[3], nil
}

func PrincipalPortfolioAssociationCreateResourceID(acceptLanguage, principalARN, portfolioID, principalType string) string {
	return strings.Join([]string{acceptLanguage, principalARN, portfolioID, principalType}, principalPortfolioAssociationResourceIDSeparator)
}

func findPrincipalForPortfolio(ctx context.Context, conn *servicecatalog.ServiceCatalog, input *servicecatalog.ListPrincipalsForPortfolioInput, filter tfslices.Predicate[*servicecatalog.Principal]) (*servicecatalog.Principal, error) {
	output, err := findPrincipalsForPortfolio(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findPrincipalsForPortfolio(ctx context.Context, conn *servicecatalog.ServiceCatalog, input *servicecatalog.ListPrincipalsForPortfolioInput, filter tfslices.Predicate[*servicecatalog.Principal]) ([]*servicecatalog.Principal, error) {
	var output []*servicecatalog.Principal

	err := conn.ListPrincipalsForPortfolioPagesWithContext(ctx, input, func(page *servicecatalog.ListPrincipalsForPortfolioOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Principals {
			if v != nil && filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindPrincipalPortfolioAssociation(ctx context.Context, conn *servicecatalog.ServiceCatalog, acceptLanguage, principalARN, portfolioID, principalType string) (*servicecatalog.Principal, error) {
	input := &servicecatalog.ListPrincipalsForPortfolioInput{
		AcceptLanguage: aws.String(acceptLanguage),
		PortfolioId:    aws.String(portfolioID),
	}
	filter := func(v *servicecatalog.Principal) bool {
		return aws.StringValue(v.PrincipalARN) == principalARN && aws.StringValue(v.PrincipalType) == principalType
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
				Default:  AcceptLanguageEnglish,
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
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  servicecatalog.PrincipalTypeIam,
			},
		},
	}
}

func principalPortfolioAssociationUpgradeV0(_ context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	if rawState == nil {
		rawState = map[string]interface{}{}
	}

	// Is resource ID in the correct format?
	if _, _, _, _, err := PrincipalPortfolioAssociationParseResourceID(rawState[names.AttrID].(string)); err != nil {
		acceptLanguage, principalARN, portfolioID, principalType := rawState["accept_language"].(string), rawState["principal_arn"].(string), rawState["portfolio_id"].(string), rawState["principal_type"].(string)
		rawState[names.AttrID] = PrincipalPortfolioAssociationCreateResourceID(acceptLanguage, principalARN, portfolioID, principalType)
	}

	return rawState, nil
}
