package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSApiGatewayStage_importBasic(t *testing.T) {
	resourceName := "tf-acc-test.dev"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayStageDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSAPIGatewayStageConfig_basic(),
			},
			resource.TestStep{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSAPIGatewayStageDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigateway

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_stage" {
			continue
		}

		req := &apigateway.GetStageInput{
			RestApiId: aws.String(s.RootModule().Resources["aws_api_gateway_rest_api.test"].Primary.ID),
			StageName: aws.String(rs.Primary.Attributes["stage_name"]),
		}
		out, err := conn.GetStage(req)
		if err == nil {
			return fmt.Errorf("API Gateway Stage still exists: %s", out)
		}

		awsErr, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if awsErr.Code() != "NotFoundException" {
			return err
		}

		return nil
	}

	return nil
}

const testAccAWSAPIGatewayStageConfig_base = `
resource "aws_api_gateway_rest_api" "test" {
  name = "tf-acc-test"
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
  uri = "https://www.google.co.uk"
  integration_http_method = "GET"
}

resource "aws_api_gateway_integration_response" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_resource.test.id}"
  http_method = "${aws_api_gateway_integration.test.http_method}"
  status_code = "${aws_api_gateway_method_response.error.status_code}"
}

resource "aws_api_gateway_deployment" "dev" {
  depends_on = ["aws_api_gateway_integration.test"]

  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  stage_name = "dev"
  description = "This is a dev env"

  variables = {
    "a" = "2"
  }
}
`

func testAccAWSAPIGatewayStageConfig_basic() string {
	return testAccAWSAPIGatewayStageConfig_base + `
resource "aws_api_gateway_stage" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  stage_name = "prod"
  deployment_id = "${aws_api_gateway_deployment.dev.id}"
  cache_cluster_enabled = true
  cache_cluster_size = "0.5"
  variables {
    one = "1"
    two = "2"
  }
}
`
}
