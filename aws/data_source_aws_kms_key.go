package aws

import "github.com/hashicorp/terraform/helper/schema"

func dataSourceAwsKmsKey() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsKmsKeyRead,
		Schema: map[string]*schema.Schema{
			"key_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateKmsKey,
			},
			"grant_tokens": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsKmsKeyRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}
