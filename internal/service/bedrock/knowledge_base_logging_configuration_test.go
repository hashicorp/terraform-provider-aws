// knowledge_base_logging_configuration_test.go

package bedrock_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbedrock "github.com/hashicorp/terraform-provider-aws/internal/service/bedrock"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKnowledgeBaseLoggingConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrock_knowledge_base_logging_configuration.test"
	logGroupResourceName := "aws_cloudwatch_log_group.knowledge_base_logs"
	iamRoleResourceName := "aws_iam_role.bedrock_logging_role"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKnowledgeBaseLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKnowledgeBaseLoggingConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKnowledgeBaseLoggingConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_id", "kb-example-id"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.embedding_data_delivery_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "logging_config.cloudwatch_config.log_group_name", logGroupResourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "logging_config.cloudwatch_config.role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "Production"),
					resource.TestCheckResourceAttr(resourceName, "tags.Project", "BedrockKnowledgeBase"),
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

func testAccCheckKnowledgeBaseLoggingConfigurationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		knowledgeBaseID := rs.Primary.Attributes["knowledge_base_id"]

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockClient(ctx)

		_, err := tfbedrock.FindKnowledgeBaseLoggingConfiguration(ctx, conn, knowledgeBaseID)

		return err
	}
}

func testAccCheckKnowledgeBaseLoggingConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrock_knowledge_base_logging_configuration" {
				continue
			}

			knowledgeBaseID := rs.Primary.Attributes["knowledge_base_id"]

			_, err := tfbedrock.FindKnowledgeBaseLoggingConfiguration(ctx, conn, knowledgeBaseID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Knowledge Base Logging Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccKnowledgeBaseLoggingConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
variable "knowledge_base_id" {
  type    = string
  default = "kb-example-id"
}

data "aws_caller_identity" "current" {}

resource "aws_cloudwatch_log_group" "knowledge_base_logs" {
  name = "/aws/vendedlogs/bedrock/knowledge-base/%s"
}

resource "aws_iam_role" "bedrock_logging_role" {
  name = "bedrock_logging_role_%s"

  assume_role_policy = data.aws_iam_policy_document.bedrock_assume_role_policy.json
}

data "aws_iam_policy_document" "bedrock_assume_role_policy" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["bedrock.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_policy" "bedrock_logging_policy" {
  name   = "bedrock_logging_policy_%s"
  policy = data.aws_iam_policy_document.bedrock_logging_policy.json
}

data "aws_iam_policy_document" "bedrock_logging_policy" {
  statement {
    effect = "Allow"

    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = [
      "${aws_cloudwatch_log_group.knowledge_base_logs.arn}:*",
    ]
  }
}

resource "aws_iam_role_policy_attachment" "bedrock_logging_attachment" {
  role       = aws_iam_role.bedrock_logging_role.name
  policy_arn = aws_iam_policy.bedrock_logging_policy.arn
}

resource "aws_bedrock_knowledge_base_logging_configuration" "test" {
  knowledge_base_id = var.knowledge_base_id

  logging_config {
    embedding_data_delivery_enabled = true

    cloudwatch_config {
      log_group_name = aws_cloudwatch_log_group.knowledge_base_logs.name
      role_arn       = aws_iam_role.bedrock_logging_role.arn
    }
  }

  tags = {
    Environment = "Production"
    Project     = "BedrockKnowledgeBase"
  }
}
`, rName, rName, rName)
}
