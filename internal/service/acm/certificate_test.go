package acm_test

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)





func TestAccACMCertificate_emailValidation(t *testing.T) {
	resourceName := "aws_acm_certificate.cert"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfig(domain, acm.ValidationMethodEmail),
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "0"),
					resource.TestMatchResourceAttr(resourceName, "validation_emails.0", regexp.MustCompile(`^[^@]+@.+$`)),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodEmail),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccACMCertificate_dnsValidation(t *testing.T) {
	resourceName := "aws_acm_certificate.cert"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfig(domain, acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          domain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccACMCertificate_root(t *testing.T) {
	resourceName := "aws_acm_certificate.cert"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfig(rootDomain, acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", rootDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccACMCertificate_privateCert(t *testing.T) {
	certificateAuthorityResourceName := "aws_acmpca_certificate_authority.test"
	resourceName := "aws_acm_certificate.cert"

	commonName := acctest.RandomDomain()
	certificateDomainName := commonName.RandomSubdomain().String()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfig_privateCert(commonName.String(), certificateDomainName),
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", certificateDomainName),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusFailed), // FailureReason: PCA_INVALID_STATE (PCA State: PENDING_CERTIFICATE)
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", "NONE"),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_authority_arn", certificateAuthorityResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccACMCertificate_Root_trailingPeriod updated in 3.0 to account for domain_name plan-time validation
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13510
func TestAccACMCertificate_Root_trailingPeriod(t *testing.T) {
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := fmt.Sprintf("%s.", rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAcmCertificateConfig(domain, acm.ValidationMethodDns),
				ExpectError: regexp.MustCompile(`invalid value for domain_name \(cannot end with a period\)`),
			},
		},
	})
}

func TestAccACMCertificate_rootAndWildcardSan(t *testing.T) {
	resourceName := "aws_acm_certificate.cert"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	wildcardDomain := fmt.Sprintf("*.%s", rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfig_subjectAlternativeNames(rootDomain, strconv.Quote(wildcardDomain), acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", rootDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          wildcardDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", wildcardDomain),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccACMCertificate_SubjectAlternativeNames_emptyString(t *testing.T) {
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAcmCertificateConfig_subjectAlternativeNames(domain, strconv.Quote(""), acm.ValidationMethodDns),
				ExpectError: regexp.MustCompile(`expected length`),
			},
		},
	})
}

