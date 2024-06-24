// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_device")
func ResourceDevice() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDeviceCreate,
		ReadWithoutTimeout:   resourceDeviceRead,
		UpdateWithoutTimeout: resourceDeviceUpdate,
		DeleteWithoutTimeout: resourceDeviceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"agent_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"device_fleet_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z](-*[0-9A-Za-z]){0,62}$`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"device": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDescription: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 40),
						},
						names.AttrDeviceName: {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 63),
						},
						"iot_thing_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 128),
						},
					},
				},
			},
		},
	}
}

func resourceDeviceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	name := d.Get("device_fleet_name").(string)
	input := &sagemaker.RegisterDevicesInput{
		DeviceFleetName: aws.String(name),
		Devices:         expandDevice(d.Get("device").([]interface{})),
	}

	_, err := conn.RegisterDevicesWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Device %s: %s", name, err)
	}

	d.SetId(fmt.Sprintf("%s/%s", name, aws.StringValue(input.Devices[0].DeviceName)))

	return append(diags, resourceDeviceRead(ctx, d, meta)...)
}

func resourceDeviceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	deviceFleetName, deviceName, err := DecodeDeviceId(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Device (%s): %s", d.Id(), err)
	}
	device, err := FindDeviceByName(ctx, conn, deviceFleetName, deviceName)
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Unable to find SageMaker Device (%s); removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Device (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(device.DeviceArn)
	d.Set("device_fleet_name", device.DeviceFleetName)
	d.Set("agent_version", device.AgentVersion)
	d.Set(names.AttrARN, arn)

	if err := d.Set("device", flattenDevice(device)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting device for SageMaker Device (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceDeviceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	deviceFleetName, _, err := DecodeDeviceId(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating SageMaker Device (%s): %s", d.Id(), err)
	}

	input := &sagemaker.UpdateDevicesInput{
		DeviceFleetName: aws.String(deviceFleetName),
		Devices:         expandDevice(d.Get("device").([]interface{})),
	}

	log.Printf("[DEBUG] SageMaker Device update config: %s", input.String())
	_, err = conn.UpdateDevicesWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating SageMaker Device (%s): %s", d.Id(), err)
	}

	return append(diags, resourceDeviceRead(ctx, d, meta)...)
}

func resourceDeviceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	deviceFleetName, deviceName, err := DecodeDeviceId(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker Device (%s): %s", d.Id(), err)
	}

	input := &sagemaker.DeregisterDevicesInput{
		DeviceFleetName: aws.String(deviceFleetName),
		DeviceNames:     []*string{aws.String(deviceName)},
	}

	if _, err := conn.DeregisterDevicesWithContext(ctx, input); err != nil {
		if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Device with name") ||
			tfawserr.ErrMessageContains(err, ErrCodeValidationException, "No device fleet with name") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker Device (%s): %s", d.Id(), err)
	}

	return diags
}

func expandDevice(l []interface{}) []*sagemaker.Device {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.Device{
		DeviceName: aws.String(m[names.AttrDeviceName].(string)),
	}

	if v, ok := m[names.AttrDescription].(string); ok && v != "" {
		config.Description = aws.String(m[names.AttrDescription].(string))
	}

	if v, ok := m["iot_thing_name"].(string); ok && v != "" {
		config.IotThingName = aws.String(m["iot_thing_name"].(string))
	}

	return []*sagemaker.Device{config}
}

func flattenDevice(config *sagemaker.DescribeDeviceOutput) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		names.AttrDeviceName: aws.StringValue(config.DeviceName),
	}

	if config.Description != nil {
		m[names.AttrDescription] = aws.StringValue(config.Description)
	}

	if config.IotThingName != nil {
		m["iot_thing_name"] = aws.StringValue(config.IotThingName)
	}

	return []map[string]interface{}{m}
}

func DecodeDeviceId(id string) (string, string, error) {
	iDParts := strings.Split(id, "/")
	if len(iDParts) != 2 {
		return "", "", fmt.Errorf("unexpected format of ID (%q), expected DEVICE-FLEET-NAME:DEVICE-NAME", id)
	}
	return iDParts[0], iDParts[1], nil
}
