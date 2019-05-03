package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsCallerIdentity() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCallerIdentityRead,

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

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

func dataSourceAwsCallerIdentityRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).stsconn

	log.Printf("[DEBUG] Reading Caller Identity")
	res, err := client.GetCallerIdentity(&sts.GetCallerIdentityInput{})

	if err != nil {
		return fmt.Errorf("Error getting Caller Identity: %v", err)
	}

	log.Printf("[DEBUG] Received Caller Identity: %s", res)

	creds, err := client.Config.Credentials.Get()
	if err != nil {
		return fmt.Errorf("Error getting credentials: %v", err)
	}

	d.SetId(time.Now().UTC().String())
	d.Set("account_id", res.Account)
	d.Set("arn", res.Arn)
	d.Set("user_id", res.UserId)
	d.Set("access_key", creds.AccessKeyID)
	d.Set("secret_key", creds.SecretAccessKey)
	d.Set("session_token", creds.SessionToken)

	return nil
}
