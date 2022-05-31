package iot

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// https://docs.aws.amazon.com/iot/latest/apireference/API_CreateThingType.html
func ResourceThingType() *schema.Resource {
	return &schema.Resource{
		Create: resourceThingTypeCreate,
		Read:   resourceThingTypeRead,
		Update: resourceThingTypeUpdate,
		Delete: resourceThingTypeDelete,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("name", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validThingTypeName,
			},
			"properties": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validThingTypeDescription,
						},
						"searchable_attributes": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							ForceNew: true,
							MaxItems: 3,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validThingTypeSearchableAttribute,
							},
						},
					},
				},
			},
			"deprecated": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceThingTypeCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	params := &iot.CreateThingTypeInput{
		ThingTypeName: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("properties"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok && config != nil {
			params.ThingTypeProperties = expandThingTypeProperties(config)
		}
	}
	if len(tags) > 0 {
		params.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating IoT Thing Type: %s", params)
	out, err := conn.CreateThingType(params)

	if err != nil {
		return err
	}

	d.SetId(aws.StringValue(out.ThingTypeName))

	if v := d.Get("deprecated").(bool); v {
		params := &iot.DeprecateThingTypeInput{
			ThingTypeName: aws.String(d.Id()),
			UndoDeprecate: aws.Bool(false),
		}

		log.Printf("[DEBUG] Deprecating IoT Thing Type: %s", params)
		_, err := conn.DeprecateThingType(params)

		if err != nil {
			return err
		}
	}

	return resourceThingTypeRead(d, meta)
}

func resourceThingTypeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	params := &iot.DescribeThingTypeInput{
		ThingTypeName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading IoT Thing Type: %s", params)
	out, err := conn.DescribeThingType(params)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
			log.Printf("[WARN] IoT Thing Type %q not found, removing from state", d.Id())
			d.SetId("")
		}
		return err
	}

	if out.ThingTypeMetadata != nil {
		d.Set("deprecated", out.ThingTypeMetadata.Deprecated)
	}

	d.Set("arn", out.ThingTypeArn)

	tags, err := ListTags(conn, aws.StringValue(out.ThingTypeArn))
	if err != nil {
		return fmt.Errorf("error listing tags for IoT Thing Type (%s): %w", aws.StringValue(out.ThingTypeArn), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	if err := d.Set("properties", flattenThingTypeProperties(out.ThingTypeProperties)); err != nil {
		return fmt.Errorf("error setting properties: %s", err)
	}

	return nil
}

func resourceThingTypeUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	if d.HasChange("deprecated") {
		params := &iot.DeprecateThingTypeInput{
			ThingTypeName: aws.String(d.Id()),
			UndoDeprecate: aws.Bool(!d.Get("deprecated").(bool)),
		}

		log.Printf("[DEBUG] Updating IoT Thing Type: %s", params)
		_, err := conn.DeprecateThingType(params)

		if err != nil {
			return err
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceThingTypeRead(d, meta)
}

func resourceThingTypeDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	// In order to delete an IoT Thing Type, you must deprecate it first and wait
	// at least 5 minutes.
	deprecateParams := &iot.DeprecateThingTypeInput{
		ThingTypeName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Deprecating IoT Thing Type: %s", deprecateParams)
	_, err := conn.DeprecateThingType(deprecateParams)

	if err != nil {
		return err
	}

	deleteParams := &iot.DeleteThingTypeInput{
		ThingTypeName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Deleting IoT Thing Type: %s", deleteParams)

	err = resource.Retry(6*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteThingType(deleteParams)

		if err != nil {
			if tfawserr.ErrMessageContains(err, iot.ErrCodeInvalidRequestException, "Please wait for 5 minutes after deprecation and then retry") {
				return resource.RetryableError(err)
			}

			// As the delay post-deprecation is about 5 minutes, it may have been
			// deleted in between, thus getting a Not Found Exception.
			if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
				return nil
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteThingType(deleteParams)
		if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
			return nil
		}
	}
	if err != nil {
		return fmt.Errorf("Error deleting IOT thing type: %s", err)
	}
	return nil
}
