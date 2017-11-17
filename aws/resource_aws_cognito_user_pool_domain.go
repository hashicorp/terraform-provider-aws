package aws

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsCognitoUserPoolDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCognitoUserPoolDomainCreate,
		Read:   resourceAwsCognitoUserPoolDomainRead,
		Delete: resourceAwsCognitoUserPoolDomainDelete,

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateCognitoUserPoolDomain,
			},
			"user_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsCognitoUserPoolDomainCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	input := &cognitoidentityprovider.CreateUserPoolDomainInput{
		Domain:     aws.String(d.Get("domain").(string)),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	_, err := conn.CreateUserPoolDomain(input)

	if err != nil {
		return err
	}

	d.SetId(d.Get("domain").(string))
	return resourceAwsCognitoUserPoolDomainRead(d, meta)
}

func resourceAwsCognitoUserPoolDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	domainStateConf := &resource.StateChangeConf{
		Pending:    []string{"CREATING", "UPDATING"},
		Target:     []string{"ACTIVE"},
		Refresh:    cognitoUserPoolDomainStateRefreshFunc(d.Id(), conn),
		Timeout:    10 * time.Minute,
		Delay:      3 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := domainStateConf.WaitForState()

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case cognitoidentityprovider.ErrCodeResourceNotFoundException:
				d.SetId("")
				return nil
			default:
				return err
			}
		}
		return err
	}

	return nil
}

func resourceAwsCognitoUserPoolDomainDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	input := &cognitoidentityprovider.DeleteUserPoolDomainInput{
		Domain:     aws.String(d.Id()),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	_, err := conn.DeleteUserPoolDomain(input)

	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}

func cognitoUserPoolDomainStateRefreshFunc(domain string, conn *cognitoidentityprovider.CognitoIdentityProvider) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &cognitoidentityprovider.DescribeUserPoolDomainInput{
			Domain: aws.String(domain),
		}
		out, err := conn.DescribeUserPoolDomain(input)
		if err != nil {
			return nil, "failed", err
		}
		if out.DomainDescription.Status == nil {
			return nil, "not found", nil
		}
		return out, *out.DomainDescription.Status, nil
	}
}
