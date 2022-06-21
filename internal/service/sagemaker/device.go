package sagemaker

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceDevice() *schema.Resource {
	return &schema.Resource{
		Create: resourceDeviceCreate,
		Read:   resourceDeviceRead,
		Update: resourceDeviceUpdate,
		Delete: resourceDeviceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
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
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,62}$`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"device": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 40),
						},
						"device_name": {
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

func resourceDeviceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	name := d.Get("device_fleet_name").(string)
	input := &sagemaker.RegisterDevicesInput{
		DeviceFleetName: aws.String(name),
		Devices:         expandDevice(d.Get("device").([]interface{})),
	}

	_, err := conn.RegisterDevices(input)
	if err != nil {
		return fmt.Errorf("error creating SageMaker Device %s: %w", name, err)
	}

	d.SetId(fmt.Sprintf("%s/%s", name, aws.StringValue(input.Devices[0].DeviceName)))

	return resourceDeviceRead(d, meta)
}

func resourceDeviceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	deviceFleetName, deviceName, err := DecodeDeviceId(d.Id())
	if err != nil {
		return err
	}
	device, err := FindDeviceByName(conn, deviceFleetName, deviceName)
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Unable to find SageMaker Device (%s); removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading SageMaker Device (%s): %w", d.Id(), err)
	}

	arn := aws.StringValue(device.DeviceArn)
	d.Set("device_fleet_name", device.DeviceFleetName)
	d.Set("agent_version", device.AgentVersion)
	d.Set("arn", arn)

	if err := d.Set("device", flattenDevice(device)); err != nil {
		return fmt.Errorf("error setting device for SageMaker Device (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceDeviceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	deviceFleetName, _, err := DecodeDeviceId(d.Id())
	if err != nil {
		return err
	}

	input := &sagemaker.UpdateDevicesInput{
		DeviceFleetName: aws.String(deviceFleetName),
		Devices:         expandDevice(d.Get("device").([]interface{})),
	}

	log.Printf("[DEBUG] sagemaker Device update config: %s", input.String())
	_, err = conn.UpdateDevices(input)
	if err != nil {
		return fmt.Errorf("error updating SageMaker Device: %w", err)
	}

	return resourceDeviceRead(d, meta)
}

func resourceDeviceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	deviceFleetName, deviceName, err := DecodeDeviceId(d.Id())
	if err != nil {
		return err
	}

	input := &sagemaker.DeregisterDevicesInput{
		DeviceFleetName: aws.String(deviceFleetName),
		DeviceNames:     []*string{aws.String(deviceName)},
	}

	if _, err := conn.DeregisterDevices(input); err != nil {
		if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Device with name") ||
			tfawserr.ErrMessageContains(err, ErrCodeValidationException, "No device fleet with name") {
			return nil
		}
		return fmt.Errorf("error deleting SageMaker Device (%s): %w", d.Id(), err)
	}

	return nil
}

func expandDevice(l []interface{}) []*sagemaker.Device {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.Device{
		DeviceName: aws.String(m["device_name"].(string)),
	}

	if v, ok := m["description"].(string); ok && v != "" {
		config.Description = aws.String(m["description"].(string))
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
		"device_name": aws.StringValue(config.DeviceName),
	}

	if config.Description != nil {
		m["description"] = aws.StringValue(config.Description)
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
