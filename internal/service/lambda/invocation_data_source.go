// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"crypto/md5"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_lambda_invocation", name="Invocation")
func dataSourceInvocation() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceInvocationRead,

		Schema: map[string]*schema.Schema{
			"function_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"input": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsJSON,
			},
			"qualifier": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  FunctionVersionLatest,
			},
			"result": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceInvocationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	functionName := d.Get("function_name").(string)
	qualifier := d.Get("qualifier").(string)
	payload := []byte(d.Get("input").(string))

	input := &lambda.InvokeInput{
		FunctionName:   aws.String(functionName),
		InvocationType: awstypes.InvocationTypeRequestResponse,
		Payload:        payload,
		Qualifier:      aws.String(qualifier),
	}

	output, err := conn.Invoke(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "invoking Lambda Function (%s): %s", functionName, err)
	}

	if output.FunctionError != nil {
		return sdkdiag.AppendErrorf(diags, `invoking Lambda Function (%s): %s`, functionName, string(output.Payload))
	}

	d.SetId(fmt.Sprintf("%s_%s_%x", functionName, qualifier, md5.Sum(payload)))
	d.Set("result", string(output.Payload))

	return diags
}
