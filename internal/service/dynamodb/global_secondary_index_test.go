// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package dynamodb_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdynamodb "github.com/hashicorp/terraform-provider-aws/internal/service/dynamodb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	warmThroughputOnDemandMixReadUnitsPerSecond  = 12_000
	warmThroughputOnDemandMixWriteUnitsPerSecond = 4_000
)

func TestAccDynamoDBGlobalSecondaryIndex_basic(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_basic(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PROVISIONED"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName, "index_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rNameTable),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("key_schema"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"attribute_name": knownvalue.StringExact(rNameTable),
							"attribute_type": knownvalue.StringExact("S"),
							"key_type":       knownvalue.StringExact("HASH"),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("projection"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"projection_type":    knownvalue.StringExact("ALL"),
							"non_key_attributes": knownvalue.Null(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"read_capacity_units":  knownvalue.Int64Exact(1),
							"write_capacity_units": knownvalue.Int64Exact(1),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(1),
						"write_units_per_second": knownvalue.Int64Exact(1),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_disappears(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_basic(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfdynamodb.ResourceGlobalSecondaryIndex, resourceName),
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

func TestAccDynamoDBGlobalSecondaryIndex_disappears_table(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_basic(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdynamodb.ResourceTable(), resourceNameTable),
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

func TestAccDynamoDBGlobalSecondaryIndex_provisioned_capacityChange(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	const (
		capacityUnitsInitial = 2
		capacityUnitsUpdated = 4
	)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_provisioned_withCapacity(rNameTable, rName, capacityUnitsInitial),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PROVISIONED"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"read_capacity_units":  knownvalue.Int64Exact(capacityUnitsInitial),
							"write_capacity_units": knownvalue.Int64Exact(capacityUnitsInitial),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(capacityUnitsInitial),
						"write_units_per_second": knownvalue.Int64Exact(capacityUnitsInitial),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_provisioned_withCapacity(rNameTable, rName, capacityUnitsUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PROVISIONED"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"read_capacity_units":  knownvalue.Int64Exact(capacityUnitsUpdated),
							"write_capacity_units": knownvalue.Int64Exact(capacityUnitsUpdated),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(capacityUnitsUpdated),
						"write_units_per_second": knownvalue.Int64Exact(capacityUnitsUpdated),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_provisioned_capacityChange_ignoreChanges(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_provisioned_withCapacityAndIgnoreChanges(rNameTable, rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PROVISIONED"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"read_capacity_units":  knownvalue.Int64Exact(2),
							"write_capacity_units": knownvalue.Int64Exact(2),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(2),
						"write_units_per_second": knownvalue.Int64Exact(2),
					})),
				},
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_provisioned_withCapacityAndIgnoreChanges(rNameTable, rName, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PROVISIONED"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"read_capacity_units":  knownvalue.Int64Exact(2),
							"write_capacity_units": knownvalue.Int64Exact(2),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(2),
						"write_units_per_second": knownvalue.Int64Exact(2),
					})),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_provisioned_changeTableCapacity_gsiSameAsTable(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_provisioned_capacity_gsiSameAsTable(rNameTable, rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PROVISIONED"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceNameTable, tfjsonpath.New("read_capacity"), knownvalue.Int64Exact(2)),
					statecheck.ExpectKnownValue(resourceNameTable, tfjsonpath.New("write_capacity"), knownvalue.Int64Exact(2)),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"read_capacity_units":  knownvalue.Int64Exact(2),
							"write_capacity_units": knownvalue.Int64Exact(2),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(2),
						"write_units_per_second": knownvalue.Int64Exact(2),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_provisioned_capacity_gsiSameAsTable(rNameTable, rName, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PROVISIONED"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceNameTable, tfjsonpath.New("read_capacity"), knownvalue.Int64Exact(4)),
					statecheck.ExpectKnownValue(resourceNameTable, tfjsonpath.New("write_capacity"), knownvalue.Int64Exact(4)),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"read_capacity_units":  knownvalue.Int64Exact(4),
							"write_capacity_units": knownvalue.Int64Exact(4),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(4),
						"write_units_per_second": knownvalue.Int64Exact(4),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_provisioned_changeTableCapacity_gsiDifferentFromTable(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_provisioned_capacity_gsiDifferentFromTable(rNameTable, rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PROVISIONED"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceNameTable, tfjsonpath.New("read_capacity"), knownvalue.Int64Exact(2)),
					statecheck.ExpectKnownValue(resourceNameTable, tfjsonpath.New("write_capacity"), knownvalue.Int64Exact(2)),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"read_capacity_units":  knownvalue.Int64Exact(4),
							"write_capacity_units": knownvalue.Int64Exact(4),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(4),
						"write_units_per_second": knownvalue.Int64Exact(4),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_provisioned_capacity_gsiDifferentFromTable(rNameTable, rName, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PROVISIONED"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceNameTable, tfjsonpath.New("read_capacity"), knownvalue.Int64Exact(4)),
					statecheck.ExpectKnownValue(resourceNameTable, tfjsonpath.New("write_capacity"), knownvalue.Int64Exact(4)),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"read_capacity_units":  knownvalue.Int64Exact(6),
							"write_capacity_units": knownvalue.Int64Exact(6),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(6),
						"write_units_per_second": knownvalue.Int64Exact(6),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_provisioned_warmThroughput_basic(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_provisioned_warmThroughput(rNameTable, rName, 1, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PROVISIONED"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName, "index_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rNameTable),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("projection"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"projection_type":    knownvalue.StringExact("ALL"),
							"non_key_attributes": knownvalue.Null(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"read_capacity_units":  knownvalue.Int64Exact(1),
							"write_capacity_units": knownvalue.Int64Exact(1),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(3),
						"write_units_per_second": knownvalue.Int64Exact(3),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_provisioned_warmThroughput_updateFromUnspecified(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_provisioned_warmThroughput_unspecified(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PROVISIONED"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"read_capacity_units":  knownvalue.Int64Exact(1),
							"write_capacity_units": knownvalue.Int64Exact(1),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(1),
						"write_units_per_second": knownvalue.Int64Exact(1),
					})),
				},
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_provisioned_warmThroughput(rNameTable, rName, 1, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PROVISIONED"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"read_capacity_units":  knownvalue.Int64Exact(1),
							"write_capacity_units": knownvalue.Int64Exact(1),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(3),
						"write_units_per_second": knownvalue.Int64Exact(3),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_provisioned_warmThroughput_updateFromUnspecified_noChange(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_provisioned_warmThroughput_unspecified(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PROVISIONED"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"read_capacity_units":  knownvalue.Int64Exact(1),
							"write_capacity_units": knownvalue.Int64Exact(1),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(1),
						"write_units_per_second": knownvalue.Int64Exact(1),
					})),
				},
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_provisioned_warmThroughput(rNameTable, rName, 1, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PROVISIONED"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"read_capacity_units":  knownvalue.Int64Exact(1),
							"write_capacity_units": knownvalue.Int64Exact(1),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(1),
						"write_units_per_second": knownvalue.Int64Exact(1),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_provisioned_throughputChanges(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	t.Parallel()

	const (
		resourceNameTable = "aws_dynamodb_table.test"
		resourceName      = "aws_dynamodb_global_secondary_index.test"
	)

	type config struct {
		provisionedCapacity int64
		warmCapacity        int64
	}
	testcases := map[string]struct {
		setup config
		apply config
	}{
		"change warm": {
			setup: config{
				provisionedCapacity: 1,
				warmCapacity:        1,
			},
			apply: config{
				provisionedCapacity: 1,
				warmCapacity:        3,
			},
		},
		"change provisioned no change warm, less than": {
			setup: config{
				provisionedCapacity: 1,
				warmCapacity:        3,
			},
			apply: config{
				provisionedCapacity: 2,
				warmCapacity:        3,
			},
		},
		"change provisioned change warm, less than": {
			setup: config{
				provisionedCapacity: 1,
				warmCapacity:        3,
			},
			apply: config{
				provisionedCapacity: 2,
				warmCapacity:        4,
			},
		},
		"change provisioned to match warm": {
			setup: config{
				provisionedCapacity: 1,
				warmCapacity:        3,
			},
			apply: config{
				provisionedCapacity: 3,
				warmCapacity:        3,
			},
		},
		"change both to same": {
			setup: config{
				provisionedCapacity: 1,
				warmCapacity:        2,
			},
			apply: config{
				provisionedCapacity: 3,
				warmCapacity:        3,
			},
		},
	}

	for name, testcase := range testcases { //nolint:paralleltest // false positive
		t.Run(name, func(t *testing.T) {
			ctx := acctest.Context(t)
			var conf awstypes.TableDescription
			var gsi awstypes.GlobalSecondaryIndexDescription

			rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
			rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

			acctest.ParallelTest(ctx, t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
				Steps: []resource.TestStep{
					{
						Config: testAccGlobalSecondaryIndexConfig_provisioned_warmThroughput(rNameTable, rName, testcase.setup.provisionedCapacity, testcase.setup.warmCapacity),
						Check: resource.ComposeAggregateTestCheckFunc(
							testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
							resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PROVISIONED"),

							testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
						),
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
							statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"read_capacity_units":  knownvalue.Int64Exact(testcase.setup.provisionedCapacity),
									"write_capacity_units": knownvalue.Int64Exact(testcase.setup.provisionedCapacity),
								}),
							})),
							statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
								"read_units_per_second":  knownvalue.Int64Exact(testcase.setup.warmCapacity),
								"write_units_per_second": knownvalue.Int64Exact(testcase.setup.warmCapacity),
							})),
						},
					},
					{
						Config: testAccGlobalSecondaryIndexConfig_provisioned_warmThroughput(rNameTable, rName, testcase.apply.provisionedCapacity, testcase.apply.warmCapacity),
						Check: resource.ComposeAggregateTestCheckFunc(
							testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
							resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PROVISIONED"),

							testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
						),
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
							statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"read_capacity_units":  knownvalue.Int64Exact(testcase.apply.provisionedCapacity),
									"write_capacity_units": knownvalue.Int64Exact(testcase.apply.provisionedCapacity),
								}),
							})),
							statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
								"read_units_per_second":  knownvalue.Int64Exact(testcase.apply.warmCapacity),
								"write_units_per_second": knownvalue.Int64Exact(testcase.apply.warmCapacity),
							})),
						},
					},
					{
						ResourceName:                         resourceName,
						ImportState:                          true,
						ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
						ImportStateVerify:                    true,
						ImportStateVerifyIdentifierAttribute: names.AttrARN,
					},
				},
			})
		})
	}
}

