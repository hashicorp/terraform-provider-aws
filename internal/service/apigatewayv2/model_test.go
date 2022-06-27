package apigatewayv2_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccAPIGatewayV2Model_basic(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetModelOutput
	resourceName := "aws_apigatewayv2_model.test"
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "")

	schema := `
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "title": "ExampleModel",
  "type": "object",
  "properties": {
    "id": {
      "type": "string"
    }
  }
}
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_basic(rName, schema),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "content_type", "application/json"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "schema", schema),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccModelImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayV2Model_disappears(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetModelOutput
	resourceName := "aws_apigatewayv2_model.test"
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "")

	schema := `
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "title": "ExampleModel",
  "type": "object",
  "properties": {
    "id": {
      "type": "string"
    }
  }
}
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_basic(rName, schema),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(resourceName, &apiId, &v),
					testAccCheckModelDisappears(&apiId, &v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayV2Model_allAttributes(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetModelOutput
	resourceName := "aws_apigatewayv2_model.test"
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "")

	schema1 := `
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "title": "ExampleModel1",
  "type": "object",
  "properties": {
    "id": {
      "type": "string"
    }
  }
}
`
	schema2 := `
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "title": "ExampleModel",
  "type": "object",
  "properties": {
    "ids": {
      "type": "array",
        "items":{
          "type": "integer"
        }
    }
  }
}
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_allAttributes(rName, schema1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "content_type", "text/x-json"),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "schema", schema1),
				),
			},
			{
				Config: testAccModelConfig_basic(rName, schema2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "content_type", "application/json"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "schema", schema2),
				),
			},
			{
				Config: testAccModelConfig_allAttributes(rName, schema1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "content_type", "text/x-json"),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "schema", schema1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccModelImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckModelDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_apigatewayv2_model" {
			continue
		}

		_, err := conn.GetModel(&apigatewayv2.GetModelInput{
			ApiId:   aws.String(rs.Primary.Attributes["api_id"]),
			ModelId: aws.String(rs.Primary.ID),
		})
		if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("API Gateway v2 model %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckModelDisappears(apiId *string, v *apigatewayv2.GetModelOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Conn

		_, err := conn.DeleteModel(&apigatewayv2.DeleteModelInput{
			ApiId:   apiId,
			ModelId: v.ModelId,
		})

		return err
	}
}

func testAccCheckModelExists(n string, vApiId *string, v *apigatewayv2.GetModelOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 model ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Conn

		apiId := aws.String(rs.Primary.Attributes["api_id"])
		resp, err := conn.GetModel(&apigatewayv2.GetModelInput{
			ApiId:   apiId,
			ModelId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*vApiId = *apiId
		*v = *resp

		return nil
	}
}

func testAccModelImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["api_id"], rs.Primary.ID), nil
	}
}

func testAccModelConfig_api(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name                       = %[1]q
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"
}
`, rName)
}

func testAccModelConfig_basic(rName, schema string) string {
	return testAccModelConfig_api(rName) + fmt.Sprintf(`
resource "aws_apigatewayv2_model" "test" {
  api_id       = aws_apigatewayv2_api.test.id
  content_type = "application/json"
  name         = %[1]q
  schema       = %[2]q
}
`, rName, schema)
}

func testAccModelConfig_allAttributes(rName, schema string) string {
	return testAccModelConfig_api(rName) + fmt.Sprintf(`
resource "aws_apigatewayv2_model" "test" {
  api_id       = aws_apigatewayv2_api.test.id
  content_type = "text/x-json"
  name         = %[1]q
  description  = "test"
  schema       = %[2]q
}
`, rName, schema)
}
