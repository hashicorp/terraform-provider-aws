package aws

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsLexIntent() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLexIntentRead,

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
			"parent_intent_signature": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "$LATEST",
				ValidateFunc: validateStringMinMaxRegex(lexVersionMinLength, lexVersionMaxLength, lexVersionRegex),
			},
		},
	}
}

func dataSourceAwsLexIntentRead(d *schema.ResourceData, meta interface{}) error {
	// The data source and resource read functions are the same except the resource read expects to have the id set.
	d.SetId(d.Get("name").(string))

	return resourceAwsLexIntentRead(d, meta)
}
