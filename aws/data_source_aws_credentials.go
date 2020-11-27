package aws

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsCredentials() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCredentialsRead,
		Schema: map[string]*schema.Schema{
			"access_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"secret_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceAwsCredentialsRead(d *schema.ResourceData, meta interface{}) error {
	providerCredentials := meta.(*AWSClient).credentials

	log.Printf("[DEBUG] Reading Provider Credentials")

	val, err := providerCredentials.Get()
	if err != nil {
		return fmt.Errorf("Error getting Provider Credentials: %v", err)
	}

	log.Printf("[DEBUG] Received Provider Credentials: %s", val.ProviderName)

	d.SetId(val.ProviderName)

	if val.HasKeys() {
		log.Printf("[DEBUG] Received provider has both AccessKeyID and SecretAccessKey value set")
		d.Set("access_key", val.AccessKeyID)
		d.Set("secret_key", val.SecretAccessKey)
	}

	d.Set("token", val.SessionToken)

	return nil
}
