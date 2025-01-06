// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayDocumentationVersion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetDocumentationVersionOutput
	rString := sdkacctest.RandString(8)
	version := fmt.Sprintf("tf-acc-test_version_%s", rString)
	apiName := fmt.Sprintf("tf-acc-test_api_doc_version_basic_%s", rString)
	resourceName := "aws_api_gateway_documentation_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentationVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentationVersionConfig_basic(version, apiName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentationVersionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, version),
					resource.TestCheckResourceAttrSet(resourceName, "rest_api_id"),
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

func TestAccAPIGatewayDocumentationVersion_allFields(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetDocumentationVersionOutput
	rString := sdkacctest.RandString(8)
	version := fmt.Sprintf("tf-acc-test_version_%s", rString)
	apiName := fmt.Sprintf("tf-acc-test_api_doc_version_method_%s", rString)
	stageName := fmt.Sprintf("tf-acc-test_stage_%s", rString)
	description := fmt.Sprintf("Tf Acc Test description %s", rString)
	uDescription := fmt.Sprintf("Tf Acc Test description updated %s", rString)
	resourceName := "aws_api_gateway_documentation_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentationVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentationVersionConfig_allFields(version, apiName, stageName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentationVersionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, version),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttrSet(resourceName, "rest_api_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentationVersionConfig_allFields(version, apiName, stageName, uDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentationVersionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, version),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, uDescription),
					resource.TestCheckResourceAttrSet(resourceName, "rest_api_id"),
				),
			},
		},
	})
}

func TestAccAPIGatewayDocumentationVersion_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetDocumentationVersionOutput
	rString := sdkacctest.RandString(8)
	version := fmt.Sprintf("tf-acc-test_version_%s", rString)
	apiName := fmt.Sprintf("tf-acc-test_api_doc_version_basic_%s", rString)
	resourceName := "aws_api_gateway_documentation_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentationVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentationVersionConfig_basic(version, apiName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentationVersionExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigateway.ResourceDocumentationVersion(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDocumentationVersionExists(ctx context.Context, n string, v *apigateway.GetDocumentationVersionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		output, err := tfapigateway.FindDocumentationVersionByTwoPartKey(ctx, conn, rs.Primary.Attributes["rest_api_id"], rs.Primary.Attributes[names.AttrVersion])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckDocumentationVersionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_documentation_version" {
				continue
			}

			_, err := tfapigateway.FindDocumentationVersionByTwoPartKey(ctx, conn, rs.Primary.Attributes["rest_api_id"], rs.Primary.Attributes[names.AttrVersion])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway Documentation Version %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDocumentationVersionConfig_basic(version, apiName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_documentation_version" "test" {
  version     = "%s"
  rest_api_id = aws_api_gateway_rest_api.test.id
  depends_on  = [aws_api_gateway_documentation_part.test]
}

resource "aws_api_gateway_documentation_part" "test" {
  location {
    type = "API"
  }

  properties  = "{\"description\":\"Terraform Acceptance Test\"}"
  rest_api_id = aws_api_gateway_rest_api.test.id
}

resource "aws_api_gateway_rest_api" "test" {
  name = "%s"
}
`, version, apiName)
}

func testAccDocumentationVersionConfig_allFields(version, apiName, stageName, description string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_documentation_version" "test" {
  version     = "%s"
  rest_api_id = aws_api_gateway_rest_api.test.id
  description = "%s"
  depends_on  = [aws_api_gateway_documentation_part.test]
}

resource "aws_api_gateway_documentation_part" "test" {
  location {
    type = "API"
  }

  properties  = "{\"description\":\"Terraform Acceptance Test\"}"
  rest_api_id = aws_api_gateway_rest_api.test.id
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

resource "aws_api_gateway_method_response" "error" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method
  status_code = "400"
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  type                    = "HTTP"
  uri                     = "https://www.google.co.uk"
  integration_http_method = "GET"
}

resource "aws_api_gateway_integration_response" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_integration.test.http_method
  status_code = aws_api_gateway_method_response.error.status_code
}

resource "aws_api_gateway_deployment" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = "first"
  depends_on  = [aws_api_gateway_integration_response.test]
}

resource "aws_api_gateway_stage" "test" {
  stage_name            = "%s"
  rest_api_id           = aws_api_gateway_rest_api.test.id
  deployment_id         = aws_api_gateway_deployment.test.id
  documentation_version = aws_api_gateway_documentation_version.test.version
}

resource "aws_api_gateway_rest_api" "test" {
  name = "%s"
}
`, version, description, stageName, apiName)
}
