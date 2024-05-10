// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"

	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_servicecatalog_portfolio_status")
func ResourceServicecatalogPortfolioStatus() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceServicecatalogPortfolioStatusPut,
		ReadWithoutTimeout:   resourceServicecatalogPortfolioStatusRead,
		UpdateWithoutTimeout: resourceServicecatalogPortfolioStatusPut,
		DeleteWithoutTimeout: schema.NoopContext,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrStatus: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(sagemaker.SagemakerServicecatalogStatus_Values(), false),
			},
		},
	}
}

func resourceServicecatalogPortfolioStatusPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	status := d.Get(names.AttrStatus).(string)
	var err error
	if status == sagemaker.SagemakerServicecatalogStatusEnabled {
		_, err = conn.EnableSagemakerServicecatalogPortfolioWithContext(ctx, &sagemaker.EnableSagemakerServicecatalogPortfolioInput{})
	} else {
		_, err = conn.DisableSagemakerServicecatalogPortfolioWithContext(ctx, &sagemaker.DisableSagemakerServicecatalogPortfolioInput{})
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting SageMaker Servicecatalog Portfolio Status: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	return append(diags, resourceServicecatalogPortfolioStatusRead(ctx, d, meta)...)
}

func resourceServicecatalogPortfolioStatusRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	resp, err := conn.GetSagemakerServicecatalogPortfolioStatusWithContext(ctx, &sagemaker.GetSagemakerServicecatalogPortfolioStatusInput{})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Getting SageMaker Servicecatalog Portfolio Status: %s", err)
	}

	d.Set(names.AttrStatus, resp.Status)

	return diags
}
