package iot_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iot"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccIoTLoggingOptions_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":  testAccLoggingOptions_basic,
		"update": testAccLoggingOptions_update,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccLoggingOptions_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_logging_options.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingOptionsConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_log_level", "WARN"),
					resource.TestCheckResourceAttr(resourceName, "disable_all_logs", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
				),
			},
		},
	})
}

func testAccLoggingOptions_update(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_logging_options.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingOptionsConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_log_level", "WARN"),
					resource.TestCheckResourceAttr(resourceName, "disable_all_logs", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
				),
			},
			{
				Config: testAccLoggingOptionsUpdatedConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_log_level", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "disable_all_logs", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
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

func testAccLoggingOptionsConfig(rName string) string {
	return acctest.ConfigCompose(testAccLoggingOptionsBaseConfig(rName), `
resource "aws_iot_logging_options" "test" {
  default_log_level = "WARN"
  role_arn          = aws_iam_role.test.arn

  depends_on = [aws_iam_role_policy.test]
}
`)
}

func testAccLoggingOptionsUpdatedConfig(rName string) string {
	return acctest.ConfigCompose(testAccLoggingOptionsBaseConfig(rName), `
resource "aws_iot_logging_options" "test" {
  default_log_level = "DISABLED"
  disable_all_logs  = true
  role_arn          = aws_iam_role.test.arn

  depends_on = [aws_iam_role_policy.test]
}
`)
}
