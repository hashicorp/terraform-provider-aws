// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_servicecatalog_portfolio", name="Portfolio")
// @Tags
// @Testing(existsType="github.com/aws/aws-sdk-go/service/servicecatalog;servicecatalog.DescribePortfolioOutput", generator="github.com/hashicorp/terraform-plugin-testing/helper/acctest;sdkacctest;sdkacctest.RandString(5)", skipEmptyTags=true)
func ResourcePortfolio() *schema.Resource {
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

		CustomizeDiff: verify.SetTagsDiff,
	}
}
func resourcePortfolioCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	name := d.Get(names.AttrName).(string)
	input := &servicecatalog.CreatePortfolioInput{
		AcceptLanguage:   aws.String(AcceptLanguageEnglish),
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

	output, err := conn.CreatePortfolioWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Service Catalog Portfolio (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.PortfolioDetail.Id))

	return append(diags, resourcePortfolioRead(ctx, d, meta)...)
}

func resourcePortfolioRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	output, err := FindPortfolioByID(ctx, conn, d.Id())

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

func resourcePortfolioUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	input := &servicecatalog.UpdatePortfolioInput{
		AcceptLanguage: aws.String(AcceptLanguageEnglish),
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

		input.AddTags = Tags(tftags.New(ctx, n).IgnoreAWS())
		input.RemoveTags = aws.StringSlice(tftags.New(ctx, o).IgnoreAWS().Keys())
	}

	_, err := conn.UpdatePortfolioWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Service Catalog Portfolio (%s): %s", d.Id(), err)
	}

	return append(diags, resourcePortfolioRead(ctx, d, meta)...)
}

func resourcePortfolioDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	log.Printf("[DEBUG] Deleting Service Catalog Portfolio: %s", d.Id())
	_, err := conn.DeletePortfolioWithContext(ctx, &servicecatalog.DeletePortfolioInput{
		Id: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Service Catalog Portfolio (%s): %s", d.Id(), err)
	}

	return diags
}

func FindPortfolioByID(ctx context.Context, conn *servicecatalog.ServiceCatalog, id string) (*servicecatalog.DescribePortfolioOutput, error) {
	input := &servicecatalog.DescribePortfolioInput{
		AcceptLanguage: aws.String(AcceptLanguageEnglish),
		Id:             aws.String(id),
	}

	output, err := conn.DescribePortfolioWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
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
