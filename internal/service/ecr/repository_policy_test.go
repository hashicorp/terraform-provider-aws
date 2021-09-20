package ecr_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccAWSEcrRepositoryPolicy_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecr_repository_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ecr.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSEcrRepositoryPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrRepositoryPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRepositoryPolicyExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "repository", "aws_ecr_repository.test", "name"),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(rName)),
					acctest.CheckResourceAttrAccountID(resourceName, "registry_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEcrRepositoryPolicyConfigUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRepositoryPolicyExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "repository", "aws_ecr_repository.test", "name"),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(rName)),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile("ecr:DescribeImages")),
					acctest.CheckResourceAttrAccountID(resourceName, "registry_id"),
				),
			},
		},
	})
}

func TestAccAWSEcrRepositoryPolicy_iam(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecr_repository_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ecr.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSEcrRepositoryPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrRepositoryPolicyWithIAMRoleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRepositoryPolicyExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(rName)),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile("iam")),
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

func TestAccAWSEcrRepositoryPolicy_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecr_repository_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ecr.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSEcrRepositoryPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrRepositoryPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRepositoryPolicyExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceRepositoryPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEcrRepositoryPolicy_disappears_repository(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecr_repository_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ecr.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSEcrRepositoryPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrRepositoryPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRepositoryPolicyExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceRepository(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSEcrRepositoryPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ECRConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ecr_repository_policy" {
			continue
		}

		_, err := conn.GetRepositoryPolicy(&ecr.GetRepositoryPolicyInput{
			RegistryId:     aws.String(rs.Primary.Attributes["registry_id"]),
			RepositoryName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			if tfawserr.ErrMessageContains(err, ecr.ErrCodeRepositoryNotFoundException, "") ||
				tfawserr.ErrMessageContains(err, ecr.ErrCodeRepositoryPolicyNotFoundException, "") {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccCheckAWSEcrRepositoryPolicyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccAWSEcrRepositoryPolicyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %[1]q
}

resource "aws_ecr_repository_policy" "test" {
  repository = aws_ecr_repository.test.name

  policy = <<EOF
{
    "Version": "2008-10-17",
    "Statement": [
        {
            "Sid": "%[1]s",
            "Effect": "Allow",
            "Principal": "*",
            "Action": [
                "ecr:ListImages"
            ]
        }
    ]
}
EOF
}
`, rName)
}

func testAccAWSEcrRepositoryPolicyConfigUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %[1]q
}

resource "aws_ecr_repository_policy" "test" {
  repository = aws_ecr_repository.test.name

  policy = <<EOF
{
    "Version": "2008-10-17",
    "Statement": [
        {
            "Sid": "%[1]s",
            "Effect": "Allow",
            "Principal": "*",
            "Action": [
                "ecr:ListImages",
                "ecr:DescribeImages"
            ]
        }
    ]
}
EOF
}
`, rName)
}

// testAccAWSEcrRepositoryPolicyWithIAMRoleConfig creates a new IAM Role and tries
// to use it's ARN in an ECR Repository Policy. IAM changes need some time to
// be propagated to other services - like ECR. So the following code should
// exercise our retry logic, since we try to use the new resource instantly.
func testAccAWSEcrRepositoryPolicyWithIAMRoleConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      }
    }
  ]
}
EOF
}

resource "aws_ecr_repository_policy" "test" {
  repository = aws_ecr_repository.test.name

  policy = <<EOF
{
    "Version": "2008-10-17",
    "Statement": [
        {
            "Sid": "%[1]s",
            "Effect": "Allow",
            "Principal": {
              "AWS": "${aws_iam_role.test.arn}"
            },
            "Action": [
                "ecr:ListImages"
            ]
        }
    ]
}
EOF
}
`, rName)
}
