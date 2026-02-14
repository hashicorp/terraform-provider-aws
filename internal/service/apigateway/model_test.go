// Copyright IBM Corp. 2014, 2026
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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayModel_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetModelOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	modelName := sdkacctest.RandString(16)
	resourceName := "aws_api_gateway_model.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_basic(rName, modelName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, t, resourceName, &conf),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	modelName := sdkacctest.RandString(16)
	resourceName := "aws_api_gateway_model.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_basic(rName, modelName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, t, resourceName, &conf),
					acctest.CheckSDKResourceDisappears(ctx, t, tfapigateway.ResourceModel(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckModelExists(ctx context.Context, t *testing.T, n string, v *apigateway.GetModelOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).APIGatewayClient(ctx)

		output, err := tfapigateway.FindModelByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes["rest_api_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckModelDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).APIGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_model" {
				continue
			}

			_, err := tfapigateway.FindModelByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes["rest_api_id"])

			if retry.NotFound(err) {
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
