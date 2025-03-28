// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dataexchange_test

import (
	"context"
	"errors"
	"fmt"
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
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrARN:       tfknownvalue.RegionalARNRegexp("dataexchange", regexache.MustCompile(`data-sets/\w+/revisions/\w+/assets/\w+`)),
							names.AttrCreatedAt: knownvalue.NotNull(),
							names.AttrID:        knownvalue.NotNull(),
							"import_assets_from_s3": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"asset_source": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											names.AttrBucket: knownvalue.StringExact(rName),
											names.AttrKey:    knownvalue.StringExact("test"),
										}),
									}),
								}),
							}),
							"import_assets_from_signed_url": knownvalue.ListExact([]knownvalue.Check{}),
							names.AttrName:                  knownvalue.StringExact("test"),
							"updated_at":                    knownvalue.NotNull(),
						}),
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
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrARN:       tfknownvalue.RegionalARNRegexp("dataexchange", regexache.MustCompile(`data-sets/\w+/revisions/\w+/assets/\w+`)),
							names.AttrCreatedAt: knownvalue.NotNull(),
							names.AttrID:        knownvalue.NotNull(),
							"import_assets_from_s3": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"asset_source": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											names.AttrBucket: knownvalue.StringExact(rName + "-0"),
											names.AttrKey:    knownvalue.StringExact("test-0"),
										}),
									}),
								}),
							}),
							"import_assets_from_signed_url": knownvalue.ListExact([]knownvalue.Check{}),
							names.AttrName:                  knownvalue.StringExact("test-0"),
							"updated_at":                    knownvalue.NotNull(),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrARN:       tfknownvalue.RegionalARNRegexp("dataexchange", regexache.MustCompile(`data-sets/\w+/revisions/\w+/assets/\w+`)),
							names.AttrCreatedAt: knownvalue.NotNull(),
							names.AttrID:        knownvalue.NotNull(),
							"import_assets_from_s3": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"asset_source": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											names.AttrBucket: knownvalue.StringExact(rName + "-1"),
											names.AttrKey:    knownvalue.StringExact("test-1"),
										}),
									}),
								}),
							}),
							"import_assets_from_signed_url": knownvalue.ListExact([]knownvalue.Check{}),
							names.AttrName:                  knownvalue.StringExact("test-1"),
							"updated_at":                    knownvalue.NotNull(),
						}),
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
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrARN:           tfknownvalue.RegionalARNRegexp("dataexchange", regexache.MustCompile(`data-sets/\w+/revisions/\w+/assets/\w+`)),
							names.AttrCreatedAt:     knownvalue.NotNull(),
							names.AttrID:            knownvalue.NotNull(),
							"import_assets_from_s3": knownvalue.ListExact([]knownvalue.Check{}),
							"import_assets_from_signed_url": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"filename": knownvalue.StringExact("./test-fixtures/data.json"),
								}),
							}),
							names.AttrName: knownvalue.StringExact("./test-fixtures/data.json"),
							"updated_at":   knownvalue.NotNull(),
						}),
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
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrARN:           tfknownvalue.RegionalARNRegexp("dataexchange", regexache.MustCompile(`data-sets/\w+/revisions/\w+/assets/\w+`)),
							names.AttrCreatedAt:     knownvalue.NotNull(),
							names.AttrID:            knownvalue.NotNull(),
							"import_assets_from_s3": knownvalue.ListExact([]knownvalue.Check{}),
							"import_assets_from_signed_url": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"filename": knownvalue.StringExact("./test-fixtures/data.json"),
								}),
							}),
							names.AttrName: knownvalue.StringExact("./test-fixtures/data.json"),
							"updated_at":   knownvalue.NotNull(),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrARN:           tfknownvalue.RegionalARNRegexp("dataexchange", regexache.MustCompile(`data-sets/\w+/revisions/\w+/assets/\w+`)),
							names.AttrCreatedAt:     knownvalue.NotNull(),
							names.AttrID:            knownvalue.NotNull(),
							"import_assets_from_s3": knownvalue.ListExact([]knownvalue.Check{}),
							"import_assets_from_signed_url": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"filename": knownvalue.StringExact("./test-fixtures/data2.json"),
								}),
							}),
							names.AttrName: knownvalue.StringExact("./test-fixtures/data2.json"),
							"updated_at":   knownvalue.NotNull(),
						}),
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
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrARN:           tfknownvalue.RegionalARNRegexp("dataexchange", regexache.MustCompile(`data-sets/\w+/revisions/\w+/assets/\w+`)),
							names.AttrCreatedAt:     knownvalue.NotNull(),
							names.AttrID:            knownvalue.NotNull(),
							"import_assets_from_s3": knownvalue.ListExact([]knownvalue.Check{}),
							"import_assets_from_signed_url": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"filename": knownvalue.StringExact("./test-fixtures/data.json"),
								}),
							}),
							names.AttrName: knownvalue.StringExact("./test-fixtures/data.json"),
							"updated_at":   knownvalue.NotNull(),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrARN:       tfknownvalue.RegionalARNRegexp("dataexchange", regexache.MustCompile(`data-sets/\w+/revisions/\w+/assets/\w+`)),
							names.AttrCreatedAt: knownvalue.NotNull(),
							names.AttrID:        knownvalue.NotNull(),
							"import_assets_from_s3": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"asset_source": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											names.AttrBucket: knownvalue.StringExact(rName),
											names.AttrKey:    knownvalue.StringExact("test"),
										}),
									}),
								}),
							}),
							"import_assets_from_signed_url": knownvalue.ListExact([]knownvalue.Check{}),
							names.AttrName:                  knownvalue.StringExact("test"),
							"updated_at":                    knownvalue.NotNull(),
						}),
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
