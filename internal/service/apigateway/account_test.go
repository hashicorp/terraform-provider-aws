package apigateway_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/apigateway"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccAPIGatewayAccount_basic(t *testing.T) {
	var conf apigateway.Account

	rInt := sdkacctest.RandInt()
	firstName := fmt.Sprintf("tf_acc_api_gateway_cloudwatch_%d", rInt)
	secondName := fmt.Sprintf("tf_acc_api_gateway_cloudwatch_modified_%d", rInt)
	resourceName := "aws_api_gateway_account.test"
	expectedRoleArn_first := regexp.MustCompile("role/" + firstName + "$")
	expectedRoleArn_second := regexp.MustCompile("role/" + secondName + "$")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_updated(firstName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(resourceName, &conf),
					testAccCheckAccountCloudWatchRoleARN(&conf, expectedRoleArn_first),
					acctest.MatchResourceAttrGlobalARN(resourceName, "cloudwatch_role_arn", "iam", expectedRoleArn_first),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"cloudwatch_role_arn"},
			},
			{
				Config: testAccAccountConfig_updated2(secondName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(resourceName, &conf),
					testAccCheckAccountCloudWatchRoleARN(&conf, expectedRoleArn_second),
					acctest.MatchResourceAttrGlobalARN(resourceName, "cloudwatch_role_arn", "iam", expectedRoleArn_second),
				),
			},
			{
				Config: testAccAccountConfig_empty,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(resourceName, &conf),
					// This resource does not un-set the value, so this will preserve the CloudWatch role ARN setting on the
					// deployed resource, but will be empty in the Terraform state
					testAccCheckAccountCloudWatchRoleARN(&conf, expectedRoleArn_second),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_role_arn", ""),
				),
			},
		},
	})
}

func testAccCheckAccountCloudWatchRoleARN(conf *apigateway.Account, expectedArn *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if expectedArn == nil && conf.CloudwatchRoleArn == nil {
			return nil
		}
		if expectedArn == nil && conf.CloudwatchRoleArn != nil {
			return fmt.Errorf("Expected empty CloudwatchRoleArn, given: %q", *conf.CloudwatchRoleArn)
		}
		if expectedArn != nil && conf.CloudwatchRoleArn == nil {
			return fmt.Errorf("Empty CloudwatchRoleArn, expected: %q", expectedArn)
		}
		if !expectedArn.MatchString(*conf.CloudwatchRoleArn) {
			return fmt.Errorf("CloudwatchRoleArn didn't match. Expected: %q, Given: %q", expectedArn, *conf.CloudwatchRoleArn)
		}
		return nil
	}
}

func testAccCheckAccountExists(n string, res *apigateway.Account) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway Account ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

		req := &apigateway.GetAccountInput{}
		describe, err := conn.GetAccount(req)
		if err != nil {
			return err
		}
		if describe == nil {
			return fmt.Errorf("Got nil account ?!")
		}

		*res = *describe

		return nil
	}
}

func testAccCheckAccountDestroy(s *terraform.State) error {
	// Intentionally noop
	// as there is no API method for deleting or resetting account settings
	return nil
}

const testAccAccountConfig_empty = `
resource "aws_api_gateway_account" "test" {
}
`

func testAccAccountConfig_updated(randName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_account" "test" {
  cloudwatch_role_arn = aws_iam_role.cloudwatch.arn
}

resource "aws_iam_role" "cloudwatch" {
  name = "%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "apigateway.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "cloudwatch" {
  name = "default"
  role = aws_iam_role.cloudwatch.id

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:DescribeLogGroups",
                "logs:DescribeLogStreams",
                "logs:PutLogEvents",
                "logs:GetLogEvents",
                "logs:FilterLogEvents"
            ],
            "Resource": "*"
        }
    ]
}
EOF
}
`, randName)
}

func testAccAccountConfig_updated2(randName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_account" "test" {
  cloudwatch_role_arn = aws_iam_role.second.arn
}

resource "aws_iam_role" "second" {
  name = "%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "apigateway.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "cloudwatch" {
  name = "default"
  role = aws_iam_role.second.id

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:DescribeLogGroups",
                "logs:DescribeLogStreams",
                "logs:PutLogEvents",
                "logs:GetLogEvents",
                "logs:FilterLogEvents"
            ],
            "Resource": "*"
        }
    ]
}
EOF
}
`, randName)
}
