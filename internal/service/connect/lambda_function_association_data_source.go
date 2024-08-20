// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_connect_lambda_function_association")
func DataSourceLambdaFunctionAssociation() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLambdaFunctionAssociationRead,
		Schema: map[string]*schema.Schema{
			names.AttrFunctionARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrInstanceID: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceLambdaFunctionAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)
	functionArn := d.Get(names.AttrFunctionARN)
	instanceID := d.Get(names.AttrInstanceID)

	lfaArn, err := FindLambdaFunctionAssociationByARNWithContext(ctx, conn, instanceID.(string), functionArn.(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "finding Connect Lambda Function Association by ARN (%s): %s", functionArn, err)
	}

	if lfaArn == "" {
		return sdkdiag.AppendErrorf(diags, "finding Connect Lambda Function Association by ARN (%s): not found", functionArn)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set(names.AttrFunctionARN, functionArn)
	d.Set(names.AttrInstanceID, instanceID)

	return diags
}
