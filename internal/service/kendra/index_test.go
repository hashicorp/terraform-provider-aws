package kendra_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kendra"
	"github.com/aws/aws-sdk-go-v2/service/kendra/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkendra "github.com/hashicorp/terraform-provider-aws/internal/service/kendra"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccPreCheck(t *testing.T) {
	acctest.PreCheckPartitionHasService(names.KendraEndpointID, t)

	conn := acctest.Provider.Meta().(*conns.AWSClient).KendraConn

	input := &kendra.ListIndicesInput{}

	_, err := conn.ListIndices(context.TODO(), input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccIndex_basic(t *testing.T) {
	var index kendra.DescribeIndexOutput

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	description := "basic"
	resourceName := "aws_kendra_index.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_basic(rName, rName2, rName3, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(resourceName, &index),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "capacity_units.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_units.0.query_capacity_units", "0"),
					resource.TestCheckResourceAttr(resourceName, "capacity_units.0.storage_capacity_units", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "document_metadata_configuration_updates.#", "13"),
					resource.TestCheckResourceAttr(resourceName, "edition", string(types.IndexEditionEnterpriseEdition)),
					resource.TestCheckResourceAttr(resourceName, "index_statistics.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "index_statistics.0.faq_statistics.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "index_statistics.0.faq_statistics.0.indexed_question_answers_count"),
					resource.TestCheckResourceAttr(resourceName, "index_statistics.0.text_document_statistics.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "index_statistics.0.text_document_statistics.0.indexed_text_bytes"),
					resource.TestCheckResourceAttrSet(resourceName, "index_statistics.0.text_document_statistics.0.indexed_text_documents_count"),
					resource.TestCheckResourceAttr(resourceName, "name", rName3),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.access_cw", "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", string(types.IndexStatusActive)),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttr(resourceName, "user_context_policy", "ATTRIBUTE_FILTER"),
					resource.TestCheckResourceAttr(resourceName, "user_group_resolution_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
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

func testAccIndex_serverSideEncryption(t *testing.T) {
	var index kendra.DescribeIndexOutput

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_index.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_serverSideEncryption(rName, rName2, rName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "server_side_encryption_configuration.0.kms_key_id", "data.aws_kms_key.this", "arn"),
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

func testAccIndex_updateCapacityUnits(t *testing.T) {
	var index kendra.DescribeIndexOutput

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	originalQueryCapacityUnits := 2
	updatedQueryCapacityUnits := 3
	originalStorageCapacityUnits := 1
	updatedStorageCapacityUnits := 2
	resourceName := "aws_kendra_index.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_capacityUnits(rName, rName2, rName3, originalQueryCapacityUnits, originalStorageCapacityUnits),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "capacity_units.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_units.0.query_capacity_units", strconv.Itoa(originalQueryCapacityUnits)),
					resource.TestCheckResourceAttr(resourceName, "capacity_units.0.storage_capacity_units", strconv.Itoa(originalStorageCapacityUnits)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIndexConfig_capacityUnits(rName, rName2, rName3, updatedQueryCapacityUnits, updatedStorageCapacityUnits),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "capacity_units.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_units.0.query_capacity_units", strconv.Itoa(updatedQueryCapacityUnits)),
					resource.TestCheckResourceAttr(resourceName, "capacity_units.0.storage_capacity_units", strconv.Itoa(updatedStorageCapacityUnits)),
				),
			},
		},
	})
}
func testAccIndex_updateDescription(t *testing.T) {
	var index kendra.DescribeIndexOutput

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	originalDescription := "original description"
	updatedDescription := "updated description"
	resourceName := "aws_kendra_index.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_basic(rName, rName2, rName3, originalDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "description", originalDescription),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIndexConfig_basic(rName, rName2, rName3, updatedDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "description", updatedDescription),
				),
			},
		},
	})
}

func testAccIndex_updateName(t *testing.T) {
	var index kendra.DescribeIndexOutput

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	description := "description"
	resourceName := "aws_kendra_index.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_basic(rName, rName2, rName3, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "name", rName3),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIndexConfig_basic(rName, rName2, rName4, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "name", rName4),
				),
			},
		},
	})
}

