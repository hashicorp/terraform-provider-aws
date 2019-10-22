package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsCognitoUserPoolIdentityProvider() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCognitoUserPoolIdentityProviderCreate,
		Read:   resourceAwsCognitoUserPoolIdentityProviderRead,
		Update: resourceAwsCognitoUserPoolIdentityProviderUpdate,
		Delete: resourceAwsCognitoUserPoolIdentityProviderDelete,

		Importer: &schema.ResourceImporter{
			State: resourceAwsCognitoUserPoolIdentityProviderImport,
		},

		// https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_CreateIdentityProvider.html
		Schema: map[string]*schema.Schema{
			"attribute_mapping": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 32),
				},
			},

			"idp_identifiers": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 50,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 40),
						validation.StringMatch(regexp.MustCompile(`^[\w\s+=.@-]+$`), "see https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_CreateIdentityProvider.html#API_CreateIdentityProvider_RequestSyntax"),
					),
				},
			},

			"provider_details": {
				Type:     schema.TypeMap,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"provider_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
					validation.StringMatch(regexp.MustCompile(`^[^_][\p{L}\p{M}\p{S}\p{N}\p{P}][^_]+$`), "see https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_CreateIdentityProvider.html#API_CreateIdentityProvider_RequestSyntax"),
				),
			},

			"provider_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					cognitoidentityprovider.IdentityProviderTypeTypeSaml,
					cognitoidentityprovider.IdentityProviderTypeTypeFacebook,
					cognitoidentityprovider.IdentityProviderTypeTypeGoogle,
					cognitoidentityprovider.IdentityProviderTypeTypeLoginWithAmazon,
					cognitoidentityprovider.IdentityProviderTypeTypeOidc,
				}, false),
			},

			"user_pool_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateCognitoUserPoolId,
			},
		},
	}
}

func resourceAwsCognitoUserPoolIdentityProviderCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.CreateIdentityProviderInput{
		ProviderName:    aws.String(d.Get("provider_name").(string)),
		ProviderType:    aws.String(d.Get("provider_type").(string)),
		UserPoolId:      aws.String(d.Get("user_pool_id").(string)),
		ProviderDetails: stringMapToPointers(d.Get("provider_details").(map[string]interface{})),
	}

	if v, ok := d.GetOk("attribute_mapping"); ok {
		params.AttributeMapping = stringMapToPointers(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("idp_identifiers"); ok {
		params.IdpIdentifiers = expandStringList(v.([]interface{}))
	}

	log.Printf("[DEBUG] Creating Cognito User Pool Identity Provider: %s", params)

	resp, err := conn.CreateIdentityProvider(params)

	if err != nil {
		return fmt.Errorf("Error creating Cognito User Pool Identity Provider: %s", err)
	}

	d.SetId(*resp.IdentityProvider.ProviderName)

	return resourceAwsCognitoUserPoolIdentityProviderRead(d, meta)
}

func resourceAwsCognitoUserPoolIdentityProviderRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.DescribeIdentityProviderInput{
		UserPoolId:   aws.String(d.Get("user_pool_id").(string)),
		ProviderName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Cognito User Pool Identity Provider: %s", params)

	resp, err := conn.DescribeIdentityProvider(params)

	if err != nil {
		if isAWSErr(err, "ResourceNotFoundException", "") {
			log.Printf("[WARN] Cognito User Pool Identity Provider %s is already gone", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.SetId(*resp.IdentityProvider.ProviderName)
	d.Set("provider_name", *resp.IdentityProvider.ProviderName)
	d.Set("user_pool_id", *resp.IdentityProvider.UserPoolId)
	d.Set("provider_type", *resp.IdentityProvider.ProviderType)
	if err := d.Set("idp_identifiers", flattenStringList(resp.IdentityProvider.IdpIdentifiers)); err != nil {
		return fmt.Errorf("error setting idp_identifiers for resource %s: %s", d.Id(), err)
	}
	if err := d.Set("attribute_mapping", pointersMapToStringList(resp.IdentityProvider.AttributeMapping)); err != nil {
		return fmt.Errorf("error setting attribute_mapping for resource %s: %s", d.Id(), err)
	}
	if err := d.Set("provider_details", pointersMapToStringList(resp.IdentityProvider.ProviderDetails)); err != nil {
		return fmt.Errorf("error setting provider_details for resource %s: %s", d.Id(), err)
	}

	return nil
}

func resourceAwsCognitoUserPoolIdentityProviderUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.UpdateIdentityProviderInput{
		ProviderName: aws.String(d.Id()),
		UserPoolId:   aws.String(d.Get("user_pool_id").(string)),
	}

	if v, ok := d.GetOk("provider_details"); ok {
		params.ProviderDetails = stringMapToPointers(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("attribute_mapping"); ok {
		params.AttributeMapping = stringMapToPointers(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("idp_identifiers"); ok {
		params.IdpIdentifiers = expandStringList(v.([]interface{}))
	}

	log.Printf("[DEBUG] Updating Cognito User Pool Identity Provider: %s", params)

	_, err := conn.UpdateIdentityProvider(params)
	if err != nil {
		return fmt.Errorf("Error updating Cognito User Pool Identity Provider: %s", err)
	}

	return resourceAwsCognitoUserPoolIdentityProviderRead(d, meta)
}

func resourceAwsCognitoUserPoolIdentityProviderDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.DeleteIdentityProviderInput{
		ProviderName: aws.String(d.Id()),
		UserPoolId:   aws.String(d.Get("user_pool_id").(string)),
	}

	log.Printf("[DEBUG] Deleting Cognito User Pool Identity Provider: %s", params)

	_, err := conn.DeleteIdentityProvider(params)

	if err != nil {
		return fmt.Errorf("Error deleting Cognito User Pool Identity Provider: %s", err)
	}

	return nil
}

func resourceAwsCognitoUserPoolIdentityProviderImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 2 || len(d.Id()) < 3 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'user-pool-id/provider-name'", d.Id())
	}
	userPoolId := strings.Split(d.Id(), "/")[0]
	providerName := strings.Split(d.Id(), "/")[1]
	d.SetId(providerName)
	d.Set("user_pool_id", userPoolId)
	log.Printf("[DEBUG] Importing identity provider %s for user pool %s", providerName, userPoolId)

	return []*schema.ResourceData{d}, nil
}
