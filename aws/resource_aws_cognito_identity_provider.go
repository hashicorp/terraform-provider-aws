package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsCognitoIdentityProvider() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCognitoIdentityProviderCreate,
		Read:   resourceAwsCognitoIdentityProviderRead,
		Update: resourceAwsCognitoIdentityProviderUpdate,
		Delete: resourceAwsCognitoIdentityProviderDelete,

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"attribute_mapping": {
				Type:     schema.TypeMap,
				Optional: true,
			},

			"idp_identifiers": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"provider_details": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
			},

			"provider_name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"provider_type": {
				Type:     schema.TypeString,
				Required: true,
			},

			"user_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsCognitoIdentityProviderCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn
	log.Print("[DEBUG] Creating Cognito Identity Provider")

	name := aws.String(d.Get("provider_name").(string))
	params := &cognitoidentityprovider.CreateIdentityProviderInput{
		ProviderName: name,
		ProviderType: aws.String(d.Get("provider_type").(string)),
		UserPoolId:   aws.String(d.Get("user_pool_id").(string)),
	}

	if v, ok := d.GetOk("attribute_mapping"); ok {
		params.AttributeMapping = expandCognitoIdentityProviderMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("provider_details"); ok {
		params.ProviderDetails = expandCognitoIdentityProviderMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("idp_identifiers"); ok {
		params.IdpIdentifiers = expandStringList(v.([]interface{}))
	}

	_, err := conn.CreateIdentityProvider(params)
	if err != nil {
		return fmt.Errorf("Error creating Cognito Identity Provider: %s", err)
	}

	d.SetId(*name)

	return resourceAwsCognitoIdentityProviderRead(d, meta)
}

func resourceAwsCognitoIdentityProviderRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn
	log.Printf("[DEBUG] Reading Cognito Identity Provider: %s", d.Id())

	ret, err := conn.DescribeIdentityProvider(&cognitoidentityprovider.DescribeIdentityProviderInput{
		ProviderName: aws.String(d.Id()),
		UserPoolId:   aws.String(d.Get("user_pool_id").(string)),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
			d.SetId("")
			return nil
		}
		return err
	}

	ip := ret.IdentityProvider
	d.Set("provider_name", ip.ProviderName)
	d.Set("provider_type", ip.ProviderType)
	d.Set("user_pool_id", ip.UserPoolId)

	if err := d.Set("attribute_mapping", flattenCognitoIdentityProviderMap(ip.AttributeMapping)); err != nil {
		return fmt.Errorf("[DEBUG] Error setting attribute_mapping error: %#v", err)
	}

	if err := d.Set("provider_details", flattenCognitoIdentityProviderMap(ip.ProviderDetails)); err != nil {
		return fmt.Errorf("[DEBUG] Error setting provider_details error: %#v", err)
	}

	if err := d.Set("idp_identifiers", flattenStringList(ip.IdpIdentifiers)); err != nil {
		return fmt.Errorf("[DEBUG] Error setting idp_identifiers error: %#v", err)
	}

	return nil
}

func resourceAwsCognitoIdentityProviderUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn
	log.Print("[DEBUG] Updating Cognito Identity Provider")

	params := &cognitoidentityprovider.UpdateIdentityProviderInput{
		UserPoolId:   aws.String(d.Get("user_pool_id").(string)),
		ProviderName: aws.String(d.Id()),
	}

	if d.HasChange("attribute_mapping") {
		params.AttributeMapping = expandCognitoIdentityProviderMap(d.Get("attribute_mapping").(map[string]interface{}))
	}

	if d.HasChange("provider_details") {
		params.ProviderDetails = expandCognitoIdentityProviderMap(d.Get("provider_details").(map[string]interface{}))
	}

	if d.HasChange("idp_identifiers") {
		params.IdpIdentifiers = expandStringList(d.Get("supported_login_providers").([]interface{}))
	}

	_, err := conn.UpdateIdentityProvider(params)
	if err != nil {
		return fmt.Errorf("Error updating Cognito Identity Provider: %s", err)
	}

	return resourceAwsCognitoIdentityProviderRead(d, meta)
}

func resourceAwsCognitoIdentityProviderDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn
	log.Printf("[DEBUG] Deleting Cognito Identity Provider: %s", d.Id())

	return resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := conn.DeleteIdentityProvider(&cognitoidentityprovider.DeleteIdentityProviderInput{
			ProviderName: aws.String(d.Id()),
			UserPoolId:   aws.String(d.Get("user_pool_id").(string)),
		})

		if err == nil {
			d.SetId("")
			return nil
		}

		return resource.NonRetryableError(err)
	})
}
