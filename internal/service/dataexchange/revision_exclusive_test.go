// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dataexchange_test

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/dataexchange"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfdataexchange "github.com/hashicorp/terraform-provider-aws/internal/service/dataexchange"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// lintignore:AT002
func TestAccDataExchangeRevisionExclusive_importFromS3(t *testing.T) {
	ctx := acctest.Context(t)

	var revision dataexchange.GetRevisionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dataexchange_revision_exclusive.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRevisionExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRevisionExclusiveConfig_importFromS3(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRevisionExclusiveExists(ctx, resourceName, &revision),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dataexchange", "data-sets/{data_set_id}/revisions/{id}"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrComment),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrPair(resourceName, "data_set_id", "aws_dataexchange_data_set.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("asset"), knownvalue.SetExact([]knownvalue.Check{
						checkAssetImportFromS3(rName, "test"),
					})),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
			// {
			// 	ResourceName:      resourceName,
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// },
		},
	})
}

func TestAccDataExchangeRevisionExclusive_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var revisionexclusive dataexchange.GetRevisionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dataexchange_revision_exclusive.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRevisionExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRevisionExclusiveConfig_importFromS3(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRevisionExclusiveExists(ctx, resourceName, &revisionexclusive),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfdataexchange.ResourceRevisionExclusive, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// lintignore:AT002
func TestAccDataExchangeRevisionExclusive_importMultipleFromS3(t *testing.T) {
	ctx := acctest.Context(t)

	var revision dataexchange.GetRevisionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dataexchange_revision_exclusive.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRevisionExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRevisionExclusiveConfig_importMultipleFromS3(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRevisionExclusiveExists(ctx, resourceName, &revision),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("asset"), knownvalue.SetExact([]knownvalue.Check{
						checkAssetImportFromS3(rName+"-0", "test-0"),
						checkAssetImportFromS3(rName+"-1", "test-1"),
					})),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
			// {
			// 	ResourceName:      resourceName,
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// },
		},
	})
}

// lintignore:AT002
func TestAccDataExchangeRevisionExclusive_importFromSignedURL(t *testing.T) {
	ctx := acctest.Context(t)

	var revision dataexchange.GetRevisionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dataexchange_revision_exclusive.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRevisionExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRevisionExclusiveConfig_importFromSignedURL(rName, "test-fixtures/data.json"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRevisionExclusiveExists(ctx, resourceName, &revision),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dataexchange", "data-sets/{data_set_id}/revisions/{id}"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrComment),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrPair(resourceName, "data_set_id", "aws_dataexchange_data_set.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("asset"), knownvalue.SetExact([]knownvalue.Check{
						checkAssetImportFromSignedURL("./test-fixtures/data.json"),
					})),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
			// {
			// 	ResourceName:      resourceName,
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// },
		},
	})
}

// lintignore:AT002
func TestAccDataExchangeRevisionExclusive_importMultipleFromSignedURL(t *testing.T) {
	ctx := acctest.Context(t)

	var revision dataexchange.GetRevisionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dataexchange_revision_exclusive.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRevisionExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRevisionExclusiveConfig_importMultipleFromSignedURL(rName, "test-fixtures/data.json", "test-fixtures/data2.json"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRevisionExclusiveExists(ctx, resourceName, &revision),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dataexchange", "data-sets/{data_set_id}/revisions/{id}"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrComment),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrPair(resourceName, "data_set_id", "aws_dataexchange_data_set.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("asset"), knownvalue.SetExact([]knownvalue.Check{
						checkAssetImportFromSignedURL("./test-fixtures/data.json"),
						checkAssetImportFromSignedURL("./test-fixtures/data2.json"),
					})),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
			// {
			// 	ResourceName:      resourceName,
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// },
		},
	})
}

// lintignore:AT002
func TestAccDataExchangeRevisionExclusive_importFromS3AndSignedURL(t *testing.T) {
	ctx := acctest.Context(t)

	var revision dataexchange.GetRevisionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dataexchange_revision_exclusive.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRevisionExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRevisionExclusiveConfig_importFromS3AndSignedURL(rName, "test-fixtures/data.json"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRevisionExclusiveExists(ctx, resourceName, &revision),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dataexchange", "data-sets/{data_set_id}/revisions/{id}"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrComment),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrPair(resourceName, "data_set_id", "aws_dataexchange_data_set.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("asset"), knownvalue.SetExact([]knownvalue.Check{
						checkAssetImportFromSignedURL("./test-fixtures/data.json"),
						checkAssetImportFromS3(rName, "test"),
					})),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
			// {
			// 	ResourceName:      resourceName,
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// },
		},
	})
}

