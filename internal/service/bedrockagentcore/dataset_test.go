// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreDataset_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := randomWithPrefixAndUnderscore(t)
	resourceName := "aws_bedrockagentcore_dataset.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckDatasets(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatasetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasetConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatasetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "dataset_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "dataset_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "schema_type", "AGENTCORE_EVALUATION_PREDEFINED_V1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "dataset_version", "DRAFT"),
					resource.TestCheckResourceAttr(resourceName, "example_count", "1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "dataset_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "dataset_id",
				// source is a required write-only input (absent from GetDataset); it can
				// not round-trip on import.
				ImportStateVerifyIgnore: []string{"source"},
			},
		},
	})
}

func TestAccBedrockAgentCoreDataset_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := randomWithPrefixAndUnderscore(t)
	resourceName := "aws_bedrockagentcore_dataset.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckDatasets(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatasetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasetConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatasetExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourceDataset, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBedrockAgentCoreDataset_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := randomWithPrefixAndUnderscore(t)
	resourceName := "aws_bedrockagentcore_dataset.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckDatasets(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatasetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasetConfig_description(rName, "initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatasetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "initial description"),
				),
			},
			{
				Config: testAccDatasetConfig_description(rName, "updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatasetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "updated description"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				// Removing description does not clear it: UpdateDataset treats a nil
				// description as "leave unchanged", so the prior value is retained
				// (Optional+Computed) rather than producing an inconsistent result.
				Config: testAccDatasetConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatasetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "updated description"),
				),
			},
		},
	})
}

func TestAccBedrockAgentCoreDataset_s3Source(t *testing.T) {
	ctx := acctest.Context(t)
	rName := randomWithPrefixAndUnderscore(t)
	resourceName := "aws_bedrockagentcore_dataset.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckDatasets(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatasetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasetConfig_s3(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatasetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "schema_type", "AGENTCORE_EVALUATION_PREDEFINED_V1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.s3_source.#", "1"),
				),
			},
		},
	})
}

func testAccCheckDatasetDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_dataset" {
				continue
			}

			_, err := tfbedrockagentcore.FindDatasetByID(ctx, conn, rs.Primary.Attributes["dataset_id"])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock AgentCore Dataset %s still exists", rs.Primary.Attributes["dataset_id"])
		}

		return nil
	}
}

func testAccCheckDatasetExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		_, err := tfbedrockagentcore.FindDatasetByID(ctx, conn, rs.Primary.Attributes["dataset_id"])
		return err
	}
}

func testAccPreCheckDatasets(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

	input := bedrockagentcorecontrol.ListDatasetsInput{}

	_, err := conn.ListDatasets(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccDatasetConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_dataset" "test" {
  name        = %[1]q
  schema_type = "AGENTCORE_EVALUATION_PREDEFINED_V1"

  source {
    inline_examples {
      examples = [
        jsonencode({
          scenario_id = "scenario-1"
          turns = [
            { input = "What is 2+2?", expected_response = "4" }
          ]
        })
      ]
    }
  }
}
`, rName)
}

func testAccDatasetConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_dataset" "test" {
  name        = %[1]q
  description = %[2]q
  schema_type = "AGENTCORE_EVALUATION_PREDEFINED_V1"

  source {
    inline_examples {
      examples = [
        jsonencode({
          scenario_id = "scenario-1"
          turns = [
            { input = "What is 2+2?", expected_response = "4" }
          ]
        })
      ]
    }
  }
}
`, rName, description)
}

func testAccDatasetConfig_s3(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket_prefix = "tf-acc-test-ds-"
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.id
  key     = "examples.jsonl"
  content = jsonencode({ scenario_id = "s1", turns = [{ input = "hi", expected_response = "ok" }] })
}

resource "aws_bedrockagentcore_dataset" "test" {
  name        = %[1]q
  schema_type = "AGENTCORE_EVALUATION_PREDEFINED_V1"

  source {
    s3_source {
      s3_uri = "s3://${aws_s3_bucket.test.id}/${aws_s3_object.test.key}"
    }
  }
}
`, rName)
}
