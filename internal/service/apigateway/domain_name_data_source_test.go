// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayDomainNameDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_api_gateway_domain_name.test"
	dataSourceName := "data.aws_api_gateway_domain_name.test"
	rName := acctest.RandomSubdomain()
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameDataSourceConfig_regionalCertificateARN(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCertificateARN, dataSourceName, names.AttrCertificateARN),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_name", dataSourceName, "certificate_name"),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_upload_date", dataSourceName, "certificate_upload_date"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudfront_domain_name", dataSourceName, "cloudfront_domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudfront_zone_id", dataSourceName, "cloudfront_zone_id"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDomainName, dataSourceName, names.AttrDomainName),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_configuration.#", dataSourceName, "endpoint_configuration.#"),
					resource.TestCheckResourceAttrPair(resourceName, "regional_certificate_arn", dataSourceName, "regional_certificate_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "regional_certificate_name", dataSourceName, "regional_certificate_name"),
					resource.TestCheckResourceAttrPair(resourceName, "regional_domain_name", dataSourceName, "regional_domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "regional_zone_id", dataSourceName, "regional_zone_id"),
					resource.TestCheckResourceAttrPair(resourceName, "security_policy", dataSourceName, "security_policy"),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
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
