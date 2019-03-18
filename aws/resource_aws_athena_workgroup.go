package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
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
			},
			"publish_cloudwatch_metrics_enable": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"output_location": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"encryption_option": {
				Type:     schema.TypeString,
				Required: true,
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

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	resp, err := conn.CreateWorkGroup(input)

	if err != nil {
		return err
	}

	d.SetId(name)

	return resourceAwsAthenaWorkgroupRead(d, meta)
}

func resourceAwsAthenaWorkgroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).athenaconn

	input := &athena.GetWorkGroupInput{
		WorkGroup: aws.String(d.Id()),
	}

	resp, err := conn.GetWorkGroup(input)

	if err != nil {
		if isAWSErr(err, athena.ErrCodeInvalidRequestException, d.Id()) {
			log.Printf("[WARN] Athena WorkGroup (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("name", resp.WorkGroup.Name)
	d.Set("description", resp.WorkGroup.Description)
	d.Set("bytes_scanned_cutoff_per_query", resp.WorkGroup.Configuration.BytesScannedCutoffPerQuery)
	d.Set("publish_cloudwatch_metrics_enabled", resp.WorkGroup.Configuration.PublishCloudWatchMetricsEnabled)
	d.Set("enforce_workgroup_configuration", resp.WorkGroup.Configuration.EnforceWorkGroupConfiguration)
	d.Set("output_location", resp.WorkGroup.Configuration.ResultConfiguration.OutputLocation)
	d.Set("encryption_option", resp.WorkGroup.Configuration.ResultConfiguration.EncryptionConfiguration.EncryptionOption)
	d.Set("kms_key", resp.WorkGroup.Configuration.ResultConfiguration.EncryptionConfiguration.KmsKey)

	return nil
}
