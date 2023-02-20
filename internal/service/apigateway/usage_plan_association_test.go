package apigateway_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccAPIGatewayUsagePlanAssociation_basic(t *testing.T) {
	var conf apigateway.ApiStage
	resourceName := "aws_api_gateway_usage_plan_assocation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUsagePlanAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUsagePlanAssociationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanAssociationExists(resourceName, &conf),
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

func testAccCheckUsagePlanAssociationExists(n string, res *apigateway.ApiStage) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway Usage Plan Stage ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

		spl := strings.Split(rs.Primary.ID, "/")
		req := &apigateway.GetUsagePlanInput{
			UsagePlanId: aws.String(spl[0]),
		}

		up, err := conn.GetUsagePlan(req)
		if err != nil {
			return err
		}

		for _, stage := range up.ApiStages {
			if aws.StringValue(stage.ApiId) == spl[1] && aws.StringValue(stage.Stage) == spl[2] {
				*res = *stage
				return nil
			}
		}

		return fmt.Errorf("APIGateway Usage Plan Stage not found")
	}
}

func testAccCheckUsagePlanAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_usage_plan_assocation" {
			continue
		}

		spl := strings.Split(rs.Primary.ID, "/")
		req := &apigateway.GetUsagePlanInput{
			UsagePlanId: aws.String(spl[0]),
		}

		up, err := conn.GetUsagePlan(req)
		if err != nil {
			return err
		}

		for _, stage := range up.ApiStages {
			if aws.StringValue(stage.ApiId) == spl[1] && aws.StringValue(stage.Stage) == spl[2] {
				return fmt.Errorf("API Gateway Usage Plan Stage still exists")
			}
		}

		return nil
	}

	return nil
}

func testAccUsagePlanAssociationConfig() string {
	return `
resource "aws_api_gateway_rest_api" "test" {
	name = "tf-acc-test-usage-plan-stage"

	endpoint_configuration {
		types = ["REGIONAL"]
	}
}

resource "aws_api_gateway_resource" "test" {
	rest_api_id = aws_api_gateway_rest_api.test.id
	parent_id   = aws_api_gateway_rest_api.test.root_resource_id
	path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
	rest_api_id   = aws_api_gateway_rest_api.test.id
	resource_id   = aws_api_gateway_resource.test.id
	http_method   = "GET"
	authorization = "NONE"
}

resource "aws_api_gateway_stage" "test" {
	deployment_id = aws_api_gateway_deployment.test.id
	rest_api_id   = aws_api_gateway_rest_api.test.id
	stage_name    = "test"
  
}

resource "aws_api_gateway_integration" "test" {
	http_method             = aws_api_gateway_method.test.http_method
	resource_id             = aws_api_gateway_resource.test.id
	rest_api_id             = aws_api_gateway_rest_api.test.id
	type                    = "MOCK"
}

resource "aws_api_gateway_deployment" "test" {
	depends_on = [aws_api_gateway_integration.test]

	rest_api_id = aws_api_gateway_rest_api.test.id
	lifecycle {
	  create_before_destroy = true
	}
}

resource "aws_api_gateway_usage_plan" "test" {
	name = "tf-acc-test-usage-plan"
	lifecycle {
		ignore_changes = [api_stages]
	}
}
`
}

func testAccUsagePlanAssociationConfig_basic() string {
	return testAccUsagePlanAssociationConfig() + `
resource "aws_api_gateway_usage_plan_assocation" "test" {
  usage_plan_id = aws_api_gateway_usage_plan.test.id
  api_id = aws_api_gateway_rest_api.test.id
  stage = aws_api_gateway_stage.test.stage_name
}`
}
