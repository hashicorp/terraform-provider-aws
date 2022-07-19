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

func TestAccAPIGatewayResource_basic(t *testing.T) {
	var conf apigateway.Resource
	rName := fmt.Sprintf("tf-test-acc-%s", sdkacctest.RandString(8))
	resourceName := "aws_api_gateway_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(resourceName, &conf),
					testAccCheckResourceAttributes(&conf, "/test"),
					resource.TestCheckResourceAttr(
						resourceName, "path_part", "test"),
					resource.TestCheckResourceAttr(
						resourceName, "path", "/test"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccResourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayResource_update(t *testing.T) {
	var conf apigateway.Resource
	rName := fmt.Sprintf("tf-test-acc-%s", sdkacctest.RandString(8))
	resourceName := "aws_api_gateway_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(resourceName, &conf),
					testAccCheckResourceAttributes(&conf, "/test"),
					resource.TestCheckResourceAttr(
						resourceName, "path_part", "test"),
					resource.TestCheckResourceAttr(
						resourceName, "path", "/test"),
				),
			},

			{
				Config: testAccResourceConfig_updatePathPart(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(resourceName, &conf),
					testAccCheckResourceAttributes(&conf, "/test_changed"),
					resource.TestCheckResourceAttr(
						resourceName, "path_part", "test_changed"),
					resource.TestCheckResourceAttr(
						resourceName, "path", "/test_changed"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccResourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayResource_disappears(t *testing.T) {
	var conf apigateway.Resource
	rName := fmt.Sprintf("tf-test-acc-%s", sdkacctest.RandString(8))
	resourceName := "aws_api_gateway_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(resourceName, &conf),
					acctest.CheckResourceDisappears(acctest.Provider, tfapigateway.ResourceResource(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckResourceAttributes(conf *apigateway.Resource, path string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *conf.Path != path {
			return fmt.Errorf("Wrong Path: %q", *conf.Path)
		}

		return nil
	}
}

func testAccCheckResourceExists(n string, res *apigateway.Resource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway Resource ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

		req := &apigateway.GetResourceInput{
			ResourceId: aws.String(rs.Primary.ID),
			RestApiId:  aws.String(s.RootModule().Resources["aws_api_gateway_rest_api.test"].Primary.ID),
		}
		describe, err := conn.GetResource(req)
		if err != nil {
			return err
		}

		if *describe.Id != rs.Primary.ID {
			return fmt.Errorf("APIGateway Resource not found")
		}

		*res = *describe

		return nil
	}
}

func testAccCheckResourceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_resource" {
			continue
		}

		req := &apigateway.GetResourcesInput{
			RestApiId: aws.String(s.RootModule().Resources["aws_api_gateway_rest_api.test"].Primary.ID),
		}
		describe, err := conn.GetResources(req)

		if err == nil {
			if len(describe.Items) != 0 &&
				*describe.Items[0].Id == rs.Primary.ID {
				return fmt.Errorf("API Gateway Resource still exists")
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

func testAccResourceImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["rest_api_id"], rs.Primary.ID), nil
	}
}

func testAccResourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "%s"
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}
`, rName)
}

func testAccResourceConfig_updatePathPart(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "%s"
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test_changed"
}
`, rName)
}
