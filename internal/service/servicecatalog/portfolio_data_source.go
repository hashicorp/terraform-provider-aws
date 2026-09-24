// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package servicecatalog

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_servicecatalog_portfolio", name="Portfolio")
// @Tags
// @Testing(tagsIdentifierAttribute="id", tagsResourceType="Portfolio")
func dataSourcePortfolio() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePortfolioRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(ConstraintReadTimeout),
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"accept_language": {
					Type:         schema.TypeString,
					Optional:     true,
					Default:      acceptLanguageEnglish,
					ValidateFunc: validation.StringInSlice(acceptLanguage_Values(), false),
				},
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrCreatedTime: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrDescription: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrID: {
					Type:         schema.TypeString,
					Optional:     true,
					Computed:     true,
					ExactlyOneOf: []string{names.AttrID, names.AttrName},
				},
				names.AttrName: {
					Type:         schema.TypeString,
					Optional:     true,
					Computed:     true,
					ExactlyOneOf: []string{names.AttrID, names.AttrName},
				},
				names.AttrProviderName: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrTags: tftags.TagsSchemaComputed(),
			}
		},
	}
}

func dataSourcePortfolioRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogClient(ctx)

	acceptLanguage := d.Get("accept_language").(string)

	id := d.Get(names.AttrID).(string)

	if id == "" {
		name := d.Get(names.AttrName).(string)

		portfolio, err := findPortfolioByName(ctx, conn, acceptLanguage, name)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Service Catalog Portfolio (%s): %s", name, err)
		}

		id = aws.ToString(portfolio.Id)
	}

	input := &servicecatalog.DescribePortfolioInput{
		Id: aws.String(id),
	}

	if acceptLanguage != "" {
		input.AcceptLanguage = aws.String(acceptLanguage)
	}

	output, err := conn.DescribePortfolio(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Service Catalog Portfolio (%s): %s", id, err)
	}

	if output == nil || output.PortfolioDetail == nil {
		return sdkdiag.AppendErrorf(diags, "getting Service Catalog Portfolio (%s): empty response", id)
	}

	detail := output.PortfolioDetail

	d.SetId(aws.ToString(detail.Id))

	if err := d.Set(names.AttrCreatedTime, aws.ToTime(detail.CreatedTime).Format(time.RFC3339)); err != nil {
		log.Printf("[DEBUG] Error setting created_time: %s", err)
	}

	d.Set(names.AttrARN, detail.ARN)
	d.Set(names.AttrDescription, detail.Description)
	d.Set(names.AttrName, detail.DisplayName)
	d.Set(names.AttrProviderName, detail.ProviderName)

	setTagsOut(ctx, output.Tags)

	return diags
}
