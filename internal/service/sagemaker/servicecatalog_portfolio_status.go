// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_servicecatalog_portfolio_status", name="Servicecatalog Portfolio Status")
func resourceServicecatalogPortfolioStatus() *schema.Resource {
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
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.SagemakerServicecatalogStatus](),
			},
		},
	}
}

func resourceServicecatalogPortfolioStatusPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	status := d.Get(names.AttrStatus).(string)
	var err error
	if status == string(awstypes.SagemakerServicecatalogStatusEnabled) {
		_, err = conn.EnableSagemakerServicecatalogPortfolio(ctx, &sagemaker.EnableSagemakerServicecatalogPortfolioInput{})
	} else {
		_, err = conn.DisableSagemakerServicecatalogPortfolio(ctx, &sagemaker.DisableSagemakerServicecatalogPortfolioInput{})
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting SageMaker AI Servicecatalog Portfolio Status: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))

	return append(diags, resourceServicecatalogPortfolioStatusRead(ctx, d, meta)...)
}

func resourceServicecatalogPortfolioStatusRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	resp, err := findServicecatalogPortfolioStatus(ctx, conn)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting SageMaker AI Servicecatalog Portfolio Status: %s", err)
	}

	d.Set(names.AttrStatus, resp.Status)

	return diags
}

func findServicecatalogPortfolioStatus(ctx context.Context, conn *sagemaker.Client) (*sagemaker.GetSagemakerServicecatalogPortfolioStatusOutput, error) {
	input := &sagemaker.GetSagemakerServicecatalogPortfolioStatusInput{}

	output, err := conn.GetSagemakerServicecatalogPortfolioStatus(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
