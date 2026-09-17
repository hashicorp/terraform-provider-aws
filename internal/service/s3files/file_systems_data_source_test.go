// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3FilesFileSystemsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_s3files_file_systems.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3FilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "file_systems.#"),
				),
			},
		},
	})
}

func testAccFileSystemsDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccFileSystemConfig_base(rName), `
resource "aws_s3files_file_system" "test" {
  bucket   = aws_s3_bucket.test.arn
  role_arn = aws_iam_role.test.arn

  depends_on = [aws_s3_bucket_versioning.test]
}

data "aws_s3files_file_systems" "test" {
  depends_on = [aws_s3files_file_system.test]
}
`)
}
