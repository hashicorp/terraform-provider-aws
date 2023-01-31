package kendra_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/backup"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccKendraThesaurusDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	datasourceName := "data.aws_kendra_thesaurus.test"
	resourceName := "aws_kendra_thesaurus.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, backup.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccThesaurusDataSourceConfig_nonExistent,
				ExpectError: regexp.MustCompile(`reading Kendra Thesaurus`),
			},
			{
				Config: testAccThesaurusDataSourceConfig_basic(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrSet(datasourceName, "created_at"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrSet(datasourceName, "file_size_bytes"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "index_id", resourceName, "index_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "role_arn", resourceName, "role_arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "source_s3_path.#", resourceName, "source_s3_path.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "source_s3_path.0.bucket", resourceName, "source_s3_path.0.bucket"),
					resource.TestCheckResourceAttrPair(datasourceName, "source_s3_path.0.key", resourceName, "source_s3_path.0.key"),
					resource.TestCheckResourceAttrPair(datasourceName, "status", resourceName, "status"),
					resource.TestCheckResourceAttrSet(datasourceName, "synonym_rule_count"),
					resource.TestCheckResourceAttrSet(datasourceName, "term_count"),
					resource.TestCheckResourceAttrPair(datasourceName, "thesaurus_id", resourceName, "thesaurus_id"),
					resource.TestCheckResourceAttrSet(datasourceName, "updated_at"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Key1", resourceName, "tags.Key1")),
			},
		},
	})
}

const testAccThesaurusDataSourceConfig_nonExistent = `
data "aws_kendra_thesaurus" "test" {
  index_id     = "tf-acc-test-does-not-exist-kendra-id"
  thesaurus_id = "tf-acc-test-does-not-exist-kendra-thesaurus-id"
}
`

func testAccThesaurusDataSourceConfig_basic(rName, rName2 string) string {
	return acctest.ConfigCompose(
		testAccThesaurusBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kendra_thesaurus" "test" {
  index_id    = aws_kendra_index.test.id
  name        = %[1]q
  description = "example description thesaurus"
  role_arn    = aws_iam_role.test.arn

  source_s3_path {
    bucket = aws_s3_bucket.test.id
    key    = aws_s3_object.test.key
  }

  tags = {
    "Key1" = "Value1"
  }
}

data "aws_kendra_thesaurus" "test" {
  index_id     = aws_kendra_index.test.id
  thesaurus_id = aws_kendra_thesaurus.test.thesaurus_id
}
`, rName2))
}
