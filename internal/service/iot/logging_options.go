// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iot/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iot_logging_options", name="Logging Options")
func resourceLoggingOptions() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLoggingOptionsPut,
		ReadWithoutTimeout:   resourceLoggingOptionsRead,
		UpdateWithoutTimeout: resourceLoggingOptionsPut,
		DeleteWithoutTimeout: schema.NoopContext,

		Schema: map[string]*schema.Schema{
			"default_log_level": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.LogLevel](),
			},
			"disable_all_logs": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceLoggingOptionsPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	input := &iot.SetV2LoggingOptionsInput{}

	if v, ok := d.GetOk("default_log_level"); ok {
		input.DefaultLogLevel = awstypes.LogLevel(v.(string))
	}

	if v, ok := d.GetOk("disable_all_logs"); ok {
		input.DisableAllLogs = v.(bool)
	}

	if v, ok := d.GetOk(names.AttrRoleARN); ok {
		input.RoleArn = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhenIsA[*awstypes.InvalidRequestException](ctx, propagationTimeout, func() (interface{}, error) {
		return conn.SetV2LoggingOptions(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting IoT Logging Options: %s", err)
	}

	if d.IsNewResource() {
		d.SetId(meta.(*conns.AWSClient).Region)
	}

	return append(diags, resourceLoggingOptionsRead(ctx, d, meta)...)
}

func resourceLoggingOptionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	output, err := conn.GetV2LoggingOptions(ctx, &iot.GetV2LoggingOptionsInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Logging Options (%s): %s", d.Id(), err)
	}

	d.Set("default_log_level", output.DefaultLogLevel)
	d.Set("disable_all_logs", output.DisableAllLogs)
	d.Set(names.AttrRoleARN, output.RoleArn)

	return diags
}
