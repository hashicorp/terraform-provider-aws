// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
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

func TestAccAPIGatewayDeployment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var deployment apigateway.GetDeploymentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_deployment.test"
	restApiResourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					acctest.MatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexache.MustCompile(".+/")),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexache.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/", acctest.Region()))),
					resource.TestCheckResourceAttrPair(resourceName, "rest_api_id", restApiResourceName, names.AttrID),
					resource.TestCheckNoResourceAttr(resourceName, "stage_description"),
					resource.TestCheckNoResourceAttr(resourceName, "stage_name"),
					resource.TestCheckNoResourceAttr(resourceName, "variables.%"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccDeploymentImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayDeployment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var deployment apigateway.GetDeploymentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigateway.ResourceDeployment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayDeployment_Disappears_restAPI(t *testing.T) {
	ctx := acctest.Context(t)
	var deployment apigateway.GetDeploymentOutput
	var restApi apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_deployment.test"
	restApiResourceName := "aws_api_gateway_rest_api.test"
	stageName := sdkacctest.RandomWithPrefix("tf-acc-test-deployment")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_stageName(rName, stageName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					testAccCheckRESTAPIExists(ctx, restApiResourceName, &restApi),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigateway.ResourceRestAPI(), restApiResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayDeployment_triggers(t *testing.T) {
	ctx := acctest.Context(t)
	var deployment1, deployment2, deployment3, deployment4 apigateway.GetDeploymentOutput
	var stage apigateway.GetStageOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_triggers(rName, "description1", "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment1),
					testAccCheckStageExists(ctx, resourceName, &stage),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
					resource.TestCheckResourceAttr(resourceName, "stage_description", "description1"),
				),
				// Due to how the Terraform state is handled for resources during creation,
				// any SHA1 of whole resources will change after first apply, then stabilize.
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccDeploymentConfig_triggers(rName, "description1", "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment2),
					testAccCheckDeploymentRecreated(&deployment1, &deployment2),
					testAccCheckStageExists(ctx, resourceName, &stage),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
					resource.TestCheckResourceAttr(resourceName, "stage_description", "description1"),
				),
			},
			{
				Config: testAccDeploymentConfig_triggers(rName, "description1", "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment3),
					testAccCheckDeploymentNotRecreated(&deployment2, &deployment3),
					testAccCheckStageExists(ctx, resourceName, &stage),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
					resource.TestCheckResourceAttr(resourceName, "stage_description", "description1"),
				),
			},
			{
				Config: testAccDeploymentConfig_triggers(rName, "description2", "https://example.org"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment4),
					testAccCheckDeploymentRecreated(&deployment3, &deployment4),
					testAccCheckStageExists(ctx, resourceName, &stage),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
					resource.TestCheckResourceAttr(resourceName, "stage_description", "description2"),
				),
			},
		},
	})
}

func TestAccAPIGatewayDeployment_description(t *testing.T) {
	ctx := acctest.Context(t)
	var deployment apigateway.GetDeploymentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				Config: testAccDeploymentConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func TestAccAPIGatewayDeployment_stageDescription(t *testing.T) {
	ctx := acctest.Context(t)
	var deployment apigateway.GetDeploymentOutput
	var stage apigateway.GetStageOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_stageDescription(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					testAccCheckStageExists(ctx, resourceName, &stage),
					resource.TestCheckResourceAttr(resourceName, "stage_description", "description1"),
				),
			},
		},
	})
}

func TestAccAPIGatewayDeployment_stageName(t *testing.T) {
	ctx := acctest.Context(t)
	var deployment apigateway.GetDeploymentOutput
	var stage apigateway.GetStageOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_deployment.test"
	stageName := sdkacctest.RandomWithPrefix("tf-acc-test-deployment")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_stageName(rName, stageName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					testAccCheckStageExists(ctx, resourceName, &stage),
					resource.TestCheckResourceAttr(resourceName, "stage_name", stageName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexache.MustCompile(fmt.Sprintf(".+/%s", stageName))),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexache.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), stageName))),
				),
			},
			{
				Config: testAccDeploymentConfig_required(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "stage_name"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexache.MustCompile(".+/")),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexache.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/", acctest.Region()))),
				),
			},
		},
	})
}

func TestAccAPIGatewayDeployment_StageName_emptyString(t *testing.T) {
	ctx := acctest.Context(t)
	var deployment apigateway.GetDeploymentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_stageName(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "stage_name", ""),
				),
			},
		},
	})
}

