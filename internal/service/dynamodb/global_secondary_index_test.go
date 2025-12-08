// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdynamodb "github.com/hashicorp/terraform-provider-aws/internal/service/dynamodb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDynamoDBGlobalSecondaryIndex_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_basic(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckGSIExists(ctx, t, resourceName, &gsi),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName, "key_schema.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rNameTable,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckResourceAttr(resourceName, "index_name", rName),
					resource.TestCheckNoResourceAttr(resourceName, "non_key_attributes"),
					resource.TestCheckResourceAttr(resourceName, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "1"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
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

func TestAccDynamoDBGlobalSecondaryIndex_billingPayPerRequest(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckGSIExists(ctx, t, resourceName, &gsi),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName, "key_schema.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rNameTable,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rName,
						"attribute_type": "S",
						"key_type":       "RANGE",
					}),
					resource.TestCheckResourceAttr(resourceName, "index_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "0"),
				),
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

func TestAccDynamoDBGlobalSecondaryIndex_billingPayPerRequest_onDemandThroughput(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughput(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckGSIExists(ctx, t, resourceName, &gsi),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName, "key_schema.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rNameTable,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rName,
						"attribute_type": "S",
						"key_type":       "RANGE",
					}),
					resource.TestCheckResourceAttr(resourceName, "index_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "on_demand_throughput.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "on_demand_throughput.*", map[string]string{
						"max_read_request_units":  "1",
						"max_write_request_units": "1",
					}),
				),
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

func TestAccDynamoDBGlobalSecondaryIndex_billingPayPerRequest_capacityChange(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_basic_withCapacity(rNameTable, rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckGSIExists(ctx, t, resourceName, &gsi),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName, "key_schema.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rNameTable,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rName,
						"attribute_type": "S",
						"key_type":       "RANGE",
					}),
					resource.TestCheckResourceAttr(resourceName, "index_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "2"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_basic_withCapacity(rNameTable, rName, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckGSIExists(ctx, t, resourceName, &gsi),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName, "key_schema.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rNameTable,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rName,
						"attribute_type": "S",
						"key_type":       "RANGE",
					}),
					resource.TestCheckResourceAttr(resourceName, "index_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "4"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "4"),
				),
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

func TestAccDynamoDBGlobalSecondaryIndex_billingPayPerRequest_capacityChange_ignoreChanges(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_basic_withCapacityAndIgnoreChanges(rNameTable, rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckGSIExists(ctx, t, resourceName, &gsi),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName, "key_schema.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rNameTable,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rName,
						"attribute_type": "S",
						"key_type":       "RANGE",
					}),
					resource.TestCheckResourceAttr(resourceName, "index_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "2"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_basic_withCapacityAndIgnoreChanges(rNameTable, rName, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckGSIExists(ctx, t, resourceName, &gsi),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName, "key_schema.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rNameTable,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rName,
						"attribute_type": "S",
						"key_type":       "RANGE",
					}),
					resource.TestCheckResourceAttr(resourceName, "index_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "2"),
				),
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

func TestAccDynamoDBGlobalSecondaryIndex_billingPayPerRequest_onDemandThroughput_capacityChange(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughputWithCapacity(rNameTable, rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckGSIExists(ctx, t, resourceName, &gsi),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName, "key_schema.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rNameTable,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rName,
						"attribute_type": "S",
						"key_type":       "RANGE",
					}),
					resource.TestCheckResourceAttr(resourceName, "index_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "on_demand_throughput.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "on_demand_throughput.*", map[string]string{
						"max_read_request_units":  "2",
						"max_write_request_units": "2",
					}),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughputWithCapacity(rNameTable, rName, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckGSIExists(ctx, t, resourceName, &gsi),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName, "key_schema.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rNameTable,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rName,
						"attribute_type": "S",
						"key_type":       "RANGE",
					}),
					resource.TestCheckResourceAttr(resourceName, "index_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "on_demand_throughput.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "on_demand_throughput.*", map[string]string{
						"max_read_request_units":  "4",
						"max_write_request_units": "4",
					}),
				),
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

