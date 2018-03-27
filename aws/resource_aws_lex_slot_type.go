package aws

import (
	"fmt"

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
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validation.StringLenBetween(0, lexDescriptionMaxLength),
			},
			"enumeration_value": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: lexSlotTypeMinEnumerationValues,
				MaxItems: lexSlotTypeMaxEnumerationValues,
				Elem:     lexEnumerationValueResource,
			},
			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateLexName,
			},
			"value_selection_strategy": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "ORIGINAL_VALUE",
				ValidateFunc: validateLexSlotSelectionStrategy,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsLexSlotTypeCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn
	name := d.Get("name").(string)

	_, err := conn.PutSlotType(&lexmodelbuildingservice.PutSlotTypeInput{
		Description:       aws.String(d.Get("description").(string)),
		EnumerationValues: expandLexEnumerationValues(d.Get("enumeration_value")),
		Name:              aws.String(name),
		ValueSelectionStrategy: aws.String(d.Get("value_selection_strategy").(string)),
	})
	if err != nil {
		return fmt.Errorf("error creating Lex slot type %s: %s", name, err)
	}

	d.SetId(name)

	return resourceAwsLexSlotTypeRead(d, meta)
}

func resourceAwsLexSlotTypeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	resp, err := conn.GetSlotType(&lexmodelbuildingservice.GetSlotTypeInput{
		Name:    aws.String(d.Id()),
		Version: aws.String("$LATEST"),
	})
	if err != nil {
		return fmt.Errorf("error getting Lex slot type: %s", err)
	}

	d.Set("checksum", resp.Checksum)
	d.Set("created_date", resp.CreatedDate.UTC().String())
	d.Set("description", resp.Description)
	d.Set("enumeration_value", flattenLexEnumerationValues(resp.EnumerationValues))
	d.Set("last_updated_date", resp.LastUpdatedDate.UTC().String())
	d.Set("name", resp.Name)
	d.Set("value_selection_strategy", resp.ValueSelectionStrategy)
	d.Set("version", resp.Version)

	return nil
}

func resourceAwsLexSlotTypeUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn
	hasChanges := false

	input := &lexmodelbuildingservice.PutSlotTypeInput{
		Name:          aws.String(d.Id()),
		Checksum:      aws.String(d.Get("checksum").(string)),
		CreateVersion: aws.Bool(true),
		Description:   aws.String(d.Get("description").(string)),
	}

	// Description is always added otherwise the API call sees an empty description and unsets it.
	if d.HasChange("description") {
		hasChanges = true
	}
	if d.HasChange("enumeration_value") {
		input.EnumerationValues = expandLexEnumerationValues(d.Get("enumeration_value"))
		hasChanges = true
	}
	if d.HasChange("value_selection_strategy") {
		input.ValueSelectionStrategy = aws.String(d.Get("value_selection_strategy").(string))
		hasChanges = true
	}

	if hasChanges {
		_, err := conn.PutSlotType(input)
		if err != nil {
			return fmt.Errorf("error updating Lex slot type %s: %s", d.Id(), err)
		}
	}

	return resourceAwsLexSlotTypeRead(d, meta)
}

func resourceAwsLexSlotTypeDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	_, err := conn.DeleteSlotType(&lexmodelbuildingservice.DeleteSlotTypeInput{
		Name: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("error deleteing Lex slot type %s: %s", d.Id(), err)
	}

	return nil
}
