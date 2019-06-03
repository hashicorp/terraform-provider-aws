package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSAPIGateway2Api_basic(t *testing.T) {
	resourceName := "aws_api_gateway_v2_api.test"
	rName := fmt.Sprintf("terraform-testacc-apigwv2api-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGateway2ApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGateway2ApiConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2ApiExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeWebsocket),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.body.action"),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
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

func TestAccAWSAPIGateway2Api_update(t *testing.T) {
	resourceName := "aws_api_gateway_v2_api.test"
	rName1 := fmt.Sprintf("terraform-testacc-apigwv2api-%d", acctest.RandInt())
	rName2 := fmt.Sprintf("terraform-testacc-apigwv2api-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGateway2ApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGateway2ApiConfig_basic(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2ApiExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeWebsocket),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.body.action"),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
				),
			},
			{
				Config: testAccAWSAPIGateway2ApiConfig_update(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2ApiExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$context.authorizer.usageIdentifierKey"),
					resource.TestCheckResourceAttr(resourceName, "description", "test description"),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.body.service"),
					resource.TestCheckResourceAttr(resourceName, "version", "v1"),
				),
			},
			{
				Config: testAccAWSAPIGateway2ApiConfig_update(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2ApiExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$context.authorizer.usageIdentifierKey"),
					resource.TestCheckResourceAttr(resourceName, "description", "test description"),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.body.service"),
					resource.TestCheckResourceAttr(resourceName, "version", "v1"),
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

func testAccCheckAWSAPIGateway2ApiDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_v2_api" {
			continue
		}

		_, err := conn.GetApi(&apigatewayv2.GetApiInput{
			ApiId: aws.String(rs.Primary.ID),
		})
		if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("API Gateway v2 API %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSAPIGateway2ApiExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 API ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		_, err := conn.GetApi(&apigatewayv2.GetApiInput{
			ApiId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAWSAPIGateway2ApiConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_v2_api" "test" {
  name                       = %[1]q
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"
}
`, rName)
}

func testAccAWSAPIGateway2ApiConfig_update(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_v2_api" "test" {
  api_key_selection_expression = "$context.authorizer.usageIdentifierKey"
  description                  = "test description"
  name                         = %[1]q
  protocol_type                = "WEBSOCKET"
  route_selection_expression   = "$request.body.service"
  version                      = "v1"
}
`, rName)
}
