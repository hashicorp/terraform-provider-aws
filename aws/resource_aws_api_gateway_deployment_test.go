package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSAPIGatewayDeployment_basic(t *testing.T) {
	var conf apigateway.Deployment

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDeploymentConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDeploymentExists("aws_api_gateway_deployment.test", &conf),
					resource.TestCheckResourceAttr(
						"aws_api_gateway_deployment.test", "stage_name", "test"),
					resource.TestCheckResourceAttr(
						"aws_api_gateway_deployment.test", "description", "This is a test"),
					resource.TestCheckResourceAttr(
						"aws_api_gateway_deployment.test", "tags.b", "z"),
					resource.TestCheckResourceAttr(
						"aws_api_gateway_deployment.test", "tags.c", "3"),
					resource.TestCheckResourceAttr(
						"aws_api_gateway_deployment.test", "variables.a", "2"),
					resource.TestCheckResourceAttr(
						"aws_api_gateway_deployment.test", "variables.b", "3"),
					resource.TestCheckResourceAttrSet(
						"aws_api_gateway_deployment.test", "created_date"),
				),
			},
			{
				Config: testAccAWSAPIGatewayDeploymentConfig_updated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDeploymentExists("aws_api_gateway_deployment.test", &conf),
					resource.TestCheckResourceAttr(
						"aws_api_gateway_deployment.test", "stage_name", "test"),
					resource.TestCheckResourceAttr(
						"aws_api_gateway_deployment.test", "description", "This is a test"),
					resource.TestCheckResourceAttr(
						"aws_api_gateway_deployment.test", "stage_description", "test stage description"),
					resource.TestCheckResourceAttr(
						"aws_api_gateway_deployment.test", "tags.b", "z"),
					resource.TestCheckResourceAttr(
						"aws_api_gateway_deployment.test", "tags.d", "4"),
					resource.TestCheckResourceAttr(
						"aws_api_gateway_deployment.test", "variables.a", "2"),
					resource.TestCheckResourceAttr(
						"aws_api_gateway_deployment.test", "variables.c", "4"),
					resource.TestCheckResourceAttrSet(
						"aws_api_gateway_deployment.test", "created_date"),
				),
			},
			{
				Config: testAccAWSAPIGatewayDeploymentConfig_empty(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDeploymentExists("aws_api_gateway_deployment.test", &conf),
					resource.TestCheckResourceAttr(
						"aws_api_gateway_deployment.test", "description", "This is a test"),
					resource.TestCheckResourceAttrSet(
						"aws_api_gateway_deployment.test", "created_date"),
				),
			},
		},
	})
}

func testAccCheckAWSAPIGatewayDeploymentExists(n string, res *apigateway.Deployment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway Deployment ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigateway

		req := &apigateway.GetDeploymentInput{
			DeploymentId: aws.String(rs.Primary.ID),
			RestApiId:    aws.String(s.RootModule().Resources["aws_api_gateway_rest_api.test"].Primary.ID),
		}
		describe, err := conn.GetDeployment(req)
		if err != nil {
			return err
		}

		if *describe.Id != rs.Primary.ID {
			return fmt.Errorf("APIGateway Deployment not found")
		}

		*res = *describe

		return nil
	}
}

func testAccCheckAWSAPIGatewayDeploymentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigateway

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_resource" {
			continue
		}

		req := &apigateway.GetDeploymentsInput{
			RestApiId: aws.String(s.RootModule().Resources["aws_api_gateway_rest_api.test"].Primary.ID),
		}
		describe, err := conn.GetDeployments(req)

		if err == nil {
			if len(describe.Items) != 0 &&
				*describe.Items[0].Id == rs.Primary.ID {
				return fmt.Errorf("API Gateway Deployment still exists")
			}
		}

		aws2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if aws2err.Code() != "NotFoundException" {
			return err
		}

		return nil
	}

	return nil
}

const testAccAWSAPIGatewayDeploymentConfig_base = `
resource "aws_api_gateway_rest_api" "test" {
  name = "test"
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  parent_id = "${aws_api_gateway_rest_api.test.root_resource_id}"
  path_part = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_resource.test.id}"
  http_method = "GET"
  authorization = "NONE"
}

resource "aws_api_gateway_method_response" "error" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_resource.test.id}"
  http_method = "${aws_api_gateway_method.test.http_method}"
  status_code = "400"
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_resource.test.id}"
  http_method = "${aws_api_gateway_method.test.http_method}"

  type = "HTTP"
  uri = "https://www.google.de"
  integration_http_method = "GET"
}

resource "aws_api_gateway_integration_response" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_resource.test.id}"
  http_method = "${aws_api_gateway_integration.test.http_method}"
  status_code = "${aws_api_gateway_method_response.error.status_code}"
}
`

func testAccAWSAPIGatewayDeploymentConfig_basic() string {
	return testAccAWSAPIGatewayDeploymentConfig_base + `
resource "aws_api_gateway_deployment" "test" {
  depends_on = ["aws_api_gateway_integration.test"]

  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  stage_name = "test"
  description = "This is a test"

  tags = {
    "b" = "z"
    "c" = "3"
  }

  variables = {
    "a" = "2"
    "b" = "3"
  }
}
`
}

func testAccAWSAPIGatewayDeploymentConfig_updated() string {
	return testAccAWSAPIGatewayDeploymentConfig_base + `
resource "aws_api_gateway_deployment" "test" {
  depends_on = ["aws_api_gateway_integration.test"]

  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  stage_name = "test"
  description = "This is a test"
  stage_description = "test stage description"

  tags = {
    "b" = "z"
    "d" = "4"
  }

  variables = {
    "a" = "2"
    "c" = "4"
  }
}
`
}

func testAccAWSAPIGatewayDeploymentConfig_empty() string {
	return testAccAWSAPIGatewayDeploymentConfig_base + `
resource "aws_api_gateway_deployment" "test" {
  depends_on = ["aws_api_gateway_integration.test"]

  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  description = "This is a test"
}
`
}