func TestAccACMCertificate_San_single(t *testing.T) {
	resourceName := "aws_acm_certificate.cert"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	sanDomain := acctest.ACMCertificateRandomSubDomain(rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfig_subjectAlternativeNames(domain, strconv.Quote(sanDomain), acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          domain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          sanDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", sanDomain),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccACMCertificate_San_multiple(t *testing.T) {
	resourceName := "aws_acm_certificate.cert"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	sanDomain1 := acctest.ACMCertificateRandomSubDomain(rootDomain)
	sanDomain2 := acctest.ACMCertificateRandomSubDomain(rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfig_subjectAlternativeNames(domain, fmt.Sprintf("%q, %q", sanDomain1, sanDomain2), acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          domain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          sanDomain1,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          sanDomain2,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", sanDomain1),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", sanDomain2),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccACMCertificate_San_trailingPeriod(t *testing.T) {
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	sanDomain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	resourceName := "aws_acm_certificate.cert"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfig_subjectAlternativeNames(domain, strconv.Quote(sanDomain), acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile(`certificate/.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          domain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          strings.TrimSuffix(sanDomain, "."),
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", strings.TrimSuffix(sanDomain, ".")),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccACMCertificate_wildcard(t *testing.T) {
	resourceName := "aws_acm_certificate.cert"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	wildcardDomain := fmt.Sprintf("*.%s", rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfig(wildcardDomain, acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", wildcardDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          wildcardDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccACMCertificate_wildcardAndRootSan(t *testing.T) {
	resourceName := "aws_acm_certificate.cert"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	wildcardDomain := fmt.Sprintf("*.%s", rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfig_subjectAlternativeNames(wildcardDomain, strconv.Quote(rootDomain), acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", wildcardDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          wildcardDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", rootDomain),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccACMCertificate_disableCTLogging(t *testing.T) {
	resourceName := "aws_acm_certificate.cert"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfig_disableCTLogging(rootDomain, acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", rootDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.certificate_transparency_logging_preference", acm.CertificateTransparencyLoggingPreferenceDisabled),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccACMCertificate_tags(t *testing.T) {
	resourceName := "aws_acm_certificate.cert"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfig(domain, acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAcmCertificateConfig_twoTags(domain, acm.ValidationMethodDns, "Hello", "World", "Foo", "Bar"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Hello", "World"),
					resource.TestCheckResourceAttr(resourceName, "tags.Foo", "Bar"),
				),
			},
			{
				Config: testAccAcmCertificateConfig_twoTags(domain, acm.ValidationMethodDns, "Hello", "World", "Foo", "Baz"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Hello", "World"),
					resource.TestCheckResourceAttr(resourceName, "tags.Foo", "Baz"),
				),
			},
			{
				Config: testAccAcmCertificateConfig_oneTag(domain, acm.ValidationMethodDns, "Environment", "Test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "Test"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

//lintignore:AT002
func TestAccACMCertificate_Imported_domainName(t *testing.T) {
	resourceName := "aws_acm_certificate.test"

	commonName := "example.com"
	caKey := acctest.TLSRSAPrivateKeyPEM(2048)
	caCertificate := acctest.TLSRSAX509SelfSignedCACertificatePEM(caKey)
	key := acctest.TLSRSAPrivateKeyPEM(2048)
	certificate := acctest.TLSRSAX509LocallySignedCertificatePEM(caKey, caCertificate, key, commonName)

	newCaKey := acctest.TLSRSAPrivateKeyPEM(2048)
	newCaCertificate := acctest.TLSRSAX509SelfSignedCACertificatePEM(newCaKey)
	newCertificate := acctest.TLSRSAX509LocallySignedCertificatePEM(newCaKey, newCaCertificate, key, commonName)

	withoutChainDomain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfigPrivateKey(certificate, key, caCertificate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "domain_name", commonName),
				),
			},
			{
				Config: testAccAcmCertificateConfigPrivateKey(newCertificate, key, newCaCertificate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "domain_name", commonName),
				),
			},
			{
				Config: testAccAcmCertificateConfigPrivateKeyWithoutChain(withoutChainDomain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "domain_name", withoutChainDomain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// These are not returned by the API
				ImportStateVerifyIgnore: []string{"private_key", "certificate_body", "certificate_chain"},
			},
		},
	})
}

//lintignore:AT002
func TestAccACMCertificate_Imported_ipAddress(t *testing.T) { // Reference: https://github.com/hashicorp/terraform-provider-aws/issues/7103
	resourceName := "aws_acm_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfigPrivateKeyWithoutChain("1.2.3.4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "domain_name", ""),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// These are not returned by the API
				ImportStateVerifyIgnore: []string{"private_key", "certificate_body"},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/15055
func TestAccACMCertificate_PrivateKey_tags(t *testing.T) {
	resourceName := "aws_acm_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfigPrivateKeyTags("1.2.3.4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_key", "certificate_body"},
			},
			{
				Config: testAccAcmCertificateConfigPrivateKeyTags("5.6.7.8"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
		},
	})
}

func testAccAcmCertificateConfig(domainName, validationMethod string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "cert" {
  domain_name       = "%s"
  validation_method = "%s"
}
`, domainName, validationMethod)
}

func testAccAcmCertificateConfig_privateCert(commonName, certificateDomainName string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}

resource "aws_acm_certificate" "cert" {
  domain_name               = %[2]q
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn
}
`, commonName, certificateDomainName)
}

func testAccAcmCertificateConfig_subjectAlternativeNames(domainName, subjectAlternativeNames, validationMethod string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "cert" {
  domain_name               = "%s"
  subject_alternative_names = [%s]
  validation_method         = "%s"
}
`, domainName, subjectAlternativeNames, validationMethod)
}

func testAccAcmCertificateConfig_oneTag(domainName, validationMethod, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "cert" {
  domain_name       = "%s"
  validation_method = "%s"

  tags = {
    "%s" = "%s"
  }
}
`, domainName, validationMethod, tag1Key, tag1Value)
}

func testAccAcmCertificateConfig_twoTags(domainName, validationMethod, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "cert" {
  domain_name       = "%s"
  validation_method = "%s"

  tags = {
    "%s" = "%s"
    "%s" = "%s"
  }
}
`, domainName, validationMethod, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccAcmCertificateConfigPrivateKeyWithoutChain(commonName string) string {
	key := acctest.TLSRSAPrivateKeyPEM(2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(key, commonName)

	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[1]s"
  private_key      = "%[2]s"
}
`, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
}

func testAccAcmCertificateConfigPrivateKeyTags(commonName string) string {
	key := acctest.TLSRSAPrivateKeyPEM(2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(key, commonName)

	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[1]s"
  private_key      = "%[2]s"

  tags = {
    key1 = "value1"
  }
}
`, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
}

func testAccAcmCertificateConfigPrivateKey(certificate, privateKey, chain string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body  = "%[1]s"
  private_key       = "%[2]s"
  certificate_chain = "%[3]s"
}
`, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(privateKey), acctest.TLSPEMEscapeNewlines(chain))
}

func testAccAcmCertificateConfig_disableCTLogging(domainName, validationMethod string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "cert" {
  domain_name       = "%s"
  validation_method = "%s"
  options {
    certificate_transparency_logging_preference = "DISABLED"
  }
}
`, domainName, validationMethod)
}

func testAccCheckAcmCertificateDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ACMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_acm_certificate" {
			continue
		}
		_, err := conn.DescribeCertificate(&acm.DescribeCertificateInput{
			CertificateArn: aws.String(rs.Primary.ID),
		})

		if err == nil {
			return fmt.Errorf("Certificate still exists.")
		}

		// Verify the error is what we want
		if !tfawserr.ErrMessageContains(err, acm.ErrCodeResourceNotFoundException, "") {
			return err
		}
	}

	return nil
}
