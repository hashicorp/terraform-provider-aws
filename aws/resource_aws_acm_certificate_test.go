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
					resource.TestMatchResourceAttr("aws_acm_certificate.cert", "arn", regexp.MustCompile(`^arn:aws:acm:[^:]+:[^:]+:certificate/.+$`)),
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
	var conf acm.DescribeCertificateOutput
	var tags acm.ListTagsForCertificateOutput
	var confAfterValidation acm.DescribeCertificateOutput
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
					testAccCheckAcmCertificateExists("aws_acm_certificate.cert", &conf, &tags),
					testAccCheckAcmCertificateAttributes("aws_acm_certificate.cert", &conf, domain, sanDomain, "PENDING_VALIDATION"),

					testAccCheckTagsACM(&tags.Tags, "Hello", "World"),
					testAccCheckTagsACM(&tags.Tags, "Foo", "Bar"),

					resource.TestMatchResourceAttr("aws_acm_certificate.cert", "arn", regexp.MustCompile(`^arn:aws:acm:[^:]+:[^:]+:certificate/.+$`)),
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
					testAccCheckAcmCertificateExists("aws_acm_certificate.cert", &conf, &tags),
					testAccCheckAcmCertificateAttributes("aws_acm_certificate.cert", &conf, domain, sanDomain, "PENDING_VALIDATION"),

					testAccCheckTagsACM(&tags.Tags, "Environment", "Test"),
					testAccCheckTagsACM(&tags.Tags, "Foo", "Baz"),

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
					testAccCheckAcmCertificateExists("aws_acm_certificate.cert", &confAfterValidation, &tags),
					testAccCheckAcmCertificateAttributes("aws_acm_certificate.cert", &confAfterValidation, domain, sanDomain, "ISSUED"),
					testAccCheckAcmCertificateValidationAttributes("aws_acm_certificate_validation.cert", &confAfterValidation),
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

func testAccCheckAcmCertificateExists(n string, res *acm.DescribeCertificateOutput, tags *acm.ListTagsForCertificateOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No id is set")
		}

		if rs.Primary.Attributes["arn"] == "" {
			return fmt.Errorf("No arn is set")
		}

		if rs.Primary.Attributes["arn"] != rs.Primary.ID {
			return fmt.Errorf("No arn and ID are different: %s %s", rs.Primary.Attributes["arn"], rs.Primary.ID)
		}

		acmconn := testAccProvider.Meta().(*AWSClient).acmconn

		resp, err := acmconn.DescribeCertificate(&acm.DescribeCertificateInput{
			CertificateArn: aws.String(rs.Primary.Attributes["arn"]),
		})

		if err != nil {
			return err
		}

		tagsResp, err := acmconn.ListTagsForCertificate(&acm.ListTagsForCertificateInput{
			CertificateArn: aws.String(rs.Primary.Attributes["arn"]),
		})

		if err != nil {
			return err
		}

		*res = *resp
		*tags = *tagsResp

		return nil
	}
}

func testAccCheckAcmCertificateAttributes(n string, cert *acm.DescribeCertificateOutput, domain string, sanDomain string, status string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		attrs := rs.Primary.Attributes

		if *cert.Certificate.DomainName != domain {
			return fmt.Errorf("Domain name is %s but expected %s", *cert.Certificate.DomainName, domain)
		}
		if *cert.Certificate.SubjectAlternativeNames[1] != sanDomain {
			return fmt.Errorf("SAN Domain name is %s but expected %s", *cert.Certificate.SubjectAlternativeNames[1], sanDomain)
		}
		if *cert.Certificate.Status != status {
			return fmt.Errorf("Status is %s but expected %s", *cert.Certificate.Status, status)
		}
		if attrs["domain_name"] != domain {
			return fmt.Errorf("Domain name in state is %s but expected %s", attrs["domain_name"], domain)
		}

		return nil
	}
}

func testAccCheckAcmCertificateValidationAttributes(n string, cert *acm.DescribeCertificateOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s.", n)
		}
		attrs := rs.Primary.Attributes

		if attrs["certificate_arn"] != *cert.Certificate.CertificateArn {
			return fmt.Errorf("Certificate ARN in state is %s but expected %s", attrs["arn"], *cert.Certificate.CertificateArn)
		}

		return nil
	}
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
