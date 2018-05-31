package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsCognitoResourceServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCognitoResourceServerCreate,
		Read:   resourceAwsCognitoResourceServerRead,
		Update: resourceAwsCognitoResourceServerUpdate,
		Delete: resourceAwsCognitoResourceServerDelete,

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
				MaxItems: 25,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"scope_description": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateCognitoResourceServerScopeDescription,
						},
						"scope_name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateCognitoResourceServerScopeName,
						},
						"scope_identifier": {
							Type:     schema.TypeString,
							Computed: true,
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

func resourceAwsCognitoResourceServerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.CreateResourceServerInput{
		Identifier: aws.String(d.Get("identifier").(string)),
		Name:       aws.String(d.Get("name").(string)),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	if v, ok := d.GetOk("scope"); ok {
		configs := v.(*schema.Set).List()
		params.Scopes = expandCognitoResourceServerScope(configs)
	}

	log.Printf("[DEBUG] Creating Cognito Resource Server: %s", params)

	resp, err := conn.CreateResourceServer(params)

	if err != nil {
		return errwrap.Wrapf("Error creating Cognito Resource Server: {{err}}", err)
	}

	d.SetId(*resp.ResourceServer.Identifier)

	return resourceAwsCognitoResourceServerRead(d, meta)
}

func resourceAwsCognitoResourceServerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.DescribeResourceServerInput{
		Identifier: aws.String(d.Id()),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	log.Printf("[DEBUG] Reading Cognito Resource Server: %s", params)

	resp, err := conn.DescribeResourceServer(params)

	if err != nil {
		if isAWSErr(err, "ResourceNotFoundException", "") {
			log.Printf("[WARN] Cognito Resource Server %s is already gone", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.SetId(*resp.ResourceServer.Identifier)
	d.Set("name", *resp.ResourceServer.Name)
	d.Set("user_pool_id", *resp.ResourceServer.UserPoolId)

	scopes := flattenCognitoResourceServerScope(*resp.ResourceServer.Identifier, resp.ResourceServer.Scopes)
	if err := d.Set("scope", scopes); err != nil {
		return fmt.Errorf("Failed setting schema: %s", err)
	}

	var scopeIdentifiers []string
	for _, elem := range scopes {

		scopeIdentifier := elem["scope_identifier"].(string)
		scopeIdentifiers = append(scopeIdentifiers, scopeIdentifier)
	}
	d.Set("scope_identifiers", scopeIdentifiers)
	return nil
}

func resourceAwsCognitoResourceServerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.UpdateResourceServerInput{
		Identifier: aws.String(d.Id()),
		Name:       aws.String(d.Get("name").(string)),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	log.Printf("[DEBUG] Updating Cognito Resource Server: %s", params)

	_, err := conn.UpdateResourceServer(params)
	if err != nil {
		return errwrap.Wrapf("Error updating Cognito Resource Server: {{err}}", err)
	}

	return resourceAwsCognitoResourceServerRead(d, meta)
}

func resourceAwsCognitoResourceServerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.DeleteResourceServerInput{
		Identifier: aws.String(d.Id()),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	log.Printf("[DEBUG] Deleting Resource Server: %s", params)

	_, err := conn.DeleteResourceServer(params)

	if err != nil {
		return errwrap.Wrapf("Error deleting Resource Server: {{err}}", err)
	}

	return nil
}
