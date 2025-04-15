// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalog"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicecatalog/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_servicecatalog_portfolio", name="Portfolio")
// @Tags
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/servicecatalog;servicecatalog.DescribePortfolioOutput", generator="github.com/hashicorp/terraform-plugin-testing/helper/acctest;sdkacctest;sdkacctest.RandString(5)", skipEmptyTags=true)
// @Testing(tagsIdentifierAttribute="id", tagsResourceType="Portfolio")
func resourcePortfolio() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePortfolioCreate,
		ReadWithoutTimeout:   resourcePortfolioRead,
		UpdateWithoutTimeout: resourcePortfolioUpdate,
		DeleteWithoutTimeout: resourcePortfolioDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(PortfolioCreateTimeout),
			Read:   schema.DefaultTimeout(PortfolioReadTimeout),
			Update: schema.DefaultTimeout(PortfolioUpdateTimeout),
			Delete: schema.DefaultTimeout(PortfolioDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(0, 2000),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			names.AttrProviderName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 50),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}
func resourcePortfolioCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &servicecatalog.CreatePortfolioInput{
		AcceptLanguage:   aws.String(acceptLanguageEnglish),
		DisplayName:      aws.String(name),
		IdempotencyToken: aws.String(id.UniqueId()),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrProviderName); ok {
		input.ProviderName = aws.String(v.(string))
	}

	output, err := conn.CreatePortfolio(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Service Catalog Portfolio (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.PortfolioDetail.Id))

	return append(diags, resourcePortfolioRead(ctx, d, meta)...)
}

func resourcePortfolioRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogClient(ctx)

	output, err := findPortfolioByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Service Catalog Portfolio (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Service Catalog Portfolio (%s): %s", d.Id(), err)
	}

	portfolioDetail := output.PortfolioDetail
	d.Set(names.AttrARN, portfolioDetail.ARN)
	d.Set(names.AttrCreatedTime, portfolioDetail.CreatedTime.Format(time.RFC3339))
	d.Set(names.AttrDescription, portfolioDetail.Description)
	d.Set(names.AttrName, portfolioDetail.DisplayName)
	d.Set(names.AttrProviderName, portfolioDetail.ProviderName)

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourcePortfolioUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogClient(ctx)

	input := &servicecatalog.UpdatePortfolioInput{
		AcceptLanguage: aws.String(acceptLanguageEnglish),
		Id:             aws.String(d.Id()),
	}

	if d.HasChange("accept_language") {
		input.AcceptLanguage = aws.String(d.Get("accept_language").(string))
	}

	if d.HasChange(names.AttrDescription) {
		input.Description = aws.String(d.Get(names.AttrDescription).(string))
	}

	if d.HasChange(names.AttrName) {
		input.DisplayName = aws.String(d.Get(names.AttrName).(string))
	}

	if d.HasChange(names.AttrProviderName) {
		input.ProviderName = aws.String(d.Get(names.AttrProviderName).(string))
	}

	if d.HasChange(names.AttrTagsAll) {
		o, n := d.GetChange(names.AttrTagsAll)

		input.AddTags = svcTags(tftags.New(ctx, n).IgnoreAWS())
		input.RemoveTags = tftags.New(ctx, o).IgnoreAWS().Keys()
	}

	_, err := conn.UpdatePortfolio(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Service Catalog Portfolio (%s): %s", d.Id(), err)
	}

	return append(diags, resourcePortfolioRead(ctx, d, meta)...)
}

func resourcePortfolioDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogClient(ctx)

	log.Printf("[DEBUG] Deleting Service Catalog Portfolio: %s", d.Id())
	_, err := conn.DeletePortfolio(ctx, &servicecatalog.DeletePortfolioInput{
		Id: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Service Catalog Portfolio (%s): %s", d.Id(), err)
	}

	return diags
}

func findPortfolioByID(ctx context.Context, conn *servicecatalog.Client, id string) (*servicecatalog.DescribePortfolioOutput, error) {
	input := &servicecatalog.DescribePortfolioInput{
		AcceptLanguage: aws.String(acceptLanguageEnglish),
		Id:             aws.String(id),
	}

	output, err := conn.DescribePortfolio(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
