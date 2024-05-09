// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ssm_parameter")
func DataSourceParameter() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataParameterRead,
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"insecure_value": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrValue: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"with_decryption": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			names.AttrVersion: {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataParameterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn(ctx)

	name := d.Get(names.AttrName).(string)

	paramInput := &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(d.Get("with_decryption").(bool)),
	}

	log.Printf("[DEBUG] Reading SSM Parameter: %s", paramInput)
	resp, err := conn.GetParameterWithContext(ctx, paramInput)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing SSM parameter (%s): %s", name, err)
	}

	param := resp.Parameter

	d.SetId(aws.StringValue(param.Name))
	d.Set(names.AttrARN, param.ARN)
	d.Set(names.AttrValue, param.Value)
	d.Set("insecure_value", nil)
	if aws.StringValue(param.Type) != ssm.ParameterTypeSecureString {
		d.Set("insecure_value", param.Value)
	}
	d.Set(names.AttrName, param.Name)
	d.Set(names.AttrType, param.Type)
	d.Set(names.AttrVersion, param.Version)

	return diags
}
