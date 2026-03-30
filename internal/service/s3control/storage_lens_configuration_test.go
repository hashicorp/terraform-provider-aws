// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfs3control "github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3ControlStorageLensConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3control_storage_lens_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStorageLensConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStorageLensConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStorageLensConfigurationExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "s3", fmt.Sprintf("storage-lens/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.activity_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.advanced_cost_optimization_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.advanced_data_protection_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.activity_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.advanced_cost_optimization_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.advanced_data_protection_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.detailed_status_code_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.prefix_level.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.detailed_status_code_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.aws_org.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.exclude.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.include.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func TestAccS3ControlStorageLensConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3control_storage_lens_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStorageLensConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStorageLensConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageLensConfigurationExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfs3control.ResourceStorageLensConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3ControlStorageLensConfiguration_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3control_storage_lens_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStorageLensConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStorageLensConfigurationConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageLensConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStorageLensConfigurationConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageLensConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccStorageLensConfigurationConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageLensConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccS3ControlStorageLensConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3control_storage_lens_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStorageLensConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStorageLensConfigurationConfig_allAttributes(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStorageLensConfigurationExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "s3", fmt.Sprintf("storage-lens/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.activity_metrics.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.activity_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.advanced_cost_optimization_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.advanced_data_protection_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.activity_metrics.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.activity_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.advanced_cost_optimization_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.advanced_data_protection_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.detailed_status_code_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.prefix_level.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.prefix_level.0.storage_metrics.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.prefix_level.0.storage_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.prefix_level.0.storage_metrics.0.selection_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.prefix_level.0.storage_metrics.0.selection_criteria.0.delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.prefix_level.0.storage_metrics.0.selection_criteria.0.max_depth", "3"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.prefix_level.0.storage_metrics.0.selection_criteria.0.min_storage_bytes_percentage", "49.5"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.detailed_status_code_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.aws_org.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.cloud_watch_metrics.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.cloud_watch_metrics.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.#", "1"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.account_id"),
					resource.TestCheckResourceAttrSet(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.arn"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.encryption.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.format", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.output_schema_version", "V_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.prefix", "p1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.exclude.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.exclude.0.buckets.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.exclude.0.regions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.include.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStorageLensConfigurationConfig_allAttributesUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStorageLensConfigurationExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "s3", fmt.Sprintf("storage-lens/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.activity_metrics.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.activity_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.advanced_cost_optimization_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.advanced_data_protection_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.activity_metrics.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.activity_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.advanced_cost_optimization_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.advanced_data_protection_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.detailed_status_code_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.prefix_level.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.detailed_status_code_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.aws_org.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.cloud_watch_metrics.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.cloud_watch_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.#", "1"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.account_id"),
					resource.TestCheckResourceAttrSet(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.arn"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.encryption.0.sse_kms.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.encryption.0.sse_s3.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.format", "Parquet"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.output_schema_version", "V_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.exclude.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.include.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.include.0.buckets.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.include.0.regions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func TestAccS3ControlStorageLensConfiguration_advancedMetrics(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3control_storage_lens_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStorageLensConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStorageLensConfigurationConfig_advancedMetrics(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStorageLensConfigurationExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "s3", fmt.Sprintf("storage-lens/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.activity_metrics.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.activity_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.advanced_cost_optimization_metrics.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.advanced_cost_optimization_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.advanced_data_protection_metrics.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.advanced_data_protection_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.activity_metrics.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.activity_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.advanced_cost_optimization_metrics.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.advanced_cost_optimization_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.advanced_data_protection_metrics.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.advanced_data_protection_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.detailed_status_code_metrics.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.detailed_status_code_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.prefix_level.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.detailed_status_code_metrics.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.detailed_status_code_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.aws_org.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.cloud_watch_metrics.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.cloud_watch_metrics.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.exclude.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.include.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStorageLensConfigurationConfig_freeMetrics(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStorageLensConfigurationExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "s3", fmt.Sprintf("storage-lens/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.activity_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.advanced_cost_optimization_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.advanced_data_protection_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.activity_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.advanced_cost_optimization_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.advanced_data_protection_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.detailed_status_code_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.prefix_level.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.detailed_status_code_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.aws_org.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.cloud_watch_metrics.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.cloud_watch_metrics.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.exclude.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.include.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func testAccCheckStorageLensConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).S3ControlClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3control_object_lambda_access_point" {
				continue
			}

			accountID, configID, err := tfs3control.StorageLensConfigurationParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfs3control.FindStorageLensConfigurationByAccountIDAndConfigID(ctx, conn, accountID, configID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Storage Lens Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckStorageLensConfigurationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		accountID, configID, err := tfs3control.StorageLensConfigurationParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.ProviderMeta(ctx, t).S3ControlClient(ctx)

		_, err = tfs3control.FindStorageLensConfigurationByAccountIDAndConfigID(ctx, conn, accountID, configID)

		return err
	}
}

func testAccStorageLensConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3control_storage_lens_configuration" "test" {
  config_id = %[1]q

  storage_lens_configuration {
    enabled = true

    account_level {
      bucket_level {}
    }
  }
}
`, rName)
}

func testAccStorageLensConfigurationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_s3control_storage_lens_configuration" "test" {
  config_id = %[1]q

  storage_lens_configuration {
    enabled = true

    account_level {
      bucket_level {}
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccStorageLensConfigurationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_s3control_storage_lens_configuration" "test" {
  config_id = %[1]q

  storage_lens_configuration {
    enabled = true

    account_level {
      bucket_level {}
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

// lintignore:AWSAT003
func testAccStorageLensConfigurationConfig_allAttributes(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket" "test" {
  count = 3
}

resource "aws_s3control_storage_lens_configuration" "test" {
  config_id = %[1]q

  storage_lens_configuration {
    enabled = true

    account_level {
      activity_metrics {
        enabled = true
      }

      bucket_level {
        activity_metrics {
          enabled = true
        }

        prefix_level {
          storage_metrics {
            enabled = true

            selection_criteria {
              delimiter                    = ","
              max_depth                    = 3
              min_storage_bytes_percentage = 49.5
            }
          }
        }
      }
    }

    data_export {
      cloud_watch_metrics {
        enabled = false
      }

      s3_bucket_destination {
        account_id            = data.aws_caller_identity.current.account_id
        arn                   = aws_s3_bucket.test[0].arn
        format                = "CSV"
        output_schema_version = "V_1"
        prefix                = "p1"
      }
    }

    exclude {
      buckets = [aws_s3_bucket.test[1].arn, aws_s3_bucket.test[2].arn]
      regions = ["eu-south-1"]
    }
  }
}
`, rName)
}