func TestAccDynamoDBGlobalSecondaryIndex_billingPayPerRequest_onDemandThroughput_capacityChange_ignoreChanges(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughputWithCapacityAndIgnoreChanges(rNameTable, rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckGSIExists(ctx, t, resourceName, &gsi),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName, "key_schema.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rNameTable,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rName,
						"attribute_type": "S",
						"key_type":       "RANGE",
					}),
					resource.TestCheckResourceAttr(resourceName, "index_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "on_demand_throughput.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "on_demand_throughput.*", map[string]string{
						"max_read_request_units":  "2",
						"max_write_request_units": "2",
					}),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughputWithCapacityAndIgnoreChanges(rNameTable, rName, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckGSIExists(ctx, t, resourceName, &gsi),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName, "key_schema.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rNameTable,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rName,
						"attribute_type": "S",
						"key_type":       "RANGE",
					}),
					resource.TestCheckResourceAttr(resourceName, "index_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "on_demand_throughput.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "on_demand_throughput.*", map[string]string{
						"max_read_request_units":  "2",
						"max_write_request_units": "2",
					}),
				),
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

func TestAccDynamoDBGlobalSecondaryIndex_billingPayPerRequest_warmThroughput(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_warmThroughput(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckGSIExists(ctx, t, resourceName, &gsi),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName, "key_schema.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rNameTable,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rName,
						"attribute_type": "S",
						"key_type":       "RANGE",
					}),
					resource.TestCheckResourceAttr(resourceName, "index_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "warm_throughput.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "warm_throughput.*", map[string]string{
						"read_units_per_second":  "15000",
						"write_units_per_second": "5000",
					}),
				),
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

func TestAccDynamoDBGlobalSecondaryIndex_attributeValidation(t *testing.T) {
	ctx := acctest.Context(t)

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccGlobalSecondaryIndexConfig_validateAttribute_missmatchedType(rNameTable, rName),
				ExpectError: regexache.MustCompile(`Changing already existing attribute`),
			},
			{
				Config:      testAccGlobalSecondaryIndexConfig_validateAttribute_numberOfKeySchemas(rNameTable, rName, 0, 0),
				ExpectError: regexache.MustCompile(`Unsupported number of hash keys`),
			},
			{
				Config:      testAccGlobalSecondaryIndexConfig_validateAttribute_numberOfKeySchemas(rNameTable, rName, 8, 0),
				ExpectError: regexache.MustCompile(`Unsupported number of hash keys`),
			},
			{
				Config:      testAccGlobalSecondaryIndexConfig_validateAttribute_numberOfKeySchemas(rNameTable, rName, 0, 8),
				ExpectError: regexache.MustCompile(`Unsupported number of range keys`),
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_payPerRequest_to_provisioned(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughputWithCapacity(rNameTable, rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckGSIExists(ctx, t, resourceName, &gsi),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName, "key_schema.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rNameTable,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rName,
						"attribute_type": "S",
						"key_type":       "RANGE",
					}),
					resource.TestCheckResourceAttr(resourceName, "index_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "on_demand_throughput.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "on_demand_throughput.*", map[string]string{
						"max_read_request_units":  "2",
						"max_write_request_units": "2",
					}),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_basic_withCapacity(rNameTable, rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckGSIExists(ctx, t, resourceName, &gsi),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName, "key_schema.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rNameTable,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rName,
						"attribute_type": "S",
						"key_type":       "RANGE",
					}),
					resource.TestCheckResourceAttr(resourceName, "index_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "2"),
				),
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

func TestAccDynamoDBGlobalSecondaryIndex_provisioned_to_payPerRequest(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_basic_withCapacity(rNameTable, rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckGSIExists(ctx, t, resourceName, &gsi),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName, "key_schema.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rNameTable,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rName,
						"attribute_type": "S",
						"key_type":       "RANGE",
					}),
					resource.TestCheckResourceAttr(resourceName, "index_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "2"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughputWithCapacity(rNameTable, rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckGSIExists(ctx, t, resourceName, &gsi),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName, "key_schema.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rNameTable,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rName,
						"attribute_type": "S",
						"key_type":       "RANGE",
					}),
					resource.TestCheckResourceAttr(resourceName, "index_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "on_demand_throughput.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "on_demand_throughput.0.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "on_demand_throughput.0.max_read_request_units", "2"),
					resource.TestCheckResourceAttr(resourceName, "on_demand_throughput.0.max_write_request_units", "2"),
				),
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

