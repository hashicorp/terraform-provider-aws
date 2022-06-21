package apigateway_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAPIGatewayDomainNameDataSource_basic(t *testing.T) {
	resourceName := "aws_api_gateway_domain_name.test"
	dataSourceName := "data.aws_api_gateway_domain_name.test"
	rName := acctest.RandomSubdomain()

	key := acctest.TLSRSAPrivateKeyPEM(2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(key, rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameDataSourceConfig_regionalCertificateARN(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_arn", dataSourceName, "certificate_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_name", dataSourceName, "certificate_name"),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_upload_date", dataSourceName, "certificate_upload_date"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudfront_domain_name", dataSourceName, "cloudfront_domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudfront_zone_id", dataSourceName, "cloudfront_zone_id"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", dataSourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_configuration.#", dataSourceName, "endpoint_configuration.#"),
					resource.TestCheckResourceAttrPair(resourceName, "regional_certificate_arn", dataSourceName, "regional_certificate_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "regional_certificate_name", dataSourceName, "regional_certificate_name"),
					resource.TestCheckResourceAttrPair(resourceName, "regional_domain_name", dataSourceName, "regional_domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "regional_zone_id", dataSourceName, "regional_zone_id"),
					resource.TestCheckResourceAttrPair(resourceName, "security_policy", dataSourceName, "security_policy"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccDomainNameDataSourceConfig_regionalCertificateARN(domainName, key, certificate string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_api_gateway_domain_name" "test" {
  domain_name              = %[1]q
  regional_certificate_arn = aws_acm_certificate.test.arn

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}

data "aws_api_gateway_domain_name" "test" {
  domain_name = aws_api_gateway_domain_name.test.domain_name
}
`, domainName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
}
