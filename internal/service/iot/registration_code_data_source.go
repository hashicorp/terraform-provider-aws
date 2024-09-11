// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_iot_registration_code", name="Registration Code")
func dataSourceRegistrationCode() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRegistrationCodeRead,

		Schema: map[string]*schema.Schema{
			"registration_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceRegistrationCodeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	input := &iot.GetRegistrationCodeInput{}

	output, err := conn.GetRegistrationCode(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Registration Code: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("registration_code", output.RegistrationCode)

	return diags
}
