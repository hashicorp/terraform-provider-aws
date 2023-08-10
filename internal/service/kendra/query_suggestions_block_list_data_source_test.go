// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kendra_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/backup"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
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
		ErrorCheck:               acctest.ErrorCheck(t, backup.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccQuerySuggestionsBlockListDataSourceConfig_nonExistent,
				ExpectError: regexp.MustCompile(`reading Kendra QuerySuggestionsBlockList`),
			},
			{
				Config: testAccQuerySuggestionsBlockListDataSourceConfig_basic(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrSet(datasourceName, "created_at"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrSet(datasourceName, "file_size_bytes"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "index_id", resourceName, "index_id"),
					resource.TestCheckResourceAttrSet(datasourceName, "item_count"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "query_suggestions_block_list_id", resourceName, "query_suggestions_block_list_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "role_arn", resourceName, "role_arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "source_s3_path.#", resourceName, "source_s3_path.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "source_s3_path.0.bucket", resourceName, "source_s3_path.0.bucket"),
					resource.TestCheckResourceAttrPair(datasourceName, "source_s3_path.0.key", resourceName, "source_s3_path.0.key"),
					resource.TestCheckResourceAttrPair(datasourceName, "status", resourceName, "status"),
					resource.TestCheckResourceAttrSet(datasourceName, "updated_at"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
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
