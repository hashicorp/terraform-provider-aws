package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appconfig"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
)

func TestAccAWSAppConfigDeployment_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appconfig_deployment.test"
	appResourceName := "aws_appconfig_application.test"
	confProfResourceName := "aws_appconfig_configuration_profile.test"
	depStrategyResourceName := "aws_appconfig_deployment_strategy.test"
	envResourceName := "aws_appconfig_environment.test"
	confVersionResourceName := "aws_appconfig_hosted_configuration_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, appconfig.EndpointsID),
		Providers:  acctest.Providers,
		// AppConfig Deployments cannot be destroyed, but we want to ensure
		// the Application and its dependents are removed.
		CheckDestroy: testAccCheckAppConfigApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigDeploymentConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigDeploymentExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "appconfig", regexp.MustCompile(`application/[a-z0-9]{4,7}/environment/[a-z0-9]{4,7}/deployment/1`)),
					resource.TestCheckResourceAttrPair(resourceName, "application_id", appResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_profile_id", confProfResourceName, "configuration_profile_id"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_version", confVersionResourceName, "version_number"),
					resource.TestCheckResourceAttr(resourceName, "deployment_number", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "deployment_strategy_id", depStrategyResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttrPair(resourceName, "environment_id", envResourceName, "environment_id"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSAppConfigDeployment_PredefinedStrategy(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appconfig_deployment.test"
	strategy := "AppConfig.Linear50PercentEvery30Seconds"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, appconfig.EndpointsID),
		Providers:  acctest.Providers,
		// AppConfig Deployments cannot be destroyed, but we want to ensure
		// the Application and its dependents are removed.
		CheckDestroy: testAccCheckAppConfigApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigDeploymentConfig_PredefinedStrategy(rName, strategy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigDeploymentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "deployment_strategy_id", strategy),
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

func TestAccAWSAppConfigDeployment_Tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appconfig_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appconfig.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigDeploymentTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigDeploymentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAppConfigDeploymentTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigDeploymentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSAppConfigDeploymentTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigDeploymentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
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

		appID, envID, deploymentNum, err := resourceAwsAppconfigDeploymentParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppConfigConn

		input := &appconfig.GetDeploymentInput{
			ApplicationId:    aws.String(appID),
			DeploymentNumber: aws.Int64(int64(deploymentNum)),
			EnvironmentId:    aws.String(envID),
		}

		output, err := conn.GetDeployment(input)

		if err != nil {
			return fmt.Errorf("error getting Appconfig Deployment (%s): %w", rs.Primary.ID, err)
		}

		if output == nil {
			return fmt.Errorf("AppConfig Deployment (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAWSAppConfigDeploymentConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_appconfig_application" "test" {
  name = %[1]q
}

resource "aws_appconfig_environment" "test" {
  name           = %[1]q
  application_id = aws_appconfig_application.test.id
}

resource "aws_appconfig_configuration_profile" "test" {
  application_id = aws_appconfig_application.test.id
  name           = %[1]q
  location_uri   = "hosted"
}

resource "aws_appconfig_deployment_strategy" "test" {
  name                           = %[1]q
  deployment_duration_in_minutes = 3
  growth_factor                  = 10
  replicate_to                   = "NONE"
}

resource "aws_appconfig_hosted_configuration_version" "test" {
  application_id           = aws_appconfig_application.test.id
  configuration_profile_id = aws_appconfig_configuration_profile.test.configuration_profile_id
  content_type             = "application/json"

  content = jsonencode({
    foo = "bar"
  })

  description = %[1]q
}
`, rName)
}

func testAccAWSAppConfigDeploymentConfigName(rName string) string {
	return acctest.ConfigCompose(
		testAccAWSAppConfigDeploymentConfigBase(rName),
		fmt.Sprintf(`
resource "aws_appconfig_deployment" "test"{
  application_id           = aws_appconfig_application.test.id
  configuration_profile_id = aws_appconfig_configuration_profile.test.configuration_profile_id
  configuration_version    = aws_appconfig_hosted_configuration_version.test.version_number
  description              = %[1]q
  deployment_strategy_id   = aws_appconfig_deployment_strategy.test.id
  environment_id           = aws_appconfig_environment.test.environment_id
}
`, rName))
}

func testAccAWSAppConfigDeploymentConfig_PredefinedStrategy(rName, strategy string) string {
	return acctest.ConfigCompose(
		testAccAWSAppConfigDeploymentConfigBase(rName),
		fmt.Sprintf(`
resource "aws_appconfig_deployment" "test"{
  application_id           = aws_appconfig_application.test.id
  configuration_profile_id = aws_appconfig_configuration_profile.test.configuration_profile_id
  configuration_version    = aws_appconfig_hosted_configuration_version.test.version_number
  description              = %[1]q
  deployment_strategy_id   = %[2]q
  environment_id           = aws_appconfig_environment.test.environment_id
}
`, rName, strategy))
}

func testAccAWSAppConfigDeploymentTags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccAWSAppConfigDeploymentConfigBase(rName),
		fmt.Sprintf(`
resource "aws_appconfig_deployment" "test"{
  application_id           = aws_appconfig_application.test.id
  configuration_profile_id = aws_appconfig_configuration_profile.test.configuration_profile_id
  configuration_version    = aws_appconfig_hosted_configuration_version.test.version_number
  deployment_strategy_id   = aws_appconfig_deployment_strategy.test.id
  environment_id           = aws_appconfig_environment.test.environment_id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccAWSAppConfigDeploymentTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccAWSAppConfigDeploymentConfigBase(rName),
		fmt.Sprintf(`
resource "aws_appconfig_deployment" "test"{
  application_id           = aws_appconfig_application.test.id
  configuration_profile_id = aws_appconfig_configuration_profile.test.configuration_profile_id
  configuration_version    = aws_appconfig_hosted_configuration_version.test.version_number
  deployment_strategy_id   = aws_appconfig_deployment_strategy.test.id
  environment_id           = aws_appconfig_environment.test.environment_id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
