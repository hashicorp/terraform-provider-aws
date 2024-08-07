// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQuickSightFolder_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var folder quicksight.Folder
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_quicksight_folder.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, quicksight.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFolderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFolderConfig_basic(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFolderExists(ctx, resourceName, &folder),
					resource.TestCheckResourceAttr(resourceName, "folder_id", rId),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "folder_type", quicksight.FolderTypeShared),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "quicksight", fmt.Sprintf("folder/%s", rId)),
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
	var folder quicksight.Folder
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_quicksight_folder.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, quicksight.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFolderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFolderConfig_basic(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFolderExists(ctx, resourceName, &folder),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfquicksight.ResourceFolder(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccQuickSightFolder_permissions(t *testing.T) {
	ctx := acctest.Context(t)
	var folder quicksight.Folder
	resourceName := "aws_quicksight_folder.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFolderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFolderConfig_permissions(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFolderExists(ctx, resourceName, &folder),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct1),
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
					testAccCheckFolderExists(ctx, resourceName, &folder),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct1),
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
					testAccCheckFolderExists(ctx, resourceName, &folder),
					resource.TestCheckResourceAttr(resourceName, "permission.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccQuickSightFolder_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var folder quicksight.Folder
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_quicksight_folder.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, quicksight.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFolderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFolderConfig_tags1(rId, rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFolderExists(ctx, resourceName, &folder),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFolderConfig_tags2(rId, rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFolderExists(ctx, resourceName, &folder),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccFolderConfig_tags1(rId, rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFolderExists(ctx, resourceName, &folder),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccQuickSightFolder_parentFolder(t *testing.T) {
	ctx := acctest.Context(t)
	var folder quicksight.Folder
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	parentId1 := rId + "-parent1"
	parentName1 := rName + "-parent1"
	parentId2 := rId + "-parent2"
	parentName2 := rName + "-parent2"
	resourceName := "aws_quicksight_folder.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, quicksight.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFolderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFolderConfig_parentFolder(rId, rName, parentId1, parentName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFolderExists(ctx, resourceName, &folder),
					acctest.CheckResourceAttrRegionalARN(resourceName, "parent_folder_arn", "quicksight", fmt.Sprintf("folder/%s", parentId1)),
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
					testAccCheckFolderExists(ctx, resourceName, &folder),
					acctest.CheckResourceAttrRegionalARN(resourceName, "parent_folder_arn", "quicksight", fmt.Sprintf("folder/%s", parentId2)),
				),
			},
		},
	})
}

func TestAccQuickSightFolder_parentFolderNested(t *testing.T) {
	ctx := acctest.Context(t)
	var folder quicksight.Folder
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	parentId1 := rId + "-parent1"
	parentName1 := rName + "-parent1"
	parentId2 := rId + "-parent2"
	parentName2 := rName + "-parent2"
	resourceName := "aws_quicksight_folder.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, quicksight.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFolderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFolderConfig_parentFolder(rId, rName, parentId1, parentName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFolderExists(ctx, resourceName, &folder),
					acctest.CheckResourceAttrRegionalARN(resourceName, "parent_folder_arn", "quicksight", fmt.Sprintf("folder/%s", parentId1)),
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
					testAccCheckFolderExists(ctx, resourceName, &folder),
					acctest.CheckResourceAttrRegionalARN(resourceName, "parent_folder_arn", "quicksight", fmt.Sprintf("folder/%s", parentId2)),
				),
			},
		},
	})
}

func testAccCheckFolderDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_folder" {
				continue
			}

			output, err := tfquicksight.FindFolderByID(ctx, conn, rs.Primary.ID)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
					return nil
				}
				return err
			}

			if output != nil {
				return fmt.Errorf("QuickSight Folder (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckFolderExists(ctx context.Context, name string, folder *quicksight.Folder) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.QuickSight, create.ErrActionCheckingExistence, tfquicksight.ResNameFolder, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.QuickSight, create.ErrActionCheckingExistence, tfquicksight.ResNameFolder, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn(ctx)
		output, err := tfquicksight.FindFolderByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.QuickSight, create.ErrActionCheckingExistence, tfquicksight.ResNameFolder, rs.Primary.ID, err)
		}

		*folder = *output

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

func testAccFolderConfig_tags1(rId, rName, key1, value1 string) string {
	return fmt.Sprintf(`
resource "aws_quicksight_folder" "test" {
  folder_id = %[1]q
  name      = %[2]q

  tags = {
    %[3]q = %[4]q
  }
}
`, rId, rName, key1, value1)
}

func testAccFolderConfig_tags2(rId, rName, key1, value1, key2, value2 string) string {
	return fmt.Sprintf(`
resource "aws_quicksight_folder" "test" {
  folder_id = %[1]q
  name      = %[2]q

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rId, rName, key1, value1, key2, value2)
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
