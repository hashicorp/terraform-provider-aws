// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kendra_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKendraFaqDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_kendra_faq.test"
	resourceName := "aws_kendra_faq.test"
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccFaqDataSourceConfig_nonExistent,
				ExpectError: regexache.MustCompile(`getting Kendra Faq`),
			},
			{
				Config: testAccFaqDataSourceConfig_basic(rName, rName2, rName3, rName4, rName5),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrCreatedAt, resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, "faq_id", resourceName, "faq_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "file_format", resourceName, "file_format"),
					resource.TestCheckResourceAttrPair(datasourceName, "index_id", resourceName, "index_id"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrLanguageCode, resourceName, names.AttrLanguageCode),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrRoleARN, resourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttrPair(datasourceName, "s3_path.#", resourceName, "s3_path.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "s3_path.0.bucket", resourceName, "s3_path.0.bucket"),
					resource.TestCheckResourceAttrPair(datasourceName, "s3_path.0.key", resourceName, "s3_path.0.key"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrStatus, resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(datasourceName, "updated_at", resourceName, "updated_at"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Key1", resourceName, "tags.Key1")),
			},
		},
	})
}

const testAccFaqDataSourceConfig_nonExistent = `
data "aws_kendra_faq" "test" {
  faq_id   = "tf-acc-test-does-not-exist-kendra-faq-id"
  index_id = "tf-acc-test-does-not-exist-kendra-id"
}
`

func testAccFaqDataSourceConfig_basic(rName, rName2, rName3, rName4, rName5 string) string {
	return acctest.ConfigCompose(
		testAccFaqConfigBase(rName, rName2, rName3, rName4),
		fmt.Sprintf(`
resource "aws_kendra_faq" "test" {
  index_id      = aws_kendra_index.test.id
  name          = %[1]q
  file_format   = "CSV"
  language_code = "en"
  role_arn      = aws_iam_role.test_faq.arn

  s3_path {
    bucket = aws_s3_bucket.test.id
    key    = aws_s3_object.test.key
  }

  tags = {
    "Key1" = "Value1"
  }
}

data "aws_kendra_faq" "test" {
  faq_id   = aws_kendra_faq.test.faq_id
  index_id = aws_kendra_index.test.id
}
`, rName5))
}
