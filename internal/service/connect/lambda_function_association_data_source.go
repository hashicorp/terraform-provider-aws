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

// @SDKDataSource("aws_connect_lambda_function_association", name="Lambda Function Association")
func dataSourceLambdaFunctionAssociation() *schema.Resource {
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

func dataSourceLambdaFunctionAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	functionARN := d.Get(names.AttrFunctionARN).(string)
	instanceID := d.Get(names.AttrInstanceID).(string)
	_, err := findLambdaFunctionAssociationByTwoPartKey(ctx, conn, instanceID, functionARN)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Lambda Function Association: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))
	d.Set(names.AttrFunctionARN, functionARN)
	d.Set(names.AttrInstanceID, instanceID)

	return diags
}
