package kendra_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/kendra/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkendra "github.com/hashicorp/terraform-provider-aws/internal/service/kendra"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDataSource_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_data_source.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_basic(rName, rName2, rName3, rName4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "kendra", regexp.MustCompile(`index/.+/data-source/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrPair(resourceName, "index_id", "aws_kendra_index.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "language_code", "en"),
					resource.TestCheckResourceAttr(resourceName, "name", rName4),
					resource.TestCheckResourceAttr(resourceName, "status", string(types.DataSourceStatusActive)),
					resource.TestCheckResourceAttr(resourceName, "type", string(types.DataSourceTypeCustom)),
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

func testAccDataSource_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_data_source.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_basic(rName, rName2, rName3, rName4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfkendra.ResourceDataSource(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDataSource_description(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	originalDescription := "Original Description"
	updatedDescription := "Updated Description"
	resourceName := "aws_kendra_data_source.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_description(rName, rName2, rName3, rName4, originalDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "kendra", regexp.MustCompile(`index/.+/data-source/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttr(resourceName, "description", originalDescription),
					resource.TestCheckResourceAttrPair(resourceName, "index_id", "aws_kendra_index.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "language_code", "en"),
					resource.TestCheckResourceAttr(resourceName, "name", rName4),
					resource.TestCheckResourceAttr(resourceName, "status", string(types.DataSourceStatusActive)),
					resource.TestCheckResourceAttr(resourceName, "type", string(types.DataSourceTypeCustom)),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_description(rName, rName2, rName3, rName4, updatedDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "kendra", regexp.MustCompile(`index/.+/data-source/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttr(resourceName, "description", updatedDescription),
					resource.TestCheckResourceAttrPair(resourceName, "index_id", "aws_kendra_index.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "language_code", "en"),
					resource.TestCheckResourceAttr(resourceName, "name", rName4),
					resource.TestCheckResourceAttr(resourceName, "status", string(types.DataSourceStatusActive)),
					resource.TestCheckResourceAttr(resourceName, "type", string(types.DataSourceTypeCustom)),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
		},
	})
}

func testAccDataSource_languageCode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	originalLanguageCode := "en"
	updatedLanguageCode := "zh"
	resourceName := "aws_kendra_data_source.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_languageCode(rName, rName2, rName3, rName4, originalLanguageCode),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "language_code", originalLanguageCode),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_languageCode(rName, rName2, rName3, rName4, updatedLanguageCode),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "language_code", updatedLanguageCode),
				),
			},
		},
	})
}

func testAccDataSource_tags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_data_source.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_tags1(rName, rName2, rName3, rName4, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_tags2(rName, rName2, rName3, rName4, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDataSourceConfig_tags1(rName, rName2, rName3, rName4, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccDataSource_typeCustomCustomizeDiff(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceConfig_typeCustomConflictRoleArn(rName, rName2, rName3, rName4, rName5),
				ExpectError: regexp.MustCompile(`role_arn must not be set when type is CUSTOM`),
			},
			{
				Config:      testAccDataSourceConfig_typeCustomConflictConfiguration(rName, rName2, rName3, rName4, rName5),
				ExpectError: regexp.MustCompile(`configuration must not be set when type is CUSTOM`),
			},
		},
	})
}

func testAccCheckDataSourceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KendraConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kendra_data_source" {
			continue
		}

		id, indexId, err := tfkendra.DataSourceParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}
		_, err = tfkendra.FindDataSourceByID(context.TODO(), conn, id, indexId)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckDataSourceExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Kendra Data Source is set")
		}

		id, indexId, err := tfkendra.DataSourceParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KendraConn

		_, err = tfkendra.FindDataSourceByID(context.TODO(), conn, id, indexId)

		if err != nil {
			return fmt.Errorf("Error describing Kendra Data Source: %s", err.Error())
		}

		return nil
	}
}

func testAccDataSourceConfigBase(rName, rName2, rName3 string) string {
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

resource "aws_iam_policy" "test_index" {
  name        = %[1]q
  description = "Kendra Index IAM permissions"
  policy      = data.aws_iam_policy_document.test_index.json
}

resource "aws_iam_role_policy_attachment" "test_index" {
  role       = aws_iam_role.test_index.name
  policy_arn = aws_iam_policy.test_index.arn
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
`, rName, rName2, rName3)
}

func testAccDataSourceConfig_basic(rName, rName2, rName3, rName4 string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		fmt.Sprintf(`
resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "CUSTOM"
}
`, rName4))
}

func testAccDataSourceConfig_description(rName, rName2, rName3, rName4, description string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		fmt.Sprintf(`
resource "aws_kendra_data_source" "test" {
  index_id    = aws_kendra_index.test.id
  name        = %[1]q
  description = %[2]q
  type        = "CUSTOM"
}
`, rName4, description))
}

func testAccDataSourceConfig_languageCode(rName, rName2, rName3, rName4, languageCode string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		fmt.Sprintf(`
resource "aws_kendra_data_source" "test" {
  index_id      = aws_kendra_index.test.id
  name          = %[1]q
  language_code = %[2]q
  type          = "CUSTOM"
}
`, rName4, languageCode))
}

func testAccDataSourceConfig_tags1(rName, rName2, rName3, rName4, tag, value string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		fmt.Sprintf(`
resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "CUSTOM"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName4, tag, value))
}

func testAccDataSourceConfig_tags2(rName, rName2, rName3, rName4, tag1, value1, tag2, value2 string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		fmt.Sprintf(`
resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "CUSTOM"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName4, tag1, value1, tag2, value2))
}

func testAccDataSourceConfig_typeCustomConflictRoleArn(rName, rName2, rName3, rName4, rName5 string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		fmt.Sprintf(`
resource "aws_iam_role" "test_data_source" {
  name               = %[2]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "CUSTOM"
  role_arn = aws_iam_role.test_data_source.arn
}
`, rName4, rName5))
}

func testAccDataSourceConfig_typeCustomConflictConfiguration(rName, rName2, rName3, rName4, rName5 string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfigBase(rName, rName2, rName3),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_kendra_data_source" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  type     = "CUSTOM"

  configuration {
    s3_configuration {
      bucket_name = aws_s3_bucket.test.id
    }
  }
}
`, rName4, rName5))
}
