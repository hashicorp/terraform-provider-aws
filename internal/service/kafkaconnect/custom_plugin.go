package kafkaconnect

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceCustomPlugin() *schema.Resource {
	return &schema.Resource{
		Create: resourceCustomPluginCreate,
		Read:   resourceCustomPluginRead,
		Delete: schema.Noop,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(kafkaconnect.CustomPluginContentType_Values(), false),
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
			"location": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3": {
							Type:     schema.TypeList,
							MaxItems: 1,
							ForceNew: true,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bucket_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: verify.ValidARN,
									},
									"file_key": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"object_version": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCustomPluginCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KafkaConnectConn

	name := d.Get("name").(string)
	input := &kafkaconnect.CreateCustomPluginInput{
		Name:        aws.String(name),
		ContentType: aws.String(d.Get("content_type").(string)),
		Location:    expandLocation(d.Get("location").([]interface{})),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating MSK Connect Custom Plugin: %s", input)
	output, err := conn.CreateCustomPlugin(input)

	if err != nil {
		return fmt.Errorf("error creating MSK Connect Custom Plugin (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.CustomPluginArn))

	_, err = waitCustomPluginCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return fmt.Errorf("error waiting for MSK Connect Custom Plugin (%s) create: %w", d.Id(), err)
	}

	return resourceCustomPluginRead(d, meta)
}

func resourceCustomPluginRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KafkaConnectConn

	plugin, err := FindCustomPluginByARN(conn, d.Id())

	if tfresource.NotFound(err) && !d.IsNewResource() {
		log.Printf("[WARN] MSK Connect Custom Plugin (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading MSK Connect Custom Plugin (%s): %w", d.Id(), err)
	}

	d.Set("arn", plugin.CustomPluginArn)
	d.Set("description", plugin.Description)
	d.Set("name", plugin.Name)
	d.Set("state", plugin.CustomPluginState)

	if plugin.LatestRevision != nil {
		d.Set("content_type", plugin.LatestRevision.ContentType)
		d.Set("latest_revision", plugin.LatestRevision.Revision)
		if err := d.Set("location", flattenLocation(plugin.LatestRevision.Location)); err != nil {
			return fmt.Errorf("error setting location: %w", err)
		}
	} else {
		d.Set("content_type", nil)
		d.Set("latest_revision", nil)
		d.Set("location", nil)
	}

	return nil
}

func expandLocation(tfList []interface{}) *kafkaconnect.CustomPluginLocation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	location := tfList[0].(map[string]interface{})

	return &kafkaconnect.CustomPluginLocation{
		S3Location: expandS3Location(location["s3"].([]interface{})),
	}
}

func expandS3Location(tfList []interface{}) *kafkaconnect.S3Location {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	s3Location := kafkaconnect.S3Location{
		BucketArn: aws.String(tfMap["bucket_arn"].(string)),
		FileKey:   aws.String(tfMap["file_key"].(string)),
	}

	if objVer, ok := tfMap["object_version"].(string); ok && objVer != "" {
		s3Location.ObjectVersion = aws.String(objVer)
	}

	return &s3Location
}

func flattenLocation(apiLocation *kafkaconnect.CustomPluginLocationDescription) []interface{} {
	location := make(map[string]interface{})

	location["s3"] = flattenS3Location(apiLocation.S3Location)

	return []interface{}{location}
}

func flattenS3Location(apiS3Location *kafkaconnect.S3LocationDescription) []interface{} {
	location := make(map[string]interface{})

	location["bucket_arn"] = aws.StringValue(apiS3Location.BucketArn)
	location["file_key"] = aws.StringValue(apiS3Location.FileKey)

	if objVer := apiS3Location.ObjectVersion; objVer != nil {
		location["object_version"] = aws.StringValue(objVer)
	}

	return []interface{}{location}
}
