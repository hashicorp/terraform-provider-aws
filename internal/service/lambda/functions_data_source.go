// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
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
			}
		},
	}
}

func dataSourceFunctionsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.LambdaClient(ctx)

	var (
		functionARNs  []string
		functionNames []string
		input         lambda.ListFunctionsInput
	)
	for v, err := range listFunctions(ctx, conn, &input) {
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		functionARNs = append(functionARNs, aws.ToString(v.FunctionArn))
		functionNames = append(functionNames, aws.ToString(v.FunctionName))
	}

	d.SetId(c.Region(ctx))
	d.Set("function_arns", functionARNs)
	d.Set("function_names", functionNames)

	return diags
}
