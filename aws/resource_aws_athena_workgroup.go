package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/hashicorp/terraform/helper/validation"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsAthenaWorkgroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAthenaWorkgroupCreate,
		Read:   resourceAwsAthenaWorkgroupRead,
		Update: resourceAwsAthenaWorkgroupUpdate,
		Delete: resourceAwsAthenaWorkgroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"bytes_scanned_cutoff_per_query": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(10485760),
			},
			"enforce_workgroup_configuration": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"publish_cloudwatch_metrics_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"output_location": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"encryption_option": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					athena.EncryptionOptionCseKms,
					athena.EncryptionOptionSseKms,
					athena.EncryptionOptionSseS3,
				}, false),
			},
			"kms_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"state": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  athena.WorkGroupStateEnabled,
				ValidateFunc: validation.StringInSlice([]string{
					athena.WorkGroupStateDisabled,
					athena.WorkGroupStateEnabled,
				}, false),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsAthenaWorkgroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).athenaconn

	name := d.Get("name").(string)

	input := &athena.CreateWorkGroupInput{
		Name: aws.String(name),
	}

	basicConfig := false
	resultConfig := false
	encryptConfig := false

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	inputConfiguration := &athena.WorkGroupConfiguration{}

	if v, ok := d.GetOk("bytes_scanned_cutoff_per_query"); ok {
		basicConfig = true
		inputConfiguration.BytesScannedCutoffPerQuery = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("enforce_workgroup_configuration"); ok {
		basicConfig = true
		inputConfiguration.EnforceWorkGroupConfiguration = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("publish_cloudwatch_metrics_enabled"); ok {
		basicConfig = true
		inputConfiguration.PublishCloudWatchMetricsEnabled = aws.Bool(v.(bool))
	}

	resultConfiguration := &athena.ResultConfiguration{}

	if v, ok := d.GetOk("output_location"); ok {
		resultConfig = true
		resultConfiguration.OutputLocation = aws.String(v.(string))
	}

	encryptionConfiguration := &athena.EncryptionConfiguration{}

	if v, ok := d.GetOk("encryption_option"); ok {
		resultConfig = true
		encryptConfig = true
		encryptionConfiguration.EncryptionOption = aws.String(v.(string))

		if v.(string) == athena.EncryptionOptionCseKms || v.(string) == athena.EncryptionOptionSseKms {
			if vkms, ok := d.GetOk("kms_key"); ok {
				encryptionConfiguration.KmsKey = aws.String(vkms.(string))
			} else {
				return fmt.Errorf("KMS Key required but not provided for encryption_option: %s", v.(string))
			}
		}
	}

	if basicConfig {
		input.Configuration = inputConfiguration
	}

	if resultConfig {
		input.Configuration.ResultConfiguration = resultConfiguration

		if encryptConfig {
			input.Configuration.ResultConfiguration.EncryptionConfiguration = encryptionConfiguration
		}
	}

	// Prevent the below error:
	// InvalidRequestException: Tags provided upon WorkGroup creation must not be empty
	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		input.Tags = tagsFromMapAthena(v)
	}

	_, err := conn.CreateWorkGroup(input)

	if err != nil {
		return fmt.Errorf("error creating Athena WorkGroup: %s", err)
	}

	d.SetId(name)

	if v := d.Get("state").(string); v == athena.WorkGroupStateDisabled {
		input := &athena.UpdateWorkGroupInput{
			State:     aws.String(v),
			WorkGroup: aws.String(d.Id()),
		}

		if _, err := conn.UpdateWorkGroup(input); err != nil {
			return fmt.Errorf("error disabling Athena WorkGroup (%s): %s", d.Id(), err)
		}
	}

	return resourceAwsAthenaWorkgroupRead(d, meta)
}

func resourceAwsAthenaWorkgroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).athenaconn

	input := &athena.GetWorkGroupInput{
		WorkGroup: aws.String(d.Id()),
	}

	resp, err := conn.GetWorkGroup(input)

	if isAWSErr(err, athena.ErrCodeInvalidRequestException, "is not found") {
		log.Printf("[WARN] Athena WorkGroup (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Athena WorkGroup (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "athena",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("workgroup/%s", d.Id()),
	}

	d.Set("arn", arn.String())
	d.Set("description", resp.WorkGroup.Description)
	d.Set("name", resp.WorkGroup.Name)
	d.Set("state", resp.WorkGroup.State)

	if resp.WorkGroup.Configuration != nil {
		d.Set("bytes_scanned_cutoff_per_query", resp.WorkGroup.Configuration.BytesScannedCutoffPerQuery)
		d.Set("enforce_workgroup_configuration", resp.WorkGroup.Configuration.EnforceWorkGroupConfiguration)
		d.Set("publish_cloudwatch_metrics_enabled", resp.WorkGroup.Configuration.PublishCloudWatchMetricsEnabled)

		if resp.WorkGroup.Configuration.ResultConfiguration != nil {
			d.Set("output_location", resp.WorkGroup.Configuration.ResultConfiguration.OutputLocation)

			if resp.WorkGroup.Configuration.ResultConfiguration.EncryptionConfiguration != nil {
				d.Set("encryption_option", resp.WorkGroup.Configuration.ResultConfiguration.EncryptionConfiguration.EncryptionOption)
				d.Set("kms_key", resp.WorkGroup.Configuration.ResultConfiguration.EncryptionConfiguration.KmsKey)
			}
		}
	}

	err = saveTagsAthena(conn, d, d.Get("arn").(string))

	if isAWSErr(err, athena.ErrCodeInvalidRequestException, "is not found") {
		log.Printf("[WARN] Athena WorkGroup (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsAthenaWorkgroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).athenaconn

	input := &athena.DeleteWorkGroupInput{
		WorkGroup: aws.String(d.Id()),
	}

	_, err := conn.DeleteWorkGroup(input)

	if err != nil {
		return fmt.Errorf("error deleting Athena WorkGroup (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceAwsAthenaWorkgroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).athenaconn

	workGroupUpdate := false
	resultConfigUpdate := false
	configUpdate := false
	encryptionUpdate := false
	removeEncryption := false

	input := &athena.UpdateWorkGroupInput{
		WorkGroup: aws.String(d.Get("name").(string)),
	}

	if d.HasChange("description") {
		workGroupUpdate = true
		input.Description = aws.String(d.Get("description").(string))
	}

	inputConfigurationUpdates := &athena.WorkGroupConfigurationUpdates{}

	if d.HasChange("bytes_scanned_cutoff_per_query") {
		workGroupUpdate = true
		configUpdate = true

		if v, ok := d.GetOk("bytes_scanned_cutoff_per_query"); ok {
			inputConfigurationUpdates.BytesScannedCutoffPerQuery = aws.Int64(int64(v.(int)))
		} else {
			inputConfigurationUpdates.RemoveBytesScannedCutoffPerQuery = aws.Bool(true)
		}
	}

	if d.HasChange("enforce_workgroup_configuration") {
		workGroupUpdate = true
		configUpdate = true

		v := d.Get("enforce_workgroup_configuration")
		inputConfigurationUpdates.EnforceWorkGroupConfiguration = aws.Bool(v.(bool))
	}

	if d.HasChange("publish_cloudwatch_metrics_enabled") {
		workGroupUpdate = true
		configUpdate = true

		v := d.Get("publish_cloudwatch_metrics_enabled")
		inputConfigurationUpdates.PublishCloudWatchMetricsEnabled = aws.Bool(v.(bool))
	}

	resultConfigurationUpdates := &athena.ResultConfigurationUpdates{}

	if d.HasChange("output_location") {
		workGroupUpdate = true
		configUpdate = true
		resultConfigUpdate = true

		if v, ok := d.GetOk("output_location"); ok {
			resultConfigurationUpdates.OutputLocation = aws.String(v.(string))
		} else {
			resultConfigurationUpdates.RemoveOutputLocation = aws.Bool(true)
		}
	}

	encryptionConfiguration := &athena.EncryptionConfiguration{}

	if d.HasChange("encryption_option") {
		workGroupUpdate = true
		configUpdate = true
		resultConfigUpdate = true
		encryptionUpdate = true

		if v, ok := d.GetOk("encryption_option"); ok {
			encryptionConfiguration.EncryptionOption = aws.String(v.(string))

			if v.(string) == athena.EncryptionOptionCseKms || v.(string) == athena.EncryptionOptionSseKms {
				if vkms, ok := d.GetOk("kms_key"); ok {
					encryptionConfiguration.KmsKey = aws.String(vkms.(string))
				} else {
					return fmt.Errorf("KMS Key required but not provided for encryption_option: %s", v.(string))
				}
			}
		} else {
			removeEncryption = true
			resultConfigurationUpdates.RemoveEncryptionConfiguration = aws.Bool(true)
		}
	}

	if d.HasChange("state") {
		workGroupUpdate = true
		input.State = aws.String(d.Get("state").(string))
	}

	if workGroupUpdate {
		if configUpdate {
			input.ConfigurationUpdates = inputConfigurationUpdates
		}

		if resultConfigUpdate {
			input.ConfigurationUpdates.ResultConfigurationUpdates = resultConfigurationUpdates

			if encryptionUpdate && !removeEncryption {
				input.ConfigurationUpdates.ResultConfigurationUpdates.EncryptionConfiguration = encryptionConfiguration
			}
		}

		_, err := conn.UpdateWorkGroup(input)

		if err != nil {
			return fmt.Errorf("error updating Athena WorkGroup (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		err := setTagsAthena(conn, d, d.Get("arn").(string))

		if err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsAthenaWorkgroupRead(d, meta)
}
