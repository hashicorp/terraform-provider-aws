package iot

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceLoggingOptions() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLoggingOptionsPut,
		ReadWithoutTimeout:   resourceLoggingOptionsRead,
		UpdateWithoutTimeout: resourceLoggingOptionsPut,
		DeleteWithoutTimeout: schema.NoopContext,

		Schema: map[string]*schema.Schema{
			"default_log_level": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(iot.LogLevel_Values(), false),
			},
			"disable_all_logs": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceLoggingOptionsPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IoTConn

	input := &iot.SetV2LoggingOptionsInput{}

	if v, ok := d.GetOk("default_log_level"); ok {
		input.DefaultLogLevel = aws.String(v.(string))
	}

	if v, ok := d.GetOk("disable_all_logs"); ok {
		input.DisableAllLogs = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("role_arn"); ok {
		input.RoleArn = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhenAWSErrMessageContainsContext(ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.SetV2LoggingOptionsWithContext(ctx, input)
		},
		iot.ErrCodeInvalidRequestException, "If the role was just created or updated, please try again in a few seconds.",
	)

	if err != nil {
		return diag.Errorf("setting IoT logging options: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	return resourceLoggingOptionsRead(ctx, d, meta)
}

func resourceLoggingOptionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IoTConn

	output, err := conn.GetV2LoggingOptionsWithContext(ctx, &iot.GetV2LoggingOptionsInput{})

	if err != nil {
		return diag.Errorf("reading IoT logging options: %s", err)
	}

	d.Set("default_log_level", output.DefaultLogLevel)
	d.Set("disable_all_logs", output.DisableAllLogs)
	d.Set("role_arn", output.RoleArn)

	return nil
}
