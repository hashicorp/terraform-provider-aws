package cognitoidp

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceIdentityProvider() *schema.Resource {
	return &schema.Resource{
		Create: resourceIdentityProviderCreate,
		Read:   resourceIdentityProviderRead,
		Update: resourceIdentityProviderUpdate,
		Delete: resourceIdentityProviderDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"attribute_mapping": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"idp_identifiers": {
				Type:     schema.TypeList,
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
				Elem:     &schema.Schema{Type: schema.TypeString},
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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(cognitoidentityprovider.IdentityProviderTypeType_Values(), false),
			},

			"user_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceIdentityProviderCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn
	log.Print("[DEBUG] Creating Cognito Identity Provider")

	providerName := d.Get("provider_name").(string)
	userPoolID := d.Get("user_pool_id").(string)
	params := &cognitoidentityprovider.CreateIdentityProviderInput{
		ProviderName: aws.String(providerName),
		ProviderType: aws.String(d.Get("provider_type").(string)),
		UserPoolId:   aws.String(userPoolID),
	}

	if v, ok := d.GetOk("attribute_mapping"); ok {
		params.AttributeMapping = flex.ExpandStringMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("provider_details"); ok {
		params.ProviderDetails = flex.ExpandStringMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("idp_identifiers"); ok {
		params.IdpIdentifiers = flex.ExpandStringList(v.([]interface{}))
	}

	_, err := conn.CreateIdentityProvider(params)
	if err != nil {
		return fmt.Errorf("Error creating Cognito Identity Provider: %w", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", userPoolID, providerName))

	return resourceIdentityProviderRead(d, meta)
}

func resourceIdentityProviderRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn
	log.Printf("[DEBUG] Reading Cognito Identity Provider: %s", d.Id())

	userPoolID, providerName, err := DecodeIdentityProviderID(d.Id())
	if err != nil {
		return err
	}

	ret, err := conn.DescribeIdentityProvider(&cognitoidentityprovider.DescribeIdentityProviderInput{
		ProviderName: aws.String(providerName),
		UserPoolId:   aws.String(userPoolID),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeResourceNotFoundException) {
		names.LogNotFoundRemoveState(names.CognitoIDP, names.ErrActionReading, ResIdentityProvider, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.Error(names.CognitoIDP, names.ErrActionReading, ResIdentityProvider, d.Id(), err)
	}

	if !d.IsNewResource() && (ret == nil || ret.IdentityProvider == nil) {
		names.LogNotFoundRemoveState(names.CognitoIDP, names.ErrActionReading, ResIdentityProvider, d.Id())
		d.SetId("")
		return nil
	}

	if d.IsNewResource() && (ret == nil || ret.IdentityProvider == nil) {
		return names.Error(names.CognitoIDP, names.ErrActionReading, ResIdentityProvider, d.Id(), errors.New("not found after creation"))
	}

	ip := ret.IdentityProvider
	d.Set("provider_name", ip.ProviderName)
	d.Set("provider_type", ip.ProviderType)
	d.Set("user_pool_id", ip.UserPoolId)

	if err := d.Set("attribute_mapping", aws.StringValueMap(ip.AttributeMapping)); err != nil {
		return fmt.Errorf("error setting attribute_mapping error: %w", err)
	}

	if err := d.Set("provider_details", aws.StringValueMap(ip.ProviderDetails)); err != nil {
		return fmt.Errorf("error setting provider_details error: %w", err)
	}

	if err := d.Set("idp_identifiers", flex.FlattenStringList(ip.IdpIdentifiers)); err != nil {
		return fmt.Errorf("error setting idp_identifiers error: %w", err)
	}

	return nil
}

func resourceIdentityProviderUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn
	log.Print("[DEBUG] Updating Cognito Identity Provider")

	userPoolID, providerName, err := DecodeIdentityProviderID(d.Id())
	if err != nil {
		return err
	}

	params := &cognitoidentityprovider.UpdateIdentityProviderInput{
		ProviderName: aws.String(providerName),
		UserPoolId:   aws.String(userPoolID),
	}

	if d.HasChange("attribute_mapping") {
		params.AttributeMapping = flex.ExpandStringMap(d.Get("attribute_mapping").(map[string]interface{}))
	}

	if d.HasChange("provider_details") {
		params.ProviderDetails = flex.ExpandStringMap(d.Get("provider_details").(map[string]interface{}))
	}

	if d.HasChange("idp_identifiers") {
		params.IdpIdentifiers = flex.ExpandStringList(d.Get("idp_identifiers").([]interface{}))
	}

	_, err = conn.UpdateIdentityProvider(params)
	if err != nil {
		return fmt.Errorf("Error updating Cognito Identity Provider: %w", err)
	}

	return resourceIdentityProviderRead(d, meta)
}

func resourceIdentityProviderDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn
	log.Printf("[DEBUG] Deleting Cognito Identity Provider: %s", d.Id())

	userPoolID, providerName, err := DecodeIdentityProviderID(d.Id())
	if err != nil {
		return err
	}

	_, err = conn.DeleteIdentityProvider(&cognitoidentityprovider.DeleteIdentityProviderInput{
		ProviderName: aws.String(providerName),
		UserPoolId:   aws.String(userPoolID),
	})

	if err != nil {
		if tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeResourceNotFoundException) {
			return nil
		}
		return err
	}

	return nil
}

func DecodeIdentityProviderID(id string) (string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("expected ID in format UserPoolID:ProviderName, received: %s", id)
	}
	return idParts[0], idParts[1], nil
}
