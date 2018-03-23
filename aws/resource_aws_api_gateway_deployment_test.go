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
				Config: testAccAWSAPIGatewayDeploymentConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDeploymentExists("aws_api_gateway_deployment.test", &conf),
					resource.TestCheckResourceAttr(
						"aws_api_gateway_deployment.test", "stage_name", "test"),
					resource.TestCheckResourceAttr(
						"aws_api_gateway_deployment.test", "description", "This is a test"),
					resource.TestCheckResourceAttr(
						"aws_api_gateway_deployment.test", "variables.a", "2"),
					resource.TestCheckResourceAttrSet(
						"aws_api_gateway_deployment.test", "created_date"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayDeployment_createBeforeDestoryUpdate(t *testing.T) {
	var conf apigateway.Deployment
	var stage apigateway.Stage

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDeploymentCreateBeforeDestroyConfig("description1", "https://www.google.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDeploymentExists("aws_api_gateway_deployment.test", &conf),
					resource.TestCheckResourceAttr(
						"aws_api_gateway_deployment.test", "stage_name", "test"),
					resource.TestCheckResourceAttr(
						"aws_api_gateway_deployment.test", "description", "description1"),
					resource.TestCheckResourceAttr(
						"aws_api_gateway_deployment.test", "variables.a", "2"),
					resource.TestCheckResourceAttrSet(
						"aws_api_gateway_deployment.test", "created_date"),
					testAccCheckAWSAPIGatewayDeploymentStageExists("test", &stage),
				),
			},
			{
				Config: testAccAWSAPIGatewayDeploymentCreateBeforeDestroyConfig("description2", "https://www.google.de"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDeploymentExists("aws_api_gateway_deployment.test", &conf),
					resource.TestCheckResourceAttr(
						"aws_api_gateway_deployment.test", "stage_name", "test"),
					resource.TestCheckResourceAttr(
						"aws_api_gateway_deployment.test", "description", "description2"),
					resource.TestCheckResourceAttr(
						"aws_api_gateway_deployment.test", "variables.a", "2"),
					resource.TestCheckResourceAttrSet(
						"aws_api_gateway_deployment.test", "created_date"),
					testAccCheckAWSAPIGatewayDeploymentStageExists("test", &stage),
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

func testAccCheckAWSAPIGatewayDeploymentStageExists(stageName string, res *apigateway.Stage) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).apigateway

		restApiId := aws.String(s.RootModule().Resources["aws_api_gateway_rest_api.test"].Primary.ID)

		req := &apigateway.GetStageInput{
			StageName: &stageName,
			RestApiId: restApiId,
		}
		stage, err := conn.GetStage(req)
		if err != nil {
			return err
		}

		*res = *stage

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

func buildAPIGatewayDeploymentConfig(description, url, extras string) string {
	return fmt.Sprintf(`
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
  uri = "%s"
  integration_http_method = "GET"
}

resource "aws_api_gateway_integration_response" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_resource.test.id}"
  http_method = "${aws_api_gateway_integration.test.http_method}"
  status_code = "${aws_api_gateway_method_response.error.status_code}"
}

resource "aws_api_gateway_deployment" "test" {
  depends_on = ["aws_api_gateway_integration.test"]

  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  stage_name = "test"
	description = "%s"
	stage_description = "%s"

	%s

  variables = {
    "a" = "2"
  }
}
`, url, description, description, extras)
}

var testAccAWSAPIGatewayDeploymentConfig = buildAPIGatewayDeploymentConfig("This is a test", "https://www.google.de", "")

func testAccAWSAPIGatewayDeploymentCreateBeforeDestroyConfig(description string, url string) string {
	return buildAPIGatewayDeploymentConfig(description, url, `
		lifecycle {
			create_before_destroy = true
		}
	`)
}
