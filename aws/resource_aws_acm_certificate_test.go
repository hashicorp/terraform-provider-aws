package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsAcmResource_certificateIssuingFlow(t *testing.T) {
	var conf acm.DescribeCertificateOutput

	root_zone_domain := "sandbox.sellmayr.net"
	domain := "certtest.sandbox.sellmayr.net"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAcmCertificateConfig(root_zone_domain, domain),
				// expect non-empty plan that's triggered by the depends_on; see https://github.com/hashicorp/terraform/issues/11806
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAcmCertificateExists("aws_acm_certificate.cert", &conf),
					testAccCheckAcmCertificateWasIssued(&conf),
					testAccCheckAcmCertificateAttributes("aws_acm_certificate.cert", &conf, domain),
					testAccCheckAcmDataSourceAttributes("data.aws_acm_certificate.cert", &conf, domain),
				),
			},
		},
	})
}
func testAccCheckAcmCertificateWasIssued(output *acm.DescribeCertificateOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *output.Certificate.Status != "ISSUED" {
			return fmt.Errorf("Expected certificate to be issued but was in status %s", output.Certificate.Status)
		}
		return nil
	}
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

func testAccCheckAcmCertificateAttributes(n string, cert *acm.DescribeCertificateOutput, domain string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		attrs := rs.Primary.Attributes

		if *cert.Certificate.DomainName != domain {
			return fmt.Errorf("Domain name is %s but expected %s", *cert.Certificate.DomainName, domain)
		}
		if attrs["domain_name"] != domain {
			return fmt.Errorf("Domain name in state is %s but expected %s", attrs["domain_name"], domain)
		}

		// TODO: check other attributes?

		return nil
	}
}

func testAccCheckAcmDataSourceAttributes(n string, cert *acm.DescribeCertificateOutput, domain string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		attrs := rs.Primary.Attributes

		if attrs["domain"] != domain {
			return fmt.Errorf("Domain name in state is %s but expected %s", attrs["domain"], domain)
		}

		if attrs["arn"] != *cert.Certificate.CertificateArn {
			return fmt.Errorf("Certificate ARN in state is %s but expected %s", attrs["arn"], *cert.Certificate.CertificateArn)
		}

		return nil
	}
}

func testAccAcmCertificateConfig(rootZoneDomain string, domain string) string {
	return fmt.Sprintf(`

data "aws_route53_zone" "zone" {
  name = "%s."
  private_zone = false
}

resource "aws_acm_certificate" "cert" {
    domain_name = "%s"
	validation_method = "DNS"
}
resource "aws_route53_record" "cert_validation" {
  name = "${aws_acm_certificate.cert.domain_validation_options.0.resource_record_name}"
  type = "${aws_acm_certificate.cert.domain_validation_options.0.resource_record_type}"
  zone_id = "${data.aws_route53_zone.zone.id}"
  records = ["${aws_acm_certificate.cert.domain_validation_options.0.resource_record_value}"]
  ttl = 60
}

data "aws_acm_certificate" "cert" {
  arn = "${aws_acm_certificate.cert.certificate_arn}"
  statuses = ["ISSUED"]
  wait_until_present = true
  depends_on = ["aws_route53_record.cert_validation"]
}

`, rootZoneDomain, domain)
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
