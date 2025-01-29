// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3tables_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3tables/types"
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
	tfs3tables "github.com/hashicorp/terraform-provider-aws/internal/service/s3tables"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3TablesTableBucket_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var tablebucket s3tables.GetTableBucketOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3tables_table_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableBucketConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableBucketExists(ctx, resourceName, &tablebucket),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "s3tables", "bucket/"+rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
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
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccS3TablesTableBucket_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var tablebucket s3tables.GetTableBucketOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3tables_table_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableBucketConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableBucketExists(ctx, resourceName, &tablebucket),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfs3tables.NewResourceTableBucket, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3TablesTableBucket_maintenanceConfiguration(t *testing.T) {
	ctx := acctest.Context(t)

	var tablebucket s3tables.GetTableBucketOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3tables_table_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableBucketConfig_maintenanceConfiguration(rName, awstypes.MaintenanceStatusEnabled, 20, 6),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableBucketExists(ctx, resourceName, &tablebucket),
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
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccTableBucketConfig_maintenanceConfiguration(rName, awstypes.MaintenanceStatusEnabled, 15, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableBucketExists(ctx, resourceName, &tablebucket),
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
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccTableBucketConfig_maintenanceConfiguration(rName, awstypes.MaintenanceStatusDisabled, 15, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableBucketExists(ctx, resourceName, &tablebucket),
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
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func testAccCheckTableBucketDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3TablesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3tables_table_bucket" {
				continue
			}

			_, err := tfs3tables.FindTableBucket(ctx, conn, rs.Primary.Attributes[names.AttrARN])
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.S3Tables, create.ErrActionCheckingDestroyed, tfs3tables.ResNameTableBucket, rs.Primary.ID, err)
			}

			return create.Error(names.S3Tables, create.ErrActionCheckingDestroyed, tfs3tables.ResNameTableBucket, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckTableBucketExists(ctx context.Context, name string, tablebucket *s3tables.GetTableBucketOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.S3Tables, create.ErrActionCheckingExistence, tfs3tables.ResNameTableBucket, name, errors.New("not found"))
		}

		if rs.Primary.Attributes[names.AttrARN] == "" {
			return create.Error(names.S3Tables, create.ErrActionCheckingExistence, tfs3tables.ResNameTableBucket, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3TablesClient(ctx)

		resp, err := tfs3tables.FindTableBucket(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return create.Error(names.S3Tables, create.ErrActionCheckingExistence, tfs3tables.ResNameTableBucket, rs.Primary.ID, err)
		}

		*tablebucket = *resp

		return nil
	}
}

func testAccTableBucketConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3tables_table_bucket" "test" {
  name = %[1]q
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
