package configservice

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceConfigurationRecorderStatus() *schema.Resource {
	return &schema.Resource{
		Create: resourceConfigurationRecorderStatusPut,
		Read:   resourceConfigurationRecorderStatusRead,
		Update: resourceConfigurationRecorderStatusPut,
		Delete: resourceConfigurationRecorderStatusDelete,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("name", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"is_enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
		},
	}
}

func resourceConfigurationRecorderStatusPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigServiceConn

	name := d.Get("name").(string)
	d.SetId(name)

	if d.HasChange("is_enabled") {
		isEnabled := d.Get("is_enabled").(bool)
		if isEnabled {
			log.Printf("[DEBUG] Starting AWSConfig Configuration recorder %q", name)
			startInput := configservice.StartConfigurationRecorderInput{
				ConfigurationRecorderName: aws.String(name),
			}
			_, err := conn.StartConfigurationRecorder(&startInput)
			if err != nil {
				return fmt.Errorf("Failed to start Configuration Recorder: %s", err)
			}
		} else {
			log.Printf("[DEBUG] Stopping AWSConfig Configuration recorder %q", name)
			stopInput := configservice.StopConfigurationRecorderInput{
				ConfigurationRecorderName: aws.String(name),
			}
			_, err := conn.StopConfigurationRecorder(&stopInput)
			if err != nil {
				return fmt.Errorf("Failed to stop Configuration Recorder: %s", err)
			}
		}
	}

	return resourceConfigurationRecorderStatusRead(d, meta)
}

func resourceConfigurationRecorderStatusRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigServiceConn

	name := d.Id()
	statusInput := configservice.DescribeConfigurationRecorderStatusInput{
		ConfigurationRecorderNames: []*string{aws.String(name)},
	}
	statusOut, err := conn.DescribeConfigurationRecorderStatus(&statusInput)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchConfigurationRecorderException) {
		names.LogNotFoundRemoveState(names.ConfigService, names.ErrActionReading, "Configuration Recorder Status", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.Error(names.ConfigService, names.ErrActionReading, "Configuration Recorder Status", d.Id(), err)
	}

	numberOfStatuses := len(statusOut.ConfigurationRecordersStatus)
	if !d.IsNewResource() && numberOfStatuses < 1 {
		names.LogNotFoundRemoveState(names.ConfigService, names.ErrActionReading, "Configuration Recorder Status", d.Id())
		d.SetId("")
		return nil
	}

	if d.IsNewResource() && numberOfStatuses < 1 {
		return names.Error(names.ConfigService, names.ErrActionReading, "Configuration Recorder Status", d.Id(), errors.New("not found after creation"))
	}

	if numberOfStatuses > 1 {
		return fmt.Errorf("Expected exactly 1 Configuration Recorder (status), received %d: %#v",
			numberOfStatuses, statusOut.ConfigurationRecordersStatus)
	}

	d.Set("is_enabled", statusOut.ConfigurationRecordersStatus[0].Recording)

	return nil
}

func resourceConfigurationRecorderStatusDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigServiceConn
	input := configservice.StopConfigurationRecorderInput{
		ConfigurationRecorderName: aws.String(d.Get("name").(string)),
	}
	_, err := conn.StopConfigurationRecorder(&input)
	if err != nil {
		return fmt.Errorf("Stopping Configuration Recorder failed: %s", err)
	}

	return nil
}
