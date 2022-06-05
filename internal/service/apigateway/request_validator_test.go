package apigateway_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigateway"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
)

func TestAccAPIGatewayRequestValidator_basic(t *testing.T) {
	var conf apigateway.UpdateRequestValidatorOutput
	rName := fmt.Sprintf("tf-test-acc-%s", sdkacctest.RandString(8))
	resourceName := "aws_api_gateway_request_validator.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRequestValidatorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRequestValidatorConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRequestValidatorExists(resourceName, &conf),
					testAccCheckRequestValidatorName(&conf, "tf-acc-test-request-validator"),
					resource.TestCheckResourceAttr(resourceName, "name", "tf-acc-test-request-validator"),
					testAccCheckRequestValidatorValidateRequestBody(&conf, false),
					resource.TestCheckResourceAttr(resourceName, "validate_request_body", "false"),
					testAccCheckRequestValidatorValidateRequestParameters(&conf, false),
					resource.TestCheckResourceAttr(resourceName, "validate_request_parameters", "false"),
				),
			},
			{
				Config: testAccRequestValidatorUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRequestValidatorExists(resourceName, &conf),
					testAccCheckRequestValidatorName(&conf, "tf-acc-test-request-validator_modified"),
					resource.TestCheckResourceAttr(resourceName, "name", "tf-acc-test-request-validator_modified"),
					testAccCheckRequestValidatorValidateRequestBody(&conf, true),
					resource.TestCheckResourceAttr(resourceName, "validate_request_body", "true"),
					testAccCheckRequestValidatorValidateRequestParameters(&conf, true),
					resource.TestCheckResourceAttr(resourceName, "validate_request_parameters", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRequestValidatorImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayRequestValidator_disappears(t *testing.T) {
	var conf apigateway.UpdateRequestValidatorOutput
	rName := fmt.Sprintf("tf-test-acc-%s", sdkacctest.RandString(8))
	resourceName := "aws_api_gateway_request_validator.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRequestValidatorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRequestValidatorConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRequestValidatorExists(resourceName, &conf),
					acctest.CheckResourceDisappears(acctest.Provider, tfapigateway.ResourceRequestValidator(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRequestValidatorName(conf *apigateway.UpdateRequestValidatorOutput, expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if conf.Name == nil {
			return fmt.Errorf("Empty Name, expected: %q", expectedName)
		}
		if *conf.Name != expectedName {
			return fmt.Errorf("Name didn't match. Expected: %q, Given: %q", expectedName, *conf.Name)
		}
		return nil
	}
}

func testAccCheckRequestValidatorValidateRequestBody(conf *apigateway.UpdateRequestValidatorOutput, expectedValue bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if conf.ValidateRequestBody == nil {
			return fmt.Errorf("Empty ValidateRequestBody, expected: %t", expectedValue)
		}
		if *conf.ValidateRequestBody != expectedValue {
			return fmt.Errorf("ValidateRequestBody didn't match. Expected: %t, Given: %t", expectedValue, *conf.ValidateRequestBody)
		}
		return nil
	}
}

func testAccCheckRequestValidatorValidateRequestParameters(conf *apigateway.UpdateRequestValidatorOutput, expectedValue bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if conf.ValidateRequestParameters == nil {
			return fmt.Errorf("Empty ValidateRequestParameters, expected: %t", expectedValue)
		}
		if *conf.ValidateRequestParameters != expectedValue {
			return fmt.Errorf("ValidateRequestParameters didn't match. Expected: %t, Given: %t", expectedValue, *conf.ValidateRequestParameters)
		}
		return nil
	}
}

func testAccCheckRequestValidatorExists(n string, res *apigateway.UpdateRequestValidatorOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Request Validator ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

		req := &apigateway.GetRequestValidatorInput{
			RequestValidatorId: aws.String(rs.Primary.ID),
			RestApiId:          aws.String(rs.Primary.Attributes["rest_api_id"]),
		}
		describe, err := conn.GetRequestValidator(req)
		if err != nil {
			return err
		}

		*res = *describe

		return nil
	}
}

func testAccCheckRequestValidatorDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_request_validator" {
			continue
		}

		req := &apigateway.GetRequestValidatorInput{
			RequestValidatorId: aws.String(rs.Primary.ID),
			RestApiId:          aws.String(rs.Primary.Attributes["rest_api_id"]),
		}
		_, err := conn.GetRequestValidator(req)

		if err == nil {
			return fmt.Errorf("API Request Validator still exists")
		}

		aws2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if aws2err.Code() != apigateway.ErrCodeNotFoundException {
			return err
		}

		return nil
	}

	return nil
}

func testAccRequestValidatorImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["rest_api_id"], rs.Primary.ID), nil
	}
}

func testAccRequestValidatorConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "%s"
}
`, rName)
}

func testAccRequestValidatorConfig(rName string) string {
	return fmt.Sprintf(testAccRequestValidatorConfig_base(rName) + `
resource "aws_api_gateway_request_validator" "test" {
  name        = "tf-acc-test-request-validator"
  rest_api_id = aws_api_gateway_rest_api.test.id
}
`)
}

func testAccRequestValidatorUpdatedConfig(rName string) string {
	return fmt.Sprintf(testAccRequestValidatorConfig_base(rName) + `
resource "aws_api_gateway_request_validator" "test" {
  name                        = "tf-acc-test-request-validator_modified"
  rest_api_id                 = aws_api_gateway_rest_api.test.id
  validate_request_body       = true
  validate_request_parameters = true
}
`)
}
