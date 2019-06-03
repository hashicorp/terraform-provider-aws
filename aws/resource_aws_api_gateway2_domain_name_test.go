package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("aws_api_gateway_v2_domain_name", &resource.Sweeper{
		Name: "aws_api_gateway_v2_domain_name",
		F:    testSweepAPIGateway2DomainNames,
	})
}

func testSweepAPIGateway2DomainNames(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).apigatewayv2conn
	input := &apigatewayv2.GetDomainNamesInput{}

	for {
		output, err := conn.GetDomainNames(input)
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping API Gateway v2 domain names sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving API Gateway v2 domain names: %s", err)
		}

		for _, domainName := range output.Items {
			log.Printf("[INFO] Deleting API Gateway v2 domain name: %s", aws.StringValue(domainName.DomainName))
			_, err := conn.DeleteDomainName(&apigatewayv2.DeleteDomainNameInput{
				DomainName: domainName.DomainName,
			})
			if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
				continue
			}
			if err != nil {
				return fmt.Errorf("error deleting API Gateway v2 domain name (%s): %s", aws.StringValue(domainName.DomainName), err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	return nil
}

func TestAccAWSAPIGateway2DomainName_basic(t *testing.T) {
	resourceName := "aws_api_gateway_v2_domain_name.test"
	certResourceName := "aws_acm_certificate.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, fmt.Sprintf("%s.example.com", rName))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGateway2DomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGateway2DomainNameConfig_basic(rName, certificate, key, 1, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2DomainNameExists(resourceName),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/domainnames/.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain_name", fmt.Sprintf("%s.example.com", rName)),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", certResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.endpoint_type", "REGIONAL"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.security_policy", "TLS_1_2"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.target_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSAPIGateway2DomainName_Tags(t *testing.T) {
	resourceName := "aws_api_gateway_v2_domain_name.test"
	certResourceName := "aws_acm_certificate.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, fmt.Sprintf("%s.example.com", rName))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGateway2DomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGateway2DomainNameConfig_tags(rName, certificate, key, 1, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2DomainNameExists(resourceName),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/domainnames/.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain_name", fmt.Sprintf("%s.example.com", rName)),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", certResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.endpoint_type", "REGIONAL"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.security_policy", "TLS_1_2"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.target_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGateway2DomainNameConfig_basic(rName, certificate, key, 1, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2DomainNameExists(resourceName),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/domainnames/.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain_name", fmt.Sprintf("%s.example.com", rName)),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", certResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.endpoint_type", "REGIONAL"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.security_policy", "TLS_1_2"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.target_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSAPIGateway2DomainName_UpdateCertificate(t *testing.T) {
	resourceName := "aws_api_gateway_v2_domain_name.test"
	certResourceName0 := "aws_acm_certificate.test.0"
	certResourceName1 := "aws_acm_certificate.test.1"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, fmt.Sprintf("%s.example.com", rName))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGateway2DomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGateway2DomainNameConfig_basic(rName, certificate, key, 2, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2DomainNameExists(resourceName),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/domainnames/.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain_name", fmt.Sprintf("%s.example.com", rName)),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", certResourceName0, "arn"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.endpoint_type", "REGIONAL"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.security_policy", "TLS_1_2"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.target_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAWSAPIGateway2DomainNameConfig_basic(rName, certificate, key, 2, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2DomainNameExists(resourceName),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/domainnames/.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain_name", fmt.Sprintf("%s.example.com", rName)),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", certResourceName1, "arn"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.endpoint_type", "REGIONAL"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.security_policy", "TLS_1_2"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.target_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAWSAPIGateway2DomainNameConfig_tags(rName, certificate, key, 2, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2DomainNameExists(resourceName),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/domainnames/.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain_name", fmt.Sprintf("%s.example.com", rName)),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", certResourceName0, "arn"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.endpoint_type", "REGIONAL"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.security_policy", "TLS_1_2"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.target_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2"),
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

func testAccCheckAWSAPIGateway2DomainNameDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_v2_domain_name" {
			continue
		}

		_, err := conn.GetDomainName(&apigatewayv2.GetDomainNameInput{
			DomainName: aws.String(rs.Primary.ID),
		})
		if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("API Gateway v2 domain name %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSAPIGateway2DomainNameExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 domain name ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		_, err := conn.GetDomainName(&apigatewayv2.GetDomainNameInput{
			DomainName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAWSAPIGateway2DomainNameConfig_base(rName, certificate, key string, count int) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  count = %[4]d

  certificate_body = %[2]q
  private_key      = %[3]q

  tags = {
    Name = %[1]q
  }
}
`, rName, certificate, key, count)
}

func testAccAWSAPIGateway2DomainNameConfig_basic(rName, certificate, key string, count, index int) string {
	return testAccAWSAPIGateway2DomainNameConfig_base(rName, certificate, key, count) + fmt.Sprintf(`
resource "aws_api_gateway_v2_domain_name" "test" {
  domain_name = "%[1]s.example.com"

  domain_name_configuration {
    certificate_arn = "${aws_acm_certificate.test.*.arn[%[2]d]}"
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }
}
`, rName, index)
}

func testAccAWSAPIGateway2DomainNameConfig_tags(rName, certificate, key string, count, index int) string {
	return testAccAWSAPIGateway2DomainNameConfig_base(rName, certificate, key, count) + fmt.Sprintf(`
resource "aws_api_gateway_v2_domain_name" "test" {
  domain_name = "%[1]s.example.com"

  domain_name_configuration {
    certificate_arn = "${aws_acm_certificate.test.*.arn[%[2]d]}"
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }

  tags = {
    Key1 = "Value1"
    Key2 = "Value2"
  }
}
`, rName, index)
}
