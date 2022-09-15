package cognitoidp_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcognitoidp "github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidp"
)

func TestAccCognitoIDPUserPoolDomain_basic(t *testing.T) {
	domainName := fmt.Sprintf("tf-acc-test-domain-%d", sdkacctest.RandInt())
	poolName := fmt.Sprintf("tf-acc-test-pool-%s", sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:               acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolDomainConfig_basic(domainName, poolName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolDomainExists("aws_cognito_user_pool_domain.main"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_domain.main", "domain", domainName),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "name", poolName),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool_domain.main", "aws_account_id"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool_domain.main", "cloudfront_distribution_arn"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool_domain.main", "s3_bucket"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool_domain.main", "version"),
				),
			},
			{
				ResourceName:      "aws_cognito_user_pool_domain.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolDomain_custom(t *testing.T) {
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	poolName := fmt.Sprintf("tf-acc-test-pool-%s", sdkacctest.RandString(10))

	acmCertificateResourceName := "aws_acm_certificate.test"
	cognitoUserPoolResourceName := "aws_cognito_user_pool.test"
	resourceName := "aws_cognito_user_pool_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckUserPoolCustomDomain(t) },
		ErrorCheck:               acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolDomainConfig_custom(rootDomain, domain, poolName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolDomainExists(resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, "aws_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_arn", acmCertificateResourceName, "arn"),
					//lintignore:AWSAT001 // Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11666
					resource.TestMatchResourceAttr(resourceName, "cloudfront_distribution_arn", regexp.MustCompile(`[a-z0-9]+.cloudfront.net$`)),
					resource.TestCheckResourceAttrPair(resourceName, "domain", acmCertificateResourceName, "domain_name"),
					resource.TestMatchResourceAttr(resourceName, "s3_bucket", regexp.MustCompile(`^.+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_id", cognitoUserPoolResourceName, "id"),
					resource.TestMatchResourceAttr(resourceName, "version", regexp.MustCompile(`^.+$`)),
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

func TestAccCognitoIDPUserPoolDomain_customCertUpdate(t *testing.T) {
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	poolName := fmt.Sprintf("tf-acc-test-pool-%s", sdkacctest.RandString(10))

	acmInitialValidationResourceName := "aws_acm_certificate_validation.initial_test"
	acmUpdatedValidationResourceName := "aws_acm_certificate_validation.updated_test"

	acmInitialCertResourceName := "aws_acm_certificate.initial"
	acmUpdatedCertResourceName := "aws_acm_certificate.updated"

	cognitoPoolResourceName := "aws_cognito_user_pool_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckUserPoolCustomDomain(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolDomainConfig_customCertUpdate(rootDomain, domain, poolName, acmInitialValidationResourceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolDomainExists(cognitoPoolResourceName),
					testAccCheckUserPoolDomainCertMatches(cognitoPoolResourceName, acmInitialCertResourceName),
					resource.TestCheckResourceAttrPair(cognitoPoolResourceName, "certificate_arn", acmInitialCertResourceName, "arn"),
				),
			},
			{
				Config: testAccUserPoolDomainConfig_customCertUpdate(rootDomain, domain, poolName, acmUpdatedValidationResourceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolDomainCertMatches(cognitoPoolResourceName, acmUpdatedCertResourceName),
					resource.TestCheckResourceAttrPair(cognitoPoolResourceName, "certificate_arn", acmUpdatedCertResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPoolDomain_disappears(t *testing.T) {
	domainName := fmt.Sprintf("tf-acc-test-domain-%d", sdkacctest.RandInt())
	poolName := fmt.Sprintf("tf-acc-test-pool-%s", sdkacctest.RandString(10))
	resourceName := "aws_cognito_user_pool_domain.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:               acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolDomainConfig_basic(domainName, poolName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolDomainExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcognitoidp.ResourceUserPoolDomain(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckUserPoolDomainExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Cognito User Pool Domain ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPConn

		_, err := conn.DescribeUserPoolDomain(&cognitoidentityprovider.DescribeUserPoolDomainInput{
			Domain: aws.String(rs.Primary.ID),
		})

		return err
	}
}

func testAccCheckUserPoolDomainCertMatches(cognitoResourceName string, certResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		cognitoResource, ok := s.RootModule().Resources[cognitoResourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", cognitoResourceName)
		}

		if cognitoResource.Primary.ID == "" {
			return errors.New("No Cognito User Pool Domain ID is set")
		}

		certResource, ok := s.RootModule().Resources[certResourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", cognitoResourceName)
		}

		if certResource.Primary.ID == "" {
			return errors.New("No ACM Certificate ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPConn

		domain, err := conn.DescribeUserPoolDomain(&cognitoidentityprovider.DescribeUserPoolDomainInput{
			Domain: aws.String(cognitoResource.Primary.ID),
		})

		if err != nil {
			return err
		}
		desc := domain.DomainDescription

		if desc.CustomDomainConfig == nil {
			return fmt.Errorf("No Custom Domain set on User pool: %s", *desc.UserPoolId)
		}

		if *desc.CustomDomainConfig.CertificateArn != certResource.Primary.ID {
			return fmt.Errorf("Certificate ARN on Custom Domain does not match, expected: %s, got: %s", certResource.Primary.ID, *desc.CustomDomainConfig.CertificateArn)
		}

		return nil
	}
}

func testAccCheckUserPoolDomainDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cognito_user_pool_domain" {
			continue
		}

		_, err := conn.DescribeUserPoolDomain(&cognitoidentityprovider.DescribeUserPoolDomainInput{
			Domain: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeResourceNotFoundException) {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccUserPoolDomainConfig_basic(domainName, poolName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool_domain" "main" {
  domain       = "%s"
  user_pool_id = aws_cognito_user_pool.main.id
}

resource "aws_cognito_user_pool" "main" {
  name = "%s"
}
`, domainName, poolName)
}

func testAccUserPoolDomainConfig_custom(rootDomain string, domain string, poolName string) string {
	return acctest.ConfigCompose(
		testAccUserPoolCustomDomainRegionProviderConfig(),
		fmt.Sprintf(`
data "aws_route53_zone" "test" {
  name         = %[1]q
  private_zone = false
}

resource "aws_acm_certificate" "test" {
  domain_name       = %[2]q
  validation_method = "DNS"
}

#
# for_each acceptance testing requires:
# https://github.com/hashicorp/terraform-plugin-sdk/issues/536
#
# resource "aws_route53_record" "test" {
#   for_each = {
#     for dvo in aws_acm_certificate.test.domain_validation_options: dvo.domain_name => {
#       name   = dvo.resource_record_name
#       record = dvo.resource_record_value
#       type   = dvo.resource_record_type
#     }
#   }

#   allow_overwrite = true
#   name            = each.value.name
#   records         = [each.value.record]
#   ttl             = 60
#   type            = each.value.type
#   zone_id         = data.aws_route53_zone.test.zone_id
# }

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

resource "aws_cognito_user_pool" "test" {
  name = %[3]q
}

resource "aws_cognito_user_pool_domain" "test" {
  certificate_arn = aws_acm_certificate_validation.test.certificate_arn
  domain          = aws_acm_certificate.test.domain_name
  user_pool_id    = aws_cognito_user_pool.test.id
}
`, rootDomain, domain, poolName))
}

func testAccUserPoolDomainConfig_customCertUpdate(rootDomain string, domain string, poolName string, appliedCertValidation string) string {
	return acctest.ConfigCompose(
		testAccUserPoolCustomDomainRegionProviderConfig(),
		fmt.Sprintf(`
data "aws_route53_zone" "test" {
  name         = %[1]q
  private_zone = false
}

resource "aws_acm_certificate" "initial" {
  domain_name       = %[2]q
  validation_method = "DNS"
}

resource "aws_acm_certificate" "updated" {
  domain_name       = %[2]q
  validation_method = "DNS"
}

resource "aws_route53_record" "initial_test" {
  allow_overwrite = true
  name            = tolist(aws_acm_certificate.initial.domain_validation_options)[0].resource_record_name
  records         = [tolist(aws_acm_certificate.initial.domain_validation_options)[0].resource_record_value]
  ttl             = 60
  type            = tolist(aws_acm_certificate.initial.domain_validation_options)[0].resource_record_type
  zone_id         = data.aws_route53_zone.test.zone_id
}

resource "aws_route53_record" "updated_test" {
  allow_overwrite = true
  name            = tolist(aws_acm_certificate.updated.domain_validation_options)[0].resource_record_name
  records         = [tolist(aws_acm_certificate.updated.domain_validation_options)[0].resource_record_value]
  ttl             = 60
  type            = tolist(aws_acm_certificate.updated.domain_validation_options)[0].resource_record_type
  zone_id         = data.aws_route53_zone.test.zone_id
}

resource "aws_acm_certificate_validation" "initial_test" {
  certificate_arn         = aws_acm_certificate.initial.arn
  validation_record_fqdns = [aws_route53_record.initial_test.fqdn]
}

resource "aws_acm_certificate_validation" "updated_test" {
  certificate_arn         = aws_acm_certificate.updated.arn
  validation_record_fqdns = [aws_route53_record.updated_test.fqdn]
}

resource "aws_cognito_user_pool" "test" {
  name = %[3]q
}

resource "aws_cognito_user_pool_domain" "test" {
  certificate_arn = %[4]s.certificate_arn
  domain          = aws_acm_certificate.initial.domain_name
  user_pool_id    = aws_cognito_user_pool.test.id
}
`, rootDomain, domain, poolName, appliedCertValidation))
}
