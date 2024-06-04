// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package evidently_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/evidently/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudwatchevidently "github.com/hashicorp/terraform-provider-aws/internal/service/evidently"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEvidentlyProject_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var project awstypes.Project

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	originalDescription := "original description"
	updatedDescription := "updated description"
	resourceName := "aws_evidently_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_basic(rName, originalDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttrSet(resourceName, "active_experiment_count"),
					resource.TestCheckResourceAttrSet(resourceName, "active_launch_count"),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "evidently", fmt.Sprintf("project/%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, originalDescription),
					resource.TestCheckResourceAttrSet(resourceName, "experiment_count"),
					resource.TestCheckResourceAttrSet(resourceName, "feature_count"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedTime),
					resource.TestCheckResourceAttrSet(resourceName, "launch_count"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.ProjectStatusAvailable)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Test Project"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_basic(rName, updatedDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "active_experiment_count"),
					resource.TestCheckResourceAttrSet(resourceName, "active_launch_count"),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "evidently", fmt.Sprintf("project/%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, updatedDescription),
					resource.TestCheckResourceAttrSet(resourceName, "experiment_count"),
					resource.TestCheckResourceAttrSet(resourceName, "feature_count"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedTime),
					resource.TestCheckResourceAttrSet(resourceName, "launch_count"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Test Project"),
				),
			},
		},
	})
}

func TestAccEvidentlyProject_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var project awstypes.Project

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	description := "example description"
	resourceName := "aws_evidently_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_basic(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Test Project"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_tags1(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Test Project"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_tags2(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Test Project"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
				),
			},
		},
	})
}

func TestAccEvidentlyProject_updateDataDeliveryCloudWatchLogGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var project awstypes.Project

	rName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	rName2 := sdkacctest.RandomWithPrefix("tf-test-bucket")
	rName3 := sdkacctest.RandomWithPrefix(fmt.Sprintf("/aws/vendedlogs/%s", acctest.ResourcePrefix))
	rName4 := sdkacctest.RandomWithPrefix(fmt.Sprintf("/aws/vendedlogs/%s", acctest.ResourcePrefix))
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_evidently_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_dataDeliveryCloudWatchLogs(rName, rName2, rName3, rName4, rName5, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "data_delivery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_delivery.0.cloudwatch_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "data_delivery.0.cloudwatch_logs.0.log_group", "aws_cloudwatch_log_group.test", names.AttrName)),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_dataDeliveryCloudWatchLogs(rName, rName2, rName3, rName4, rName5, "second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "data_delivery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_delivery.0.cloudwatch_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "data_delivery.0.cloudwatch_logs.0.log_group", "aws_cloudwatch_log_group.test2", names.AttrName)),
			},
		},
	})
}

func TestAccEvidentlyProject_updateDataDeliveryS3Bucket(t *testing.T) {
	ctx := acctest.Context(t)
	var project awstypes.Project

	rName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	rName2 := sdkacctest.RandomWithPrefix("tf-test-bucket")
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName4 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_evidently_project.test"
	prefix := "tests3prefix"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_dataDeliveryS3Bucket(rName, rName2, rName3, rName4, rName5, prefix, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "data_delivery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_delivery.0.s3_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "data_delivery.0.s3_destination.0.bucket", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "data_delivery.0.s3_destination.0.prefix", prefix),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_dataDeliveryS3Bucket(rName, rName2, rName3, rName4, rName5, prefix, "second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "data_delivery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_delivery.0.s3_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "data_delivery.0.s3_destination.0.bucket", "aws_s3_bucket.test2", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "data_delivery.0.s3_destination.0.prefix", prefix),
				),
			},
		},
	})
}

func TestAccEvidentlyProject_updateDataDeliveryS3Prefix(t *testing.T) {
	ctx := acctest.Context(t)
	var project awstypes.Project

	rName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	rName2 := sdkacctest.RandomWithPrefix("tf-test-bucket")
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName4 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_evidently_project.test"
	originalPrefix := "original-prefix"
	updatedPrefix := "updated-prefix"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_dataDeliveryS3Bucket(rName, rName2, rName3, rName4, rName5, originalPrefix, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "data_delivery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_delivery.0.s3_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "data_delivery.0.s3_destination.0.bucket", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "data_delivery.0.s3_destination.0.prefix", originalPrefix),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_dataDeliveryS3Bucket(rName, rName2, rName3, rName4, rName5, updatedPrefix, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "data_delivery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_delivery.0.s3_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "data_delivery.0.s3_destination.0.bucket", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "data_delivery.0.s3_destination.0.prefix", updatedPrefix),
				),
			},
		},
	})
}

