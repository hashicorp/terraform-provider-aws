package iot

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceThing() *schema.Resource {
	return &schema.Resource{
		Create: resourceThingCreate,
		Read:   resourceThingRead,
		Update: resourceThingUpdate,
		Delete: resourceThingDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attributes": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"default_client_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"thing_type_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceThingCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	name := d.Get("name").(string)
	input := &iot.CreateThingInput{
		ThingName: aws.String(name),
	}

	if v, ok := d.GetOk("attributes"); ok && len(v.(map[string]interface{})) > 0 {
		input.AttributePayload = &iot.AttributePayload{
			Attributes: flex.ExpandStringMap(v.(map[string]interface{})),
		}
	}

	if v, ok := d.GetOk("thing_type_name"); ok {
		input.ThingTypeName = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating IoT Thing: %s", input)
	output, err := conn.CreateThing(input)

	if err != nil {
		return fmt.Errorf("error creating IoT Thing (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.ThingName))

	return resourceThingRead(d, meta)
}

func resourceThingRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	output, err := FindThingByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT Thing (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading IoT Thing (%s): %w", d.Id(), err)
	}

	d.Set("arn", output.ThingArn)
	d.Set("default_client_id", output.DefaultClientId)
	d.Set("name", output.ThingName)
	d.Set("attributes", aws.StringValueMap(output.Attributes))
	d.Set("thing_type_name", output.ThingTypeName)
	d.Set("version", output.Version)

	return nil
}

func resourceThingUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	input := &iot.UpdateThingInput{
		ThingName: aws.String(d.Get("name").(string)),
	}

	if d.HasChange("attributes") {
		attributes := map[string]*string{}

		if v, ok := d.GetOk("attributes"); ok && len(v.(map[string]interface{})) > 0 {
			attributes = flex.ExpandStringMap(v.(map[string]interface{}))
		}

		input.AttributePayload = &iot.AttributePayload{
			Attributes: attributes,
		}
	}

	if d.HasChange("thing_type_name") {
		if v, ok := d.GetOk("thing_type_name"); ok {
			input.ThingTypeName = aws.String(v.(string))
		} else {
			input.RemoveThingType = aws.Bool(true)
		}
	}

	log.Printf("[DEBUG] Updating IoT Thing: %s", input)
	_, err := conn.UpdateThing(input)

	if err != nil {
		return fmt.Errorf("error updating IoT Thing (%s): %w", d.Id(), err)
	}

	return resourceThingRead(d, meta)
}

func resourceThingDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	log.Printf("[DEBUG] Deleting IoT Thing: %s", d.Id())
	_, err := conn.DeleteThing(&iot.DeleteThingInput{
		ThingName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting IoT Thing (%s): %w", d.Id(), err)
	}

	return nil
}
