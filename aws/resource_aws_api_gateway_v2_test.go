package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSAPIGatewayV2_basic(t *testing.T) {
	var conf apigatewayv2.Api

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2Config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2Exists("aws_api_gateway_v2.test", &conf),
					resource.TestCheckResourceAttr("aws_api_gateway_v2.test", "description", "testing things"),
					resource.TestCheckResourceAttr("aws_api_gateway_v2.test", "name", "test"),
					resource.TestCheckResourceAttr("aws_api_gateway_v2.test", "protocol_type", "WEBSOCKET"),
					resource.TestCheckResourceAttr("aws_api_gateway_v2.test", "route_selection_expression", "$request.body.action"),
				),
			},
			{
				ResourceName:      "aws_api_gateway_v2.test",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSAPIGatewayV2ImportStateIdFunc("aws_api_gateway_v2.test"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayV2_update(t *testing.T) {
	var conf apigatewayv2.Api

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2Config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2Exists("aws_api_gateway_v2.test", &conf),
					resource.TestCheckResourceAttr("aws_api_gateway_v2.test", "name", "test"),
					resource.TestCheckResourceAttr("aws_api_gateway_v2.test", "description", "testing things"),
					resource.TestCheckResourceAttr("aws_api_gateway_v2.test", "protocol_type", "WEBSOCKET"),
					resource.TestCheckResourceAttr("aws_api_gateway_v2.test", "route_selection_expression", "$request.body.action"),
				),
			},
			{
				Config: testAccAWSAPIGatewayV2Config_updatePathPart,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2Exists("aws_api_gateway_v2.test", &conf),
					resource.TestCheckResourceAttr("aws_api_gateway_v2.test", "name", "new test name"),
					resource.TestCheckResourceAttr("aws_api_gateway_v2.test", "description", "new testing things"),
					resource.TestCheckResourceAttr("aws_api_gateway_v2.test", "protocol_type", "HTTP"),
					resource.TestCheckResourceAttr("aws_api_gateway_v2.test", "route_selection_expression", "$request.body.type"),
				),
			},
			{
				ResourceName:      "aws_api_gateway_v2.test",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSAPIGatewayV2ImportStateIdFunc("aws_api_gateway_v2.test"),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSAPIGatewayV2Exists(n string, res *apigatewayv2.Api) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway V2 ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		req := &apigatewayv2.GetApiInput{
			ApiId: aws.String(rs.Primary.ID),
		}
		api, err := conn.GetApi(req)
		if err != nil {
			return err
		}

		if *api.ApiId != rs.Primary.ID {
			return fmt.Errorf("APIGateway V2 not found")
		}

		//*res = *api

		return nil
	}
}

func testAccCheckAWSAPIGatewayV2Destroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_v2" {
			continue
		}

		req := &apigatewayv2.GetApisInput{}
		api, err := conn.GetApis(req)

		if err == nil {
			if len(api.Items) != 0 &&
				*api.Items[0].ApiId == rs.Primary.ID {
				return fmt.Errorf("API Gateway Api still exists")
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

func testAccAWSAPIGatewayV2ImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["api_id"], rs.Primary.ID), nil
	}
}

const testAccAWSAPIGatewayV2Config = `
resource "aws_api_gateway_v2" "test" {
	name                       = "test"
	description                = "testing things"
	protocol_type              = "WEBSOCKET"
	route_selection_expression = "$request.body.action"
}
`

const testAccAWSAPIGatewayV2Config_updatePathPart = `
resource "aws_api_gateway_v2" "test" {
	name                       = "new test name"
	description                = "new testing things"
	protocol_type              = "HTTP"
	route_selection_expression = "$request.body.type"
}
`
