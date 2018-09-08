package aws

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAWSPolicy_namePrefix(t *testing.T) {
	var out iam.GetPolicyOutput

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSPolicyPrefixNameConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPolicyExists("aws_iam_policy.policy", &out),
					testAccCheckAWSPolicyGeneratedNamePrefix(
						"aws_iam_policy.policy", "test-policy-"),
				),
			},
		},
	})
}

func TestAWSPolicy_invalidJson(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSPolicyInvalidJsonConfig,
				ExpectError: regexp.MustCompile("invalid JSON"),
			},
		},
	})
}

func TestAWSPolicy_checkDescription(t *testing.T) {
	var out iam.GetPolicyOutput
	rName := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSPolicyDescriptionCheckConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPolicyExists("aws_iam_policy.policy", &out),
					resource.TestCheckResourceAttrSet("aws_iam_policy.policy", "description"),
				),
			},
		},
	})
}

func testAccCheckAWSPolicyExists(resource string, res *iam.GetPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Policy name is set")
		}

		iamconn := testAccProvider.Meta().(*AWSClient).iamconn

		resp, err := iamconn.GetPolicy(&iam.GetPolicyInput{
			PolicyArn: aws.String(rs.Primary.Attributes["arn"]),
		})
		if err != nil {
			return err
		}

		*res = *resp

		return nil
	}
}

func testAccCheckAWSPolicyGeneratedNamePrefix(resource, prefix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Resource not found")
		}
		name, ok := r.Primary.Attributes["name"]
		if !ok {
			return fmt.Errorf("Name attr not found: %#v", r.Primary.Attributes)
		}
		if !strings.HasPrefix(name, prefix) {
			return fmt.Errorf("Name: %q, does not have prefix: %q", name, prefix)
		}
		return nil
	}
}

const testAccAWSPolicyPrefixNameConfig = `
resource "aws_iam_policy" "policy" {
	name_prefix = "test-policy-"
	path = "/"
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
`
const testAccAWSPolicyInvalidJsonConfig = `
resource "aws_iam_policy" "policy" {
	name_prefix = "test-policy-"
	path = "/"
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
`

func testAccAWSPolicyDescriptionCheckConfig(rName string) string {
	return fmt.Sprintf(`
  resource "aws_iam_policy" "policy" {
    name_prefix = "test-policy-%s"
    path = "/"
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
`, rName)
}
