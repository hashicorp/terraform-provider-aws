// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent_test

import (
	"context"
	"fmt"
	"testing"

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

// Prerequisites:
// * psql run via null_resource/provisioner "local-exec"
// * jq for parsing output from aws cli to retrieve postgres password
func testAccKnowledgeBase_basicRDS(t *testing.T) {
	acctest.SkipIfExeNotOnPath(t, "psql")
	acctest.SkipIfExeNotOnPath(t, "jq")
	acctest.SkipIfExeNotOnPath(t, "aws")

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
		ExternalProviders: map[string]resource.ExternalProvider{
			"null": {
				Source:            "hashicorp/null",
				VersionConstraint: "3.2.2",
			},
		},
		CheckDestroy: testAccCheckKnowledgeBaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKnowledgeBaseConfig_basicRDS(rName, foundationModel),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, resourceName, &knowledgebase),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.0.vector_knowledge_base_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.0.type", "VECTOR"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.type", "RDS"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.rds_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.rds_configuration.0.table_name", "bedrock_integration.bedrock_kb"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.rds_configuration.0.field_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.rds_configuration.0.field_mapping.0.vector_field", "embedding"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.rds_configuration.0.field_mapping.0.text_field", "chunks"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.rds_configuration.0.field_mapping.0.metadata_field", "metadata"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.rds_configuration.0.field_mapping.0.primary_key_field", names.AttrID),
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

// Prerequisites:
// * psql run via null_resource/provisioner "local-exec"
// * jq for parsing output from aws cli to retrieve postgres password
func testAccKnowledgeBase_disappears(t *testing.T) {
	acctest.SkipIfExeNotOnPath(t, "psql")
	acctest.SkipIfExeNotOnPath(t, "jq")
	acctest.SkipIfExeNotOnPath(t, "aws")

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
		ExternalProviders: map[string]resource.ExternalProvider{
			"null": {
				Source:            "hashicorp/null",
				VersionConstraint: "3.2.2",
			},
		},
		CheckDestroy: testAccCheckKnowledgeBaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKnowledgeBaseConfig_basicRDS(rName, foundationModel),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, resourceName, &knowledgebase),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbedrockagent.ResourceKnowledgeBase, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Prerequisites:
// * psql run via null_resource/provisioner "local-exec"
// * jq for parsing output from aws cli to retrieve postgres password
func testAccKnowledgeBase_tags(t *testing.T) {
	acctest.SkipIfExeNotOnPath(t, "psql")
	acctest.SkipIfExeNotOnPath(t, "jq")
	acctest.SkipIfExeNotOnPath(t, "aws")

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
		ExternalProviders: map[string]resource.ExternalProvider{
			"null": {
				Source:            "hashicorp/null",
				VersionConstraint: "3.2.2",
			},
		},
		CheckDestroy: testAccCheckKnowledgeBaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKnowledgeBaseConfig_tags1(rName, foundationModel, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, resourceName, &knowledgebase),
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
				Config: testAccKnowledgeBaseConfig_tags2(rName, foundationModel, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, resourceName, &knowledgebase),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccKnowledgeBaseConfig_tags1(rName, foundationModel, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, resourceName, &knowledgebase),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccKnowledgeBase_basicOpenSearch(t *testing.T) {
	acctest.Skip(t, "Bedrock Agent Knowledge Base requires external configuration of a vector index")

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
				Config: testAccKnowledgeBaseConfig_basicOpenSearch(rName, foundationModel),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, resourceName, &knowledgebase),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.0.vector_knowledge_base_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.0.type", "VECTOR"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.type", "OPENSEARCH_SERVERLESS"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.0.vector_index_name", "bedrock-knowledge-base-default-index"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.0.field_mapping.#", acctest.Ct1),
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

func testAccKnowledgeBase_updateOpenSearch(t *testing.T) {
	acctest.Skip(t, "Bedrock Agent Knowledge Base requires external configuration of a vector index")

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
				Config: testAccKnowledgeBaseConfig_basicOpenSearch(rName, foundationModel),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, resourceName, &knowledgebase),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.0.vector_knowledge_base_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.0.type", "VECTOR"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.type", "OPENSEARCH_SERVERLESS"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.0.vector_index_name", "bedrock-knowledge-base-default-index"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.0.field_mapping.#", acctest.Ct1),
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
				Config: testAccKnowledgeBaseConfig_updateOpenSearch(rName, foundationModel),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, resourceName, &knowledgebase),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName+"-updated"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.0.vector_knowledge_base_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.0.type", "VECTOR"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.type", "OPENSEARCH_SERVERLESS"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.0.vector_index_name", "bedrock-knowledge-base-default-index"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.opensearch_serverless_configuration.0.field_mapping.#", acctest.Ct1),
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

			return fmt.Errorf("Bedrock Agent Knowledge Base %s still exists", rs.Primary.ID)
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

