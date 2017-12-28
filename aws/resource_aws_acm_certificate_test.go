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

func TestAccAcmCertificate_basic(t *testing.T) {
	var conf acm.DescribeCertificateOutput

	domain := "certtest.hashicorp.com"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAcmCertificateConfig(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAcmCertificateExists("aws_acm_certificate.cert", &conf),
					testAccCheckAcmCertificateAttributes("aws_acm_certificate.cert", &conf, domain),
				),
			},
		},
	})
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
			CertificateArn: aws.String(rs.Primary.ID),
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
			return fmt.Errorf("Domain name is %s but expected %s", cert.Certificate.DomainName, domain)
		}
		if attrs["domain_name"] != domain {
			return fmt.Errorf("Domain name in state is %s but expected %s", attrs["domain_name"], domain)
		}

		// TODO: check other attributes?

		return nil
	}
}

func testAccAcmCertificateConfig(domain string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "cert" {
    domain_name = "%s"
	validation_method = "DNS"
}
`, domain)
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