func testAccIndex_updateUserTokenJSON(t *testing.T) {
	var index kendra.DescribeIndexOutput

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	originalGroupAttributeField := "groups"
	originalUserNameAttributeField := "username"
	updatedGroupAttributeField := "groupings"
	updatedUserNameAttributeField := "usernames"
	resourceName := "aws_kendra_index.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_userTokenJSON(rName, rName2, rName3, originalGroupAttributeField, originalUserNameAttributeField),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "user_token_configurations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_token_configurations.0.json_token_type_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_token_configurations.0.json_token_type_configuration.0.group_attribute_field", originalGroupAttributeField),
					resource.TestCheckResourceAttr(resourceName, "user_token_configurations.0.json_token_type_configuration.0.user_name_attribute_field", originalUserNameAttributeField),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIndexConfig_userTokenJSON(rName, rName2, rName3, updatedGroupAttributeField, originalUserNameAttributeField),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "user_token_configurations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_token_configurations.0.json_token_type_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_token_configurations.0.json_token_type_configuration.0.group_attribute_field", updatedGroupAttributeField),
					resource.TestCheckResourceAttr(resourceName, "user_token_configurations.0.json_token_type_configuration.0.user_name_attribute_field", originalUserNameAttributeField),
				),
			},
			{
				Config: testAccIndexConfig_userTokenJSON(rName, rName2, rName3, updatedGroupAttributeField, updatedUserNameAttributeField),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "user_token_configurations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_token_configurations.0.json_token_type_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_token_configurations.0.json_token_type_configuration.0.group_attribute_field", updatedGroupAttributeField),
					resource.TestCheckResourceAttr(resourceName, "user_token_configurations.0.json_token_type_configuration.0.user_name_attribute_field", updatedUserNameAttributeField),
				),
			},
		},
	})
}

func testAccIndex_updateTags(t *testing.T) {
	var index kendra.DescribeIndexOutput

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	description := "description"
	resourceName := "aws_kendra_index.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_basic(rName, rName2, rName3, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIndexConfig_tags(rName, rName2, rName3, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
				),
			},
			{
				Config: testAccIndexConfig_tagsUpdated(rName, rName2, rName3, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
				),
			},
		},
	})
}

func testAccIndex_updateRoleARN(t *testing.T) {
	var index kendra.DescribeIndexOutput

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	description := "description"
	resourceName := "aws_kendra_index.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_basic(rName, rName2, rName3, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(resourceName, &index),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.access_cw", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIndexConfig_secretsManagerRole(rName, rName2, rName3, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(resourceName, &index),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.access_sm", "arn"),
				),
			},
		},
	})
}

func testAccIndex_disappears(t *testing.T) {
	var index kendra.DescribeIndexOutput

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	description := "disappears"
	resourceName := "aws_kendra_index.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_basic(rName, rName2, rName3, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(resourceName, &index),
					acctest.CheckResourceDisappears(acctest.Provider, tfkendra.ResourceIndex(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckIndexDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KendraConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kendra_index" {
			continue
		}

		input := &kendra.DescribeIndexInput{
			Id: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeIndex(context.TODO(), input)

		if err == nil {
			if aws.ToString(resp.Id) == rs.Primary.ID {
				return fmt.Errorf("Index '%s' was not deleted properly", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckIndexExists(name string, index *kendra.DescribeIndexOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KendraConn
		input := &kendra.DescribeIndexInput{
			Id: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeIndex(context.TODO(), input)

		if err != nil {
			return err
		}

		*index = *resp

		return nil
	}
}

func testAccIndexConfigBase(rName, rName2 string) string {
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

resource "aws_iam_role" "access_cw" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json

  inline_policy {
    name = "access_cw"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action   = ["cloudwatch:PutMetricData"]
          Effect   = "Allow"
          Resource = "*"
          Condition = {
            StringEquals = {
              "cloudwatch:namespace" = "Kendra"
            }
          }
        },
        {
          Action   = ["logs:DescribeLogGroups"]
          Effect   = "Allow"
          Resource = "*"
        },
        {
          Action   = ["logs:CreateLogGroup"]
          Effect   = "Allow"
          Resource = "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/kendra/*"
        },
        {
          Action = [
            "logs:DescribeLogStreams",
            "logs:CreateLogStream",
            "logs:PutLogEvents"
          ]
          Effect   = "Allow"
          Resource = "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/kendra/*:log-stream:*"
        },
      ]
    })
  }
}

resource "aws_iam_role" "access_sm" {
  name               = %[2]q
  assume_role_policy = data.aws_iam_policy_document.test.json

  inline_policy {
    name = "access_sm"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action   = ["cloudwatch:PutMetricData"]
          Effect   = "Allow"
          Resource = "*"
          Condition = {
            StringEquals = {
              "cloudwatch:namespace" = "Kendra"
            }
          }
        },
        {
          Action   = ["logs:DescribeLogGroups"]
          Effect   = "Allow"
          Resource = "*"
        },
        {
          Action   = ["logs:CreateLogGroup"]
          Effect   = "Allow"
          Resource = "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/kendra/*"
        },
        {
          Action = [
            "logs:DescribeLogStreams",
            "logs:CreateLogStream",
            "logs:PutLogEvents"
          ]
          Effect   = "Allow"
          Resource = "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/kendra/*:log-stream:*"
        },
        {
          Action   = ["secretsmanager:GetSecretValue"]
          Effect   = "Allow"
          Resource = "arn:${data.aws_partition.current.partition}:secretsmanager:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:secret:example"
        },
        {
          Action   = ["kms:Decrypt"]
          Effect   = "Allow"
          Resource = "arn:${data.aws_partition.current.partition}:kms:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:key/example"
          Condition = {
            StringLike = {
              "kms:ViaService" = ["secretsmanager.*.amazonaws.com"]
            }
          }
        }
      ]
    })
  }
}
`, rName, rName2)
}

func testAccIndexConfig_basic(rName, rName2, rName3, description string) string {
	return acctest.ConfigCompose(
		testAccIndexConfigBase(rName, rName2),
		fmt.Sprintf(`
