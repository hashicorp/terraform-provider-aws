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
func testAccDataSource_basic(t *testing.T) {
	acctest.SkipIfExeNotOnPath(t, "psql")
	acctest.SkipIfExeNotOnPath(t, "jq")
	acctest.SkipIfExeNotOnPath(t, "aws")

	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dataSource types.DataSource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_data_source.test"
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
		CheckDestroy: testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_basic(rName, foundationModel),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName, &dataSource),
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
func testAccDataSource_full(t *testing.T) {
	acctest.SkipIfExeNotOnPath(t, "psql")
	acctest.SkipIfExeNotOnPath(t, "jq")
	acctest.SkipIfExeNotOnPath(t, "aws")

	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dataSource types.DataSource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_data_source.test"
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
		CheckDestroy: testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_full(rName, foundationModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "data_deletion_policy", "RETAIN"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "data_source_configuration.0.s3_configuration.0.bucket_arn"),
					resource.TestCheckNoResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.0.bucket_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.0.inclusion_prefixes.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "data_source_configuration.0.s3_configuration.0.inclusion_prefixes.*", "Europe/France/Nouvelle-Aquitaine/Bordeaux"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.type", "S3"),
					resource.TestCheckResourceAttrSet(resourceName, "data_source_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "testing"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.chunking_strategy", "FIXED_SIZE"),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.fixed_size_chunking_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.0.chunking_configuration.0.fixed_size_chunking_configuration.0.max_tokens", acctest.Ct3),
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

// Prerequisites:
// * psql run via null_resource/provisioner "local-exec"
// * jq for parsing output from aws cli to retrieve postgres password
func testAccDataSource_disappears(t *testing.T) {
	acctest.SkipIfExeNotOnPath(t, "psql")
	acctest.SkipIfExeNotOnPath(t, "jq")
	acctest.SkipIfExeNotOnPath(t, "aws")

	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dataSource types.DataSource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_data_source.test"
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
		CheckDestroy: testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_basic(rName, foundationModel),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName, &dataSource),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbedrockagent.ResourceDataSource, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Prerequisites:
// * psql run via null_resource/provisioner "local-exec"
// * jq for parsing output from aws cli to retrieve postgres password
func testAccDataSource_update(t *testing.T) {
	acctest.SkipIfExeNotOnPath(t, "psql")
	acctest.SkipIfExeNotOnPath(t, "jq")
	acctest.SkipIfExeNotOnPath(t, "aws")

	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dataSource types.DataSource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_data_source.test"
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
		CheckDestroy: testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_basic(rName, foundationModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName, &dataSource),
					resource.TestCheckResourceAttrSet(resourceName, "data_deletion_policy"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "data_source_configuration.0.s3_configuration.0.bucket_arn"),
					resource.TestCheckNoResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.0.bucket_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.0.inclusion_prefixes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.type", "S3"),
					resource.TestCheckResourceAttrSet(resourceName, "data_source_id"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.#", acctest.Ct0),
				),
			},
			{
				Config: testAccDataSourceConfig_updated(rName, foundationModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "data_deletion_policy", "RETAIN"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "data_source_configuration.0.s3_configuration.0.bucket_arn"),
					resource.TestCheckNoResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.0.bucket_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.s3_configuration.0.inclusion_prefixes.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "data_source_configuration.0.s3_configuration.0.inclusion_prefixes.*", "Europe/France/Nouvelle-Aquitaine/Bordeaux"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.type", "S3"),
					resource.TestCheckResourceAttrSet(resourceName, "data_source_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "testing"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "vector_ingestion_configuration.#", acctest.Ct0),
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

func testAccCheckDataSourceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagent_data_source" {
				continue
			}

			_, err := tfbedrockagent.FindDataSourceByTwoPartKey(ctx, conn, rs.Primary.Attributes["data_source_id"], rs.Primary.Attributes["knowledge_base_id"])

			if tfresource.NotFound(err) {
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

func testAccCheckDataSourceExists(ctx context.Context, n string, v *types.DataSource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		output, err := tfbedrockagent.FindDataSourceByTwoPartKey(ctx, conn, rs.Primary.Attributes["data_source_id"], rs.Primary.Attributes["knowledge_base_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccDataSourceConfig_base(rName, embeddingModel string) string {
	return acctest.ConfigCompose(testAccKnowledgeBaseConfig_basicRDS(rName, embeddingModel), fmt.Sprintf(`
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
