// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cognito_user_pool_domain", name="User Pool Domain")
func resourceUserPoolDomain() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserPoolDomainCreate,
		ReadWithoutTimeout:   resourceUserPoolDomainRead,
		UpdateWithoutTimeout: resourceUserPoolDomainUpdate,
		DeleteWithoutTimeout: resourceUserPoolDomainDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrAWSAccountID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCertificateARN: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"cloudfront_distribution": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloudfront_distribution_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloudfront_distribution_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDomain: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 63),
			},
			names.AttrS3Bucket: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrUserPoolID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: customdiff.ForceNewIfChange(names.AttrCertificateARN, func(_ context.Context, old, new, meta interface{}) bool {
			// If the cert arn is being changed to a new arn, don't force new.
			return !(old.(string) != "" && new.(string) != "")
		}),
	}
}

func resourceUserPoolDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	domain := d.Get(names.AttrDomain).(string)
	timeout := 1 * time.Minute
	input := &cognitoidentityprovider.CreateUserPoolDomainInput{
		Domain:     aws.String(domain),
		UserPoolId: aws.String(d.Get(names.AttrUserPoolID).(string)),
	}

	if v, ok := d.GetOk(names.AttrCertificateARN); ok {
		input.CustomDomainConfig = &awstypes.CustomDomainConfigType{
			CertificateArn: aws.String(v.(string)),
		}
		timeout = 60 * time.Minute // Custom domains take more time to become active.
	}

	_, err := conn.CreateUserPoolDomain(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Cognito User Pool Domain (%s): %s", domain, err)
	}

	d.SetId(domain)

	if _, err := waitUserPoolDomainCreated(ctx, conn, d.Id(), timeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Cognito User Pool Domain (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceUserPoolDomainRead(ctx, d, meta)...)
}

func resourceUserPoolDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	desc, err := findUserPoolDomain(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Cognito User Pool Domain %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Cognito User Pool Domain (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAWSAccountID, desc.AWSAccountId)
	d.Set(names.AttrCertificateARN, "")
	if desc.CustomDomainConfig != nil {
		d.Set(names.AttrCertificateARN, desc.CustomDomainConfig.CertificateArn)
	}
	d.Set("cloudfront_distribution", desc.CloudFrontDistribution)
	d.Set("cloudfront_distribution_arn", desc.CloudFrontDistribution)
	d.Set("cloudfront_distribution_zone_id", meta.(*conns.AWSClient).CloudFrontDistributionHostedZoneID(ctx))
	d.Set(names.AttrDomain, d.Id())
	d.Set(names.AttrS3Bucket, desc.S3Bucket)
	d.Set(names.AttrUserPoolID, desc.UserPoolId)
	d.Set(names.AttrVersion, desc.Version)

	return diags
}

func resourceUserPoolDomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	input := &cognitoidentityprovider.UpdateUserPoolDomainInput{
		CustomDomainConfig: &awstypes.CustomDomainConfigType{
			CertificateArn: aws.String(d.Get(names.AttrCertificateARN).(string)),
		},
		Domain:     aws.String(d.Id()),
		UserPoolId: aws.String(d.Get(names.AttrUserPoolID).(string)),
	}

	_, err := conn.UpdateUserPoolDomain(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Cognito User Pool Domain (%s): %s", d.Id(), err)
	}

	const (
		timeout = 60 * time.Minute // Update is only for cert arns on custom domains, which take more time to become active.
	)
	if _, err := waitUserPoolDomainUpdated(ctx, conn, d.Id(), timeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Cognito User Pool Domain (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceUserPoolDomainRead(ctx, d, meta)...)
}

func resourceUserPoolDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	log.Printf("[DEBUG] Deleting Cognito User Pool Domain: %s", d.Id())
	_, err := conn.DeleteUserPoolDomain(ctx, &cognitoidentityprovider.DeleteUserPoolDomainInput{
		Domain:     aws.String(d.Id()),
		UserPoolId: aws.String(d.Get(names.AttrUserPoolID).(string)),
	})

	if errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "No such domain") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Cognito User Pool Domain (%s): %s", d.Id(), err)
	}

	const (
		timeout = 1 * time.Minute
	)
	if _, err := waitUserPoolDomainDeleted(ctx, conn, d.Id(), timeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Cognito User Pool Domain (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findUserPoolDomain(ctx context.Context, conn *cognitoidentityprovider.Client, domain string) (*awstypes.DomainDescriptionType, error) {
	input := &cognitoidentityprovider.DescribeUserPoolDomainInput{
		Domain: aws.String(domain),
	}

	output, err := conn.DescribeUserPoolDomain(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	// e.g.
	// {
	// 	"DomainDescription": {}
	// }
	if output == nil || output.DomainDescription == nil || output.DomainDescription.Status == "" {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DomainDescription, nil
}

func statusUserPoolDomain(ctx context.Context, conn *cognitoidentityprovider.Client, domain string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findUserPoolDomain(ctx, conn, domain)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitUserPoolDomainCreated(ctx context.Context, conn *cognitoidentityprovider.Client, domain string, timeout time.Duration) (*awstypes.DomainDescriptionType, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DomainStatusTypeCreating, awstypes.DomainStatusTypeUpdating),
		Target:  enum.Slice(awstypes.DomainStatusTypeActive),
		Refresh: statusUserPoolDomain(ctx, conn, domain),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DomainDescriptionType); ok {
		return output, err
	}

	return nil, err
}

func waitUserPoolDomainUpdated(ctx context.Context, conn *cognitoidentityprovider.Client, domain string, timeout time.Duration) (*awstypes.DomainDescriptionType, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DomainStatusTypeUpdating),
		Target:  enum.Slice(awstypes.DomainStatusTypeActive),
		Refresh: statusUserPoolDomain(ctx, conn, domain),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DomainDescriptionType); ok {
		return output, err
	}

	return nil, err
}

func waitUserPoolDomainDeleted(ctx context.Context, conn *cognitoidentityprovider.Client, domain string, timeout time.Duration) (*awstypes.DomainDescriptionType, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DomainStatusTypeUpdating, awstypes.DomainStatusTypeDeleting),
		Target:  []string{},
		Refresh: statusUserPoolDomain(ctx, conn, domain),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DomainDescriptionType); ok {
		return output, err
	}

	return nil, err
}