func TestAccEvidentlyProject_updateDataDeliveryCloudWatchToS3(t *testing.T) {
	ctx := acctest.Context(t)
	var project awstypes.Project

	rName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	rName2 := sdkacctest.RandomWithPrefix("tf-test-bucket")
	rName3 := sdkacctest.RandomWithPrefix(fmt.Sprintf("/aws/vendedlogs/%s", acctest.ResourcePrefix))
	rName4 := sdkacctest.RandomWithPrefix(fmt.Sprintf("/aws/vendedlogs/%s", acctest.ResourcePrefix))
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_evidently_project.test"
	prefix := "tests3prefix"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_dataDeliveryCloudWatchLogs(rName, rName2, rName3, rName4, rName5, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "data_delivery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_delivery.0.cloudwatch_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "data_delivery.0.cloudwatch_logs.0.log_group", "aws_cloudwatch_log_group.test", names.AttrName)),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_dataDeliveryS3Bucket(rName, rName2, rName3, rName4, rName5, prefix, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "data_delivery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_delivery.0.s3_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "data_delivery.0.s3_destination.0.bucket", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "data_delivery.0.s3_destination.0.prefix", prefix),
				),
			},
		},
	})
}

func TestAccEvidentlyProject_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var project awstypes.Project

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_evidently_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_basic(rName, acctest.CtDisappears),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudwatchevidently.ResourceProject(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckProjectDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EvidentlyClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_evidently_project" {
				continue
			}

			_, err := tfcloudwatchevidently.FindProjectByNameOrARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Evidently Project %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckProjectExists(ctx context.Context, n string, v *awstypes.Project) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudWatch Evidently Project ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EvidentlyClient(ctx)

		output, err := tfcloudwatchevidently.FindProjectByNameOrARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccProjectConfig_basic(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_evidently_project" "test" {
  name        = %[1]q
  description = %[2]q

  tags = {
    "Key1" = "Test Project"
  }
}
`, rName, description)
}

func testAccProjectConfig_tags1(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_evidently_project" "test" {
  name        = %[1]q
  description = %[2]q

  tags = {
    "Key1" = "Test Project"
    "Key2" = "Value2a"
  }
}
`, rName, description)
}

func testAccProjectConfig_tags2(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_evidently_project" "test" {
  name        = %[1]q
  description = %[2]q

  tags = {
    "Key1" = "Test Project"
    "Key2" = "Value2b"
    "Key3" = "Value3"
  }
}
`, rName, description)
}

func testAccProjectBaseConfig(rName, rName2, rName3, rName4 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket" "test2" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[3]q
}

resource "aws_cloudwatch_log_group" "test2" {
  name = %[4]q
}
`, rName, rName2, rName3, rName4)
}

func testAccProjectConfig_dataDeliveryCloudWatchLogs(rName, rName2, rName3, rName4, rName5, selectLogGroup string) string {
	return acctest.ConfigCompose(
		testAccProjectBaseConfig(rName, rName2, rName3, rName4),
		fmt.Sprintf(`
locals {
  select_log_group = %[2]q
}

resource "aws_evidently_project" "test" {
  name = %[1]q

  data_delivery {
    cloudwatch_logs {
      log_group = local.select_log_group == "first" ? aws_cloudwatch_log_group.test.name : aws_cloudwatch_log_group.test2.name
    }
  }
}
`, rName5, selectLogGroup))
}

func testAccProjectConfig_dataDeliveryS3Bucket(rName, rName2, rName3, rName4, rName5, prefix, selectBucket string) string {
	return acctest.ConfigCompose(
		testAccProjectBaseConfig(rName, rName2, rName3, rName4),
		fmt.Sprintf(`
locals {
  select_bucket = %[3]q
}

resource "aws_evidently_project" "test" {
  name = %[1]q

  data_delivery {
    s3_destination {
      bucket = local.select_bucket == "first" ? aws_s3_bucket.test.id : aws_s3_bucket.test2.id
      prefix = %[2]q
    }
  }
}
`, rName5, prefix, selectBucket))
}
