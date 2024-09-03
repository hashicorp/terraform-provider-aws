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

func TestAccKendraQuerySuggestionsBlockListDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	datasourceName := "data.aws_kendra_query_suggestions_block_list.test"
	resourceName := "aws_kendra_query_suggestions_block_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccQuerySuggestionsBlockListDataSourceConfig_nonExistent,
				ExpectError: regexache.MustCompile(`reading Kendra QuerySuggestionsBlockList`),
			},
			{
				Config: testAccQuerySuggestionsBlockListDataSourceConfig_basic(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrSet(datasourceName, "file_size_bytes"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, "index_id", resourceName, "index_id"),
					resource.TestCheckResourceAttrSet(datasourceName, "item_count"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, "query_suggestions_block_list_id", resourceName, "query_suggestions_block_list_id"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrRoleARN, resourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttrPair(datasourceName, "source_s3_path.#", resourceName, "source_s3_path.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "source_s3_path.0.bucket", resourceName, "source_s3_path.0.bucket"),
					resource.TestCheckResourceAttrPair(datasourceName, "source_s3_path.0.key", resourceName, "source_s3_path.0.key"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrStatus, resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(datasourceName, "updated_at"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Key1", resourceName, "tags.Key1")),
			},
		},
	})
}

const testAccQuerySuggestionsBlockListDataSourceConfig_nonExistent = `
data "aws_kendra_query_suggestions_block_list" "test" {
  index_id                        = "tf-acc-test-does-not-exist-kendra-id"
  query_suggestions_block_list_id = "tf-acc-test-does-not-exist-kendra-id"
}
`

func testAccQuerySuggestionsBlockListDataSourceConfig_basic(rName, rName2 string) string {
	return acctest.ConfigCompose(
		testAccQuerySuggestionsBlockListBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kendra_query_suggestions_block_list" "test" {
  index_id    = aws_kendra_index.test.id
  name        = %[1]q
  description = "example description query suggestions block list"
  role_arn    = aws_iam_role.test.arn

  source_s3_path {
    bucket = aws_s3_bucket.test.id
    key    = aws_s3_object.test.key
  }

  tags = {
    "Key1" = "Value1"
  }
}

data "aws_kendra_query_suggestions_block_list" "test" {
  index_id                        = aws_kendra_index.test.id
  query_suggestions_block_list_id = aws_kendra_query_suggestions_block_list.test.query_suggestions_block_list_id
}
`, rName2))
}
