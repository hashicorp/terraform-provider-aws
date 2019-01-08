package aws

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsLexBot() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLexBotRead,

		Schema: map[string]*schema.Schema{
			"checksum": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"child_directed": {
				Type:     schema.TypeBool,
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
			"failure_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"idle_session_ttl_in_seconds": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"locale": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateStringMinMaxRegex(lexNameMinLength, lexNameMaxLength, lexNameRegex),
			},
			"process_behavior": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateStringMinMaxRegex(lexVersionMinLength, lexVersionMaxLength, lexVersionRegex),
			},
			"voice_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsLexBotRead(d *schema.ResourceData, meta interface{}) error {
	// The data source and resource read functions are the same except the resource read expects to have the id set.
	d.SetId(d.Get("name").(string))

	return resourceAwsLexBotRead(d, meta)
}
