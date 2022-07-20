package apigateway_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
)

func TestAccAPIGatewayDomainName_certificateARN(t *testing.T) {
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)

	var domainName apigateway.DomainName
	acmCertificateResourceName := "aws_acm_certificate.test"
	resourceName := "aws_api_gateway_domain_name.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckEdgeDomainName(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEdgeDomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_certificateARN(rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEdgeDomainNameExists(resourceName, &domainName),
					testAccCheckResourceAttrRegionalARNEdgeDomainName(resourceName, "arn", "apigateway", domain),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_arn", acmCertificateResourceName, "arn"),
					resource.TestMatchResourceAttr(resourceName, "cloudfront_domain_name", regexp.MustCompile(`[a-z0-9]+.cloudfront.net`)),
					resource.TestCheckResourceAttr(resourceName, "cloudfront_zone_id", "Z2FDTNDATAQYW2"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", acmCertificateResourceName, "domain_name"),
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

func TestAccAPIGatewayDomainName_certificateName(t *testing.T) {
	certificateBody := os.Getenv("AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_BODY")
	if certificateBody == "" {
		t.Skip(
			"Environment variable AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_BODY is not set. " +
				"This environment variable must be set to any non-empty value " +
				"with a publicly trusted certificate body to enable the test.")
	}

	certificateChain := os.Getenv("AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_CHAIN")
	if certificateChain == "" {
		t.Skip(
			"Environment variable AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_CHAIN is not set. " +
				"This environment variable must be set to any non-empty value " +
				"with a chain certificate acceptable for the certificate to enable the test.")
	}

	certificatePrivateKey := os.Getenv("AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_PRIVATE_KEY")
	if certificatePrivateKey == "" {
		t.Skip(
			"Environment variable AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_PRIVATE_KEY is not set. " +
				"This environment variable must be set to any non-empty value " +
				"with a private key of a publicly trusted certificate to enable the test.")
	}

	domainName := os.Getenv("AWS_API_GATEWAY_DOMAIN_NAME_DOMAIN_NAME")
	if domainName == "" {
		t.Skip(
			"Environment variable AWS_API_GATEWAY_DOMAIN_NAME_DOMAIN_NAME is not set. " +
				"This environment variable must be set to any non-empty value " +
				"with a domain name acceptable for the certificate to enable the test.")
	}

	var conf apigateway.DomainName
	resourceName := "aws_api_gateway_domain_name.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_certificate(domainName, certificatePrivateKey, certificateBody, certificateChain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/domainnames/+.`)),
					resource.TestCheckResourceAttr(resourceName, "certificate_name", "tf-acc-apigateway-domain-name"),
					resource.TestCheckResourceAttrSet(resourceName, "cloudfront_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "cloudfront_zone_id", "Z2FDTNDATAQYW2"),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domainName),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_upload_date"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"certificate_body", "certificate_chain", "certificate_private_key"},
			},
		},
	})
}

func TestAccAPIGatewayDomainName_regionalCertificateARN(t *testing.T) {
	var domainName apigateway.DomainName
	resourceName := "aws_api_gateway_domain_name.test"
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
				Config: testAccDomainNameConfig_regionalCertificateARN(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(resourceName, &domainName),
					testAccCheckResourceAttrRegionalARNRegionalDomainName(resourceName, "arn", "apigateway", rName),
					resource.TestCheckResourceAttr(resourceName, "domain_name", rName),
					acctest.MatchResourceAttrRegionalHostname(resourceName, "regional_domain_name", "execute-api", regexp.MustCompile(`d-[a-z0-9]+`)),
					resource.TestMatchResourceAttr(resourceName, "regional_zone_id", regexp.MustCompile(`^Z`)),
				),
			},
		},
	})
}

