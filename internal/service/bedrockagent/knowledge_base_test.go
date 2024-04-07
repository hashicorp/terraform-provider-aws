// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbedrockagent "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagent"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccKnowledgeBase_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var knowledgebase types.KnowledgeBase
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_knowledge_base.test"
	foundationModel := "amazon.titan-embed-text-v1"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKnowledgeBaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKnowledgeBase_base(rName, foundationModel),
				// Withought Sleep role couldnt assumed.
				Check: acctest.CheckSleep(t, 5*time.Second),
			},
			{
				Config: testAccKnowledgeBaseConfig_basic(rName, foundationModel),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, resourceName, &knowledgebase),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.0.vector_knowledge_base_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.0.type", "VECTOR"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.type", "OPENSEARCH_SERVERLESS"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.0.vector_index_name", "bedrock-knowledge-base-default-index"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.0.field_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.0.field_mapping.0.vector_field", "bedrock-knowledge-base-default-vector"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.0.field_mapping.0.text_field", "AMAZON_BEDROCK_TEXT_CHUNK"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.0.field_mapping.0.metadata_field", "AMAZON_BEDROCK_METADATA"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccKnowledgeBase_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var knowledgebase types.KnowledgeBase
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_knowledge_base.test"
	foundationModel := "amazon.titan-embed-g1-text-02"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKnowledgeBaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKnowledgeBase_base(rName, foundationModel),
				// Withought Sleep role couldnt assumed.
				Check: acctest.CheckSleep(t, 5*time.Second),
			},
			{
				Config: testAccKnowledgeBaseConfig_basic(rName, foundationModel),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, resourceName, &knowledgebase),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbedrockagent.ResourceKnowledgeBase, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.0.vector_knowledge_base_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.0.type", "VECTOR"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.type", "OPENSEARCH_SERVERLESS"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.0.vector_index_name", "bedrock-knowledge-base-default-index"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.0.field_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.0.field_mapping.0.vector_field", "bedrock-knowledge-base-default-vector"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.0.field_mapping.0.text_field", "AMAZON_BEDROCK_TEXT_CHUNK"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.0.field_mapping.0.metadata_field", "AMAZON_BEDROCK_METADATA"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccKnowledgeBase_update(t *testing.T) {
	ctx := acctest.Context(t)
	var knowledgebase types.KnowledgeBase
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_knowledge_base.test"
	foundationModel := "amazon.titan-embed-g1-text-02"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKnowledgeBaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKnowledgeBase_base(rName, foundationModel),
				// Withought Sleep role couldnt assumed.
				Check: acctest.CheckSleep(t, 5*time.Second),
			},
			{
				Config: testAccKnowledgeBaseConfig_basic(rName, foundationModel),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, resourceName, &knowledgebase),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.0.vector_knowledge_base_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.0.type", "VECTOR"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.type", "OPENSEARCH_SERVERLESS"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.0.vector_index_name", "bedrock-knowledge-base-default-index"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.0.field_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.0.field_mapping.0.vector_field", "bedrock-knowledge-base-default-vector"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.0.field_mapping.0.text_field", "AMAZON_BEDROCK_TEXT_CHUNK"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.0.field_mapping.0.metadata_field", "AMAZON_BEDROCK_METADATA"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testAccKnowledgeBaseConfig_update(rName, foundationModel),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, resourceName, &knowledgebase),
					resource.TestCheckResourceAttr(resourceName, "name", "updated-name"),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.0.vector_knowledge_base_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.0.type", "VECTOR"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.type", "OPENSEARCH_SERVERLESS"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.0.vector_index_name", "bedrock-knowledge-base-default-index"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.0.field_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.0.field_mapping.0.vector_field", "bedrock-knowledge-base-default-vector"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.0.field_mapping.0.text_field", "AMAZON_BEDROCK_TEXT_CHUNK"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.0.field_mapping.0.metadata_field", "AMAZON_BEDROCK_METADATA"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccCheckKnowledgeBaseDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagent_knowledge_base" {
				continue
			}

			_, err := tfbedrockagent.FindKnowledgeBaseByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent knowledge base %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckKnowledgeBaseExists(ctx context.Context, n string, v *types.KnowledgeBase) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		output, err := tfbedrockagent.FindKnowledgeBaseByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccKnowledgeBase_base(rName, model string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

resource "aws_iam_role" "test" {
	name               = "AmazonBedrockExecutionRoleForKnowledgeBase_tf"
	path               = "/service-role/"
	assume_role_policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [{
		"Action": "sts:AssumeRole",
		"Principal": {
		"Service": "bedrock.amazonaws.com"
		},
		"Effect": "Allow"
	}]
}
POLICY
}

