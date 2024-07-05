// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecrpublic_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfecrpublic "github.com/hashicorp/terraform-provider-aws/internal/service/ecrpublic"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECRPublicRepositoryPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, names.USEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRPublicServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
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
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
				),
			},
		},
	})
}

func TestAccECRPublicRepositoryPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, names.USEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRPublicServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfecrpublic.ResourceRepositoryPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccECRPublicRepositoryPolicy_Disappears_repository(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository_policy.test"
	repositoryResourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, names.USEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRPublicServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfecrpublic.ResourceRepository(), repositoryResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccECRPublicRepositoryPolicy_iam(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, names.USEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRPublicServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryPolicyConfig_iamRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryPolicyConfig_iamRoleUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
				),
			},
		},
	})
}

func testAccCheckRepositoryPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRPublicClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecrpublic_repository_policy" {
				continue
			}

			_, err := tfecrpublic.FindRepositoryPolicyByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ECR Public Repository Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRepositoryPolicyExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ECR Public Repository Policy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRPublicClient(ctx)

		_, err := tfecrpublic.FindRepositoryPolicyByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccRepositoryPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %[1]q
}

resource "aws_ecrpublic_repository_policy" "test" {
  repository_name = aws_ecrpublic_repository.test.repository_name

  policy = <<EOF
{
    "Version": "2008-10-17",
    "Statement": [
        {
            "Sid": "testpolicy",
            "Effect": "Allow",
            "Principal": "*",
            "Action": "ecr-public:ListImages"
        }
    ]
}
EOF
}
`, rName)
}

func testAccRepositoryPolicyConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %[1]q
}

resource "aws_ecrpublic_repository_policy" "test" {
  repository_name = aws_ecrpublic_repository.test.repository_name

  policy = <<EOF
{
    "Version": "2008-10-17",
    "Statement": [
        {
            "Sid": "testpolicy",
            "Effect": "Allow",
            "Principal": "*",
            "Action": [
                "ecr-public:ListImages",
                "ecr-public:PutImage"
            ]
        }
    ]
}
EOF
}
`, rName)
}

// testAccRepositoryPolicyConfig_iamRole creates a new IAM Role and tries
// to use it's ARN in an ECR Repository Policy. IAM changes need some time to
// be propagated to other services - like ECR. So the following code should
// exercise our retry logic, since we try to use the new resource instantly.
func testAccRepositoryPolicyConfig_iamRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %[1]q
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

resource "aws_ecrpublic_repository_policy" "test" {
  repository_name = aws_ecrpublic_repository.test.repository_name

  policy = <<EOF
{
    "Version": "2008-10-17",
    "Statement": [
        {
            "Sid": "testpolicy",
            "Effect": "Allow",
            "Principal": {
              "AWS": "${aws_iam_role.test.arn}"
            },
            "Action": "ecr-public:ListImages"
        }
    ]
}
EOF
}
`, rName)
}

func testAccRepositoryPolicyConfig_iamRoleUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %[1]q
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

resource "aws_ecrpublic_repository_policy" "test" {
  repository_name = aws_ecrpublic_repository.test.repository_name

  policy = <<EOF
{
    "Version": "2008-10-17",
    "Statement": [
        {
            "Sid": "testpolicy",
            "Effect": "Allow",
            "Principal": {
              "AWS": "${aws_iam_role.test.arn}"
            },
            "Action": [
                "ecr-public:ListImages",
                "ecr-public:PutImage"
            ]
        }
    ]
}
EOF
}
`, rName)
}
