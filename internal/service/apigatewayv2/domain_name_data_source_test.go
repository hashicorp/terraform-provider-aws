// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayV2DomainNameDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	testName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_apigatewayv2_domain_name.test"

	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	domainName := fmt.Sprintf("%s.example.com", testName)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domainName)

	resourceName := fmt.Sprintf("aws_apigatewayv2_domain_name.%s", testName)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameDataSourceConfig_basic(testName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", dataSourceName, "domain_name"),
					resource.TestCheckResourceAttr(dataSourceName, "domain_name_configurations.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", dataSourceName, "domain_name_configurations.0.certificate_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_name", dataSourceName, "domain_name_configurations.0.certificate_name"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_upload_date", dataSourceName, "domain_name_configurations.0.certificate_upload_date"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.endpoint_type", dataSourceName, "domain_name_configurations.0.endpoint_type"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttr(dataSourceName, "domain_name_configurations.0.domain_name_status", "AVAILABLE"),
					resource.TestCheckResourceAttr(dataSourceName, "domain_name_configurations.0.security_policy", "TLS_1_2"),
					resource.TestCheckNoResourceAttr(dataSourceName, "domain_name_configurations.0.domain_name_status_message"),
					resource.TestCheckNoResourceAttr(dataSourceName, "domain_name_configurations.0.ownership_verification_certificate_arn"),
				),
			},
		},
	})
}

func testAccDomainNameDataSourceConfig_basic(rName, key, certificate string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "%[1]s" {	
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_apigatewayv2_domain_name" "%[1]s" {
  domain_name              = aws_acm_certificate.%[1]s.domain_name

  domain_name_configuration {
    certificate_arn = aws_acm_certificate.%[1]s.arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }

  tags = {
    Description = "No mutual TLS"
  }
}

data "aws_apigatewayv2_domain_name" "test" {
  domain_name = aws_apigatewayv2_domain_name.%[1]s.domain_name
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
}