func testAccKnowledgeBase_baseRDS(rName, model string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnetsEnableDNSHostnames(rName, 2), fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_iam_role" "test" {
  name               = %[1]q
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

# See https://docs.aws.amazon.com/bedrock/latest/userguide/kb-permissions.html.
resource "aws_iam_role_policy" "test" {
  name   = %[1]q
  role   = aws_iam_role.test.name
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:ListFoundationModels",
        "bedrock:ListCustomModels"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeModel"
      ],
      "Resource": [
        "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}::foundation-model/%[2]s"
      ]
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "rds_data_full_access" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_partition.current.partition}:policy/AmazonRDSDataFullAccess"
}

resource "aws_iam_role_policy_attachment" "secrets_manager_read_write" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_partition.current.partition}:policy/SecretsManagerReadWrite"
}

resource "aws_rds_cluster" "test" {
  cluster_identifier          = %[1]q
  engine                      = "aurora-postgresql"
  engine_mode                 = "provisioned"
  engine_version              = "15.4"
  database_name               = "test"
  master_username             = "test"
  manage_master_user_password = true
  enable_http_endpoint        = true
  vpc_security_group_ids      = [aws_security_group.test.id]
  skip_final_snapshot         = true
  db_subnet_group_name        = aws_db_subnet_group.test.name

  serverlessv2_scaling_configuration {
    max_capacity = 1.0
    min_capacity = 0.5
  }
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier  = aws_rds_cluster.test.id
  instance_class      = "db.serverless"
  engine              = aws_rds_cluster.test.engine
  engine_version      = aws_rds_cluster.test.engine_version
  publicly_accessible = true
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}

data "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_rds_cluster.test.master_user_secret[0].secret_arn
  version_stage = "AWSCURRENT"
  depends_on    = [aws_rds_cluster.test]
}