func TestAccDynamoDBGlobalSecondaryIndex_billingPayPerRequest_basic(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_basic(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PAY_PER_REQUEST"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(warmThroughputOnDemandMixReadUnitsPerSecond),
						"write_units_per_second": knownvalue.Int64Exact(warmThroughputOnDemandMixWriteUnitsPerSecond),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_billingPayPerRequest_onDemandThroughput(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughput(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PAY_PER_REQUEST"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"max_read_request_units":  knownvalue.Int64Exact(1),
							"max_write_request_units": knownvalue.Int64Exact(1),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(warmThroughputOnDemandMixReadUnitsPerSecond),
						"write_units_per_second": knownvalue.Int64Exact(warmThroughputOnDemandMixWriteUnitsPerSecond),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_billingPayPerRequest_onDemandThroughput_capacityChange(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughputWithCapacity(rNameTable, rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PAY_PER_REQUEST"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"max_read_request_units":  knownvalue.Int64Exact(2),
							"max_write_request_units": knownvalue.Int64Exact(2),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(warmThroughputOnDemandMixReadUnitsPerSecond),
						"write_units_per_second": knownvalue.Int64Exact(warmThroughputOnDemandMixWriteUnitsPerSecond),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughputWithCapacity(rNameTable, rName, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PAY_PER_REQUEST"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"max_read_request_units":  knownvalue.Int64Exact(4),
							"max_write_request_units": knownvalue.Int64Exact(4),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(warmThroughputOnDemandMixReadUnitsPerSecond),
						"write_units_per_second": knownvalue.Int64Exact(warmThroughputOnDemandMixWriteUnitsPerSecond),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_billingPayPerRequest_onDemandThroughput_capacityChange_ignoreChanges(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughputWithCapacityAndIgnoreChanges(rNameTable, rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PAY_PER_REQUEST"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"max_read_request_units":  knownvalue.Int64Exact(2),
							"max_write_request_units": knownvalue.Int64Exact(2),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(warmThroughputOnDemandMixReadUnitsPerSecond),
						"write_units_per_second": knownvalue.Int64Exact(warmThroughputOnDemandMixWriteUnitsPerSecond),
					})),
				},
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughputWithCapacityAndIgnoreChanges(rNameTable, rName, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PAY_PER_REQUEST"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"max_read_request_units":  knownvalue.Int64Exact(2),
							"max_write_request_units": knownvalue.Int64Exact(2),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(warmThroughputOnDemandMixReadUnitsPerSecond),
						"write_units_per_second": knownvalue.Int64Exact(warmThroughputOnDemandMixWriteUnitsPerSecond),
					})),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_billingPayPerRequest_onDemandThroughput_unset_read(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughput_both(rNameTable, rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PAY_PER_REQUEST"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"max_read_request_units":  knownvalue.Int64Exact(1),
							"max_write_request_units": knownvalue.Int64Exact(1),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(warmThroughputOnDemandMixReadUnitsPerSecond),
						"write_units_per_second": knownvalue.Int64Exact(warmThroughputOnDemandMixWriteUnitsPerSecond),
					})),
				},
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughput_writeUnitsOnly(rNameTable, rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PAY_PER_REQUEST"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"max_read_request_units":  knownvalue.Null(),
							"max_write_request_units": knownvalue.Int64Exact(2),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(warmThroughputOnDemandMixReadUnitsPerSecond),
						"write_units_per_second": knownvalue.Int64Exact(warmThroughputOnDemandMixWriteUnitsPerSecond),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_billingPayPerRequest_onDemandThroughput_unset_write(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughput_both(rNameTable, rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PAY_PER_REQUEST"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"max_read_request_units":  knownvalue.Int64Exact(1),
							"max_write_request_units": knownvalue.Int64Exact(1),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(warmThroughputOnDemandMixReadUnitsPerSecond),
						"write_units_per_second": knownvalue.Int64Exact(warmThroughputOnDemandMixWriteUnitsPerSecond),
					})),
				},
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughput_readUnitsOnly(rNameTable, rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PAY_PER_REQUEST"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"max_read_request_units":  knownvalue.Int64Exact(2),
							"max_write_request_units": knownvalue.Null(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(warmThroughputOnDemandMixReadUnitsPerSecond),
						"write_units_per_second": knownvalue.Int64Exact(warmThroughputOnDemandMixWriteUnitsPerSecond),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_billingPayPerRequest_warmThroughput_basic(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_warmThroughput(rNameTable, rName, 15000, 5000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PAY_PER_REQUEST"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(15000),
						"write_units_per_second": knownvalue.Int64Exact(5000),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateVerifyIgnore:              []string{"warm_throughput"},
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_billingPayPerRequest_warmThroughput_explicitDefault(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_warmThroughput(rNameTable, rName, 12000, 4000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PAY_PER_REQUEST"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(warmThroughputOnDemandMixReadUnitsPerSecond),
						"write_units_per_second": knownvalue.Int64Exact(warmThroughputOnDemandMixWriteUnitsPerSecond),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateVerifyIgnore:              []string{"warm_throughput"},
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_billingPayPerRequest_warmThroughput_updateFromUnspecified(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_warmThroughput_unspecified(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PAY_PER_REQUEST"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(warmThroughputOnDemandMixReadUnitsPerSecond),
						"write_units_per_second": knownvalue.Int64Exact(warmThroughputOnDemandMixWriteUnitsPerSecond),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_warmThroughput(rNameTable, rName, 15000, 5000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PAY_PER_REQUEST"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(15000),
						"write_units_per_second": knownvalue.Int64Exact(5000),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateVerifyIgnore:              []string{"warm_throughput"},
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_billingPayPerRequest_warmThroughput_updateFromUnspecified_noChange(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_warmThroughput_unspecified(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PAY_PER_REQUEST"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(warmThroughputOnDemandMixReadUnitsPerSecond),
						"write_units_per_second": knownvalue.Int64Exact(warmThroughputOnDemandMixWriteUnitsPerSecond),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_warmThroughput(rNameTable, rName, warmThroughputOnDemandMixReadUnitsPerSecond, warmThroughputOnDemandMixWriteUnitsPerSecond),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PAY_PER_REQUEST"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(warmThroughputOnDemandMixReadUnitsPerSecond),
						"write_units_per_second": knownvalue.Int64Exact(warmThroughputOnDemandMixWriteUnitsPerSecond),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateVerifyIgnore:              []string{"warm_throughput"},
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_payPerRequest_throughputChanges(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	t.Parallel()

	const (
		resourceNameTable = "aws_dynamodb_table.test"
		resourceName      = "aws_dynamodb_global_secondary_index.test"
	)

	type config struct {
		onDemandReadCapacity  int64
		onDemandWriteCapacity int64
		warmReadCapacity      int64
		warmWriteCapacity     int64
	}
	testcases := map[string]struct {
		setup config
		apply config
	}{
		"change warm": {
			setup: config{
				onDemandReadCapacity:  1,
				onDemandWriteCapacity: 1,
				warmReadCapacity:      warmThroughputOnDemandMixReadUnitsPerSecond,
				warmWriteCapacity:     warmThroughputOnDemandMixWriteUnitsPerSecond,
			},
			apply: config{
				onDemandReadCapacity:  1,
				onDemandWriteCapacity: 1,
				warmReadCapacity:      warmThroughputOnDemandMixReadUnitsPerSecond + 100,
				warmWriteCapacity:     warmThroughputOnDemandMixWriteUnitsPerSecond + 100,
			},
		},
		"change on-demand no change warm, less than": {
			setup: config{
				onDemandReadCapacity:  1,
				onDemandWriteCapacity: 1,
				warmReadCapacity:      warmThroughputOnDemandMixReadUnitsPerSecond,
				warmWriteCapacity:     warmThroughputOnDemandMixWriteUnitsPerSecond,
			},
			apply: config{
				onDemandReadCapacity:  2,
				onDemandWriteCapacity: 2,
				warmReadCapacity:      warmThroughputOnDemandMixReadUnitsPerSecond,
				warmWriteCapacity:     warmThroughputOnDemandMixWriteUnitsPerSecond,
			},
		},
		"change on-demand change warm, less than": {
			setup: config{
				onDemandReadCapacity:  1,
				onDemandWriteCapacity: 1,
				warmReadCapacity:      warmThroughputOnDemandMixReadUnitsPerSecond,
				warmWriteCapacity:     warmThroughputOnDemandMixWriteUnitsPerSecond,
			},
			apply: config{
				onDemandReadCapacity:  2,
				onDemandWriteCapacity: 2,
				warmReadCapacity:      warmThroughputOnDemandMixReadUnitsPerSecond + 100,
				warmWriteCapacity:     warmThroughputOnDemandMixWriteUnitsPerSecond + 100,
			},
		},
		"change on-demand to match warm": {
			setup: config{
				onDemandReadCapacity:  1,
				onDemandWriteCapacity: 1,
				warmReadCapacity:      warmThroughputOnDemandMixReadUnitsPerSecond,
				warmWriteCapacity:     warmThroughputOnDemandMixWriteUnitsPerSecond,
			},
			apply: config{
				onDemandReadCapacity:  warmThroughputOnDemandMixReadUnitsPerSecond,
				onDemandWriteCapacity: warmThroughputOnDemandMixWriteUnitsPerSecond,
				warmReadCapacity:      warmThroughputOnDemandMixReadUnitsPerSecond,
				warmWriteCapacity:     warmThroughputOnDemandMixWriteUnitsPerSecond,
			},
		},
		"change both to same": {
			setup: config{
				onDemandReadCapacity:  1,
				onDemandWriteCapacity: 1,
				warmReadCapacity:      warmThroughputOnDemandMixReadUnitsPerSecond,
				warmWriteCapacity:     warmThroughputOnDemandMixWriteUnitsPerSecond,
			},
			apply: config{
				onDemandReadCapacity:  warmThroughputOnDemandMixReadUnitsPerSecond + 100,
				onDemandWriteCapacity: warmThroughputOnDemandMixWriteUnitsPerSecond + 100,
				warmReadCapacity:      warmThroughputOnDemandMixReadUnitsPerSecond + 100,
				warmWriteCapacity:     warmThroughputOnDemandMixWriteUnitsPerSecond + 100,
			},
		},
	}

	for name, testcase := range testcases { //nolint:paralleltest // false positive
		t.Run(name, func(t *testing.T) {
			ctx := acctest.Context(t)
			var conf awstypes.TableDescription
			var gsi awstypes.GlobalSecondaryIndexDescription

			rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
			rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

			acctest.ParallelTest(ctx, t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
				Steps: []resource.TestStep{
					{
						Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_throughputChanges(rNameTable, rName, testcase.setup.onDemandReadCapacity, testcase.setup.onDemandWriteCapacity, testcase.setup.warmReadCapacity, testcase.setup.warmWriteCapacity),
						Check: resource.ComposeAggregateTestCheckFunc(
							testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
							resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PAY_PER_REQUEST"),

							testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
						),
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"max_read_request_units":  knownvalue.Int64Exact(testcase.setup.onDemandReadCapacity),
									"max_write_request_units": knownvalue.Int64Exact(testcase.setup.onDemandWriteCapacity),
								}),
							})),
							statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
							statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
								"read_units_per_second":  knownvalue.Int64Exact(testcase.setup.warmReadCapacity),
								"write_units_per_second": knownvalue.Int64Exact(testcase.setup.warmWriteCapacity),
							})),
						},
					},
					{
						Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_throughputChanges(rNameTable, rName, testcase.apply.onDemandReadCapacity, testcase.apply.onDemandWriteCapacity, testcase.apply.warmReadCapacity, testcase.apply.warmWriteCapacity),
						Check: resource.ComposeAggregateTestCheckFunc(
							testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
							resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PAY_PER_REQUEST"),

							testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
						),
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"max_read_request_units":  knownvalue.Int64Exact(testcase.apply.onDemandReadCapacity),
									"max_write_request_units": knownvalue.Int64Exact(testcase.apply.onDemandWriteCapacity),
								}),
							})),
							statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
							statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
								"read_units_per_second":  knownvalue.Int64Exact(testcase.apply.warmReadCapacity),
								"write_units_per_second": knownvalue.Int64Exact(testcase.apply.warmWriteCapacity),
							})),
						},
					},
					{
						ResourceName:                         resourceName,
						ImportState:                          true,
						ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
						ImportStateVerify:                    true,
						ImportStateVerifyIdentifierAttribute: names.AttrARN,
					},
				},
			})
		})
	}
}

