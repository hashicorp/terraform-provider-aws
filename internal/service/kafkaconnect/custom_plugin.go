package kafkaconnect

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
			Create: schema.DefaultTimeout(customPluginCreateDefaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content_type": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice(kafkaconnect.CustomPluginContentType_Values(), false),
				Required:     true,
				ForceNew:     true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"location": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3": {
							Type: schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bucket_arn": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
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
							ForceNew: true,
							Required: true,
							MaxItems: 1,
						},
					},
				},
				Required: true,
				ForceNew: true,
				MaxItems: 1,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"latest_revision": {
				Type:     schema.TypeInt,
				Computed: true,
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

	log.Print("[DEBUG] Creating MSK Connect Custom Plugin")

	name := d.Get("name").(string)
	params := &kafkaconnect.CreateCustomPluginInput{
		Name:        aws.String(name),
		ContentType: aws.String(d.Get("content_type").(string)),
		Location:    expandLocation(d.Get("location").([]interface{})),
	}

	log.Print("[DEBUG] Setting Description")
	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	resp, err := conn.CreateCustomPlugin(params)
	if err != nil {
		return fmt.Errorf("Error creating custom plugin (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(resp.CustomPluginArn))

	_, err = waitCustomPluginCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return fmt.Errorf("Error waiting for MSK Connect Custom Plugin (%s) create: %w", d.Id(), err)
	}

	return resourceCustomPluginRead(d, meta)
}

func resourceCustomPluginRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KafkaConnectConn

	params := &kafkaconnect.DescribeCustomPluginInput{
		CustomPluginArn: aws.String(d.Id()),
	}

	plugin, err := conn.DescribeCustomPlugin(params)
	if err != nil {
		if tfresource.NotFound(&kafkaconnect.NotFoundException{}) {
			log.Printf("[WARN] Custom Plugin (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
	}

	d.Set("arn", plugin.CustomPluginArn)
	d.Set("description", plugin.Description)
	d.Set("content_type", plugin.LatestRevision.ContentType)
	d.Set("latest_revision", plugin.LatestRevision.Revision)
	d.Set("name", plugin.Name)
	d.Set("state", plugin.CustomPluginState)

	if err := d.Set("location", flattenLocation(plugin.LatestRevision.Location)); err != nil {
		return fmt.Errorf("error setting custom plugin's location (%s): %w", d.Id(), err)
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
		log.Printf("[DEBUG] Assigning object_version: %v", *objVer)
		location["object_version"] = aws.StringValue(objVer)
	}

	return []interface{}{location}
}
