package ses

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceDomainIdentityVerification() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainIdentityVerificationCreate,
		ReadWithoutTimeout:   resourceDomainIdentityVerificationRead,
		DeleteWithoutTimeout: resourceDomainIdentityVerificationDelete,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringDoesNotMatch(regexp.MustCompile(`\.$`), "cannot end with a period"),
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(45 * time.Minute),
		},
	}
}

func getIdentityVerificationAttributes(ctx context.Context, conn *ses.SES, domainName string) (*ses.IdentityVerificationAttributes, error) {
	input := &ses.GetIdentityVerificationAttributesInput{
		Identities: []*string{
			aws.String(domainName),
		},
	}

	response, err := conn.GetIdentityVerificationAttributesWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("Error getting identity verification attributes: %s", err)
	}

	return response.VerificationAttributes[domainName], nil
}

func resourceDomainIdentityVerificationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn()
	domainName := d.Get("domain").(string)
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		att, err := getIdentityVerificationAttributes(ctx, conn, domainName)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error getting identity verification attributes: %s", err))
		}

		if att == nil {
			return resource.NonRetryableError(fmt.Errorf("SES Domain Identity %s not found in AWS", domainName))
		}

		if aws.StringValue(att.VerificationStatus) != ses.VerificationStatusSuccess {
			return resource.RetryableError(fmt.Errorf("Expected domain verification Success, but was in state %s", aws.StringValue(att.VerificationStatus)))
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		var att *ses.IdentityVerificationAttributes
		att, err = getIdentityVerificationAttributes(ctx, conn, domainName)

		if att != nil && aws.StringValue(att.VerificationStatus) != ses.VerificationStatusSuccess {
			return sdkdiag.AppendErrorf(diags, "Expected domain verification Success, but was in state %s", aws.StringValue(att.VerificationStatus))
		}
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SES domain identity verification: %s", err)
	}

	log.Printf("[INFO] Domain verification successful for %s", domainName)
	d.SetId(domainName)
	return append(diags, resourceDomainIdentityVerificationRead(ctx, d, meta)...)
}

func resourceDomainIdentityVerificationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn()

	domainName := d.Id()
	d.Set("domain", domainName)

	att, err := getIdentityVerificationAttributes(ctx, conn, domainName)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Domain Identity Verification (%s): %s", domainName, err)
	}

	if att == nil {
		log.Printf("[WARN] SES Domain Identity Verification (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if aws.StringValue(att.VerificationStatus) != ses.VerificationStatusSuccess {
		log.Printf("[WARN] Expected domain verification Success, but was %s, tainting verification", aws.StringValue(att.VerificationStatus))
		d.SetId("")
		return diags
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("identity/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	return diags
}

func resourceDomainIdentityVerificationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var
	// No need to do anything, domain identity will be deleted when aws_ses_domain_identity is deleted
	diags diag.Diagnostics

	return diags
}
