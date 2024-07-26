// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kendra_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/kendra/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkendra "github.com/hashicorp/terraform-provider-aws/internal/service/kendra"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKendraFaq_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_faq.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFaqDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFaqConfig_basic(rName, rName2, rName3, rName4, rName5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFaqExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "kendra", regexache.MustCompile(`index/.+/faq/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "faq_id"),
					resource.TestCheckResourceAttrPair(resourceName, "index_id", "aws_kendra_index.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrLanguageCode, "en"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName5),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_faq", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "s3_path.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "s3_path.0.bucket", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "s3_path.0.key", "aws_s3_object.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.FaqStatusActive)),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccKendraFaq_description(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	description := "example description for kendra faq"
	resourceName := "aws_kendra_faq.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFaqDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFaqConfig_description(rName, rName2, rName3, rName4, rName5, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFaqExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccKendraFaq_fileFormat(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	fileFormat := string(types.FaqFileFormatCsv)
	resourceName := "aws_kendra_faq.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFaqDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFaqConfig_fileFormat(rName, rName2, rName3, rName4, rName5, fileFormat),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFaqExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "file_format", fileFormat),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccKendraFaq_languageCode(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	languageCode := "en"
	resourceName := "aws_kendra_faq.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFaqDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFaqConfig_languageCode(rName, rName2, rName3, rName4, rName5, languageCode),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFaqExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrLanguageCode, languageCode),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccKendraFaq_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_faq.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFaqDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFaqConfig_tags1(rName, rName2, rName3, rName4, rName5, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFaqExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFaqConfig_tags2(rName, rName2, rName3, rName4, rName5, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFaqExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccFaqConfig_tags1(rName, rName2, rName3, rName4, rName5, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFaqExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccKendraFaq_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_faq.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFaqDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFaqConfig_basic(rName, rName2, rName3, rName4, rName5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFaqExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfkendra.ResourceFaq(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFaqDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KendraClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kendra_faq" {
				continue
			}

			id, indexId, err := tfkendra.FaqParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}
			_, err = tfkendra.FindFaqByID(ctx, conn, id, indexId)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}
		}

		return nil
	}
}

func testAccCheckFaqExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Kendra Faq is set")
		}

		id, indexId, err := tfkendra.FaqParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KendraClient(ctx)

		_, err = tfkendra.FindFaqByID(ctx, conn, id, indexId)

		if err != nil {
			return fmt.Errorf("Error describing Kendra Faq: %s", err.Error())
		}

		return nil
	}
}

