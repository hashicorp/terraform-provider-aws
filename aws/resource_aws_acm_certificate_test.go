package aws

import (
	"fmt"
	"strconv"
	"testing"

	"os"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

var certificateArnRegex = regexp.MustCompile(`^arn:aws:acm:[^:]+:[^:]+:certificate/.+$`)

func testAccAwsAcmCertificateDomainFromEnv(t *testing.T) string {
	if os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN") == "" {
		t.Skip(
			"Environment variable ACM_CERTIFICATE_ROOT_DOMAIN is not set. " +
				"For DNS validation requests, this domain must be publicly " +
				"accessible and configurable via Route53 during the testing. " +
				"For email validation requests, you must have access to one of " +
				"the five standard email addresses used (admin|administrator|" +
				"hostmaster|postmaster|webmaster)@domain or one of the WHOIS " +
				"contact addresses.")
	}
	return os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN")
}

func TestAccAWSAcmCertificate_emailValidation(t *testing.T) {
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)

	rInt1 := acctest.RandInt()

	domain := fmt.Sprintf("tf-acc-%d.%s", rInt1, rootDomain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAcmCertificateConfig(domain, acm.ValidationMethodEmail),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_acm_certificate.cert", "arn", certificateArnRegex),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_name", domain),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.#", "0"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "subject_alternative_names.#", "0"),
					resource.TestMatchResourceAttr("aws_acm_certificate.cert", "validation_emails.0", regexp.MustCompile(`^[^@]+@.+$`)),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "validation_method", acm.ValidationMethodEmail),
				),
			},
			resource.TestStep{
				ResourceName:      "aws_acm_certificate.cert",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})

}

