// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfecr "github.com/hashicorp/terraform-provider-aws/internal/service/ecr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECRRepositoryPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "repository", "aws_ecr_repository.test", names.AttrName),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(rName)),
					acctest.CheckResourceAttrAccountID(resourceName, "registry_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryPolicyConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "repository", "aws_ecr_repository.test", names.AttrName),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(rName)),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile("ecr:DescribeImages")),
					acctest.CheckResourceAttrAccountID(resourceName, "registry_id"),
				),
			},
		},
	})
}

func TestAccECRRepositoryPolicy_IAM_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryPolicyConfig_iamRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPolicyExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(rName)),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile("iam")),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryPolicyConfig_iamRoleOrderJSONEncode(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPolicyExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(rName)),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile("iam")),
				),
			},
			{
				Config: testAccRepositoryPolicyConfig_iamRoleNewOrderJSONEncode(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPolicyExists(ctx, resourceName),
				),
			},
			{
				Config:   testAccRepositoryPolicyConfig_iamRoleOrderJSONEncode(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccECRRepositoryPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPolicyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfecr.ResourceRepositoryPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccECRRepositoryPolicy_Disappears_repository(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPolicyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfecr.ResourceRepository(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRepositoryPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecr_repository_policy" {
				continue
			}

			_, err := tfecr.FindRepositoryPolicyByRepositoryName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ECR Repository Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRepositoryPolicyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRClient(ctx)

		_, err := tfecr.FindRepositoryPolicyByRepositoryName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccRepositoryPolicyConfig_basic(rName string) string {
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

func testAccRepositoryPolicyConfig_updated(rName string) string {
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

// testAccRepositoryPolicyConfig_iamRole creates a new IAM Role and tries
// to use it's ARN in an ECR Repository Policy. IAM changes need some time to
// be propagated to other services - like ECR. So the following code should
// exercise our retry logic, since we try to use the new resource instantly.
func testAccRepositoryPolicyConfig_iamRole(rName string) string {
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

func testAccRepositoryPolicyConfig_iamRoleOrderJSONEncode(rName string) string {
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

func testAccRepositoryPolicyConfig_iamRoleNewOrderJSONEncode(rName string) string {
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
