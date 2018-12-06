package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsSecurityHubAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSecurityHubAccountCreate,
		Read:   resourceAwsSecurityHubAccountRead,
		Delete: resourceAwsSecurityHubAccountDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{},
	}
}

func resourceAwsSecurityHubAccountCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	log.Print("[DEBUG] Enabling Security Hub for account")

	_, err := conn.EnableSecurityHub(&securityhub.EnableSecurityHubInput{})

	if err != nil {
		// Ignore SerializationError: https://github.com/aws/aws-sdk-go/issues/2332
		if !isAWSErr(err, request.ErrCodeSerialization, "") {
			return fmt.Errorf("Error enabling Security Hub for account: %s", err)
		}
	}

	d.SetId("securityhub-account")

	return resourceAwsSecurityHubAccountRead(d, meta)
}

func resourceAwsSecurityHubAccountRead(d *schema.ResourceData, meta interface{}) error {
	// AWS does not currently publically expose API methods to check if an account is
	// enabled for Security Hub. The console uses a private API method called
	// isSecurityHubEnabled - this may become public in the future.
	return nil
}

func resourceAwsSecurityHubAccountDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	log.Print("[DEBUG] Disabling Security Hub for account")

	_, err := conn.DisableSecurityHub(&securityhub.DisableSecurityHubInput{})

	if err != nil {
		// Ignore SerializationError: https://github.com/aws/aws-sdk-go/issues/2332
		if !isAWSErr(err, request.ErrCodeSerialization, "") {
			return fmt.Errorf("Error disabling Security Hub for account: %s", err)
		}
	}

	return nil
}
