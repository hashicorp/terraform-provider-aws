// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3tables_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3tables/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfs3tables "github.com/hashicorp/terraform-provider-aws/internal/service/s3tables"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3TablesTableBucket_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3tables.GetTableBucketOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3tables_table_bucket.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableBucketDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableBucketConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableBucketExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "s3tables", "bucket/"+rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("maintenance_configuration"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"iceberg_unreferenced_file_removal": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"settings": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"non_current_days":  knownvalue.Int32Exact(10),
								"unreferenced_days": knownvalue.Int32Exact(3),
							}),
							names.AttrStatus: tfknownvalue.StringExact(awstypes.MaintenanceStatusEnabled),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccTableBucketImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateVerifyIgnore:              []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccS3TablesTableBucket_encryptionConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3tables.GetTableBucketOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3tables_table_bucket.test"
	resourceKeyOne := "aws_kms_key.test"
	resourceKeyTwo := "aws_kms_key.test2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableBucketDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableBucketConfig_encryptionConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableBucketExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "s3tables", "bucket/"+rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_configuration.kms_key_arn", resourceKeyOne, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.sse_algorithm", "aws:kms"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("maintenance_configuration"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"iceberg_unreferenced_file_removal": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"settings": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"non_current_days":  knownvalue.Int32Exact(10),
								"unreferenced_days": knownvalue.Int32Exact(3),
							}),
							names.AttrStatus: tfknownvalue.StringExact(awstypes.MaintenanceStatusEnabled),
						}),
					})),
				},
			},
			{
				Config: testAccTableBucketConfig_encryptionConfigurationUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableBucketExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "s3tables", "bucket/"+rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_configuration.kms_key_arn", resourceKeyTwo, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.sse_algorithm", "aws:kms"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("maintenance_configuration"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"iceberg_unreferenced_file_removal": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"settings": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"non_current_days":  knownvalue.Int32Exact(10),
								"unreferenced_days": knownvalue.Int32Exact(3),
							}),
							names.AttrStatus: tfknownvalue.StringExact(awstypes.MaintenanceStatusEnabled),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccTableBucketImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateVerifyIgnore:              []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccS3TablesTableBucket_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3tables.GetTableBucketOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3tables_table_bucket.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableBucketDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableBucketConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableBucketExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfs3tables.ResourceTableBucket, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3TablesTableBucket_maintenanceConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3tables.GetTableBucketOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3tables_table_bucket.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableBucketDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableBucketConfig_maintenanceConfiguration(rName, awstypes.MaintenanceStatusEnabled, 20, 6),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableBucketExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("maintenance_configuration"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"iceberg_unreferenced_file_removal": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"settings": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"non_current_days":  knownvalue.Int32Exact(20),
								"unreferenced_days": knownvalue.Int32Exact(6),
							}),
							names.AttrStatus: tfknownvalue.StringExact(awstypes.MaintenanceStatusEnabled),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccTableBucketImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateVerifyIgnore:              []string{names.AttrForceDestroy},
			},
			{
				Config: testAccTableBucketConfig_maintenanceConfiguration(rName, awstypes.MaintenanceStatusEnabled, 15, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableBucketExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("maintenance_configuration"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"iceberg_unreferenced_file_removal": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"settings": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"non_current_days":  knownvalue.Int32Exact(15),
								"unreferenced_days": knownvalue.Int32Exact(4),
							}),
							names.AttrStatus: tfknownvalue.StringExact(awstypes.MaintenanceStatusEnabled),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccTableBucketImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateVerifyIgnore:              []string{names.AttrForceDestroy},
			},
			{
				Config: testAccTableBucketConfig_maintenanceConfiguration(rName, awstypes.MaintenanceStatusDisabled, 15, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableBucketExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("maintenance_configuration"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"iceberg_unreferenced_file_removal": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"settings": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"non_current_days":  knownvalue.Int32Exact(15),
								"unreferenced_days": knownvalue.Int32Exact(4),
							}),
							names.AttrStatus: tfknownvalue.StringExact(awstypes.MaintenanceStatusDisabled),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccTableBucketImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateVerifyIgnore:              []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccS3TablesTableBucket_forceDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3tables.GetTableBucketOutput
	resourceName := "aws_s3tables_table_bucket.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableBucketDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableBucketConfig_forceDestroy(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableBucketExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtTrue),
					testAccCheckTableBucketAddTables(ctx, t, resourceName, "namespace1", "table1"),
				),
			},
		},
	})
}

