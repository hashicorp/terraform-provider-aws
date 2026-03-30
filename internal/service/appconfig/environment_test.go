// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appconfig_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfappconfig "github.com/hashicorp/terraform-provider-aws/internal/service/appconfig"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppConfigEnvironment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appconfig_environment.test"
	appResourceName := "aws_appconfig_application.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "appconfig", regexache.MustCompile(`application/[0-9a-z]{4,7}/environment/[0-9a-z]{4,7}`)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrApplicationID, appResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "monitor.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrState),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appconfig_environment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfappconfig.ResourceEnvironment, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppConfigEnvironment_updateName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, "tf-acc-test-update")
	resourceName := "aws_appconfig_environment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName),
				),
			},
			{
				Config: testAccEnvironmentConfig_basic(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	description := acctest.RandomWithPrefix(t, "tf-acc-test-update")
	resourceName := "aws_appconfig_environment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_description(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
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
					testAccCheckEnvironmentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
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
					testAccCheckEnvironmentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
				),
			},
		},
	})
}

func TestAccAppConfigEnvironment_monitors(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appconfig_environment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_monitors(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "monitor.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "monitor.*.alarm_arn", "aws_cloudwatch_metric_alarm.test.0", names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "monitor.*.alarm_role_arn", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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
					testAccCheckEnvironmentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "monitor.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "monitor.*.alarm_arn", "aws_cloudwatch_metric_alarm.test.0", names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "monitor.*.alarm_role_arn", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "monitor.*.alarm_arn", "aws_cloudwatch_metric_alarm.test.1", names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "monitor.*.alarm_role_arn", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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
					testAccCheckEnvironmentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "monitor.#", "0"),
				),
			},
		},
	})
}

func TestAccAppConfigEnvironment_multipleEnvironments(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName1 := "aws_appconfig_environment.test"
	resourceName2 := "aws_appconfig_environment.test2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName1),
					testAccCheckEnvironmentExists(ctx, t, resourceName2),
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
					testAccCheckEnvironmentExists(ctx, t, resourceName1),
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

func TestAccAppConfigEnvironment_frameworkMigration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appconfig_environment.test"
	description := "Description"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.AppConfigServiceID),
		CheckDestroy: testAccCheckEnvironmentDestroy(ctx, t),
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
					testAccCheckEnvironmentExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccEnvironmentConfig_description(rName, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccAppConfigEnvironment_frameworkMigration_monitors(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appconfig_environment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.AppConfigServiceID),
		CheckDestroy: testAccCheckEnvironmentDestroy(ctx, t),
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
					testAccCheckEnvironmentExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccEnvironmentConfig_monitors(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func testAccCheckEnvironmentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AppConfigClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appconfig_environment" {
				continue
			}

			_, err := tfappconfig.FindEnvironmentByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrApplicationID], rs.Primary.Attributes["environment_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AppConfig Environment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckEnvironmentExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AppConfigClient(ctx)

		_, err := tfappconfig.FindEnvironmentByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrApplicationID], rs.Primary.Attributes["environment_id"])

		return err
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