func testAccStorageLensConfigurationConfig_allAttributesUpdated(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket" "test" {
  count = 3
}

resource "aws_s3control_storage_lens_configuration" "test" {
  config_id = %[1]q

  storage_lens_configuration {
    enabled = true

    account_level {
      activity_metrics {
        enabled = true
      }

      bucket_level {
        activity_metrics {
          enabled = true
        }
      }
    }

    data_export {
      cloud_watch_metrics {
        enabled = true
      }

      s3_bucket_destination {
        account_id            = data.aws_caller_identity.current.account_id
        arn                   = aws_s3_bucket.test[0].arn
        format                = "Parquet"
        output_schema_version = "V_1"

        encryption {
          sse_s3 {}
        }
      }
    }

    include {
      buckets = [aws_s3_bucket.test[1].arn]
    }
  }
}
`, rName)
}

func testAccStorageLensConfigurationConfig_advancedMetrics(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3control_storage_lens_configuration" "test" {
  config_id = %[1]q

  storage_lens_configuration {
    enabled = true

    account_level {
      activity_metrics {
        enabled = true
      }

      advanced_cost_optimization_metrics {
        enabled = true
      }

      advanced_data_protection_metrics {
        enabled = true
      }

      detailed_status_code_metrics {
        enabled = true
      }

      bucket_level {
        activity_metrics {
          enabled = true
        }

        advanced_cost_optimization_metrics {
          enabled = true
        }

        advanced_data_protection_metrics {
          enabled = true
        }

        detailed_status_code_metrics {
          enabled = true
        }
      }
    }

    data_export {
      cloud_watch_metrics {
        enabled = false
      }
    }
  }
}
`, rName)
}

func testAccStorageLensConfigurationConfig_freeMetrics(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3control_storage_lens_configuration" "test" {
  config_id = %[1]q

  storage_lens_configuration {
    enabled = true

    account_level {
      bucket_level {}
    }

    data_export {
      cloud_watch_metrics {
        enabled = false
      }
    }
  }
}
`, rName)
}