resource "null_resource" "db_setup" {
  depends_on = [aws_rds_cluster_instance.test, aws_rds_cluster.test, data.aws_secretsmanager_secret_version.test]

  provisioner "local-exec" {
    command = <<EOT
      sleep 60
      export PGPASSWORD=$(aws secretsmanager get-secret-value --secret-id '${aws_rds_cluster.test.master_user_secret[0].secret_arn}' --version-stage AWSCURRENT --region ${data.aws_region.current.name} --query SecretString --output text | jq -r '."password"')
      psql -h ${aws_rds_cluster.test.endpoint} -U ${aws_rds_cluster.test.master_username} -d ${aws_rds_cluster.test.database_name} -c "CREATE EXTENSION IF NOT EXISTS vector;"
      psql -h ${aws_rds_cluster.test.endpoint} -U ${aws_rds_cluster.test.master_username} -d ${aws_rds_cluster.test.database_name} -c "CREATE SCHEMA IF NOT EXISTS bedrock_integration;"
      psql -h ${aws_rds_cluster.test.endpoint} -U ${aws_rds_cluster.test.master_username} -d ${aws_rds_cluster.test.database_name} -c "CREATE SCHEMA IF NOT EXISTS bedrock_new;"
      psql -h ${aws_rds_cluster.test.endpoint} -U ${aws_rds_cluster.test.master_username} -d ${aws_rds_cluster.test.database_name} -c "CREATE ROLE bedrock_user WITH PASSWORD '$PGPASSWORD' LOGIN;"
      psql -h ${aws_rds_cluster.test.endpoint} -U ${aws_rds_cluster.test.master_username} -d ${aws_rds_cluster.test.database_name} -c "GRANT ALL ON SCHEMA bedrock_integration TO bedrock_user;"
      psql -h ${aws_rds_cluster.test.endpoint} -U ${aws_rds_cluster.test.master_username} -d ${aws_rds_cluster.test.database_name} -c "CREATE TABLE bedrock_integration.bedrock_kb (id uuid PRIMARY KEY, embedding vector(1536), chunks text, metadata json);"
      psql -h ${aws_rds_cluster.test.endpoint} -U ${aws_rds_cluster.test.master_username} -d ${aws_rds_cluster.test.database_name} -c "CREATE INDEX ON bedrock_integration.bedrock_kb USING hnsw (embedding vector_cosine_ops);"
    EOT
  }
}
`, rName, model))
}

func testAccKnowledgeBaseConfig_basicRDS(rName, model string) string {
	return acctest.ConfigCompose(testAccKnowledgeBase_baseRDS(rName, model), fmt.Sprintf(`
resource "aws_bedrockagent_knowledge_base" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  knowledge_base_configuration {
    vector_knowledge_base_configuration {
      embedding_model_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}::foundation-model/%[2]s"
    }
    type = "VECTOR"
  }

  storage_configuration {
    type = "RDS"
    rds_configuration {
      resource_arn           = aws_rds_cluster.test.arn
      credentials_secret_arn = tolist(aws_rds_cluster.test.master_user_secret)[0].secret_arn
      database_name          = aws_rds_cluster.test.database_name
      table_name             = "bedrock_integration.bedrock_kb"
      field_mapping {
        vector_field      = "embedding"
        text_field        = "chunks"
        metadata_field    = "metadata"
        primary_key_field = "id"
      }
    }
  }

  depends_on = [aws_iam_role_policy.test, null_resource.db_setup]
}
`, rName, model))
}

func testAccKnowledgeBaseConfig_tags1(rName, model, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(testAccKnowledgeBase_baseRDS(rName, model), fmt.Sprintf(`
resource "aws_bedrockagent_knowledge_base" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  knowledge_base_configuration {
    vector_knowledge_base_configuration {
      embedding_model_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}::foundation-model/%[2]s"
    }
    type = "VECTOR"
  }

  storage_configuration {
    type = "RDS"
    rds_configuration {
      resource_arn           = aws_rds_cluster.test.arn
      credentials_secret_arn = tolist(aws_rds_cluster.test.master_user_secret)[0].secret_arn
      database_name          = aws_rds_cluster.test.database_name
      table_name             = "bedrock_integration.bedrock_kb"
      field_mapping {
        vector_field      = "embedding"
        text_field        = "chunks"
        metadata_field    = "metadata"
        primary_key_field = "id"
      }
    }
  }

  tags = {
    %[3]q = %[4]q
  }

  depends_on = [aws_iam_role_policy.test, null_resource.db_setup]
}
`, rName, model, tag1Key, tag1Value))
}

func testAccKnowledgeBaseConfig_tags2(rName, model, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(testAccKnowledgeBase_baseRDS(rName, model), fmt.Sprintf(`
resource "aws_bedrockagent_knowledge_base" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  knowledge_base_configuration {
    vector_knowledge_base_configuration {
      embedding_model_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}::foundation-model/%[2]s"
    }
    type = "VECTOR"
  }

  storage_configuration {
    type = "RDS"
    rds_configuration {
      resource_arn           = aws_rds_cluster.test.arn
      credentials_secret_arn = tolist(aws_rds_cluster.test.master_user_secret)[0].secret_arn
      database_name          = aws_rds_cluster.test.database_name
      table_name             = "bedrock_integration.bedrock_kb"
      field_mapping {
        vector_field      = "embedding"
        text_field        = "chunks"
        metadata_field    = "metadata"
        primary_key_field = "id"
      }
    }
  }

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }

  depends_on = [aws_iam_role_policy.test, null_resource.db_setup]
}
`, rName, model, tag1Key, tag1Value, tag2Key, tag2Value))
}

func testAccKnowledgeBaseConfig_baseOpenSearch(rName, model string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_opensearchserverless_security_policy" "test" {
  name = %[1]q
  type = "encryption"

  policy = jsonencode({
    "Rules" = [
      {
        "Resource" = [
          "collection/%[1]s"
        ],
        "ResourceType" = "collection"
      }
    ],
    "AWSOwnedKey" = true
  })
}

resource "aws_opensearchserverless_collection" "test" {
  name = %[1]q

  depends_on = [aws_opensearchserverless_security_policy.test]
}

resource "aws_iam_role" "test" {
  name               = %[1]q
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

# See https://docs.aws.amazon.com/bedrock/latest/userguide/kb-permissions.html.
resource "aws_iam_role_policy" "test" {
  name   = %[1]q
  role   = aws_iam_role.test.name
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:ListFoundationModels",
        "bedrock:ListCustomModels"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeModel"
      ],
      "Resource": [
        "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}::foundation-model/%[2]s"
      ]
    },
    {
      "Action": [
        "aoss:APIAccessAll"
      ],
      "Effect": "Allow",
      "Resource": [
        "${aws_opensearchserverless_collection.test.arn}"
      ]
    }
  ]
}
POLICY
}

`, rName, model)
}

