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

func TestAccAWSAPIGateway2Deployment_basic(t *testing.T) {
	resourceName := "aws_api_gateway_v2_deployment.test"
	rName := fmt.Sprintf("tf-testacc-apigwv2-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGateway2DeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGateway2DeploymentConfig_basic(rName, "Test description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2DeploymentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "Test description"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGateway2DeploymentImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGateway2DeploymentConfig_basic(rName, "Test description updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2DeploymentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "Test description updated"),
				),
			},
		},
	})
}

func testAccCheckAWSAPIGateway2DeploymentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_v2_deployment" {
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

func testAccCheckAWSAPIGateway2DeploymentExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 deployment ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		_, err := conn.GetDeployment(&apigatewayv2.GetDeploymentInput{
			ApiId:        aws.String(rs.Primary.Attributes["api_id"]),
			DeploymentId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAWSAPIGateway2DeploymentImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["api_id"], rs.Primary.ID), nil
	}
}

func testAccAWSAPIGateway2DeploymentConfig_basic(rName, description string) string {
	return testAccAWSAPIGatewayV2RouteConfig_target(rName) + fmt.Sprintf(`
resource "aws_api_gateway_v2_deployment" "test" {
  api_id      = "${aws_api_gateway_v2_api.test.id}"
  description = %[1]q

  depends_on  = ["aws_api_gateway_v2_route.test"]
}
`, description)
}
