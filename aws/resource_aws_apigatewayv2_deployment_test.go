package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSAPIGatewayV2Deployment_basic(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetDeploymentOutput
	resourceName := "aws_apigatewayv2_deployment.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2DeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2DeploymentConfig_basic(rName, "Test description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2DeploymentExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deployed", "false"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test description"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2DeploymentImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayV2DeploymentConfig_basic(rName, "Test description updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2DeploymentExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deployed", "false"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test description updated"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Deployment_disappears(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetDeploymentOutput
	resourceName := "aws_apigatewayv2_deployment.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2DeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2DeploymentConfig_basic(rName, "Test description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2DeploymentExists(resourceName, &apiId, &v),
					testAccCheckAWSAPIGatewayV2DeploymentDisappears(&apiId, &v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSAPIGatewayV2DeploymentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_apigatewayv2_deployment" {
			continue
		}

		_, err := conn.GetDeployment(&apigatewayv2.GetDeploymentInput{
			ApiId:        aws.String(rs.Primary.Attributes["api_id"]),
			DeploymentId: aws.String(rs.Primary.ID),
		})
		if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("API Gateway v2 deployment %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSAPIGatewayV2DeploymentDisappears(apiId *string, v *apigatewayv2.GetDeploymentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		_, err := conn.DeleteDeployment(&apigatewayv2.DeleteDeploymentInput{
			ApiId:        apiId,
			DeploymentId: v.DeploymentId,
		})

		return err
	}
}

func testAccCheckAWSAPIGatewayV2DeploymentExists(n string, vApiId *string, v *apigatewayv2.GetDeploymentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 deployment ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		apiId := aws.String(rs.Primary.Attributes["api_id"])
		resp, err := conn.GetDeployment(&apigatewayv2.GetDeploymentInput{
			ApiId:        apiId,
			DeploymentId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*vApiId = *apiId
		*v = *resp

		return nil
	}
}

func testAccAWSAPIGatewayV2DeploymentImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["api_id"], rs.Primary.ID), nil
	}
}

func testAccAWSAPIGatewayV2DeploymentConfig_basic(rName, description string) string {
	return testAccAWSAPIGatewayV2RouteConfig_target(rName) + fmt.Sprintf(`
resource "aws_apigatewayv2_deployment" "test" {
  api_id      = "${aws_apigatewayv2_api.test.id}"
  description = %[1]q

  depends_on  = ["aws_apigatewayv2_route.test"]
}
`, description)
}
