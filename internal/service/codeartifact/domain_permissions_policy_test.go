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

func testAccDomainPermissionsPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_domain_permissions_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CodeArtifactEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeArtifactServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainPermissionsPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, "aws_codeartifact_domain.test", names.AttrARN),
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
				Config: testAccDomainPermissionsPolicyConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, "aws_codeartifact_domain.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomain, rName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile("codeartifact:CreateRepository")),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile("codeartifact:ListRepositoriesInDomain")),
					resource.TestCheckResourceAttrPair(resourceName, "domain_owner", "aws_codeartifact_domain.test", names.AttrOwner),
				),
			},
		},
	})
}

func testAccDomainPermissionsPolicy_ignoreEquivalent(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_domain_permissions_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CodeArtifactEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeArtifactServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainPermissionsPolicyConfig_order(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, "aws_codeartifact_domain.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomain, rName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile("codeartifact:CreateRepository")),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile("codeartifact:ListRepositoriesInDomain")),
					resource.TestCheckResourceAttrPair(resourceName, "domain_owner", "aws_codeartifact_domain.test", names.AttrOwner),
				),
			},
			{
				Config:   testAccDomainPermissionsPolicyConfig_newOrder(rName),
				PlanOnly: true,
			},
		},
	})
}

func testAccDomainPermissionsPolicy_owner(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_domain_permissions_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CodeArtifactEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeArtifactServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainPermissionsPolicyConfig_owner(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainPermissionsExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, "aws_codeartifact_domain.test", names.AttrARN),
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

func testAccDomainPermissionsPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_domain_permissions_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CodeArtifactEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeArtifactServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainPermissionsPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainPermissionsExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcodeartifact.ResourceDomainPermissionsPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDomainPermissionsPolicy_Disappears_domain(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_domain_permissions_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CodeArtifactEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeArtifactServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainPermissionsPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainPermissionsExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcodeartifact.ResourceDomain(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDomainPermissionsExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeArtifactClient(ctx)

		_, err := tfcodeartifact.FindDomainPermissionsPolicyByTwoPartKey(ctx, conn, rs.Primary.Attributes["domain_owner"], rs.Primary.Attributes[names.AttrDomain])

		return err
	}
}

func testAccCheckDomainPermissionsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codeartifact_domain_permissions_policy" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).CodeArtifactClient(ctx)

			_, err := tfcodeartifact.FindDomainPermissionsPolicyByTwoPartKey(ctx, conn, rs.Primary.Attributes["domain_owner"], rs.Primary.Attributes[names.AttrDomain])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CodeArtifact Domain Permissions Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDomainPermissionsPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codeartifact_domain" "test" {
  domain         = %[1]q
  encryption_key = aws_kms_key.test.arn
}

resource "aws_codeartifact_domain_permissions_policy" "test" {
  domain          = aws_codeartifact_domain.test.domain
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

func testAccDomainPermissionsPolicyConfig_owner(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codeartifact_domain" "test" {
  domain         = %[1]q
  encryption_key = aws_kms_key.test.arn
}

resource "aws_codeartifact_domain_permissions_policy" "test" {
  domain          = aws_codeartifact_domain.test.domain
  domain_owner    = aws_codeartifact_domain.test.owner
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

func testAccDomainPermissionsPolicyConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codeartifact_domain" "test" {
  domain         = %[1]q
  encryption_key = aws_kms_key.test.arn
}

resource "aws_codeartifact_domain_permissions_policy" "test" {
  domain          = aws_codeartifact_domain.test.domain
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

func testAccDomainPermissionsPolicyConfig_order(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codeartifact_domain" "test" {
  domain         = %[1]q
  encryption_key = aws_kms_key.test.arn
}

resource "aws_codeartifact_domain_permissions_policy" "test" {
  domain = aws_codeartifact_domain.test.domain
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

func testAccDomainPermissionsPolicyConfig_newOrder(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codeartifact_domain" "test" {
  domain         = %[1]q
  encryption_key = aws_kms_key.test.arn
}

resource "aws_codeartifact_domain_permissions_policy" "test" {
  domain = aws_codeartifact_domain.test.domain
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
