// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package signer_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/signer"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSignerSigningProfileDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_signer_signing_profile.test"
	resourceName := "aws_signer_signing_profile.test"
	rString := sdkacctest.RandString(48)
	profileName := fmt.Sprintf("tf_acc_sp_basic_%s", rString)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSingerSigningProfile(ctx, t, "AWSLambda-SHA384-ECDSA")
		},
		ErrorCheck:               acctest.ErrorCheck(t, signer.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSigningProfileDataSourceConfig_basic(profileName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "platform_id", resourceName, "platform_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "signature_validity_period.value", resourceName, "signature_validity_period.value"),
					resource.TestCheckResourceAttrPair(dataSourceName, "signature_validity_period.type", resourceName, "signature_validity_period.type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "platform_display_name", resourceName, "platform_display_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStatus, resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrTags, resourceName, names.AttrTags),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccSignerSigningProfileDataSource_signingParameters(t *testing.T) {
	ctx := acctest.Context(t)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domainName := acctest.ACMCertificateRandomSubDomain(rootDomain)
	var conf signer.GetSigningProfileOutput
	rName := fmt.Sprintf("tf_acc_test_%d", acctest.RandInt(t))
	dataSourceName := "data.aws_signer_signing_profile.test"
	resourceName := "aws_signer_signing_profile.test_sp"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSingerSigningProfile(ctx, t, "AmazonFreeRTOS-Default")
		},
		ErrorCheck:               acctest.ErrorCheck(t, signer.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSigningProfileDataSourceConfig_signingParameters(rName, rootDomain, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSigningProfileExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttrPair(dataSourceName, "platform_id", resourceName, "platform_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "signing_material.#", resourceName, "signing_material.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "signing_material.0.certificate_arn", resourceName, "signing_material.0.certificate_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "signing_parameters.%", resourceName, "signing_parameters.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "signing_parameters.param1", resourceName, "signing_parameters.param1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "signing_parameters.param2", resourceName, "signing_parameters.param2"),
				),
			},
		},
	})
}

func testAccSigningProfileDataSourceConfig_basic(profileName string) string {
	return fmt.Sprintf(`
resource "aws_signer_signing_profile" "test" {
  platform_id = "AWSLambda-SHA384-ECDSA"
  name        = "%s"
}

data "aws_signer_signing_profile" "test" {
  name = aws_signer_signing_profile.test.name
}`, profileName)
}

func testAccSigningProfileDataSourceConfig_signingParameters(rName, rootDomain, domainName string) string {
	return fmt.Sprintf(`
data "aws_route53_zone" "test" {
  name         = %[2]q
  private_zone = false
}

resource "aws_acm_certificate" "test" {
  domain_name       = %[3]q
  validation_method = "DNS"
}

resource "aws_route53_record" "test" {
  allow_overwrite = true
  name            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_name
  records         = [tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_value]
  ttl             = 60
  type            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_type
  zone_id         = data.aws_route53_zone.test.zone_id
}

resource "aws_acm_certificate_validation" "test" {
  certificate_arn         = aws_acm_certificate.test.arn
  validation_record_fqdns = [aws_route53_record.test.fqdn]
}

resource "aws_signer_signing_profile" "test_sp" {
  platform_id = "AmazonFreeRTOS-Default"
  name        = %[1]q

  signing_material {
    certificate_arn = aws_acm_certificate.test.arn
  }
  signing_parameters = {
    "param1" = "value1"
    "param2" = "value2"
  }
  depends_on = [aws_acm_certificate_validation.test]
}

data "aws_signer_signing_profile" "test" {
  name = aws_signer_signing_profile.test_sp.name
}`, rName, rootDomain, domainName)
}