func TestAccAPIGatewayDomainName_regionalCertificateName(t *testing.T) {
	// For now, use an environment variable to limit running this test
	// BadRequestException: Uploading certificates is not supported for REGIONAL.
	// See Remarks section of https://docs.aws.amazon.com/apigateway/api-reference/link-relation/domainname-create/
	// which suggests this configuration should be possible somewhere, e.g. AWS China?
	regionalCertificateArn := os.Getenv("AWS_API_GATEWAY_DOMAIN_NAME_REGIONAL_CERTIFICATE_NAME_ENABLED")
	if regionalCertificateArn == "" {
		t.Skip(
			"Environment variable AWS_API_GATEWAY_DOMAIN_NAME_REGIONAL_CERTIFICATE_NAME_ENABLED is not set. " +
				"This environment variable must be set to any non-empty value " +
				"in a region where uploading REGIONAL certificates is allowed " +
				"to enable the test.")
	}

	var domainName apigateway.DomainName
	resourceName := "aws_api_gateway_domain_name.test"

	domain := acctest.RandomDomainName()
	domainWildcard := fmt.Sprintf("*.%s", domain)
	rName := fmt.Sprintf("%s.%s", sdkacctest.RandString(8), domain)

	caKey := acctest.TLSRSAPrivateKeyPEM(2048)
	caCertificate := acctest.TLSRSAX509SelfSignedCACertificatePEM(caKey)
	key := acctest.TLSRSAPrivateKeyPEM(2048)
	certificate := acctest.TLSRSAX509LocallySignedCertificatePEM(caKey, caCertificate, key, domainWildcard)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_regionalCertificate(rName, key, certificate, caCertificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(resourceName, &domainName),
					testAccCheckResourceAttrRegionalARNRegionalDomainName(resourceName, "arn", "apigateway", rName),
					resource.TestCheckResourceAttr(resourceName, "certificate_body", certificate),
					resource.TestCheckResourceAttr(resourceName, "certificate_chain", caCertificate),
					resource.TestCheckResourceAttr(resourceName, "certificate_name", "tf-acc-apigateway-domain-name"),
					resource.TestCheckResourceAttr(resourceName, "certificate_private_key", key),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_upload_date"),
					resource.TestCheckResourceAttr(resourceName, "domain_name", rName),
					acctest.MatchResourceAttrRegionalHostname(resourceName, "regional_domain_name", "execute-api", regexp.MustCompile(`d-[a-z0-9]+`)),
					resource.TestMatchResourceAttr(resourceName, "regional_zone_id", regexp.MustCompile(`^Z`)),
				),
			},
		},
	})
}

