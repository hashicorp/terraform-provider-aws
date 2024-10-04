// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
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

func TestAccAPIGatewayV2Deployment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v apigatewayv2.GetDeploymentOutput
	resourceName := "aws_apigatewayv2_deployment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_basic(rName, "Test description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deployed", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Test description"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccDeploymentImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeploymentConfig_basic(rName, "Test description updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deployed", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Test description updated"),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2Deployment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v apigatewayv2.GetDeploymentOutput
	resourceName := "aws_apigatewayv2_deployment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_basic(rName, "Test description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigatewayv2.ResourceDeployment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayV2Deployment_triggers(t *testing.T) {
	ctx := acctest.Context(t)
	var deployment1, deployment2, deployment3, deployment4 apigatewayv2.GetDeploymentOutput
	resourceName := "aws_apigatewayv2_deployment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_triggers(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment1),
				),
				// Due to how the Terraform state is handled for resources during creation,
				// any SHA1 of whole resources will change after first apply, then stabilize.
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccDeploymentConfig_triggers(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment2),
					testAccCheckDeploymentRecreated(&deployment1, &deployment2),
				),
			},
			{
				Config: testAccDeploymentConfig_triggers(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment3),
					testAccCheckDeploymentNotRecreated(&deployment2, &deployment3),
				),
			},
			{
				ResourceName:            resourceName,
				ImportStateIdFunc:       testAccDeploymentImportStateIdFunc(resourceName),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrTriggers},
			},
			{
				Config: testAccDeploymentConfig_triggers(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment4),
					testAccCheckDeploymentRecreated(&deployment3, &deployment4),
				),
			},
		},
	})
}

func testAccCheckDeploymentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_apigatewayv2_deployment" {
				continue
			}

			_, err := tfapigatewayv2.FindDeploymentByTwoPartKey(ctx, conn, rs.Primary.Attributes["api_id"], rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway v2 Deployment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDeploymentExists(ctx context.Context, n string, v *apigatewayv2.GetDeploymentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Client(ctx)

		output, err := tfapigatewayv2.FindDeploymentByTwoPartKey(ctx, conn, rs.Primary.Attributes["api_id"], rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckDeploymentNotRecreated(i, j *apigatewayv2.GetDeploymentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.ToTime(i.CreatedDate).Equal(aws.ToTime(j.CreatedDate)) {
			return fmt.Errorf("API Gateway v2 Deployment recreated")
		}

		return nil
	}
}

func testAccCheckDeploymentRecreated(i, j *apigatewayv2.GetDeploymentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToTime(i.CreatedDate).Equal(aws.ToTime(j.CreatedDate)) {
			return fmt.Errorf("API Gateway v2 Deployment not recreated")
		}

		return nil
	}
}

func testAccDeploymentImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["api_id"], rs.Primary.ID), nil
	}
}

func testAccDeploymentConfig_basic(rName, description string) string {
	return acctest.ConfigCompose(testAccRouteConfig_target(rName), fmt.Sprintf(`
resource "aws_apigatewayv2_deployment" "test" {
  api_id      = aws_apigatewayv2_api.test.id
  description = %[1]q

  depends_on = [aws_apigatewayv2_route.test]
}
`, description))
}

func testAccDeploymentConfig_triggers(rName string, apiKeyRequired bool) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name                       = %[1]q
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"
}

resource "aws_apigatewayv2_integration" "test" {
  api_id           = aws_apigatewayv2_api.test.id
  integration_type = "MOCK"
}

resource "aws_apigatewayv2_route" "test" {
  api_id           = aws_apigatewayv2_api.test.id
  api_key_required = %[2]t
  route_key        = "$default"
  target           = "integrations/${aws_apigatewayv2_integration.test.id}"
}

resource "aws_apigatewayv2_deployment" "test" {
  api_id = aws_apigatewayv2_api.test.id

  triggers = {
    redeployment = sha1(jsonencode([
      aws_apigatewayv2_integration.test,
      aws_apigatewayv2_route.test,
    ]))
  }

  lifecycle {
    create_before_destroy = true
  }
}
`, rName, apiKeyRequired)
}