func TestAccDynamoDBGlobalSecondaryIndex_validate_Attributes(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccGlobalSecondaryIndexConfig_validateAttribute_missmatchedType(rNameTable, rName),
				ExpectError: regexache.MustCompile(`Invalid Key Schema Type Change`),
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_validate_KeySchemas(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccGlobalSecondaryIndexConfig_validateAttribute_numberOfKeySchemas(rNameTable, rName, 0, 0),
				ExpectError: regexache.MustCompile(`Attribute key_schema must contain at least 1 and at most 4 elements with a\s+"key_type" of "HASH", got: 0`),
			},
			{
				Config:      testAccGlobalSecondaryIndexConfig_validateAttribute_numberOfKeySchemas(rNameTable, rName, 5, 0),
				ExpectError: regexache.MustCompile(`Attribute key_schema must contain at least 1 and at most 4 elements with a\s+"key_type" of "HASH", got: 5`),
			},
			{
				Config:      testAccGlobalSecondaryIndexConfig_validateAttribute_numberOfKeySchemas(rNameTable, rName, 1, 5),
				ExpectError: regexache.MustCompile(`Attribute key_schema must contain at most 4 elements with a "key_type" of\s+"RANGE", got: 5`),
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_payPerRequest_to_provisioned(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughputWithCapacity(rNameTable, rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PAY_PER_REQUEST"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"max_read_request_units":  knownvalue.Int64Exact(2),
							"max_write_request_units": knownvalue.Int64Exact(2),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(warmThroughputOnDemandMixReadUnitsPerSecond),
						"write_units_per_second": knownvalue.Int64Exact(warmThroughputOnDemandMixWriteUnitsPerSecond),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_provisioned_withCapacity(rNameTable, rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PROVISIONED"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"read_capacity_units":  knownvalue.Int64Exact(2),
							"write_capacity_units": knownvalue.Int64Exact(2),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(2),
						"write_units_per_second": knownvalue.Int64Exact(2),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_provisioned_to_payPerRequest_unspecifiedCapacity(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_provisioned_withCapacity(rNameTable, rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PROVISIONED"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"read_capacity_units":  knownvalue.Int64Exact(2),
							"write_capacity_units": knownvalue.Int64Exact(2),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(2),
						"write_units_per_second": knownvalue.Int64Exact(2),
					})),
				},
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_basic(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PAY_PER_REQUEST"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(warmThroughputOnDemandMixReadUnitsPerSecond),
						"write_units_per_second": knownvalue.Int64Exact(warmThroughputOnDemandMixWriteUnitsPerSecond),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_provisioned_to_payPerRequest_specifiedCapacity(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_provisioned_withCapacity(rNameTable, rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PROVISIONED"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"read_capacity_units":  knownvalue.Int64Exact(2),
							"write_capacity_units": knownvalue.Int64Exact(2),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(2),
						"write_units_per_second": knownvalue.Int64Exact(2),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughputWithCapacity(rNameTable, rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					resource.TestCheckResourceAttr(resourceNameTable, "billing_mode", "PAY_PER_REQUEST"),

					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"max_read_request_units":  knownvalue.Int64Exact(2),
							"max_write_request_units": knownvalue.Int64Exact(2),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(warmThroughputOnDemandMixReadUnitsPerSecond),
						"write_units_per_second": knownvalue.Int64Exact(warmThroughputOnDemandMixWriteUnitsPerSecond),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_keysNotOnTable_onCreate_hashOnly(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rHashKey := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Create GSI with hash key not on table
			{
				Config: testAccGlobalSecondaryIndexConfig_keysNotOnTable_hashOnly(rNameTable, rName, rHashKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("key_schema"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"attribute_name": knownvalue.StringExact(rHashKey),
							"attribute_type": knownvalue.StringExact("S"),
							"key_type":       knownvalue.StringExact("HASH"),
						}),
					})),
				},
			},
			// Step 1a: Import check
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			// Step 1b: Validate table after refresh
			{
				RefreshState: true,
				RefreshPlanChecks: resource.RefreshPlanChecks{
					PostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceNameTable, plancheck.ResourceActionNoop),
						plancheck.ExpectKnownValue(resourceNameTable, tfjsonpath.New("attribute"), knownvalue.SetExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrName: knownvalue.StringExact(rNameTable),
								names.AttrType: knownvalue.StringExact("S"),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrName: knownvalue.StringExact(rHashKey),
								names.AttrType: knownvalue.StringExact("S"),
							}),
						})),
					},
				},
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_keysNotOnTable_onCreate_hashAndSort(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rHashKey := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rSortKey := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Create GSI with keys not on table
			{
				Config: testAccGlobalSecondaryIndexConfig_keysNotOnTable_hashAndSort(rNameTable, rName, rHashKey, rSortKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("key_schema"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"attribute_name": knownvalue.StringExact(rHashKey),
							"attribute_type": knownvalue.StringExact("S"),
							"key_type":       knownvalue.StringExact("HASH"),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"attribute_name": knownvalue.StringExact(rSortKey),
							"attribute_type": knownvalue.StringExact("S"),
							"key_type":       knownvalue.StringExact("RANGE"),
						}),
					})),
				},
			},
			// Step 1a: Import check
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			// Step 1b: Validate table after refresh
			{
				RefreshState: true,
				RefreshPlanChecks: resource.RefreshPlanChecks{
					PostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceNameTable, plancheck.ResourceActionNoop),
						plancheck.ExpectKnownValue(resourceNameTable, tfjsonpath.New("attribute"), knownvalue.SetExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrName: knownvalue.StringExact(rNameTable),
								names.AttrType: knownvalue.StringExact("S"),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrName: knownvalue.StringExact(rHashKey),
								names.AttrType: knownvalue.StringExact("S"),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrName: knownvalue.StringExact(rSortKey),
								names.AttrType: knownvalue.StringExact("S"),
							}),
						})),
					},
				},
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_keysNotOnTable_onUpdate_hashOnly(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rHashKey1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rHashKey2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Create GSI with hash key same as table's hash key
			{
				Config: testAccGlobalSecondaryIndexConfig_keysNotOnTable_hashOnly(rNameTable, rName, rNameTable),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("key_schema"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"attribute_name": knownvalue.StringExact(rNameTable),
							"attribute_type": knownvalue.StringExact("S"),
							"key_type":       knownvalue.StringExact("HASH"),
						}),
					})),
				},
			},

			// Step 2: Modify GSI to use a new hash key not on the table
			{
				Config: testAccGlobalSecondaryIndexConfig_keysNotOnTable_hashOnly(rNameTable, rName, rHashKey1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("key_schema"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"attribute_name": knownvalue.StringExact(rHashKey1),
							"attribute_type": knownvalue.StringExact("S"),
							"key_type":       knownvalue.StringExact("HASH"),
						}),
					})),
				},
			},
			// Step 2a: Import check
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			// Step 2b: Validate table after refresh
			{
				RefreshState: true,
				RefreshPlanChecks: resource.RefreshPlanChecks{
					PostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceNameTable, plancheck.ResourceActionNoop),
						plancheck.ExpectKnownValue(resourceNameTable, tfjsonpath.New("attribute"), knownvalue.SetExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrName: knownvalue.StringExact(rNameTable),
								names.AttrType: knownvalue.StringExact("S"),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrName: knownvalue.StringExact(rHashKey1),
								names.AttrType: knownvalue.StringExact("S"),
							}),
						})),
					},
				},
			},

			// Step 3: Modify GSI to use a new hash key not on the table
			{
				Config: testAccGlobalSecondaryIndexConfig_keysNotOnTable_hashOnly(rNameTable, rName, rHashKey2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("key_schema"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"attribute_name": knownvalue.StringExact(rHashKey2),
							"attribute_type": knownvalue.StringExact("S"),
							"key_type":       knownvalue.StringExact("HASH"),
						}),
					})),
				},
			},
			// Step 3a: Import check
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			// Step 3b: Validate table after refresh
			{
				RefreshState: true,
				RefreshPlanChecks: resource.RefreshPlanChecks{
					PostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceNameTable, plancheck.ResourceActionNoop),
						plancheck.ExpectKnownValue(resourceNameTable, tfjsonpath.New("attribute"), knownvalue.SetExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrName: knownvalue.StringExact(rNameTable),
								names.AttrType: knownvalue.StringExact("S"),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrName: knownvalue.StringExact(rHashKey2),
								names.AttrType: knownvalue.StringExact("S"),
							}),
						})),
					},
				},
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_keysNotOnTable_onUpdate_hashAndSort(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rHashKey1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rSortKey1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rHashKey2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rSortKey2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Create GSI with hash key same as table's hash key
			{
				Config: testAccGlobalSecondaryIndexConfig_keysNotOnTable_hashOnly(rNameTable, rName, rNameTable),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("key_schema"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"attribute_name": knownvalue.StringExact(rNameTable),
							"attribute_type": knownvalue.StringExact("S"),
							"key_type":       knownvalue.StringExact("HASH"),
						}),
					})),
				},
			},

			// Step 2: Modify GSI to use new keys not on the table
			{
				Config: testAccGlobalSecondaryIndexConfig_keysNotOnTable_hashAndSort(rNameTable, rName, rHashKey1, rSortKey1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("key_schema"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"attribute_name": knownvalue.StringExact(rHashKey1),
							"attribute_type": knownvalue.StringExact("S"),
							"key_type":       knownvalue.StringExact("HASH"),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"attribute_name": knownvalue.StringExact(rSortKey1),
							"attribute_type": knownvalue.StringExact("S"),
							"key_type":       knownvalue.StringExact("RANGE"),
						}),
					})),
				},
			},
			// Step 2a: Import check
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			// Step 2b: Validate table after refresh
			{
				RefreshState: true,
				RefreshPlanChecks: resource.RefreshPlanChecks{
					PostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceNameTable, plancheck.ResourceActionNoop),
						plancheck.ExpectKnownValue(resourceNameTable, tfjsonpath.New("attribute"), knownvalue.SetExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrName: knownvalue.StringExact(rNameTable),
								names.AttrType: knownvalue.StringExact("S"),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrName: knownvalue.StringExact(rHashKey1),
								names.AttrType: knownvalue.StringExact("S"),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrName: knownvalue.StringExact(rSortKey1),
								names.AttrType: knownvalue.StringExact("S"),
							}),
						})),
					},
				},
			},

			// Step 3: Modify GSI to use new hash keys not on the table
			{
				Config: testAccGlobalSecondaryIndexConfig_keysNotOnTable_hashAndSort(rNameTable, rName, rHashKey2, rSortKey2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("key_schema"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"attribute_name": knownvalue.StringExact(rHashKey2),
							"attribute_type": knownvalue.StringExact("S"),
							"key_type":       knownvalue.StringExact("HASH"),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"attribute_name": knownvalue.StringExact(rSortKey2),
							"attribute_type": knownvalue.StringExact("S"),
							"key_type":       knownvalue.StringExact("RANGE"),
						}),
					})),
				},
			},
			// Step 3a: Import check
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			// Step 3b: Validate table after refresh
			{
				RefreshState: true,
				RefreshPlanChecks: resource.RefreshPlanChecks{
					PostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceNameTable, plancheck.ResourceActionNoop),
						plancheck.ExpectKnownValue(resourceNameTable, tfjsonpath.New("attribute"), knownvalue.SetExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrName: knownvalue.StringExact(rNameTable),
								names.AttrType: knownvalue.StringExact("S"),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrName: knownvalue.StringExact(rHashKey2),
								names.AttrType: knownvalue.StringExact("S"),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrName: knownvalue.StringExact(rSortKey2),
								names.AttrType: knownvalue.StringExact("S"),
							}),
						})),
					},
				},
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_nonKeyAttributes_onCreate(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_nonKeyAttributes(rNameTable, rName, []string{"test1", "test2"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("projection"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"projection_type": knownvalue.StringExact("INCLUDE"),
							"non_key_attributes": knownvalue.SetExact([]knownvalue.Check{
								knownvalue.StringExact("test1"),
								knownvalue.StringExact("test2"),
							}),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_nonKeyAttributes_onUpdate(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_nonKeyAttributes_updateSetup(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("projection"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"projection_type":    knownvalue.StringExact("ALL"),
							"non_key_attributes": knownvalue.Null(),
						}),
					})),
				},
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_nonKeyAttributes(rNameTable, rName, []string{"test1", "test2"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("projection"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"projection_type": knownvalue.StringExact("INCLUDE"),
							"non_key_attributes": knownvalue.SetExact([]knownvalue.Check{
								knownvalue.StringExact("test1"),
								knownvalue.StringExact("test2"),
							}),
						}),
					})),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_concurrentGSI_create(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi1 awstypes.GlobalSecondaryIndexDescription
	var gsi2 awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName1 := "aws_dynamodb_global_secondary_index.test1"
	resourceName2 := "aws_dynamodb_global_secondary_index.test2"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_multipleGsi_create(rNameTable, rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName1, &gsi1),
					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName2, &gsi2),
				),
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_concurrentGSI_update(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi1 awstypes.GlobalSecondaryIndexDescription
	var gsi2 awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName1 := "aws_dynamodb_global_secondary_index.test1"
	resourceName2 := "aws_dynamodb_global_secondary_index.test2"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_multipleGsi_update(rNameTable, rName1, rName2, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName1, &gsi1),
					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName2, &gsi2),
				),
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_multipleGsi_update(rNameTable, rName1, rName2, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName1, &gsi1),
					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName2, &gsi2),
				),
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_concurrentGSI_delete(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi1 awstypes.GlobalSecondaryIndexDescription
	var gsi2 awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName1 := "aws_dynamodb_global_secondary_index.test1"
	resourceName2 := "aws_dynamodb_global_secondary_index.test2"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_multipleGsi_create(rNameTable, rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName1, &gsi1),
					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName2, &gsi2),
				),
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_multipleGsi_tableOnly(rNameTable),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckGSINotExists(ctx, t, resourceName1),
					testAccCheckGSINotExists(ctx, t, resourceName2),
				),
			},
		},
	})
}

