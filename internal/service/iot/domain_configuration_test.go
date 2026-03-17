// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iot_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIoTDomainConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	resourceName := "aws_iot_domain_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfigurationConfig_basic(rName, rootDomain, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainConfigurationExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "iot", regexache.MustCompile(`domainconfiguration/`+rName+`/[a-z0-9]+$`)),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domain),
					resource.TestCheckResourceAttr(resourceName, "domain_type", "CUSTOMER_MANAGED"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "server_certificate_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "service_type", "DATA"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "tls_config.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "tls_config.0.security_policy"),
					resource.TestCheckResourceAttr(resourceName, "validation_certificate_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIoTDomainConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	resourceName := "aws_iot_domain_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfigurationConfig_basic(rName, rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainConfigurationExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "iot", regexache.MustCompile(`domainconfiguration/`+rName+`/[a-z0-9]+$`)),
					acctest.CheckSDKResourceDisappears(ctx, t, tfiot.ResourceDomainConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIoTDomainConfiguration_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	resourceName := "aws_iot_domain_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfigurationConfig_tags1(rName, rootDomain, domain, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfigurationConfig_tags2(rName, rootDomain, domain, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDomainConfigurationConfig_tags1(rName, rootDomain, domain, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccIoTDomainConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	resourceName := "aws_iot_domain_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfigurationConfig_securityPolicy(rName, rootDomain, domain, "IoTSecurityPolicy_TLS13_1_3_2022_10", true, "CUSTOM_AUTH", "MQTT_WSS"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "application_protocol", "MQTT_WSS"),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "CUSTOM_AUTH"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_config.0.allow_authorizer_override", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "tls_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_config.0.security_policy", "IoTSecurityPolicy_TLS13_1_3_2022_10"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfigurationConfig_securityPolicy(rName, rootDomain, domain, "IoTSecurityPolicy_TLS13_1_2_2022_10", false, "CUSTOM_AUTH_X509", "HTTPS"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "application_protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "CUSTOM_AUTH_X509"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_config.0.allow_authorizer_override", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "tls_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_config.0.security_policy", "IoTSecurityPolicy_TLS13_1_2_2022_10"),
				),
			},
		},
	})
}

func TestAccIoTDomainConfiguration_awsManaged(t *testing.T) { // nosemgrep:ci.aws-in-func-name
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iot_domain_configuration.test"

	acctest.SkipIfEnvVarNotSet(t, "IOT_DOMAIN_CONFIGURATION_TEST_AWS_MANAGED")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfigurationConfig_awsManaged(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainConfigurationExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "iot", regexache.MustCompile(`domainconfiguration/`+rName+`/[a-z0-9]+$`)),
					resource.TestCheckResourceAttr(resourceName, "authorizer_config.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDomainName),
					resource.TestCheckResourceAttr(resourceName, "domain_type", "AWS_MANAGED"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "server_certificate_arns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "service_type", "DATA"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "tls_config.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "tls_config.0.security_policy"),
					resource.TestCheckResourceAttr(resourceName, "validation_certificate_arn", ""),
				),
			},
		},
	})
}

func testAccCheckDomainConfigurationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).IoTClient(ctx)

		_, err := tfiot.FindDomainConfigurationByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckDomainConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IoTClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iot_domain_configuration" {
				continue
			}

			_, err := tfiot.FindDomainConfigurationByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IoT Domain Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDomainConfigurationConfig_base(rootDomain, domain string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name       = %[2]q
  validation_method = "DNS"
}

data "aws_route53_zone" "test" {
  name         = %[1]q
  private_zone = false
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
  depends_on = [aws_route53_record.test]

  certificate_arn = aws_acm_certificate.test.arn
}
`, rootDomain, domain)
}

func testAccDomainConfigurationConfig_basic(rName, rootDomain, domain string) string {
	return acctest.ConfigCompose(testAccDomainConfigurationConfig_base(rootDomain, domain), fmt.Sprintf(`
resource "aws_iot_domain_configuration" "test" {
  depends_on = [aws_acm_certificate_validation.test]

  name                    = %[1]q
  domain_name             = %[2]q
  server_certificate_arns = [aws_acm_certificate.test.arn]
}
`, rName, domain))
}

func testAccDomainConfigurationConfig_tags1(rName, rootDomain, domain, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccDomainConfigurationConfig_base(rootDomain, domain), fmt.Sprintf(`
resource "aws_iot_domain_configuration" "test" {
  depends_on = [aws_acm_certificate_validation.test]

  name                    = %[1]q
  domain_name             = %[2]q
  server_certificate_arns = [aws_acm_certificate.test.arn]

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, domain, tagKey1, tagValue1))
}

func testAccDomainConfigurationConfig_tags2(rName, rootDomain, domain, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccDomainConfigurationConfig_base(rootDomain, domain), fmt.Sprintf(`
resource "aws_iot_domain_configuration" "test" {
  depends_on = [aws_acm_certificate_validation.test]

  name                    = %[1]q
  domain_name             = %[2]q
  server_certificate_arns = [aws_acm_certificate.test.arn]

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, domain, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccDomainConfigurationConfig_securityPolicy(rName, rootDomain, domain, securityPolicy string, allowAuthorizerOverride bool, authenticationType, applicationProtocol string) string {
	return acctest.ConfigCompose(testAccAuthorizerConfig_basic(rName), testAccDomainConfigurationConfig_base(rootDomain, domain), fmt.Sprintf(`
resource "aws_iot_domain_configuration" "test" {
  depends_on = [aws_acm_certificate_validation.test]

  authentication_type  = %[5]q
  application_protocol = %[6]q

  authorizer_config {
    allow_authorizer_override = %[4]t
    default_authorizer_name   = aws_iot_authorizer.test.name
  }

  name                    = %[1]q
  domain_name             = %[2]q
  server_certificate_arns = [aws_acm_certificate.test.arn]

  tls_config {
    security_policy = %[3]q
  }
}
`, rName, domain, securityPolicy, allowAuthorizerOverride, authenticationType, applicationProtocol))
}

func testAccDomainConfigurationConfig_awsManaged(rName string) string { // nosemgrep:ci.aws-in-func-name
	return fmt.Sprintf(`
resource "aws_iot_domain_configuration" "test" {
  name = %[1]q
}
`, rName)
}
