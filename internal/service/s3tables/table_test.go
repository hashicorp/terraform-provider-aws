// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3tables_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3tables/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfs3tables "github.com/hashicorp/terraform-provider-aws/internal/service/s3tables"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3TablesTable_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var table s3tables.GetTableOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bucketName := rName
	nsName := strings.ReplaceAll(rName, "-", "_")
	tableName := strings.ReplaceAll(rName, "-", "_")
	resourceName := "aws_s3tables_table.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_basic(tableName, nsName, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &table),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "s3tables", regexache.MustCompile("bucket/"+bucketName+"/table/"+verify.UUIDRegexPattern+"$")),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "created_by"),
					resource.TestCheckResourceAttr(resourceName, names.AttrFormat, "ICEBERG"),
					resource.TestCheckNoResourceAttr(resourceName, "metadata_location"),
					resource.TestCheckResourceAttrPair(resourceName, "modified_at", resourceName, names.AttrCreatedAt),
					resource.TestCheckNoResourceAttr(resourceName, "modified_by"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, tableName),
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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bucketName := rName
	nsName := strings.ReplaceAll(rName, "-", "_")
	tableName := strings.ReplaceAll(rName, "-", "_")
	resourceName := "aws_s3tables_table.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_basic(tableName, nsName, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &table),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfs3tables.ResourceTable, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccS3TablesTable_rename(t *testing.T) {
	ctx := acctest.Context(t)
	var table s3tables.GetTableOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bucketName := rName
	nsName := strings.ReplaceAll(rName, "-", "_")
	tableName := strings.ReplaceAll(rName, "-", "_")
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	tableNameUpdated := strings.ReplaceAll(rNameUpdated, "-", "_")
	resourceName := "aws_s3tables_table.test"

	createdAtNoChange := statecheck.CompareValue(compare.ValuesSame())
	createdByNoChange := statecheck.CompareValue(compare.ValuesSame())
	modifiedAtChange := statecheck.CompareValue(compare.ValuesDiffer())
	modifiedByChange := statecheck.CompareValue(compare.ValuesDiffer())
	versionNoChange := statecheck.CompareValue(compare.ValuesSame())
	warehouseLocationNoChange := statecheck.CompareValue(compare.ValuesSame())

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_basic(tableName, nsName, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, tableName),
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
				Config: testAccTableConfig_basic(tableNameUpdated, nsName, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, tableNameUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "modified_at"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "modified_by"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(tableNameUpdated)),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bucketName := rName
	nsName := strings.ReplaceAll(rName, "-", "_")
	tableName := strings.ReplaceAll(rName, "-", "_")
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	nsNameUpdated := strings.ReplaceAll(rNameUpdated, "-", "_")
	resourceName := "aws_s3tables_table.test"

	createdAtNoChange := statecheck.CompareValue(compare.ValuesSame())
	createdByNoChange := statecheck.CompareValue(compare.ValuesSame())
	modifiedAtChange := statecheck.CompareValue(compare.ValuesDiffer())
	modifiedByChange := statecheck.CompareValue(compare.ValuesDiffer())
	versionNoChange := statecheck.CompareValue(compare.ValuesSame())
	warehouseLocationNoChange := statecheck.CompareValue(compare.ValuesSame())

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_basic(tableName, nsName, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamespace, nsName),
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
				Config: testAccTableConfig_basic(tableName, nsNameUpdated, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamespace, nsNameUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "modified_at"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "modified_by"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrNamespace), knownvalue.StringExact(nsNameUpdated)),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bucketName := rName
	nsName := strings.ReplaceAll(rName, "-", "_")
	tableName := strings.ReplaceAll(rName, "-", "_")
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	nsNameUpdated := strings.ReplaceAll(rNameUpdated, "-", "_")
	tableNameUpdated := strings.ReplaceAll(rNameUpdated, "-", "_")
	resourceName := "aws_s3tables_table.test"

	createdAtNoChange := statecheck.CompareValue(compare.ValuesSame())
	createdByNoChange := statecheck.CompareValue(compare.ValuesSame())
	modifiedAtChange := statecheck.CompareValue(compare.ValuesDiffer())
	modifiedByChange := statecheck.CompareValue(compare.ValuesDiffer())
	versionNoChange := statecheck.CompareValue(compare.ValuesSame())
	warehouseLocationNoChange := statecheck.CompareValue(compare.ValuesSame())

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_basic(tableName, nsName, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, tableName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamespace, nsName),
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
				Config: testAccTableConfig_basic(tableNameUpdated, nsNameUpdated, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, tableNameUpdated),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamespace, nsNameUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "modified_at"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "modified_by"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(tableNameUpdated)),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrNamespace), knownvalue.StringExact(nsNameUpdated)),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bucketName := rName
	nsName := strings.ReplaceAll(rName, "-", "_")
	tableName := strings.ReplaceAll(rName, "-", "_")
	resourceName := "aws_s3tables_table.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_maintenanceConfiguration(tableName, nsName, bucketName, 64, 24, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &table),
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
				Config: testAccTableConfig_maintenanceConfiguration(tableName, nsName, bucketName, 128, 48, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &table),
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

