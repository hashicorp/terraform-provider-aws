package aws

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsLexBotAlias() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLexBotAliasRead,

		Schema: map[string]*schema.Schema{
			"bot_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateStringMinMaxRegex(lexBotNameMinLength, lexBotNameMaxLength, lexNameRegex),
			},
			"bot_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
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
				ValidateFunc: validateStringMinMaxRegex(lexNameMinLength, lexNameMaxLength, lexNameRegex),
			},
		},
	}
}

func dataSourceAwsLexBotAliasRead(d *schema.ResourceData, meta interface{}) error {
	// The data source and resource read functions are the same except the resource read expects to have the id set.
	d.SetId(d.Get("name").(string))

	return resourceAwsLexBotAliasRead(d, meta)
}
