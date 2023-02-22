package acm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceCertificateValidation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCertificateValidationCreate,
		ReadWithoutTimeout:   resourceCertificateValidationRead,
		DeleteWithoutTimeout: schema.NoopContext,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(75 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"certificate_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"validation_record_fqdns": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceCertificateValidationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ACMConn()

	arn := d.Get("certificate_arn").(string)
	certificate, err := FindCertificateByARN(ctx, conn, arn)

	if err != nil {
		return diag.Errorf("reading ACM Certificate (%s): %s", arn, err)
	}

	if v := aws.StringValue(certificate.Type); v != acm.CertificateTypeAmazonIssued {
		return diag.Errorf("ACM Certificate (%s) has type %s, no validation necessary", arn, v)
	}

	if v, ok := d.GetOk("validation_record_fqdns"); ok && v.(*schema.Set).Len() > 0 {
		fqdns := make(map[string]*acm.DomainValidation)

		for _, domainValidation := range certificate.DomainValidationOptions {
			if v := aws.StringValue(domainValidation.ValidationMethod); v != acm.ValidationMethodDns {
				return diag.Errorf("validation_record_fqdns is not valid for %s validation", v)
			}

			if v := domainValidation.ResourceRecord; v != nil {
				if v := aws.StringValue(v.Name); v != "" {
					fqdns[strings.TrimSuffix(v, ".")] = domainValidation
				}
			}
		}

		for _, v := range v.(*schema.Set).List() {
			delete(fqdns, strings.TrimSuffix(v.(string), "."))
		}

		if len(fqdns) > 0 {
			var errs *multierror.Error

			for fqdn, domainValidation := range fqdns {
				errs = multierror.Append(errs, fmt.Errorf("missing %s DNS validation record: %s", aws.StringValue(domainValidation.DomainName), fqdn))
			}

			return diag.FromErr(errs)
		}
	}

	if _, err := waitCertificateIssued(ctx, conn, arn, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for ACM Certificate (%s) to be issued: %s", arn, err)
	}

	d.SetId(aws.TimeValue(certificate.IssuedAt).String())

	return resourceCertificateValidationRead(ctx, d, meta)
}

func resourceCertificateValidationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ACMConn()

	arn := d.Get("certificate_arn").(string)
	certificate, err := FindCertificateValidationByARN(ctx, conn, arn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ACM Certificate %s not found, removing from state", arn)
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading ACM Certificate (%s): %s", arn, err)
	}

	d.Set("certificate_arn", certificate.CertificateArn)

	return nil
}

func FindCertificateValidationByARN(ctx context.Context, conn *acm.ACM, arn string) (*acm.CertificateDetail, error) {
	output, err := FindCertificateByARN(ctx, conn, arn)

	if err != nil {
		return nil, err
	}

	if status := aws.StringValue(output.Status); status != acm.CertificateStatusIssued {
		return nil, &resource.NotFoundError{
			Message:     status,
			LastRequest: arn,
		}
	}

	return output, nil
}

func statusCertificate(ctx context.Context, conn *acm.ACM, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Don't call FindCertificateByARN as it maps useful status codes to NotFoundError.
		input := &acm.DescribeCertificateInput{
			CertificateArn: aws.String(arn),
		}

		output, err := findCertificate(ctx, conn, input)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitCertificateIssued(ctx context.Context, conn *acm.ACM, arn string, timeout time.Duration) (*acm.CertificateDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{acm.CertificateStatusPendingValidation},
		Target:  []string{acm.CertificateStatusIssued},
		Refresh: statusCertificate(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*acm.CertificateDetail); ok {
		switch aws.StringValue(output.Status) {
		case acm.CertificateStatusFailed:
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureReason)))
		case acm.CertificateStatusRevoked:
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.RevocationReason)))
		}

		return output, err
	}

	return nil, err
}
