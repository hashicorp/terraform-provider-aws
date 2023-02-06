package apigateway_test

import (
	"context"
	"fmt"
	"regexp"
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

func TestAccAPIGatewayDeployment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var deployment apigateway.Deployment
	resourceName := "aws_api_gateway_deployment.test"
	restApiResourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_required(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					acctest.MatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(".+/")),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/", acctest.Region()))),
					resource.TestCheckResourceAttrPair(resourceName, "rest_api_id", restApiResourceName, "id"),
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

func TestAccAPIGatewayDeployment_Disappears_restAPI(t *testing.T) {
	ctx := acctest.Context(t)
	var deployment apigateway.Deployment
	var restApi apigateway.RestApi
	resourceName := "aws_api_gateway_deployment.test"
	restApiResourceName := "aws_api_gateway_rest_api.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test-deployment")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_stageName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					testAccCheckRestAPIExists(ctx, restApiResourceName, &restApi),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigateway.ResourceRestAPI(), restApiResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayDeployment_triggers(t *testing.T) {
	ctx := acctest.Context(t)
	var deployment1, deployment2, deployment3, deployment4 apigateway.Deployment
	var stage apigateway.Stage
	resourceName := "aws_api_gateway_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_triggers("description1", "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment1),
					testAccCheckStageExists(ctx, resourceName, &stage),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
					resource.TestCheckResourceAttr(resourceName, "stage_description", "description1"),
				),
				// Due to how the Terraform state is handled for resources during creation,
				// any SHA1 of whole resources will change after first apply, then stabilize.
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccDeploymentConfig_triggers("description1", "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment2),
					testAccCheckDeploymentRecreated(&deployment1, &deployment2),
					testAccCheckStageExists(ctx, resourceName, &stage),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
					resource.TestCheckResourceAttr(resourceName, "stage_description", "description1"),
				),
			},
			{
				Config: testAccDeploymentConfig_triggers("description1", "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment3),
					testAccCheckDeploymentNotRecreated(&deployment2, &deployment3),
					testAccCheckStageExists(ctx, resourceName, &stage),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
					resource.TestCheckResourceAttr(resourceName, "stage_description", "description1"),
				),
			},
			{
				Config: testAccDeploymentConfig_triggers("description2", "https://example.org"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment4),
					testAccCheckDeploymentRecreated(&deployment3, &deployment4),
					testAccCheckStageExists(ctx, resourceName, &stage),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
					resource.TestCheckResourceAttr(resourceName, "stage_description", "description2"),
				),
			},
		},
	})
}

func TestAccAPIGatewayDeployment_description(t *testing.T) {
	ctx := acctest.Context(t)
	var deployment apigateway.Deployment
	resourceName := "aws_api_gateway_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_description("description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				Config: testAccDeploymentConfig_description("description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccAPIGatewayDeployment_stageDescription(t *testing.T) {
	ctx := acctest.Context(t)
	var deployment apigateway.Deployment
	var stage apigateway.Stage
	resourceName := "aws_api_gateway_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_stageDescription("description1"),
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
	var deployment apigateway.Deployment
	var stage apigateway.Stage
	resourceName := "aws_api_gateway_deployment.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test-deployment")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_stageName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					testAccCheckStageExists(ctx, resourceName, &stage),
					resource.TestCheckResourceAttr(resourceName, "stage_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(fmt.Sprintf(".+/%s", rName))),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
				),
			},
			{
				Config: testAccDeploymentConfig_required(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "stage_name"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(".+/")),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/", acctest.Region()))),
				),
			},
		},
	})
}

func TestAccAPIGatewayDeployment_StageName_emptyString(t *testing.T) {
	ctx := acctest.Context(t)
	var deployment apigateway.Deployment
	resourceName := "aws_api_gateway_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_stageName(""),
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
	var deployment apigateway.Deployment
	resourceName := "aws_api_gateway_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_variables("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "variables.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "variables.key1", "value1"),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/28997.
func TestAccAPIGatewayDeployment_conflictingConnectionType(t *testing.T) {
	ctx := acctest.Context(t)
	var deployment apigateway.Deployment
	resourceName := "aws_api_gateway_deployment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, apigateway.EndpointsID),
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

func testAccCheckDeploymentExists(ctx context.Context, n string, res *apigateway.Deployment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway Deployment ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn()

		req := &apigateway.GetDeploymentInput{
			DeploymentId: aws.String(rs.Primary.ID),
			RestApiId:    aws.String(s.RootModule().Resources["aws_api_gateway_rest_api.test"].Primary.ID),
		}
		describe, err := conn.GetDeploymentWithContext(ctx, req)
		if err != nil {
			return err
		}

		if *describe.Id != rs.Primary.ID {
			return fmt.Errorf("APIGateway Deployment not found")
		}

		*res = *describe

		return nil
	}
}

func testAccCheckDeploymentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_deployment" {
				continue
			}

			req := &apigateway.GetDeploymentsInput{
				RestApiId: aws.String(rs.Primary.Attributes["rest_api_id"]),
			}
			describe, err := conn.GetDeploymentsWithContext(ctx, req)

			if tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
				continue
			}

			if err != nil {
				return err
			}

			if describe != nil {
				for _, deployment := range describe.Items {
					if aws.StringValue(deployment.Id) == rs.Primary.ID {
						return fmt.Errorf("API Gateway Deployment still exists")
					}
				}
			}
		}

		return nil
	}
}

func testAccCheckDeploymentNotRecreated(i, j *apigateway.Deployment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.CreatedDate).Equal(aws.TimeValue(j.CreatedDate)) {
			return fmt.Errorf("API Gateway Deployment recreated")
		}

		return nil
	}
}

func testAccCheckDeploymentRecreated(i, j *apigateway.Deployment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreatedDate).Equal(aws.TimeValue(j.CreatedDate)) {
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

func testAccDeploymentConfig_base(uri string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "tf-acc-test-deployment"
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
  uri                     = %[1]q
  integration_http_method = "GET"
}

resource "aws_api_gateway_integration_response" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_integration.test.http_method
  status_code = aws_api_gateway_method_response.error.status_code
}
`, uri)
}

func testAccDeploymentConfig_triggers(description string, url string) string {
	return acctest.ConfigCompose(testAccDeploymentConfig_base(url), fmt.Sprintf(`
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

func testAccDeploymentConfig_description(description string) string {
	return acctest.ConfigCompose(testAccDeploymentConfig_base("http://example.com"), fmt.Sprintf(`
resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  description = %[1]q
  rest_api_id = aws_api_gateway_rest_api.test.id
}
`, description))
}

func testAccDeploymentConfig_required() string {
	return acctest.ConfigCompose(testAccDeploymentConfig_base("http://example.com"), `
resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id = aws_api_gateway_rest_api.test.id
}
`)
}

func testAccDeploymentConfig_stageDescription(stageDescription string) string {
	return acctest.ConfigCompose(testAccDeploymentConfig_base("http://example.com"), fmt.Sprintf(`
resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id       = aws_api_gateway_rest_api.test.id
  stage_description = %[1]q
  stage_name        = "tf-acc-test"
}
`, stageDescription))
}

func testAccDeploymentConfig_stageName(stageName string) string {
	return acctest.ConfigCompose(testAccDeploymentConfig_base("http://example.com"), fmt.Sprintf(`
resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = %[1]q
}
`, stageName))
}

func testAccDeploymentConfig_variables(key1, value1 string) string {
	return acctest.ConfigCompose(testAccDeploymentConfig_base("http://example.com"), fmt.Sprintf(`
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
