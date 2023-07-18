// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// @SDKDataSource("aws_lambda_function_url")
func DataSourceFunctionURL() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceFunctionURLRead,

		Schema: map[string]*schema.Schema{
			"authorization_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cors": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_credentials": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"allow_headers": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"allow_methods": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"allow_origins": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"expose_headers": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"max_age": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"function_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"function_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"function_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"invoke_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"qualifier": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"url_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceFunctionURLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LambdaConn(ctx)

	name := d.Get("function_name").(string)
	qualifier := d.Get("qualifier").(string)
	id := FunctionURLCreateResourceID(name, qualifier)
	output, err := FindFunctionURLByNameAndQualifier(ctx, conn, name, qualifier)

	if err != nil {
		return diag.Errorf("reading Lambda Function URL (%s): %s", id, err)
	}

	functionURL := aws.StringValue(output.FunctionUrl)

	d.SetId(id)
	d.Set("authorization_type", output.AuthType)
	if output.Cors != nil {
		if err := d.Set("cors", []interface{}{flattenCors(output.Cors)}); err != nil {
			return diag.Errorf("setting cors: %s", err)
		}
	} else {
		d.Set("cors", nil)
	}
	d.Set("creation_time", output.CreationTime)
	d.Set("function_arn", output.FunctionArn)
	d.Set("function_name", name)
	d.Set("function_url", functionURL)
	d.Set("invoke_mode", output.InvokeMode)
	d.Set("last_modified_time", output.LastModifiedTime)
	d.Set("qualifier", qualifier)

	// Function URL endpoints have the following format:
	// https://<url-id>.lambda-url.<region>.on.aws
	if v, err := url.Parse(functionURL); err != nil {
		return diag.Errorf("parsing URL (%s): %s", functionURL, err)
	} else if v := strings.Split(v.Host, "."); len(v) > 0 {
		d.Set("url_id", v[0])
	} else {
		d.Set("url_id", nil)
	}

	return nil
}
