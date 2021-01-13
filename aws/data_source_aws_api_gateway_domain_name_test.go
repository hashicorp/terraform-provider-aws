package aws

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataSourceAwsApiGatewayDomainName_CertificateArn(t *testing.T) {
	certificateArn := os.Getenv("AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_ARN")
	if certificateArn == "" {
		t.Skip(
			"Environment variable AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_ARN is not set. " +
				"This environment variable must be set to the ARN of " +
				"an ISSUED ACM certificate in us-east-1 to enable this test.")
	}

	// This test must always run in us-east-1
	// BadRequestException: Invalid certificate ARN: arn:aws:acm:us-west-2:123456789012:certificate/xxxxx. Certificate must be in 'us-east-1'.
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	resourceName := "aws_api_gateway_domain_name.test"
	dataSourceName := "data.aws_api_gateway_domain_name.test"
	rName := fmt.Sprintf("tf-acc-%s.terraformtest.com", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsApiGatewayDomainNameConfig_CertificateArn(rName, certificateArn),
				Check: resource.ComposeTestCheckFunc(
					testAccMatchResourceAttrRegionalARNNoAccount(dataSourceName, "arn", "apigateway", regexp.MustCompile(`/domainnames/+.`)),
					resource.TestCheckResourceAttr(dataSourceName, "domain_name", rName),
					resource.TestCheckResourceAttr(dataSourceName, "cloudfront_zone_id", "Z2FDTNDATAQYW2"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", dataSourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudfront_domain_name", dataSourceName, "cloudfront_domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudfront_zone_id", dataSourceName, "cloudfront_zone_id"),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_upload_date", dataSourceName, "certificate_upload_date"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsApiGatewayDomainName_RegionalCertificateArn(t *testing.T) {
	resourceName := "aws_api_gateway_domain_name.test"
	dataSourceName := "data.aws_api_gateway_domain_name.test"
	rName := fmt.Sprintf("tf-acc-%s.terraformtest.com", acctest.RandString(8))

	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsApiGatewayDomainNameConfig_RegionalCertificateArn(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccMatchResourceAttrRegionalARNNoAccount(dataSourceName, "arn", "apigateway", regexp.MustCompile(`/domainnames/+.`)),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", dataSourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "regional_domain_name", dataSourceName, "regional_domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "regional_zone_id", dataSourceName, "regional_zone_id"),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_upload_date", dataSourceName, "certificate_upload_date"),
				),
			},
		},
	})
}

func testAccDataSourceAwsApiGatewayDomainNameConfig_CertificateArn(domainName, certificateArn string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_domain_name" "test" {
 domain_name     = "%s"
 certificate_arn = "%s"

 endpoint_configuration {
   types = ["EDGE"]
 }
}

data "aws_api_gateway_domain_name" "test" {
 domain_name = "${aws_api_gateway_domain_name.test.domain_name}"
}
`, domainName, certificateArn)
}

func testAccDataSourceAwsApiGatewayDomainNameConfig_RegionalCertificateArn(domainName, key, certificate string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
 certificate_body = "%[2]s"
 private_key      = "%[3]s"
}

resource "aws_api_gateway_domain_name" "test" {
 domain_name              = %[1]q
 regional_certificate_arn = "${aws_acm_certificate.test.arn}"

 endpoint_configuration {
   types = ["REGIONAL"]
 }
}

data "aws_api_gateway_domain_name" "test" {
 domain_name = "${aws_api_gateway_domain_name.test.domain_name}"
}
`, domainName, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(key))
}
