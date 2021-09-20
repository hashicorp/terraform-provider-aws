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
	resource.AddTestSweepers("aws_appconfig_environment", &resource.Sweeper{
		Name: "aws_appconfig_environment",
		F:    testSweepAppConfigEnvironments,
	})
}

func testSweepAppConfigEnvironments(region string) error {
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

			appId := aws.StringValue(item.Id)

			envInput := &appconfig.ListEnvironmentsInput{
				ApplicationId: item.Id,
			}

			err := conn.ListEnvironmentsPages(envInput, func(page *appconfig.ListEnvironmentsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, item := range page.Items {
					if item == nil {
						continue
					}

					id := fmt.Sprintf("%s:%s", aws.StringValue(item.Id), appId)

					log.Printf("[INFO] Deleting AppConfig Environment (%s)", id)
					r := resourceAwsAppconfigEnvironment()
					d := r.Data(nil)
					d.SetId(id)

					sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
				}

				return !lastPage
			})

			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("error listing AppConfig Environments for Application (%s): %w", appId, err))
			}
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing AppConfig Applications: %w", err))
	}

	if err = testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AppConfig Environments for %s: %w", region, err))
	}

	if testSweepSkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping AppConfig Environments sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSAppConfigEnvironment_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appconfig_environment.test"
	appResourceName := "aws_appconfig_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigEnvironmentConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigEnvironmentExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "appconfig", regexp.MustCompile(`application/[a-z0-9]{4,7}/environment/[a-z0-9]{4,7}`)),
					resource.TestCheckResourceAttrPair(resourceName, "application_id", appResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "monitor.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccAWSAppConfigEnvironment_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appconfig_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigEnvironmentConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigEnvironmentExists(resourceName),
					acctest.CheckResourceDisappears(testAccProvider, resourceAwsAppconfigEnvironment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSAppConfigEnvironment_updateName(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rNameUpdated := sdkacctest.RandomWithPrefix("tf-acc-test-update")
	resourceName := "aws_appconfig_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigEnvironmentConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigEnvironmentExists(resourceName),
				),
			},
			{
				Config: testAccAWSAppConfigEnvironmentConfigBasic(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigEnvironmentExists(resourceName),
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

func TestAccAWSAppConfigEnvironment_updateDescription(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	description := sdkacctest.RandomWithPrefix("tf-acc-test-update")
	resourceName := "aws_appconfig_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigEnvironmentConfigDescription(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigEnvironmentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAppConfigEnvironmentConfigDescription(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigEnvironmentExists(resourceName),
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
				Config: testAccAWSAppConfigEnvironmentConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigEnvironmentExists(resourceName),
				),
			},
		},
	})
}

func TestAccAWSAppConfigEnvironment_Monitors(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appconfig_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigEnvironmentWithMonitors(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigEnvironmentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "monitor.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "monitor.*.alarm_arn", "aws_cloudwatch_metric_alarm.test.0", "arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "monitor.*.alarm_role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAppConfigEnvironmentWithMonitors(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigEnvironmentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "monitor.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "monitor.*.alarm_arn", "aws_cloudwatch_metric_alarm.test.0", "arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "monitor.*.alarm_role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "monitor.*.alarm_arn", "aws_cloudwatch_metric_alarm.test.1", "arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "monitor.*.alarm_role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test Monitor Removal
				Config: testAccAWSAppConfigEnvironmentConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigEnvironmentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "monitor.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSAppConfigEnvironment_MultipleEnvironments(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName1 := "aws_appconfig_environment.test"
	resourceName2 := "aws_appconfig_environment.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigEnvironmentConfigMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigEnvironmentExists(resourceName1),
					testAccCheckAWSAppConfigEnvironmentExists(resourceName2),
				),
			},
			{
				ResourceName:      resourceName1,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      resourceName2,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAppConfigEnvironmentConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigEnvironmentExists(resourceName1),
				),
			},
			{
				ResourceName:      resourceName1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAppConfigEnvironment_Tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appconfig_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigEnvironmentTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigEnvironmentExists(resourceName),
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
				Config: testAccAWSAppConfigEnvironmentTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigEnvironmentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSAppConfigEnvironmentTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigEnvironmentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAppConfigEnvironmentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appconfigconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appconfig_environment" {
			continue
		}

		envID, appID, err := resourceAwsAppconfigEnvironmentParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		input := &appconfig.GetEnvironmentInput{
			ApplicationId: aws.String(appID),
			EnvironmentId: aws.String(envID),
		}

		output, err := conn.GetEnvironment(input)

		if tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading AppConfig Environment (%s) for Application (%s): %w", envID, appID, err)
		}

		if output != nil {
			return fmt.Errorf("AppConfig Environment (%s) for Application (%s) still exists", envID, appID)
		}
	}

	return nil
}

