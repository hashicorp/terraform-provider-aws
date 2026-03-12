// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQuickSightFolder_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var folder awstypes.Folder
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_quicksight_folder.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.QuickSightEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFolderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFolderConfig_basic(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFolderExists(ctx, t, resourceName, &folder),
					resource.TestCheckResourceAttr(resourceName, "folder_id", rId),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "folder_type", string(awstypes.FolderTypeShared)),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "quicksight", fmt.Sprintf("folder/%s", rId)),
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

func TestAccQuickSightFolder_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var folder awstypes.Folder
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_quicksight_folder.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.QuickSightEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFolderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFolderConfig_basic(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFolderExists(ctx, t, resourceName, &folder),
					acctest.CheckSDKResourceDisappears(ctx, t, tfquicksight.ResourceFolder(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccQuickSightFolder_permissions(t *testing.T) {
	ctx := acctest.Context(t)
	var folder awstypes.Folder
	resourceName := "aws_quicksight_folder.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFolderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFolderConfig_permissions(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFolderExists(ctx, t, resourceName, &folder),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "permissions.*", map[string]*regexp.Regexp{
						names.AttrPrincipal: regexache.MustCompile(fmt.Sprintf(`user/default/%s`, rName)),
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:DescribeFolder"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFolderConfig_permissionsUpdate(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFolderExists(ctx, t, resourceName, &folder),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "permissions.*", map[string]*regexp.Regexp{
						names.AttrPrincipal: regexache.MustCompile(fmt.Sprintf(`user/default/%s`, rName)),
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:CreateFolder"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:DescribeFolder"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:UpdateFolder"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:DeleteFolder"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:CreateFolderMembership"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:DeleteFolderMembership"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:DescribeFolderPermissions"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:UpdateFolderPermissions"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFolderConfig_basic(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFolderExists(ctx, t, resourceName, &folder),
					resource.TestCheckResourceAttr(resourceName, "permission.#", "0"),
				),
			},
		},
	})
}

func TestAccQuickSightFolder_parentFolder(t *testing.T) {
	ctx := acctest.Context(t)
	var folder awstypes.Folder
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	parentId1 := rId + "-parent1"
	parentName1 := rName + "-parent1"
	parentId2 := rId + "-parent2"
	parentName2 := rName + "-parent2"
	resourceName := "aws_quicksight_folder.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.QuickSightEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFolderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFolderConfig_parentFolder(rId, rName, parentId1, parentName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFolderExists(ctx, t, resourceName, &folder),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, "parent_folder_arn", "quicksight", fmt.Sprintf("folder/%s", parentId1)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFolderConfig_parentFolder(rId, rName, parentId2, parentName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFolderExists(ctx, t, resourceName, &folder),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, "parent_folder_arn", "quicksight", fmt.Sprintf("folder/%s", parentId2)),
				),
			},
		},
	})
}

func TestAccQuickSightFolder_parentFolderNested(t *testing.T) {
	ctx := acctest.Context(t)
	var folder awstypes.Folder
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	parentId1 := rId + "-parent1"
	parentName1 := rName + "-parent1"
	parentId2 := rId + "-parent2"
	parentName2 := rName + "-parent2"
	resourceName := "aws_quicksight_folder.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.QuickSightEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFolderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFolderConfig_parentFolder(rId, rName, parentId1, parentName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFolderExists(ctx, t, resourceName, &folder),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, "parent_folder_arn", "quicksight", fmt.Sprintf("folder/%s", parentId1)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFolderConfig_parentFolder2(rId, rName, parentId1, parentName1, parentId2, parentName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFolderExists(ctx, t, resourceName, &folder),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, "parent_folder_arn", "quicksight", fmt.Sprintf("folder/%s", parentId2)),
				),
			},
		},
	})
}

func testAccCheckFolderDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_folder" {
				continue
			}

			_, err := tfquicksight.FindFolderByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes["folder_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("QuickSight Folder (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckFolderExists(ctx context.Context, t *testing.T, n string, v *awstypes.Folder) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		output, err := tfquicksight.FindFolderByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes["folder_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccFolderConfig_basic(rId, rName string) string {
	return fmt.Sprintf(`
resource "aws_quicksight_folder" "test" {
  folder_id = %[1]q
  name      = %[2]q
}
`, rId, rName)
}

func testAccFolderConfigUserBase(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_quicksight_user" "test" {
  aws_account_id = data.aws_caller_identity.current.account_id
  user_name      = %[1]q
  email          = %[2]q
  identity_type  = "QUICKSIGHT"
  user_role      = "AUTHOR"

  lifecycle {
    create_before_destroy = true
  }
}
`, rName, acctest.DefaultEmailAddress)
}

func testAccFolderConfig_permissions(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccFolderConfigUserBase(rName),
		fmt.Sprintf(`
resource "aws_quicksight_folder" "test" {
  folder_id = %[1]q
  name      = %[2]q
  permissions {
    actions = [
      "quicksight:DescribeFolder",
    ]
    principal = aws_quicksight_user.test.arn
  }
}
`, rId, rName))
}

func testAccFolderConfig_permissionsUpdate(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccFolderConfigUserBase(rName),
		fmt.Sprintf(`
resource "aws_quicksight_folder" "test" {
  folder_id = %[1]q
  name      = %[2]q
  permissions {
    actions = [
      "quicksight:CreateFolder",
      "quicksight:DescribeFolder",
      "quicksight:UpdateFolder",
      "quicksight:DeleteFolder",
      "quicksight:CreateFolderMembership",
      "quicksight:DeleteFolderMembership",
      "quicksight:DescribeFolderPermissions",
      "quicksight:UpdateFolderPermissions",
    ]
    principal = aws_quicksight_user.test.arn
  }
}
`, rId, rName))
}

func testAccFolderConfig_parentFolder(rId, rName, parentId, parentName string) string {
	return fmt.Sprintf(`
resource "aws_quicksight_folder" "parent" {
  folder_id = %[3]q
  name      = %[4]q
}

resource "aws_quicksight_folder" "test" {
  folder_id         = %[1]q
  name              = %[2]q
  parent_folder_arn = aws_quicksight_folder.parent.arn
}
`, rId, rName, parentId, parentName)
}

func testAccFolderConfig_parentFolder2(rId, rName, parentId1, parentName1, parentId2, parentName2 string) string {
	return fmt.Sprintf(`
resource "aws_quicksight_folder" "parent" {
  folder_id = %[3]q
  name      = %[4]q
}

resource "aws_quicksight_folder" "parent2" {
  folder_id         = %[5]q
  name              = %[6]q
  parent_folder_arn = aws_quicksight_folder.parent.arn
}

resource "aws_quicksight_folder" "test" {
  folder_id         = %[1]q
  name              = %[2]q
  parent_folder_arn = aws_quicksight_folder.parent2.arn
}
`, rId, rName, parentId1, parentName1, parentId2, parentName2)
}
