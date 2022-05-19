package apigateway_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
)

func TestAccAPIGatewayDocumentationPart_basic(t *testing.T) {
	var conf apigateway.DocumentationPart

	rString := sdkacctest.RandString(8)
	apiName := fmt.Sprintf("tf-acc-test_api_doc_part_basic_%s", rString)
	properties := `{"description":"Terraform Acceptance Test"}`
	uProperties := `{"description":"Terraform Acceptance Test Updated"}`

	resourceName := "aws_api_gateway_documentation_part.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDocumentationPartDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentationPartConfig(apiName, strconv.Quote(properties)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentationPartExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "location.0.type", "API"),
					resource.TestCheckResourceAttr(resourceName, "properties", properties),
					resource.TestCheckResourceAttrSet(resourceName, "rest_api_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentationPartConfig(apiName, strconv.Quote(uProperties)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentationPartExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "location.0.type", "API"),
					resource.TestCheckResourceAttr(resourceName, "properties", uProperties),
					resource.TestCheckResourceAttrSet(resourceName, "rest_api_id"),
				),
			},
		},
	})
}

func TestAccAPIGatewayDocumentationPart_method(t *testing.T) {
	var conf apigateway.DocumentationPart

	rString := sdkacctest.RandString(8)
	apiName := fmt.Sprintf("tf-acc-test_api_doc_part_method_%s", rString)
	properties := `{"description":"Terraform Acceptance Test"}`
	uProperties := `{"description":"Terraform Acceptance Test Updated"}`

	resourceName := "aws_api_gateway_documentation_part.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDocumentationPartDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentationPartMethodConfig(apiName, strconv.Quote(properties)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentationPartExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "location.0.type", "METHOD"),
					resource.TestCheckResourceAttr(resourceName, "location.0.method", "GET"),
					resource.TestCheckResourceAttr(resourceName, "location.0.path", "/terraform-acc-test"),
					resource.TestCheckResourceAttr(resourceName, "properties", properties),
					resource.TestCheckResourceAttrSet(resourceName, "rest_api_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentationPartMethodConfig(apiName, strconv.Quote(uProperties)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentationPartExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "location.0.type", "METHOD"),
					resource.TestCheckResourceAttr(resourceName, "location.0.method", "GET"),
					resource.TestCheckResourceAttr(resourceName, "location.0.path", "/terraform-acc-test"),
					resource.TestCheckResourceAttr(resourceName, "properties", uProperties),
					resource.TestCheckResourceAttrSet(resourceName, "rest_api_id"),
				),
			},
		},
	})
}

func TestAccAPIGatewayDocumentationPart_responseHeader(t *testing.T) {
	var conf apigateway.DocumentationPart

	rString := sdkacctest.RandString(8)
	apiName := fmt.Sprintf("tf-acc-test_api_doc_part_resp_header_%s", rString)
	properties := `{"description":"Terraform Acceptance Test"}`
	uProperties := `{"description":"Terraform Acceptance Test Updated"}`

	resourceName := "aws_api_gateway_documentation_part.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDocumentationPartDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentationPartResponseHeaderConfig(apiName, strconv.Quote(properties)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentationPartExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "location.0.type", "RESPONSE_HEADER"),
					resource.TestCheckResourceAttr(resourceName, "location.0.method", "GET"),
					resource.TestCheckResourceAttr(resourceName, "location.0.name", "tfacc"),
					resource.TestCheckResourceAttr(resourceName, "location.0.path", "/terraform-acc-test"),
					resource.TestCheckResourceAttr(resourceName, "location.0.status_code", "200"),
					resource.TestCheckResourceAttr(resourceName, "properties", properties),
					resource.TestCheckResourceAttrSet(resourceName, "rest_api_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentationPartResponseHeaderConfig(apiName, strconv.Quote(uProperties)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentationPartExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "location.0.type", "RESPONSE_HEADER"),
					resource.TestCheckResourceAttr(resourceName, "location.0.method", "GET"),
					resource.TestCheckResourceAttr(resourceName, "location.0.name", "tfacc"),
					resource.TestCheckResourceAttr(resourceName, "location.0.path", "/terraform-acc-test"),
					resource.TestCheckResourceAttr(resourceName, "location.0.status_code", "200"),
					resource.TestCheckResourceAttr(resourceName, "properties", uProperties),
					resource.TestCheckResourceAttrSet(resourceName, "rest_api_id"),
				),
			},
		},
	})
}

func TestAccAPIGatewayDocumentationPart_disappears(t *testing.T) {
	var conf apigateway.DocumentationPart

	rString := sdkacctest.RandString(8)
	apiName := fmt.Sprintf("tf-acc-test_api_doc_part_basic_%s", rString)
	properties := `{"description":"Terraform Acceptance Test"}`

	resourceName := "aws_api_gateway_documentation_part.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDocumentationPartDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentationPartConfig(apiName, strconv.Quote(properties)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentationPartExists(resourceName, &conf),
					acctest.CheckResourceDisappears(acctest.Provider, tfapigateway.ResourceDocumentationPart(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDocumentationPartExists(n string, res *apigateway.DocumentationPart) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway Documentation Part ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

		apiId, id, err := tfapigateway.DecodeDocumentationPartID(rs.Primary.ID)
		if err != nil {
			return err
		}

		req := &apigateway.GetDocumentationPartInput{
			DocumentationPartId: aws.String(id),
			RestApiId:           aws.String(apiId),
		}
		docPart, err := conn.GetDocumentationPart(req)
		if err != nil {
			return err
		}

		*res = *docPart

		return nil
	}
}

func testAccCheckDocumentationPartDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_documentation_part" {
			continue
		}

		apiId, id, err := tfapigateway.DecodeDocumentationPartID(rs.Primary.ID)
		if err != nil {
			return err
		}

		req := &apigateway.GetDocumentationPartInput{
			DocumentationPartId: aws.String(id),
			RestApiId:           aws.String(apiId),
		}
		_, err = conn.GetDocumentationPart(req)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
				return nil
			}
			return err
		}

		return fmt.Errorf("API Gateway Documentation Part %q still exists.", rs.Primary.ID)
	}
	return nil
}

func testAccDocumentationPartConfig(apiName, properties string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_documentation_part" "test" {
  location {
    type = "API"
  }
  properties  = %s
  rest_api_id = aws_api_gateway_rest_api.test.id
}

resource "aws_api_gateway_rest_api" "test" {
  name = "%s"
}
`, properties, apiName)
}

func testAccDocumentationPartMethodConfig(apiName, properties string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_documentation_part" "test" {
  location {
    type   = "METHOD"
    method = "GET"
    path   = "/terraform-acc-test"
  }
  properties  = %s
  rest_api_id = aws_api_gateway_rest_api.test.id
}

resource "aws_api_gateway_rest_api" "test" {
  name = "%s"
}
`, properties, apiName)
}

func testAccDocumentationPartResponseHeaderConfig(apiName, properties string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_documentation_part" "test" {
  location {
    type        = "RESPONSE_HEADER"
    method      = "GET"
    name        = "tfacc"
    path        = "/terraform-acc-test"
    status_code = "200"
  }
  properties  = %s
  rest_api_id = aws_api_gateway_rest_api.test.id
}

resource "aws_api_gateway_rest_api" "test" {
  name = "%s"
}
`, properties, apiName)
}