func TestAccAWSAcmCertificate_dnsValidation(t *testing.T) {
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)

	rInt1 := acctest.RandInt()

	domain := fmt.Sprintf("tf-acc-%d.%s", rInt1, rootDomain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAcmCertificateConfig(domain, acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_acm_certificate.cert", "arn", certificateArnRegex),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_name", domain),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.#", "1"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.0.domain_name", domain),
					resource.TestCheckResourceAttrSet("aws_acm_certificate.cert", "domain_validation_options.0.resource_record_name"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.0.resource_record_type", "CNAME"),
					resource.TestCheckResourceAttrSet("aws_acm_certificate.cert", "domain_validation_options.0.resource_record_value"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "subject_alternative_names.#", "0"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "validation_emails.#", "0"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "validation_method", acm.ValidationMethodDns),
				),
			},
			resource.TestStep{
				ResourceName:      "aws_acm_certificate.cert",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAcmCertificate_root(t *testing.T) {
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAcmCertificateConfig(rootDomain, acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_acm_certificate.cert", "arn", certificateArnRegex),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_name", rootDomain),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.#", "1"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.0.domain_name", rootDomain),
					resource.TestCheckResourceAttrSet("aws_acm_certificate.cert", "domain_validation_options.0.resource_record_name"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.0.resource_record_type", "CNAME"),
					resource.TestCheckResourceAttrSet("aws_acm_certificate.cert", "domain_validation_options.0.resource_record_value"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "subject_alternative_names.#", "0"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "validation_emails.#", "0"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "validation_method", acm.ValidationMethodDns),
				),
			},
			resource.TestStep{
				ResourceName:      "aws_acm_certificate.cert",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAcmCertificate_rootAndWildcardSan(t *testing.T) {
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)
	wildcardDomain := fmt.Sprintf("*.%s", rootDomain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAcmCertificateConfig_subjectAlternativeNames(rootDomain, strconv.Quote(wildcardDomain), acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_acm_certificate.cert", "arn", certificateArnRegex),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_name", rootDomain),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.#", "2"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.0.domain_name", rootDomain),
					resource.TestCheckResourceAttrSet("aws_acm_certificate.cert", "domain_validation_options.0.resource_record_name"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.0.resource_record_type", "CNAME"),
					resource.TestCheckResourceAttrSet("aws_acm_certificate.cert", "domain_validation_options.0.resource_record_value"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.1.domain_name", wildcardDomain),
					resource.TestCheckResourceAttrSet("aws_acm_certificate.cert", "domain_validation_options.1.resource_record_name"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.1.resource_record_type", "CNAME"),
					resource.TestCheckResourceAttrSet("aws_acm_certificate.cert", "domain_validation_options.1.resource_record_value"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "subject_alternative_names.#", "1"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "subject_alternative_names.0", wildcardDomain),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "validation_emails.#", "0"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "validation_method", acm.ValidationMethodDns),
				),
			},
			resource.TestStep{
				ResourceName:      "aws_acm_certificate.cert",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAcmCertificate_san_single(t *testing.T) {
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)

	rInt1 := acctest.RandInt()

	domain := fmt.Sprintf("tf-acc-%d.%s", rInt1, rootDomain)
	sanDomain := fmt.Sprintf("tf-acc-%d-san.%s", rInt1, rootDomain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAcmCertificateConfig_subjectAlternativeNames(domain, strconv.Quote(sanDomain), acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_acm_certificate.cert", "arn", certificateArnRegex),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_name", domain),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.#", "2"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.0.domain_name", domain),
					resource.TestCheckResourceAttrSet("aws_acm_certificate.cert", "domain_validation_options.0.resource_record_name"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.0.resource_record_type", "CNAME"),
					resource.TestCheckResourceAttrSet("aws_acm_certificate.cert", "domain_validation_options.0.resource_record_value"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.1.domain_name", sanDomain),
					resource.TestCheckResourceAttrSet("aws_acm_certificate.cert", "domain_validation_options.1.resource_record_name"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.1.resource_record_type", "CNAME"),
					resource.TestCheckResourceAttrSet("aws_acm_certificate.cert", "domain_validation_options.1.resource_record_value"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "subject_alternative_names.#", "1"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "subject_alternative_names.0", sanDomain),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "validation_emails.#", "0"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "validation_method", acm.ValidationMethodDns),
				),
			},
			resource.TestStep{
				ResourceName:      "aws_acm_certificate.cert",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAcmCertificate_san_multiple(t *testing.T) {
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)

	rInt1 := acctest.RandInt()

	domain := fmt.Sprintf("tf-acc-%d.%s", rInt1, rootDomain)
	sanDomain1 := fmt.Sprintf("tf-acc-%d-san1.%s", rInt1, rootDomain)
	sanDomain2 := fmt.Sprintf("tf-acc-%d-san2.%s", rInt1, rootDomain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAcmCertificateConfig_subjectAlternativeNames(domain, fmt.Sprintf("%q, %q", sanDomain1, sanDomain2), acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_acm_certificate.cert", "arn", certificateArnRegex),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_name", domain),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.#", "3"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.0.domain_name", domain),
					resource.TestCheckResourceAttrSet("aws_acm_certificate.cert", "domain_validation_options.0.resource_record_name"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.0.resource_record_type", "CNAME"),
					resource.TestCheckResourceAttrSet("aws_acm_certificate.cert", "domain_validation_options.0.resource_record_value"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.1.domain_name", sanDomain1),
					resource.TestCheckResourceAttrSet("aws_acm_certificate.cert", "domain_validation_options.1.resource_record_name"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.1.resource_record_type", "CNAME"),
					resource.TestCheckResourceAttrSet("aws_acm_certificate.cert", "domain_validation_options.1.resource_record_value"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.2.domain_name", sanDomain2),
					resource.TestCheckResourceAttrSet("aws_acm_certificate.cert", "domain_validation_options.2.resource_record_name"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.2.resource_record_type", "CNAME"),
					resource.TestCheckResourceAttrSet("aws_acm_certificate.cert", "domain_validation_options.2.resource_record_value"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "subject_alternative_names.#", "2"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "subject_alternative_names.0", sanDomain1),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "subject_alternative_names.1", sanDomain2),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "validation_emails.#", "0"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "validation_method", acm.ValidationMethodDns),
				),
			},
			resource.TestStep{
				ResourceName:      "aws_acm_certificate.cert",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAcmCertificate_wildcard(t *testing.T) {
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)
	wildcardDomain := fmt.Sprintf("*.%s", rootDomain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAcmCertificateConfig(wildcardDomain, acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_acm_certificate.cert", "arn", certificateArnRegex),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_name", wildcardDomain),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.#", "1"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.0.domain_name", wildcardDomain),
					resource.TestCheckResourceAttrSet("aws_acm_certificate.cert", "domain_validation_options.0.resource_record_name"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.0.resource_record_type", "CNAME"),
					resource.TestCheckResourceAttrSet("aws_acm_certificate.cert", "domain_validation_options.0.resource_record_value"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "subject_alternative_names.#", "0"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "validation_emails.#", "0"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "validation_method", acm.ValidationMethodDns),
				),
			},
			resource.TestStep{
				ResourceName:      "aws_acm_certificate.cert",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAcmCertificate_wildcardAndRootSan(t *testing.T) {
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)
	wildcardDomain := fmt.Sprintf("*.%s", rootDomain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAcmCertificateConfig_subjectAlternativeNames(wildcardDomain, strconv.Quote(rootDomain), acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_acm_certificate.cert", "arn", certificateArnRegex),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_name", wildcardDomain),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.#", "2"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.0.domain_name", wildcardDomain),
					resource.TestCheckResourceAttrSet("aws_acm_certificate.cert", "domain_validation_options.0.resource_record_name"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.0.resource_record_type", "CNAME"),
					resource.TestCheckResourceAttrSet("aws_acm_certificate.cert", "domain_validation_options.0.resource_record_value"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.1.domain_name", rootDomain),
					resource.TestCheckResourceAttrSet("aws_acm_certificate.cert", "domain_validation_options.1.resource_record_name"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_validation_options.1.resource_record_type", "CNAME"),
					resource.TestCheckResourceAttrSet("aws_acm_certificate.cert", "domain_validation_options.1.resource_record_value"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "subject_alternative_names.#", "1"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "subject_alternative_names.0", rootDomain),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "validation_emails.#", "0"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "validation_method", acm.ValidationMethodDns),
				),
			},
			resource.TestStep{
				ResourceName:      "aws_acm_certificate.cert",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAcmCertificate_tags(t *testing.T) {
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)

	rInt1 := acctest.RandInt()

	domain := fmt.Sprintf("tf-acc-%d.%s", rInt1, rootDomain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAcmCertificateConfig(domain, acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "tags.%", "0"),
				),
			},
			resource.TestStep{
				Config: testAccAcmCertificateConfig_twoTags(domain, acm.ValidationMethodDns, "Hello", "World", "Foo", "Bar"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "tags.%", "2"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "tags.Hello", "World"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "tags.Foo", "Bar"),
				),
			},
			resource.TestStep{
				Config: testAccAcmCertificateConfig_twoTags(domain, acm.ValidationMethodDns, "Hello", "World", "Foo", "Baz"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "tags.%", "2"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "tags.Hello", "World"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "tags.Foo", "Baz"),
				),
			},
			resource.TestStep{
				Config: testAccAcmCertificateConfig_oneTag(domain, acm.ValidationMethodDns, "Environment", "Test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "tags.Environment", "Test"),
				),
			},
			resource.TestStep{
				ResourceName:      "aws_acm_certificate.cert",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAcmCertificate_imported(t *testing.T) {
	resourceName := "aws_acm_certificate.cert"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfig_selfSigned("example"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com"),
				),
			},
			{
				Config: testAccAcmCertificateConfig_selfSigned("example2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example2.com"),
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

func testAccAcmCertificateConfig(domainName, validationMethod string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "cert" {
  domain_name       = "%s"
  validation_method = "%s"
}
`, domainName, validationMethod)

}

func testAccAcmCertificateConfig_subjectAlternativeNames(domainName, subjectAlternativeNames, validationMethod string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "cert" {
  domain_name               = "%s"
  subject_alternative_names = [%s]
  validation_method = "%s"
}
`, domainName, subjectAlternativeNames, validationMethod)
}

func testAccAcmCertificateConfig_oneTag(domainName, validationMethod, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "cert" {
  domain_name       = "%s"
  validation_method = "%s"

  tags {
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

  tags {
    "%s" = "%s"
    "%s" = "%s"
  }
}
`, domainName, validationMethod, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccAcmCertificateConfig_selfSigned(certName string) string {
	return fmt.Sprintf(`
resource "tls_private_key" "%[1]s" {
	algorithm = "RSA"
}

resource "tls_self_signed_cert" "%[1]s" {
	key_algorithm   = "RSA"
	private_key_pem = "${tls_private_key.%[1]s.private_key_pem}"

	subject {
		common_name  = "%[1]s.com"
		organization = "ACME Examples, Inc"
	}

	validity_period_hours = 12

	allowed_uses = [
		"key_encipherment",
		"digital_signature",
		"server_auth",
	]
}

resource "aws_acm_certificate" "cert" {
  private_key 		= "${tls_private_key.%[1]s.private_key_pem}"
  certificate_body  = "${tls_self_signed_cert.%[1]s.cert_pem}"
}
`, certName)
}

func testAccCheckAcmCertificateDestroy(s *terraform.State) error {
	acmconn := testAccProvider.Meta().(*AWSClient).acmconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_acm_certificate" {
			continue
		}
		_, err := acmconn.DescribeCertificate(&acm.DescribeCertificateInput{
			CertificateArn: aws.String(rs.Primary.ID),
		})

		if err == nil {
			return fmt.Errorf("Certificate still exists.")
		}

		// Verify the error is what we want
		if !isAWSErr(err, acm.ErrCodeResourceNotFoundException, "") {
			return err
		}
	}

	return nil
}
