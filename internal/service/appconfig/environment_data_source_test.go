// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppConfigEnvironmentDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	appName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	appResourceName := "aws_appconfig_application.test"
	dataSourceName := "data.aws_appconfig_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppConfigEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentDataSourceConfig_basic(appName, rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrRegionalARN(dataSourceName, names.AttrARN, "appconfig", regexache.MustCompile(`application/([a-z\d]{4,7})/environment/+.`)),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrApplicationID, appResourceName, names.AttrID),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDescription, "Example AppConfig Environment"),
					resource.TestMatchResourceAttr(dataSourceName, "environment_id", regexache.MustCompile(`[a-z\d]{4,7}`)),
					resource.TestCheckResourceAttr(dataSourceName, "monitor.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "monitor.*.alarm_arn", "aws_cloudwatch_metric_alarm.test", names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "monitor.*.alarm_role_arn", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrState),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
		},
	})
}

func testAccEnvironmentDataSourceConfig_basic(appName, rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_name(appName),
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
  alarm_name                = "%[1]s"
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
  application_id = aws_appconfig_application.test.id
  name           = %[1]q
  description    = "Example AppConfig Environment"

  monitor {
    alarm_arn      = aws_cloudwatch_metric_alarm.test.arn
    alarm_role_arn = aws_iam_role.test.arn
  }
}

data "aws_appconfig_environment" "test" {
  application_id = aws_appconfig_application.test.id
  environment_id = aws_appconfig_environment.test.environment_id
}
`, rName))
}
