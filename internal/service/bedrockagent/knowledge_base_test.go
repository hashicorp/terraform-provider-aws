// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent_test

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbedrockagent "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagent"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Prerequisites:
// * psql run via null_resource/provisioner "local-exec"
// * jq for parsing output from aws cli to retrieve postgres password
func testAccKnowledgeBase_basic(t *testing.T) {
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
				Config: testAccKnowledgeBaseConfig_basicRDS(rName, foundationModel, ""),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, resourceName, &knowledgebase),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.0.vector_knowledge_base_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.0.type", "VECTOR"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.type", "RDS"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.rds_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.rds_configuration.0.table_name", "bedrock_integration.bedrock_kb"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.rds_configuration.0.field_mapping.#", "1"),
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
			{
				Config: testAccKnowledgeBaseConfig_updateRDS(rName, foundationModel, "test description"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, resourceName, &knowledgebase),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.0.vector_knowledge_base_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.0.type", "VECTOR"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName+"-update"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_update", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.type", "RDS"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.rds_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.rds_configuration.0.table_name", "bedrock_integration.bedrock_kb"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.rds_configuration.0.field_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.rds_configuration.0.field_mapping.0.vector_field", "embedding"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.rds_configuration.0.field_mapping.0.text_field", "chunks"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.rds_configuration.0.field_mapping.0.metadata_field", "metadata"),
					resource.TestCheckResourceAttr(resourceName, "storage_configuration.0.rds_configuration.0.field_mapping.0.primary_key_field", names.AttrID),
				),
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
				Config: testAccKnowledgeBaseConfig_basicRDS(rName, foundationModel, ""),
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
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, resourceName, &knowledgebase),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKnowledgeBaseConfig_tags2(rName, foundationModel, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, resourceName, &knowledgebase),
				),
			},
			{
				Config: testAccKnowledgeBaseConfig_tags1(rName, foundationModel, acctest.CtKey2, acctest.CtValue2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, resourceName, &knowledgebase),
				),
			},
		},
	})
}

func testAccKnowledgeBase_OpenSearch_basic(t *testing.T) {
	ctx := acctest.Context(t)
	collectionName := skipIfOSSCollectionNameEnvVarNotSet(t)

	var knowledgebase types.KnowledgeBase
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_knowledge_base.test"
	foundationModel := "amazon.titan-embed-text-v2:0"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKnowledgeBaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKnowledgeBaseConfig_OpenSearch_basic(rName, collectionName, foundationModel),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, resourceName, &knowledgebase),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
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
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccKnowledgeBase_OpenSearch_update(t *testing.T) {
	ctx := acctest.Context(t)
	collectionName := skipIfOSSCollectionNameEnvVarNotSet(t)

	var knowledgebase types.KnowledgeBase
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_knowledge_base.test"
	foundationModel := "amazon.titan-embed-text-v2:0"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKnowledgeBaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKnowledgeBaseConfig_OpenSearch_basic(rName, collectionName, foundationModel),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, resourceName, &knowledgebase),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
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
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKnowledgeBaseConfig_OpenSearch_update(rName, collectionName, foundationModel),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, resourceName, &knowledgebase),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName+"-updated"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
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
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccKnowledgeBase_OpenSearch_supplementalDataStorage(t *testing.T) {
	ctx := acctest.Context(t)
	collectionName := skipIfOSSCollectionNameEnvVarNotSet(t)

	var knowledgebase types.KnowledgeBase
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_knowledge_base.test"
	foundationModel := "amazon.titan-embed-text-v2:0"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKnowledgeBaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKnowledgeBaseConfig_OpenSearch_supplementalDataStorage(rName, collectionName, foundationModel),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, resourceName, &knowledgebase),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.0.vector_knowledge_base_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.0.type", "VECTOR"),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.0.vector_knowledge_base_configuration.0.embedding_model_configuration.0.bedrock_embedding_model_configuration.0.dimensions", "1024"),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.0.vector_knowledge_base_configuration.0.embedding_model_configuration.0.bedrock_embedding_model_configuration.0.embedding_data_type", "FLOAT32"),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.0.vector_knowledge_base_configuration.0.supplemental_data_storage_configuration.0.storage_location.0.type", "S3"),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_configuration.0.vector_knowledge_base_configuration.0.supplemental_data_storage_configuration.0.storage_location.0.s3_location.0.uri", fmt.Sprintf("s3://%s", rName)),
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