resource "aws_iam_role_policy" "policy" {
	name = "AmazonBedrockExecutionRoleForKnowledgeBasePolicy"
	role = aws_iam_role.test.name
	policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
	  {
		"Sid": "BedrockInvokeModelStatement",
		"Effect": "Allow",
		"Action": [
			"bedrock:InvokeModel"
		],
		"Resource": [
			"arn:aws:bedrock:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:foundation-model/%[2]s"
		]
	  },
	  {
		"Sid": "OpenSearchServerlessAPIAccessAllStatement",
		"Action": [
			"aoss:APIAccessAll"
		],
		"Effect": "Allow",
		"Resource": [
			"arn:aws:aoss:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:collection/142bezjddq707i5stcrf"
		]
	  }
	]
}
POLICY
}

`, rName, model)
}

func testAccKnowledgeBaseConfig_basic(rName, model string) string {
	return acctest.ConfigCompose(testAccKnowledgeBase_base(rName, model), fmt.Sprintf(`
	
resource "aws_bedrockagent_knowledge_base" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  knowledge_base_configuration {
    vector_knowledge_base_configuration {
      embedding_model_arn = "arn:aws:bedrock:${data.aws_region.current.name}::foundation-model/%[2]s"
    }
    type = "VECTOR"
  }
  storage_configuration {
    type = "OPENSEARCH_SERVERLESS"
    opensearch_serverless_configuration {
      collection_arn    = "arn:aws:aoss:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:collection/142bezjddq707i5stcrf"
      vector_index_name = "bedrock-knowledge-base-default-index"
      field_mapping {
		vector_field   = "bedrock-knowledge-base-default-vector"
        text_field     = "AMAZON_BEDROCK_TEXT_CHUNK"
        metadata_field = "AMAZON_BEDROCK_METADATA"
      }
    }
  }
  depends_on = [aws_iam_role.test]
}
`, rName, model))
}

func testAccKnowledgeBaseConfig_update(rName, model string) string {
	return acctest.ConfigCompose(testAccKnowledgeBase_base(rName, model), fmt.Sprintf(`
	
resource "aws_bedrockagent_knowledge_base" "test" {
  name     = "updated-name"
  description = %[1]q
  role_arn = aws_iam_role.test.arn
  knowledge_base_configuration {
    vector_knowledge_base_configuration {
      embedding_model_arn = "arn:aws:bedrock:${data.aws_region.current.name}::foundation-model/%[2]s"
    }
    type = "VECTOR"
  }
  storage_configuration {
    type = "OPENSEARCH_SERVERLESS"
    opensearch_serverless_configuration {
      collection_arn    = "arn:aws:aoss:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:collection/142bezjddq707i5stcrf"
      vector_index_name = "bedrock-knowledge-base-default-index"
      field_mapping {
		vector_field   = "bedrock-knowledge-base-default-vector"
        text_field     = "AMAZON_BEDROCK_TEXT_CHUNK"
        metadata_field = "AMAZON_BEDROCK_METADATA"
      }
    }
  }
  depends_on = [aws_iam_role.test]
}
`, rName, model))
}