func TestAccS3TablesTable_encryptionConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var table s3tables.GetTableOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bucketName := rName
	nsName := strings.ReplaceAll(rName, "-", "_")
	tableName := strings.ReplaceAll(rName, "-", "_")
	resourceName := "aws_s3tables_table.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_encryptionConfiguration(tableName, nsName, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &table),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "s3tables", regexache.MustCompile("bucket/"+bucketName+"/table/"+verify.UUIDRegexPattern+"$")),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "created_by"),
					resource.TestCheckResourceAttr(resourceName, names.AttrFormat, "ICEBERG"),
					resource.TestCheckNoResourceAttr(resourceName, "metadata_location"),
					resource.TestCheckResourceAttrPair(resourceName, "modified_at", resourceName, names.AttrCreatedAt),
					resource.TestCheckNoResourceAttr(resourceName, "modified_by"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, tableName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrNamespace, "aws_s3tables_namespace.test", names.AttrNamespace),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttrPair(resourceName, "table_bucket_arn", "aws_s3tables_table_bucket.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.TableTypeCustomer)),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_configuration.kms_key_arn", "aws_kms_key.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.sse_algorithm", "aws:kms"),
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

func TestAccS3TablesTable_metadata(t *testing.T) {
	ctx := acctest.Context(t)
	var table s3tables.GetTableOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bucketName := rName
	nsName := strings.ReplaceAll(rName, "-", "_")
	tableName := strings.ReplaceAll(rName, "-", "_")
	resourceName := "aws_s3tables_table.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_metadata(tableName, nsName, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "metadata.0.iceberg.0.schema.0.field.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "metadata.0.iceberg.0.schema.0.field.0.name", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "metadata.0.iceberg.0.schema.0.field.0.type", "int"),
					resource.TestCheckResourceAttr(resourceName, "metadata.0.iceberg.0.schema.0.field.0.required", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "metadata.0.iceberg.0.schema.0.field.1.name", acctest.CtName),
					resource.TestCheckResourceAttr(resourceName, "metadata.0.iceberg.0.schema.0.field.1.type", "string"),
					resource.TestCheckResourceAttr(resourceName, "metadata.0.iceberg.0.schema.0.field.1.required", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "metadata.0.iceberg.0.schema.0.field.2.name", names.AttrCreatedAt),
					resource.TestCheckResourceAttr(resourceName, "metadata.0.iceberg.0.schema.0.field.2.type", "timestamp"),
					resource.TestCheckResourceAttr(resourceName, "metadata.0.iceberg.0.schema.0.field.2.required", acctest.CtTrue),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccTableImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateVerifyIgnore:              []string{"metadata"},
			},
		},
	})
}

func testAccCheckTableDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).S3TablesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3tables_table" {
				continue
			}

			_, err := tfs3tables.FindTableByThreePartKey(ctx, conn,
				rs.Primary.Attributes["table_bucket_arn"],
				rs.Primary.Attributes[names.AttrNamespace],
				rs.Primary.Attributes[names.AttrName],
			)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Tables Table %s still exists", rs.Primary.Attributes[names.AttrName])
		}

		return nil
	}
}

