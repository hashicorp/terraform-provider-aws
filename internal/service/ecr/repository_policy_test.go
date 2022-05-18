package ecr_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfecr "github.com/hashicorp/terraform-provider-aws/internal/service/ecr"
)

func TestAccECRRepositoryPolicy_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRepositoryPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPolicyExists(resourceName),
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
				Config: testAccRepositoryPolicyUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPolicyExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "repository", "aws_ecr_repository.test", "name"),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(rName)),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile("ecr:DescribeImages")),
					acctest.CheckResourceAttrAccountID(resourceName, "registry_id"),
				),
			},
		},
	})
}

func TestAccECRRepositoryPolicy_IAM_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRepositoryPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryPolicyIAMRoleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPolicyExists(resourceName),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/19365
func TestAccECRRepositoryPolicy_IAM_principalOrder(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRepositoryPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryPolicyIAMRoleOrderJSONEncodeConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPolicyExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(rName)),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile("iam")),
				),
			},
			{
				Config: testAccRepositoryPolicyIAMRoleNewOrderJSONEncodeConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPolicyExists(resourceName),
				),
			},
			{
				Config:   testAccRepositoryPolicyIAMRoleOrderJSONEncodeConfig(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccECRRepositoryPolicy_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRepositoryPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPolicyExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfecr.ResourceRepositoryPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccECRRepositoryPolicy_Disappears_repository(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRepositoryPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPolicyExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfecr.ResourceRepository(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRepositoryPolicyDestroy(s *terraform.State) error {
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
			if tfawserr.ErrCodeEquals(err, ecr.ErrCodeRepositoryNotFoundException) ||
				tfawserr.ErrCodeEquals(err, ecr.ErrCodeRepositoryPolicyNotFoundException) {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccCheckRepositoryPolicyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccRepositoryPolicyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %[1]q
}

resource "aws_ecr_repository_policy" "test" {
  repository = aws_ecr_repository.test.name

  policy = jsonencode({
    Version = "2008-10-17"
    Statement = [{
      Sid       = %[1]q
      Effect    = "Allow"
      Principal = "*"
      Action    = "ecr:ListImages"
    }]
  })
}
`, rName)
}

func testAccRepositoryPolicyUpdatedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %[1]q
}

resource "aws_ecr_repository_policy" "test" {
  repository = aws_ecr_repository.test.name

  policy = jsonencode({
    Version = "2008-10-17"
    Statement = [{
      Sid       = %[1]q
      Effect    = "Allow"
      Principal = "*"
      Action = [
        "ecr:ListImages",
        "ecr:DescribeImages",
      ]
    }]
  })
}
`, rName)
}

// testAccRepositoryPolicyIAMRoleConfig creates a new IAM Role and tries
// to use it's ARN in an ECR Repository Policy. IAM changes need some time to
// be propagated to other services - like ECR. So the following code should
// exercise our retry logic, since we try to use the new resource instantly.
func testAccRepositoryPolicyIAMRoleConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.amazonaws.com"
      }
    }]
  })
}

resource "aws_ecr_repository_policy" "test" {
  repository = aws_ecr_repository.test.name

  policy = jsonencode({
    Version = "2008-10-17"
    Statement = [{
      Sid    = %[1]q
      Effect = "Allow",
      Principal = {
        AWS = aws_iam_role.test.arn
      }
      Action = "ecr:ListImages"
    }]
  })
}
`, rName)
}

func testAccRepositoryPolicyIAMRoleOrderBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test1" {
  name = "%[1]s-mercedes"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role" "test2" {
  name = "%[1]s-redbull"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role" "test3" {
  name = "%[1]s-mclaren"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role" "test4" {
  name = "%[1]s-ferrari"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role" "test5" {
  name = "%[1]s-astonmartin"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_ecr_repository" "test" {
  name = %[1]q
}
`, rName)
}

func testAccRepositoryPolicyIAMRoleOrderJSONEncodeConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccRepositoryPolicyIAMRoleOrderBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_ecr_repository_policy" "test" {
  repository = aws_ecr_repository.test.name

  policy = jsonencode({
    Statement = [{
      Sid    = %[1]q
      Action = "ecr:ListImages"
      Effect = "Allow"
      Principal = {
        AWS = [
          aws_iam_role.test1.arn,
          aws_iam_role.test3.arn,
          aws_iam_role.test2.arn,
          aws_iam_role.test4.arn,
          aws_iam_role.test5.arn,
        ]
      }
    }]
    Version = "2012-10-17"
  })
}
`, rName))
}

func testAccRepositoryPolicyIAMRoleNewOrderJSONEncodeConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccRepositoryPolicyIAMRoleOrderBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_ecr_repository_policy" "test" {
  repository = aws_ecr_repository.test.name

  policy = jsonencode({
    Statement = [{
      Sid    = %[1]q
      Action = "ecr:ListImages"
      Effect = "Allow"
      Principal = {
        AWS = [
          aws_iam_role.test1.arn,
          aws_iam_role.test5.arn,
          aws_iam_role.test4.arn,
          aws_iam_role.test2.arn,
          aws_iam_role.test3.arn,
        ]
      }
    }]
    Version = "2012-10-17"
  })
}
`, rName))
}
