package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

func resourceAwsCognitoUserPoolAppClient() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCognitoUserPoolAppClientCreate,
		Read:   resourceAwsCognitoUserPoolAppClientRead,
		Update: resourceAwsCognitoUserPoolAppClientUpdate,
		Delete: resourceAwsCognitoUserPoolAppClientDelete,

		Schema: map[string]*schema.Schema{
			"client_secret": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"generate_secret": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"read_attributes": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"refresh_token_validity": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  30,
			},
			"user_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"write_attributes": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceAwsCognitoUserPoolAppClientCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.CreateUserPoolClientInput{
		ClientName: aws.String(d.Get("name").(string)),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	if v, ok := d.GetOk("generate_secret"); ok {
		params.GenerateSecret = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("refresh_token_validity"); ok {
		params.RefreshTokenValidity = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("read_attributes"); ok {
		params.ReadAttributes = expandStringList(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("write_attributes"); ok {
		params.WriteAttributes = expandStringList(v.(*schema.Set).List())
	}

	log.Printf("[DEBUG] Creating Cognito User Pool App Client: %s", params)

	resp, err := conn.CreateUserPoolClient(params)
	if err != nil {
		return errwrap.Wrapf("Error creating Cognito User Pool App Client: {{err}}", err)
	}

	d.SetId(*resp.UserPoolClient.ClientId)

	return resourceAwsCognitoUserPoolAppClientRead(d, meta)
}

func resourceAwsCognitoUserPoolAppClientRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.DescribeUserPoolClientInput{
		ClientId:   aws.String(d.Id()),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	log.Printf("[DEBUG] Reading Cognito User Pool: %s", params)

	resp, err := conn.DescribeUserPoolClient(params)
	if err != nil {
		return errwrap.Wrapf("Error reading Cognito User Pool App Client: {{err}}", err)
	}

	if resp.UserPoolClient.ClientSecret != nil {
		d.Set("client_secret", *resp.UserPoolClient.ClientName)
	}

	d.Set("name", *resp.UserPoolClient.ClientName)
	d.Set("refresh_token_validity", *resp.UserPoolClient.RefreshTokenValidity)

	d.Set("read_attributes", flattenStringList(resp.UserPoolClient.ReadAttributes))
	d.Set("write_attributes", flattenStringList(resp.UserPoolClient.WriteAttributes))

	return nil
}

func resourceAwsCognitoUserPoolAppClientUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.UpdateUserPoolClientInput{
		ClientId:   aws.String(d.Id()),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	if d.HasChange("name") {
		params.ClientName = aws.String(d.Get("name").(string))
	}

	if d.HasChange("refresh_token_validity") {
		params.RefreshTokenValidity = aws.Int64(int64(d.Get("refresh_token_validity").(int)))
	}

	if v, ok := d.GetOk("read_attributes"); ok {
		params.ReadAttributes = expandStringList(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("write_attributes"); ok {
		params.WriteAttributes = expandStringList(v.(*schema.Set).List())
	}

	log.Printf("[DEBUG] Updating Cognito User Pool: %s", params)

	_, err := conn.UpdateUserPoolClient(params)
	if err != nil {
		return errwrap.Wrapf("Error updating Cognito User Pool App Client: {{err}}", err)
	}

	return resourceAwsCognitoUserPoolAppClientRead(d, meta)
}

func resourceAwsCognitoUserPoolAppClientDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.DeleteUserPoolClientInput{
		ClientId:   aws.String(d.Id()),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	log.Printf("[DEBUG] Deleting Cognito User Pool App Client: %s", params)

	_, err := conn.DeleteUserPoolClient(params)

	if err != nil {
		return errwrap.Wrapf("Error deleting user pool App Client: {{err}}", err)
	}

	return nil
}