func testAccCheckTableExists(ctx context.Context, t *testing.T, n string, v *s3tables.GetTableOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).S3TablesClient(ctx)

		output, err := tfs3tables.FindTableByThreePartKey(ctx, conn,
			rs.Primary.Attributes["table_bucket_arn"],
			rs.Primary.Attributes[names.AttrNamespace],
			rs.Primary.Attributes[names.AttrName],
		)

		if err != nil {
			return err
		}

		*v = *output

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

func testAccTableConfig_base(nsName, bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3tables_namespace" "test" {
  namespace        = %[1]q
  table_bucket_arn = aws_s3tables_table_bucket.test.arn

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_s3tables_table_bucket" "test" {
  name = %[2]q
}
`, nsName, bucketName)
}

func testAccTableConfig_basic(tableName, nsName, bucketName string) string {
	return acctest.ConfigCompose(testAccTableConfig_base(nsName, bucketName), fmt.Sprintf(`
resource "aws_s3tables_table" "test" {
  name             = %[1]q
  namespace        = aws_s3tables_namespace.test.namespace
  table_bucket_arn = aws_s3tables_namespace.test.table_bucket_arn
  format           = "ICEBERG"
}
`, tableName))
}

func testAccTableConfig_maintenanceConfiguration(tableName, nsName, bucketName string, targetSize, maxSnapshotAge, minSnapshots int32) string {
	return acctest.ConfigCompose(testAccTableConfig_base(nsName, bucketName), fmt.Sprintf(`
resource "aws_s3tables_table" "test" {
  name             = %[1]q
  namespace        = aws_s3tables_namespace.test.namespace
  table_bucket_arn = aws_s3tables_namespace.test.table_bucket_arn
  format           = "ICEBERG"

  maintenance_configuration = {
    iceberg_compaction = {
      settings = {
        target_file_size_mb = %[2]d
      }
      status = "enabled"
    }
    iceberg_snapshot_management = {
      settings = {
        max_snapshot_age_hours = %[3]d
        min_snapshots_to_keep  = %[4]d
      }
      status = "enabled"
    }
  }
}
`, tableName, targetSize, maxSnapshotAge, minSnapshots))
}

func testAccTableConfig_encryptionConfiguration(tableName, nsName, bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3tables_table" "test" {
  name             = %[1]q
  namespace        = aws_s3tables_namespace.test.namespace
  table_bucket_arn = aws_s3tables_namespace.test.table_bucket_arn
  format           = "ICEBERG"

  encryption_configuration = {
    kms_key_arn   = aws_kms_key.test.arn
    sse_algorithm = "aws:kms"
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
  encryption_configuration = {
    kms_key_arn   = aws_kms_key.test2.arn
    sse_algorithm = "aws:kms"
  }
}

resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  policy                  = data.aws_iam_policy_document.key_policy.json
}

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_iam_policy_document" "key_policy" {
  statement {
    sid = "EnableUserAccess"
    principals {
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
      type        = "AWS"
    }
    actions   = ["kms:*"]
    resources = ["*"]
  }
  statement {
    sid = "EnableMaintenace"
    principals {
      identifiers = ["maintenance.s3tables.amazonaws.com"]
      type        = "Service"
    }
    actions = [
      "kms:GenerateDataKey",
      "kms:Decrypt"
    ]
    resources = ["*"]
    condition {
      test     = "StringLike"
      values   = ["${aws_s3tables_table_bucket.test.arn}/*"]
      variable = "kms:EncryptionContext:aws:s3:arn"
    }
  }
}

resource "aws_kms_key" "test2" {
  deletion_window_in_days = 7
}
`, tableName, nsName, bucketName)
}

func testAccTableConfig_metadata(tableName, nsName, bucketName string) string {
	return acctest.ConfigCompose(testAccTableConfig_base(nsName, bucketName), fmt.Sprintf(`
resource "aws_s3tables_table" "test" {
  name             = %[1]q
  namespace        = aws_s3tables_namespace.test.namespace
  table_bucket_arn = aws_s3tables_namespace.test.table_bucket_arn
  format           = "ICEBERG"

  metadata {
    iceberg {
      schema {
        field {
          name     = "id"
          type     = "int"
          required = true
        }
        field {
          name = "name"
          type = "string"
        }
        field {
          name     = "created_at"
          type     = "timestamp"
          required = true
        }
      }
    }
  }
}
`, tableName))
}
