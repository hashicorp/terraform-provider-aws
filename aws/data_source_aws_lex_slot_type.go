package aws

import (
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func dataSourceAwsLexSlotType() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLexSlotTypeRead,

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
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(lexNameMinLength, lexNameMaxLength),
					validation.StringMatch(regexp.MustCompile(lexNameRegex), ""),
				),
			},
			"value_selection_strategy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(lexVersionMinLength, lexVersionMaxLength),
					validation.StringMatch(regexp.MustCompile(lexVersionRegex), ""),
				),
			},
		},
	}
}

func dataSourceAwsLexSlotTypeRead(d *schema.ResourceData, meta interface{}) error {
	slotTypeName := d.Get("name").(string)
	slotTypeVersion := "$LATEST"
	if v, ok := d.GetOk("version"); ok {
		slotTypeVersion = v.(string)
	}

	conn := meta.(*AWSClient).lexmodelconn

	resp, err := conn.GetSlotType(&lexmodelbuildingservice.GetSlotTypeInput{
		Name:    aws.String(slotTypeName),
		Version: aws.String(slotTypeVersion),
	})
	if err != nil {
		return fmt.Errorf("error getting slot type %s: %s", slotTypeName, err)
	}

	d.Set("checksum", resp.Checksum)
	d.Set("created_date", resp.CreatedDate.UTC().String())
	d.Set("description", resp.Description)
	d.Set("last_updated_date", resp.LastUpdatedDate.UTC().String())
	d.Set("name", resp.Name)
	d.Set("value_selection_strategy", resp.ValueSelectionStrategy)
	d.Set("version", resp.Version)

	d.SetId(slotTypeName)

	return nil
}
