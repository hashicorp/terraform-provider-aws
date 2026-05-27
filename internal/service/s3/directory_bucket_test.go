// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDirectoryBucketPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).S3ExpressClient(ctx)

	input := s3.ListDirectoryBucketsInput{}

	_, err := conn.ListDirectoryBuckets(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func TestAccS3DirectoryBucket_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_directory_bucket.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDirectoryBucketPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryBucketDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryBucketConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryBucketExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("data_redundancy"), tfknownvalue.StringExact(awstypes.DataRedundancySingleAvailabilityZone)),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectRegionalARNFormat(resourceName, tfjsonpath.New(names.AttrARN), "s3express", "bucket/{bucket}"),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrBucket), knownvalue.StringRegexp(directoryBucketFullNameRegex(rName))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("data_redundancy"), tfknownvalue.StringExact(awstypes.DataRedundancySingleAvailabilityZone)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrForceDestroy), knownvalue.Bool(false)),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrBucket), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrLocation), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrName: knownvalue.NotNull(),
							names.AttrType: tfknownvalue.StringExact(awstypes.LocationTypeAvailabilityZone),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.BucketTypeDirectory)),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_directory_bucket.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDirectoryBucketPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryBucketDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryBucketConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryBucketExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfs3.ResourceDirectoryBucket, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3DirectoryBucket_forceDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_directory_bucket.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDirectoryBucketPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryBucketDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryBucketConfig_forceDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryBucketExists(ctx, t, resourceName),
					testAccCheckBucketAddObjects(ctx, t, resourceName, "data.txt", "prefix/more_data.txt"),
				),
			},
		},
	})
}

func TestAccS3DirectoryBucket_forceDestroyWithUnusualKeyBytes(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_directory_bucket.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDirectoryBucketPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryBucketDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryBucketConfig_forceDestroyUnusualKeyBytes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryBucketExists(ctx, t, resourceName),
					testAccCheckBucketAddObjects(ctx, t, resourceName, "unusual-key-bytes\x10.txt"),
				),
			},
		},
	})
}

func TestAccS3DirectoryBucket_defaultDataRedundancy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_directory_bucket.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDirectoryBucketPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryBucketDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryBucketConfig_defaultDataRedundancy(rName, awstypes.LocationTypeAvailabilityZone),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryBucketExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("data_redundancy"), tfknownvalue.StringExact(awstypes.DataRedundancySingleAvailabilityZone)),
					},
				},
			},
			{
				Config: testAccDirectoryBucketConfig_defaultDataRedundancy(rName, awstypes.LocationTypeLocalZone),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("data_redundancy"), tfknownvalue.StringExact(awstypes.DataRedundancySingleLocalZone)),
					},
				},
				ExpectError: regexache.MustCompile(`InvalidRequest: Invalid Data Redundancy value`),
			},
		},
	})
}

func TestAccS3DirectoryBucket_upgradeDefaultDataRedundancy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_directory_bucket.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t); testAccDirectoryBucketPreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: testAccCheckDirectoryBucketDestroy(ctx, t),
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
					testAccCheckDirectoryBucketExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("data_redundancy"), tfknownvalue.StringExact(awstypes.DataRedundancySingleAvailabilityZone)),
					},
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccDirectoryBucketConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryBucketExists(ctx, t, resourceName),
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

func testAccCheckDirectoryBucketDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).S3ExpressClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_directory_bucket" {
				continue
			}

			_, err := tfs3.FindBucket(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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

func testAccCheckDirectoryBucketExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).S3ExpressClient(ctx)

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
  location_name     = data.aws_availability_zones.available.zone_ids[0]
  bucket            = "%[1]s--${local.location_name}--x-s3"
  access_point_name = "%[1]s--${local.location_name}--xa-s3"
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

func testAccDirectoryBucketConfig_defaultDataRedundancy(rName string, locationType awstypes.LocationType) string {
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

func directoryBucketFullNameRegex(name string) *regexp.Regexp {
	return regexache.MustCompile(`^` + name + tfs3.DirectoryBucketNameSuffixRegexPattern + `$`)
}