// lintignore:AT002
func TestAccDynamoDBGlobalSecondaryIndex_migrate_single_importcmd(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	expectTableGSINoChange := statecheck.CompareValue(compare.ValuesSame())

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_migrate_single_setup(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectTableGSINoChange.AddStateValue(resourceNameTable, tfjsonpath.New("global_secondary_index")),
				},
			},
			{
				Config:             testAccGlobalSecondaryIndexConfig_migrate_single(rNameTable, rName),
				ResourceName:       resourceName,
				ImportState:        true,
				ImportStateIdFunc:  testAccGlobalSecondaryIndexImportCmdIdFunc(rNameTable, rName),
				ImportStatePersist: true,
				ImportStateVerify:  false,
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_migrate_single(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectTableGSINoChange.AddStateValue(resourceNameTable, tfjsonpath.New("global_secondary_index")),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("key_schema"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"attribute_name": knownvalue.StringExact(rName),
							"attribute_type": knownvalue.StringExact("S"),
							"key_type":       knownvalue.StringExact("HASH"),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(1),
						"write_units_per_second": knownvalue.Int64Exact(1),
					})),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceNameTable, plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

// lintignore:AT002
func TestAccDynamoDBGlobalSecondaryIndex_migrate_single_importblock(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	expectTableGSINoChange := statecheck.CompareValue(compare.ValuesSame())

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_migrate_single_setup(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectTableGSINoChange.AddStateValue(resourceNameTable, tfjsonpath.New("global_secondary_index")),
				},
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_migrate_single_importblock(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName, &gsi),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectTableGSINoChange.AddStateValue(resourceNameTable, tfjsonpath.New("global_secondary_index")),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("key_schema"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"attribute_name": knownvalue.StringExact(rName),
							"attribute_type": knownvalue.StringExact("S"),
							"key_type":       knownvalue.StringExact("HASH"),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(1),
						"write_units_per_second": knownvalue.Int64Exact(1),
					})),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceNameTable, plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

