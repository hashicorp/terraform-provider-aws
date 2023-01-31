package acmpca

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceCertificateAuthority() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCertificateAuthorityRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"certificate": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_chain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_signing_request": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"not_after": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"not_before": {
				Type:     schema.TypeString,
				Computed: true,
			},
			// https://docs.aws.amazon.com/privateca/latest/APIReference/API_RevocationConfiguration.html
			"revocation_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// https://docs.aws.amazon.com/privateca/latest/APIReference/API_CrlConfiguration.html
						"crl_configuration": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"custom_cname": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"enabled": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"expiration_in_days": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"s3_bucket_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"s3_object_acl": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						// https://docs.aws.amazon.com/privateca/latest/APIReference/API_OcspConfiguration.html
						"ocsp_configuration": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"ocsp_custom_cname": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"serial": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"usage_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceCertificateAuthorityRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAConn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	certificateAuthorityARN := d.Get("arn").(string)

	describeCertificateAuthorityInput := &acmpca.DescribeCertificateAuthorityInput{
		CertificateAuthorityArn: aws.String(certificateAuthorityARN),
	}

	log.Printf("[DEBUG] Reading ACM PCA Certificate Authority: %s", describeCertificateAuthorityInput)

	describeCertificateAuthorityOutput, err := conn.DescribeCertificateAuthorityWithContext(ctx, describeCertificateAuthorityInput)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ACM PCA Certificate Authority (%s): %s", certificateAuthorityARN, err)
	}

	if describeCertificateAuthorityOutput.CertificateAuthority == nil {
		return sdkdiag.AppendErrorf(diags, "reading ACM PCA Certificate Authority: not found")
	}
	certificateAuthority := describeCertificateAuthorityOutput.CertificateAuthority

	d.Set("arn", certificateAuthority.Arn)
	d.Set("not_after", aws.TimeValue(certificateAuthority.NotAfter).Format(time.RFC3339))
	d.Set("not_before", aws.TimeValue(certificateAuthority.NotBefore).Format(time.RFC3339))
	if err := d.Set("revocation_configuration", flattenRevocationConfiguration(certificateAuthority.RevocationConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting revocation_configuration: %s", err)
	}
	d.Set("serial", certificateAuthority.Serial)
	d.Set("status", certificateAuthority.Status)
	d.Set("type", certificateAuthority.Type)
	d.Set("usage_mode", certificateAuthority.UsageMode)

	getCertificateAuthorityCertificateInput := &acmpca.GetCertificateAuthorityCertificateInput{
		CertificateAuthorityArn: aws.String(certificateAuthorityARN),
	}

	log.Printf("[DEBUG] Reading ACM PCA Certificate Authority Certificate: %s", getCertificateAuthorityCertificateInput)

	getCertificateAuthorityCertificateOutput, err := conn.GetCertificateAuthorityCertificateWithContext(ctx, getCertificateAuthorityCertificateInput)
	if err != nil {
		// Returned when in PENDING_CERTIFICATE status
		// InvalidStateException: The certificate authority XXXXX is not in the correct state to have a certificate signing request.
		if !tfawserr.ErrCodeEquals(err, acmpca.ErrCodeInvalidStateException) {
			return sdkdiag.AppendErrorf(diags, "reading ACM PCA Certificate Authority Certificate: %s", err)
		}
	}

	d.Set("certificate", "")
	d.Set("certificate_chain", "")
	if getCertificateAuthorityCertificateOutput != nil {
		d.Set("certificate", getCertificateAuthorityCertificateOutput.Certificate)
		d.Set("certificate_chain", getCertificateAuthorityCertificateOutput.CertificateChain)
	}

	getCertificateAuthorityCsrInput := &acmpca.GetCertificateAuthorityCsrInput{
		CertificateAuthorityArn: aws.String(certificateAuthorityARN),
	}

	log.Printf("[DEBUG] Reading ACM PCA Certificate Authority Certificate Signing Request: %s", getCertificateAuthorityCsrInput)

	getCertificateAuthorityCsrOutput, err := conn.GetCertificateAuthorityCsrWithContext(ctx, getCertificateAuthorityCsrInput)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ACM PCA Certificate Authority Certificate Signing Request: %s", err)
	}

	d.Set("certificate_signing_request", "")
	if getCertificateAuthorityCsrOutput != nil {
		d.Set("certificate_signing_request", getCertificateAuthorityCsrOutput.Csr)
	}

	tags, err := ListTags(ctx, conn, certificateAuthorityARN)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for ACM PCA Certificate Authority (%s): %s", certificateAuthorityARN, err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	d.SetId(certificateAuthorityARN)

	return diags
}
