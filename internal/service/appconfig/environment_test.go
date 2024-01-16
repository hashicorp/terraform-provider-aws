// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappconfig "github.com/hashicorp/terraform-provider-aws/internal/service/appconfig"
)

func TestAccAppConfigEnvironment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_environment.test"
	appResourceName := "aws_appconfig_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "appconfig", regexache.MustCompile(`application/[0-9a-z]{4,7}/environment/[0-9a-z]{4,7}`)),
					resource.TestCheckResourceAttrPair(resourceName, "application_id", appResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
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

func TestAccAppConfigEnvironment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfappconfig.ResourceEnvironmentFW, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppConfigEnvironment_updateName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix("tf-acc-test-update")
	resourceName := "aws_appconfig_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName),
				),
			},
			{
				Config: testAccEnvironmentConfig_basic(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName),
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

func TestAccAppConfigEnvironment_updateDescription(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := sdkacctest.RandomWithPrefix("tf-acc-test-update")
	resourceName := "aws_appconfig_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_description(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEnvironmentConfig_description(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName),
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
				Config: testAccEnvironmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
		},
	})
}

func TestAccAppConfigEnvironment_monitors(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_monitors(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName),
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
				Config: testAccEnvironmentConfig_monitors(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName),
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
				Config: testAccEnvironmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "monitor.#", "0"),
				),
			},
		},
	})
}

func TestAccAppConfigEnvironment_multipleEnvironments(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName1 := "aws_appconfig_environment.test"
	resourceName2 := "aws_appconfig_environment.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName1),
					testAccCheckEnvironmentExists(ctx, resourceName2),
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
				Config: testAccEnvironmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName1),
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

func TestAccAppConfigEnvironment_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName),
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
				Config: testAccEnvironmentConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccEnvironmentConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAppConfigEnvironment_tagsWithNullValue(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_tagsWithNullValue(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckNoResourceAttr(resourceName, "tags.key2"),
				),
				// ~ tags           = {
				// 	~ "key2" = "" -> null
				// 	  # (1 unchanged element hidden)
				//   }
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppConfigEnvironment_frameworkMigration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_environment.test"
	description := "Description"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, appconfig.EndpointsID),
		CheckDestroy: testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.3.0",
					},
				},
				Config: testAccEnvironmentConfig_description(rName, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccEnvironmentConfig_description(rName, description),
				PlanOnly:                 true,
			},
		},
	})
}

func TestAccAppConfigEnvironment_frameworkMigration_monitors(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, appconfig.EndpointsID),
		CheckDestroy: testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.3.0",
					},
				},
				Config: testAccEnvironmentConfig_monitors(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccEnvironmentConfig_monitors(rName, 2),
				PlanOnly:                 true,
			},
		},
	})
}

func testAccCheckEnvironmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppConfigConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appconfig_environment" {
				continue
			}

			appID := rs.Primary.Attributes["application_id"]
			envID := rs.Primary.Attributes["environment_id"]

			input := &appconfig.GetEnvironmentInput{
				ApplicationId: aws.String(appID),
				EnvironmentId: aws.String(envID),
			}

			output, err := conn.GetEnvironmentWithContext(ctx, input)

			if tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
				continue
			}

			if err != nil {
				return fmt.Errorf("reading AppConfig Environment (%s) for Application (%s): %w", envID, appID, err)
			}

			if output != nil {
				return fmt.Errorf("AppConfig Environment (%s) for Application (%s) still exists", envID, appID)
			}
		}

		return nil
	}
}

func testAccCheckEnvironmentExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		appID := rs.Primary.Attributes["application_id"]
		envID := rs.Primary.Attributes["environment_id"]

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppConfigConn(ctx)

		input := &appconfig.GetEnvironmentInput{
			ApplicationId: aws.String(appID),
			EnvironmentId: aws.String(envID),
		}

		output, err := conn.GetEnvironmentWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("reading AppConfig Environment (%s) for Application (%s): %w", envID, appID, err)
		}

		if output == nil {
			return fmt.Errorf("AppConfig Environment (%s) for Application (%s) not found", envID, appID)
		}

		return nil
	}
}

func testAccEnvironmentConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccApplicationConfig_name(rName), fmt.Sprintf(`
resource "aws_appconfig_environment" "test" {
  name           = %[1]q
  application_id = aws_appconfig_application.test.id
}
`, rName))
}

func testAccEnvironmentConfig_description(rName, description string) string {
	return acctest.ConfigCompose(testAccApplicationConfig_name(rName), fmt.Sprintf(`
resource "aws_appconfig_environment" "test" {
  name           = %[1]q
  description    = %[2]q
  application_id = aws_appconfig_application.test.id
}
`, rName, description))
}

func testAccEnvironmentConfig_monitors(rName string, count int) string {
	return acctest.ConfigCompose(testAccApplicationConfig_name(rName), fmt.Sprintf(`
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
    for_each = aws_cloudwatch_metric_alarm.test[*].arn
    content {
      alarm_arn      = monitor.value
      alarm_role_arn = aws_iam_role.test.arn
    }
  }
}
`, rName, count))
}

func testAccEnvironmentConfig_multiple(rName string) string {
	return acctest.ConfigCompose(testAccApplicationConfig_name(rName), fmt.Sprintf(`
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

func testAccEnvironmentConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccApplicationConfig_name(rName), fmt.Sprintf(`
resource "aws_appconfig_environment" "test" {
  name           = %[1]q
  application_id = aws_appconfig_application.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccEnvironmentConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccApplicationConfig_name(rName), fmt.Sprintf(`
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

func testAccEnvironmentConfig_tagsWithNullValue(rName string) string {
	return acctest.ConfigCompose(testAccApplicationConfig_name(rName), fmt.Sprintf(`
resource "aws_appconfig_environment" "test" {
  name           = %[1]q
  application_id = aws_appconfig_application.test.id

  tags = {
    key1 = "value1"
    key2 = null
  }
}
`, rName))
}