resource "aws_kendra_index" "test" {
  name        = %[1]q
  description = %[2]q
  role_arn    = aws_iam_role.access_cw.arn

  tags = {
    "Key1" = "Value1"
  }
}
`, rName3, description))
}

func testAccIndexConfig_capacityUnits(rName, rName2, rName3 string, queryCapacityUnits, storageCapacityUnits int) string {
	return acctest.ConfigCompose(
		testAccIndexConfigBase(rName, rName2),
		fmt.Sprintf(`
resource "aws_kendra_index" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.access_cw.arn

  capacity_units {
    query_capacity_units   = %[2]d
    storage_capacity_units = %[3]d
  }

  tags = {
    "Key1" = "Value1"
  }
}
`, rName3, queryCapacityUnits, storageCapacityUnits))
}

func testAccIndexConfig_secretsManagerRole(rName, rName2, rName3, description string) string {
	return acctest.ConfigCompose(
		testAccIndexConfigBase(rName, rName2),
		fmt.Sprintf(`
resource "aws_kendra_index" "test" {
  name        = %[1]q
  description = %[2]q
  role_arn    = aws_iam_role.access_sm.arn

  tags = {
    "Key1" = "Value1"
  }
}
`, rName3, description))
}

func testAccIndexConfig_serverSideEncryption(rName, rName2, rName3 string) string {
	return acctest.ConfigCompose(
		testAccIndexConfigBase(rName, rName2),
		fmt.Sprintf(`
resource "aws_kendra_index" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.access_cw.arn

  server_side_encryption_configuration {
    kms_key_id = data.aws_kms_key.this.arn
  }
}
`, rName3))
}

func testAccIndexConfig_userTokenJSON(rName, rName2, rName3, groupAttributeField, userNameAttributeField string) string {
	return acctest.ConfigCompose(
		testAccIndexConfigBase(rName, rName2),
		fmt.Sprintf(`
resource "aws_kendra_index" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.access_cw.arn

  user_token_configurations {
    json_token_type_configuration {
      group_attribute_field     = %[2]q
      user_name_attribute_field = %[3]q
    }
  }
}
`, rName3, groupAttributeField, userNameAttributeField))
}

func testAccIndexConfig_tags(rName, rName2, rName3, description string) string {
	return acctest.ConfigCompose(
		testAccIndexConfigBase(rName, rName2),
		fmt.Sprintf(`
resource "aws_kendra_index" "test" {
  name        = %[1]q
  description = %[2]q
  role_arn    = aws_iam_role.access_cw.arn

  tags = {
    "Key1" = "Value1"
    "Key2" = "Value2a",
  }
}
`, rName3, description))
}

func testAccIndexConfig_tagsUpdated(rName, rName2, rName3, description string) string {
	return acctest.ConfigCompose(
		testAccIndexConfigBase(rName, rName2),
		fmt.Sprintf(`
resource "aws_kendra_index" "test" {
  name        = %[1]q
  description = %[2]q
  role_arn    = aws_iam_role.access_cw.arn

  tags = {
    "Key1" = "Value1",
    "Key2" = "Value2b"
    "Key3" = "Value3"
  }
}
`, rName3, description))
}
