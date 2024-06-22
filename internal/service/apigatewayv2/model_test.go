// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigatewayv2 "github.com/hashicorp/terraform-provider-aws/internal/service/apigatewayv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayV2Model_basic(t *testing.T) {
	ctx := acctest.Context(t)
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_basic(rName, schema),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrContentType, "application/json"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, names.AttrSchema, schema),
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
	ctx := acctest.Context(t)
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_basic(rName, schema),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName, &apiId, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigatewayv2.ResourceModel(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayV2Model_allAttributes(t *testing.T) {
	ctx := acctest.Context(t)
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_allAttributes(rName, schema1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrContentType, "text/x-json"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, names.AttrSchema, schema1),
				),
			},
			{
				Config: testAccModelConfig_basic(rName, schema2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrContentType, "application/json"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, names.AttrSchema, schema2),
				),
			},
			{
				Config: testAccModelConfig_allAttributes(rName, schema1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrContentType, "text/x-json"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, names.AttrSchema, schema1),
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

func testAccCheckModelDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_apigatewayv2_model" {
				continue
			}

			_, err := tfapigatewayv2.FindModelByTwoPartKey(ctx, conn, rs.Primary.Attributes["api_id"], rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway v2 Model %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckModelExists(ctx context.Context, n string, apiID *string, v *apigatewayv2.GetModelOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Client(ctx)

		output, err := tfapigatewayv2.FindModelByTwoPartKey(ctx, conn, rs.Primary.Attributes["api_id"], rs.Primary.ID)

		if err != nil {
			return err
		}

		*apiID = rs.Primary.Attributes["api_id"]
		*v = *output

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
