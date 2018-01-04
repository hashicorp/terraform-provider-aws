package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSAPIGatewayDocumentationPart_basic(t *testing.T) {
	var conf apigateway.DocumentationPart

	rString := acctest.RandString(8)
	apiName := fmt.Sprintf("tf_acc_api_%s", rString)
	properties := `{\"description\":\"Terraform Acceptance Test\"}`
	uProperties := `{\"description\":\"Terraform Acceptance Test Updated\"}`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDocumentationPartDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSAPIGatewayDocumentationPartConfig(apiName, properties),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDocumentationPartExists("aws_api_gateway_documentation_part.test", &conf),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "location.#", "1"),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "location.0.type", "API"),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "properties", `{"description":"Terraform Acceptance Test"}`),
					resource.TestCheckResourceAttrSet("aws_api_gateway_documentation_part.test", "rest_api_id"),
				),
			},
			resource.TestStep{
				Config: testAccAWSAPIGatewayDocumentationPartConfig(apiName, uProperties),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDocumentationPartExists("aws_api_gateway_documentation_part.test", &conf),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "location.#", "1"),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "location.0.type", "API"),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "properties", `{"description":"Terraform Acceptance Test Updated"}`),
					resource.TestCheckResourceAttrSet("aws_api_gateway_documentation_part.test", "rest_api_id"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayDocumentationPart_method(t *testing.T) {
	var conf apigateway.DocumentationPart

	rString := acctest.RandString(8)
	apiName := fmt.Sprintf("tf_acc_api_%s", rString)
	properties := `{\"description\":\"Terraform Acceptance Test\"}`
	uProperties := `{\"description\":\"Terraform Acceptance Test Updated\"}`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDocumentationPartDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSAPIGatewayDocumentationPartMethodConfig(apiName, properties),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDocumentationPartExists("aws_api_gateway_documentation_part.test", &conf),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "location.#", "1"),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "location.0.type", "METHOD"),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "location.0.method", "GET"),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "location.0.path", "/terraform-acc-test"),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "properties", `{"description":"Terraform Acceptance Test"}`),
					resource.TestCheckResourceAttrSet("aws_api_gateway_documentation_part.test", "rest_api_id"),
				),
			},
			resource.TestStep{
				Config: testAccAWSAPIGatewayDocumentationPartMethodConfig(apiName, uProperties),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDocumentationPartExists("aws_api_gateway_documentation_part.test", &conf),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "location.#", "1"),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "location.0.type", "METHOD"),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "location.0.method", "GET"),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "location.0.path", "/terraform-acc-test"),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "properties", `{"description":"Terraform Acceptance Test Updated"}`),
					resource.TestCheckResourceAttrSet("aws_api_gateway_documentation_part.test", "rest_api_id"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayDocumentationPart_responseHeader(t *testing.T) {
	var conf apigateway.DocumentationPart

	rString := acctest.RandString(8)
	apiName := fmt.Sprintf("tf_acc_api_%s", rString)
	properties := `{\"description\":\"Terraform Acceptance Test\"}`
	uProperties := `{\"description\":\"Terraform Acceptance Test Updated\"}`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDocumentationPartDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSAPIGatewayDocumentationPartResponseHeaderConfig(apiName, properties),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDocumentationPartExists("aws_api_gateway_documentation_part.test", &conf),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "location.#", "1"),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "location.0.type", "RESPONSE_HEADER"),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "location.0.method", "GET"),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "location.0.name", "tfacc"),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "location.0.path", "/terraform-acc-test"),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "location.0.status_code", "200"),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "properties", `{"description":"Terraform Acceptance Test"}`),
					resource.TestCheckResourceAttrSet("aws_api_gateway_documentation_part.test", "rest_api_id"),
				),
			},
			resource.TestStep{
				Config: testAccAWSAPIGatewayDocumentationPartResponseHeaderConfig(apiName, uProperties),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDocumentationPartExists("aws_api_gateway_documentation_part.test", &conf),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "location.#", "1"),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "location.0.type", "RESPONSE_HEADER"),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "location.0.method", "GET"),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "location.0.name", "tfacc"),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "location.0.path", "/terraform-acc-test"),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "location.0.status_code", "200"),
					resource.TestCheckResourceAttr("aws_api_gateway_documentation_part.test", "properties", `{"description":"Terraform Acceptance Test Updated"}`),
					resource.TestCheckResourceAttrSet("aws_api_gateway_documentation_part.test", "rest_api_id"),
				),
			},
		},
	})
}

func testAccCheckAWSAPIGatewayDocumentationPartExists(n string, res *apigateway.DocumentationPart) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway Documentation Part ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigateway

		req := &apigateway.GetDocumentationPartInput{
			DocumentationPartId: aws.String(rs.Primary.ID),
			RestApiId:           aws.String(rs.Primary.Attributes["rest_api_id"]),
		}
		docPart, err := conn.GetDocumentationPart(req)
		if err != nil {
			return err
		}

		*res = *docPart

		return nil
	}
}

func testAccCheckAWSAPIGatewayDocumentationPartDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigateway

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_documentation_part" {
			continue
		}

		req := &apigateway.GetDocumentationPartInput{
			DocumentationPartId: aws.String(rs.Primary.ID),
			RestApiId:           aws.String(rs.Primary.Attributes["rest_api_id"]),
		}
		_, err := conn.GetDocumentationPart(req)
		if err != nil {
			if isAWSErr(err, "NotFoundException", "") {
				return nil
			}
			return err
		}

		return fmt.Errorf("API Gateway Documentation Part %s/%s still exists.",
			rs.Primary.ID, rs.Primary.Attributes["rest_api_id"])
	}
	return nil
}

func testAccAWSAPIGatewayDocumentationPartConfig(apiName, properties string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_documentation_part" "test" {
  location {
    type = "API"
  }
  properties = "%s"
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
}

resource "aws_api_gateway_rest_api" "test" {
  name = "%s"
}
`, properties, apiName)
}

func testAccAWSAPIGatewayDocumentationPartMethodConfig(apiName, properties string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_documentation_part" "test" {
  location {
    type = "METHOD"
    method = "GET"
    path = "/terraform-acc-test"
  }
  properties = "%s"
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
}

resource "aws_api_gateway_rest_api" "test" {
  name = "%s"
}
`, properties, apiName)
}

func testAccAWSAPIGatewayDocumentationPartResponseHeaderConfig(apiName, properties string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_documentation_part" "test" {
  location {
    type = "RESPONSE_HEADER"
    method = "GET"
    name = "tfacc"
    path = "/terraform-acc-test"
    status_code = "200"
  }
  properties = "%s"
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
}

resource "aws_api_gateway_rest_api" "test" {
  name = "%s"
}
`, properties, apiName)
}
