package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func init() {
	resource.AddTestSweepers("aws_appconfig_deployment_strategy", &resource.Sweeper{
		Name: "aws_appconfig_deployment_strategy",
		F:    testSweepAppConfigDeploymentStrategies,
	})
}

func testSweepAppConfigDeploymentStrategies(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).appconfigconn
	sweepResources := make([]*testSweepResource, 0)
	var errs *multierror.Error

	input := &appconfig.ListDeploymentStrategiesInput{}

	err = conn.ListDeploymentStrategiesPages(input, func(page *appconfig.ListDeploymentStrategiesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, item := range page.Items {
			if item == nil {
				continue
			}

			id := aws.StringValue(item.Id)

			// Deleting AppConfig Predefined Strategies is not supported; returns BadRequestException
			if regexp.MustCompile(`^AppConfig\.[A-Za-z0-9]{9,40}$`).MatchString(id) {
				log.Printf("[DEBUG] Skipping AppConfig Deployment Strategy (%s): predefined strategy cannot be deleted", id)
				continue
			}

			log.Printf("[INFO] Deleting AppConfig Deployment Strategy (%s)", id)
			r := resourceAwsAppconfigDeploymentStrategy()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing AppConfig Deployment Strategies: %w", err))
	}

	if err = testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AppConfig Deployment Strategies for %s: %w", region, err))
	}

	if testSweepSkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping AppConfig Deployment Strategies sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSAppConfigDeploymentStrategy_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appconfig_deployment_strategy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigDeploymentStrategyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigDeploymentStrategyConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigDeploymentStrategyExists(resourceName),
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

func TestAccAWSAppConfigDeploymentStrategy_updateDescription(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	description := sdkacctest.RandomWithPrefix("tf-acc-test-update")
	resourceName := "aws_appconfig_deployment_strategy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigDeploymentStrategyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigDeploymentStrategyConfigDescription(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigDeploymentStrategyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
				),
			},
			{
				Config: testAccAWSAppConfigDeploymentStrategyConfigDescription(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigDeploymentStrategyExists(resourceName),
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

func TestAccAWSAppConfigDeploymentStrategy_updateFinalBakeTime(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appconfig_deployment_strategy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigDeploymentStrategyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigDeploymentStrategyConfigFinalBakeTime(rName, 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigDeploymentStrategyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "final_bake_time_in_minutes", "60"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAppConfigDeploymentStrategyConfigFinalBakeTime(rName, 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigDeploymentStrategyExists(resourceName),
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
				Config: testAccAWSAppConfigDeploymentStrategyConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigDeploymentStrategyExists(resourceName),
				),
			},
		},
	})
}

func TestAccAWSAppConfigDeploymentStrategy_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appconfig_deployment_strategy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigDeploymentStrategyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigDeploymentStrategyConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigDeploymentStrategyExists(resourceName),
					acctest.CheckResourceDisappears(testAccProvider, resourceAwsAppconfigDeploymentStrategy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSAppConfigDeploymentStrategy_Tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appconfig_deployment_strategy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigDeploymentStrategyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigDeploymentStrategyTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigDeploymentStrategyExists(resourceName),
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
				Config: testAccAWSAppConfigDeploymentStrategyTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigDeploymentStrategyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSAppConfigDeploymentStrategyTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigDeploymentStrategyExists(resourceName),
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

func testAccCheckAWSAppConfigDeploymentStrategyExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).appconfigconn

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

func testAccAWSAppConfigDeploymentStrategyConfigName(rName string) string {
	return fmt.Sprintf(`
resource "aws_appconfig_deployment_strategy" "test" {
  name                           = %[1]q
  deployment_duration_in_minutes = 3
  growth_factor                  = 10
  replicate_to                   = "NONE"
}
`, rName)
}

func testAccAWSAppConfigDeploymentStrategyConfigDescription(rName, description string) string {
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

func testAccAWSAppConfigDeploymentStrategyConfigFinalBakeTime(rName string, time int) string {
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

func testAccAWSAppConfigDeploymentStrategyTags1(rName, tagKey1, tagValue1 string) string {
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

func testAccAWSAppConfigDeploymentStrategyTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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