func TestAccAPIGatewayDeployment_variables(t *testing.T) {
	ctx := acctest.Context(t)
	var deployment apigateway.GetDeploymentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_variables(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "variables.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "variables.key1", acctest.CtValue1),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/28997.
func TestAccAPIGatewayDeployment_conflictingConnectionType(t *testing.T) {
	ctx := acctest.Context(t)
	var deployment apigateway.GetDeploymentOutput
	resourceName := "aws_api_gateway_deployment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_conflictingConnectionType(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayDeployment_deploymentCanarySettings(t *testing.T) {
	ctx := acctest.Context(t)
	var deployment apigateway.GetDeploymentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	url := "https://example.com"
	resourceName := "aws_api_gateway_deployment.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_deploymentCanarySettings(rName, url),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "variables.one", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "canary_settings.0.percent_traffic", "33.33"),
					resource.TestCheckResourceAttr(resourceName, "canary_settings.0.stage_variable_overrides.one", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "canary_settings.0.use_stage_cache", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccCheckDeploymentExists(ctx context.Context, n string, v *apigateway.GetDeploymentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		output, err := tfapigateway.FindDeploymentByTwoPartKey(ctx, conn, rs.Primary.Attributes["rest_api_id"], rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckDeploymentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_deployment" {
				continue
			}

			_, err := tfapigateway.FindDeploymentByTwoPartKey(ctx, conn, rs.Primary.Attributes["rest_api_id"], rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway Deployment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDeploymentNotRecreated(i, j *apigateway.GetDeploymentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.ToTime(i.CreatedDate).Equal(aws.ToTime(j.CreatedDate)) {
			return fmt.Errorf("API Gateway Deployment recreated")
		}

		return nil
	}
}

func testAccCheckDeploymentRecreated(i, j *apigateway.GetDeploymentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToTime(i.CreatedDate).Equal(aws.ToTime(j.CreatedDate)) {
			return fmt.Errorf("API Gateway Deployment not recreated")
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

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["rest_api_id"], rs.Primary.ID), nil
	}
}

func testAccDeploymentConfig_base(rName, uri string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
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

resource "aws_api_gateway_method_response" "test" {
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
  uri                     = %[2]q
  integration_http_method = "GET"
}

resource "aws_api_gateway_integration_response" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_integration.test.http_method
  status_code = aws_api_gateway_method_response.test.status_code
}
`, rName, uri)
}

func testAccDeploymentConfig_triggers(rName, description, url string) string {
	return acctest.ConfigCompose(testAccDeploymentConfig_base(rName, url), fmt.Sprintf(`
resource "aws_api_gateway_deployment" "test" {
  description       = %[1]q
  rest_api_id       = aws_api_gateway_rest_api.test.id
  stage_description = %[1]q
  stage_name        = "tf-acc-test"

  triggers = {
    redeployment = sha1(jsonencode(aws_api_gateway_integration.test))
  }

  lifecycle {
    create_before_destroy = true
  }
}
`, description))
}

func testAccDeploymentConfig_description(rName, description string) string {
	return acctest.ConfigCompose(testAccDeploymentConfig_base(rName, "http://example.com"), fmt.Sprintf(`
resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  description = %[1]q
  rest_api_id = aws_api_gateway_rest_api.test.id
}
`, description))
}

func testAccDeploymentConfig_required(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentConfig_base(rName, "http://example.com"), `
resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id = aws_api_gateway_rest_api.test.id
}
`)
}

func testAccDeploymentConfig_stageDescription(rName, stageDescription string) string {
	return acctest.ConfigCompose(testAccDeploymentConfig_base(rName, "http://example.com"), fmt.Sprintf(`
resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id       = aws_api_gateway_rest_api.test.id
  stage_description = %[1]q
  stage_name        = "tf-acc-test"
}
`, stageDescription))
}

func testAccDeploymentConfig_stageName(rName, stageName string) string {
	return acctest.ConfigCompose(testAccDeploymentConfig_base(rName, "http://example.com"), fmt.Sprintf(`
resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = %[1]q
}
`, stageName))
}

func testAccDeploymentConfig_variables(rName, key1, value1 string) string {
	return acctest.ConfigCompose(testAccDeploymentConfig_base(rName, "http://example.com"), fmt.Sprintf(`
resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id = aws_api_gateway_rest_api.test.id

  variables = {
    %[1]q = %[2]q
  }
}
`, key1, value1))
}

func testAccDeploymentConfig_conflictingConnectionType(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLambdaBase(rName, rName, rName), fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
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
resource "aws_api_gateway_deployment" "test" {
  description = "The deployment"
  rest_api_id = aws_api_gateway_rest_api.test.id
  triggers = {
    redeployment = sha1(join(",", tolist([
      jsonencode(aws_api_gateway_integration.test),
    ])))
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.test.invoke_arn
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "index.handler"
  runtime       = "nodejs16.x"
}
`, rName))
}

func testAccStageConfig_deploymentCanarySettings(rName, url string) string {
	return acctest.ConfigCompose(testAccDeploymentConfig_base(rName, url), `
resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id = aws_api_gateway_rest_api.test.id
  canary_settings {
    percent_traffic = "33.33"
    stage_variable_overrides = {
      one = "3"
    }
    use_stage_cache = "true"
  }
  variables = {
    one = "1"
    two = "2"
  }
}
`)
}
