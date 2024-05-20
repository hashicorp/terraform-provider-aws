// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"context"
	"fmt"
	"strconv"
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

func TestAccAPIGatewayDocumentationPart_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetDocumentationPartOutput
	rString := sdkacctest.RandString(8)
	apiName := fmt.Sprintf("tf-acc-test_api_doc_part_basic_%s", rString)
	properties := `{"description":"Terraform Acceptance Test"}`
	uProperties := `{"description":"Terraform Acceptance Test Updated"}`
	resourceName := "aws_api_gateway_documentation_part.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentationPartDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentationPartConfig_basic(apiName, strconv.Quote(properties)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDocumentationPartExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "documentation_part_id"),
					resource.TestCheckResourceAttr(resourceName, "location.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "location.0.type", "API"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProperties, properties),
					resource.TestCheckResourceAttrSet(resourceName, "rest_api_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentationPartConfig_basic(apiName, strconv.Quote(uProperties)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentationPartExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "location.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "location.0.type", "API"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProperties, uProperties),
					resource.TestCheckResourceAttrSet(resourceName, "rest_api_id"),
				),
			},
		},
	})
}

func TestAccAPIGatewayDocumentationPart_method(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetDocumentationPartOutput
	rString := sdkacctest.RandString(8)
	apiName := fmt.Sprintf("tf-acc-test_api_doc_part_method_%s", rString)
	properties := `{"description":"Terraform Acceptance Test"}`
	uProperties := `{"description":"Terraform Acceptance Test Updated"}`
	resourceName := "aws_api_gateway_documentation_part.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentationPartDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentationPartConfig_method(apiName, strconv.Quote(properties)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentationPartExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "location.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "location.0.type", "METHOD"),
					resource.TestCheckResourceAttr(resourceName, "location.0.method", "GET"),
					resource.TestCheckResourceAttr(resourceName, "location.0.path", "/terraform-acc-test"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProperties, properties),
					resource.TestCheckResourceAttrSet(resourceName, "rest_api_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentationPartConfig_method(apiName, strconv.Quote(uProperties)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentationPartExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "location.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "location.0.type", "METHOD"),
					resource.TestCheckResourceAttr(resourceName, "location.0.method", "GET"),
					resource.TestCheckResourceAttr(resourceName, "location.0.path", "/terraform-acc-test"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProperties, uProperties),
					resource.TestCheckResourceAttrSet(resourceName, "rest_api_id"),
				),
			},
		},
	})
}

func TestAccAPIGatewayDocumentationPart_responseHeader(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetDocumentationPartOutput
	rString := sdkacctest.RandString(8)
	apiName := fmt.Sprintf("tf-acc-test_api_doc_part_resp_header_%s", rString)
	properties := `{"description":"Terraform Acceptance Test"}`
	uProperties := `{"description":"Terraform Acceptance Test Updated"}`
	resourceName := "aws_api_gateway_documentation_part.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentationPartDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentationPartConfig_responseHeader(apiName, strconv.Quote(properties)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentationPartExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "location.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "location.0.type", "RESPONSE_HEADER"),
					resource.TestCheckResourceAttr(resourceName, "location.0.method", "GET"),
					resource.TestCheckResourceAttr(resourceName, "location.0.name", "tfacc"),
					resource.TestCheckResourceAttr(resourceName, "location.0.path", "/terraform-acc-test"),
					resource.TestCheckResourceAttr(resourceName, "location.0.status_code", "200"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProperties, properties),
					resource.TestCheckResourceAttrSet(resourceName, "rest_api_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentationPartConfig_responseHeader(apiName, strconv.Quote(uProperties)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentationPartExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "location.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "location.0.type", "RESPONSE_HEADER"),
					resource.TestCheckResourceAttr(resourceName, "location.0.method", "GET"),
					resource.TestCheckResourceAttr(resourceName, "location.0.name", "tfacc"),
					resource.TestCheckResourceAttr(resourceName, "location.0.path", "/terraform-acc-test"),
					resource.TestCheckResourceAttr(resourceName, "location.0.status_code", "200"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProperties, uProperties),
					resource.TestCheckResourceAttrSet(resourceName, "rest_api_id"),
				),
			},
		},
	})
}

func TestAccAPIGatewayDocumentationPart_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetDocumentationPartOutput
	rString := sdkacctest.RandString(8)
	apiName := fmt.Sprintf("tf-acc-test_api_doc_part_basic_%s", rString)
	properties := `{"description":"Terraform Acceptance Test"}`
	resourceName := "aws_api_gateway_documentation_part.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentationPartDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentationPartConfig_basic(apiName, strconv.Quote(properties)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentationPartExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigateway.ResourceDocumentationPart(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDocumentationPartExists(ctx context.Context, n string, v *apigateway.GetDocumentationPartOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		output, err := tfapigateway.FindDocumentationPartByTwoPartKey(ctx, conn, rs.Primary.Attributes["rest_api_id"], rs.Primary.Attributes["documentation_part_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckDocumentationPartDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_documentation_part" {
				continue
			}

			_, err := tfapigateway.FindDocumentationPartByTwoPartKey(ctx, conn, rs.Primary.Attributes["rest_api_id"], rs.Primary.Attributes["documentation_part_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway Documentation Part %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDocumentationPartConfig_basic(apiName, properties string) string {
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

func testAccDocumentationPartConfig_method(apiName, properties string) string {
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

func testAccDocumentationPartConfig_responseHeader(apiName, properties string) string {
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
