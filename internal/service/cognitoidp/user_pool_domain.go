package cognitoidp

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceUserPoolDomain() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserPoolDomainCreate,
		ReadWithoutTimeout:   resourceUserPoolDomainRead,
		DeleteWithoutTimeout: resourceUserPoolDomainDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourceUserPoolDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPConn()

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

	_, err := conn.CreateUserPoolDomainWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error creating Cognito User Pool Domain: %s", err)
	}

	d.SetId(domain)

	if _, err := waitUserPoolDomainCreated(ctx, conn, d.Id(), timeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for User Pool Domain (%s) creation: %s", d.Id(), err)
	}

	return append(diags, resourceUserPoolDomainRead(ctx, d, meta)...)
}

func resourceUserPoolDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPConn()
	log.Printf("[DEBUG] Reading Cognito User Pool Domain: %s", d.Id())

	domain, err := conn.DescribeUserPoolDomainWithContext(ctx, &cognitoidentityprovider.DescribeUserPoolDomainInput{
		Domain: aws.String(d.Id()),
	})
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeResourceNotFoundException) {
		create.LogNotFoundRemoveState(names.CognitoIDP, create.ErrActionReading, ResNameUserPoolDomain, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.DiagError(names.CognitoIDP, create.ErrActionReading, ResNameUserPoolDomain, d.Id(), err)
	}

	desc := domain.DomainDescription

	if !d.IsNewResource() && desc.Status == nil {
		create.LogNotFoundRemoveState(names.CognitoIDP, create.ErrActionReading, ResNameUserPoolDomain, d.Id())
		d.SetId("")
		return diags
	}

	if d.IsNewResource() && desc.Status == nil {
		return create.DiagError(names.CognitoIDP, create.ErrActionReading, ResNameUserPoolDomain, d.Id(), errors.New("not found after creation"))
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

	return diags
}

func resourceUserPoolDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPConn()
	log.Printf("[DEBUG] Deleting Cognito User Pool Domain: %s", d.Id())

	_, err := conn.DeleteUserPoolDomainWithContext(ctx, &cognitoidentityprovider.DeleteUserPoolDomainInput{
		Domain:     aws.String(d.Id()),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error deleting User Pool Domain: %s", err)
	}

	if _, err := waitUserPoolDomainDeleted(ctx, conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeResourceNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "waiting for User Pool Domain (%s) deletion: %s", d.Id(), err)
	}

	return diags
}
