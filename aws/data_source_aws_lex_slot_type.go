package aws

import (
	"github.com/hashicorp/terraform/helper/schema"
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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateStringMinMaxRegex(lexNameMinLength, lexNameMaxLength, lexNameRegex),
			},
			"value_selection_strategy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateStringMinMaxRegex(lexVersionMinLength, lexVersionMaxLength, lexVersionRegex),
			},
		},
	}
}

func dataSourceAwsLexSlotTypeRead(d *schema.ResourceData, meta interface{}) error {
	// The data source and resource read functions are the same except the resource read expects to have the id set.
	d.SetId(d.Get("name").(string))

	return resourceAwsLexSlotTypeRead(d, meta)
}