// skipIfOSSCollectionNameEnvVarNotSet handles skipping tests when an environment
// variable providing a valid OSS collection name is unset
//
// This should be called in all acceptance tests currently dependent on an OpenSearch
// Serverless vector collection.
//
// To create a collection to be used with this environment variable, do the following.
//
// 1. In the AWS console, navigate to the OpenSearch service. Choose the "Collections"
// entry on the left navbar and select "Create collection" above the collections table.
// 2. Enter a collection name. Choose "Vector search" as the collection type. Choose
// "Easy create" in the security section. Click "Next" to review, then click "Submit".
// 3. Once the collection is available, select the "Indexes" tab. Click "Create vector
// index". Name the index "bedrock-knowledge-base-default-index".
// 4. In the "Vector fields" section, click "Add vector field".  Name the field
// "bedrock-knowledge-base-default-vector". Choose "faiss" for engine, "FP32" for
// precision, "1024" for dimensions, and "Euclidean" for distance metric. Click
// "Confirm" to create the field.
//  5. In the "Metadata management" section, add two fields.
//     "AMAZON_BEDROCK_METADATA" - string type, filterable is false.
//     "AMAZON_BEDROCK_TEXT_CHUNK" - string type, filterable is true.
//  6. Click "Create" to finish index creation.
//
// At this point the collection is usable with this test. Set the collection name to the
// environment variable below.
func skipIfOSSCollectionNameEnvVarNotSet(t *testing.T) string {
	t.Helper()

	v := os.Getenv("TF_AWS_BEDROCK_OSS_COLLECTION_NAME")
	if v == "" {
		acctest.Skip(t, "This test requires external configuration of an OpenSearch collection vector index. "+
			"Set the TF_AWS_BEDROCK_OSS_COLLECTION_NAME environment variable to the OpenSearch collection name "+
			"where the vector index is configured.")
	}
	return v
}

func testAccKnowledgeBaseConfig_basicRDS(rName, model, description string) string {
	if description == "" {
		description = "null"
	} else {
		description = strconv.Quote(description)
	}

	return acctest.ConfigCompose(acctest.ConfigBedrockAgentKnowledgeBaseRDSBase(rName, model), fmt.Sprintf(`
resource "aws_bedrockagent_knowledge_base" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  description = %[3]s

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
`, rName, model, description))
}

func testAccKnowledgeBaseConfig_updateRDS(rName, model, description string) string {
	return acctest.ConfigCompose(acctest.ConfigBedrockAgentKnowledgeBaseRDSUpdateBase(rName, model), fmt.Sprintf(`
resource "aws_bedrockagent_knowledge_base" "test" {
  name     = "%[1]s-update"
  role_arn = aws_iam_role.test_update.arn

  description = %[3]q

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

  depends_on = [aws_iam_role_policy.test_update, null_resource.db_setup]
}
`, rName, model, description))
}

