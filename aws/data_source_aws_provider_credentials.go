package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsProviderCredentials() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsProviderCredentialsRead,

		Schema: map[string]*schema.Schema{
			"access_key": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"secret_key": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"session_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsProviderCredentialsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).stsconn

	log.Printf("[DEBUG] Provider Credentials fetching credentials from session")
	creds, err := client.Config.Credentials.Get()
	if err != nil {
		return fmt.Errorf("Error getting credentials: %v", err)
	}

	d.SetId(time.Now().UTC().String())
	d.Set("access_key", creds.AccessKeyID)
	d.Set("secret_key", creds.SecretAccessKey)
	d.Set("session_token", creds.SessionToken)

	return nil
}
