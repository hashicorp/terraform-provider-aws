package cognitoidp

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceResourceServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceResourceServerCreate,
		Read:   resourceResourceServerRead,
		Update: resourceResourceServerUpdate,
		Delete: resourceResourceServerDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		// https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_CreateResourceServer.html
		Schema: map[string]*schema.Schema{
			"identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"scope": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 100,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"scope_description": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 256),
						},
						"scope_name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validResourceServerScopeName,
						},
					},
				},
			},
			"user_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"scope_identifiers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceResourceServerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn

	identifier := d.Get("identifier").(string)
	userPoolID := d.Get("user_pool_id").(string)

	params := &cognitoidentityprovider.CreateResourceServerInput{
		Identifier: aws.String(identifier),
		Name:       aws.String(d.Get("name").(string)),
		UserPoolId: aws.String(userPoolID),
	}

	if v, ok := d.GetOk("scope"); ok {
		configs := v.(*schema.Set).List()
		params.Scopes = expandServerScope(configs)
	}

	log.Printf("[DEBUG] Creating Cognito Resource Server: %s", params)

	_, err := conn.CreateResourceServer(params)

	if err != nil {
		return fmt.Errorf("Error creating Cognito Resource Server: %s", err)
	}

	d.SetId(fmt.Sprintf("%s|%s", userPoolID, identifier))

	return resourceResourceServerRead(d, meta)
}

func resourceResourceServerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn

	userPoolID, identifier, err := DecodeResourceServerID(d.Id())
	if err != nil {
		return err
	}

	params := &cognitoidentityprovider.DescribeResourceServerInput{
		Identifier: aws.String(identifier),
		UserPoolId: aws.String(userPoolID),
	}

	log.Printf("[DEBUG] Reading Cognito Resource Server: %s", params)

	resp, err := conn.DescribeResourceServer(params)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeResourceNotFoundException) {
		names.LogNotFoundRemoveState(names.CognitoIDP, names.ErrActionReading, ResResourceServer, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.Error(names.CognitoIDP, names.ErrActionReading, ResResourceServer, d.Id(), err)
	}

	if !d.IsNewResource() && (resp == nil || resp.ResourceServer == nil) {
		names.LogNotFoundRemoveState(names.CognitoIDP, names.ErrActionReading, ResResourceServer, d.Id())
		d.SetId("")
		return nil
	}

	if d.IsNewResource() && (resp == nil || resp.ResourceServer == nil) {
		return names.Error(names.CognitoIDP, names.ErrActionReading, ResResourceServer, d.Id(), errors.New("not found after creation"))
	}

	d.Set("identifier", resp.ResourceServer.Identifier)
	d.Set("name", resp.ResourceServer.Name)
	d.Set("user_pool_id", resp.ResourceServer.UserPoolId)

	scopes := flattenServerScope(resp.ResourceServer.Scopes)
	if err := d.Set("scope", scopes); err != nil {
		return fmt.Errorf("Failed setting schema: %s", err)
	}

	var scopeIdentifiers []string
	for _, elem := range scopes {

		scopeIdentifier := fmt.Sprintf("%s/%s", aws.StringValue(resp.ResourceServer.Identifier), elem["scope_name"].(string))
		scopeIdentifiers = append(scopeIdentifiers, scopeIdentifier)
	}
	if err := d.Set("scope_identifiers", scopeIdentifiers); err != nil {
		return fmt.Errorf("error setting scope_identifiers: %s", err)
	}
	return nil
}

func resourceResourceServerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn

	userPoolID, identifier, err := DecodeResourceServerID(d.Id())
	if err != nil {
		return err
	}

	params := &cognitoidentityprovider.UpdateResourceServerInput{
		Identifier: aws.String(identifier),
		Name:       aws.String(d.Get("name").(string)),
		Scopes:     expandServerScope(d.Get("scope").(*schema.Set).List()),
		UserPoolId: aws.String(userPoolID),
	}

	log.Printf("[DEBUG] Updating Cognito Resource Server: %s", params)

	_, err = conn.UpdateResourceServer(params)
	if err != nil {
		return fmt.Errorf("Error updating Cognito Resource Server: %s", err)
	}

	return resourceResourceServerRead(d, meta)
}

func resourceResourceServerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn

	userPoolID, identifier, err := DecodeResourceServerID(d.Id())
	if err != nil {
		return err
	}

	params := &cognitoidentityprovider.DeleteResourceServerInput{
		Identifier: aws.String(identifier),
		UserPoolId: aws.String(userPoolID),
	}

	log.Printf("[DEBUG] Deleting Resource Server: %s", params)

	_, err = conn.DeleteResourceServer(params)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeResourceNotFoundException) {
			return nil
		}
		return fmt.Errorf("Error deleting Resource Server: %s", err)
	}

	return nil
}

func DecodeResourceServerID(id string) (string, string, error) {
	idParts := strings.Split(id, "|")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("expected ID in format UserPoolID|Identifier, received: %s", id)
	}
	return idParts[0], idParts[1], nil
}
