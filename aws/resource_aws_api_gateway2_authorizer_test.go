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

func TestAccAWSAPIGateway2Authorizer_basic(t *testing.T) {
	resourceName := "aws_api_gateway_v2_authorizer.test"
	lambdaResourceName := "aws_lambda_function.test"
	rName := fmt.Sprintf("tf-testacc-apigwv2-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGateway2AuthorizerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGateway2AuthorizerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2AuthorizerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "authorizer_credentials_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "authorizer_type", "REQUEST"),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.645907014", "route.request.header.Auth"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGateway2AuthorizerImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGateway2Authorizer_Credentials(t *testing.T) {
	resourceName := "aws_api_gateway_v2_authorizer.test"
	iamRoleResourceName := "aws_iam_role.test"
	lambdaResourceName := "aws_lambda_function.test"
	rName := fmt.Sprintf("tf-testacc-apigwv2-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGateway2AuthorizerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGateway2AuthorizerConfig_credentials(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2AuthorizerExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_credentials_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_type", "REQUEST"),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.645907014", "route.request.header.Auth"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGateway2AuthorizerImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGateway2AuthorizerConfig_credentialsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2AuthorizerExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_credentials_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_type", "REQUEST"),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.645907014", "route.request.header.Auth"),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.4138478046", "route.request.querystring.Name"),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("%s-updated", rName)),
				),
			},
			{
				Config: testAccAWSAPIGateway2AuthorizerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2AuthorizerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "authorizer_credentials_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "authorizer_type", "REQUEST"),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.645907014", "route.request.header.Auth"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
		},
	})
}

func testAccCheckAWSAPIGateway2AuthorizerDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_v2_authorizer" {
			continue
		}

		_, err := conn.GetAuthorizer(&apigatewayv2.GetAuthorizerInput{
			ApiId:        aws.String(rs.Primary.Attributes["api_id"]),
			AuthorizerId: aws.String(rs.Primary.ID),
		})
		if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("API Gateway v2 authorizer %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSAPIGateway2AuthorizerExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 authorizer ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		_, err := conn.GetAuthorizer(&apigatewayv2.GetAuthorizerInput{
			ApiId:        aws.String(rs.Primary.Attributes["api_id"]),
			AuthorizerId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAWSAPIGateway2AuthorizerImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["api_id"], rs.Primary.ID), nil
	}
}

func testAccAWSAPIGateway2AuthorizerConfig_base(rName string) string {
	return baseAccAWSLambdaConfig(rName, rName, rName) + fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = "${aws_iam_role.iam_for_lambda.arn}"
  handler       = "index.handler"
  runtime       = "nodejs10.x"
}

resource "aws_api_gateway_v2_api" "test" {
  name                       = %[1]q
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"
}

resource "aws_iam_role" "test" {
  name = "%[1]s_auth_invocation_role"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {"Service": "apigateway.amazonaws.com"},
    "Effect": "Allow"
  }]
}
EOF
}
`, rName)
}

func testAccAWSAPIGateway2AuthorizerConfig_basic(rName string) string {
	return testAccAWSAPIGateway2AuthorizerConfig_base(rName) + fmt.Sprintf(`
resource "aws_api_gateway_v2_authorizer" "test" {
  api_id           = "${aws_api_gateway_v2_api.test.id}"
  authorizer_type  = "REQUEST"
  authorizer_uri   = "${aws_lambda_function.test.invoke_arn}"
  identity_sources = ["route.request.header.Auth"]
  name             = %[1]q
}
`, rName)
}

func testAccAWSAPIGateway2AuthorizerConfig_credentials(rName string) string {
	return testAccAWSAPIGateway2AuthorizerConfig_base(rName) + fmt.Sprintf(`
resource "aws_api_gateway_v2_authorizer" "test" {
  api_id           = "${aws_api_gateway_v2_api.test.id}"
  authorizer_type  = "REQUEST"
  authorizer_uri   = "${aws_lambda_function.test.invoke_arn}"
  identity_sources = ["route.request.header.Auth"]
  name             = %[1]q

  authorizer_credentials_arn = "${aws_iam_role.test.arn}"
}
`, rName)
}

func testAccAWSAPIGateway2AuthorizerConfig_credentialsUpdated(rName string) string {
	return testAccAWSAPIGateway2AuthorizerConfig_base(rName) + fmt.Sprintf(`
resource "aws_api_gateway_v2_authorizer" "test" {
  api_id           = "${aws_api_gateway_v2_api.test.id}"
  authorizer_type  = "REQUEST"
  authorizer_uri   = "${aws_lambda_function.test.invoke_arn}"
  identity_sources = ["route.request.header.Auth", "route.request.querystring.Name"]
  name             = "%[1]s-updated"

  authorizer_credentials_arn = "${aws_iam_role.test.arn}"
}
`, rName)
}