func testAccKnowledgeBaseConfig_tags1(rName, model, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(acctest.ConfigBedrockAgentKnowledgeBaseRDSBase(rName, model), fmt.Sprintf(`
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
	return acctest.ConfigCompose(acctest.ConfigBedrockAgentKnowledgeBaseRDSBase(rName, model), fmt.Sprintf(`
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

func testAccKnowledgeBaseConfigBase_openSearch(rName, collectionName, model string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "aws_partition" "current" {}

# As a future enhancment, the creation and preparation of the OSS collection
# will be done within this configuration. Creation of the collection is
# possible today, but creation of the appropriate vector index must be done
# out of band via awscurl or some other mechanism.
#
# Ref: https://docs.aws.amazon.com/opensearch-service/latest/developerguide/serverless-vector-search.html
data "aws_opensearchserverless_collection" "test" {
  name = %[2]q
}

# See the Amazon Bedrock documentation for creating a service role:
# https://docs.aws.amazon.com/bedrock/latest/userguide/kb-permissions.html
data "aws_iam_policy_document" "test_trust" {
  statement {
    effect = "Allow"
    actions = [
      "sts:AssumeRole",
    ]
    principals {
      type        = "Service"
      identifiers = ["bedrock.amazonaws.com"]
    }
    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values = [
        data.aws_caller_identity.current.account_id,
      ]
    }
    condition {
      test     = "ArnLike"
      variable = "aws:SourceArn"
      values = [
        "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:knowledge-base/*"
      ]
    }
  }
}

data "aws_iam_policy_document" "test" {
  statement {
    effect = "Allow"
    actions = [
      "bedrock:ListFoundationModels",
      "bedrock:ListCustomModels",
    ]
    resources = [
      "*",
    ]
  }

  statement {
    effect = "Allow"
    actions = [
      "bedrock:InvokeModel",
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:foundation-model/%[3]s",
    ]
  }

  statement {
    effect = "Allow"
    actions = [
      "bedrock:RetreiveAndGenerate",
    ]
    resources = [
      "*",
    ]
  }

  statement {
    effect = "Allow"
    actions = [
      "aoss:APIAccessAll",
    ]
    resources = [
      data.aws_opensearchserverless_collection.test.arn
    ]
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test_trust.json
}

resource "aws_iam_policy" "test" {
  name   = %[1]q
  policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test.arn
}

resource "aws_opensearchserverless_access_policy" "test" {
  name = %[1]q
  type = "data"
  policy = jsonencode([
    {
      Rules = [
        {
          ResourceType = "index",
          Resource = [
            "index/%[2]s/*"
          ],
          Permission = [
            "aoss:*",
          ]
        },
        {
          ResourceType = "collection",
          Resource = [
            "collection/%[2]s"
          ],
          Permission = [
            "aoss:*",
          ]
        }
      ],
      Principal = [
        data.aws_caller_identity.current.arn,
        aws_iam_role.test.arn,
      ]
    }
  ])
}
`, rName, collectionName, model)
}

func testAccKnowledgeBaseConfig_OpenSearch_basic(rName, collectionName, model string) string {
	return acctest.ConfigCompose(
		testAccKnowledgeBaseConfigBase_openSearch(rName, collectionName, model),
		fmt.Sprintf(`
resource "aws_bedrockagent_knowledge_base" "test" {
  depends_on = [
    aws_iam_role_policy_attachment.test,
    aws_opensearchserverless_access_policy.test,
  ]

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
      collection_arn    = data.aws_opensearchserverless_collection.test.arn
      vector_index_name = "bedrock-knowledge-base-default-index"
      field_mapping {
        vector_field   = "bedrock-knowledge-base-default-vector"
        text_field     = "AMAZON_BEDROCK_TEXT_CHUNK"
        metadata_field = "AMAZON_BEDROCK_METADATA"
      }
    }
  }
}
`, rName, model))
}

func testAccKnowledgeBaseConfig_OpenSearch_update(rName, collectionName, model string) string {
	return acctest.ConfigCompose(
		testAccKnowledgeBaseConfigBase_openSearch(rName, collectionName, model),
		fmt.Sprintf(`
resource "aws_bedrockagent_knowledge_base" "test" {
  depends_on = [
    aws_iam_role_policy_attachment.test,
    aws_opensearchserverless_access_policy.test,
  ]

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
      collection_arn    = data.aws_opensearchserverless_collection.test.arn
      vector_index_name = "bedrock-knowledge-base-default-index"
      field_mapping {
        vector_field   = "bedrock-knowledge-base-default-vector"
        text_field     = "AMAZON_BEDROCK_TEXT_CHUNK"
        metadata_field = "AMAZON_BEDROCK_METADATA"
      }
    }
  }
}
`, rName, model))
}

func testAccKnowledgeBaseConfig_OpenSearch_supplementalDataStorage(rName, collectionName, model string) string {
	return acctest.ConfigCompose(
		testAccKnowledgeBaseConfigBase_openSearch(rName, collectionName, model),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

data "aws_iam_policy_document" "test_s3" {
  statement {
    effect = "Allow"
    actions = [
      "s3:DeleteObject",
      "s3:GetObject",
      "s3:ListBucket",
      "s3:PutObject",
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.bucket}",
      "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.bucket}/*",
    ]
    condition {
      test     = "StringEquals"
      variable = "aws:PrincipalAccount"
      values = [
        data.aws_caller_identity.current.account_id,
      ]
    }
  }
}

resource "aws_iam_policy" "test_s3" {
  name   = "%[1]s-s3"
  policy = data.aws_iam_policy_document.test_s3.json
}

resource "aws_iam_role_policy_attachment" "test_s3" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test_s3.arn
}

resource "aws_bedrockagent_knowledge_base" "test" {
  depends_on = [
    aws_iam_role_policy_attachment.test,
    aws_iam_role_policy_attachment.test_s3,
    aws_opensearchserverless_access_policy.test,
  ]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  knowledge_base_configuration {
    type = "VECTOR"
    vector_knowledge_base_configuration {
      embedding_model_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}::foundation-model/%[2]s"
      embedding_model_configuration {
        bedrock_embedding_model_configuration {
          dimensions          = 1024
          embedding_data_type = "FLOAT32"
        }
      }
      supplemental_data_storage_configuration {
        storage_location {
          type = "S3"
          s3_location {
            uri = "s3://${aws_s3_bucket.test.bucket}"
          }
        }
      }
    }
  }

  storage_configuration {
    type = "OPENSEARCH_SERVERLESS"
    opensearch_serverless_configuration {
      collection_arn    = data.aws_opensearchserverless_collection.test.arn
      vector_index_name = "bedrock-knowledge-base-default-index"
      field_mapping {
        vector_field   = "bedrock-knowledge-base-default-vector"
        text_field     = "AMAZON_BEDROCK_TEXT_CHUNK"
        metadata_field = "AMAZON_BEDROCK_METADATA"
      }
    }
  }
}
`, rName, model))
}
