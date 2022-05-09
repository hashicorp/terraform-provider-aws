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

func TestAccAPIGatewayMethodResponse_basic(t *testing.T) {
	var conf apigateway.MethodResponse
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(10))
	resourceName := "aws_api_gateway_method_response.error"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMethodResponseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMethodResponseConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodResponseExists(resourceName, &conf),
					testAccCheckMethodResponseAttributes(&conf),
					resource.TestCheckResourceAttr(
						resourceName, "status_code", "400"),
					resource.TestCheckResourceAttr(
						resourceName, "response_models.application/json", "Error"),
				),
			},

			{
				Config: testAccMethodResponseUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodResponseExists(resourceName, &conf),
					testAccCheckMethodResponseAttributesUpdate(&conf),
					resource.TestCheckResourceAttr(
						resourceName, "status_code", "400"),
					resource.TestCheckResourceAttr(
						resourceName, "response_models.application/json", "Empty"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMethodResponseImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayMethodResponse_disappears(t *testing.T) {
	var conf apigateway.MethodResponse
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(10))
	resourceName := "aws_api_gateway_method_response.error"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMethodResponseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMethodResponseConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodResponseExists(resourceName, &conf),
					acctest.CheckResourceDisappears(acctest.Provider, tfapigateway.ResourceMethodResponse(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckMethodResponseAttributes(conf *apigateway.MethodResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *conf.StatusCode == "" {
			return fmt.Errorf("empty StatusCode")
		}
		if val, ok := conf.ResponseModels["application/json"]; !ok {
			return fmt.Errorf("missing application/json ResponseModel")
		} else {
			if *val != "Error" {
				return fmt.Errorf("wrong application/json ResponseModel")
			}
		}
		if val, ok := conf.ResponseParameters["method.response.header.Content-Type"]; !ok {
			return fmt.Errorf("missing Content-Type ResponseParameters")
		} else {
			if !*val {
				return fmt.Errorf("wrong ResponseParameters value")
			}
		}
		return nil
	}
}

func testAccCheckMethodResponseAttributesUpdate(conf *apigateway.MethodResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *conf.StatusCode == "" {
			return fmt.Errorf("empty StatusCode")
		}
		if val, ok := conf.ResponseModels["application/json"]; !ok {
			return fmt.Errorf("missing application/json ResponseModel")
		} else {
			if *val != "Empty" {
				return fmt.Errorf("wrong application/json ResponseModel")
			}
		}
		if conf.ResponseParameters["method.response.header.Content-Type"] != nil {
			return fmt.Errorf("Content-Type ResponseParameters shouldn't exist")
		}
		return nil
	}
}

func testAccCheckMethodResponseExists(n string, res *apigateway.MethodResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway Method ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

		req := &apigateway.GetMethodResponseInput{
			HttpMethod: aws.String("GET"),
			ResourceId: aws.String(s.RootModule().Resources["aws_api_gateway_resource.test"].Primary.ID),
			RestApiId:  aws.String(s.RootModule().Resources["aws_api_gateway_rest_api.test"].Primary.ID),
			StatusCode: aws.String(rs.Primary.Attributes["status_code"]),
		}
		describe, err := conn.GetMethodResponse(req)
		if err != nil {
			return err
		}

		*res = *describe

		return nil
	}
}

func testAccCheckMethodResponseDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_method_response" {
			continue
		}

		req := &apigateway.GetMethodResponseInput{
			HttpMethod: aws.String("GET"),
			ResourceId: aws.String(s.RootModule().Resources["aws_api_gateway_resource.test"].Primary.ID),
			RestApiId:  aws.String(s.RootModule().Resources["aws_api_gateway_rest_api.test"].Primary.ID),
			StatusCode: aws.String(rs.Primary.Attributes["status_code"]),
		}
		_, err := conn.GetMethodResponse(req)

		if err == nil {
			return fmt.Errorf("API Gateway Method still exists")
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

func testAccMethodResponseImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s/%s", rs.Primary.Attributes["rest_api_id"], rs.Primary.Attributes["resource_id"], rs.Primary.Attributes["http_method"], rs.Primary.Attributes["status_code"]), nil
	}
}

func testAccMethodResponseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "%s"
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

  request_models = {
    "application/json" = "Error"
  }
}

resource "aws_api_gateway_method_response" "error" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method
  status_code = "400"

  response_models = {
    "application/json" = "Error"
  }

  response_parameters = {
    "method.response.header.Content-Type" = true
  }
}
`, rName)
}

func testAccMethodResponseUpdateConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "%s"
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

  request_models = {
    "application/json" = "Error"
  }
}

resource "aws_api_gateway_method_response" "error" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method
  status_code = "400"

  response_models = {
    "application/json" = "Empty"
  }

  response_parameters = {
    "method.response.header.Host" = true
  }
}
`, rName)
}
