package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
				ValidateFunc: validateIotThingTypeName,
			},
			"properties": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validateIotThingTypeDescription,
						},
						"searchable_attributes": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							ForceNew: true,
							MaxItems: 3,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validateIotThingTypeSearchableAttribute,
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceThingTypeCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

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

	params := &iot.DescribeThingTypeInput{
		ThingTypeName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading IoT Thing Type: %s", params)
	out, err := conn.DescribeThingType(params)

	if err != nil {
		if tfawserr.ErrMessageContains(err, iot.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] IoT Thing Type %q not found, removing from state", d.Id())
			d.SetId("")
		}
		return err
	}

	if out.ThingTypeMetadata != nil {
		d.Set("deprecated", out.ThingTypeMetadata.Deprecated)
	}

	d.Set("arn", out.ThingTypeArn)

	if err := d.Set("properties", flattenIoTThingTypeProperties(out.ThingTypeProperties)); err != nil {
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
			if tfawserr.ErrMessageContains(err, iot.ErrCodeResourceNotFoundException, "") {
				return nil
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteThingType(deleteParams)
		if tfawserr.ErrMessageContains(err, iot.ErrCodeResourceNotFoundException, "") {
			return nil
		}
	}
	if err != nil {
		return fmt.Errorf("Error deleting IOT thing type: %s", err)
	}
	return nil
}
