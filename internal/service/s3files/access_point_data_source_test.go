// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3FilesAccessPointDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_s3files_access_point.test"
	resourceName := "aws_s3files_access_point.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3FilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrFileSystemID, resourceName, names.AttrFileSystemID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrOwnerID, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStatus, resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(dataSourceName, "posix_user.#", resourceName, "posix_user.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "posix_user.0.gid", resourceName, "posix_user.0.gid"),
					resource.TestCheckResourceAttrPair(dataSourceName, "posix_user.0.uid", resourceName, "posix_user.0.uid"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrTags, resourceName, names.AttrTags),
				),
			},
		},
	})
}

func testAccAccessPointDataSourceConfig_basic(rName string) string {
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

data "aws_s3files_access_point" "test" {
  id = aws_s3files_access_point.test.id
}
`, rName))
}