func testAccCheckAWSAppConfigEnvironmentExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		envID, appID, err := resourceAwsAppconfigEnvironmentParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).appconfigconn

		input := &appconfig.GetEnvironmentInput{
			ApplicationId: aws.String(appID),
			EnvironmentId: aws.String(envID),
		}

		output, err := conn.GetEnvironment(input)

		if err != nil {
			return fmt.Errorf("error reading AppConfig Environment (%s) for Application (%s): %w", envID, appID, err)
		}

		if output == nil {
			return fmt.Errorf("AppConfig Environment (%s) for Application (%s) not found", envID, appID)
		}

		return nil
	}
}

func testAccAWSAppConfigEnvironmentConfigBasic(rName string) string {
	return acctest.ConfigCompose(
		testAccAWSAppConfigApplicationConfigName(rName),
		fmt.Sprintf(`
resource "aws_appconfig_environment" "test" {
  name           = %q
  application_id = aws_appconfig_application.test.id
}
`, rName))
}

func testAccAWSAppConfigEnvironmentConfigDescription(rName, description string) string {
	return acctest.ConfigCompose(
		testAccAWSAppConfigApplicationConfigName(rName),
		fmt.Sprintf(`
resource "aws_appconfig_environment" "test" {
  name           = %q
  description    = %q
  application_id = aws_appconfig_application.test.id
}
`, rName, description))
}

func testAccAWSAppConfigEnvironmentWithMonitors(rName string, count int) string {
	return acctest.ConfigCompose(
		testAccAWSAppConfigApplicationConfigName(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "appconfig.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "cloudwatch:DescribeAlarms"
            ],
            "Resource": "*"
        }
    ]
}
POLICY
}

resource "aws_cloudwatch_metric_alarm" "test" {
  count = %[2]d

  alarm_name                = "%[1]s-${count.index}"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = "120"
  statistic                 = "Average"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abc123"
  }
}

resource "aws_appconfig_environment" "test" {
  name           = %[1]q
  application_id = aws_appconfig_application.test.id

  dynamic "monitor" {
    for_each = aws_cloudwatch_metric_alarm.test.*.arn
    content {
      alarm_arn      = monitor.value
      alarm_role_arn = aws_iam_role.test.arn
    }
  }
}
`, rName, count))
}

func testAccAWSAppConfigEnvironmentConfigMultiple(rName string) string {
	return acctest.ConfigCompose(
		testAccAWSAppConfigApplicationConfigName(rName),
		fmt.Sprintf(`
resource "aws_appconfig_environment" "test" {
  name           = %[1]q
  application_id = aws_appconfig_application.test.id
}

resource "aws_appconfig_environment" "test2" {
  name           = "%[1]s-2"
  application_id = aws_appconfig_application.test.id
}
`, rName))
}

func testAccAWSAppConfigEnvironmentTags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccAWSAppConfigApplicationConfigName(rName),
		fmt.Sprintf(`
resource "aws_appconfig_environment" "test" {
  name           = %[1]q
  application_id = aws_appconfig_application.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccAWSAppConfigEnvironmentTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccAWSAppConfigApplicationConfigName(rName),
		fmt.Sprintf(`
resource "aws_appconfig_environment" "test" {
  name           = %[1]q
  application_id = aws_appconfig_application.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
