package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSDataSourceIAMPolicy_basic(t *testing.T) {
	datasourceName := "data.aws_iam_policy.test"
	resourceName := "aws_iam_policy.test"
	policyName := fmt.Sprintf("test-policy-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, iam.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDataSourceIamPolicyConfig(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "path", resourceName, "path"),
					resource.TestCheckResourceAttrPair(datasourceName, "policy", resourceName, "policy"),
					resource.TestCheckResourceAttrPair(datasourceName, "policy_id", resourceName, "policy_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags", resourceName, "tags"),
				),
			},
		},
	})

}

func TestAccAWSDataSourceIAMPolicy_withName(t *testing.T) {
	policyName := fmt.Sprintf("test-policy-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDataSourceIamPolicyConfig_withName(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_iam_policy.test", "name", policyName),
					resource.TestCheckResourceAttr("data.aws_iam_policy.test", "description", "My test policy"),
					resource.TestCheckResourceAttr("data.aws_iam_policy.test", "path", "/"),
					resource.TestCheckResourceAttrSet("data.aws_iam_policy.test", "policy"),
					resource.TestMatchResourceAttr("data.aws_iam_policy.test", "arn",
						regexp.MustCompile(`^arn:[\w-]+:([a-zA-Z0-9\-])+:([a-z]{2}-(gov-)?[a-z]+-\d{1})?:(\d{12})?:(.*)$`)),
				),
			},
		},
	})

}

func testAccAwsDataSourceIamPolicyConfig(policyName string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name        = "%s"
  path        = "/"
  description = "My test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

data "aws_iam_policy" "test" {
  arn = aws_iam_policy.test.arn
}
`, policyName)
}

func testAccAwsDataSourceIamPolicyConfig_withName(policyName string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test_policy" {
    name = "%s"
    path = "/"
    description = "My test policy"
    policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

data "aws_iam_policy" "test" {
  name = "${aws_iam_policy.test_policy.name}"
}
`, policyName)
}
