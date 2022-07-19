package appconfig_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappconfig "github.com/hashicorp/terraform-provider-aws/internal/service/appconfig"
)

func TestAccAppConfigDeploymentStrategy_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_deployment_strategy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDeploymentStrategyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentStrategyConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentStrategyExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "appconfig", regexp.MustCompile(`deploymentstrategy/[a-z0-9]{4,7}`)),
					resource.TestCheckResourceAttr(resourceName, "deployment_duration_in_minutes", "3"),
					resource.TestCheckResourceAttr(resourceName, "growth_factor", "10"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "replicate_to", appconfig.ReplicateToNone),
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

func TestAccAppConfigDeploymentStrategy_updateDescription(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := sdkacctest.RandomWithPrefix("tf-acc-test-update")
	resourceName := "aws_appconfig_deployment_strategy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDeploymentStrategyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentStrategyConfig_description(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentStrategyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
				),
			},
			{
				Config: testAccDeploymentStrategyConfig_description(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentStrategyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", description),
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

func TestAccAppConfigDeploymentStrategy_updateFinalBakeTime(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_deployment_strategy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDeploymentStrategyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentStrategyConfig_finalBakeTime(rName, 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentStrategyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "final_bake_time_in_minutes", "60"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeploymentStrategyConfig_finalBakeTime(rName, 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentStrategyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "final_bake_time_in_minutes", "30"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test FinalBakeTimeInMinutes Removal
				Config: testAccDeploymentStrategyConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentStrategyExists(resourceName),
				),
			},
		},
	})
}

func TestAccAppConfigDeploymentStrategy_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_deployment_strategy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDeploymentStrategyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentStrategyConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentStrategyExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfappconfig.ResourceDeploymentStrategy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppConfigDeploymentStrategy_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_deployment_strategy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDeploymentStrategyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentStrategyConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentStrategyExists(resourceName),
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
				Config: testAccDeploymentStrategyConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentStrategyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDeploymentStrategyConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentStrategyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckDeploymentStrategyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppConfigConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appconfig_deployment_strategy" {
			continue
		}

		input := &appconfig.GetDeploymentStrategyInput{
			DeploymentStrategyId: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetDeploymentStrategy(input)

		if tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting Appconfig Deployment Strategy (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("AppConfig Deployment Strategy (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckDeploymentStrategyExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppConfigConn

		input := &appconfig.GetDeploymentStrategyInput{
			DeploymentStrategyId: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetDeploymentStrategy(input)

		if err != nil {
			return fmt.Errorf("error getting Appconfig Deployment Strategy (%s): %w", rs.Primary.ID, err)
		}

		if output == nil {
			return fmt.Errorf("AppConfig Deployment Strategy (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDeploymentStrategyConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_appconfig_deployment_strategy" "test" {
  name                           = %[1]q
  deployment_duration_in_minutes = 3
  growth_factor                  = 10
  replicate_to                   = "NONE"
}
`, rName)
}

func testAccDeploymentStrategyConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_appconfig_deployment_strategy" "test" {
  name                           = %[1]q
  deployment_duration_in_minutes = 3
  description                    = %[2]q
  growth_factor                  = 10
  replicate_to                   = "NONE"
}
`, rName, description)
}

func testAccDeploymentStrategyConfig_finalBakeTime(rName string, time int) string {
	return fmt.Sprintf(`
resource "aws_appconfig_deployment_strategy" "test" {
  name                           = %[1]q
  deployment_duration_in_minutes = 3
  final_bake_time_in_minutes     = %[2]d
  growth_factor                  = 10
  replicate_to                   = "NONE"
}
`, rName, time)
}

func testAccDeploymentStrategyConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_appconfig_deployment_strategy" "test" {
  name                           = %[1]q
  deployment_duration_in_minutes = 3
  growth_factor                  = 10
  replicate_to                   = "NONE"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDeploymentStrategyConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_appconfig_deployment_strategy" "test" {
  name                           = %[1]q
  deployment_duration_in_minutes = 3
  growth_factor                  = 10
  replicate_to                   = "NONE"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
