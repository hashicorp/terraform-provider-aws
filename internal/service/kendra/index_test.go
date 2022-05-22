package kendra_test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kendra"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkendra "github.com/hashicorp/terraform-provider-aws/internal/service/kendra"
)

func TestAccKendraIndex_basic(t *testing.T) {
	var index kendra.DescribeIndexOutput

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	description := "basic"
	resourceName := "aws_kendra_index.test"

	propagationSleep := func() resource.TestCheckFunc {
		return func(s *terraform.State) error {
			log.Print("[DEBUG] Test: Sleep to allow IAM role to become visible to Kendra")
			time.Sleep(30 * time.Second)
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(kendra.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, kendra.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexBaseConfig(rName),
				Check:  propagationSleep(),
			},
			{
				Config: testAccIndexConfig_basic(rName, rName2, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(resourceName, &index),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "capacity_units.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_units.0.query_capacity_units", "0"),
					resource.TestCheckResourceAttr(resourceName, "capacity_units.0.storage_capacity_units", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "document_metadata_configuration_updates.#", "13"),
					resource.TestCheckResourceAttr(resourceName, "edition", kendra.IndexEditionEnterpriseEdition),
					resource.TestCheckResourceAttr(resourceName, "index_statistics.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "index_statistics.0.faq_statistics.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "index_statistics.0.faq_statistics.0.indexed_question_answers_count"),
					resource.TestCheckResourceAttr(resourceName, "index_statistics.0.text_document_statistics.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "index_statistics.0.text_document_statistics.0.indexed_text_bytes"),
					resource.TestCheckResourceAttrSet(resourceName, "index_statistics.0.text_document_statistics.0.indexed_text_documents_count"),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.access_cw", "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", kendra.IndexStatusActive),
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

func TestAccKendraIndex_updateDescription(t *testing.T) {
	var index kendra.DescribeIndexOutput

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	originalDescription := "original description"
	updatedDescription := "updated description"
	resourceName := "aws_kendra_index.test"

	propagationSleep := func() resource.TestCheckFunc {
		return func(s *terraform.State) error {
			log.Print("[DEBUG] Test: Sleep to allow IAM role to become visible to Kendra")
			time.Sleep(30 * time.Second)
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(kendra.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, kendra.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexBaseConfig(rName),
				Check:  propagationSleep(),
			},
			{
				Config: testAccIndexConfig_basic(rName, rName2, originalDescription),
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
				Config: testAccIndexConfig_basic(rName, rName2, updatedDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "description", updatedDescription),
				),
			},
		},
	})
}

func TestAccKendraIndex_updateName(t *testing.T) {
	var index kendra.DescribeIndexOutput

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	description := "description"
	resourceName := "aws_kendra_index.test"

	propagationSleep := func() resource.TestCheckFunc {
		return func(s *terraform.State) error {
			log.Print("[DEBUG] Test: Sleep to allow IAM role to become visible to Kendra")
			time.Sleep(30 * time.Second)
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(kendra.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, kendra.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexBaseConfig(rName),
				Check:  propagationSleep(),
			},
			{
				Config: testAccIndexConfig_basic(rName, rName2, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIndexConfig_basic(rName, rName3, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists(resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "name", rName3),
				),
			},
		},
	})
}

func TestAccKendraIndex_disappears(t *testing.T) {
	var index kendra.DescribeIndexOutput

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	description := "disappears"
	resourceName := "aws_kendra_index.test"

	propagationSleep := func() resource.TestCheckFunc {
		return func(s *terraform.State) error {
			log.Print("[DEBUG] Test: Sleep to allow IAM role to become visible to Kendra")
			time.Sleep(30 * time.Second)
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(kendra.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, kendra.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexBaseConfig(rName),
				Check:  propagationSleep(),
			},
			{
				Config: testAccIndexConfig_basic(rName, rName2, description),
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

		resp, err := conn.DescribeIndex(input)

		if err == nil {
			if aws.StringValue(resp.Id) == rs.Primary.ID {
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
		resp, err := conn.DescribeIndex(input)

		if err != nil {
			return err
		}

		*index = *resp

		return nil
	}
}

func testAccIndexBaseConfig(rName string) string {
	// Kendra IAM policies: https://docs.aws.amazon.com/kendra/latest/dg/iam-roles.html
	return fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}
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
          Resource = "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/kendra/*"
        },
        {
          Action = [
            "logs:DescribeLogStreams",
            "logs:CreateLogStream",
            "logs:PutLogEvents"
          ]
          Effect   = "Allow"
          Resource = "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/kendra/*:log-stream:*"
        },
      ]
    })
  }
}
`, rName)
}

func testAccIndexConfig_basic(rName, rName2, description string) string {
	return acctest.ConfigCompose(
		testAccIndexBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kendra_index" "test" {
  name        = %[1]q
  description = %[2]q
  role_arn    = aws_iam_role.access_cw.arn

  tags = {
    "Key1" = "Value1"
  }
}
`, rName2, description))
}
