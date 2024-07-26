// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3control "github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3ControlStorageLensConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_storage_lens_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStorageLensConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStorageLensConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStorageLensConfigurationExists(ctx, resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrAccountID),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "s3", fmt.Sprintf("storage-lens/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.activity_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.advanced_cost_optimization_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.advanced_data_protection_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.activity_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.advanced_cost_optimization_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.advanced_data_protection_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.detailed_status_code_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.prefix_level.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.detailed_status_code_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.aws_org.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.exclude.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.include.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_storage_lens_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStorageLensConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStorageLensConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageLensConfigurationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3control.ResourceStorageLensConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3ControlStorageLensConfiguration_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_storage_lens_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStorageLensConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStorageLensConfigurationConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageLensConfigurationExists(ctx, resourceName),
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
				Config: testAccStorageLensConfigurationConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageLensConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccStorageLensConfigurationConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageLensConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccS3ControlStorageLensConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_storage_lens_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStorageLensConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStorageLensConfigurationConfig_allAttributes(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStorageLensConfigurationExists(ctx, resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrAccountID),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "s3", fmt.Sprintf("storage-lens/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.activity_metrics.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.activity_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.advanced_cost_optimization_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.advanced_data_protection_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.activity_metrics.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.activity_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.advanced_cost_optimization_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.advanced_data_protection_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.detailed_status_code_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.prefix_level.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.prefix_level.0.storage_metrics.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.prefix_level.0.storage_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.prefix_level.0.storage_metrics.0.selection_criteria.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.prefix_level.0.storage_metrics.0.selection_criteria.0.delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.prefix_level.0.storage_metrics.0.selection_criteria.0.max_depth", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.prefix_level.0.storage_metrics.0.selection_criteria.0.min_storage_bytes_percentage", "49.5"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.detailed_status_code_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.aws_org.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.cloud_watch_metrics.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.cloud_watch_metrics.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.#", acctest.Ct1),
					acctest.CheckResourceAttrAccountID(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.account_id"),
					resource.TestCheckResourceAttrSet(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.arn"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.encryption.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.format", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.output_schema_version", "V_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.prefix", "p1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.exclude.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.exclude.0.buckets.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.exclude.0.regions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.include.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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
					testAccCheckStorageLensConfigurationExists(ctx, resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrAccountID),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "s3", fmt.Sprintf("storage-lens/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.activity_metrics.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.activity_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.advanced_cost_optimization_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.advanced_data_protection_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.activity_metrics.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.activity_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.advanced_cost_optimization_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.advanced_data_protection_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.detailed_status_code_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.prefix_level.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.detailed_status_code_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.aws_org.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.cloud_watch_metrics.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.cloud_watch_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.#", acctest.Ct1),
					acctest.CheckResourceAttrAccountID(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.account_id"),
					resource.TestCheckResourceAttrSet(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.arn"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.encryption.0.sse_kms.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.encryption.0.sse_s3.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.format", "Parquet"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.output_schema_version", "V_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.exclude.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.include.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.include.0.buckets.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.include.0.regions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccS3ControlStorageLensConfiguration_advancedMetrics(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_storage_lens_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStorageLensConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStorageLensConfigurationConfig_advancedMetrics(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStorageLensConfigurationExists(ctx, resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrAccountID),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "s3", fmt.Sprintf("storage-lens/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.activity_metrics.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.activity_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.advanced_cost_optimization_metrics.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.advanced_cost_optimization_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.advanced_data_protection_metrics.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.advanced_data_protection_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.activity_metrics.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.activity_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.advanced_cost_optimization_metrics.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.advanced_cost_optimization_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.advanced_data_protection_metrics.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.advanced_data_protection_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.detailed_status_code_metrics.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.detailed_status_code_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.prefix_level.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.detailed_status_code_metrics.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.detailed_status_code_metrics.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.aws_org.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.cloud_watch_metrics.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.cloud_watch_metrics.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.exclude.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.include.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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
					testAccCheckStorageLensConfigurationExists(ctx, resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrAccountID),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "s3", fmt.Sprintf("storage-lens/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.activity_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.advanced_cost_optimization_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.advanced_data_protection_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.activity_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.advanced_cost_optimization_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.advanced_data_protection_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.detailed_status_code_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.prefix_level.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.detailed_status_code_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.aws_org.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.cloud_watch_metrics.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.cloud_watch_metrics.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.data_export.0.s3_bucket_destination.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.exclude.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.include.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func testAccCheckStorageLensConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3control_object_lambda_access_point" {
				continue
			}

			accountID, configID, err := tfs3control.StorageLensConfigurationParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfs3control.FindStorageLensConfigurationByAccountIDAndConfigID(ctx, conn, accountID, configID)

			if tfresource.NotFound(err) {
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

func testAccCheckStorageLensConfigurationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		accountID, configID, err := tfs3control.StorageLensConfigurationParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

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
