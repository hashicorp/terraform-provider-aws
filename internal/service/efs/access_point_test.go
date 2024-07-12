// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfefs "github.com/hashicorp/terraform-provider-aws/internal/service/efs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEFSAccessPoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var ap awstypes.AccessPointDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_efs_access_point.test"
	fsResourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(ctx, resourceName, &ap),
					resource.TestCheckResourceAttrPair(resourceName, "file_system_arn", fsResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrFileSystemID, fsResourceName, names.AttrID),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticfilesystem", regexache.MustCompile(`access-point/fsap-.+`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, "tags.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "posix_user.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "root_directory.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.path", "/"),
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

func TestAccEFSAccessPoint_Root_directory(t *testing.T) {
	ctx := acctest.Context(t)
	var ap awstypes.AccessPointDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_efs_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointConfig_rootDirectory(rName, "/home/test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(ctx, resourceName, &ap),
					resource.TestCheckResourceAttr(resourceName, "root_directory.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.path", "/home/test"),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.creation_info.#", acctest.Ct0),
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

func TestAccEFSAccessPoint_RootDirectoryCreation_info(t *testing.T) {
	ctx := acctest.Context(t)
	var ap awstypes.AccessPointDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_efs_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointConfig_rootDirectoryCreationInfo(rName, "/home/test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(ctx, resourceName, &ap),
					resource.TestCheckResourceAttr(resourceName, "root_directory.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.path", "/home/test"),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.creation_info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.creation_info.0.owner_gid", "1001"),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.creation_info.0.owner_uid", "1001"),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.creation_info.0.permissions", "755"),
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

func TestAccEFSAccessPoint_POSIX_user(t *testing.T) {
	ctx := acctest.Context(t)
	var ap awstypes.AccessPointDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_efs_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointConfig_posixUser(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(ctx, resourceName, &ap),
					resource.TestCheckResourceAttr(resourceName, "posix_user.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "posix_user.0.gid", "1001"),
					resource.TestCheckResourceAttr(resourceName, "posix_user.0.uid", "1001"),
					resource.TestCheckResourceAttr(resourceName, "posix_user.0.secondary_gids.%", acctest.Ct0),
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

func TestAccEFSAccessPoint_POSIXUserSecondary_gids(t *testing.T) {
	ctx := acctest.Context(t)
	var ap awstypes.AccessPointDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_efs_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointConfig_posixUserSecondaryGids(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(ctx, resourceName, &ap),
					resource.TestCheckResourceAttr(resourceName, "posix_user.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "posix_user.0.gid", "1001"),
					resource.TestCheckResourceAttr(resourceName, "posix_user.0.uid", "1001"),
					resource.TestCheckResourceAttr(resourceName, "posix_user.0.secondary_gids.#", acctest.Ct1)),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEFSAccessPoint_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var ap awstypes.AccessPointDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_efs_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(ctx, resourceName, &ap),
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
				Config: testAccAccessPointConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(ctx, resourceName, &ap),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAccessPointConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(ctx, resourceName, &ap),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccEFSAccessPoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var ap awstypes.AccessPointDescription
	resourceName := "aws_efs_access_point.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(ctx, resourceName, &ap),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfefs.ResourceAccessPoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAccessPointDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EFSClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_efs_access_point" {
				continue
			}

			_, err := tfefs.FindAccessPointByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EFS Access Point %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAccessPointExists(ctx context.Context, n string, v *awstypes.AccessPointDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EFSClient(ctx)

		output, err := tfefs.FindAccessPointByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAccessPointConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id
}
`, rName)
}

func testAccAccessPointConfig_rootDirectory(rName, dir string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id
  root_directory {
    path = %[2]q
  }
}
`, rName, dir)
}

func testAccAccessPointConfig_rootDirectoryCreationInfo(rName, dir string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id
  root_directory {
    path = %[2]q
    creation_info {
      owner_gid   = 1001
      owner_uid   = 1001
      permissions = "755"
    }
  }
}
`, rName, dir)
}

func testAccAccessPointConfig_posixUser(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id
  posix_user {
    gid = 1001
    uid = 1001
  }
}
`, rName)
}

func testAccAccessPointConfig_posixUserSecondaryGids(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id
  posix_user {
    gid            = 1001
    uid            = 1001
    secondary_gids = [1002]
  }
}
`, rName)
}

func testAccAccessPointConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAccessPointConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
