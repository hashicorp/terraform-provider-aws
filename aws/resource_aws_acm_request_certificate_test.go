package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAcmCertificateRequest_createReq(t *testing.T) {
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateRequestConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAcmCertificateRequestExists(),
				),
			},
		},
	})
}

func testAccCheckAcmCertificateRequestDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).acmconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_acm_request_certificate" {
			continue
		}

		response, err := conn.DescribeCertificate(&acm.DescribeCertificateInput{
			CertificateArn: aws.String(rs.Primary.Attributes["arn"]),
		})
		if err != nil {
			if acmerr, ok := err.(awserr.Error); ok {
				if acmerr.Code() == "ResourceNotFoundException" {
					return nil
				}
			}
			return fmt.Errorf("Error occured when get certificate description.")
		}

		if response.Certificate != nil {
			return fmt.Errorf("Certificate still exists")
		}
	}

	return nil
}

func testAccCheckAcmCertificateRequestExists() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).acmconn
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_acm_request_certificate" {
				continue
			}

			response, err := conn.DescribeCertificate(&acm.DescribeCertificateInput{
				CertificateArn: aws.String(rs.Primary.Attributes["arn"]),
			})

			if err != nil {
				if acmerr, ok := err.(awserr.Error); ok {
					if acmerr.Code() == "ResourceNotFoundException" {
						return fmt.Errorf("Certificate does not exists")
					}
				}
				return fmt.Errorf("Error occured when get certificate description.")
			}

			if response.Certificate == nil {
				return fmt.Errorf("Certificate does not exists")
			}
		}
		return nil
	}
}

func testAccAcmCertificateRequestConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_acm_request_certificate" "this" {
  domain_name = "test-domaini-%d.example.com"
}
  `, rInt)
}
