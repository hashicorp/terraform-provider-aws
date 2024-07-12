// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_servicecatalog_portfolio")
// @Tags
func DataSourcePortfolio() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePortfolioRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(ConstraintReadTimeout),
		},

		Schema: map[string]*schema.Schema{
			"accept_language": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "en",
				ValidateFunc: validation.StringInSlice(AcceptLanguage_Values(), false),
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
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrProviderName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourcePortfolioRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	input := &servicecatalog.DescribePortfolioInput{
		Id: aws.String(d.Get(names.AttrID).(string)),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	output, err := conn.DescribePortfolioWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Service Catalog Portfolio (%s): %s", d.Get(names.AttrID).(string), err)
	}

	if output == nil || output.PortfolioDetail == nil {
		return sdkdiag.AppendErrorf(diags, "getting Service Catalog Portfolio (%s): empty response", d.Get(names.AttrID).(string))
	}

	detail := output.PortfolioDetail

	d.SetId(aws.StringValue(detail.Id))

	if err := d.Set(names.AttrCreatedTime, aws.TimeValue(detail.CreatedTime).Format(time.RFC3339)); err != nil {
		log.Printf("[DEBUG] Error setting created_time: %s", err)
	}

	d.Set(names.AttrARN, detail.ARN)
	d.Set(names.AttrDescription, detail.Description)
	d.Set(names.AttrName, detail.DisplayName)
	d.Set(names.AttrProviderName, detail.ProviderName)

	setTagsOut(ctx, output.Tags)

	return diags
}