func testAccKnowledgeBaseConfig_basicOpenSearch(rName, model string) string {
	return acctest.ConfigCompose(testAccKnowledgeBaseConfig_baseOpenSearch(rName, model), fmt.Sprintf(`
resource "aws_bedrockagent_knowledge_base" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  knowledge_base_configuration {
    vector_knowledge_base_configuration {
      embedding_model_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}::foundation-model/%[2]s"
    }
    type = "VECTOR"
  }

  storage_configuration {
    type = "OPENSEARCH_SERVERLESS"
    opensearch_serverless_configuration {
      collection_arn    = aws_opensearchserverless_collection.test.arn
      vector_index_name = "bedrock-knowledge-base-default-index"
      field_mapping {
        vector_field   = "bedrock-knowledge-base-default-vector"
        text_field     = "AMAZON_BEDROCK_TEXT_CHUNK"
        metadata_field = "AMAZON_BEDROCK_METADATA"
      }
    }
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, model))
}

func testAccKnowledgeBaseConfig_updateOpenSearch(rName, model string) string {
	return acctest.ConfigCompose(testAccKnowledgeBaseConfig_baseOpenSearch(rName, model), fmt.Sprintf(`
resource "aws_bedrockagent_knowledge_base" "test" {
  name        = "%[1]s-updated"
  description = %[1]q
  role_arn    = aws_iam_role.test.arn

  knowledge_base_configuration {
    vector_knowledge_base_configuration {
      embedding_model_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}::foundation-model/%[2]s"
    }
    type = "VECTOR"
  }

  storage_configuration {
    type = "OPENSEARCH_SERVERLESS"
    opensearch_serverless_configuration {
      collection_arn    = aws_opensearchserverless_collection.test.arn
      vector_index_name = "bedrock-knowledge-base-default-index"
      field_mapping {
        vector_field   = "bedrock-knowledge-base-default-vector"
        text_field     = "AMAZON_BEDROCK_TEXT_CHUNK"
        metadata_field = "AMAZON_BEDROCK_METADATA"
      }
    }
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, model))
}
