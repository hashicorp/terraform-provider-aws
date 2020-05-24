package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsSecurityHubAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSecurityHubAccountCreate,
		Read:   resourceAwsSecurityHubAccountRead,
		Update: resourceAwsSecurityHubAccountUpdate,
		Delete: resourceAwsSecurityHubAccountDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"enable_default_standards": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceAwsSecurityHubAccountCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	defaultStandard := d.Get("enable_default_standards").(bool)
	log.Print("[DEBUG] Enabling Security Hub for account")

	_, err := conn.EnableSecurityHub(&securityhub.EnableSecurityHubInput{
		EnableDefaultStandards: aws.Bool(defaultStandard),
	})

	if err != nil {
		return fmt.Errorf("Error enabling Security Hub for account: %s", err)
	}

	d.SetId(meta.(*AWSClient).accountid)

	return resourceAwsSecurityHubAccountRead(d, meta)
}

func resourceAwsSecurityHubAccountRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn

	log.Printf("[DEBUG] Checking if Security Hub is enabled")
	_, err := conn.GetEnabledStandards(&securityhub.GetEnabledStandardsInput{})

	if err != nil {
		// Can only read enabled standards if Security Hub is enabled
		if isAWSErr(err, "InvalidAccessException", "not subscribed to AWS Security Hub") {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error checking if Security Hub is enabled: %s", err)
	}

	return nil
}

func resourceAwsSecurityHubAccountUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn

	name := d.Id()
	log.Printf("[DEBUG] Updating Security Hub account: %q", name)

	if d.HasChange("enable_default_standards") {
		defaultStandard := d.Get("enable_default_standards").(bool)
		input := securityhub.EnableSecurityHubInput{
			EnableDefaultStandards: aws.Bool(defaultStandard),
		}
		log.Printf("[DEBUG] Setting default standards for Security Hub: %q: %s", name, input)
		if _, err := conn.EnableSecurityHub(&input); err != nil {
			return fmt.Errorf("[ERROR] Error updating Security Hub account (%s): %s", d.Id(), err)
		}
	}

	return nil
}

func resourceAwsSecurityHubAccountDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	log.Print("[DEBUG] Disabling Security Hub for account")

	_, err := conn.DisableSecurityHub(&securityhub.DisableSecurityHubInput{})

	if err != nil {
		return fmt.Errorf("Error disabling Security Hub for account: %s", err)
	}

	return nil
}
