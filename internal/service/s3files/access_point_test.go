// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/s3files"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3files/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfs3files "github.com/hashicorp/terraform-provider-aws/internal/service/s3files"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3FilesAccessPoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var accessPoint s3files.GetAccessPointOutput
	resourceName := "aws_s3files_access_point.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3FilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(ctx, t, resourceName, &accessPoint),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "s3files", regexache.MustCompile(`file-system/.+/access-point/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.LifeCycleStateAvailable)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, "posix_user.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "posix_user.0.gid", "1001"),
					resource.TestCheckResourceAttr(resourceName, "posix_user.0.uid", "1001"),
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

func TestAccS3FilesAccessPoint_rootDirectory(t *testing.T) {
	ctx := acctest.Context(t)
	var accessPoint s3files.GetAccessPointOutput
	resourceName := "aws_s3files_access_point.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3FilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointConfig_rootDirectory(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(ctx, t, resourceName, &accessPoint),
					resource.TestCheckResourceAttr(resourceName, "root_directory.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.path", "/restricted"),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.creation_permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.creation_permissions.0.owner_gid", "1001"),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.creation_permissions.0.owner_uid", "1001"),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.creation_permissions.0.permissions", "755"),
				),
			},
		},
	})
}

func testAccCheckAccessPointExists(ctx context.Context, t *testing.T, n string, v *s3files.GetAccessPointOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).S3FilesClient(ctx)

		output, err := tfs3files.FindAccessPointByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAccessPointDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).S3FilesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3files_access_point" {
				continue
			}

			_, err := tfs3files.FindAccessPointByID(ctx, conn, rs.Primary.ID)

			if err == nil {
				return fmt.Errorf("S3 Files Access Point %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccAccessPointConfig_base(rName string) string {
	return acctest.ConfigCompose(
		testAccFileSystemConfig_base(rName),
		`
resource "aws_s3files_file_system" "test" {
  bucket   = aws_s3_bucket.test.arn
  role_arn = aws_iam_role.test.arn

  depends_on = [aws_s3_bucket_versioning.test]
}
`)
}

func testAccAccessPointConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccAccessPointConfig_base(rName),
		fmt.Sprintf(`
resource "aws_s3files_access_point" "test" {
  file_system_id = aws_s3files_file_system.test.id

  posix_user {
    gid = 1001
    uid = 1001
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccAccessPointConfig_rootDirectory(rName string) string {
	return acctest.ConfigCompose(
		testAccAccessPointConfig_base(rName),
		fmt.Sprintf(`
resource "aws_s3files_access_point" "test" {
  file_system_id = aws_s3files_file_system.test.id

  posix_user {
    gid = 1001
    uid = 1001
  }

  root_directory {
    path = "/restricted"

    creation_permissions {
      owner_gid   = 1001
      owner_uid   = 1001
      permissions = "755"
    }
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}