func TestAccDynamoDBGlobalSecondaryIndex_differentKeys(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rHashKey := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rRangeKey := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_differentKeys(rNameTable, rName, rNameTable, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckGSIExists(ctx, t, resourceName, &gsi),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName, "key_schema.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rNameTable,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckResourceAttr(resourceName, "index_name", rName),
					resource.TestCheckNoResourceAttr(resourceName, "non_key_attributes"),
					resource.TestCheckResourceAttr(resourceName, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "1"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
				},
			},
			// When using the new resource it is necessary to ignore the global_secondary_index blocks during tests
			// due to these blocks being still computed in the old aws_dynamodb_table resource.
			// During normal usage these do not produce diffs due to custom ignore logic.
			{
				ResourceName:      resourceNameTable,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"global_secondary_index",
				},
				ImportStateCheck: acctest.ComposeAggregateImportStateCheckFunc(
					acctest.ImportCheckResourceAttr(names.AttrName, rNameTable),
					acctest.ImportCheckResourceAttr("attribute.#", "1"),
					acctest.ImportCheckResourceAttr("attribute.0.%", "2"),
					acctest.ImportCheckResourceAttr("attribute.0.name", rNameTable),
					acctest.ImportCheckResourceAttr("attribute.0.type", "S"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_differentKeys(rNameTable, rName, rHashKey, rRangeKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckGSIExists(ctx, t, resourceName, &gsi),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName, "key_schema.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rHashKey,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rRangeKey,
						"attribute_type": "S",
						"key_type":       "RANGE",
					}),
					resource.TestCheckResourceAttr(resourceName, "index_name", rName),
					resource.TestCheckNoResourceAttr(resourceName, "non_key_attributes"),
					resource.TestCheckResourceAttr(resourceName, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "1"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
				},
			},
			{
				ResourceName:      resourceNameTable,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"attribute",
					"global_secondary_index",
				},
				ImportStateCheck: acctest.ComposeAggregateImportStateCheckFunc(
					acctest.ImportCheckResourceAttr(names.AttrName, rNameTable),
					acctest.ImportCheckResourceAttr("attribute.#", "3"),
					acctest.ImportCheckResourceAttr("attribute.0.%", "2"),
					acctest.ImportCheckResourceAttrSet("attribute.0.name"),
					acctest.ImportCheckResourceAttr("attribute.0.type", "S"),
					acctest.ImportCheckResourceAttr("attribute.1.%", "2"),
					acctest.ImportCheckResourceAttrSet("attribute.1.name"),
					acctest.ImportCheckResourceAttr("attribute.1.type", "S"),
					acctest.ImportCheckResourceAttr("attribute.2.%", "2"),
					acctest.ImportCheckResourceAttrSet("attribute.2.name"),
					acctest.ImportCheckResourceAttr("attribute.2.type", "S"),
				),
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

func TestAccDynamoDBGlobalSecondaryIndex_nonKeyAttributes(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_nonKeyAttributes(rNameTable, rName, []string{}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckGSIExists(ctx, t, resourceName, &gsi),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName, "key_schema.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rNameTable,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckResourceAttr(resourceName, "index_name", rName),
					resource.TestCheckNoResourceAttr(resourceName, "non_key_attributes"),
					resource.TestCheckResourceAttr(resourceName, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "1"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
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
				Config: testAccGlobalSecondaryIndexConfig_nonKeyAttributes(rNameTable, rName, []string{"test1", "test2"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckGSIExists(ctx, t, resourceName, &gsi),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName, "key_schema.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "key_schema.*", map[string]string{
						"attribute_name": rNameTable,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckResourceAttr(resourceName, "index_name", rName),
					resource.TestCheckResourceAttr(resourceName, "non_key_attributes.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "projection_type", "INCLUDE"),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "1"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("warm_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
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

func TestAccDynamoDBGlobalSecondaryIndex_multipleGsi_create(t *testing.T) {
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_multipleGsi_create(rNameTable, rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckGSIExists(ctx, t, resourceName1, &gsi1),
					testAccCheckGSIExists(ctx, t, resourceName2, &gsi2),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName1, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName1, "key_schema.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName1, "key_schema.*", map[string]string{
						"attribute_name": rName1,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckResourceAttr(resourceName1, "index_name", rName1),
					resource.TestCheckNoResourceAttr(resourceName1, "non_key_attributes"),
					resource.TestCheckResourceAttr(resourceName1, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName1, "read_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName1, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName1, "write_capacity", "1"),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName2, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName2, "key_schema.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName2, "key_schema.*", map[string]string{
						"attribute_name": rName2,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckResourceAttr(resourceName2, "index_name", rName2),
					resource.TestCheckNoResourceAttr(resourceName2, "non_key_attributes"),
					resource.TestCheckResourceAttr(resourceName2, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName2, "read_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName2, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName2, "write_capacity", "1"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("warm_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("warm_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
				},
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_multipleGsi_update(t *testing.T) {
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_multipleGsi_update(rNameTable, rName1, rName2, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckGSIExists(ctx, t, resourceName1, &gsi1),
					testAccCheckGSIExists(ctx, t, resourceName2, &gsi2),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName1, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName1, "key_schema.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName1, "key_schema.*", map[string]string{
						"attribute_name": rName1,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckResourceAttr(resourceName1, "index_name", rName1),
					resource.TestCheckNoResourceAttr(resourceName1, "non_key_attributes"),
					resource.TestCheckResourceAttr(resourceName1, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName1, "read_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName1, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName1, "write_capacity", "1"),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName2, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName2, "key_schema.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName2, "key_schema.*", map[string]string{
						"attribute_name": rName2,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckResourceAttr(resourceName2, "index_name", rName2),
					resource.TestCheckNoResourceAttr(resourceName2, "non_key_attributes"),
					resource.TestCheckResourceAttr(resourceName2, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName2, "read_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName2, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName2, "write_capacity", "1"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("warm_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("warm_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
				},
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_multipleGsi_update(rNameTable, rName1, rName2, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckGSIExists(ctx, t, resourceName1, &gsi1),
					testAccCheckGSIExists(ctx, t, resourceName2, &gsi2),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName1, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName1, "key_schema.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName1, "key_schema.*", map[string]string{
						"attribute_name": rName1,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckResourceAttr(resourceName1, "index_name", rName1),
					resource.TestCheckNoResourceAttr(resourceName1, "non_key_attributes"),
					resource.TestCheckResourceAttr(resourceName1, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName1, "read_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName1, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName1, "write_capacity", "2"),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName2, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName2, "key_schema.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName2, "key_schema.*", map[string]string{
						"attribute_name": rName2,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckResourceAttr(resourceName2, "index_name", rName2),
					resource.TestCheckNoResourceAttr(resourceName2, "non_key_attributes"),
					resource.TestCheckResourceAttr(resourceName2, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName2, "read_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName2, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName2, "write_capacity", "2"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("warm_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("warm_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
				},
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_multipleGsi_delete(t *testing.T) {
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_multipleGsi_create(rNameTable, rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckGSIExists(ctx, t, resourceName1, &gsi1),
					testAccCheckGSIExists(ctx, t, resourceName2, &gsi2),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName1, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName1, "key_schema.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName1, "key_schema.*", map[string]string{
						"attribute_name": rName1,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckResourceAttr(resourceName1, "index_name", rName1),
					resource.TestCheckNoResourceAttr(resourceName1, "non_key_attributes"),
					resource.TestCheckResourceAttr(resourceName1, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName1, "read_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName1, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName1, "write_capacity", "1"),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName2, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName2, "key_schema.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName2, "key_schema.*", map[string]string{
						"attribute_name": rName2,
						"attribute_type": "S",
						"key_type":       "HASH",
					}),
					resource.TestCheckResourceAttr(resourceName2, "index_name", rName2),
					resource.TestCheckNoResourceAttr(resourceName2, "non_key_attributes"),
					resource.TestCheckResourceAttr(resourceName2, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName2, "read_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName2, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName2, "write_capacity", "1"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("warm_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("on_demand_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("warm_throughput"), knownvalue.ListExact([]knownvalue.Check{})),
				},
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_multipleGsi_tableOnly(rNameTable),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckGSINotExists(ctx, t, resourceName1),
					testAccCheckGSINotExists(ctx, t, resourceName2),
				),
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_multipleGsi_badKeys(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName1 := "aws_dynamodb_global_secondary_index.test1"
	resourceName2 := "aws_dynamodb_global_secondary_index.test2"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_multipleGsi_tableOnly(rNameTable),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckGSINotExists(ctx, t, resourceName1),
					testAccCheckGSINotExists(ctx, t, resourceName2),
				),
			},
			{
				Config:      testAccGlobalSecondaryIndexConfig_multipleGsi_badKeys(rNameTable, rName1, rName2),
				ExpectError: regexache.MustCompile(`Changing already existing attribute`),
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_migrate_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName1 := "aws_dynamodb_global_secondary_index.test1"
	resourceName2 := "aws_dynamodb_global_secondary_index.test2"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_migrate_allOld(rNameTable, rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),

					resource.TestCheckResourceAttr(resourceNameTable, "global_secondary_index.#", "2"),
				),
			},
			{
				Config:       testAccGlobalSecondaryIndexConfig_multipleGsi_create(rNameTable, rName1, rName2),
				ResourceName: resourceName1,
				ImportState:  true,
				ImportStateIdFunc: func(t *awstypes.TableDescription, idxName string) resource.ImportStateIdFunc {
					return func(s *terraform.State) (string, error) {
						return fmt.Sprintf("%s/index/%s", *conf.TableArn, rName1), nil
					}
				}(&conf, rName1),
				ImportStatePersist: true,
				ImportStateVerify:  false,
			},
			{
				Config:       testAccGlobalSecondaryIndexConfig_multipleGsi_create(rNameTable, rName1, rName2),
				ResourceName: resourceName2,
				ImportState:  true,
				ImportStateIdFunc: func(t *awstypes.TableDescription, idxName string) resource.ImportStateIdFunc {
					return func(s *terraform.State) (string, error) {
						return fmt.Sprintf("%s/index/%s", *conf.TableArn, idxName), nil
					}
				}(&conf, rName2),
				ImportStatePersist: true,
				ImportStateVerify:  false,
			},
			// nosemgrep:ci.semgrep.acctest.checks.replace-planonly-checks
			{
				Config:   testAccGlobalSecondaryIndexConfig_multipleGsi_create(rNameTable, rName1, rName2),
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

func TestAccDynamoDBGlobalSecondaryIndex_migrate_partial(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_global_secondary_index.test1"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_migrate_allOld(rNameTable, rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),

					resource.TestCheckResourceAttr(resourceNameTable, "global_secondary_index.#", "2"),
					resource.TestCheckResourceAttr(resourceNameTable, "attribute.#", "3"),
				),
			},
			{
				Config:       testAccGlobalSecondaryIndexConfig_migrate_partial(rNameTable, rName1, rName2),
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/index/%s", *conf.TableArn, rName1), nil
				},
				ImportStatePersist: true,
				ImportStateVerify:  false,
			},
			// nosemgrep:ci.semgrep.acctest.checks.replace-planonly-checks
			{
				Config:   testAccGlobalSecondaryIndexConfig_migrate_partial(rNameTable, rName1, rName2),
				PlanOnly: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),

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

func testAccCheckGSIExists(ctx context.Context, t *testing.T, n string, gsi *awstypes.GlobalSecondaryIndexDescription) resource.TestCheckFunc {
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

func testAccGlobalSecondaryIndexConfig_basic(tableName, indexName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name      = aws_dynamodb_table.test.name
  index_name      = %[2]q
  projection_type = "ALL"

  key_schema {
    attribute_name = %[1]q
    attribute_type = "S"
    key_type = "HASH"
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

func testAccGlobalSecondaryIndexConfig_basic_withCapacity(tableName, indexName string, capacity int) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name      = aws_dynamodb_table.test.name
  index_name      = %[2]q
  projection_type = "ALL"
  read_capacity   = %[3]d
  write_capacity  = %[3]d

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

func testAccGlobalSecondaryIndexConfig_basic_withCapacityAndIgnoreChanges(tableName, indexName string, capacity int) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name      = aws_dynamodb_table.test.name
  index_name      = %[2]q
  projection_type = "ALL"
  read_capacity   = %[3]d
  write_capacity  = %[3]d

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
      read_capacity,
      write_capacity
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
      write_capacity
    ]
  }
}
`, tableName, indexName, capacity)
}

func testAccGlobalSecondaryIndexConfig_billingPayPerRequest(tableName, indexName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name      = aws_dynamodb_table.test.name
  index_name      = %[2]q
  projection_type = "ALL"

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
  table_name      = aws_dynamodb_table.test.name
  index_name      = %[2]q
  projection_type = "ALL"

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
  table_name      = aws_dynamodb_table.test.name
  index_name      = %[2]q
  projection_type = "ALL"

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
  table_name      = aws_dynamodb_table.test.name
  index_name      = %[2]q
  projection_type = "ALL"

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

func testAccGlobalSecondaryIndexConfig_billingPayPerRequest_warmThroughput(tableName, indexName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name      = aws_dynamodb_table.test.name
  index_name      = %[2]q
  projection_type = "ALL"

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

  warm_throughput {
    read_units_per_second  = 15000
    write_units_per_second = 5000
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

func testAccGlobalSecondaryIndexConfig_validateAttribute_missmatchedType(tableName, indexName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name      = aws_dynamodb_table.test.name
  index_name      = %[2]q
  projection_type = "ALL"

  key_schema {
    attribute_name = %[1]q
    attribute_type = "N"
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
  table_name      = aws_dynamodb_table.test.name
  index_name      = %[2]q
  projection_type = "ALL"

  # hashes
  %[3]s

  # ranges
  %[4]s

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
`, tableName, indexName, hashes, ranges)
}

func testAccGlobalSecondaryIndexConfig_nonKeyAttributes(tableName, indexName string, nonKeyAttributes []string) string {
	nka := ""
	projectionType := "ALL"
	if len(nonKeyAttributes) > 0 {
		nka = strings.Join(nonKeyAttributes, `", "`)
		nka = fmt.Sprintf(`non_key_attributes = ["%s"]`, nka)
		projectionType = "INCLUDE"
	}

	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name      = aws_dynamodb_table.test.name
  index_name      = %[2]q
  projection_type = %[3]q

  key_schema {
    attribute_name = %[1]q
    attribute_type = "S"
    key_type       = "HASH"
  }

  %[4]s
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
`, tableName, indexName, projectionType, nka)
}

func testAccGlobalSecondaryIndexConfig_differentKeys(tableName, indexName, pkGsi, skGsi string) string {
	if skGsi == "" {
		return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name      = aws_dynamodb_table.test.name
  index_name      = %[2]q
  projection_type = "ALL"

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
`, tableName, indexName, pkGsi)
	}

	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test" {
  table_name      = aws_dynamodb_table.test.name
  index_name      = %[2]q
  projection_type = "ALL"

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
`, tableName, indexName, pkGsi, skGsi)
}

func testAccGlobalSecondaryIndexConfig_multipleGsi_create(tableName, indexName1, indexName2 string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test1" {
  table_name      = aws_dynamodb_table.test.name
  index_name      = %[2]q
  projection_type = "ALL"

  key_schema {
    attribute_name = %[2]q
    attribute_type = "S"
    key_type       = "HASH"
  }
}

resource "aws_dynamodb_global_secondary_index" "test2" {
  table_name      = aws_dynamodb_table.test.name
  index_name      = %[3]q
  projection_type = "ALL"

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
  table_name      = aws_dynamodb_table.test.name
  index_name      = %[2]q
  projection_type = "ALL"
  read_capacity   = %[4]d
  write_capacity  = %[4]d

  key_schema {
    attribute_name = %[2]q
    attribute_type = "S"
    key_type       = "HASH"
  }
}
resource "aws_dynamodb_global_secondary_index" "test2" {
  table_name      = aws_dynamodb_table.test.name
  index_name      = %[3]q
  projection_type = "ALL"
  read_capacity   = %[4]d
  write_capacity  = %[4]d

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

func testAccGlobalSecondaryIndexConfig_multipleGsi_badKeys(tableName, indexName1, indexName2 string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test1" {
  table_name      = aws_dynamodb_table.test.name
  index_name      = %[2]q
  projection_type = "ALL"

  key_schema {
    attribute_name = %[1]q
    attribute_type = "B"
    key_type       = "HASH"
  }
}

resource "aws_dynamodb_global_secondary_index" "test2" {
  table_name      = aws_dynamodb_table.test.name
  index_name      = %[3]q
  projection_type = "ALL"

  key_schema {
    attribute_name = %[1]q
    attribute_type = "N"
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

func testAccGlobalSecondaryIndexConfig_migrate_allOld(tableName, indexName1, indexName2 string) string {
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

func testAccGlobalSecondaryIndexConfig_migrate_partial(tableName, indexName1, indexName2 string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_secondary_index" "test1" {
  table_name      = aws_dynamodb_table.test.name
  index_name      = %[2]q
  projection_type = "ALL"

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
