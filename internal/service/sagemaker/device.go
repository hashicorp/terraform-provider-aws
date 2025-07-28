// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_device", name="Device")
func resourceDevice() *schema.Resource {
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

func resourceDeviceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	name := d.Get("device_fleet_name").(string)
	input := &sagemaker.RegisterDevicesInput{
		DeviceFleetName: aws.String(name),
		Devices:         expandDevice(d.Get("device").([]any)),
	}

	_, err := conn.RegisterDevices(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI Device %s: %s", name, err)
	}

	d.SetId(fmt.Sprintf("%s/%s", name, aws.ToString(input.Devices[0].DeviceName)))

	return append(diags, resourceDeviceRead(ctx, d, meta)...)
}

func resourceDeviceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	deviceFleetName, deviceName, err := decodeDeviceId(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Device (%s): %s", d.Id(), err)
	}
	device, err := findDeviceByName(ctx, conn, deviceFleetName, deviceName)
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Unable to find SageMaker AI Device (%s); removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Device (%s): %s", d.Id(), err)
	}

	d.Set("device_fleet_name", device.DeviceFleetName)
	d.Set("agent_version", device.AgentVersion)
	d.Set(names.AttrARN, device.DeviceArn)

	if err := d.Set("device", flattenDevice(device)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting device for SageMaker AI Device (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceDeviceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	deviceFleetName, _, err := decodeDeviceId(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating SageMaker AI Device (%s): %s", d.Id(), err)
	}

	input := &sagemaker.UpdateDevicesInput{
		DeviceFleetName: aws.String(deviceFleetName),
		Devices:         expandDevice(d.Get("device").([]any)),
	}

	log.Printf("[DEBUG] SageMaker AI Device update config: %#v", input)
	_, err = conn.UpdateDevices(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating SageMaker AI Device (%s): %s", d.Id(), err)
	}

	return append(diags, resourceDeviceRead(ctx, d, meta)...)
}

func resourceDeviceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	deviceFleetName, deviceName, err := decodeDeviceId(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI Device (%s): %s", d.Id(), err)
	}

	input := &sagemaker.DeregisterDevicesInput{
		DeviceFleetName: aws.String(deviceFleetName),
		DeviceNames:     []string{deviceName},
	}

	if _, err := conn.DeregisterDevices(ctx, input); err != nil {
		if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Device with name") ||
			tfawserr.ErrMessageContains(err, ErrCodeValidationException, "No device fleet with name") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI Device (%s): %s", d.Id(), err)
	}

	return diags
}

func findDeviceByName(ctx context.Context, conn *sagemaker.Client, deviceFleetName, deviceName string) (*sagemaker.DescribeDeviceOutput, error) {
	input := &sagemaker.DescribeDeviceInput{
		DeviceFleetName: aws.String(deviceFleetName),
		DeviceName:      aws.String(deviceName),
	}

	output, err := conn.DescribeDevice(ctx, input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "No device with name") ||
		tfawserr.ErrMessageContains(err, ErrCodeValidationException, "No device fleet with name") {
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

	return output, nil
}

func expandDevice(l []any) []awstypes.Device {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := awstypes.Device{
		DeviceName: aws.String(m[names.AttrDeviceName].(string)),
	}

	if v, ok := m[names.AttrDescription].(string); ok && v != "" {
		config.Description = aws.String(m[names.AttrDescription].(string))
	}

	if v, ok := m["iot_thing_name"].(string); ok && v != "" {
		config.IotThingName = aws.String(m["iot_thing_name"].(string))
	}

	return []awstypes.Device{config}
}

func flattenDevice(config *sagemaker.DescribeDeviceOutput) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		names.AttrDeviceName: aws.ToString(config.DeviceName),
	}

	if config.Description != nil {
		m[names.AttrDescription] = aws.ToString(config.Description)
	}

	if config.IotThingName != nil {
		m["iot_thing_name"] = aws.ToString(config.IotThingName)
	}

	return []map[string]any{m}
}

func decodeDeviceId(id string) (string, string, error) {
	iDParts := strings.Split(id, "/")
	if len(iDParts) != 2 {
		return "", "", fmt.Errorf("unexpected format of ID (%q), expected DEVICE-FLEET-NAME:DEVICE-NAME", id)
	}
	return iDParts[0], iDParts[1], nil
}
