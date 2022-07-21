package comprehend_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/comprehend/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcomprehend "github.com/hashicorp/terraform-provider-aws/internal/service/comprehend"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccComprehendEntityRecognizer_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var entityrecognizer types.EntityRecognizerProperties
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_comprehend_entity_recognizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.ComprehendEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntityRecognizerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEntityRecognizerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(resourceName, &entityrecognizer),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "data_access_role_arn", "aws_iam_role.test", "arn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "comprehend", regexp.MustCompile(fmt.Sprintf(`entity-recognizer/%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.entity_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.annotations.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.augmented_manifests.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.data_format", string(types.EntityRecognizerDataFormatComprehendCsv)),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.documents.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.entity_list.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "language_code", "en"),
					resource.TestCheckResourceAttr(resourceName, "model_kms_key_id", ""),
					resource.TestCheckNoResourceAttr(resourceName, "model_policy"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_name", ""),
					resource.TestCheckResourceAttr(resourceName, "volume_kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "0"),
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

func TestAccComprehendEntityRecognizer_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var entityrecognizer types.EntityRecognizerProperties
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_comprehend_entity_recognizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.ComprehendEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntityRecognizerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEntityRecognizerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(resourceName, &entityrecognizer),
					acctest.CheckResourceDisappears(acctest.Provider, tfcomprehend.ResourceEntityRecognizer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccComprehendEntityRecognizer_tags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2, v3 types.EntityRecognizerProperties
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_comprehend_entity_recognizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.ComprehendEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntityRecognizerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEntityRecognizerConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(resourceName, &v1),
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
				Config: testAccEntityRecognizerConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(resourceName, &v2),
					testAccCheckEntityRecognizerNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccEntityRecognizerConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntityRecognizerExists(resourceName, &v3),
					testAccCheckEntityRecognizerNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccComprehendEntityRecognizer_DefaultTags_providerOnly(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2, v3 types.EntityRecognizerProperties
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_comprehend_entity_recognizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.ComprehendEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComprehendEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntityRecognizerDestroy,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("providerkey1", "providervalue1"),
					testAccEntityRecognizerConfig_tags0(rName),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEntityRecognizerExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "providervalue1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags2("providerkey1", "providervalue1", "providerkey2", "providervalue2"),
					testAccEntityRecognizerConfig_tags0(rName),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEntityRecognizerExists(resourceName, &v2),
					testAccCheckEntityRecognizerNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "providervalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey2", "providervalue2"),
				),
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("providerkey1", "value1"),
					testAccEntityRecognizerConfig_tags0(rName),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEntityRecognizerExists(resourceName, &v3),
					testAccCheckEntityRecognizerNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "value1"),
				),
			},
		},
	})
}

// TODO: test deletion from in-error state. Try insufficient permissions to force error

// TODO: add test for catching, e.g. permission errors in training

func testAccCheckEntityRecognizerDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ComprehendConn
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_comprehend_entity_recognizer" {
			continue
		}

		_, err := tfcomprehend.FindEntityRecognizerByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			if tfresource.NotFound(err) {
				return nil
			}
			return err
		}

		return fmt.Errorf("Expected Comprehend Entity Recognizer to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckEntityRecognizerExists(name string, entityrecognizer *types.EntityRecognizerProperties) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Comprehend Entity Recognizer is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ComprehendConn
		ctx := context.Background()

		resp, err := tfcomprehend.FindEntityRecognizerByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error describing Comprehend Entity Recognizer: %w", err)
		}

		*entityrecognizer = *resp

		return nil
	}
}

// func testAccCheckEntityRecognizerRecreated(before, after *types.EntityRecognizerProperties) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		if entityRecognizerIdentity(before, after) {
// 			return fmt.Errorf("Comprehend Entity Recognizer not recreated")
// 		}

// 		return nil
// 	}
// }

func testAccCheckEntityRecognizerNotRecreated(before, after *types.EntityRecognizerProperties) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !entityRecognizerIdentity(before, after) {
			return fmt.Errorf("Comprehend Entity Recognizer recreated")
		}

		return nil
	}
}

func entityRecognizerIdentity(before, after *types.EntityRecognizerProperties) bool {
	return aws.ToTime(before.SubmitTime).Equal(aws.ToTime(after.SubmitTime))
}

func testAccEntityRecognizerConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccEntityRecognizerBasicRoleConfig(rName),
		testAccEntityRecognizerS3BucketConfig(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_entity_recognizer" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  language_code = "en"
  input_data_config {
    entity_types {
      type = "ENGINEER"
    }
    entity_types {
      type = "MANAGER"
    }

    documents {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
    }

    entity_list {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.entities.id}"
    }
  }

  depends_on = [
    aws_iam_role_policy.test
  ]
}

resource "aws_s3_object" "documents" {
  bucket = aws_s3_bucket.test.bucket
  key    = "documents.txt"
  source = "test-fixtures/entity_recognizer/documents.txt"
}

resource "aws_s3_object" "entities" {
  bucket = aws_s3_bucket.test.bucket
  key    = "entitylist.csv"
  source = "test-fixtures/entity_recognizer/entitylist.csv"
}
`, rName))
}

func testAccEntityRecognizerConfig_tags0(rName string) string {
	return acctest.ConfigCompose(
		testAccEntityRecognizerBasicRoleConfig(rName),
		testAccEntityRecognizerS3BucketConfig(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_entity_recognizer" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  tags = {}

  language_code = "en"
  input_data_config {
    entity_types {
      type = "ENGINEER"
    }
    entity_types {
      type = "MANAGER"
    }

    documents {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
    }

    entity_list {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.entities.id}"
    }
  }

  depends_on = [
    aws_iam_role_policy.test
  ]
}

resource "aws_s3_object" "documents" {
  bucket = aws_s3_bucket.test.bucket
  key    = "documents.txt"
  source = "test-fixtures/entity_recognizer/documents.txt"
}

resource "aws_s3_object" "entities" {
  bucket = aws_s3_bucket.test.bucket
  key    = "entitylist.csv"
  source = "test-fixtures/entity_recognizer/entitylist.csv"
}
`, rName))
}

func testAccEntityRecognizerConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccEntityRecognizerBasicRoleConfig(rName),
		testAccEntityRecognizerS3BucketConfig(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_entity_recognizer" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  tags = {
	%[2]q = %[3]q
  }

  language_code = "en"
  input_data_config {
    entity_types {
      type = "ENGINEER"
    }
    entity_types {
      type = "MANAGER"
    }

    documents {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
    }

    entity_list {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.entities.id}"
    }
  }

  depends_on = [
    aws_iam_role_policy.test
  ]
}

resource "aws_s3_object" "documents" {
  bucket = aws_s3_bucket.test.bucket
  key    = "documents.txt"
  source = "test-fixtures/entity_recognizer/documents.txt"
}

resource "aws_s3_object" "entities" {
  bucket = aws_s3_bucket.test.bucket
  key    = "entitylist.csv"
  source = "test-fixtures/entity_recognizer/entitylist.csv"
}
`, rName, tagKey1, tagValue1))
}

func testAccEntityRecognizerConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccEntityRecognizerBasicRoleConfig(rName),
		testAccEntityRecognizerS3BucketConfig(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_comprehend_entity_recognizer" "test" {
  name = %[1]q

  data_access_role_arn = aws_iam_role.test.arn

  tags = {
	%[2]q = %[3]q
	%[4]q = %[5]q
}

  language_code = "en"
  input_data_config {
    entity_types {
      type = "ENGINEER"
    }
    entity_types {
      type = "MANAGER"
    }

    documents {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
    }

    entity_list {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.entities.id}"
    }
  }

  depends_on = [
    aws_iam_role_policy.test
  ]
}

resource "aws_s3_object" "documents" {
  bucket = aws_s3_bucket.test.bucket
  key    = "documents.txt"
  source = "test-fixtures/entity_recognizer/documents.txt"
}

resource "aws_s3_object" "entities" {
  bucket = aws_s3_bucket.test.bucket
  key    = "entitylist.csv"
  source = "test-fixtures/entity_recognizer/entitylist.csv"
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccEntityRecognizerS3BucketConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_public_access_block" "test" {
  bucket = aws_s3_bucket.test.bucket

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}
`, rName)
}

func testAccEntityRecognizerBasicRoleConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "comprehend.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = data.aws_iam_policy_document.role.json
}


data "aws_iam_policy_document" "role" {
  statement {
    actions = [
      "s3:GetObject",
    ]

    resources = [
      "${aws_s3_bucket.test.arn}/*",
    ]
  }
  statement {
    actions = [
      "s3:ListBucket",
    ]

    resources = [
      aws_s3_bucket.test.arn,
    ]
  }
}
`, rName)
}
