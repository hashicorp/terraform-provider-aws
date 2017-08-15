package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsCognitoUserPool() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCognitoUserPoolCreate,
		Read:   resourceAwsCognitoUserPoolRead,
		Update: resourceAwsCognitoUserPoolUpdate,
		Delete: resourceAwsCognitoUserPoolDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"email_verification_subject": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"email_verification_message": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, es []error) {
					value := v.(string)
					if !strings.Contains(value, "{####}") {
						es = append(es, fmt.Errorf(
							"%q does not contain {####}", k))
					}
					return
				},
			},

			"mfa_configuration": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "OFF",
				ValidateFunc: func(v interface{}, k string) (ws []string, es []error) {
					value := v.(string)

					valid := map[string]bool{"OFF": true, "ON": true, "OPTIONAL": true}
					if !valid[value] {
						es = append(es, fmt.Errorf(
							"%q must be equal to OFF, ON, or OPTIONAL", k))
					}
					return
				},
			},
		},
	}
}

func resourceAwsCognitoUserPoolCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.CreateUserPoolInput{
		PoolName: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("email_verification_subject"); ok {
		params.EmailVerificationSubject = aws.String(v.(string))
	}

	if v, ok := d.GetOk("email_verification_message"); ok {
		params.EmailVerificationMessage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("mfa_configuration"); ok {
		params.MfaConfiguration = aws.String(v.(string))
	}

	resp, err := conn.CreateUserPool(params)

	if err != nil {
		return fmt.Errorf("Error creating Cognito User Pool: %s", err)
	}

	d.SetId(*resp.UserPool.Id)

	return nil
}

func resourceAwsCognitoUserPoolRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.DescribeUserPoolInput{
		UserPoolId: aws.String(d.Id()),
	}

	resp, err := conn.DescribeUserPool(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
			d.SetId("")
			return nil
		}
		return err
	}

	if resp.UserPool.EmailVerificationSubject != nil {
		d.Set("email_verification_subject", *resp.UserPool.EmailVerificationSubject)
	}

	if resp.UserPool.EmailVerificationMessage != nil {
		d.Set("email_verification_message", *resp.UserPool.EmailVerificationMessage)
	}

	if resp.UserPool.MfaConfiguration != nil {
		d.Set("mfa_configuration", *resp.UserPool.MfaConfiguration)
	}

	return nil
}

func resourceAwsCognitoUserPoolUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.UpdateUserPoolInput{
		UserPoolId: aws.String(d.Id()),
	}

	if d.HasChange("email_verification_subject") {
		params.EmailVerificationSubject = aws.String(d.Get("email_verification_subject").(string))
	}

	if d.HasChange("email_verification_message") {
		params.EmailVerificationMessage = aws.String(d.Get("email_verification_message").(string))
	}

	if d.HasChange("mfa_configuration") {
		params.MfaConfiguration = aws.String(d.Get("mfa_configuration").(string))
	}

	_, err := conn.UpdateUserPool(params)
	if err != nil {
		return fmt.Errorf("Error updating Cognito User pool: %s", err)
	}

	return nil
}

func resourceAwsCognitoUserPoolDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.DeleteUserPoolInput{
		UserPoolId: aws.String(d.Id()),
	}

	_, err := conn.DeleteUserPool(params)

	if err != nil {
		return fmt.Errorf("Error deleting user pool: %s", err)
	}

	return nil
}