func TestAccDataExchangeRevisionExclusive_createS3DataAccessFromS3Bucket_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var revision dataexchange.GetRevisionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dataexchange_revision_exclusive.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRevisionExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRevisionExclusiveConfig_createS3DataAccessFromS3Bucket_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRevisionExclusiveExists(ctx, resourceName, &revision),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dataexchange", "data-sets/{data_set_id}/revisions/{id}"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrComment),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrPair(resourceName, "data_set_id", "aws_dataexchange_data_set.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("asset"), knownvalue.SetExact([]knownvalue.Check{
						checkAssetS3DataAccess(rName),
					})),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
			// {
			// 	ResourceName:      resourceName,
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// },
		},
	})
}

func TestAccDataExchangeRevisionExclusive_createS3DataAccessFromS3Bucket_multiple(t *testing.T) {
	ctx := acctest.Context(t)

	var revision dataexchange.GetRevisionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dataexchange_revision_exclusive.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRevisionExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRevisionExclusiveConfig_createS3DataAccessFromS3Bucket_multiple(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRevisionExclusiveExists(ctx, resourceName, &revision),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dataexchange", "data-sets/{data_set_id}/revisions/{id}"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrComment),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrPair(resourceName, "data_set_id", "aws_dataexchange_data_set.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("asset"), knownvalue.SetExact([]knownvalue.Check{
						checkAssetS3DataAccess(rName + "-0"),
						checkAssetS3DataAccess(rName + "-1"),
					})),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
			// {
			// 	ResourceName:      resourceName,
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// },
		},
	})
}

// Basic
// TODO: empty bucket
// TODO: no permissions on bucket

// Full Bucket, customer managed keys
// TODO: basic
// TODO: can't share keys

// Partial Bucket
// TODO: basic

// Other tests:
// Wrong data type

func testAccCheckRevisionExclusiveDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataExchangeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dataexchange_revision_exclusive" {
				continue
			}

			// TIP: ==== FINDERS ====
			// The find function should be exported. Since it won't be used outside of the package, it can be exported
			// in the `exports_test.go` file.
			_, err := tfdataexchange.FindRevisionByID(ctx, conn, rs.Primary.Attributes["data_set_id"], rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.DataExchange, create.ErrActionCheckingDestroyed, tfdataexchange.ResNameRevisionExclusive, rs.Primary.ID, err)
			}

			return create.Error(names.DataExchange, create.ErrActionCheckingDestroyed, tfdataexchange.ResNameRevisionExclusive, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckRevisionExclusiveExists(ctx context.Context, name string, revisionexclusive *dataexchange.GetRevisionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DataExchange, create.ErrActionCheckingExistence, tfdataexchange.ResNameRevisionExclusive, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.DataExchange, create.ErrActionCheckingExistence, tfdataexchange.ResNameRevisionExclusive, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataExchangeClient(ctx)

		resp, err := tfdataexchange.FindRevisionByID(ctx, conn, rs.Primary.Attributes["data_set_id"], rs.Primary.ID)
		if err != nil {
			return create.Error(names.DataExchange, create.ErrActionCheckingExistence, tfdataexchange.ResNameRevisionExclusive, rs.Primary.ID, err)
		}

		*revisionexclusive = *resp

		return nil
	}
}

func checkAssetImportFromS3(bucket, key string) knownvalue.Check {
	checks := assetDefaults()
	maps.Copy(checks, map[string]knownvalue.Check{
		"import_assets_from_s3": knownvalue.ListExact([]knownvalue.Check{
			knownvalue.ObjectExact(map[string]knownvalue.Check{
				"asset_source": knownvalue.ListExact([]knownvalue.Check{
					knownvalue.ObjectExact(map[string]knownvalue.Check{
						names.AttrBucket: knownvalue.StringExact(bucket),
						names.AttrKey:    knownvalue.StringExact(key),
					}),
				}),
			}),
		}),
		names.AttrName: knownvalue.StringExact(key),
	})
	return knownvalue.ObjectExact(
		checks,
	)
}

func checkAssetImportFromSignedURL(filename string) knownvalue.Check {
	checks := assetDefaults()
	maps.Copy(checks, map[string]knownvalue.Check{
		"import_assets_from_signed_url": knownvalue.ListExact([]knownvalue.Check{
			knownvalue.ObjectExact(map[string]knownvalue.Check{
				"filename": knownvalue.StringExact(filename),
			}),
		}),
		names.AttrName: knownvalue.StringExact(filename),
	})
	return knownvalue.ObjectExact(
		checks,
	)
}

