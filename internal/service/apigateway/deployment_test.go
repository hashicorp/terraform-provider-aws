package apigateway_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
)

func TestAccAWSAPIGatewayDeployment_basic(t *testing.T) {
	var deployment apigateway.Deployment
	resourceName := "aws_api_gateway_deployment.test"
	restApiResourceName := "aws_api_gateway_rest_api.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test-deployment")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:   acctest.ErrorCheck(t, apigateway.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentStageNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(resourceName, &deployment),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					acctest.MatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(fmt.Sprintf(".+/%s", rName))),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttrPair(resourceName, "rest_api_id", restApiResourceName, "id"),
					resource.TestCheckNoResourceAttr(resourceName, "stage_description"),
					resource.TestCheckResourceAttr(resourceName, "stage_name", rName),
					resource.TestCheckNoResourceAttr(resourceName, "variables.%"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayDeployment_disappears_RestApi(t *testing.T) {
	var deployment apigateway.Deployment
	var restApi apigateway.RestApi
	resourceName := "aws_api_gateway_deployment.test"
	restApiResourceName := "aws_api_gateway_rest_api.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test-deployment")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:   acctest.ErrorCheck(t, apigateway.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentStageNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(resourceName, &deployment),
					testAccCheckRestAPIExists(restApiResourceName, &restApi),
					acctest.CheckResourceDisappears(acctest.Provider, tfapigateway.ResourceRestAPI(), restApiResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayDeployment_Triggers(t *testing.T) {
	var deployment1, deployment2, deployment3, deployment4 apigateway.Deployment
	var stage apigateway.Stage
	resourceName := "aws_api_gateway_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:   acctest.ErrorCheck(t, apigateway.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentTriggersConfig("description1", "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(resourceName, &deployment1),
					testAccCheckDeploymentStageExists(resourceName, &stage),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
					resource.TestCheckResourceAttr(resourceName, "stage_description", "description1"),
				),
				// Due to how the Terraform state is handled for resources during creation,
				// any SHA1 of whole resources will change after first apply, then stabilize.
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccDeploymentTriggersConfig("description1", "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(resourceName, &deployment2),
					testAccCheckDeploymentRecreated(&deployment1, &deployment2),
					testAccCheckDeploymentStageExists(resourceName, &stage),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
					resource.TestCheckResourceAttr(resourceName, "stage_description", "description1"),
				),
			},
			{
				Config: testAccDeploymentTriggersConfig("description1", "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(resourceName, &deployment3),
					testAccCheckDeploymentNotRecreated(&deployment2, &deployment3),
					testAccCheckDeploymentStageExists(resourceName, &stage),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
					resource.TestCheckResourceAttr(resourceName, "stage_description", "description1"),
				),
			},
			{
				Config: testAccDeploymentTriggersConfig("description2", "https://example.org"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(resourceName, &deployment4),
					testAccCheckDeploymentRecreated(&deployment3, &deployment4),
					testAccCheckDeploymentStageExists(resourceName, &stage),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
					resource.TestCheckResourceAttr(resourceName, "stage_description", "description2"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayDeployment_Description(t *testing.T) {
	var deployment apigateway.Deployment
	resourceName := "aws_api_gateway_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:   acctest.ErrorCheck(t, apigateway.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentDescriptionConfig("description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				Config: testAccDeploymentDescriptionConfig("description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayDeployment_StageDescription(t *testing.T) {
	var deployment apigateway.Deployment
	var stage apigateway.Stage
	resourceName := "aws_api_gateway_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:   acctest.ErrorCheck(t, apigateway.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentStageDescriptionConfig("description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(resourceName, &deployment),
					testAccCheckDeploymentStageExists(resourceName, &stage),
					resource.TestCheckResourceAttr(resourceName, "stage_description", "description1"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayDeployment_StageName(t *testing.T) {
	var deployment apigateway.Deployment
	var stage apigateway.Stage
	resourceName := "aws_api_gateway_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:   acctest.ErrorCheck(t, apigateway.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentStageNameConfig("test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(resourceName, &deployment),
					testAccCheckDeploymentStageExists(resourceName, &stage),
					resource.TestCheckResourceAttr(resourceName, "stage_name", "test"),
				),
			},
			{
				Config: testAccDeploymentRequiredConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "stage_name"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayDeployment_StageName_EmptyString(t *testing.T) {
	var deployment apigateway.Deployment
	resourceName := "aws_api_gateway_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:   acctest.ErrorCheck(t, apigateway.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentStageNameConfig(""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "stage_name", ""),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayDeployment_Variables(t *testing.T) {
	var deployment apigateway.Deployment
	resourceName := "aws_api_gateway_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:   acctest.ErrorCheck(t, apigateway.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentVariablesConfig("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "variables.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "variables.key1", "value1"),
				),
			},
		},
	})
}

func testAccCheckDeploymentExists(n string, res *apigateway.Deployment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway Deployment ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

		req := &apigateway.GetDeploymentInput{
			DeploymentId: aws.String(rs.Primary.ID),
			RestApiId:    aws.String(s.RootModule().Resources["aws_api_gateway_rest_api.test"].Primary.ID),
		}
		describe, err := conn.GetDeployment(req)
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

func testAccCheckDeploymentStageExists(resourceName string, res *apigateway.Stage) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Deployment not found: %s", resourceName)
		}

		req := &apigateway.GetStageInput{
			StageName: aws.String(rs.Primary.Attributes["stage_name"]),
			RestApiId: aws.String(rs.Primary.Attributes["rest_api_id"]),
		}
		stage, err := conn.GetStage(req)
		if err != nil {
			return err
		}

		*res = *stage

		return nil
	}
}

func testAccCheckDeploymentDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_deployment" {
			continue
		}

		req := &apigateway.GetDeploymentsInput{
			RestApiId: aws.String(rs.Primary.Attributes["rest_api_id"]),
		}
		describe, err := conn.GetDeployments(req)

		if tfawserr.ErrMessageContains(err, apigateway.ErrCodeNotFoundException, "") {
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

func testAccDeploymentBaseConfig(uri string) string {
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
  uri                     = "%s"
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

func testAccDeploymentTriggersConfig(description string, url string) string {
	return testAccDeploymentBaseConfig(url) + fmt.Sprintf(`
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
`, description)
}

func testAccDeploymentDescriptionConfig(description string) string {
	return testAccDeploymentBaseConfig("http://example.com") + fmt.Sprintf(`
resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  description = %q
  rest_api_id = aws_api_gateway_rest_api.test.id
}
`, description)
}

func testAccDeploymentRequiredConfig() string {
	return testAccDeploymentBaseConfig("http://example.com") + `
resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id = aws_api_gateway_rest_api.test.id
}
`
}

func testAccDeploymentStageDescriptionConfig(stageDescription string) string {
	return testAccDeploymentBaseConfig("http://example.com") + fmt.Sprintf(`
resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id       = aws_api_gateway_rest_api.test.id
  stage_description = %q
  stage_name        = "tf-acc-test"
}
`, stageDescription)
}

func testAccDeploymentStageNameConfig(stageName string) string {
	return testAccDeploymentBaseConfig("http://example.com") + fmt.Sprintf(`
resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = %q
}
`, stageName)
}

func testAccDeploymentVariablesConfig(key1, value1 string) string {
	return testAccDeploymentBaseConfig("http://example.com") + fmt.Sprintf(`
resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id = aws_api_gateway_rest_api.test.id

  variables = {
    %q = %q
  }
}
`, key1, value1)
}
