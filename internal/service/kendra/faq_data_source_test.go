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

func TestAccKendraFaqDataSource_basic(t *testing.T) {
	datasourceName := "data.aws_kendra_faq.test"
	resourceName := "aws_kendra_faq.test"
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccFaqDataSourceConfig_nonExistent,
				ExpectError: regexp.MustCompile(`getting Kendra Faq`),
			},
			{
				Config: testAccFaqDataSourceConfig_basic(rName, rName2, rName3, rName4, rName5),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "created_at", resourceName, "created_at"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "faq_id", resourceName, "faq_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "file_format", resourceName, "file_format"),
					resource.TestCheckResourceAttrPair(datasourceName, "index_id", resourceName, "index_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "language_code", resourceName, "language_code"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "role_arn", resourceName, "role_arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "s3_path.#", resourceName, "s3_path.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "s3_path.0.bucket", resourceName, "s3_path.0.bucket"),
					resource.TestCheckResourceAttrPair(datasourceName, "s3_path.0.key", resourceName, "s3_path.0.key"),
					resource.TestCheckResourceAttrPair(datasourceName, "status", resourceName, "status"),
					resource.TestCheckResourceAttrPair(datasourceName, "updated_at", resourceName, "updated_at"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
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
