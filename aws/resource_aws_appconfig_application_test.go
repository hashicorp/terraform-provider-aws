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
	resource.AddTestSweepers("aws_appconfig_application", &resource.Sweeper{
		Name: "aws_appconfig_application",
		F:    testSweepAppConfigApplications,
		Dependencies: []string{
			"aws_appconfig_configuration_profile",
			"aws_appconfig_environment",
		},
	})
}

func testSweepAppConfigApplications(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).appconfigconn
	sweepResources := make([]*testSweepResource, 0)
	var errs *multierror.Error

	input := &appconfig.ListApplicationsInput{}

	err = conn.ListApplicationsPages(input, func(page *appconfig.ListApplicationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, item := range page.Items {
			if item == nil {
				continue
			}

			id := aws.StringValue(item.Id)

			log.Printf("[INFO] Deleting AppConfig Application (%s)", id)
			r := resourceAwsAppconfigApplication()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing AppConfig Applications: %w", err))
	}

	if err = testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AppConfig Applications for %s: %w", region, err))
	}

	if testSweepSkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping AppConfig Applications sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSAppConfigApplication_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appconfig_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigApplicationConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigApplicationExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "appconfig", regexp.MustCompile(`application/[a-z0-9]{4,7}`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccAWSAppConfigApplication_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appconfig_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigApplicationConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigApplicationExists(resourceName),
					acctest.CheckResourceDisappears(testAccProvider, resourceAwsAppconfigApplication(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSAppConfigApplication_updateName(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rNameUpdated := sdkacctest.RandomWithPrefix("tf-acc-test-update")
	resourceName := "aws_appconfig_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigApplicationConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigApplicationExists(resourceName),
				),
			},
			{
				Config: testAccAWSAppConfigApplicationConfigName(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigApplicationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
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

func TestAccAWSAppConfigApplication_updateDescription(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	description := sdkacctest.RandomWithPrefix("tf-acc-test-update")
	resourceName := "aws_appconfig_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigApplicationConfigDescription(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigApplicationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAppConfigApplicationConfigDescription(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigApplicationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", description),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test Description Removal
				Config: testAccAWSAppConfigApplicationConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigApplicationExists(resourceName),
				),
			},
		},
	})
}

func TestAccAWSAppConfigApplication_Tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appconfig_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigApplicationTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigApplicationExists(resourceName),
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
				Config: testAccAWSAppConfigApplicationTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigApplicationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSAppConfigApplicationTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigApplicationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAppConfigApplicationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appconfigconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appconfig_application" {
			continue
		}

		input := &appconfig.GetApplicationInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetApplication(input)

		if tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading AppConfig Application (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("AppConfig Application (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSAppConfigApplicationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).appconfigconn

		input := &appconfig.GetApplicationInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetApplication(input)

		if err != nil {
			return fmt.Errorf("error reading AppConfig Application (%s): %w", rs.Primary.ID, err)
		}

		if output == nil {
			return fmt.Errorf("AppConfig Application (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAWSAppConfigApplicationConfigName(rName string) string {
	return fmt.Sprintf(`
resource "aws_appconfig_application" "test" {
  name = %[1]q
}
`, rName)
}

func testAccAWSAppConfigApplicationConfigDescription(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_appconfig_application" "test" {
  name        = %q
  description = %q
}
`, rName, description)
}

func testAccAWSAppConfigApplicationTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_appconfig_application" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSAppConfigApplicationTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_appconfig_application" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
