// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagent_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagent "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagent"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSource types.DataSource
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_data_source.test"
	foundationModel := "amazon.titan-embed-text-v2:0"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_basic(rName, foundationModel),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, t, resourceName, &dataSource),
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

func testAccDataSource_full(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSource types.DataSource
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_data_source.test"
	foundationModel := "amazon.titan-embed-text-v2:0"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_full(rName, foundationModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataSourceExists(ctx, t, resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "data_deletion_policy", "RETAIN"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "data_source_configuration.0.s3_configuration.0.bucket_arn"),
					resource.TestCheckNoResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.0.bucket_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.0.inclusion_prefixes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "data_source_configuration.0.s3_configuration.0.inclusion_prefixes.*", "Europe/France/Nouvelle-Aquitaine/Bordeaux"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.type", "S3"),
					resource.TestCheckResourceAttrSet(resourceName, "data_source_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "testing"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.chunking_strategy", "FIXED_SIZE"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.fixed_size_chunking_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.fixed_size_chunking_configuration.0.max_tokens", "3"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.fixed_size_chunking_configuration.0.overlap_percentage", "80"),
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

func testAccDataSource_fullSemantic(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSource types.DataSource
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_data_source.test"
	foundationModel := "amazon.titan-embed-text-v2:0"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_fullSemantic(rName, foundationModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataSourceExists(ctx, t, resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "data_deletion_policy", "RETAIN"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "data_source_configuration.0.s3_configuration.0.bucket_arn"),
					resource.TestCheckNoResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.0.bucket_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.0.inclusion_prefixes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "data_source_configuration.0.s3_configuration.0.inclusion_prefixes.*", "Europe/France/Nouvelle-Aquitaine/Bordeaux"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.type", "S3"),
					resource.TestCheckResourceAttrSet(resourceName, "data_source_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "testing"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.chunking_strategy", "SEMANTIC"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.semantic_chunking_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.semantic_chunking_configuration.0.breakpoint_percentile_threshold", "80"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.semantic_chunking_configuration.0.buffer_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.semantic_chunking_configuration.0.max_token", "10"),
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

func testAccDataSource_fullHierarchical(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSource types.DataSource
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_data_source.test"
	foundationModel := "amazon.titan-embed-text-v2:0"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_fullHierarchical(rName, foundationModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataSourceExists(ctx, t, resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "data_deletion_policy", "RETAIN"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "data_source_configuration.0.s3_configuration.0.bucket_arn"),
					resource.TestCheckNoResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.0.bucket_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.0.inclusion_prefixes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "data_source_configuration.0.s3_configuration.0.inclusion_prefixes.*", "Europe/France/Nouvelle-Aquitaine/Bordeaux"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.type", "S3"),
					resource.TestCheckResourceAttrSet(resourceName, "data_source_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "testing"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.chunking_strategy", "HIERARCHICAL"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.hierarchical_chunking_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.hierarchical_chunking_configuration.0.overlap_tokens", "2"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.hierarchical_chunking_configuration.0.level_configuration.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.hierarchical_chunking_configuration.0.level_configuration.0.max_tokens", "15"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.hierarchical_chunking_configuration.0.level_configuration.1.max_tokens", "10"),
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

func testAccDataSource_fullCustomTranformation(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSource types.DataSource
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_data_source.test"
	foundationModel := "amazon.titan-embed-text-v2:0"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_fullCustomTransformation(rName, foundationModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataSourceExists(ctx, t, resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "data_deletion_policy", "RETAIN"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "data_source_configuration.0.s3_configuration.0.bucket_arn"),
					resource.TestCheckNoResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.0.bucket_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.0.inclusion_prefixes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "data_source_configuration.0.s3_configuration.0.inclusion_prefixes.*", "Europe/France/Nouvelle-Aquitaine/Bordeaux"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.type", "S3"),
					resource.TestCheckResourceAttrSet(resourceName, "data_source_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "testing"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.chunking_strategy", "FIXED_SIZE"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.fixed_size_chunking_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.fixed_size_chunking_configuration.0.max_tokens", "3"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.fixed_size_chunking_configuration.0.overlap_percentage", "80"),
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

func testAccDataSource_parsing(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSource types.DataSource
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_data_source.test"
	foundationModel := "amazon.titan-embed-text-v2:0"
	parsingModel := "anthropic.claude-3-sonnet-20240229-v1:0"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_parsing(rName, foundationModel, parsingModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataSourceExists(ctx, t, resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "data_deletion_policy", "RETAIN"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "data_source_configuration.0.s3_configuration.0.bucket_arn"),
					resource.TestCheckNoResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.0.bucket_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.0.inclusion_prefixes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "data_source_configuration.0.s3_configuration.0.inclusion_prefixes.*", "Europe/France/Nouvelle-Aquitaine/Bordeaux"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.type", "S3"),
					resource.TestCheckResourceAttrSet(resourceName, "data_source_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "testing"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.chunking_strategy", "FIXED_SIZE"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.fixed_size_chunking_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.fixed_size_chunking_configuration.0.max_tokens", "3"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.fixed_size_chunking_configuration.0.overlap_percentage", "80"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.parsing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.parsing_configuration.0.parsing_strategy", "BEDROCK_FOUNDATION_MODEL"),
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

func testAccDataSource_parsingModality(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSource types.DataSource
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_data_source.test"
	foundationModel := "amazon.titan-embed-text-v2:0"
	parsingModel := "anthropic.claude-3-sonnet-20240229-v1:0"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_parsingModality(rName, foundationModel, parsingModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataSourceExists(ctx, t, resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "data_deletion_policy", "RETAIN"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "data_source_configuration.0.s3_configuration.0.bucket_arn"),
					resource.TestCheckNoResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.0.bucket_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.0.inclusion_prefixes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "data_source_configuration.0.s3_configuration.0.inclusion_prefixes.*", "Europe/France/Nouvelle-Aquitaine/Bordeaux"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.type", "S3"),
					resource.TestCheckResourceAttrSet(resourceName, "data_source_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "testing"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.chunking_strategy", "FIXED_SIZE"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.fixed_size_chunking_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.fixed_size_chunking_configuration.0.max_tokens", "3"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.fixed_size_chunking_configuration.0.overlap_percentage", "80"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.parsing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.parsing_configuration.0.parsing_strategy", "BEDROCK_FOUNDATION_MODEL"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.parsing_configuration.0.bedrock_foundation_model_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.parsing_configuration.0.bedrock_foundation_model_configuration.0.parsing_modality", "MULTIMODAL"),
					resource.TestCheckResourceAttrSet(resourceName, "vector_ingestion_configuration.0.parsing_configuration.0.bedrock_foundation_model_configuration.0.model_arn"),
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
	ctx := acctest.Context(t)
	var dataSource types.DataSource
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_data_source.test"
	foundationModel := "amazon.titan-embed-text-v2:0"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_basic(rName, foundationModel),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, t, resourceName, &dataSource),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagent.ResourceDataSource, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDataSource_update(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSource types.DataSource
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_data_source.test"
	foundationModel := "amazon.titan-embed-text-v2:0"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_basic(rName, foundationModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataSourceExists(ctx, t, resourceName, &dataSource),
					resource.TestCheckResourceAttrSet(resourceName, "data_deletion_policy"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "data_source_configuration.0.s3_configuration.0.bucket_arn"),
					resource.TestCheckNoResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.0.bucket_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.0.inclusion_prefixes.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.type", "S3"),
					resource.TestCheckResourceAttrSet(resourceName, "data_source_id"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.#", "0"),
				),
			},
			{
				Config: testAccDataSourceConfig_updated(rName, foundationModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataSourceExists(ctx, t, resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "data_deletion_policy", "RETAIN"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "data_source_configuration.0.s3_configuration.0.bucket_arn"),
					resource.TestCheckNoResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.0.bucket_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.0.inclusion_prefixes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "data_source_configuration.0.s3_configuration.0.inclusion_prefixes.*", "Europe/France/Nouvelle-Aquitaine/Bordeaux"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.type", "S3"),
					resource.TestCheckResourceAttrSet(resourceName, "data_source_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "testing"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.#", "0"),
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

func testAccDataSource_webConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	collectionName := skipIfOSSCollectionNameEnvVarNotSet(t)
	var dataSource types.DataSource
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_data_source.test"
	foundationModel := "amazon.titan-embed-text-v2:0"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_webConfiguration(rName, collectionName, foundationModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataSourceExists(ctx, t, resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.web_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.web_configuration.0.source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.web_configuration.0.crawler_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.web_configuration.0.crawler_configuration.0.crawler_limits.0.max_pages", "25000"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.web_configuration.0.crawler_configuration.0.crawler_limits.0.rate_limit", "300"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.web_configuration.0.crawler_configuration.0.exclusion_filters.#", "5"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.web_configuration.0.crawler_configuration.0.inclusion_filters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.web_configuration.0.crawler_configuration.0.user_agent", "bedrockbot_UUID test"),
				),
			},
		},
	})
}

func testAccDataSource_bedrockDataAutomation(t *testing.T) {
	acctest.SkipIfExeNotOnPath(t, "psql")
	acctest.SkipIfExeNotOnPath(t, "jq")
	acctest.SkipIfExeNotOnPath(t, "aws")

	ctx := acctest.Context(t)
	var dataSource types.DataSource
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_data_source.test"
	foundationModel := "amazon.titan-embed-text-v1"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"null": {
				Source:            "hashicorp/null",
				VersionConstraint: "3.2.2",
			},
		},
		CheckDestroy: testAccCheckDataSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_bedrockDataAutomation(rName, foundationModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataSourceExists(ctx, t, resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.parsing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.parsing_configuration.0.parsing_strategy", "BEDROCK_DATA_AUTOMATION"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.parsing_configuration.0.bedrock_data_automation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.parsing_configuration.0.bedrock_data_automation_configuration.0.parsing_modality", "MULTIMODAL"),
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

func testAccCheckDataSourceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagent_data_source" {
				continue
			}

			_, err := tfbedrockagent.FindDataSourceByTwoPartKey(ctx, conn, rs.Primary.Attributes["data_source_id"], rs.Primary.Attributes["knowledge_base_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Data Source %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDataSourceExists(ctx context.Context, t *testing.T, n string, v *types.DataSource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentClient(ctx)

		output, err := tfbedrockagent.FindDataSourceByTwoPartKey(ctx, conn, rs.Primary.Attributes["data_source_id"], rs.Primary.Attributes["knowledge_base_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccDataSourceConfig_base(rName, embeddingModel string) string {
	return acctest.ConfigCompose(testAccKnowledgeBaseConfig_S3VectorsByIndexARN(rName, embeddingModel), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, rName))
}

func testAccDataSourceConfig_basic(rName, embeddingModel string) string {
	return acctest.ConfigCompose(testAccDataSourceConfig_base(rName, embeddingModel), fmt.Sprintf(`
resource "aws_bedrockagent_data_source" "test" {
  name              = %[1]q
  knowledge_base_id = aws_bedrockagent_knowledge_base.test.id

  data_source_configuration {
    type = "S3"

    s3_configuration {
      bucket_arn = aws_s3_bucket.test.arn
    }
  }
}
`, rName))
}

func testAccDataSourceConfig_full(rName, embeddingModel string) string {
	return acctest.ConfigCompose(testAccDataSourceConfig_base(rName, embeddingModel), fmt.Sprintf(`
resource "aws_bedrockagent_data_source" "test" {
  name                 = %[1]q
  knowledge_base_id    = aws_bedrockagent_knowledge_base.test.id
  data_deletion_policy = "RETAIN"
  description          = "testing"

  data_source_configuration {
    type = "S3"

    s3_configuration {
      bucket_arn         = aws_s3_bucket.test.arn
      inclusion_prefixes = ["Europe/France/Nouvelle-Aquitaine/Bordeaux"]
    }
  }

  vector_ingestion_configuration {
    chunking_configuration {
      chunking_strategy = "FIXED_SIZE"

      fixed_size_chunking_configuration {
        max_tokens         = 3
        overlap_percentage = 80
      }
    }
  }
}
`, rName))
}

func testAccDataSourceConfig_parsing(rName, embeddingModel, parsingModel string) string {
	return acctest.ConfigCompose(testAccDataSourceConfig_base(rName, embeddingModel), fmt.Sprintf(`
resource "aws_bedrockagent_data_source" "test" {
  name                 = %[1]q
  knowledge_base_id    = aws_bedrockagent_knowledge_base.test.id
  data_deletion_policy = "RETAIN"
  description          = "testing"

  data_source_configuration {
    type = "S3"

    s3_configuration {
      bucket_arn         = aws_s3_bucket.test.arn
      inclusion_prefixes = ["Europe/France/Nouvelle-Aquitaine/Bordeaux"]
    }
  }

  vector_ingestion_configuration {
    chunking_configuration {
      chunking_strategy = "FIXED_SIZE"

      fixed_size_chunking_configuration {
        max_tokens         = 3
        overlap_percentage = 80
      }
    }

    parsing_configuration {
      parsing_strategy = "BEDROCK_FOUNDATION_MODEL"
      bedrock_foundation_model_configuration {
        model_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}::foundation-model/%[2]s"

        parsing_prompt {
          parsing_prompt_string = "Transcribe the text content from an image page and output in Markdown syntax (not code blocks)."
        }
      }
    }
  }
}
`, rName, parsingModel))
}

func testAccDataSourceConfig_parsingModality(rName, embeddingModel, parsingModel string) string {
	return acctest.ConfigCompose(testAccDataSourceConfig_base(rName, embeddingModel), fmt.Sprintf(`
resource "aws_bedrockagent_data_source" "test" {
  name                 = %[1]q
  knowledge_base_id    = aws_bedrockagent_knowledge_base.test.id
  data_deletion_policy = "RETAIN"
  description          = "testing"

  data_source_configuration {
    type = "S3"

    s3_configuration {
      bucket_arn         = aws_s3_bucket.test.arn
      inclusion_prefixes = ["Europe/France/Nouvelle-Aquitaine/Bordeaux"]
    }
  }

  vector_ingestion_configuration {
    chunking_configuration {
      chunking_strategy = "FIXED_SIZE"

      fixed_size_chunking_configuration {
        max_tokens         = 3
        overlap_percentage = 80
      }
    }

    parsing_configuration {
      parsing_strategy = "BEDROCK_FOUNDATION_MODEL"
      bedrock_foundation_model_configuration {
        model_arn        = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}::foundation-model/%[2]s"
        parsing_modality = "MULTIMODAL"

        parsing_prompt {
          parsing_prompt_string = "Transcribe the text content from an image page and output in Markdown syntax (not code blocks)."
        }
      }
    }
  }
}
`, rName, parsingModel))
}

func testAccDataSourceConfig_fullSemantic(rName, embeddingModel string) string {
	return acctest.ConfigCompose(testAccDataSourceConfig_base(rName, embeddingModel), fmt.Sprintf(`
resource "aws_bedrockagent_data_source" "test" {
  name                 = %[1]q
  knowledge_base_id    = aws_bedrockagent_knowledge_base.test.id
  data_deletion_policy = "RETAIN"
  description          = "testing"

  data_source_configuration {
    type = "S3"

    s3_configuration {
      bucket_arn         = aws_s3_bucket.test.arn
      inclusion_prefixes = ["Europe/France/Nouvelle-Aquitaine/Bordeaux"]
    }
  }

  vector_ingestion_configuration {
    chunking_configuration {
      chunking_strategy = "SEMANTIC"

      semantic_chunking_configuration {
        breakpoint_percentile_threshold = 80
        buffer_size                     = 1
        max_token                       = 10
      }
    }
  }
}
`, rName))
}

func testAccDataSourceConfig_fullHierarchical(rName, embeddingModel string) string {
	return acctest.ConfigCompose(testAccDataSourceConfig_base(rName, embeddingModel), fmt.Sprintf(`
resource "aws_bedrockagent_data_source" "test" {
  name                 = %[1]q
  knowledge_base_id    = aws_bedrockagent_knowledge_base.test.id
  data_deletion_policy = "RETAIN"
  description          = "testing"

  data_source_configuration {
    type = "S3"

    s3_configuration {
      bucket_arn         = aws_s3_bucket.test.arn
      inclusion_prefixes = ["Europe/France/Nouvelle-Aquitaine/Bordeaux"]
    }
  }

  vector_ingestion_configuration {
    chunking_configuration {
      chunking_strategy = "HIERARCHICAL"

      hierarchical_chunking_configuration {
        overlap_tokens = 2
        level_configuration {
          max_tokens = 15
        }
        level_configuration {
          max_tokens = 10
        }
      }
    }
  }
}
`, rName))
}

func testAccDataSourceConfig_fullCustomTransformation(rName, embeddingModel string) string {
	return acctest.ConfigCompose(testAccDataSourceConfig_base(rName, embeddingModel),
		testAccAgentActionGroupConfig_lambda(rName), fmt.Sprintf(`
resource "aws_bedrockagent_data_source" "test" {
  name                 = %[1]q
  knowledge_base_id    = aws_bedrockagent_knowledge_base.test.id
  data_deletion_policy = "RETAIN"
  description          = "testing"

  data_source_configuration {
    type = "S3"

    s3_configuration {
      bucket_arn         = aws_s3_bucket.test.arn
      inclusion_prefixes = ["Europe/France/Nouvelle-Aquitaine/Bordeaux"]
    }
  }

  vector_ingestion_configuration {
    chunking_configuration {
      chunking_strategy = "FIXED_SIZE"

      fixed_size_chunking_configuration {
        max_tokens         = 3
        overlap_percentage = 80
      }
    }
    custom_transformation_configuration {
      intermediate_storage {
        s3_location {
          uri = "s3://${aws_s3_bucket.test_im.bucket}/customTransform"
        }
      }
      transformation {
        step_to_apply = "POST_CHUNKING"
        transformation_function {
          transformation_lambda_configuration {
            lambda_arn = aws_lambda_function.test_lambda.arn
          }
        }
      }
    }
  }
}

resource "aws_s3_bucket" "test_im" {
  bucket = "%[1]s-im"
}
`, rName))
}

func testAccDataSourceConfig_updated(rName, embeddingModel string) string {
	return acctest.ConfigCompose(testAccDataSourceConfig_base(rName, embeddingModel), fmt.Sprintf(`
resource "aws_bedrockagent_data_source" "test" {
  name                 = %[1]q
  knowledge_base_id    = aws_bedrockagent_knowledge_base.test.id
  data_deletion_policy = "RETAIN"
  description          = "testing"

  data_source_configuration {
    type = "S3"

    s3_configuration {
      bucket_arn         = aws_s3_bucket.test.arn
      inclusion_prefixes = ["Europe/France/Nouvelle-Aquitaine/Bordeaux"]
    }
  }
}
`, rName))
}

func testAccDataSourceConfig_webConfiguration(rName, collectionName, embeddingModel string) string {
	return acctest.ConfigCompose(testAccKnowledgeBaseConfig_OpenSearchServerless_basic(rName, collectionName, embeddingModel), fmt.Sprintf(`
resource "aws_bedrockagent_data_source" "test" {
  name              = %[1]q
  knowledge_base_id = aws_bedrockagent_knowledge_base.test.id

  data_source_configuration {
    type = "WEB"

    web_configuration {
      source_configuration {
        url_configuration {
          seed_urls {
            url = "https://aws.amazon.com/blogs/compute/category/compute/aws-outposts/"
          }
          seed_urls {
            url = "https://aws.amazon.com/blogs/networking-and-content-delivery/category/compute/aws-outposts/"
          }
        }
      }

      crawler_configuration {
        crawler_limits {
          max_pages  = 25000
          rate_limit = 300
        }
        exclusion_filters = [
          ".*\\.(txt|csv|md|pdf|doc|docx|xls|xlsx).*",
          ".*/(users|topics|products|contact\\-us|about\\-aws|pricing|privacy)/.*",
          ".*/(terms|getting\\-started)$",
          ".*\\.(github|pages\\.awscloud|awsstatic|oracle)\\.com.*",
          ".*\\.(gov|edu).*"
        ]
        inclusion_filters = [
          ".*/blogs/(compute|containers|networking\\-and\\-content\\-delivery|storage|publicsector|media|awsmarketplace|apn|machine\\-learning|industries|mt|aws|architecture|database)/.*"
        ]
        user_agent = "bedrockbot_UUID test"
      }
    }
  }
}
`, rName))
}

func testAccDataSourceConfig_bedrockDataAutomation(rName, embeddingModel string) string {
	return acctest.ConfigCompose(testAccKnowledgeBaseConfig_RDS_supplementalDataStorage(rName, embeddingModel), fmt.Sprintf(`
resource "aws_s3_bucket" "test2" {
  bucket        = "%[1]s-2"
  force_destroy = true
}

resource "aws_bedrockagent_data_source" "test" {
  knowledge_base_id = aws_bedrockagent_knowledge_base.test.id
  name              = %[1]q

  data_source_configuration {
    type = "S3"
    s3_configuration {
      bucket_arn = aws_s3_bucket.test2.arn
    }
  }

  vector_ingestion_configuration {
    parsing_configuration {
      parsing_strategy = "BEDROCK_DATA_AUTOMATION"
      bedrock_data_automation_configuration {
        parsing_modality = "MULTIMODAL"
      }
    }
  }
}
`, rName))
}
