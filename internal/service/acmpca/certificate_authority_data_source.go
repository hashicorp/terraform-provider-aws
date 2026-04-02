// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package acmpca

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acmpca"
	"github.com/aws/aws-sdk-go-v2/service/acmpca/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_acmpca_certificate_authority", name="Certificate Authority")
// @Tags(identifierAttribute="arn")
func dataSourceCertificateAuthority() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCertificateAuthorityRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{names.AttrARN, "subject"},
			},
			"subject": {
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: []string{names.AttrARN, "subject"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"common_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 64),
						},
						"country": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 2),
						},
						"distinguished_name_qualifier": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 64),
						},
						"generation_qualifier": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 3),
						},
						"given_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 16),
						},
						"initials": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 5),
						},
						"locality": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 128),
						},
						"organization": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 64),
						},
						"organizational_unit": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 64),
						},
						"pseudonym": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 128),
						},
						names.AttrState: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 128),
						},
						"surname": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 40),
						},
						"title": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 64),
						},
					},
				},
			},
			names.AttrCertificate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCertificateChain: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_signing_request": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_storage_security_standard": {
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
									"custom_path": {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrEnabled: {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"expiration_in_days": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									names.AttrS3BucketName: {
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
									names.AttrEnabled: {
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
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrType: {
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

func dataSourceCertificateAuthorityRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAClient(ctx)
	var certificateAuthority *types.CertificateAuthority
	var certificateAuthorityARN string

	if v, ok := d.GetOk(names.AttrARN); ok {
		certificateAuthorityARN = v.(string)
		var err error
		input := acmpca.DescribeCertificateAuthorityInput{
			CertificateAuthorityArn: aws.String(certificateAuthorityARN),
		}
		certificateAuthority, err = findCertificateAuthority(ctx, conn, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading ACM PCA Certificate Authority (%s): %s", certificateAuthorityARN, err)
		}
	} else {
		subject := d.Get("subject").([]any)
		paginator := acmpca.NewListCertificateAuthoritiesPaginator(conn, &acmpca.ListCertificateAuthoritiesInput{})

		var output []types.CertificateAuthority
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "reading ACM PCA Certificate Authority: %s", err)
			}
			for _, ca := range page.CertificateAuthorities {
				desc, err := conn.DescribeCertificateAuthority(ctx, &acmpca.DescribeCertificateAuthorityInput{
					CertificateAuthorityArn: ca.Arn,
				})
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "reading ACM PCA Certificate Authority (%s): %s", aws.ToString(ca.Arn), err)
				}

				actual := desc.CertificateAuthority.CertificateAuthorityConfiguration.Subject

				match := true

				if len(subject) > 0 {
					expected := subject[0].(map[string]any)

					if v, ok := expected["common_name"]; ok && v.(string) != "" {
						if aws.ToString(actual.CommonName) != v.(string) {
							match = false
						}
					}
					if v, ok := expected["country"]; ok && v.(string) != "" {
						if aws.ToString(actual.Country) != v.(string) {
							match = false
						}
					}
					if v, ok := expected["distinguished_name_qualifier"]; ok && v.(string) != "" {
						if aws.ToString(actual.DistinguishedNameQualifier) != v.(string) {
							match = false
						}
					}
					if v, ok := expected["generation_qualifier"]; ok && v.(string) != "" {
						if aws.ToString(actual.GenerationQualifier) != v.(string) {
							match = false
						}
					}
					if v, ok := expected["given_name"]; ok && v.(string) != "" {
						if aws.ToString(actual.GivenName) != v.(string) {
							match = false
						}
					}
					if v, ok := expected["initials"]; ok && v.(string) != "" {
						if aws.ToString(actual.Initials) != v.(string) {
							match = false
						}
					}
					if v, ok := expected["locality"]; ok && v.(string) != "" {
						if aws.ToString(actual.Locality) != v.(string) {
							match = false
						}
					}
					if v, ok := expected["organization"]; ok && v.(string) != "" {
						if aws.ToString(actual.Organization) != v.(string) {
							match = false
						}
					}
					if v, ok := expected["organizational_unit"]; ok && v.(string) != "" {
						if aws.ToString(actual.OrganizationalUnit) != v.(string) {
							match = false
						}
					}
					if v, ok := expected["pseudonym"]; ok && v.(string) != "" {
						if aws.ToString(actual.Pseudonym) != v.(string) {
							match = false
						}
					}
					if v, ok := expected[names.AttrState]; ok && v.(string) != "" {
						if aws.ToString(actual.State) != v.(string) {
							match = false
						}
					}
					if v, ok := expected["surname"]; ok && v.(string) != "" {
						if aws.ToString(actual.Surname) != v.(string) {
							match = false
						}
					}
					if v, ok := expected["title"]; ok && v.(string) != "" {
						if aws.ToString(actual.Title) != v.(string) {
							match = false
						}
					}
				}

				if match {
					output = append(output, *desc.CertificateAuthority)
				}
			}
		}

		if len(output) == 0 {
			return sdkdiag.AppendErrorf(diags, "reading ACM PCA Certificate Authority: no matching ACM PCA Certificate Authority found")
		}

		if len(output) > 1 {
			return sdkdiag.AppendErrorf(diags, "reading ACM PCA Certificate Authority: multiple ACM PCA Certificate Authorities matched")
		}
		certificateAuthority = &output[0]
		certificateAuthorityARN = aws.ToString(certificateAuthority.Arn)
	}

	d.SetId(certificateAuthorityARN)
	d.Set(names.AttrARN, certificateAuthority.Arn)
	d.Set("key_storage_security_standard", certificateAuthority.KeyStorageSecurityStandard)
	d.Set("not_after", aws.ToTime(certificateAuthority.NotAfter).Format(time.RFC3339))
	d.Set("not_before", aws.ToTime(certificateAuthority.NotBefore).Format(time.RFC3339))
	if err := d.Set("revocation_configuration", flattenRevocationConfiguration(certificateAuthority.RevocationConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting revocation_configuration: %s", err)
	}
	d.Set("serial", certificateAuthority.Serial)
	d.Set(names.AttrStatus, certificateAuthority.Status)
	d.Set(names.AttrType, certificateAuthority.Type)
	d.Set("usage_mode", certificateAuthority.UsageMode)

	getCACertInput := acmpca.GetCertificateAuthorityCertificateInput{
		CertificateAuthorityArn: aws.String(certificateAuthorityARN),
	}
	outputGCACert, err := conn.GetCertificateAuthorityCertificate(ctx, &getCACertInput)

	// Returned when in PENDING_CERTIFICATE status
	// InvalidStateException: The certificate authority XXXXX is not in the correct state to have a certificate signing request.
	if err != nil && !errs.IsA[*types.InvalidStateException](err) {
		return sdkdiag.AppendErrorf(diags, "reading ACM PCA Certificate Authority (%s) Certificate: %s", d.Id(), err)
	}

	d.Set(names.AttrCertificate, "")
	d.Set(names.AttrCertificateChain, "")
	if outputGCACert != nil {
		d.Set(names.AttrCertificate, outputGCACert.Certificate)
		d.Set(names.AttrCertificateChain, outputGCACert.CertificateChain)
	}

	// Attempt to get the CSR (if permitted).
	getCACSRInput := acmpca.GetCertificateAuthorityCsrInput{
		CertificateAuthorityArn: aws.String(certificateAuthorityARN),
	}
	outputGCACsr, err := conn.GetCertificateAuthorityCsr(ctx, &getCACSRInput)

	switch {
	case tfawserr.ErrCodeEquals(err, "AccessDeniedException"):
		// Handle permission issues gracefully for Resource Access Manager shared CAs.
		// arn:aws:ram::aws:permission/AWSRAMDefaultPermissionCertificateAuthority does not include acm-pca:GetCertificateAuthorityCsr.
	case errs.IsA[*types.InvalidStateException](err):
		// Returned when in PENDING_CERTIFICATE status
		// InvalidStateException: The certificate authority XXXXX is not in the correct state to have a certificate signing request.
	case err != nil:
		return sdkdiag.AppendErrorf(diags, "reading ACM PCA Certificate Authority (%s) Certificate Signing Request: %s", d.Id(), err)
	}

	d.Set("certificate_signing_request", "")
	if outputGCACsr != nil {
		d.Set("certificate_signing_request", outputGCACsr.Csr)
	}

	return diags
}
