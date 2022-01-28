package kafkaconnect

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceWorkerConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceWorkerConfigurationCreate,
		Read:   resourceWorkerConfigurationRead,
		Delete: schema.Noop,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"latest_revision": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"properties_file_content": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				StateFunc: func(v interface{}) string {
					switch v := v.(type) {
					case string:
						return decodePropertiesFileContent(v)
					default:
						return ""
					}
				},
			},
		},
	}
}

func resourceWorkerConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KafkaConnectConn

	name := d.Get("name").(string)
	properties := d.Get("properties_file_content").(string)

	input := &kafkaconnect.CreateWorkerConfigurationInput{
		Name:                  aws.String(name),
		PropertiesFileContent: aws.String(verify.Base64Encode([]byte(properties))),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	log.Print("[DEBUG] Creating MSK Connect Worker Configuration")
	output, err := conn.CreateWorkerConfiguration(input)
	if err != nil {
		return fmt.Errorf("error creating MSK Connect Worker Configuration (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.WorkerConfigurationArn))

	return resourceWorkerConfigurationRead(d, meta)
}

func resourceWorkerConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KafkaConnectConn

	config, err := FindWorkerConfigurationByARN(conn, d.Id())

	if tfresource.NotFound(err) && !d.IsNewResource() {
		log.Printf("[WARN] MSK Connect Worker Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading MSK Connect Worker Configuration (%s): %w", d.Id(), err)
	}

	d.Set("arn", config.WorkerConfigurationArn)
	d.Set("name", config.Name)
	d.Set("description", config.Description)

	if config.LatestRevision != nil {
		d.Set("latest_revision", config.LatestRevision.Revision)
		d.Set("properties_file_content", decodePropertiesFileContent(aws.StringValue(config.LatestRevision.PropertiesFileContent)))
	} else {
		d.Set("latest_revision", nil)
		d.Set("properties_file_content", nil)
	}

	return nil
}

func decodePropertiesFileContent(content string) string {
	result, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return content
	}

	return string(result)
}
