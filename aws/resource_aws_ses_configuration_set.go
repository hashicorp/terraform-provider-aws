package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsSesConfigurationSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSesConfigurationSetCreate,
		Update: resourceAwsSesConfigurationSetUpdate,
		Read:   resourceAwsSesConfigurationSetRead,
		Delete: resourceAwsSesConfigurationSetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"delivery_options": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tls_policy": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  ses.TlsPolicyOptional,
							ValidateFunc: validation.StringInSlice([]string{
								ses.TlsPolicyRequire,
								ses.TlsPolicyOptional,
							}, false),
						},
					},
				},
			},
		},
	}
}

func resourceAwsSesConfigurationSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	configurationSetName := d.Get("name").(string)

	createOpts := &ses.CreateConfigurationSetInput{
		ConfigurationSet: &ses.ConfigurationSet{
			Name: aws.String(configurationSetName),
		},
	}

	_, err := conn.CreateConfigurationSet(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating SES configuration set: %s", err)
	}

	d.SetId(configurationSetName)

	return resourceAwsSesConfigurationSetUpdate(d, meta)
}

func resourceAwsSesConfigurationSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	configurationSetName := d.Get("name").(string)

	updateOpts := &ses.PutConfigurationSetDeliveryOptionsInput{
		ConfigurationSetName: aws.String(configurationSetName),
	}

	if v, ok := d.GetOk("delivery_options"); ok {
		options := v.(*schema.Set).List()
		delivery := options[0].(map[string]interface{})
		updateOpts.DeliveryOptions = &ses.DeliveryOptions{
			TlsPolicy: aws.String(delivery["tls_policy"].(string)),
		}
	}

	_, err := conn.PutConfigurationSetDeliveryOptions(updateOpts)
	if err != nil {
		return fmt.Errorf("Error updating SES configuration set: %s", err)
	}

	return resourceAwsSesConfigurationSetRead(d, meta)
}

func resourceAwsSesConfigurationSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	configSetInput := &ses.DescribeConfigurationSetInput{
		ConfigurationSetName:           aws.String(d.Id()),
		ConfigurationSetAttributeNames: aws.StringSlice([]string{ses.ConfigurationSetAttributeDeliveryOptions}),
	}

	response, err := conn.DescribeConfigurationSet(configSetInput)

	if err != nil {
		if isAWSErr(err, ses.ErrCodeConfigurationSetDoesNotExistException, "") {
			log.Printf("[WARN] SES Configuration Set (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if response.DeliveryOptions != nil {
		var deliveryOptions []map[string]interface{}
		tlsPolicy := map[string]interface{}{
			"tls_policy": response.DeliveryOptions.TlsPolicy,
		}
		deliveryOptions = append(deliveryOptions, tlsPolicy)

		if err := d.Set("delivery_options", deliveryOptions); err != nil {
			return fmt.Errorf("Error setting delivery_options for SES configuration set %s: %s", d.Id(), err)
		}
	}

	d.Set("name", aws.StringValue(response.ConfigurationSet.Name))

	return nil
}

func resourceAwsSesConfigurationSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	log.Printf("[DEBUG] SES Delete Configuration Rule Set: %s", d.Id())
	_, err := conn.DeleteConfigurationSet(&ses.DeleteConfigurationSetInput{
		ConfigurationSetName: aws.String(d.Id()),
	})

	return err
}