// lintignore:AT002
func TestAccDynamoDBGlobalSecondaryIndex_migrate_multiple_importcmd(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName1 := "aws_dynamodb_global_secondary_index.test1"
	resourceName2 := "aws_dynamodb_global_secondary_index.test2"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	expectTableGSINoChange := statecheck.CompareValue(compare.ValuesSame())

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_migrate_multiple_setup(rNameTable, rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectTableGSINoChange.AddStateValue(resourceNameTable, tfjsonpath.New("global_secondary_index")),
				},
			},
			{
				Config:             testAccGlobalSecondaryIndexConfig_migrate_multiple(rNameTable, rName1, rName2),
				ResourceName:       resourceName1,
				ImportState:        true,
				ImportStateIdFunc:  testAccGlobalSecondaryIndexImportCmdIdFunc(rNameTable, rName1),
				ImportStatePersist: true,
				ImportStateVerify:  false,
			},
			{
				Config:             testAccGlobalSecondaryIndexConfig_migrate_multiple(rNameTable, rName1, rName2),
				ResourceName:       resourceName2,
				ImportState:        true,
				ImportStateIdFunc:  testAccGlobalSecondaryIndexImportCmdIdFunc(rNameTable, rName2),
				ImportStatePersist: true,
				ImportStateVerify:  false,
			},
			// nosemgrep:ci.semgrep.acctest.checks.replace-planonly-checks
			{
				Config:   testAccGlobalSecondaryIndexConfig_migrate_multiple(rNameTable, rName1, rName2),
				PlanOnly: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName1, plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction(resourceName2, plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction(resourceNameTable, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

// lintignore:AT002
func TestAccDynamoDBGlobalSecondaryIndex_migrate_multiple_importblock(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi1 awstypes.GlobalSecondaryIndexDescription
	var gsi2 awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName1 := "aws_dynamodb_global_secondary_index.test1"
	resourceName2 := "aws_dynamodb_global_secondary_index.test2"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	expectTableGSINoChange := statecheck.CompareValue(compare.ValuesSame())

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_migrate_multiple_setup(rNameTable, rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectTableGSINoChange.AddStateValue(resourceNameTable, tfjsonpath.New("global_secondary_index")),
				},
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_migrate_multiple_importblock(rNameTable, rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName1, &gsi1),
					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceName2, &gsi2),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectTableGSINoChange.AddStateValue(resourceNameTable, tfjsonpath.New("global_secondary_index")),

					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("key_schema"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"attribute_name": knownvalue.StringExact(rName1),
							"attribute_type": knownvalue.StringExact("S"),
							"key_type":       knownvalue.StringExact("HASH"),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(1),
						"write_units_per_second": knownvalue.Int64Exact(1),
					})),

					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("key_schema"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"attribute_name": knownvalue.StringExact(rName2),
							"attribute_type": knownvalue.StringExact("S"),
							"key_type":       knownvalue.StringExact("HASH"),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("warm_throughput"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"read_units_per_second":  knownvalue.Int64Exact(1),
						"write_units_per_second": knownvalue.Int64Exact(1),
					})),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceNameTable, plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction(resourceName1, plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction(resourceName2, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_migrate_partial(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar)

	ctx := acctest.Context(t)
	var conf awstypes.TableDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test1"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_migrate_multiple_setup(rNameTable, rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),

					resource.TestCheckResourceAttr(resourceNameTable, "global_secondary_index.#", "2"),
					resource.TestCheckResourceAttr(resourceNameTable, "attribute.#", "3"),
				),
			},
			{
				Config:             testAccGlobalSecondaryIndexConfig_migrate_partial(rNameTable, rName1, rName2),
				ResourceName:       resourceName,
				ImportState:        true,
				ImportStateIdFunc:  testAccGlobalSecondaryIndexImportCmdIdFunc(rNameTable, rName1),
				ImportStatePersist: true,
				ImportStateVerify:  false,
			},
			// nosemgrep:ci.semgrep.acctest.checks.replace-planonly-checks
			{
				Config:   testAccGlobalSecondaryIndexConfig_migrate_partial(rNameTable, rName1, rName2),
				PlanOnly: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),

					resource.TestCheckResourceAttr(resourceNameTable, "global_secondary_index.#", "1"),
					resource.TestCheckResourceAttr(resourceNameTable, "attribute.#", "2"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction(resourceNameTable, plancheck.ResourceActionUpdate),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_featureFlagNotSet(t *testing.T) {
	ctx := acctest.Context(t)

	// Explicitly clear `TF_AWS_EXPERIMENT_dynamodb_global_secondary_index`
	t.Setenv(tfdynamodb.GlobalSecondaryIndexExperimentalFlagEnvVar, "")

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccGlobalSecondaryIndexConfig_basic(rNameTable, rName),
				ExpectError: regexache.MustCompile(`Experimental Resource Type Not Enabled: "aws_dynamodb_global_secondary_index"`),
			},
		},
	})
}

func testAccCheckGlobalSecondaryIndexDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DynamoDBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dynamodb_global_secondary_index" {
				continue
			}

			tableName := rs.Primary.Attributes[names.AttrTableName]
			indexName := rs.Primary.Attributes["index_name"]

			_, err := tfdynamodb.FindGSIByTwoPartKey(ctx, conn, tableName, indexName)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("DynamoDB Global Secondary Index %s for Table %s still exists", indexName, tableName)
		}

		return nil
	}
}

func testAccCheckGlobalSecondaryIndexExists(ctx context.Context, t *testing.T, n string, gsi *awstypes.GlobalSecondaryIndexDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DynamoDBClient(ctx)
		tableName := rs.Primary.Attributes[names.AttrTableName]
		indexName := rs.Primary.Attributes["index_name"]

		output, err := tfdynamodb.FindGSIByTwoPartKey(ctx, conn, tableName, indexName)
		if err != nil {
			return err
		}

		*gsi = *output
		return nil
	}
}

func testAccCheckGSINotExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return nil
		}

		conn := acctest.ProviderMeta(ctx, t).DynamoDBClient(ctx)
		tableName := rs.Primary.Attributes[names.AttrTableName]
		indexName := rs.Primary.Attributes["index_name"]
		_, err := tfdynamodb.FindGSIByTwoPartKey(ctx, conn, tableName, indexName)
		if retry.NotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("Found: %s", n)
	}
}

func testAccGlobalSecondaryIndexImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		parts := []string{
			rs.Primary.Attributes[names.AttrTableName],
			rs.Primary.Attributes["index_name"],
		}

		return strings.Join(parts, intflex.ResourceIdSeparator), nil
	}
}

func testAccGlobalSecondaryIndexImportCmdIdFunc(tableName, indexName string) resource.ImportStateIdFunc {
	return func(_ *terraform.State) (string, error) {
		parts := []string{
			tableName,
			indexName,
		}

		return strings.Join(parts, intflex.ResourceIdSeparator), nil
	}
}

func testAccGlobalSecondaryIndexConfig_basic(tableName, indexName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type = "ALL"
  }

  provisioned_throughput {
    read_capacity_units  = 1
    write_capacity_units = 1
  }

  key_schema {
    attribute_name = %[1]q
    attribute_type = "S"
    key_type       = "HASH"
  }
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  hash_key       = %[1]q
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, tableName, indexName)
}

func testAccGlobalSecondaryIndexConfig_provisioned_withCapacity(tableName, indexName string, capacity int) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type = "ALL"
  }

  provisioned_throughput {
    read_capacity_units  = %[3]d
    write_capacity_units = %[3]d
  }

  key_schema {
    attribute_name = %[1]q
    attribute_type = "S"
    key_type       = "HASH"
  }
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 1
  hash_key       = %[1]q

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, tableName, indexName, capacity)
}

