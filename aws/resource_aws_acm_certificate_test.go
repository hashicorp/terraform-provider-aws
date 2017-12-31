package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"regexp"
)

func TestAccAwsAcmResource_certificateIssuingFlow(t *testing.T) {
	var conf acm.DescribeCertificateOutput
	var confAfterValidation acm.DescribeCertificateOutput

	root_zone_domain := "sandbox.sellmayr.net"
	domain := "certtest.sandbox.sellmayr.net"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			// Test that we can request a certificate
			resource.TestStep{
				Config: testAccAcmCertificateConfig(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAcmCertificateExists("aws_acm_certificate.cert", &conf),
					testAccCheckAcmCertificateAttributes("aws_acm_certificate.cert", &conf, domain, "PENDING_VALIDATION"),
				),
			},
			// Test that validation times out if certificate can't be validated
			resource.TestStep{
				Config:      testAccAcmCertificateWithValidationConfig(domain),
				ExpectError: regexp.MustCompile("Expected certificate to be issued but was in state PENDING_VALIDATION"),
			},
			// Test that validation succeeds once we provide the right DNS validation records
			resource.TestStep{
				Config: testAccAcmCertificateWithValidationAndRecordsConfig(root_zone_domain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAcmCertificateExists("aws_acm_certificate.cert", &confAfterValidation),
					testAccCheckAcmCertificateAttributes("aws_acm_certificate.cert", &confAfterValidation, domain, "ISSUED"),
					testAccCheckAcmCertificateValidationAttributes("aws_acm_certificate_validation.cert", &confAfterValidation),
				),
			},
		},
	})
}

func testAccAcmCertificateConfig(domain string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "cert" {
    domain_name = "%s"
	validation_method = "DNS"
}
`, domain)
}

func testAccAcmCertificateWithValidationConfig(domain string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "cert" {
    domain_name = "%s"
	validation_method = "DNS"
}

resource "aws_acm_certificate_validation" "cert" {
  certificate_arn = "${aws_acm_certificate.cert.certificate_arn}"
  timeout = "20s"
}
`, domain)
}

func testAccAcmCertificateWithValidationAndRecordsConfig(rootZoneDomain string, domain string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "cert" {
    domain_name = "%s"
	validation_method = "DNS"
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

resource "aws_acm_certificate_validation" "cert" {
  certificate_arn = "${aws_acm_certificate.cert.certificate_arn}"
  validation_record_fqdn = "${aws_route53_record.cert_validation.fqdn}" # This wouldn't strictly be necessary but it can enforce a dependency
}
`, domain, rootZoneDomain)
}

func testAccCheckAcmCertificateExists(n string, res *acm.DescribeCertificateOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No id is set")
		}

		if rs.Primary.Attributes["certificate_arn"] == "" {
			return fmt.Errorf("No certificate_arn is set")
		}

		if rs.Primary.Attributes["certificate_arn"] != rs.Primary.ID {
			return fmt.Errorf("No certificate_arn and ID are different: %s %s", rs.Primary.Attributes["certificate_arn"], rs.Primary.ID)
		}

		acmconn := testAccProvider.Meta().(*AWSClient).acmconn

		resp, err := acmconn.DescribeCertificate(&acm.DescribeCertificateInput{
			CertificateArn: aws.String(rs.Primary.Attributes["certificate_arn"]),
		})

		if err != nil {
			return err
		}

		*res = *resp

		return nil
	}
}

func testAccCheckAcmCertificateAttributes(n string, cert *acm.DescribeCertificateOutput, domain string, status string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		attrs := rs.Primary.Attributes

		if *cert.Certificate.DomainName != domain {
			return fmt.Errorf("Domain name is %s but expected %s", *cert.Certificate.DomainName, domain)
		}
		if *cert.Certificate.Status != status {
			return fmt.Errorf("Status is %s but expected %s", *cert.Certificate.Status, status)
		}
		if attrs["domain_name"] != domain {
			return fmt.Errorf("Domain name in state is %s but expected %s", attrs["domain_name"], domain)
		}

		// TODO: check other attributes?

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
		acmerr, ok := err.(awserr.Error)

		if !ok {
			return err
		}
		if acmerr.Code() != "ResourceNotFoundException" {
			return err
		}
	}

	return nil
}
