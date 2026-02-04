// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3BucketMetadataConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.MetadataConfigurationResult
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_metadata_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketMetadataConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetadataConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketMetadataConfigurationExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("metadata_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrDestination: knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"table_bucket_arn":  tfknownvalue.RegionalARNExact("s3tables", "bucket/aws-s3"),
									"table_bucket_type": tfknownvalue.StringExact(awstypes.S3TablesBucketTypeAws),
									"table_namespace":   knownvalue.NotNull(),
								}),
							}),
							"inventory_table_configuration": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"table_arn":         knownvalue.Null(),
									names.AttrTableName: knownvalue.Null(),
								}),
							}),
							"journal_table_configuration": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"table_arn":         tfknownvalue.RegionalARNRegexp("s3tables", regexache.MustCompile(`bucket/aws-s3/table/.+`)),
									names.AttrTableName: knownvalue.NotNull(),
								}),
							}),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrBucket),
				ImportStateVerifyIdentifierAttribute: names.AttrBucket,
			},
		},
	})
}

func TestAccS3BucketMetadataConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.MetadataConfigurationResult
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_metadata_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketMetadataConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetadataConfigurationConfig_encryption1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketMetadataConfigurationExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("metadata_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"inventory_table_configuration": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"configuration_state": tfknownvalue.StringExact(awstypes.InventoryConfigurationStateEnabled),
									names.AttrEncryptionConfiguration: knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											names.AttrKMSKeyARN: knownvalue.Null(),
											"sse_algorithm":     tfknownvalue.StringExact(awstypes.TableSseAlgorithmAes256),
										}),
									}),
									"table_arn":         tfknownvalue.RegionalARNRegexp("s3tables", regexache.MustCompile(`bucket/aws-s3/table/.+`)),
									names.AttrTableName: knownvalue.NotNull(),
								}),
							}),
							"journal_table_configuration": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"record_expiration": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"days":       knownvalue.Null(),
											"expiration": tfknownvalue.StringExact(awstypes.ExpirationStateDisabled),
										}),
									}),
									"table_arn":         tfknownvalue.RegionalARNRegexp("s3tables", regexache.MustCompile(`bucket/aws-s3/table/.+`)),
									names.AttrTableName: knownvalue.NotNull(),
								}),
							}),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrBucket),
				ImportStateVerifyIdentifierAttribute: names.AttrBucket,
				ImportStateVerifyIgnore: []string{
					"metadata_configuration.0.inventory_table_configuration.0.encryption_configuration",
					"metadata_configuration.0.journal_table_configuration.0.encryption_configuration",
				},
			},
			{
				Config: testAccBucketMetadataConfigurationConfig_encryption2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketMetadataConfigurationExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("metadata_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"inventory_table_configuration": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"configuration_state":             tfknownvalue.StringExact(awstypes.InventoryConfigurationStateDisabled),
									names.AttrEncryptionConfiguration: knownvalue.ListSizeExact(0),
									"table_arn":                       knownvalue.Null(),
									names.AttrTableName:               knownvalue.Null(),
								}),
							}),
							"journal_table_configuration": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									names.AttrEncryptionConfiguration: knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											names.AttrKMSKeyARN: knownvalue.NotNull(),
											"sse_algorithm":     tfknownvalue.StringExact(awstypes.TableSseAlgorithmAwsKms),
										}),
									}),
									"record_expiration": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"days":       knownvalue.Int32Exact(30),
											"expiration": tfknownvalue.StringExact(awstypes.ExpirationStateEnabled),
										}),
									}),
									"table_arn":         tfknownvalue.RegionalARNRegexp("s3tables", regexache.MustCompile(`bucket/aws-s3/table/.+`)),
									names.AttrTableName: knownvalue.NotNull(),
								}),
							}),
						}),
					})),
				},
			},
		},
	})
}

func TestAccS3BucketMetadataConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.MetadataConfigurationResult
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_metadata_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketMetadataConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetadataConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketMetadataConfigurationExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfs3.ResourceBucketMetadataConfiguration, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketMetadataConfiguration_expectedBucketOwner(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.MetadataConfigurationResult
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_metadata_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketMetadataConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetadataConfigurationConfig_expectedBucketOwner(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketMetadataConfigurationExists(ctx, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), tfknownvalue.AccountID()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrBucket),
				ImportStateVerifyIdentifierAttribute: names.AttrBucket,
				ImportStateVerifyIgnore: []string{
					names.AttrExpectedBucketOwner,
				},
			},
		},
	})
}

func testAccCheckBucketMetadataConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_bucket_metadata_configuration" {
				continue
			}

			_, err := tfs3.FindBucketMetadataConfigurationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrBucket], rs.Primary.Attributes[names.AttrExpectedBucketOwner])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Bucket Metadata Configuration %s still exists", rs.Primary.Attributes[names.AttrBucket])
		}

		return nil
	}
}

func testAccCheckBucketMetadataConfigurationExists(ctx context.Context, n string, v *awstypes.MetadataConfigurationResult) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		output, err := tfs3.FindBucketMetadataConfigurationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrBucket], rs.Primary.Attributes[names.AttrExpectedBucketOwner])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccBucketMetadataConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_metadata_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  metadata_configuration {
    inventory_table_configuration {
      configuration_state = "DISABLED"
    }

    journal_table_configuration {
      record_expiration {
        days       = 7
        expiration = "ENABLED"
      }
    }
  }
}
`, rName)
}

func testAccBucketMetadataConfigurationConfig_encryption1(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_s3_bucket_metadata_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  metadata_configuration {
    inventory_table_configuration {
      configuration_state = "ENABLED"

      encryption_configuration {
        sse_algorithm = "AES256"
      }
    }

    journal_table_configuration {
      record_expiration {
        expiration = "DISABLED"
      }
    }
  }
}
`, rName)
}

func testAccBucketMetadataConfigurationConfig_encryption2(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_s3_bucket_metadata_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  metadata_configuration {
    inventory_table_configuration {
      configuration_state = "DISABLED"
    }

    journal_table_configuration {
      record_expiration {
        days       = 30
        expiration = "ENABLED"
      }

      encryption_configuration {
        sse_algorithm = "aws:kms"
        kms_key_arn   = aws_kms_key.test.arn
      }
    }
  }
}
`, rName)
}

func testAccBucketMetadataConfigurationConfig_expectedBucketOwner(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_metadata_configuration" "test" {
  bucket                = aws_s3_bucket.test.bucket
  expected_bucket_owner = data.aws_caller_identity.current.account_id

  metadata_configuration {
    inventory_table_configuration {
      configuration_state = "DISABLED"
    }

    journal_table_configuration {
      record_expiration {
        days       = 7
        expiration = "ENABLED"
      }
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

data "aws_caller_identity" "current" {}
`, rName)
}