func testAccGlobalSecondaryIndexConfig_provisioned_withCapacityAndIgnoreChanges(tableName, indexName string, capacity int) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type = "ALL"
  }
  provisioned_throughput {
    read_capacity_units  = %[3]d
    write_capacity_units = %[3]d
  }

  key_schema {
    attribute_name = %[1]q
    attribute_type = "S"
    key_type       = "HASH"
  }

  key_schema {
    attribute_name = %[2]q
    attribute_type = "S"
    key_type       = "RANGE"
  }

  lifecycle {
    ignore_changes = [
      provisioned_throughput,
    ]
  }
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = %[3]d
  write_capacity = %[3]d
  hash_key       = %[1]q

  attribute {
    name = %[1]q
    type = "S"
  }

  lifecycle {
    ignore_changes = [
      read_capacity,
      write_capacity,
    ]
  }
}
`, tableName, indexName, capacity)
}

func testAccGlobalSecondaryIndexConfig_provisioned_capacity_gsiSameAsTable(tableName, indexName string, capacity int) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type = "ALL"
  }

  provisioned_throughput {
    read_capacity_units  = %[3]d
    write_capacity_units = %[3]d
  }

  key_schema {
    attribute_name = %[1]q
    attribute_type = "S"
    key_type       = "HASH"
  }
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = %[3]d
  write_capacity = %[3]d
  hash_key       = %[1]q

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, tableName, indexName, capacity)
}

func testAccGlobalSecondaryIndexConfig_provisioned_capacity_gsiDifferentFromTable(tableName, indexName string, capacity int) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type = "ALL"
  }

  provisioned_throughput {
    read_capacity_units  = %[4]d
    write_capacity_units = %[4]d
  }

  key_schema {
    attribute_name = %[1]q
    attribute_type = "S"
    key_type       = "HASH"
  }
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = %[3]d
  write_capacity = %[3]d
  hash_key       = %[1]q

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, tableName, indexName, capacity, capacity+2)
}

func testAccGlobalSecondaryIndexConfig_provisioned_warmThroughput(tableName, indexName string, provisionedCapacity, warmCapacity int64) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type = "ALL"
  }

  provisioned_throughput {
    read_capacity_units  = %[3]d
    write_capacity_units = %[3]d
  }

  warm_throughput = {
    read_units_per_second  = %[4]d
    write_units_per_second = %[4]d
  }

  key_schema {
    attribute_name = %[1]q
    attribute_type = "S"
    key_type       = "HASH"
  }
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  hash_key       = %[1]q
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, tableName, indexName, provisionedCapacity, warmCapacity)
}

func testAccGlobalSecondaryIndexConfig_provisioned_warmThroughput_unspecified(tableName, indexName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type = "ALL"
  }

  provisioned_throughput {
    read_capacity_units  = 1
    write_capacity_units = 1
  }

  key_schema {
    attribute_name = %[1]q
    attribute_type = "S"
    key_type       = "HASH"
  }
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  hash_key       = %[1]q
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, tableName, indexName)
}

func testAccGlobalSecondaryIndexConfig_billingPayPerRequest_basic(tableName, indexName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type = "ALL"
  }

  key_schema {
    attribute_name = %[1]q
    attribute_type = "S"
    key_type       = "HASH"
  }

  key_schema {
    attribute_name = %[2]q
    attribute_type = "S"
    key_type       = "RANGE"
  }
}

resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  hash_key     = %[1]q
  billing_mode = "PAY_PER_REQUEST"

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, tableName, indexName)
}

func testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughput(tableName, indexName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type = "ALL"
  }

  key_schema {
    attribute_name = %[1]q
    attribute_type = "S"
    key_type       = "HASH"
  }

  key_schema {
    attribute_name = %[2]q
    attribute_type = "S"
    key_type       = "RANGE"
  }

  on_demand_throughput {
    max_read_request_units  = 1
    max_write_request_units = 1
  }
}

resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  hash_key     = %[1]q
  billing_mode = "PAY_PER_REQUEST"

  on_demand_throughput {
    max_read_request_units  = 1
    max_write_request_units = 1
  }

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, tableName, indexName)
}

func testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughputWithCapacity(tableName, indexName string, capacity int) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type = "ALL"
  }

  key_schema {
    attribute_name = %[1]q
    attribute_type = "S"
    key_type       = "HASH"
  }

  key_schema {
    attribute_name = %[2]q
    attribute_type = "S"
    key_type       = "RANGE"
  }

  on_demand_throughput {
    max_read_request_units  = %[3]d
    max_write_request_units = %[3]d
  }
}

resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  hash_key     = %[1]q
  billing_mode = "PAY_PER_REQUEST"

  on_demand_throughput {
    max_read_request_units  = %[3]d
    max_write_request_units = %[3]d
  }

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, tableName, indexName, capacity)
}

func testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughputWithCapacityAndIgnoreChanges(tableName, indexName string, capacity int) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type = "ALL"
  }

  key_schema {
    attribute_name = %[1]q
    attribute_type = "S"
    key_type       = "HASH"
  }

  key_schema {
    attribute_name = %[2]q
    attribute_type = "S"
    key_type       = "RANGE"
  }

  on_demand_throughput {
    max_read_request_units  = %[3]d
    max_write_request_units = %[3]d
  }

  lifecycle {
    ignore_changes = [
      on_demand_throughput
    ]
  }
}

resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  hash_key     = %[1]q
  billing_mode = "PAY_PER_REQUEST"

  on_demand_throughput {
    max_read_request_units  = %[3]d
    max_write_request_units = %[3]d
  }

  attribute {
    name = %[1]q
    type = "S"
  }

  lifecycle {
    ignore_changes = [
      on_demand_throughput
    ]
  }
}
`, tableName, indexName, capacity)
}

func testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughput_both(tableName, indexName string, capacity int) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type = "ALL"
  }

  key_schema {
    attribute_name = %[1]q
    attribute_type = "S"
    key_type       = "HASH"
  }

  on_demand_throughput {
    max_read_request_units  = %[3]d
    max_write_request_units = %[3]d
  }
}

resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  hash_key     = %[1]q
  billing_mode = "PAY_PER_REQUEST"

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, tableName, indexName, capacity)
}

func testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughput_readUnitsOnly(tableName, indexName string, capacity int) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type = "ALL"
  }

  key_schema {
    attribute_name = %[1]q
    attribute_type = "S"
    key_type       = "HASH"
  }

  on_demand_throughput {
    max_read_request_units = %[3]d
  }
}

resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  hash_key     = %[1]q
  billing_mode = "PAY_PER_REQUEST"

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, tableName, indexName, capacity)
}

func testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughput_writeUnitsOnly(tableName, indexName string, capacity int) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type = "ALL"
  }

  key_schema {
    attribute_name = %[1]q
    attribute_type = "S"
    key_type       = "HASH"
  }

  on_demand_throughput {
    max_write_request_units = %[3]d
  }
}

resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  hash_key     = %[1]q
  billing_mode = "PAY_PER_REQUEST"

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, tableName, indexName, capacity)
}

func testAccGlobalSecondaryIndexConfig_billingPayPerRequest_warmThroughput(tableName, indexName string, readUnits, writeUnits int) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type = "ALL"
  }

  key_schema {
    attribute_name = %[1]q
    attribute_type = "S"
    key_type       = "HASH"
  }

  key_schema {
    attribute_name = %[2]q
    attribute_type = "S"
    key_type       = "RANGE"
  }

  warm_throughput = {
    read_units_per_second  = %[3]d
    write_units_per_second = %[4]d
  }
}

resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  hash_key     = %[1]q
  billing_mode = "PAY_PER_REQUEST"

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, tableName, indexName, readUnits, writeUnits)
}

func testAccGlobalSecondaryIndexConfig_billingPayPerRequest_warmThroughput_unspecified(tableName, indexName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type = "ALL"
  }

  key_schema {
    attribute_name = %[1]q
    attribute_type = "S"
    key_type       = "HASH"
  }

  key_schema {
    attribute_name = %[2]q
    attribute_type = "S"
    key_type       = "RANGE"
  }
}

resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  hash_key     = %[1]q
  billing_mode = "PAY_PER_REQUEST"

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, tableName, indexName)
}

func testAccGlobalSecondaryIndexConfig_billingPayPerRequest_throughputChanges(tableName, indexName string, onDemandReadUnits, onDemandWriteUnits, warmReadUnits, warmWriteUnits int64) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type = "ALL"
  }

  on_demand_throughput {
    max_read_request_units  = %[3]d
    max_write_request_units = %[4]d
  }

  warm_throughput = {
    read_units_per_second  = %[5]d
    write_units_per_second = %[6]d
  }

  key_schema {
    attribute_name = %[1]q
    attribute_type = "S"
    key_type       = "HASH"
  }

  key_schema {
    attribute_name = %[2]q
    attribute_type = "S"
    key_type       = "RANGE"
  }
}

resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  hash_key     = %[1]q
  billing_mode = "PAY_PER_REQUEST"

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, tableName, indexName, onDemandReadUnits, onDemandWriteUnits, warmReadUnits, warmWriteUnits)
}

func testAccGlobalSecondaryIndexConfig_validateAttribute_missmatchedType(tableName, indexName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type = "ALL"
  }

  key_schema {
    attribute_name = %[1]q
    attribute_type = "N"
    key_type       = "HASH"
  }

  on_demand_throughput {
    max_read_request_units  = 1
    max_write_request_units = 1
  }
}

resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  hash_key     = %[1]q
  billing_mode = "PAY_PER_REQUEST"

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, tableName, indexName)
}

func testAccGlobalSecondaryIndexConfig_validateAttribute_numberOfKeySchemas(tableName, indexName string, numHashes, numRanges int) string {
	hashes := ""
	ranges := ""

	for c := range numHashes {
		name := fmt.Sprintf("key%d", c)
		hashes += fmt.Sprintf("  key_schema {\n    attribute_name = %[1]q\n    attribute_type = \"S\"\n    key_type = \"HASH\"\n  }\n", name)
	}

	for c := range numRanges {
		name := fmt.Sprintf("key%d", c+numHashes)
		ranges += fmt.Sprintf("  key_schema {\n    attribute_name = %[1]q\n    attribute_type = \"S\"\n    key_type = \"RANGE\"\n  }\n", name)
	}

	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type = "ALL"
  }

  # hashes
  %[3]s

  # ranges
  %[4]s
}

resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  hash_key     = %[1]q
  billing_mode = "PAY_PER_REQUEST"

  on_demand_throughput {
    max_read_request_units  = 1
    max_write_request_units = 1
  }

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, tableName, indexName, hashes, ranges)
}

