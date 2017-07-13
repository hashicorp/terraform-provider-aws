package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsCognitoUserPoolClient() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCognitoUserPoolClientCreate,
		Read:   resourceAwsCognitoUserPoolClientRead,
		Delete: resourceAwsCognitoUserPoolClientDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"user_pool": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"generate_secret": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true,
			},

			"secret": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsCognitoUserPoolClientCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.CreateUserPoolClientInput{
		ClientName:     aws.String(d.Get("name").(string)),
		UserPoolId:     aws.String(d.Get("user_pool").(string)),
		GenerateSecret: aws.Bool(d.Get("generate_secret").(bool)),
	}

	resp, err := conn.CreateUserPoolClient(params)

	if err != nil {
		return fmt.Errorf("Error creating Cognito User Pool Client: %s", err)
	}

	d.SetId(*resp.UserPoolClient.ClientId)

	return resourceAwsCognitoUserPoolClientRead(d, meta)
}

func resourceAwsCognitoUserPoolClientRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.DescribeUserPoolClientInput{
		ClientId:   aws.String(d.Id()),
		UserPoolId: aws.String(d.Get("user_pool").(string)),
	}

	resp, err := conn.DescribeUserPoolClient(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
			d.SetId("")
			return nil
		}
		return err
	}

	if resp.UserPoolClient.ClientSecret != nil {
		d.Set("secret", *resp.UserPoolClient.ClientSecret)
	}

	return nil
}

func resourceAwsCognitoUserPoolClientDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.DeleteUserPoolClientInput{
		ClientId:   aws.String(d.Id()),
		UserPoolId: aws.String(d.Get("user_pool").(string)),
	}

	_, err := conn.DeleteUserPoolClient(params)

	if err != nil {
		return fmt.Errorf("Error deleteing user pool client: %s", err)
	}

	return nil
}