func TestAccAPIGatewayDomainName_securityPolicy(t *testing.T) {
	var domainName apigateway.DomainName
	resourceName := "aws_api_gateway_domain_name.test"
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
				Config: testAccDomainNameConfig_securityPolicy(rName, key, certificate, apigateway.SecurityPolicyTls12),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(resourceName, &domainName),
					resource.TestCheckResourceAttr(resourceName, "security_policy", apigateway.SecurityPolicyTls12),
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

func TestAccAPIGatewayDomainName_tags(t *testing.T) {
	var domainName apigateway.DomainName
	resourceName := "aws_api_gateway_domain_name.test"
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
				Config: testAccDomainNameConfig_tags1(rName, key, certificate, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(resourceName, &domainName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccDomainNameConfig_tags2(rName, key, certificate, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(resourceName, &domainName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDomainNameConfig_tags1(rName, key, certificate, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(resourceName, &domainName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
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

func TestAccAPIGatewayDomainName_disappears(t *testing.T) {
	var domainName apigateway.DomainName
	resourceName := "aws_api_gateway_domain_name.test"
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
				Config: testAccDomainNameConfig_regionalCertificateARN(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(resourceName, &domainName),
					acctest.CheckResourceDisappears(acctest.Provider, tfapigateway.ResourceDomainName(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayDomainName_MutualTLSAuthentication_basic(t *testing.T) {
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := fmt.Sprintf("%s.%s", acctest.RandomSubdomain(), rootDomain)

	var v apigateway.DomainName
	resourceName := "aws_api_gateway_domain_name.test"
	acmCertificateResourceName := "aws_acm_certificate.test"
	s3ObjectResourceName := "aws_s3_object.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_mutualTLSAuthentication(rName, rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/domainnames/+.`)),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", acmCertificateResourceName, "domain_name"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.0.truststore_uri", fmt.Sprintf("s3://%s/%s", rName, rName)),
					resource.TestCheckResourceAttrPair(resourceName, "mutual_tls_authentication.0.truststore_version", s3ObjectResourceName, "version_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Test disabling mutual TLS authentication.
			{
				Config: testAccDomainNameConfig_mutualTLSAuthenticationMissing(rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", acmCertificateResourceName, "domain_name"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "0"),
				),
			},
		},
	})
}

func TestAccAPIGatewayDomainName_MutualTLSAuthentication_ownership(t *testing.T) {
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := fmt.Sprintf("%s.%s", acctest.RandomSubdomain(), rootDomain)
	key := acctest.TLSRSAPrivateKeyPEM(2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(key, domain)

	var v apigateway.DomainName
	resourceName := "aws_api_gateway_domain_name.test"
	publicAcmCertificateResourceName := "aws_acm_certificate.test"
	s3ObjectResourceName := "aws_s3_object.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_mutualTLSOwnership(rName, rootDomain, domain, certificate, key),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/domainnames/+.`)),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", publicAcmCertificateResourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "ownership_verification_certificate_arn", publicAcmCertificateResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.0.truststore_uri", fmt.Sprintf("s3://%s/%s", rName, rName)),
					resource.TestCheckResourceAttrPair(resourceName, "mutual_tls_authentication.0.truststore_version", s3ObjectResourceName, "version_id"),
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

func testAccCheckDomainNameExists(n string, res *apigateway.DomainName) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway DomainName ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

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

func testAccCheckDomainNameDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_domain_name" {
			continue
		}

		_, err := conn.GetDomainName(&apigateway.GetDomainNameInput{
			DomainName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
				return nil
			}
			return err
		}

		return fmt.Errorf("API Gateway Domain Name still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccCheckEdgeDomainNameExists(resourceName string, domainName *apigateway.DomainName) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID not set")
		}

		conn := testAccProviderEdgeDomainName.Meta().(*conns.AWSClient).APIGatewayConn

		input := &apigateway.GetDomainNameInput{
			DomainName: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetDomainName(input)

		if err != nil {
			return fmt.Errorf("error reading API Gateway Domain Name (%s): %w", rs.Primary.ID, err)
		}

		*domainName = *output

		return nil
	}
}

func testAccCheckEdgeDomainNameDestroy(s *terraform.State) error {
	conn := testAccProviderEdgeDomainName.Meta().(*conns.AWSClient).APIGatewayConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_domain_name" {
			continue
		}

		input := &apigateway.GetDomainNameInput{
			DomainName: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetDomainName(input)

		if tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading API Gateway Domain Name (%s): %w", rs.Primary.ID, err)
		}

		if output != nil && aws.StringValue(output.DomainName) == rs.Primary.ID {
			return fmt.Errorf("API Gateway Domain Name (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccDomainNamePublicCertConfig(rootDomain, domain string) string {
	return fmt.Sprintf(`
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
`, rootDomain, domain)
}

func testAccDomainNameConfig_certificateARN(rootDomain string, domain string) string {
	return acctest.ConfigCompose(
		testAccEdgeDomainNameRegionProviderConfig(),
		testAccDomainNamePublicCertConfig(rootDomain, domain),
		`
resource "aws_api_gateway_domain_name" "test" {
  domain_name     = aws_acm_certificate.test.domain_name
  certificate_arn = aws_acm_certificate_validation.test.certificate_arn

  endpoint_configuration {
    types = ["EDGE"]
  }
}
`)
}

func testAccDomainNameConfig_certificate(domainName, key, certificate, chainCertificate string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_domain_name" "test" {
  domain_name             = "%[1]s"
  certificate_body        = "%[2]s"
  certificate_chain       = "%[3]s"
  certificate_name        = "tf-acc-apigateway-domain-name"
  certificate_private_key = "%[4]s"
}
`, domainName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(chainCertificate), acctest.TLSPEMEscapeNewlines(key))
}

func testAccDomainNameConfig_regionalCertificateARN(domainName, key, certificate string) string {
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
`, domainName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
}

func testAccDomainNameConfig_regionalCertificate(domainName, key, certificate, chainCertificate string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_domain_name" "test" {
  certificate_body          = "%[2]s"
  certificate_chain         = "%[3]s"
  certificate_private_key   = "%[4]s"
  domain_name               = "%[1]s"
  regional_certificate_name = "tf-acc-apigateway-domain-name"

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}
`, domainName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(chainCertificate), acctest.TLSPEMEscapeNewlines(key))
}

func testAccDomainNameConfig_securityPolicy(domainName, key, certificate, securityPolicy string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_api_gateway_domain_name" "test" {
  domain_name              = %[1]q
  regional_certificate_arn = aws_acm_certificate.test.arn
  security_policy          = %[4]q

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}
`, domainName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key), securityPolicy)
}

func testAccDomainNameConfig_tags1(domainName, key, certificate, tagKey1, tagValue1 string) string {
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

  tags = {
    %[4]q = %[5]q
  }
}
`, domainName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key), tagKey1, tagValue1)
}

func testAccDomainNameConfig_tags2(domainName, key, certificate, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

  tags = {
    %[4]q = %[5]q
    %[6]q = %[7]q
  }
}
`, domainName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key), tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccDomainNameConfig_mutualTLSAuthentication(rName, rootDomain, domain string) string {
	return acctest.ConfigCompose(
		testAccDomainNamePublicCertConfig(rootDomain, domain),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  force_destroy = true
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket_versioning.test.bucket
  key    = %[1]q
  source = "test-fixtures/apigateway-domain-name-truststore-1.pem"
}

resource "aws_api_gateway_domain_name" "test" {
  domain_name              = aws_acm_certificate.test.domain_name
  regional_certificate_arn = aws_acm_certificate_validation.test.certificate_arn
  security_policy          = "TLS_1_2"

  endpoint_configuration {
    types = ["REGIONAL"]
  }

  mutual_tls_authentication {
    truststore_uri     = "s3://${aws_s3_object.test.bucket}/${aws_s3_object.test.key}"
    truststore_version = aws_s3_object.test.version_id
  }
}
`, rName))
}

func testAccDomainNameConfig_mutualTLSAuthenticationMissing(rootDomain, domain string) string {
	return acctest.ConfigCompose(
		testAccDomainNamePublicCertConfig(rootDomain, domain),
		`
resource "aws_api_gateway_domain_name" "test" {
  domain_name              = aws_acm_certificate.test.domain_name
  regional_certificate_arn = aws_acm_certificate_validation.test.certificate_arn
  security_policy          = "TLS_1_2"

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}
`)
}

func testAccDomainNameConfig_mutualTLSOwnership(rName, rootDomain, domain, certificate, key string) string {
	return acctest.ConfigCompose(
		testAccDomainNamePublicCertConfig(rootDomain, domain),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  force_destroy = true
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket_versioning.test.bucket
  key    = %[1]q
  source = "test-fixtures/apigateway-domain-name-truststore-1.pem"
}

resource "aws_acm_certificate" "imported" {
  certificate_body = %[2]q
  private_key      = %[3]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_api_gateway_domain_name" "test" {
  domain_name                            = aws_acm_certificate.test.domain_name
  regional_certificate_arn               = aws_acm_certificate.imported.arn
  security_policy                        = "TLS_1_2"
  ownership_verification_certificate_arn = aws_acm_certificate_validation.test.certificate_arn

  endpoint_configuration {
    types = ["REGIONAL"]
  }

  mutual_tls_authentication {
    truststore_uri     = "s3://${aws_s3_object.test.bucket}/${aws_s3_object.test.key}"
    truststore_version = aws_s3_object.test.version_id
  }
}
`, rName, certificate, key))
}
