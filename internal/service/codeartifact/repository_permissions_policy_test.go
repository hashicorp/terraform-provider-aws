// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codeartifact_test

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
	tfcodeartifact "github.com/hashicorp/terraform-provider-aws/internal/service/codeartifact"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccRepositoryPermissionsPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_repository_permissions_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CodeArtifactEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeArtifactServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryPermissionsPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, "aws_codeartifact_repository.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomain, rName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile("codeartifact:CreateRepository")),
					resource.TestCheckResourceAttrPair(resourceName, "domain_owner", "aws_codeartifact_domain.test", names.AttrOwner),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryPermissionsPolicyConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, "aws_codeartifact_repository.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomain, rName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile("codeartifact:CreateRepository")),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile("codeartifact:ListRepositoriesInDomain")),
					resource.TestCheckResourceAttrPair(resourceName, "domain_owner", "aws_codeartifact_domain.test", names.AttrOwner),
				),
			},
		},
	})
}

func testAccRepositoryPermissionsPolicy_ignoreEquivalent(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_repository_permissions_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CodeArtifactEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeArtifactServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryPermissionsPolicyConfig_order(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, "aws_codeartifact_repository.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomain, rName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile("codeartifact:CreateRepository")),
					resource.TestCheckResourceAttrPair(resourceName, "domain_owner", "aws_codeartifact_domain.test", names.AttrOwner),
				),
			},
			{
				Config:   testAccRepositoryPermissionsPolicyConfig_newOrder(rName),
				PlanOnly: true,
			},
		},
	})
}

func testAccRepositoryPermissionsPolicy_owner(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_repository_permissions_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CodeArtifactEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeArtifactServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryPermissionsPolicyConfig_owner(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, "aws_codeartifact_repository.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomain, rName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile("codeartifact:CreateRepository")),
					resource.TestCheckResourceAttrPair(resourceName, "domain_owner", "aws_codeartifact_domain.test", names.AttrOwner),
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

func testAccRepositoryPermissionsPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_repository_permissions_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CodeArtifactEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeArtifactServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryPermissionsPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPermissionsExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcodeartifact.ResourceRepositoryPermissionsPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccRepositoryPermissionsPolicy_Disappears_domain(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_repository_permissions_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CodeArtifactEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeArtifactServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryPermissionsPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPermissionsExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcodeartifact.ResourceRepositoryPermissionsPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRepositoryPermissionsExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeArtifactClient(ctx)

		_, err := tfcodeartifact.FindRepositoryPermissionsPolicyByThreePartKey(ctx, conn, rs.Primary.Attributes["domain_owner"], rs.Primary.Attributes[names.AttrDomain], rs.Primary.Attributes["repository"])

		return err
	}
}

func testAccCheckRepositoryPermissionsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codeartifact_repository_permissions_policy" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).CodeArtifactClient(ctx)

			_, err := tfcodeartifact.FindRepositoryPermissionsPolicyByThreePartKey(ctx, conn, rs.Primary.Attributes["domain_owner"], rs.Primary.Attributes[names.AttrDomain], rs.Primary.Attributes["repository"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CodeArtifact Repository Permissions Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccRepositoryPermissionsPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codeartifact_domain" "test" {
  domain         = %[1]q
  encryption_key = aws_kms_key.test.arn
}

resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain
}

resource "aws_codeartifact_repository_permissions_policy" "test" {
  domain          = aws_codeartifact_domain.test.domain
  repository      = aws_codeartifact_repository.test.repository
  policy_document = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "codeartifact:CreateRepository",
            "Effect": "Allow",
            "Principal": "*",
            "Resource": "${aws_codeartifact_domain.test.arn}"
        }
    ]
}
EOF
}
`, rName)
}

func testAccRepositoryPermissionsPolicyConfig_owner(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codeartifact_domain" "test" {
  domain         = %[1]q
  encryption_key = aws_kms_key.test.arn
}

resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain
}

resource "aws_codeartifact_repository_permissions_policy" "test" {
  domain          = aws_codeartifact_domain.test.domain
  domain_owner    = aws_codeartifact_domain.test.owner
  repository      = aws_codeartifact_repository.test.repository
  policy_document = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "codeartifact:CreateRepository",
            "Effect": "Allow",
            "Principal": "*",
            "Resource": "${aws_codeartifact_domain.test.arn}"
        }
    ]
}
EOF
}
`, rName)
}

func testAccRepositoryPermissionsPolicyConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codeartifact_domain" "test" {
  domain         = %[1]q
  encryption_key = aws_kms_key.test.arn
}

resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain
}

resource "aws_codeartifact_repository_permissions_policy" "test" {
  domain          = aws_codeartifact_domain.test.domain
  repository      = aws_codeartifact_repository.test.repository
  policy_document = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": [
 				"codeartifact:CreateRepository",
				"codeartifact:ListRepositoriesInDomain"
			],
            "Effect": "Allow",
            "Principal": "*",
            "Resource": "${aws_codeartifact_domain.test.arn}"
        }
    ]
}
EOF
}
`, rName)
}

func testAccRepositoryPermissionsPolicyConfig_order(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codeartifact_domain" "test" {
  domain         = %[1]q
  encryption_key = aws_kms_key.test.arn
}

resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain
}

resource "aws_codeartifact_repository_permissions_policy" "test" {
  domain     = aws_codeartifact_domain.test.domain
  repository = aws_codeartifact_repository.test.repository
  policy_document = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = [
        "codeartifact:CreateRepository",
        "codeartifact:ListRepositoriesInDomain",
      ]
      Effect    = "Allow"
      Principal = "*"
      Resource  = aws_codeartifact_domain.test.arn
    }]
  })
}
`, rName)
}

func testAccRepositoryPermissionsPolicyConfig_newOrder(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codeartifact_domain" "test" {
  domain         = %[1]q
  encryption_key = aws_kms_key.test.arn
}

resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain
}

resource "aws_codeartifact_repository_permissions_policy" "test" {
  domain     = aws_codeartifact_domain.test.domain
  repository = aws_codeartifact_repository.test.repository
  policy_document = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = [
        "codeartifact:ListRepositoriesInDomain",
        "codeartifact:CreateRepository",
      ]
      Effect    = "Allow"
      Principal = "*"
      Resource  = aws_codeartifact_domain.test.arn
    }]
  })
}
`, rName)
}