func TestAccS3TablesTableBucket_forceDestroyMultipleNamespacesAndTables(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3tables.GetTableBucketOutput
	resourceName := "aws_s3tables_table_bucket.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableBucketDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableBucketConfig_forceDestroy(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableBucketExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtTrue),
					testAccCheckTableBucketAddTables(ctx, t, resourceName, "namespace1", "table1"),
					testAccCheckTableBucketAddTables(ctx, t, resourceName, "namespace2", "table2", "table3"),
					testAccCheckTableBucketAddTables(ctx, t, resourceName, "namespace3", "table4", "table5", "table6"),
				),
			},
		},
	})
}

func testAccCheckTableBucketAddTables(ctx context.Context, t *testing.T, n string, namespace string, tableNames ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := acctest.ProviderMeta(ctx, t).S3TablesClient(ctx)

		// First, create the namespace if it doesn't exist
		_, err := conn.CreateNamespace(ctx, &s3tables.CreateNamespaceInput{
			TableBucketARN: aws.String(rs.Primary.Attributes[names.AttrARN]),
			Namespace:      []string{namespace},
		})
		if err != nil {
			// Ignore if namespace already exists
			if !errs.IsA[*awstypes.ConflictException](err) {
				return fmt.Errorf("CreateNamespace error: %w", err)
			}
		}

		// Create each table
		for _, tableName := range tableNames {
			_, err := conn.CreateTable(ctx, &s3tables.CreateTableInput{
				TableBucketARN: aws.String(rs.Primary.Attributes[names.AttrARN]),
				Namespace:      aws.String(namespace),
				Name:           aws.String(tableName),
				Format:         awstypes.OpenTableFormatIceberg,
			})
			if err != nil {
				// Ignore if table already exists
				if !errs.IsA[*awstypes.ConflictException](err) {
					return fmt.Errorf("CreateTable error for table %s: %w", tableName, err)
				}
			}
		}

		return nil
	}
}

func testAccCheckTableBucketDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).S3TablesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3tables_table_bucket" {
				continue
			}

			_, err := tfs3tables.FindTableBucketByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Tables Table Bucket %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckTableBucketExists(ctx context.Context, t *testing.T, n string, v *s3tables.GetTableBucketOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).S3TablesClient(ctx)

		output, err := tfs3tables.FindTableBucketByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccTableBucketImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return acctest.AttrImportStateIdFunc(resourceName, names.AttrARN)
}

func testAccTableBucketConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3tables_table_bucket" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTableBucketConfig_encryptionConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3tables_table_bucket" "test" {
  name = %[1]q

  encryption_configuration = {
    kms_key_arn   = aws_kms_key.test.arn
    sse_algorithm = "aws:kms"
  }
}

resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}
`, rName)
}

func testAccTableBucketConfig_encryptionConfigurationUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3tables_table_bucket" "test" {
  name = %[1]q

  encryption_configuration = {
    kms_key_arn   = aws_kms_key.test2.arn
    sse_algorithm = "aws:kms"
  }
}

resource "aws_kms_key" "test2" {
  deletion_window_in_days = 7
}
`, rName)
}

func testAccTableBucketConfig_maintenanceConfiguration(rName string, status awstypes.MaintenanceStatus, nonCurrentDays, unreferencedDays int32) string {
	return fmt.Sprintf(`
resource "aws_s3tables_table_bucket" "test" {
  name = %[1]q

  maintenance_configuration = {
    iceberg_unreferenced_file_removal = {
      settings = {
        non_current_days  = %[3]d
        unreferenced_days = %[4]d
      }
      status = %[2]q
    }
  }
}
`, rName, status, nonCurrentDays, unreferencedDays)
}

func testAccTableBucketConfig_forceDestroy(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3tables_table_bucket" "test" {
  name          = %[1]q
  force_destroy = true
}
`, rName)
}
