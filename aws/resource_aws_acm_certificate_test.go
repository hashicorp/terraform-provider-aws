package aws

import (
	"fmt"
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

func TestAccAwsAcmResource_emailValidation(t *testing.T) {
	if os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN") == "" {
		t.Skip("Environment variable ACM_CERTIFICATE_ROOT_DOMAIN is not set")
	}

	root_zone_domain := os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN")

	rInt1 := acctest.RandInt()

	domain := fmt.Sprintf("tf-acc-%d.%s", rInt1, root_zone_domain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			// Test that we can request a certificate
			resource.TestStep{
				Config: testAccAcmCertificateConfigWithEMailValidation(domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_acm_certificate.cert", "arn", certificateArnRegex),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_name", domain),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "subject_alternative_names.#", "0"),
					resource.TestMatchResourceAttr("aws_acm_certificate.cert", "validation_emails.0", regexp.MustCompile(`^[^@]+@.+$`)),
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
func TestAccAwsAcmResource_certificateIssuingFlow(t *testing.T) {
	if os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN") == "" {
		t.Skip("Environment variable ACM_CERTIFICATE_ROOT_DOMAIN is not set")
	}

	root_zone_domain := os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN")

	rInt1 := acctest.RandInt()

	domain := fmt.Sprintf("tf-acc-%d.%s", rInt1, root_zone_domain)
	sanDomain := fmt.Sprintf("tf-acc-%d-san.%s", rInt1, root_zone_domain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			// Test that we can request a certificate
			resource.TestStep{
				Config: testAccAcmCertificateConfig(domain, sanDomain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_acm_certificate.cert", "arn", certificateArnRegex),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "domain_name", domain),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "subject_alternative_names.#", "1"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "subject_alternative_names.0", sanDomain),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "tags.%", "2"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "tags.Hello", "World"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "tags.Foo", "Bar"),
				),
			},
			// Test that we can change the tags
			resource.TestStep{
				Config: testAccAcmCertificateConfigWithChangedTags(domain, sanDomain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "tags.%", "2"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "tags.Environment", "Test"),
					resource.TestCheckResourceAttr("aws_acm_certificate.cert", "tags.Foo", "Baz"),
				),
			},
			// Test that validation times out if certificate can't be validated
			resource.TestStep{
				Config:      testAccAcmCertificateWithValidationConfig(domain, sanDomain),
				ExpectError: regexp.MustCompile("Expected certificate to be issued but was in state PENDING_VALIDATION"),
			},
			// Test that validation fails if given validation_fqdns don't match
			resource.TestStep{
				Config:      testAccAcmCertificateWithValidationConfigAndWrongFQDN(domain, sanDomain),
				ExpectError: regexp.MustCompile("Certificate needs .* to be set but only .* was passed to validation_record_fqdns"),
			},
			// Test that validation succeeds once we provide the right DNS validation records
			resource.TestStep{
				Config: testAccAcmCertificateWithValidationAndRecordsConfig(root_zone_domain, domain, sanDomain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_acm_certificate_validation.cert", "certificate_arn", certificateArnRegex),
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

func testAccAcmCertificateConfigWithEMailValidation(domain string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "cert" {
  domain_name = "%s"
  validation_method = "EMAIL"
}
`, domain)

}

func testAccAcmCertificateConfig(domain string, sanDomain string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "cert" {
  domain_name = "%s"
  validation_method = "DNS"
  subject_alternative_names = ["%s"]

  tags {
    "Hello" = "World"
    "Foo" = "Bar"
  }
}
`, domain, sanDomain)
}

func testAccAcmCertificateConfigWithChangedTags(domain string, sanDomain string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "cert" {
  domain_name = "%s"
  validation_method = "DNS"
  subject_alternative_names = ["%s"]

  tags {
    "Environment" = "Test"
    "Foo" = "Baz"
  }
}
`, domain, sanDomain)
}

func testAccAcmCertificateWithValidationConfig(domain string, sanDomain string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "cert" {
  domain_name = "%s"
  validation_method = "DNS"
  subject_alternative_names = ["%s"]

  tags {
    "Environment" = "Test"
    "Foo" = "Baz"
  }
}


resource "aws_acm_certificate_validation" "cert" {
  certificate_arn = "${aws_acm_certificate.cert.arn}"
  timeout = "20s"
}
`, domain, sanDomain)
}

func testAccAcmCertificateWithValidationConfigAndWrongFQDN(domain string, sanDomain string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "cert" {
  domain_name = "%s"
  validation_method = "DNS"
  subject_alternative_names = ["%s"]

  tags {
    "Environment" = "Test"
    "Foo" = "Baz"
  }
}


resource "aws_acm_certificate_validation" "cert" {
  certificate_arn = "${aws_acm_certificate.cert.arn}"
  validation_record_fqdns = ["some-wrong-fqdn.example.com"]
  timeout = "20s"
}
`, domain, sanDomain)
}

func testAccAcmCertificateWithValidationAndRecordsConfig(rootZoneDomain string, domain string, sanDomain string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "cert" {
  domain_name = "%s"
  validation_method = "DNS"
  subject_alternative_names = ["%s"]

  tags {
    "Environment" = "Test"
    "Foo" = "Baz"
  }
}


data "aws_route53_zone" "zone" {
  name = "%s."
  private_zone = false
}

resource "aws_route53_record" "cert_validation" {
  name = "${aws_acm_certificate.cert.domain_validation_options.0.resource_record_name}"
  type = "${aws_acm_certificate.cert.domain_validation_options.0.resource_record_type}"
  zone_id = "${data.aws_route53_zone.zone.id}"
  records = ["${aws_acm_certificate.cert.domain_validation_options.0.resource_record_value}"]
  ttl = 60
}

resource "aws_route53_record" "cert_validation_san" {
  name = "${aws_acm_certificate.cert.domain_validation_options.1.resource_record_name}"
  type = "${aws_acm_certificate.cert.domain_validation_options.1.resource_record_type}"
  zone_id = "${data.aws_route53_zone.zone.id}"
  records = ["${aws_acm_certificate.cert.domain_validation_options.1.resource_record_value}"]
  ttl = 60
}

resource "aws_acm_certificate_validation" "cert" {
  certificate_arn = "${aws_acm_certificate.cert.arn}"
  validation_record_fqdns = [
	"${aws_route53_record.cert_validation.fqdn}",
	"${aws_route53_record.cert_validation_san.fqdn}"
  ]
}
`, domain, sanDomain, rootZoneDomain)
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
