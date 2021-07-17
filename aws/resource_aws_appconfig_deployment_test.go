package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSAppConfigDeployment_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appconfig_deployment.test"
	appResourceName := "aws_appconfig_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigConfigurationProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigDeployment_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigDeploymentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_id", appResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "description", "My test deployment"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSAppConfigDeploymentExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}
		appID, envID, depNum, err := parseAwsAppconfigDeploymentID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).appconfigconn

		output, err := conn.GetDeployment(&appconfig.GetDeploymentInput{
			ApplicationId:    aws.String(appID),
			EnvironmentId:    aws.String(envID),
			DeploymentNumber: aws.Int64(depNum),
		})

		if err != nil {
			return fmt.Errorf("error reading AppConfig Deployment (%s): %w", rs.Primary.ID, err)
		}

		if output == nil {
			return fmt.Errorf("AppConfig Deployment (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAWSAppConfigDeployment_basic(rName string) string {
	return composeConfig(
		testAccAWSAppConfigConfigurationProfileConfigName(rName),
		`resource "aws_appconfig_environment" "test" {
  name           = "test"
  application_id = aws_appconfig_application.test.id
}

resource "aws_appconfig_hosted_configuration_version" "test" {
  application_id           = aws_appconfig_application.test.id
  configuration_profile_id = aws_appconfig_configuration_profile.test.configuration_profile_id
  content_type             = "application/json"

  content = jsonencode({
    foo = "bar"
  })
}

resource "aws_appconfig_deployment" "test" {
  application_id           = aws_appconfig_application.test.id
  configuration_profile_id = aws_appconfig_configuration_profile.test.configuration_profile_id
  configuration_version    = aws_appconfig_hosted_configuration_version.test.version_number
  deployment_strategy_id   = aws_appconfig_deployment_strategy.test.id
  description              = "My test deployment"
  environment_id           = aws_appconfig_environment.test.environment_id
  tags = {
    Env = "test"
  }
}
`,
		fmt.Sprintf(`
resource "aws_appconfig_deployment_strategy" "test" {
  name                           = "%s"
  deployment_duration_in_minutes = 0
  final_bake_time_in_minutes     = 0
  growth_factor                  = 100
  replicate_to                   = "NONE"
}
`, rName))
}
