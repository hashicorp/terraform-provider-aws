// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/aws/aws-sdk-go-v2/service/configservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_config_configuration_recorder_status", name="Configuration Recorder Status")
func resourceConfigurationRecorderStatus() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConfigurationRecorderStatusPut,
		ReadWithoutTimeout:   resourceConfigurationRecorderStatusRead,
		UpdateWithoutTimeout: resourceConfigurationRecorderStatusPut,
		DeleteWithoutTimeout: resourceConfigurationRecorderStatusDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"is_enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceConfigurationRecorderStatusPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	name := d.Get(names.AttrName).(string)

	if d.HasChange("is_enabled") {
		if d.Get("is_enabled").(bool) {
			input := &configservice.StartConfigurationRecorderInput{
				ConfigurationRecorderName: aws.String(name),
			}

			_, err := conn.StartConfigurationRecorder(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "starting ConfigService Configuration Recorder (%s): %s", name, err)
			}
		} else {
			input := &configservice.StopConfigurationRecorderInput{
				ConfigurationRecorderName: aws.String(name),
			}

			_, err := conn.StopConfigurationRecorder(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "stopping ConfigService Configuration Recorder (%s): %s", name, err)
			}
		}
	}

	d.SetId(name)

	return append(diags, resourceConfigurationRecorderStatusRead(ctx, d, meta)...)
}

func resourceConfigurationRecorderStatusRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	recorderStatus, err := findConfigurationRecorderStatusByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ConfigService Configuration Recorder Status (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ConfigService Configuration Recorder Status (%s): %s", d.Id(), err)
	}

	d.Set("is_enabled", recorderStatus.Recording)
	d.Set(names.AttrName, d.Id())

	return diags
}

func resourceConfigurationRecorderStatusDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	_, err := conn.StopConfigurationRecorder(ctx, &configservice.StopConfigurationRecorderInput{
		ConfigurationRecorderName: aws.String(d.Id()),
	})

	if errs.IsA[*types.NoSuchConfigurationRecorderException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "stopping ConfigService Configuration Recorder (%s): %s", d.Id(), err)
	}

	return diags
}

func findConfigurationRecorderStatusByName(ctx context.Context, conn *configservice.Client, name string) (*types.ConfigurationRecorderStatus, error) {
	input := &configservice.DescribeConfigurationRecorderStatusInput{
		ConfigurationRecorderNames: []string{name},
	}

	return findConfigurationRecorderStatus(ctx, conn, input)
}

func findConfigurationRecorderStatus(ctx context.Context, conn *configservice.Client, input *configservice.DescribeConfigurationRecorderStatusInput) (*types.ConfigurationRecorderStatus, error) {
	output, err := findConfigurationRecorderStatuses(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findConfigurationRecorderStatuses(ctx context.Context, conn *configservice.Client, input *configservice.DescribeConfigurationRecorderStatusInput) ([]types.ConfigurationRecorderStatus, error) {
	output, err := conn.DescribeConfigurationRecorderStatus(ctx, input)

	if errs.IsA[*types.NoSuchConfigurationRecorderException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ConfigurationRecordersStatus, nil
}
