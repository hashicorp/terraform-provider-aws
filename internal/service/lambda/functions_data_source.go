// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	DSNameFunctions = "Functions Data Source"
)

// @SDKDataSource("aws_lambda_functions")
func DataSourceFunctions() *schema.Resource {
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
	conn := meta.(*conns.AWSClient).LambdaConn(ctx)

	input := &lambda.ListFunctionsInput{}

	var functionARNs []string
	var functionNames []string

	err := conn.ListFunctionsPagesWithContext(ctx, input, func(page *lambda.ListFunctionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, function := range page.Functions {
			if function == nil {
				continue
			}

			functionARNs = append(functionARNs, aws.StringValue(function.FunctionArn))
			functionNames = append(functionNames, aws.StringValue(function.FunctionName))
		}

		return !lastPage
	})

	if err != nil {
		return create.DiagError(names.Lambda, create.ErrActionReading, DSNameFunctions, "", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("function_arns", functionARNs)
	d.Set("function_names", functionNames)

	return nil
}
