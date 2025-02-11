// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3tables_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3tables/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
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
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfs3tables "github.com/hashicorp/terraform-provider-aws/internal/service/s3tables"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3TablesTable_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var table s3tables.GetTableOutput
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	namespace := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_s3tables_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_basic(rName, namespace, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &table),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "s3tables", regexache.MustCompile("bucket/"+bucketName+"/table/"+verify.UUIDRegexPattern+"$")),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "created_by"),
					resource.TestCheckResourceAttr(resourceName, names.AttrFormat, "ICEBERG"),
					resource.TestCheckNoResourceAttr(resourceName, "metadata_location"),
					resource.TestCheckResourceAttrPair(resourceName, "modified_at", resourceName, names.AttrCreatedAt),
					resource.TestCheckNoResourceAttr(resourceName, "modified_by"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrNamespace, "aws_s3tables_namespace.test", names.AttrNamespace),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttrPair(resourceName, "table_bucket_arn", "aws_s3tables_table_bucket.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.TableTypeCustomer)),
					resource.TestCheckResourceAttrSet(resourceName, "version_token"),
					func(s *terraform.State) error {
						tableID, err := tfs3tables.TableIDFromTableARN(aws.ToString(table.TableARN))
						if err != nil {
							return err
						}
						return resource.TestMatchResourceAttr(resourceName, "warehouse_location", regexache.MustCompile("^s3://"+tableID[:19]+".+--table-s3$"))(s)
					},
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("maintenance_configuration"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"iceberg_compaction": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"settings": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"target_file_size_mb": knownvalue.Int32Exact(512),
							}),
							names.AttrStatus: tfknownvalue.StringExact(awstypes.MaintenanceStatusEnabled),
						}),
						"iceberg_snapshot_management": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"settings": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"max_snapshot_age_hours": knownvalue.Int32Exact(120),
								"min_snapshots_to_keep":  knownvalue.Int32Exact(1),
							}),
							names.AttrStatus: tfknownvalue.StringExact(awstypes.MaintenanceStatusEnabled),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccTableImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccS3TablesTable_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var table s3tables.GetTableOutput
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	namespace := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_s3tables_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_basic(rName, namespace, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &table),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfs3tables.NewResourceTable, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3TablesTable_rename(t *testing.T) {
	ctx := acctest.Context(t)

	var table s3tables.GetTableOutput
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	namespace := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	rNameUpdated := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_s3tables_table.test"

	createdAtNoChange := statecheck.CompareValue(compare.ValuesSame())
	createdByNoChange := statecheck.CompareValue(compare.ValuesSame())
	modifiedAtChange := statecheck.CompareValue(compare.ValuesDiffer())
	modifiedByChange := statecheck.CompareValue(compare.ValuesDiffer())
	versionNoChange := statecheck.CompareValue(compare.ValuesSame())
	warehouseLocationNoChange := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_basic(rName, namespace, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					createdAtNoChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrCreatedAt)),
					createdByNoChange.AddStateValue(resourceName, tfjsonpath.New("created_by")),
					modifiedAtChange.AddStateValue(resourceName, tfjsonpath.New("modified_at")),
					modifiedByChange.AddStateValue(resourceName, tfjsonpath.New("modified_by")),
					versionNoChange.AddStateValue(resourceName, tfjsonpath.New("version_token")),
					warehouseLocationNoChange.AddStateValue(resourceName, tfjsonpath.New("warehouse_location")),
				},
			},
			{
				Config: testAccTableConfig_basic(rNameUpdated, namespace, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "modified_at"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "modified_by"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rNameUpdated)),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					createdAtNoChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrCreatedAt)),
					createdByNoChange.AddStateValue(resourceName, tfjsonpath.New("created_by")),
					modifiedAtChange.AddStateValue(resourceName, tfjsonpath.New("modified_at")),
					modifiedByChange.AddStateValue(resourceName, tfjsonpath.New("modified_by")),
					versionNoChange.AddStateValue(resourceName, tfjsonpath.New("version_token")),
					warehouseLocationNoChange.AddStateValue(resourceName, tfjsonpath.New("warehouse_location")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccTableImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccS3TablesTable_updateNamespace(t *testing.T) {
	ctx := acctest.Context(t)

	var table s3tables.GetTableOutput
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	namespace := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	namespaceUpdated := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_s3tables_table.test"

	createdAtNoChange := statecheck.CompareValue(compare.ValuesSame())
	createdByNoChange := statecheck.CompareValue(compare.ValuesSame())
	modifiedAtChange := statecheck.CompareValue(compare.ValuesDiffer())
	modifiedByChange := statecheck.CompareValue(compare.ValuesDiffer())
	versionNoChange := statecheck.CompareValue(compare.ValuesSame())
	warehouseLocationNoChange := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_basic(rName, namespace, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamespace, namespace),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					createdAtNoChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrCreatedAt)),
					createdByNoChange.AddStateValue(resourceName, tfjsonpath.New("created_by")),
					modifiedAtChange.AddStateValue(resourceName, tfjsonpath.New("modified_at")),
					modifiedByChange.AddStateValue(resourceName, tfjsonpath.New("modified_by")),
					versionNoChange.AddStateValue(resourceName, tfjsonpath.New("version_token")),
					warehouseLocationNoChange.AddStateValue(resourceName, tfjsonpath.New("warehouse_location")),
				},
			},
			{
				Config: testAccTableConfig_basic(rName, namespaceUpdated, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamespace, namespaceUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "modified_at"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "modified_by"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrNamespace), knownvalue.StringExact(namespaceUpdated)),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					createdAtNoChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrCreatedAt)),
					createdByNoChange.AddStateValue(resourceName, tfjsonpath.New("created_by")),
					modifiedAtChange.AddStateValue(resourceName, tfjsonpath.New("modified_at")),
					modifiedByChange.AddStateValue(resourceName, tfjsonpath.New("modified_by")),
					versionNoChange.AddStateValue(resourceName, tfjsonpath.New("version_token")),
					warehouseLocationNoChange.AddStateValue(resourceName, tfjsonpath.New("warehouse_location")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccTableImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccS3TablesTable_updateNameAndNamespace(t *testing.T) {
	ctx := acctest.Context(t)

	var table s3tables.GetTableOutput
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	namespace := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	namespaceUpdated := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	rNameUpdated := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_s3tables_table.test"

	createdAtNoChange := statecheck.CompareValue(compare.ValuesSame())
	createdByNoChange := statecheck.CompareValue(compare.ValuesSame())
	modifiedAtChange := statecheck.CompareValue(compare.ValuesDiffer())
	modifiedByChange := statecheck.CompareValue(compare.ValuesDiffer())
	versionNoChange := statecheck.CompareValue(compare.ValuesSame())
	warehouseLocationNoChange := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_basic(rName, namespace, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamespace, namespace),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					createdAtNoChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrCreatedAt)),
					createdByNoChange.AddStateValue(resourceName, tfjsonpath.New("created_by")),
					modifiedAtChange.AddStateValue(resourceName, tfjsonpath.New("modified_at")),
					modifiedByChange.AddStateValue(resourceName, tfjsonpath.New("modified_by")),
					versionNoChange.AddStateValue(resourceName, tfjsonpath.New("version_token")),
					warehouseLocationNoChange.AddStateValue(resourceName, tfjsonpath.New("warehouse_location")),
				},
			},
			{
				Config: testAccTableConfig_basic(rNameUpdated, namespaceUpdated, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamespace, namespaceUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "modified_at"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "modified_by"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rNameUpdated)),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrNamespace), knownvalue.StringExact(namespaceUpdated)),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					createdAtNoChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrCreatedAt)),
					createdByNoChange.AddStateValue(resourceName, tfjsonpath.New("created_by")),
					modifiedAtChange.AddStateValue(resourceName, tfjsonpath.New("modified_at")),
					modifiedByChange.AddStateValue(resourceName, tfjsonpath.New("modified_by")),
					versionNoChange.AddStateValue(resourceName, tfjsonpath.New("version_token")),
					warehouseLocationNoChange.AddStateValue(resourceName, tfjsonpath.New("warehouse_location")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccTableImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccS3TablesTable_maintenanceConfiguration(t *testing.T) {
	ctx := acctest.Context(t)

	var table s3tables.GetTableOutput
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	namespace := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_s3tables_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_maintenanceConfiguration(rName, namespace, bucketName, 64, 24, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &table),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("maintenance_configuration"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"iceberg_compaction": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"settings": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"target_file_size_mb": knownvalue.Int32Exact(64),
							}),
							names.AttrStatus: tfknownvalue.StringExact(awstypes.MaintenanceStatusEnabled),
						}),
						"iceberg_snapshot_management": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"settings": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"max_snapshot_age_hours": knownvalue.Int32Exact(24),
								"min_snapshots_to_keep":  knownvalue.Int32Exact(2),
							}),
							names.AttrStatus: tfknownvalue.StringExact(awstypes.MaintenanceStatusEnabled),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccTableImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccTableConfig_maintenanceConfiguration(rName, namespace, bucketName, 128, 48, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &table),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("maintenance_configuration"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"iceberg_compaction": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"settings": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"target_file_size_mb": knownvalue.Int32Exact(128),
							}),
							names.AttrStatus: tfknownvalue.StringExact(awstypes.MaintenanceStatusEnabled),
						}),
						"iceberg_snapshot_management": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"settings": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"max_snapshot_age_hours": knownvalue.Int32Exact(48),
								"min_snapshots_to_keep":  knownvalue.Int32Exact(1),
							}),
							names.AttrStatus: tfknownvalue.StringExact(awstypes.MaintenanceStatusEnabled),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccTableImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func testAccCheckTableDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3TablesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3tables_table" {
				continue
			}

			_, err := tfs3tables.FindTable(ctx, conn,
				rs.Primary.Attributes["table_bucket_arn"],
				rs.Primary.Attributes[names.AttrNamespace],
				rs.Primary.Attributes[names.AttrName],
			)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.S3Tables, create.ErrActionCheckingDestroyed, tfs3tables.ResNameTable, rs.Primary.ID, err)
			}

			return create.Error(names.S3Tables, create.ErrActionCheckingDestroyed, tfs3tables.ResNameTable, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckTableExists(ctx context.Context, name string, table *s3tables.GetTableOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.S3Tables, create.ErrActionCheckingExistence, tfs3tables.ResNameTable, name, errors.New("not found"))
		}

		if rs.Primary.Attributes["table_bucket_arn"] == "" || rs.Primary.Attributes[names.AttrNamespace] == "" || rs.Primary.Attributes[names.AttrName] == "" {
			return create.Error(names.S3Tables, create.ErrActionCheckingExistence, tfs3tables.ResNameTable, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3TablesClient(ctx)

		resp, err := tfs3tables.FindTable(ctx, conn,
			rs.Primary.Attributes["table_bucket_arn"],
			rs.Primary.Attributes[names.AttrNamespace],
			rs.Primary.Attributes[names.AttrName],
		)
		if err != nil {
			return create.Error(names.S3Tables, create.ErrActionCheckingExistence, tfs3tables.ResNameTable, rs.Primary.ID, err)
		}

		*table = *resp

		return nil
	}
}

func testAccTableImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		identifier := tfs3tables.TableIdentifier{
			TableBucketARN: rs.Primary.Attributes["table_bucket_arn"],
			Namespace:      rs.Primary.Attributes[names.AttrNamespace],
			Name:           rs.Primary.Attributes[names.AttrName],
		}

		return identifier.String(), nil
	}
}

func testAccTableConfig_basic(rName, namespace, bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3tables_table" "test" {
  name             = %[1]q
  namespace        = aws_s3tables_namespace.test.namespace
  table_bucket_arn = aws_s3tables_namespace.test.table_bucket_arn
  format           = "ICEBERG"
}

resource "aws_s3tables_namespace" "test" {
  namespace        = %[2]q
  table_bucket_arn = aws_s3tables_table_bucket.test.arn

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_s3tables_table_bucket" "test" {
  name = %[3]q
}
`, rName, namespace, bucketName)
}

func testAccTableConfig_maintenanceConfiguration(rName, namespace, bucketName string, targetSize, maxSnapshotAge, minSnapshots int32) string {
	return fmt.Sprintf(`
resource "aws_s3tables_table" "test" {
  name             = %[1]q
  namespace        = aws_s3tables_namespace.test.namespace
  table_bucket_arn = aws_s3tables_namespace.test.table_bucket_arn
  format           = "ICEBERG"

  maintenance_configuration = {
    iceberg_compaction = {
      settings = {
        target_file_size_mb = %[4]d
      }
      status = "enabled"
    }
    iceberg_snapshot_management = {
      settings = {
        max_snapshot_age_hours = %[5]d
        min_snapshots_to_keep  = %[6]d
      }
      status = "enabled"
    }
  }
}

resource "aws_s3tables_namespace" "test" {
  namespace        = %[2]q
  table_bucket_arn = aws_s3tables_table_bucket.test.arn

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_s3tables_table_bucket" "test" {
  name = %[3]q
}
`, rName, namespace, bucketName, targetSize, maxSnapshotAge, minSnapshots)
}
