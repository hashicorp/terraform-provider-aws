package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSAPIGatewayDomainName_basic(t *testing.T) {
	var conf apigateway.DomainName

	rString := acctest.RandString(8)
	name := fmt.Sprintf("tf-acc-%s.terraformtest.com", rString)
	nameModified := fmt.Sprintf("tf-acc-%s-modified.terraformtest.com", rString)
	commonName := "*.terraformtest.com"
	certRe := regexp.MustCompile("^-----BEGIN CERTIFICATE-----\n")
	keyRe := regexp.MustCompile("^-----BEGIN RSA PRIVATE KEY-----\n")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAWSAPIGatewayDomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDomainNameConfig(name, commonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDomainNameExists("aws_api_gateway_domain_name.test", &conf),
					resource.TestMatchResourceAttr("aws_api_gateway_domain_name.test", "certificate_body", certRe),
					resource.TestMatchResourceAttr("aws_api_gateway_domain_name.test", "certificate_chain", certRe),
					resource.TestCheckResourceAttr("aws_api_gateway_domain_name.test", "certificate_name", "tf-acc-apigateway-domain-name"),
					resource.TestMatchResourceAttr("aws_api_gateway_domain_name.test", "certificate_private_key", keyRe),
					resource.TestCheckResourceAttr("aws_api_gateway_domain_name.test", "domain_name", name),
					resource.TestCheckResourceAttrSet("aws_api_gateway_domain_name.test", "certificate_upload_date"),
				),
			},
			{
				Config: testAccAWSAPIGatewayDomainNameConfig(nameModified, commonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDomainNameExists("aws_api_gateway_domain_name.test", &conf),
					resource.TestMatchResourceAttr("aws_api_gateway_domain_name.test", "certificate_body", certRe),
					resource.TestMatchResourceAttr("aws_api_gateway_domain_name.test", "certificate_chain", certRe),
					resource.TestCheckResourceAttr("aws_api_gateway_domain_name.test", "certificate_name", "tf-acc-apigateway-domain-name"),
					resource.TestMatchResourceAttr("aws_api_gateway_domain_name.test", "certificate_private_key", keyRe),
					resource.TestCheckResourceAttr("aws_api_gateway_domain_name.test", "domain_name", nameModified),
					resource.TestCheckResourceAttrSet("aws_api_gateway_domain_name.test", "certificate_upload_date"),
				),
			},
		},
	})
}

func testAccCheckAWSAPIGatewayDomainNameExists(n string, res *apigateway.DomainName) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway DomainName ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigateway

		req := &apigateway.GetDomainNameInput{
			DomainName: aws.String(rs.Primary.ID),
		}
		describe, err := conn.GetDomainName(req)
		if err != nil {
			return err
		}

		if *describe.DomainName != rs.Primary.ID {
			return fmt.Errorf("APIGateway DomainName not found")
		}

		*res = *describe

		return nil
	}
}

func testAccCheckAWSAPIGatewayDomainNameDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigateway

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_api_key" {
			continue
		}

		describe, err := conn.GetDomainNames(&apigateway.GetDomainNamesInput{})

		if err == nil {
			if len(describe.Items) != 0 &&
				*describe.Items[0].DomainName == rs.Primary.ID {
				return fmt.Errorf("API Gateway DomainName still exists")
			}
		}

		aws2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if aws2err.Code() != "NotFoundException" {
			return err
		}

		return nil
	}

	return nil
}

func testAccAWSAPIGatewayCerts(commonName string) string {
	return fmt.Sprintf(`
resource "tls_private_key" "test" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "ca" {
  key_algorithm         = "RSA"
  private_key_pem       = "${tls_private_key.test.private_key_pem}"
  is_ca_certificate     = true
  validity_period_hours = 12

  subject {
    common_name  = "ACME Root CA"
    organization = "ACME Example Holdings"
  }

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "tls_cert_request" "test" {
  key_algorithm   = "RSA"
  private_key_pem = "${tls_private_key.test.private_key_pem}"

  subject {
    common_name  = "%s"
    organization = "ACME Example Holdings, Inc"
  }
}

resource "tls_locally_signed_cert" "leaf" {
  cert_request_pem      = "${tls_cert_request.test.cert_request_pem}"
  ca_key_algorithm      = "RSA"
  ca_private_key_pem    = "${tls_private_key.test.private_key_pem}"
  ca_cert_pem           = "${tls_self_signed_cert.ca.cert_pem}"
  validity_period_hours = 12

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}
`, commonName)
}

func testAccAWSAPIGatewayDomainNameConfig(domainName, commonName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_domain_name" "test" {
  domain_name = "%s"
  certificate_body = "${tls_locally_signed_cert.leaf.cert_pem}"
  certificate_chain = "${tls_self_signed_cert.ca.cert_pem}"
  certificate_name = "tf-acc-apigateway-domain-name"
  certificate_private_key = "${tls_private_key.test.private_key_pem}"
}
%s
`, domainName, testAccAWSAPIGatewayCerts(commonName))
}