func testAccGlobalSecondaryIndexConfig_nonKeyAttributes(tableName, indexName string, nonKeyAttributes []string) string {
	nka := ""
	if len(nonKeyAttributes) > 0 {
		nka = strings.Join(nonKeyAttributes, `", "`)
	}

	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type    = "INCLUDE"
    non_key_attributes = ["%[3]s"]
  }

  provisioned_throughput {
    read_capacity_units  = 1
    write_capacity_units = 1
  }

  key_schema {
    attribute_name = %[1]q
    attribute_type = "S"
    key_type       = "HASH"
  }

}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  hash_key       = %[1]q
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, tableName, indexName, nka)
}

func testAccGlobalSecondaryIndexConfig_nonKeyAttributes_updateSetup(tableName, indexName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type = "ALL"
  }

  provisioned_throughput {
    read_capacity_units  = 1
    write_capacity_units = 1
  }

  key_schema {
    attribute_name = %[1]q
    attribute_type = "S"
    key_type       = "HASH"
  }
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  hash_key       = %[1]q
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, tableName, indexName)
}

func testAccGlobalSecondaryIndexConfig_keysNotOnTable_hashOnly(tableName, indexName, hashKey string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type = "ALL"
  }

  provisioned_throughput {
    read_capacity_units  = 1
    write_capacity_units = 1
  }

  key_schema {
    attribute_name = %[3]q
    attribute_type = "S"
    key_type       = "HASH"
  }
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  hash_key       = %[1]q
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, tableName, indexName, hashKey)
}

func testAccGlobalSecondaryIndexConfig_keysNotOnTable_hashAndSort(tableName, indexName, hashKey, sortKey string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type = "ALL"
  }

  provisioned_throughput {
    read_capacity_units  = 1
    write_capacity_units = 1
  }

  key_schema {
    attribute_name = %[3]q
    attribute_type = "S"
    key_type       = "HASH"
  }
  key_schema {
    attribute_name = %[4]q
    attribute_type = "S"
    key_type       = "RANGE"
  }
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  hash_key       = %[1]q
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, tableName, indexName, hashKey, sortKey)
}

func testAccGlobalSecondaryIndexConfig_multipleGsi_create(tableName, indexName1, indexName2 string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test1" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type = "ALL"
  }

  provisioned_throughput {
    read_capacity_units  = 1
    write_capacity_units = 1
  }

  key_schema {
    attribute_name = %[2]q
    attribute_type = "S"
    key_type       = "HASH"
  }
}

resource "aws_dynamodb_global_secondary_index" "test2" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[3]q
  projection {
    projection_type = "ALL"
  }

  provisioned_throughput {
    read_capacity_units  = 1
    write_capacity_units = 1
  }

  key_schema {
    attribute_name = %[3]q
    attribute_type = "S"
    key_type       = "HASH"
  }
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  hash_key       = %[1]q
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, tableName, indexName1, indexName2)
}

func testAccGlobalSecondaryIndexConfig_multipleGsi_tableOnly(tableName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  hash_key       = %[1]q
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, tableName)
}

func testAccGlobalSecondaryIndexConfig_multipleGsi_update(tableName, indexName1, indexName2 string, capacity int) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test1" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type = "ALL"
  }
  provisioned_throughput {
    read_capacity_units  = %[4]d
    write_capacity_units = %[4]d
  }

  key_schema {
    attribute_name = %[2]q
    attribute_type = "S"
    key_type       = "HASH"
  }
}

resource "aws_dynamodb_global_secondary_index" "test2" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[3]q
  projection {
    projection_type = "ALL"
  }
  provisioned_throughput {
    read_capacity_units  = %[4]d
    write_capacity_units = %[4]d
  }

  key_schema {
    attribute_name = %[3]q
    attribute_type = "S"
    key_type       = "HASH"
  }
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  hash_key       = %[1]q
  read_capacity  = %[4]d
  write_capacity = %[4]d

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, tableName, indexName1, indexName2, capacity)
}

func testAccGlobalSecondaryIndexConfig_migrate_single_setup(tableName, indexName1 string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  hash_key       = %[1]q
  read_capacity  = 1
  write_capacity = 1

  global_secondary_index {
    name            = %[2]q
    projection_type = "ALL"
    hash_key        = %[2]q
    read_capacity   = 1
    write_capacity  = 1
  }

  attribute {
    name = %[1]q
    type = "S"
  }
  attribute {
    name = %[2]q
    type = "S"
  }
}
`, tableName, indexName1)
}

func testAccGlobalSecondaryIndexConfig_migrate_single(tableName, indexName1 string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type = "ALL"
  }

  provisioned_throughput {
    read_capacity_units  = 1
    write_capacity_units = 1
  }

  key_schema {
    attribute_name = %[2]q
    attribute_type = "S"
    key_type       = "HASH"
  }
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  hash_key       = %[1]q
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, tableName, indexName1)
}

func testAccGlobalSecondaryIndexConfig_migrate_single_importblock(tableName, indexName string) string {
	return acctest.ConfigCompose(
		testAccGlobalSecondaryIndexConfig_migrate_single(tableName, indexName),
		fmt.Sprintf(`
import {
  to = aws_dynamodb_global_secondary_index.test
  id = "${aws_dynamodb_table.test.name},%[1]s"
}
`, indexName))
}

func testAccGlobalSecondaryIndexConfig_migrate_multiple_setup(tableName, indexName1, indexName2 string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  hash_key       = %[1]q
  read_capacity  = 1
  write_capacity = 1

  global_secondary_index {
    name            = %[2]q
    projection_type = "ALL"
    hash_key        = %[2]q
    read_capacity   = 1
    write_capacity  = 1
  }

  global_secondary_index {
    name            = %[3]q
    projection_type = "ALL"
    hash_key        = %[3]q
    read_capacity   = 1
    write_capacity  = 1
  }

  attribute {
    name = %[1]q
    type = "S"
  }

  attribute {
    name = %[2]q
    type = "S"
  }

  attribute {
    name = %[3]q
    type = "S"
  }
}
`, tableName, indexName1, indexName2)
}

func testAccGlobalSecondaryIndexConfig_migrate_multiple(tableName, indexName1, indexName2 string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test1" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type = "ALL"
  }

  provisioned_throughput {
    read_capacity_units  = 1
    write_capacity_units = 1
  }

  key_schema {
    attribute_name = %[2]q
    attribute_type = "S"
    key_type       = "HASH"
  }
}

resource "aws_dynamodb_global_secondary_index" "test2" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[3]q
  projection {
    projection_type = "ALL"
  }

  provisioned_throughput {
    read_capacity_units  = 1
    write_capacity_units = 1
  }

  key_schema {
    attribute_name = %[3]q
    attribute_type = "S"
    key_type       = "HASH"
  }
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  hash_key       = %[1]q
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, tableName, indexName1, indexName2)
}

func testAccGlobalSecondaryIndexConfig_migrate_multiple_importblock(tableName, indexName1, indexName2 string) string {
	return acctest.ConfigCompose(
		testAccGlobalSecondaryIndexConfig_migrate_multiple(tableName, indexName1, indexName2),
		fmt.Sprintf(`
import {
  to = aws_dynamodb_global_secondary_index.test1
  id = "${aws_dynamodb_table.test.name},%[1]s"
}

import {
  to = aws_dynamodb_global_secondary_index.test2
  id = "${aws_dynamodb_table.test.name},%[2]s"
}
`, indexName1, indexName2))
}

func testAccGlobalSecondaryIndexConfig_migrate_partial(tableName, indexName1, indexName2 string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test1" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
  projection {
    projection_type = "ALL"
  }

  provisioned_throughput {
    read_capacity_units  = 1
    write_capacity_units = 1
  }

  key_schema {
    attribute_name = %[2]q
    attribute_type = "S"
    key_type       = "HASH"
  }
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  hash_key       = %[1]q
  read_capacity  = 1
  write_capacity = 1

  global_secondary_index {
    name            = %[3]q
    projection_type = "ALL"
    hash_key        = %[3]q
    read_capacity   = 1
    write_capacity  = 1
  }

  attribute {
    name = %[1]q
    type = "S"
  }

  attribute {
    name = %[3]q
    type = "S"
  }
}
`, tableName, indexName1, indexName2)
}
