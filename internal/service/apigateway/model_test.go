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

func TestAccAPIGatewayModel_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetModelOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	modelName := sdkacctest.RandString(16)
	resourceName := "aws_api_gateway_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_basic(rName, modelName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrContentType, "application/json"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "a test schema"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, modelName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccModelImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSchema},
			},
		},
	})
}

func TestAccAPIGatewayModel_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetModelOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	modelName := sdkacctest.RandString(16)
	resourceName := "aws_api_gateway_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_basic(rName, modelName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigateway.ResourceModel(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckModelExists(ctx context.Context, n string, v *apigateway.GetModelOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		output, err := tfapigateway.FindModelByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes["rest_api_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckModelDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_model" {
				continue
			}

			_, err := tfapigateway.FindModelByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes["rest_api_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway Model %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccModelImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["rest_api_id"], rs.Primary.Attributes[names.AttrName]), nil
	}
}

func testAccModelConfig_basic(rName, modelName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_model" "test" {
  rest_api_id  = aws_api_gateway_rest_api.test.id
  name         = %[2]q
  description  = "a test schema"
  content_type = "application/json"
  schema       = <<EOF
{
  "type": "object"
}
EOF
}
`, rName, modelName)
}
