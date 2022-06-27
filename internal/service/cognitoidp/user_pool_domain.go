package cognitoidp

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceUserPoolDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceUserPoolDomainCreate,
		Read:   resourceUserPoolDomainRead,
		Delete: resourceUserPoolDomainDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 63),
			},
			"certificate_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"user_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"aws_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloudfront_distribution_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"s3_bucket": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUserPoolDomainCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn

	domain := d.Get("domain").(string)

	timeout := 1 * time.Minute //Default timeout for a basic domain

	params := &cognitoidentityprovider.CreateUserPoolDomainInput{
		Domain:     aws.String(domain),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	if v, ok := d.GetOk("certificate_arn"); ok {
		customDomainConfig := &cognitoidentityprovider.CustomDomainConfigType{
			CertificateArn: aws.String(v.(string)),
		}
		params.CustomDomainConfig = customDomainConfig
		timeout = 60 * time.Minute //Custom domains take more time to become active
	}

	log.Printf("[DEBUG] Creating Cognito User Pool Domain: %s", params)

	_, err := conn.CreateUserPoolDomain(params)
	if err != nil {
		return fmt.Errorf("Error creating Cognito User Pool Domain: %w", err)
	}

	d.SetId(domain)

	if _, err := waitUserPoolDomainCreated(conn, d.Id(), timeout); err != nil {
		return fmt.Errorf("error waiting for User Pool Domain (%s) creation: %w", d.Id(), err)
	}

	return resourceUserPoolDomainRead(d, meta)
}

func resourceUserPoolDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn
	log.Printf("[DEBUG] Reading Cognito User Pool Domain: %s", d.Id())

	domain, err := conn.DescribeUserPoolDomain(&cognitoidentityprovider.DescribeUserPoolDomainInput{
		Domain: aws.String(d.Id()),
	})
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeResourceNotFoundException) {
		names.LogNotFoundRemoveState(names.CognitoIDP, names.ErrActionReading, ResUserPoolDomain, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.Error(names.CognitoIDP, names.ErrActionReading, ResUserPoolDomain, d.Id(), err)
	}

	desc := domain.DomainDescription

	if !d.IsNewResource() && desc.Status == nil {
		names.LogNotFoundRemoveState(names.CognitoIDP, names.ErrActionReading, ResUserPoolDomain, d.Id())
		d.SetId("")
		return nil
	}

	if d.IsNewResource() && desc.Status == nil {
		return names.Error(names.CognitoIDP, names.ErrActionReading, ResUserPoolDomain, d.Id(), errors.New("not found after creation"))
	}

	d.Set("domain", d.Id())
	d.Set("certificate_arn", "")
	if desc.CustomDomainConfig != nil {
		d.Set("certificate_arn", desc.CustomDomainConfig.CertificateArn)
	}
	d.Set("aws_account_id", desc.AWSAccountId)
	d.Set("cloudfront_distribution_arn", desc.CloudFrontDistribution)
	d.Set("s3_bucket", desc.S3Bucket)
	d.Set("user_pool_id", desc.UserPoolId)
	d.Set("version", desc.Version)

	return nil
}

func resourceUserPoolDomainDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn
	log.Printf("[DEBUG] Deleting Cognito User Pool Domain: %s", d.Id())

	_, err := conn.DeleteUserPoolDomain(&cognitoidentityprovider.DeleteUserPoolDomainInput{
		Domain:     aws.String(d.Id()),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	})
	if err != nil {
		return fmt.Errorf("Error deleting User Pool Domain: %w", err)
	}

	if _, err := waitUserPoolDomainDeleted(conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeResourceNotFoundException) {
			return nil
		}
		return fmt.Errorf("error waiting for User Pool Domain (%s) deletion: %w", d.Id(), err)
	}

	return nil

}
