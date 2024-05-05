// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_lambda_functions", name="Functions")
func dataSourceFunctions() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceFunctionsRead,

		Schema: map[string]*schema.Schema{
			"function_arns": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"function_names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceFunctionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	var functionARNs []string
	var functionNames []string

	input := &lambda.ListFunctionsInput{}
	pages := lambda.NewListFunctionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing Lambda Functions: %s", err)
		}

		for _, v := range page.Functions {
			functionARNs = append(functionARNs, aws.ToString(v.FunctionArn))
			functionNames = append(functionNames, aws.ToString(v.FunctionName))
		}
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("function_arns", functionARNs)
	d.Set("function_names", functionNames)

	return diags
}
