// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kendra_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkendra "github.com/hashicorp/terraform-provider-aws/internal/service/kendra"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKendraQuerySuggestionsBlockList_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kendra_query_suggestions_block_list.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.KendraEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQuerySuggestionsBlockListDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQuerySuggestionsBlockListConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuerySuggestionsBlockListExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "index_id", "aws_kendra_index.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "source_s3_path.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "source_s3_path.0.bucket", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "source_s3_path.0.key", "aws_s3_object.test", names.AttrKey),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "kendra", regexache.MustCompile(`index/.+/query-suggestions-block-list/.+$`)),
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

func TestAccKendraQuerySuggestionsBlockList_Description(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kendra_query_suggestions_block_list.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.KendraEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQuerySuggestionsBlockListDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQuerySuggestionsBlockListConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuerySuggestionsBlockListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				Config: testAccQuerySuggestionsBlockListConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuerySuggestionsBlockListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccQuerySuggestionsBlockListConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuerySuggestionsBlockListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
				),
			},
		},
	})
}

func TestAccKendraQuerySuggestionsBlockList_Name(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kendra_query_suggestions_block_list.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.KendraEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQuerySuggestionsBlockListDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQuerySuggestionsBlockListConfig_basic(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuerySuggestionsBlockListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
				),
			},
			{
				Config: testAccQuerySuggestionsBlockListConfig_name(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuerySuggestionsBlockListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
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

func TestAccKendraQuerySuggestionsBlockList_RoleARN(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kendra_query_suggestions_block_list.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.KendraEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQuerySuggestionsBlockListDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQuerySuggestionsBlockListConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuerySuggestionsBlockListExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
				),
			},
			{
				Config: testAccQuerySuggestionsBlockListConfig_roleARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuerySuggestionsBlockListExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test2", names.AttrARN),
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

func TestAccKendraQuerySuggestionsBlockList_SourceS3Path(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kendra_query_suggestions_block_list.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.KendraEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQuerySuggestionsBlockListDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQuerySuggestionsBlockListConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuerySuggestionsBlockListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "source_s3_path.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "source_s3_path.0.bucket", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "source_s3_path.0.key", "aws_s3_object.test", names.AttrKey)),
			},
			{
				Config: testAccQuerySuggestionsBlockListConfig_sourceS3Path(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuerySuggestionsBlockListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "source_s3_path.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "source_s3_path.0.bucket", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "source_s3_path.0.key", "aws_s3_object.test2", names.AttrKey)),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccKendraQuerySuggestionsBlockList_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kendra_query_suggestions_block_list.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.KendraEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQuerySuggestionsBlockListDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQuerySuggestionsBlockListConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuerySuggestionsBlockListExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfkendra.ResourceQuerySuggestionsBlockList(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKendraQuerySuggestionsBlockList_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kendra_query_suggestions_block_list.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.KendraEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQuerySuggestionsBlockListDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQuerySuggestionsBlockListConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuerySuggestionsBlockListExists(ctx, resourceName),
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
				Config: testAccQuerySuggestionsBlockListConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuerySuggestionsBlockListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccQuerySuggestionsBlockListConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuerySuggestionsBlockListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckQuerySuggestionsBlockListDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KendraClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kendra_query_suggestions_block_list" {
				continue
			}

			id, indexId, err := tfkendra.QuerySuggestionsBlockListParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfkendra.FindQuerySuggestionsBlockListByID(ctx, conn, id, indexId)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Expected Kendra QuerySuggestionsBlockList to be destroyed, %s found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckQuerySuggestionsBlockListExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Kendra QuerySuggestionsBlockList is set")
		}

		id, indexId, err := tfkendra.QuerySuggestionsBlockListParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KendraClient(ctx)

		_, err = tfkendra.FindQuerySuggestionsBlockListByID(ctx, conn, id, indexId)

		if err != nil {
			return fmt.Errorf("Error describing Kendra QuerySuggestionsBlockList: %s", err.Error())
		}

		return nil
	}
}

func testAccQuerySuggestionsBlockListBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.bucket
  content = "test"
  key     = "test/suggestions.txt"
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"
    principals {
      type        = "Service"
      identifiers = ["kendra.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "test" {
  statement {
    effect = "Allow"
    actions = [
      "kendra:*",
      "s3:GetBucketLocation",
      "s3:GetObject",
      "s3:ListBucket"
    ]
    resources = [
      aws_s3_bucket.test.arn,
      "${aws_s3_bucket.test.arn}/*"
    ]
  }
}

resource "aws_iam_policy" "test" {
  name        = %[1]q
  description = "Allow Kendra to access S3"
  policy      = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test.arn
}

resource "aws_kendra_index" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
}
`, rName)
}

func testAccQuerySuggestionsBlockListConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccQuerySuggestionsBlockListBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kendra_query_suggestions_block_list" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  source_s3_path {
    bucket = aws_s3_bucket.test.id
    key    = aws_s3_object.test.key
  }
}
`, rName))
}

func testAccQuerySuggestionsBlockListConfig_description(rName, description string) string {
	return acctest.ConfigCompose(
		testAccQuerySuggestionsBlockListBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kendra_query_suggestions_block_list" "test" {
  description = %[1]q
  index_id    = aws_kendra_index.test.id
  name        = %[2]q
  role_arn    = aws_iam_role.test.arn

  source_s3_path {
    bucket = aws_s3_bucket.test.id
    key    = aws_s3_object.test.key
  }
}
`, description, rName))
}

func testAccQuerySuggestionsBlockListConfig_name(rName, name string) string {
	return acctest.ConfigCompose(
		testAccQuerySuggestionsBlockListBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kendra_query_suggestions_block_list" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  source_s3_path {
    bucket = aws_s3_bucket.test.id
    key    = aws_s3_object.test.key
  }
}
`, name))
}

func testAccQuerySuggestionsBlockListConfig_roleARN(rName string) string {
	return acctest.ConfigCompose(
		testAccQuerySuggestionsBlockListBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_iam_role" "test2" {
  name               = "%[1]s-2"
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_policy" "test2" {
  name        = "%[1]s-2"
  description = "Allow Kendra to access S3"
  policy      = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role_policy_attachment" "test2" {
  role       = aws_iam_role.test2.name
  policy_arn = aws_iam_policy.test2.arn
}

resource "aws_kendra_query_suggestions_block_list" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  role_arn = aws_iam_role.test2.arn

  source_s3_path {
    bucket = aws_s3_bucket.test.id
    key    = aws_s3_object.test.key
  }
}
`, rName))
}

func testAccQuerySuggestionsBlockListConfig_sourceS3Path(rName string) string {
	return acctest.ConfigCompose(
		testAccQuerySuggestionsBlockListBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_s3_object" "test2" {
  bucket  = aws_s3_bucket.test.bucket
  content = "test2"
  key     = "test/new_suggestions.txt"
}

resource "aws_kendra_query_suggestions_block_list" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  source_s3_path {
    bucket = aws_s3_bucket.test.id
    key    = aws_s3_object.test2.key
  }
}
`, rName))
}

func testAccQuerySuggestionsBlockListConfig_tags1(rName, tag, value string) string {
	return acctest.ConfigCompose(
		testAccQuerySuggestionsBlockListBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kendra_query_suggestions_block_list" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  source_s3_path {
    bucket = aws_s3_bucket.test.id
    key    = aws_s3_object.test.key
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag, value))
}

func testAccQuerySuggestionsBlockListConfig_tags2(rName, tag1, value1, tag2, value2 string) string {
	return acctest.ConfigCompose(
		testAccQuerySuggestionsBlockListBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kendra_query_suggestions_block_list" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  source_s3_path {
    bucket = aws_s3_bucket.test.id
    key    = aws_s3_object.test.key
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1, value1, tag2, value2))
}
