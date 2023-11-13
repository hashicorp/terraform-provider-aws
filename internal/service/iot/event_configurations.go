// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_iot_event_configurations", name="Event Configurations")
func ResourceEventConfigurations() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEventConfigurationsPut,
		ReadWithoutTimeout:   resourceEventConfigurationsRead,
		UpdateWithoutTimeout: resourceEventConfigurationsPut,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"event_configurations": {
				Type:             schema.TypeMap,
				Required:         true,
				Elem:             &schema.Schema{Type: schema.TypeBool},
				ValidateDiagFunc: verify.MapKeysAre(validation.ToDiagFunc(validation.StringInSlice(iot.EventType_Values(), false))),
			},
		},
	}
}

func resourceEventConfigurationsPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	input := &iot.UpdateEventConfigurationsInput{}

	if v, ok := d.GetOk("event_configurations"); ok && len(v.(map[string]interface{})) > 0 {
		input.EventConfigurations = tfmaps.ApplyToAllValues(v.(map[string]interface{}), func(v interface{}) *iot.Configuration {
			return &iot.Configuration{
				Enabled: aws.Bool(v.(bool)),
			}
		})
	}

	_, err := conn.UpdateEventConfigurationsWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating IoT Event Configurations (%s): %s", meta.(*conns.AWSClient).Region, err)
	}

	if d.IsNewResource() {
		d.SetId(meta.(*conns.AWSClient).Region)
	}

	return append(diags, resourceEventConfigurationsRead(ctx, d, meta)...)
}

func resourceEventConfigurationsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	output, err := findEventConfigurations(ctx, conn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT Event Configurations (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Event Configurations (%s): %s", d.Id(), err)
	}

	d.Set("event_configurations", tfmaps.ApplyToAllValues(output, func(v *iot.Configuration) bool {
		return aws.BoolValue(v.Enabled)
	}))

	return diags
}

func findEventConfigurations(ctx context.Context, conn *iot.IoT) (map[string]*iot.Configuration, error) {
	input := &iot.DescribeEventConfigurationsInput{}
	output, err := conn.DescribeEventConfigurationsWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.EventConfigurations) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.EventConfigurations, nil
}
