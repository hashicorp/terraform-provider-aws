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

func TestAccKendraIndexDataSource_basic(t *testing.T) {
	datasourceName := "data.aws_kendra_index.test"
	resourceName := "aws_kendra_index.test"
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccIndexDataSourceConfig_nonExistent,
				ExpectError: regexp.MustCompile(`error getting Kendra Index`),
			},
			{
				Config: testAccIndexDataSourceConfig_userTokenJSON(rName, rName2, rName3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "capacity_units.#", resourceName, "capacity_units.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "capacity_units.0.query_capacity_units", resourceName, "capacity_units.0.query_capacity_units"),
					resource.TestCheckResourceAttrPair(datasourceName, "capacity_units.0.storage_capacity_units", resourceName, "capacity_units.0.storage_capacity_units"),
					resource.TestCheckResourceAttrPair(datasourceName, "created_at", resourceName, "created_at"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "document_metadata_configuration_updates.#", resourceName, "document_metadata_configuration_updates.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "edition", resourceName, "edition"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "index_statistics.#", resourceName, "index_statistics.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "index_statistics.0.faq_statistics.#", resourceName, "index_statistics.0.faq_statistics.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "index_statistics.0.faq_statistics.0.indexed_question_answers_count", resourceName, "index_statistics.0.faq_statistics.0.indexed_question_answers_count"),
					resource.TestCheckResourceAttrPair(datasourceName, "index_statistics.0.text_document_statistics.#", resourceName, "index_statistics.0.text_document_statistics.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "index_statistics.0.text_document_statistics.0.indexed_text_bytes", resourceName, "index_statistics.0.text_document_statistics.0.indexed_text_bytes"),
					resource.TestCheckResourceAttrPair(datasourceName, "index_statistics.0.text_document_statistics.0.indexed_text_documents_count", resourceName, "index_statistics.0.text_document_statistics.0.indexed_text_documents_count"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "role_arn", resourceName, "role_arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "server_side_encryption_configuration.#", resourceName, "server_side_encryption_configuration.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "server_side_encryption_configuration.0.kms_key_id", resourceName, "server_side_encryption_configuration.0.kms_key_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "status", resourceName, "status"),
					resource.TestCheckResourceAttrPair(datasourceName, "updated_at", resourceName, "updated_at"),
					resource.TestCheckResourceAttrPair(datasourceName, "user_context_policy", resourceName, "user_context_policy"),
					resource.TestCheckResourceAttrPair(datasourceName, "user_group_resolution_configuration.#", resourceName, "user_group_resolution_configuration.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "user_token_configurations.#", resourceName, "user_token_configurations.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "user_token_configurations.0.json_token_type_configuration.#", resourceName, "user_token_configurations.0.json_token_type_configuration.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "user_token_configurations.0.json_token_type_configuration.0.group_attribute_field", resourceName, "user_token_configurations.0.json_token_type_configuration.0.group_attribute_field"),
					resource.TestCheckResourceAttrPair(datasourceName, "user_token_configurations.0.json_token_type_configuration.0.user_name_attribute_field", resourceName, "user_token_configurations.0.json_token_type_configuration.0.user_name_attribute_field"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Key1", resourceName, "tags.Key1"),
				),
			},
		},
	})
}

const testAccIndexDataSourceConfig_nonExistent = `
data "aws_kendra_index" "test" {
  id = "tf-acc-test-does-not-exist-kendra-id"
}
`

func testAccIndexDataSourceConfig_userTokenJSON(rName, rName2, rName3 string) string {
	return acctest.ConfigCompose(
		testAccIndexConfigBase(rName, rName2),
		fmt.Sprintf(`
resource "aws_kendra_index" "test" {
  name        = %[1]q
  description = "example"
  role_arn    = aws_iam_role.access_cw.arn

  server_side_encryption_configuration {
    kms_key_id = data.aws_kms_key.this.arn
  }

  user_token_configurations {
    json_token_type_configuration {
      group_attribute_field     = "groups"
      user_name_attribute_field = "username"
    }
  }

  tags = {
    "Key1" = "Value1"
  }
}

data "aws_kendra_index" "test" {
  id = aws_kendra_index.test.id
}
`, rName3))
}
