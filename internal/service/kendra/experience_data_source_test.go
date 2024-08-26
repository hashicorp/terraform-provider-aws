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

func TestAccKendraExperienceDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	datasourceName := "data.aws_kendra_experience.test"
	resourceName := "aws_kendra_experience.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccExperienceDataSourceConfig_nonExistent,
				ExpectError: regexache.MustCompile(`reading Kendra Experience`),
			},
			{
				Config: testAccExperienceDataSourceConfig_basic(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "configuration.#", resourceName, "configuration.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "configuration.0.content_source_configuration.#", resourceName, "configuration.0.content_source_configuration.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "configuration.0.content_source_configuration.0.faq_ids.#", resourceName, "configuration.0.content_source_configuration.0.faq_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "configuration.0.content_source_configuration.0.faq_ids.0", resourceName, "configuration.0.content_source_configuration.0.faq_ids.0"),
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, "endpoints.#", resourceName, "endpoints.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "endpoints.0.endpoint", resourceName, "endpoints.0.endpoint"),
					resource.TestCheckResourceAttrPair(datasourceName, "endpoints.0.endpoint_type", resourceName, "endpoints.0.endpoint_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "experience_id", resourceName, "experience_id"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, "index_id", resourceName, "index_id"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrRoleARN, resourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrStatus, resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(datasourceName, "updated_at"),
				)},
		},
	})
}

const testAccExperienceDataSourceConfig_nonExistent = `
data "aws_kendra_experience" "test" {
  experience_id = "tf-acc-test-does-not-exist-exp-id"
  index_id      = "tf-acc-test-does-not-exist-kendra-id"
}
`

func testAccExperienceDataSourceConfig_basic(rName, rName2 string) string {
	return acctest.ConfigCompose(
		testAccExperienceBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  source = "test-fixtures/basic.csv"
  key    = "test/basic.csv"
}

data "aws_iam_policy_document" "faq" {
  statement {
    sid    = "AllowKendraToAccessS3"
    effect = "Allow"
    actions = [
      "s3:GetObject"
    ]
    resources = [
      "${aws_s3_bucket.test.arn}/*"
    ]
  }
}

resource "aws_iam_policy" "faq" {
  name        = "%[1]s-faq"
  description = "Allow Kendra to access S3"
  policy      = data.aws_iam_policy_document.faq.json
}

resource "aws_iam_role_policy_attachment" "faq" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.faq.arn
}

resource "aws_kendra_faq" "test" {
  depends_on = [aws_iam_role_policy_attachment.faq]

  index_id = aws_kendra_index.test.id
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  s3_path {
    bucket = aws_s3_bucket.test.id
    key    = aws_s3_object.test.key
  }
}

resource "aws_kendra_experience" "test" {
  depends_on = [aws_iam_role_policy_attachment.experience]

  index_id = aws_kendra_index.test.id
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  configuration {
    content_source_configuration {
      faq_ids = [aws_kendra_faq.test.faq_id]
    }
  }
}

data "aws_kendra_experience" "test" {
  experience_id = aws_kendra_experience.test.experience_id
  index_id      = aws_kendra_index.test.id
}
`, rName2))
}
