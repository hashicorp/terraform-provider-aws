// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_servicecatalog_portfolio_share")
func ResourcePortfolioShare() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePortfolioShareCreate,
		ReadWithoutTimeout:   resourcePortfolioShareRead,
		UpdateWithoutTimeout: resourcePortfolioShareUpdate,
		DeleteWithoutTimeout: resourcePortfolioShareDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(PortfolioShareCreateTimeout),
			Read:   schema.DefaultTimeout(PortfolioShareReadTimeout),
			Update: schema.DefaultTimeout(PortfolioShareUpdateTimeout),
			Delete: schema.DefaultTimeout(PortfolioShareDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			"accept_language": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      AcceptLanguageEnglish,
				ValidateFunc: validation.StringInSlice(AcceptLanguage_Values(), false),
			},
			"accepted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"portfolio_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			// maintaining organization_node as a separate config block makes weird configs with duplicate types
			// also, principal_id is true to API since describe gives "PrincipalId"
			"principal_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validSharePrincipal,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					newARN, err := arn.Parse(new)

					if err != nil {
						return old == new
					}

					parts := strings.Split(newARN.Resource, "/")

					return old == parts[len(parts)-1]
				},
			},
			"share_principals": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"share_tag_options": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrType: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(servicecatalog.DescribePortfolioShareType_Values(), false),
			},
			"wait_for_acceptance": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourcePortfolioShareCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	input := &servicecatalog.CreatePortfolioShareInput{
		PortfolioId:     aws.String(d.Get("portfolio_id").(string)),
		SharePrincipals: aws.Bool(d.Get("share_principals").(bool)),
		AcceptLanguage:  aws.String(d.Get("accept_language").(string)),
	}

	if v, ok := d.GetOk(names.AttrType); ok && v.(string) == servicecatalog.DescribePortfolioShareTypeAccount {
		input.AccountId = aws.String(d.Get("principal_id").(string))
	} else {
		orgNode := &servicecatalog.OrganizationNode{}
		orgNode.Value = aws.String(d.Get("principal_id").(string))

		if v.(string) == servicecatalog.DescribePortfolioShareTypeOrganizationMemberAccount {
			// portfolio_share type ORGANIZATION_MEMBER_ACCOUNT = org node type ACCOUNT
			orgNode.Type = aws.String(servicecatalog.OrganizationNodeTypeAccount)
		} else {
			orgNode.Type = aws.String(d.Get(names.AttrType).(string))
		}

		input.OrganizationNode = orgNode
	}

	if v, ok := d.GetOk("share_tag_options"); ok {
		input.ShareTagOptions = aws.Bool(v.(bool))
	}

	var output *servicecatalog.CreatePortfolioShareOutput
	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		var err error

		output, err = conn.CreatePortfolioShareWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreatePortfolioShareWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Service Catalog Portfolio Share: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating Service Catalog Portfolio Share: empty response")
	}

	d.SetId(PortfolioShareCreateResourceID(d.Get("portfolio_id").(string), d.Get(names.AttrType).(string), d.Get("principal_id").(string)))

	waitForAcceptance := false
	if v, ok := d.GetOk("wait_for_acceptance"); ok {
		waitForAcceptance = v.(bool)
	}

	// only get a token if organization node, otherwise check without token
	if output.PortfolioShareToken != nil {
		if _, err := WaitPortfolioShareCreatedWithToken(ctx, conn, aws.StringValue(output.PortfolioShareToken), waitForAcceptance, d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Service Catalog Portfolio Share (%s) to be ready: %s", d.Id(), err)
		}
	} else {
		if _, err := WaitPortfolioShareReady(ctx, conn, d.Get("portfolio_id").(string), d.Get(names.AttrType).(string), d.Get("principal_id").(string), waitForAcceptance, d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Service Catalog Portfolio Share (%s) to be ready: %s", d.Id(), err)
		}
	}

	return append(diags, resourcePortfolioShareRead(ctx, d, meta)...)
}

func resourcePortfolioShareRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	portfolioID, shareType, principalID, err := PortfolioShareParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "could not parse ID (%s): %s", d.Id(), err)
	}

	waitForAcceptance := false
	if v, ok := d.GetOk("wait_for_acceptance"); ok {
		waitForAcceptance = v.(bool)
	}

	output, err := WaitPortfolioShareReady(ctx, conn, portfolioID, shareType, principalID, waitForAcceptance, d.Timeout(schema.TimeoutRead))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Service Catalog Portfolio Share (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "readingService Catalog Portfolio Share (%s): %s", d.Id(), err)
	}

	d.Set("accepted", output.Accepted)
	d.Set("portfolio_id", portfolioID)
	d.Set("principal_id", output.PrincipalId)
	d.Set("share_principals", output.SharePrincipals)
	d.Set("share_tag_options", output.ShareTagOptions)
	d.Set(names.AttrType, output.Type)
	d.Set("wait_for_acceptance", waitForAcceptance)

	return diags
}

func resourcePortfolioShareUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	input := &servicecatalog.UpdatePortfolioShareInput{
		PortfolioId:    aws.String(d.Get("portfolio_id").(string)),
		AcceptLanguage: aws.String(d.Get("accept_language").(string)),
	}

	if d.HasChange("share_principals") {
		input.SharePrincipals = aws.Bool(d.Get("share_principals").(bool))
	}

	if d.HasChange("share_tag_options") {
		input.ShareTagOptions = aws.Bool(d.Get("share_tag_options").(bool))
	}

	if v, ok := d.GetOk(names.AttrType); ok && v.(string) == servicecatalog.DescribePortfolioShareTypeAccount {
		input.AccountId = aws.String(d.Get("principal_id").(string))
	} else {
		orgNode := &servicecatalog.OrganizationNode{}
		orgNode.Value = aws.String(d.Get("principal_id").(string))

		if v.(string) == servicecatalog.DescribePortfolioShareTypeOrganizationMemberAccount {
			// portfolio_share type ORGANIZATION_MEMBER_ACCOUNT = org node type ACCOUNT
			orgNode.Type = aws.String(servicecatalog.OrganizationNodeTypeAccount)
		} else {
			orgNode.Type = aws.String(d.Get(names.AttrType).(string))
		}

		input.OrganizationNode = orgNode
	}

	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *retry.RetryError {
		_, err := conn.UpdatePortfolioShareWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.UpdatePortfolioShareWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Service Catalog Portfolio Share (%s): %s", d.Id(), err)
	}

	return append(diags, resourcePortfolioShareRead(ctx, d, meta)...)
}

func resourcePortfolioShareDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	input := &servicecatalog.DeletePortfolioShareInput{
		PortfolioId: aws.String(d.Get("portfolio_id").(string)),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrType); ok && v.(string) == servicecatalog.DescribePortfolioShareTypeAccount {
		input.AccountId = aws.String(d.Get("principal_id").(string))
	} else {
		orgNode := &servicecatalog.OrganizationNode{}
		orgNode.Value = aws.String(d.Get("principal_id").(string))

		if v.(string) == servicecatalog.DescribePortfolioShareTypeOrganizationMemberAccount {
			// portfolio_share type ORGANIZATION_MEMBER_ACCOUNT = org node type ACCOUNT
			orgNode.Type = aws.String(servicecatalog.OrganizationNodeTypeAccount)
		} else {
			orgNode.Type = aws.String(d.Get(names.AttrType).(string))
		}

		input.OrganizationNode = orgNode
	}

	output, err := conn.DeletePortfolioShareWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Service Catalog Portfolio Share (%s): %s", d.Id(), err)
	}

	// only get a token if organization node, otherwise check without token
	if output.PortfolioShareToken != nil {
		if _, err := WaitPortfolioShareDeletedWithToken(ctx, conn, aws.StringValue(output.PortfolioShareToken), d.Timeout(schema.TimeoutDelete)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Service Catalog Portfolio Share (%s) to be deleted: %s", d.Id(), err)
		}
	} else {
		if _, err := WaitPortfolioShareDeleted(ctx, conn, d.Get("portfolio_id").(string), d.Get(names.AttrType).(string), d.Get("principal_id").(string), d.Timeout(schema.TimeoutDelete)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Service Catalog Portfolio Share (%s) to be deleted: %s", d.Id(), err)
		}
	}

	return diags
}
