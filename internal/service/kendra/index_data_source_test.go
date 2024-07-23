// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kendra_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKendraIndexDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_kendra_index.test"
	resourceName := "aws_kendra_index.test"
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexDataSourceConfig_userTokenJSON(rName, rName2, rName3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "capacity_units.#", resourceName, "capacity_units.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "capacity_units.0.query_capacity_units", resourceName, "capacity_units.0.query_capacity_units"),
					resource.TestCheckResourceAttrPair(datasourceName, "capacity_units.0.storage_capacity_units", resourceName, "capacity_units.0.storage_capacity_units"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrCreatedAt, resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, "document_metadata_configuration_updates.#", resourceName, "document_metadata_configuration_updates.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "edition", resourceName, "edition"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, "index_statistics.#", resourceName, "index_statistics.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "index_statistics.0.faq_statistics.#", resourceName, "index_statistics.0.faq_statistics.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "index_statistics.0.faq_statistics.0.indexed_question_answers_count", resourceName, "index_statistics.0.faq_statistics.0.indexed_question_answers_count"),
					resource.TestCheckResourceAttrPair(datasourceName, "index_statistics.0.text_document_statistics.#", resourceName, "index_statistics.0.text_document_statistics.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "index_statistics.0.text_document_statistics.0.indexed_text_bytes", resourceName, "index_statistics.0.text_document_statistics.0.indexed_text_bytes"),
					resource.TestCheckResourceAttrPair(datasourceName, "index_statistics.0.text_document_statistics.0.indexed_text_documents_count", resourceName, "index_statistics.0.text_document_statistics.0.indexed_text_documents_count"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrRoleARN, resourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttrPair(datasourceName, "server_side_encryption_configuration.#", resourceName, "server_side_encryption_configuration.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "server_side_encryption_configuration.0.kms_key_id", resourceName, "server_side_encryption_configuration.0.kms_key_id"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrStatus, resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(datasourceName, "updated_at", resourceName, "updated_at"),
					resource.TestCheckResourceAttrPair(datasourceName, "user_context_policy", resourceName, "user_context_policy"),
					resource.TestCheckResourceAttrPair(datasourceName, "user_group_resolution_configuration.#", resourceName, "user_group_resolution_configuration.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "user_token_configurations.#", resourceName, "user_token_configurations.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "user_token_configurations.0.json_token_type_configuration.#", resourceName, "user_token_configurations.0.json_token_type_configuration.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "user_token_configurations.0.json_token_type_configuration.0.group_attribute_field", resourceName, "user_token_configurations.0.json_token_type_configuration.0.group_attribute_field"),
					resource.TestCheckResourceAttrPair(datasourceName, "user_token_configurations.0.json_token_type_configuration.0.user_name_attribute_field", resourceName, "user_token_configurations.0.json_token_type_configuration.0.user_name_attribute_field"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Key1", resourceName, "tags.Key1"),
				),
			},
		},
	})
}

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
