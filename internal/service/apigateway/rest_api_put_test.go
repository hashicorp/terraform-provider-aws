// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayRestAPIPut_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var restAPI apigateway.GetRestApiOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api_put.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIPutConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIPutExists(ctx, t, resourceName, &restAPI),
					resource.TestCheckResourceAttr(resourceName, "fail_on_warnings", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "rest_api_id"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "rest_api_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "rest_api_id",
				ImportStateVerifyIgnore:              []string{"body", names.AttrTriggers, "fail_on_warnings"},
			},
		},
	})
}

// func TestAccAPIGatewayRestAPIPut_disappears(t *testing.T) {
// Delete is no-op

func TestAccAPIGatewayRestAPIPut_multistage(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var restAPI apigateway.GetRestApiOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api_put.testv1"
	resourceName2 := "aws_api_gateway_rest_api_put.testv2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIPutConfig_multistage(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIPutExists(ctx, t, resourceName, &restAPI),
					testAccCheckRESTAPIPutExists(ctx, t, resourceName2, &restAPI),
				),
			},
		},
	})
}

func testAccCheckRESTAPIPutExists(ctx context.Context, t *testing.T, n string, v *apigateway.GetRestApiOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).APIGatewayClient(ctx)

		output, err := tfapigateway.FindRestAPIByID(ctx, conn, rs.Primary.Attributes["rest_api_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccRestAPIPutConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_rest_api_put" "test" {
  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = %[1]q
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://api.example.com/"
          }
        }
      }
    }
  })

  fail_on_warnings = true
  rest_api_id      = aws_api_gateway_rest_api.test.id
}
`, rName)
}

func testAccRestAPIPutConfig_multistage(rName string) string {
	return `
resource "aws_api_gateway_rest_api" "test" {
  name = "Simple API"
}

resource "aws_api_gateway_rest_api_put" "testv1" {
  body             = file("test-fixtures/rest_api_put_v1.yaml")
  fail_on_warnings = true
  rest_api_id      = aws_api_gateway_rest_api.test.id

  triggers = {
    redeployment = sha1(file("test-fixtures/rest_api_put_v1.yaml"))
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_deployment" "testv1" {
  rest_api_id = aws_api_gateway_rest_api.test.id

  triggers = {
    redeployment = aws_api_gateway_rest_api_put.testv1.triggers.redeployment
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_stage" "testv1" {
  stage_name    = "v1"
  rest_api_id   = aws_api_gateway_rest_api.test.id
  deployment_id = aws_api_gateway_deployment.testv1.id
}

resource "aws_api_gateway_rest_api_put" "testv2" {
  depends_on = [
    aws_api_gateway_stage.testv1
  ]

  body             = file("test-fixtures/rest_api_put_v2.yaml")
  fail_on_warnings = true
  rest_api_id      = aws_api_gateway_rest_api.test.id

  triggers = {
    redeployment = sha1(file("test-fixtures/rest_api_put_v2.yaml"))
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_deployment" "testv2" {
  rest_api_id = aws_api_gateway_rest_api.test.id

  triggers = {
    redeployment = aws_api_gateway_rest_api_put.testv2.triggers.redeployment
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_stage" "testv2" {
  stage_name    = "v2"
  rest_api_id   = aws_api_gateway_rest_api.test.id
  deployment_id = aws_api_gateway_deployment.testv2.id
}
`
}
