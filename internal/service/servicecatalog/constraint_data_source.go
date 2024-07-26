// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_servicecatalog_constraint")
func DataSourceConstraint() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceConstraintRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(ConstraintReadTimeout),
		},

		Schema: map[string]*schema.Schema{
			"accept_language": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      AcceptLanguageEnglish,
				ValidateFunc: validation.StringInSlice(AcceptLanguage_Values(), false),
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrID: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrOwner: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrParameters: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"portfolio_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"product_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrType: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceConstraintRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	output, err := WaitConstraintReady(ctx, conn, d.Get("accept_language").(string), d.Get(names.AttrID).(string), d.Timeout(schema.TimeoutRead))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Service Catalog Constraint: %s", err)
	}

	if output == nil || output.ConstraintDetail == nil {
		return sdkdiag.AppendErrorf(diags, "getting Service Catalog Constraint: empty response")
	}

	acceptLanguage := d.Get("accept_language").(string)

	if acceptLanguage == "" {
		acceptLanguage = AcceptLanguageEnglish
	}

	d.Set("accept_language", acceptLanguage)

	d.Set(names.AttrParameters, output.ConstraintParameters)
	d.Set(names.AttrStatus, output.Status)

	detail := output.ConstraintDetail

	d.Set(names.AttrDescription, detail.Description)
	d.Set(names.AttrOwner, detail.Owner)
	d.Set("portfolio_id", detail.PortfolioId)
	d.Set("product_id", detail.ProductId)
	d.Set(names.AttrType, detail.Type)

	d.SetId(aws.StringValue(detail.ConstraintId))

	return diags
}