func testAccFaqConfigBase(rName, rName2, rName3, rName4 string) string {
	// Kendra IAM policies: https://docs.aws.amazon.com/kendra/latest/dg/iam-roles.html
	return fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_kms_key" "this" {
  key_id = "alias/aws/kendra"
}
data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["kendra.amazonaws.com"]
    }
  }
}
data "aws_iam_policy_document" "test_faq" {
  statement {
    effect = "Allow"
    actions = [
      "s3:GetObject"
    ]
    resources = [
      "${aws_s3_bucket.test.arn}/*"
    ]
  }
}
data "aws_iam_policy_document" "test_index" {
  statement {
    effect = "Allow"
    actions = [
      "cloudwatch:PutMetricData"
    ]
    resources = ["*"]
    condition {
      test     = "StringEquals"
      variable = "cloudwatch:namespace"

      values = [
        "Kendra"
      ]
    }
  }

  statement {
    effect = "Allow"
    actions = [
      "logs:DescribeLogGroups"
    ]
    resources = ["*"]
  }

  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup"
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/kendra/*"
    ]
  }

  statement {
    effect = "Allow"
    actions = [
      "logs:DescribeLogStreams",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/kendra/*:log-stream:*"
    ]
  }
}

resource "aws_iam_policy" "test_faq" {
  name        = %[1]q
  description = "Allow Kendra to access S3"
  policy      = data.aws_iam_policy_document.test_faq.json
}

resource "aws_iam_policy" "test_index" {
  name        = %[2]q
  description = "Kendra Index IAM permissions"
  policy      = data.aws_iam_policy_document.test_index.json
}

resource "aws_iam_role_policy_attachment" "test_faq" {
  role       = aws_iam_role.test_faq.name
  policy_arn = aws_iam_policy.test_faq.arn
}

resource "aws_iam_role_policy_attachment" "test_index" {
  role       = aws_iam_role.test_index.name
  policy_arn = aws_iam_policy.test_index.arn
}

resource "aws_iam_role" "test_faq" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role" "test_index" {
  name               = %[2]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_kendra_index" "test" {
  depends_on = [aws_iam_role_policy_attachment.test_index]
  name       = %[3]q
  role_arn   = aws_iam_role.test_index.arn
}

resource "aws_s3_bucket" "test" {
  bucket        = %[4]q
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  source = "test-fixtures/basic.csv"
  key    = "test/basic.csv"
}
`, rName, rName2, rName3, rName4)
}

func testAccFaqConfig_basic(rName, rName2, rName3, rName4, rName5 string) string {
	return acctest.ConfigCompose(
		testAccFaqConfigBase(rName, rName2, rName3, rName4),
		fmt.Sprintf(`
resource "aws_kendra_faq" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  role_arn = aws_iam_role.test_faq.arn

  s3_path {
    bucket = aws_s3_bucket.test.id
    key    = aws_s3_object.test.key
  }
}
`, rName5))
}

func testAccFaqConfig_description(rName, rName2, rName3, rName4, rName5, description string) string {
	return acctest.ConfigCompose(
		testAccFaqConfigBase(rName, rName2, rName3, rName4),
		fmt.Sprintf(`
resource "aws_kendra_faq" "test" {
  index_id    = aws_kendra_index.test.id
  name        = %[1]q
  description = %[2]q
  role_arn    = aws_iam_role.test_faq.arn

  s3_path {
    bucket = aws_s3_bucket.test.id
    key    = aws_s3_object.test.key
  }
}
`, rName5, description))
}

func testAccFaqConfig_fileFormat(rName, rName2, rName3, rName4, rName5, fileFormat string) string {
	return acctest.ConfigCompose(
		testAccFaqConfigBase(rName, rName2, rName3, rName4),
		fmt.Sprintf(`
resource "aws_kendra_faq" "test" {
  index_id    = aws_kendra_index.test.id
  name        = %[1]q
  file_format = %[2]q
  role_arn    = aws_iam_role.test_faq.arn

  s3_path {
    bucket = aws_s3_bucket.test.id
    key    = aws_s3_object.test.key
  }
}
`, rName5, fileFormat))
}

func testAccFaqConfig_languageCode(rName, rName2, rName3, rName4, rName5, languageCode string) string {
	return acctest.ConfigCompose(
		testAccFaqConfigBase(rName, rName2, rName3, rName4),
		fmt.Sprintf(`
resource "aws_kendra_faq" "test" {
  index_id      = aws_kendra_index.test.id
  name          = %[1]q
  language_code = %[2]q
  role_arn      = aws_iam_role.test_faq.arn

  s3_path {
    bucket = aws_s3_bucket.test.id
    key    = aws_s3_object.test.key
  }
}
`, rName5, languageCode))
}

func testAccFaqConfig_tags1(rName, rName2, rName3, rName4, rName5, tag, value string) string {
	return acctest.ConfigCompose(
		testAccFaqConfigBase(rName, rName2, rName3, rName4),
		fmt.Sprintf(`
resource "aws_kendra_faq" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  role_arn = aws_iam_role.test_faq.arn

  s3_path {
    bucket = aws_s3_bucket.test.id
    key    = aws_s3_object.test.key
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName5, tag, value))
}

func testAccFaqConfig_tags2(rName, rName2, rName3, rName4, rName5, tag1, value1, tag2, value2 string) string {
	return acctest.ConfigCompose(
		testAccFaqConfigBase(rName, rName2, rName3, rName4),
		fmt.Sprintf(`
resource "aws_kendra_faq" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  role_arn = aws_iam_role.test_faq.arn

  s3_path {
    bucket = aws_s3_bucket.test.id
    key    = aws_s3_object.test.key
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName5, tag1, value1, tag2, value2))
}