func checkAssetS3DataAccess(bucket string) knownvalue.Check {
	checks := assetDefaults()
	maps.Copy(checks, map[string]knownvalue.Check{
		"create_s3_data_access_from_s3_bucket": knownvalue.ListExact([]knownvalue.Check{
			knownvalue.ObjectExact(map[string]knownvalue.Check{
				"asset_source": knownvalue.ListExact([]knownvalue.Check{
					knownvalue.ObjectExact(map[string]knownvalue.Check{
						names.AttrBucket: knownvalue.StringExact(bucket),
					}),
				}),
			}),
		}),
		names.AttrName: knownvalue.StringRegexp(regexache.MustCompile(`^s3-data-access-[a-f0-9]{32}$`)), // `s3-data-access-<asset id>`
	})
	return knownvalue.ObjectExact(
		checks,
	)
}

func assetDefaults() map[string]knownvalue.Check {
	return map[string]knownvalue.Check{
		// ARN format `data-sets/<dataset id>/revisions/<revision id>/assets/<asset id>`
		names.AttrARN:                          tfknownvalue.RegionalARNRegexp("dataexchange", regexache.MustCompile(`data-sets/\w+/revisions/\w+/assets/[a-f0-9]{32}`)),
		names.AttrCreatedAt:                    knownvalue.NotNull(),
		names.AttrID:                           knownvalue.NotNull(),
		"create_s3_data_access_from_s3_bucket": knownvalue.ListExact([]knownvalue.Check{}),
		"import_assets_from_s3":                knownvalue.ListExact([]knownvalue.Check{}),
		"import_assets_from_signed_url":        knownvalue.ListExact([]knownvalue.Check{}),
		"updated_at":                           knownvalue.NotNull(),
	}
}

func testAccRevisionExclusiveConfig_importFromS3(rName string) string {
	return fmt.Sprintf(`
resource "aws_dataexchange_revision_exclusive" "test" {
  data_set_id = aws_dataexchange_data_set.test.id

  asset {
    import_assets_from_s3 {
      asset_source {
        bucket = aws_s3_object.test.bucket
        key    = aws_s3_object.test.key
      }
    }
  }
}

resource "aws_dataexchange_data_set" "test" {
  asset_type  = "S3_SNAPSHOT"
  description = %[1]q
  name        = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "test"
  content = "test"
}
`, rName)
}

func testAccRevisionExclusiveConfig_importMultipleFromS3(rName string) string {
	return fmt.Sprintf(`
resource "aws_dataexchange_revision_exclusive" "test" {
  data_set_id = aws_dataexchange_data_set.test.id

  asset {
    import_assets_from_s3 {
      asset_source {
        bucket = aws_s3_object.test[0].bucket
        key    = aws_s3_object.test[0].key
      }
    }
  }
  asset {
    import_assets_from_s3 {
      asset_source {
        bucket = aws_s3_object.test[1].bucket
        key    = aws_s3_object.test[1].key
      }
    }
  }
}

resource "aws_dataexchange_data_set" "test" {
  asset_type  = "S3_SNAPSHOT"
  description = %[1]q
  name        = %[1]q
}

resource "aws_s3_bucket" "test" {
  count = 2

  bucket        = "%[1]s-${count.index}"
  force_destroy = true
}

resource "aws_s3_object" "test" {
  count = 2

  bucket  = aws_s3_bucket.test[count.index].bucket
  key     = "test-${count.index}"
  content = "test"
}
`, rName)
}

func testAccRevisionExclusiveConfig_importFromSignedURL(rName, filename string) string {
	return fmt.Sprintf(`
resource "aws_dataexchange_revision_exclusive" "test" {
  data_set_id = aws_dataexchange_data_set.test.id

  asset {
    import_assets_from_signed_url {
      filename = "${path.module}/%[2]s"
    }
  }
}

resource "aws_dataexchange_data_set" "test" {
  asset_type  = "S3_SNAPSHOT"
  description = %[1]q
  name        = %[1]q
}
`, rName, filename)
}

