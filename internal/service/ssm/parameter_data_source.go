// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ssm_parameter", name="Parameter")
func dataSourceParameter() *schema.Resource {
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
			names.AttrVersion: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"with_decryption": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func dataParameterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	name := d.Get(names.AttrName).(string)
	param, err := findParameterByName(ctx, conn, name, d.Get("with_decryption").(bool))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSM Parameter (%s): %s", name, err)
	}

	d.SetId(aws.ToString(param.Name))
	d.Set(names.AttrARN, param.ARN)
	d.Set("insecure_value", nil)
	if param.Type != awstypes.ParameterTypeSecureString {
		d.Set("insecure_value", param.Value)
	}
	d.Set(names.AttrName, param.Name)
	d.Set(names.AttrType, param.Type)
	d.Set(names.AttrValue, param.Value)
	d.Set(names.AttrVersion, param.Version)

	return diags
}
