// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIoTLoggingOptions_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic: testAccLoggingOptions_basic,
		"update":        testAccLoggingOptions_update,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccLoggingOptions_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_logging_options.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingOptionsConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_log_level", "WARN"),
					resource.TestCheckResourceAttr(resourceName, "disable_all_logs", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
				),
			},
		},
	})
}

func testAccLoggingOptions_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_logging_options.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingOptionsConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_log_level", "WARN"),
					resource.TestCheckResourceAttr(resourceName, "disable_all_logs", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
				),
			},
			{
				Config: testAccLoggingOptionsConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_log_level", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "disable_all_logs", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
				),
			},
		},
	})
}

func testAccLoggingOptionsBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {"Service": "iot.amazonaws.com"},
    "Effect": "Allow"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:PutMetricFilter",
      "logs:PutRetentionPolicy"
    ],
    "Resource": ["*"]
  }]
}
EOF
}
`, rName)
}

func testAccLoggingOptionsConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccLoggingOptionsBaseConfig(rName), `
resource "aws_iot_logging_options" "test" {
  default_log_level = "WARN"
  role_arn          = aws_iam_role.test.arn

  depends_on = [aws_iam_role_policy.test]
}
`)
}

func testAccLoggingOptionsConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccLoggingOptionsBaseConfig(rName), `
resource "aws_iot_logging_options" "test" {
  default_log_level = "DISABLED"
  disable_all_logs  = true
  role_arn          = aws_iam_role.test.arn

  depends_on = [aws_iam_role_policy.test]
}
`)
}
