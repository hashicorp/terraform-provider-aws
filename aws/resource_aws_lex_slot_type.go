package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsLexSlotType() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLexSlotTypeCreate,
		Read:   resourceAwsLexSlotTypeRead,
		Update: resourceAwsLexSlotTypeUpdate,
		Delete: resourceAwsLexSlotTypeDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"checksum": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(lexDescriptionMinLength, lexDescriptionMaxLength),
			},
			"enumeration_value": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: lexEnumerationValuesMin,
				MaxItems: lexEnumerationValuesMax,
				Elem:     lexEnumerationValueResource,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(lexSlotTypeMinLength, lexSlotTypeMaxLength),
					validation.StringMatch(regexp.MustCompile(lexSlotTypeRegex), ""),
				),
			},
			"value_selection_strategy": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  lexSlotTypeValueSelectionStrategyDefault,
				ValidateFunc: validation.StringInSlice([]string{
					lexmodelbuildingservice.SlotValueSelectionStrategyOriginalValue,
					lexmodelbuildingservice.SlotValueSelectionStrategyTopResolution,
				}, false),
			},
			"version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  lexVersionDefault,
				ValidateFunc: validation.All(
					validation.StringLenBetween(lexVersionMinLength, lexVersionMaxLength),
					validation.StringMatch(regexp.MustCompile(lexVersionRegex), ""),
				),
			},
		},
	}
}

func resourceAwsLexSlotTypeCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn
	name := d.Get("name").(string)

	input := &lexmodelbuildingservice.PutSlotTypeInput{
		Name:                   aws.String(name),
		ValueSelectionStrategy: aws.String(d.Get("value_selection_strategy").(string)),
	}

	// optional attributes

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("enumeration_value"); ok {
		input.EnumerationValues = expandLexEnumerationValues(expandLexSet(v.(*schema.Set)))
	}

	if _, err := conn.PutSlotType(input); err != nil {
		return fmt.Errorf("error creating slot type %s: %s", name, err)
	}

	d.SetId(name)

	return resourceAwsLexSlotTypeRead(d, meta)
}

func resourceAwsLexSlotTypeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	version := "$LATEST"
	if v, ok := d.GetOk("version"); ok {
		version = v.(string)
	}

	resp, err := conn.GetSlotType(&lexmodelbuildingservice.GetSlotTypeInput{
		Name:    aws.String(d.Id()),
		Version: aws.String(version),
	})
	if err != nil {
		if isAWSErr(err, "NotFoundException", "") {
			log.Printf("[WARN] Slot type (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error getting slot type %s: %s", d.Id(), err)
	}

	d.Set("checksum", resp.Checksum)
	d.Set("name", resp.Name)
	d.Set("value_selection_strategy", resp.ValueSelectionStrategy)
	d.Set("version", resp.Version)

	// optional attributes

	if resp.Description != nil {
		d.Set("description", resp.Description)
	}

	if resp.EnumerationValues != nil {
		d.Set("enumeration_value", flattenLexEnumerationValues(resp.EnumerationValues))
	}

	return nil
}

func resourceAwsLexSlotTypeUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	input := &lexmodelbuildingservice.PutSlotTypeInput{
		Name:                   aws.String(d.Id()),
		Checksum:               aws.String(d.Get("checksum").(string)),
		CreateVersion:          aws.Bool(true),
		ValueSelectionStrategy: aws.String(d.Get("value_selection_strategy").(string)),
	}

	// optional attributes

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("enumeration_value"); ok {
		input.EnumerationValues = expandLexEnumerationValues(expandLexSet(v.(*schema.Set)))
	}

	_, err := RetryOnAwsCodes([]string{"ConflictException"}, func() (interface{}, error) {
		return conn.PutSlotType(input)
	})
	if err != nil {
		return fmt.Errorf("error updating slot type %s: %s", d.Id(), err)
	}

	return resourceAwsLexSlotTypeRead(d, meta)
}

func resourceAwsLexSlotTypeDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	out, err := RetryOnAwsCodes([]string{"ConflictException"}, func() (interface{}, error) {
		return conn.DeleteSlotType(&lexmodelbuildingservice.DeleteSlotTypeInput{
			Name: aws.String(d.Id()),
		})
	})

	if err != nil {
		return fmt.Errorf("error deleteing slot type %s: %s %#v", d.Id(), err, out)
	}

	return nil
}
