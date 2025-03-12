// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3DirectoryBucket_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_directory_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryBucketConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryBucketExists(ctx, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("data_redundancy"), knownvalue.StringExact("SingleAvailabilityZone")),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("s3express", regexache.MustCompile(`bucket/.+--x-s3`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrBucket), knownvalue.StringRegexp(tfs3.DirectoryBucketNameRegex)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("data_redundancy"), knownvalue.StringExact("SingleAvailabilityZone")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrForceDestroy), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrLocation), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrName: knownvalue.NotNull(),
							names.AttrType: knownvalue.StringExact("AvailabilityZone"),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), knownvalue.StringExact("Directory")),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccS3DirectoryBucket_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_directory_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryBucketConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryBucketExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfs3.ResourceDirectoryBucket, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3DirectoryBucket_forceDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_directory_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryBucketConfig_forceDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryBucketExists(ctx, resourceName),
					testAccCheckBucketAddObjects(ctx, resourceName, "data.txt", "prefix/more_data.txt"),
				),
			},
		},
	})
}

func TestAccS3DirectoryBucket_forceDestroyWithUnusualKeyBytes(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_directory_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryBucketConfig_forceDestroyUnusualKeyBytes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryBucketExists(ctx, resourceName),
					testAccCheckBucketAddObjects(ctx, resourceName, "unusual-key-bytes\x10.txt"),
				),
			},
		},
	})
}

func TestAccS3DirectoryBucket_defaultDataRedundancy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_directory_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryBucketConfig_defaultDataRedundancy(rName, "AvailabilityZone"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryBucketExists(ctx, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("data_redundancy"), knownvalue.StringExact("SingleAvailabilityZone")),
					},
				},
			},
			{
				Config: testAccDirectoryBucketConfig_defaultDataRedundancy(rName, "LocalZone"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("data_redundancy"), knownvalue.StringExact("SingleLocalZone")),
					},
				},
				ExpectError: regexache.MustCompile(`InvalidRequest: Invalid Data Redundancy value`),
			},
		},
	})
}

func TestAccS3DirectoryBucket_upgradeDefaultDataRedundancy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_directory_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: testAccCheckDirectoryBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.88.0",
					},
				},
				Config: testAccDirectoryBucketConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryBucketExists(ctx, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("data_redundancy"), knownvalue.StringExact("SingleAvailabilityZone")),
					},
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccDirectoryBucketConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryBucketExists(ctx, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func testAccCheckDirectoryBucketDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ExpressClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_directory_bucket" {
				continue
			}

			_, err := tfs3.FindBucket(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Directory Bucket %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDirectoryBucketExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ExpressClient(ctx)

		_, err := tfs3.FindBucket(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccConfigDirectoryBucket_availableAZs() string {
	// https://docs.aws.amazon.com/AmazonS3/latest/userguide/directory-bucket-az-networking.html#s3-express-endpoints-az.
	return acctest.ConfigAvailableAZsNoOptInExclude("use1-az1", "use1-az2", "use1-az3", "use2-az2", "usw2-az2", "aps1-az3", "apne1-az2", "euw1-az2")
}

func testAccDirectoryBucketConfig_baseAZ(rName string) string {
	return acctest.ConfigCompose(testAccConfigDirectoryBucket_availableAZs(), fmt.Sprintf(`
locals {
  location_name = data.aws_availability_zones.available.zone_ids[0]
  bucket        = "%[1]s--${local.location_name}--x-s3"
}
`, rName))
}

func testAccDirectoryBucketConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_baseAZ(rName), `
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }
}
`)
}

func testAccDirectoryBucketConfig_forceDestroy(rName string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_baseAZ(rName), `
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }

  force_destroy = true
}
`)
}

func testAccDirectoryBucketConfig_forceDestroyUnusualKeyBytes(rName string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_baseAZ(rName), `
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }

  force_destroy = true
}
`)
}

func testAccDirectoryBucketConfig_defaultDataRedundancy(rName, locationType string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_baseAZ(rName), fmt.Sprintf(`
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
    type = %[1]q
  }
}
`, locationType))
}
