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

func TestAccAWSAppConfigDeploymentStrategy_basic(t *testing.T) {
	var strategy appconfig.GetDeploymentStrategyOutput
	resourceName := "aws_appconfig_deployment_strategy.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigDeploymentStrategyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigDeploymentStrategy(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigDeploymentStrategyExists(resourceName, &strategy),
					testAccCheckAWSAppConfigDeploymentStrategyARN(resourceName, &strategy),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", "deployment strategy description"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAppConfigDeploymentStrategyImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAppConfigDeploymentStrategy_disappears(t *testing.T) {
	var strategy appconfig.GetDeploymentStrategyOutput
	resourceName := "aws_appconfig_deployment_strategy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigDeploymentStrategyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigDeploymentStrategy(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigDeploymentStrategyExists(resourceName, &strategy),
					testAccCheckAWSAppConfigDeploymentStrategyDisappears(&strategy),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSAppConfigDeploymentStrategy_Tags(t *testing.T) {
	var strategy appconfig.GetDeploymentStrategyOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appconfig_deployment_strategy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigDeploymentStrategyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigDeploymentStrategyTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigDeploymentStrategyExists(resourceName, &strategy),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAppConfigDeploymentStrategyImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAppConfigDeploymentStrategyTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigDeploymentStrategyExists(resourceName, &strategy),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSAppConfigDeploymentStrategyTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigDeploymentStrategyExists(resourceName, &strategy),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAppConfigDeploymentStrategyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appconfigconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appconfig_deployment_strategy" {
			continue
		}

		input := &appconfig.GetDeploymentStrategyInput{
			DeploymentStrategyId: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetDeploymentStrategy(input)

		if isAWSErr(err, appconfig.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil {
			return fmt.Errorf("AppConfig DeploymentStrategy (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSAppConfigDeploymentStrategyDisappears(strategy *appconfig.GetDeploymentStrategyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).appconfigconn

		_, err := conn.DeleteDeploymentStrategy(&appconfig.DeleteDeploymentStrategyInput{
			DeploymentStrategyId: strategy.Id,
		})

		return err
	}
}

func testAccCheckAWSAppConfigDeploymentStrategyExists(resourceName string, strategy *appconfig.GetDeploymentStrategyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).appconfigconn

		output, err := conn.GetDeploymentStrategy(&appconfig.GetDeploymentStrategyInput{
			DeploymentStrategyId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*strategy = *output

		return nil
	}
}

func testAccCheckAWSAppConfigDeploymentStrategyARN(resourceName string, strategy *appconfig.GetDeploymentStrategyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appconfig", fmt.Sprintf("deploymentstrategy/%s", aws.StringValue(strategy.Id)))(s)
	}
}

func testAccAWSAppConfigDeploymentStrategy() string {
	return fmt.Sprintf(`
resource "aws_appconfig_deployment_strategy" "test" {
  name                           = "%s"
  description                    = "deployment strategy description"
  deployment_duration_in_minutes = 3
  final_bake_time_in_minutes     = 4
  growth_factor                  = 10
  growth_type                    = "LINEAR"
  replicate_to                   = "NONE"
  tags = {
    Env = "Test"
  }
}
`, acctest.RandomWithPrefix("tf-acc-test"))
}

func testAccAWSAppConfigDeploymentStrategyTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_appconfig_deployment_strategy" "test" {
  name                           = %[1]q
  deployment_duration_in_minutes = 3
  final_bake_time_in_minutes     = 4
  growth_factor                  = 10
  growth_type                    = "LINEAR"
  replicate_to                   = "NONE"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSAppConfigDeploymentStrategyTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSAppConfigApplicationTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2) + fmt.Sprintf(`
resource "aws_appconfig_deployment_strategy" "test" {
  name                           = %[1]q
  deployment_duration_in_minutes = 3
  final_bake_time_in_minutes     = 4
  growth_factor                  = 10
  growth_type                    = "LINEAR"
  replicate_to                   = "NONE"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSAppConfigDeploymentStrategyImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return rs.Primary.ID, nil
	}
}
