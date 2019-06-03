package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSAPIGateway2ApiMapping_basic(t *testing.T) {
	resourceName := "aws_api_gateway_v2_api_mapping.test"
	domainNameResourceName := "aws_api_gateway_v2_domain_name.test"
	stageResourceName := "aws_api_gateway_v2_stage.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, fmt.Sprintf("%s.example.com", rName))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGateway2ApiMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGateway2ApiMappingConfig_basic(rName, certificate, key),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2ApiMappingExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", domainNameResourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "stage", stageResourceName, "name")),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGateway2ApiMappingImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGateway2ApiMapping_ApiMappingKey(t *testing.T) {
	resourceName := "aws_api_gateway_v2_api_mapping.test"
	domainNameResourceName := "aws_api_gateway_v2_domain_name.test"
	stageResourceName := "aws_api_gateway_v2_stage.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, fmt.Sprintf("%s.example.com", rName))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGateway2ApiMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGateway2ApiMappingConfig_apiMappingKey(rName, certificate, key, "$context.domainName"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2ApiMappingExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "api_mapping_key", "$context.domainName"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", domainNameResourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "stage", stageResourceName, "name")),
			},
			{
				Config: testAccAWSAPIGateway2ApiMappingConfig_apiMappingKey(rName, certificate, key, "$context.apiId"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2ApiMappingExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "api_mapping_key", "$context.apiId"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", domainNameResourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "stage", stageResourceName, "name")),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGateway2ApiMappingImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSAPIGateway2ApiMappingDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_v2_api_mapping" {
			continue
		}

		_, err := conn.GetApiMapping(&apigatewayv2.GetApiMappingInput{
			ApiMappingId: aws.String(rs.Primary.ID),
			DomainName:   aws.String(rs.Primary.Attributes["domain_name"]),
		})
		if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("API Gateway v2 API mapping %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSAPIGateway2ApiMappingExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 API mapping ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		_, err := conn.GetApiMapping(&apigatewayv2.GetApiMappingInput{
			ApiMappingId: aws.String(rs.Primary.ID),
			DomainName:   aws.String(rs.Primary.Attributes["domain_name"]),
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAWSAPIGateway2ApiMappingImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.ID, rs.Primary.Attributes["domain_name"]), nil
	}
}

func testAccAWSAPIGateway2ApiMappingConfig_basic(rName, certificate, key string) string {
	return testAccAWSAPIGatewayV2DomainNameConfig_basic(rName, certificate, key, 1, 0) + testAccAWSAPIGatewayV2StageConfig_basic(rName) + fmt.Sprintf(`
resource "aws_api_gateway_v2_api_mapping" "test" {
  api_id      = "${aws_api_gateway_v2_api.test.id}"
  domain_name = "${aws_api_gateway_v2_domain_name.test.id}"
  stage       = "${aws_api_gateway_v2_stage.test.id}"
}
`)
}

func testAccAWSAPIGateway2ApiMappingConfig_apiMappingKey(rName, certificate, key, apiMappingKey string) string {
	return testAccAWSAPIGatewayV2DomainNameConfig_basic(rName, certificate, key, 1, 0) + testAccAWSAPIGatewayV2StageConfig_basic(rName) + fmt.Sprintf(`
resource "aws_api_gateway_v2_api_mapping" "test" {
  api_id      = "${aws_api_gateway_v2_api.test.id}"
  domain_name = "${aws_api_gateway_v2_domain_name.test.id}"
  stage       = "${aws_api_gateway_v2_stage.test.id}"

  api_mapping_key = %[1]q
}
`, apiMappingKey)
}
