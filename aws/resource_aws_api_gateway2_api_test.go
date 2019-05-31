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

func TestAccAWSAPIGateway2Api_basic(t *testing.T) {
	var conf apigatewayv2.Api
	resourceName := "aws_api_gateway_v2_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGateway2ApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGateway2ApiConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2ApiExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "testing things"),
					resource.TestCheckResourceAttr(resourceName, "name", "test"),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", "WEBSOCKET"),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.body.action"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSAPIGateway2ApiImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayV2_update(t *testing.T) {
	var conf apigatewayv2.Api
	resourceName := "aws_api_gateway_v2_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGateway2ApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGateway2ApiConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2ApiExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", "test"),
					resource.TestCheckResourceAttr(resourceName, "description", "testing things"),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", "WEBSOCKET"),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.body.action"),
				),
			},
			{
				Config: testAccAWSAPIGateway2ApiConfig_updatePathPart,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2ApiExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", "new test name"),
					resource.TestCheckResourceAttr(resourceName, "description", "new testing things"),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.body.type"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSAPIGateway2ApiImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSAPIGateway2ApiExists(n string, res *apigatewayv2.Api) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway V2 API ID is set")
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
			return fmt.Errorf("APIGateway V2 API not found")
		}

		//*res = *api

		return nil
	}
}

func testAccCheckAWSAPIGateway2ApiDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_v2_api" {
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

func testAccAWSAPIGateway2ApiImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["api_id"], rs.Primary.ID), nil
	}
}

const testAccAWSAPIGateway2ApiConfig = `
resource "aws_api_gateway_v2_api" "test" {
	name                       = "test"
	description                = "testing things"
	protocol_type              = "WEBSOCKET"
	route_selection_expression = "$request.body.action"
}
`

const testAccAWSAPIGateway2ApiConfig_updatePathPart = `
resource "aws_api_gateway_v2_api" "test" {
	name                       = "new test name"
	description                = "new testing things"
	protocol_type              = "HTTP"
	route_selection_expression = "$request.body.type"
}
`
