// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediapackagevod_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/mediapackagevod"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfmediapackagevod "github.com/hashicorp/terraform-provider-aws/internal/service/mediapackagevod"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccMediaPackageVODPackagingGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var packagingGroup mediapackagevod.DescribePackagingGroupOutput

	packagingGroupRName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_mediapackagevod_packaging_group.packagingGroup"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaPackageVODServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPackagingGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPackagingGroupConfig_basic(packagingGroupRName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPackagingGroupExists(ctx, resourceName, &packagingGroup),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "mediapackage-vod", regexache.MustCompile(`packaging-groups/[a-zA-Z0-9_-]+$`)),
					resource.TestMatchResourceAttr(resourceName, "name", regexache.MustCompile(packagingGroupRName)),
					resource.TestMatchResourceAttr(resourceName, "domain_name", regexache.MustCompile(fmt.Sprintf("^(https://[0-9a-z]+.egress.mediapackage-vod.%s.amazonaws.com(/.*)?)$", acctest.Region()))),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateIdFunc:                    testAccPackagingGroupImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
		},
	})
}

func testAccMediaPackageVODPackagingGroup_logging(t *testing.T) {
	ctx := acctest.Context(t)

	var packagingGroup mediapackagevod.DescribePackagingGroupOutput

	packagingGroupRName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_mediapackagevod_packaging_group.packagingGroup"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaPackageVODServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPackagingGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPackagingGroupConfig_logging(packagingGroupRName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPackagingGroupExists(ctx, resourceName, &packagingGroup),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "mediapackage-vod", regexache.MustCompile(`packaging-groups/[a-zA-Z0-9_-]+$`)),
					resource.TestMatchResourceAttr(resourceName, "name", regexache.MustCompile(packagingGroupRName)),
					resource.TestMatchResourceAttr(resourceName, "domain_name", regexache.MustCompile(fmt.Sprintf("^(https://[0-9a-z]+.egress.mediapackage-vod.%s.amazonaws.com(/.*)?)$", acctest.Region()))),
					resource.TestMatchResourceAttr(resourceName, "egress_access_logs.log_group_name", regexache.MustCompile(fmt.Sprintf(`^/aws/MediaPackage/tf-acc-test-\d+$`))),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateIdFunc:                    testAccPackagingGroupImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
		},
	})
}

func testAccMediaPackageVODPackagingGroup_authorization(t *testing.T) {
	ctx := acctest.Context(t)

	var packagingGroup mediapackagevod.DescribePackagingGroupOutput

	packagingGroupRName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_mediapackagevod_packaging_group.packagingGroup"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaPackageVODServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPackagingGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPackagingGroupConfig_authorization(packagingGroupRName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPackagingGroupExists(ctx, resourceName, &packagingGroup),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "mediapackage-vod", regexache.MustCompile(`packaging-groups/[a-zA-Z0-9_-]+$`)),
					resource.TestMatchResourceAttr(resourceName, "name", regexache.MustCompile(packagingGroupRName)),
					resource.TestMatchResourceAttr(resourceName, "domain_name", regexache.MustCompile(fmt.Sprintf("^(https://[0-9a-z]+.egress.mediapackage-vod.%s.amazonaws.com(/.*)?)$", acctest.Region()))),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateIdFunc:                    testAccPackagingGroupImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
		},
	})
}

func testAccMediaPackageVODPackagingGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var packagingGroup mediapackagevod.DescribePackagingGroupOutput

	packagingGroupRName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_mediapackagevod_packaging_group.packagingGroup"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaPackageV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPackagingGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPackagingGroupConfig_basic(packagingGroupRName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPackagingGroupExists(ctx, resourceName, &packagingGroup),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfmediapackagevod.ResourcePackagingGroup, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccPackagingGroupImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return acctest.AttrImportStateIdFunc(resourceName, names.AttrName)
}

func testAccCheckPackagingGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaPackageVODClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_mediapackagevod_packaging_group" {
				continue
			}

			_, err := tfmediapackagevod.FindPackagingGroupByID(ctx, conn, rs.Primary.Attributes[names.AttrName])
			if err == nil {
				return fmt.Errorf("MediaPackageVOD Packaging Group: %s not deleted", rs.Primary.ID)
			}

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return create.Error(names.MediaPackageVOD, create.ErrActionCheckingDestroyed, tfmediapackagevod.ResNamePackagingGroup, rs.Primary.Attributes[names.AttrName], err)
			}
		}

		return nil
	}
}

func testAccCheckPackagingGroupExists(ctx context.Context, packagingGroupName string, packagingGroup *mediapackagevod.DescribePackagingGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[packagingGroupName]
		if !ok {
			return fmt.Errorf("Not found: %s", packagingGroupName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaPackageVODClient(ctx)
		resp, err := tfmediapackagevod.FindPackagingGroupByID(ctx, conn, rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return create.Error(names.MediaPackageVOD, create.ErrActionCheckingExistence, tfmediapackagevod.ResNamePackagingGroup, rs.Primary.Attributes[names.AttrName], err)
		}

		*packagingGroup = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MediaPackageVODClient(ctx)

	input := &mediapackagevod.ListPackagingGroupsInput{}

	_, err := conn.ListPackagingGroups(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPackagingGroupConfig_basic(packagingGroupName string) string {
	return fmt.Sprintf(`
resource "aws_mediapackagevod_packaging_group" "packagingGroup" {
  name = %[1]q
}

`, packagingGroupName)
}

func testAccPackagingGroupConfig_logging(packagingGroupName string) string {
	return fmt.Sprintf(`
resource "aws_mediapackagevod_packaging_group" "packagingGroup" {
  name = %[1]q
  egress_access_logs = {
    log_group_name = aws_cloudwatch_log_group.packagingGroupLogGroup.name
  }
}

resource "aws_cloudwatch_log_group" "packagingGroupLogGroup" {
  name = "/aws/MediaPackage/%[1]s"
}

`, packagingGroupName)
}

func testAccPackagingGroupConfig_authorization(packagingGroupName string) string {
	return fmt.Sprintf(`
resource "aws_mediapackagevod_packaging_group" "packagingGroup" {
  name = %[1]q
  authorization = {
    cdn_identifier_secret = aws_secretsmanager_secret.packagingGroupSecret.arn
    secrets_role_arn      = aws_iam_role.packagingGroupRole.arn
  }

  depends_on = [
	  aws_secretsmanager_secret_version.packagingGroupSecretVersion,
	  aws_iam_role_policy.packagingGroupRolePolicy
	]
}

resource "aws_secretsmanager_secret" "packagingGroupSecret" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "packagingGroupSecretVersion" {
  secret_id = aws_secretsmanager_secret.packagingGroupSecret.id
  secret_string = jsonencode({
    MediaPackageCDNIdentifier = "testCDNIdentifier"
  })
}

resource "aws_iam_role" "packagingGroupRole" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "mediapackage.amazonaws.com"
        }
      },
    ]
  })

}

resource "aws_iam_role_policy" "packagingGroupRolePolicy" {
  name = %[1]q
  role = aws_iam_role.packagingGroupRole.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "secretsmanager:*",
        ]
        Effect   = "Allow"
        Resource = "*"
      },
    ]
  })
}

`, packagingGroupName)
}
