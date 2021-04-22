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
				Config: testAccAwsDataSourceIamPolicyConfig(policyName, "/"),
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

func TestAccAWSDataSourceIAMPolicy_Name(t *testing.T) {
	datasourceName := "data.aws_iam_policy.test"
	resourceName := "aws_iam_policy.test"
	policyName := fmt.Sprintf("test-policy-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, iam.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDataSourceIamPolicyConfig_Name(policyName, "/"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "name", policyName),
					resource.TestCheckResourceAttr(datasourceName, "description", "My test policy"),
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

func TestAccAWSDataSourceIAMPolicy_PathPrefix(t *testing.T) {
	datasourceName := "data.aws_iam_policy.test"
	resourceName := "aws_iam_policy.test"

	policyName := fmt.Sprintf("test-policy-%s", acctest.RandString(10))
	policyPath := "/test-path/"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, iam.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDataSourceIamPolicyConfig_PathPrefix(policyName, policyPath),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "name", policyName),
					resource.TestCheckResourceAttr(datasourceName, "description", "My test policy"),
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

func testAccAwsDataSourceIamPolicyBaseConfig(policyName, policyPath string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name        = %q
  path        = %q
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
}`, policyName, policyPath)
}

func testAccAwsDataSourceIamPolicyConfig(policyName, policyPath string) string {
	return composeConfig(
		testAccAwsDataSourceIamPolicyBaseConfig(policyName, policyPath),
		`
data "aws_iam_policy" "test" {
  arn = aws_iam_policy.test.arn
}
`)
}

func testAccAwsDataSourceIamPolicyConfig_Name(policyName, policyPath string) string {
	return composeConfig(
		testAccAwsDataSourceIamPolicyBaseConfig(policyName, policyPath),
		`
data "aws_iam_policy" "test" {
  name = aws_iam_policy.test.name
}
`)
}

func testAccAwsDataSourceIamPolicyConfig_PathPrefix(policyName, policyPath string) string {
	return composeConfig(
		testAccAwsDataSourceIamPolicyBaseConfig(policyName, policyPath),
		fmt.Sprintf(`
data "aws_iam_policy" "test" {
  name        = aws_iam_policy.test.name
  path_prefix = %q
}
`, policyPath))
}
