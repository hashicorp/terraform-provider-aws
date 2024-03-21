// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_servicecatalog_portfolio_share")
func DataSourceSharePortfolio() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePortfolioShareRead,

		// Timeouts: &schema.ResourceTimeout{
		//	Read: schema.DefaultTimeout(ConstraintReadTimeout),
		// },

		Schema: map[string]*schema.Schema{
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"portfolio_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"accepted": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"share_principals": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"principal_ids": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourcePortfolioShareRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	PortfolioId := strings.Split(d.Get("portfolio_id").(string), ":")

	input := &servicecatalog.DescribePortfolioSharesInput{
		PortfolioId: aws.String(PortfolioId[0]),
		Type:        aws.String(d.Get("type").(string)),
	}

	output, err := conn.DescribePortfolioSharesWithContext(ctx, input)

	detail := output.PortfolioShareDetails

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Service Catalog Portfolio %s: %s", *detail[1].PrincipalId, err) // d.Get("portfolio_id"), err)
	}

	if output == nil || output.PortfolioShareDetails == nil {
		return sdkdiag.AppendErrorf(diags, "getting Service Catalog Portfolio (%s): empty response", d.Get("portfolio_id").(string))
	}

	principalIds := []string{}
	accepted := []string{}
	sharePrincipals := []string{}
	accountType := []string{}

	for i := 0; i < len(detail); i++ {
		principalIds = append(principalIds, *detail[i].PrincipalId)
		accepted = append(accepted, strconv.FormatBool(*detail[i].Accepted))
		sharePrincipals = append(sharePrincipals, strconv.FormatBool(*detail[i].SharePrincipals))
		accountType = append(accountType, *detail[i].Type)
	}

	PrincipalIds := strings.Join(principalIds, ",")
	Accepted := strings.Join(accepted, ",")
	SharePrincipals := strings.Join(sharePrincipals, ",")
	AccountType := strings.Join(accountType, ",")

	d.SetId(PortfolioId[0])
	d.Set("type", AccountType)
	d.Set("accepted", Accepted)
	d.Set("share_principals", SharePrincipals)
	d.Set("principal_ids", PrincipalIds)
	d.Set("wait_for_acceptance", false)
	return diags
}
