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

func TestAccAWSAppConfigEnvironment_basic(t *testing.T) {
	var environment appconfig.GetEnvironmentOutput
	roleName := acctest.RandomWithPrefix("tf-acc-test")
	alarmName := acctest.RandomWithPrefix("tf-acc-test")
	appName := acctest.RandomWithPrefix("tf-acc-test")
	appDesc := acctest.RandomWithPrefix("desc")
	envName := acctest.RandomWithPrefix("tf-acc-test")
	envDesc := acctest.RandomWithPrefix("desc")
	resourceName := "aws_appconfig_environment.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigEnvironmentWithMonitors(roleName, alarmName, appName, appDesc, envName, envDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigEnvironmentExists(resourceName, &environment),
					resource.TestCheckResourceAttr(resourceName, "name", envName),
					testAccCheckAWSAppConfigEnvironmentARN(resourceName, &environment),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", envDesc),
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
	var environment appconfig.GetEnvironmentOutput

	appName := acctest.RandomWithPrefix("tf-acc-test")
	appDesc := acctest.RandomWithPrefix("desc")
	envName := acctest.RandomWithPrefix("tf-acc-test")
	envDesc := acctest.RandomWithPrefix("desc")
	resourceName := "aws_appconfig_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigEnvironment(appName, appDesc, envName, envDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigEnvironmentExists(resourceName, &environment),
					testAccCheckAWSAppConfigEnvironmentDisappears(&environment),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSAppConfigEnvironment_Tags(t *testing.T) {
	var environment appconfig.GetEnvironmentOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appconfig_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigEnvironmentTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigEnvironmentExists(resourceName, &environment),
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
					testAccCheckAWSAppConfigEnvironmentExists(resourceName, &environment),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSAppConfigEnvironmentTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigEnvironmentExists(resourceName, &environment),
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

		input := &appconfig.GetEnvironmentInput{
			ApplicationId: aws.String(rs.Primary.Attributes["application_id"]),
			EnvironmentId: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetEnvironment(input)

		if isAWSErr(err, appconfig.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil {
			return fmt.Errorf("AppConfig Environment (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSAppConfigEnvironmentDisappears(environment *appconfig.GetEnvironmentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).appconfigconn

		_, err := conn.DeleteEnvironment(&appconfig.DeleteEnvironmentInput{
			ApplicationId: environment.ApplicationId,
			EnvironmentId: environment.Id,
		})

		return err
	}
}

func testAccCheckAWSAppConfigEnvironmentExists(resourceName string, environment *appconfig.GetEnvironmentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).appconfigconn

		output, err := conn.GetEnvironment(&appconfig.GetEnvironmentInput{
			ApplicationId: aws.String(rs.Primary.Attributes["application_id"]),
			EnvironmentId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*environment = *output

		return nil
	}
}

func testAccCheckAWSAppConfigEnvironmentARN(resourceName string, environment *appconfig.GetEnvironmentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appconfig", fmt.Sprintf("application/%s/environment/%s", aws.StringValue(environment.ApplicationId), aws.StringValue(environment.Id)))(s)
	}
}

func testAccAWSAppConfigEnvironmentWithMonitors(roleName, alarmName, appName, appDesc, envName, envDesc string) string {
	return testAccAWSAppConfigMonitor_ServiceRole(roleName) + testAccAWSCloudWatchMetricAlarmConfig(alarmName) + testAccAWSAppConfigApplicationName(appName, appDesc) + fmt.Sprintf(`
resource "aws_appconfig_environment" "test" {
  name           = %[1]q
  description    = %[2]q
  application_id = aws_appconfig_application.test.id

  monitors {
    alarm_arn      = aws_cloudwatch_metric_alarm.test.arn
    alarm_role_arn = aws_iam_role.test.arn
  }
}
`, envName, envDesc)
}

func testAccAWSAppConfigEnvironment(appName, appDesc, envName, envDesc string) string {
	return testAccAWSAppConfigApplicationName(appName, appDesc) + fmt.Sprintf(`
resource "aws_appconfig_environment" "test" {
  name           = %[1]q
  description    = %[2]q
  application_id = aws_appconfig_application.test.id
}
`, envName, envDesc)
}

func testAccAWSAppConfigEnvironmentTags1(rName, tagKey1, tagValue1 string) string {
	return testAccAWSAppConfigApplicationTags1(rName, tagKey1, tagValue1) + fmt.Sprintf(`
resource "aws_appconfig_environment" "test" {
  name           = %[1]q
  application_id = aws_appconfig_application.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSAppConfigEnvironmentTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSAppConfigApplicationTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2) + fmt.Sprintf(`
resource "aws_appconfig_environment" "test" {
  name           = %[1]q
  application_id = aws_appconfig_application.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSAppConfigMonitor_ServiceRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = "%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "appconfig.amazonaws.com"
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
`, rName)
}