func testAccRevisionExclusiveConfig_importMultipleFromSignedURL(rName, filename, filename2 string) string {
	return fmt.Sprintf(`
resource "aws_dataexchange_revision_exclusive" "test" {
  data_set_id = aws_dataexchange_data_set.test.id

  asset {
    import_assets_from_signed_url {
      filename = "${path.module}/%[2]s"
    }
  }

  asset {
    import_assets_from_signed_url {
      filename = "${path.module}/%[3]s"
    }
  }
}

resource "aws_dataexchange_data_set" "test" {
  asset_type  = "S3_SNAPSHOT"
  description = %[1]q
  name        = %[1]q
}
`, rName, filename, filename2)
}

func testAccRevisionExclusiveConfig_importFromS3AndSignedURL(rName, filename string) string {
	return fmt.Sprintf(`
resource "aws_dataexchange_revision_exclusive" "test" {
  data_set_id = aws_dataexchange_data_set.test.id

  asset {
    import_assets_from_signed_url {
      filename = "${path.module}/%[2]s"
    }
  }

  asset {
    import_assets_from_s3 {
      asset_source {
        bucket = aws_s3_object.test.bucket
        key    = aws_s3_object.test.key
      }
    }
  }
}

resource "aws_dataexchange_data_set" "test" {
  asset_type  = "S3_SNAPSHOT"
  description = %[1]q
  name        = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "test"
  content = "test"
}
`, rName, filename)
}

func testAccRevisionExclusiveConfig_createS3DataAccessFromS3Bucket_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_dataexchange_revision_exclusive" "test" {
  data_set_id = aws_dataexchange_data_set.test.id

  asset {
    create_s3_data_access_from_s3_bucket {
      asset_source {
        bucket = aws_s3_object.test.bucket
      }
    }
  }

  depends_on = [
    aws_s3_object.test,
    aws_s3_bucket_policy.test,
  ]
}

resource "aws_dataexchange_data_set" "test" {
  asset_type  = "S3_DATA_ACCESS"
  description = %[1]q
  name        = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "test"
  content = "test"
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.bucket
  policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    effect = "Allow"

    actions = [
      "s3:GetObject",
      "s3:ListBucket",
    ]

    resources = [
      aws_s3_bucket.test.arn,
      "${aws_s3_bucket.test.arn}/*",
    ]

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    condition {
      test     = "StringEquals"
      variable = "s3:DataAccessPointAccount"
      values = [
        "337040091392",
        "504002150500",
        "366362662752",
        "330489627928",
        "291973504423",
        "461002523379",
        "036905324694",
        "540564263739",
        "675969394711",
        "108584782536",
        "844053218156",
      ]
    }
  }
}
`, rName)
}

func testAccRevisionExclusiveConfig_createS3DataAccessFromS3Bucket_multiple(rName string) string {
	return fmt.Sprintf(`
resource "aws_dataexchange_revision_exclusive" "test" {
  data_set_id = aws_dataexchange_data_set.test.id

  asset {
    create_s3_data_access_from_s3_bucket {
      asset_source {
        bucket = aws_s3_object.test[0].bucket
      }
    }
  }

  asset {
    create_s3_data_access_from_s3_bucket {
      asset_source {
        bucket = aws_s3_object.test[1].bucket
      }
    }
  }

  depends_on = [
    aws_s3_object.test,
    aws_s3_bucket_policy.test,
  ]
}

resource "aws_dataexchange_data_set" "test" {
  asset_type  = "S3_DATA_ACCESS"
  description = %[1]q
  name        = %[1]q
}

resource "aws_s3_bucket" "test" {
  count = 2

  bucket        = "%[1]s-${count.index}"
  force_destroy = true
}

resource "aws_s3_object" "test" {
  count = 2

  bucket  = aws_s3_bucket.test[count.index].bucket
  key     = "test-${count.index}"
  content = "test"
}

resource "aws_s3_bucket_policy" "test" {
  count = 2

  bucket  = aws_s3_bucket.test[count.index].bucket
  policy = data.aws_iam_policy_document.test[count.index].json
}

data "aws_iam_policy_document" "test" {
  count = 2

  statement {
    effect = "Allow"

    actions = [
      "s3:GetObject",
      "s3:ListBucket",
    ]

    resources = [
      aws_s3_bucket.test[count.index].arn,
      "${aws_s3_bucket.test[count.index].arn}/*",
    ]

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    condition {
      test     = "StringEquals"
      variable = "s3:DataAccessPointAccount"
      values = [
        "337040091392",
        "504002150500",
        "366362662752",
        "330489627928",
        "291973504423",
        "461002523379",
        "036905324694",
        "540564263739",
        "675969394711",
        "108584782536",
        "844053218156",
      ]
    }
  }
}
`, rName)
}
