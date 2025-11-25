// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDynamoDBGlobalSecondaryIndex_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceNameGSI := "aws_dynamodb_global_secondary_index.test"

	rNameTable := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameGSI := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_basic(rNameTable, rNameGSI),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckInitialGSIExists(resourceNameGSI, &conf, &gsi),

					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key_type", "S"),
					func(s *terraform.State) error {
						return resource.TestCheckResourceAttr(resourceNameGSI, names.AttrARN, aws.ToString(gsi.IndexArn))(s)
					},
					resource.TestCheckResourceAttr(resourceNameGSI, names.AttrName, rNameGSI),
					resource.TestCheckNoResourceAttr(resourceNameGSI, "non_key_attributes"),
					resource.TestCheckResourceAttr(resourceNameGSI, "projection_type", "ALL"),
					resource.TestCheckNoResourceAttr(resourceNameGSI, "range_key"),
					resource.TestCheckNoResourceAttr(resourceNameGSI, "range_key_type"),
					resource.TestCheckResourceAttr(resourceNameGSI, "read_capacity", "1"),
					resource.TestCheckResourceAttr(resourceNameGSI, "table", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "write_capacity", "1"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.#", "0"),
					resource.TestCheckResourceAttr(resourceNameGSI, "warm_throughput.#", "0"),
				),
			},
			{
				ResourceName:                         resourceNameGSI,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateFunc(resourceNameGSI),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_billingPayPerRequest(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceNameGSI := "aws_dynamodb_global_secondary_index.test"

	rNameTable := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameGSI := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest(rNameTable, rNameGSI),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckInitialGSIExists(resourceNameGSI, &conf, &gsi),

					resource.TestCheckResourceAttr(resourceNameGSI, names.AttrName, rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "table", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key", rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceNameGSI, "read_capacity", "0"),
					resource.TestCheckResourceAttr(resourceNameGSI, "write_capacity", "0"),
				),
			},
			{
				ResourceName:                         resourceNameGSI,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateFunc(resourceNameGSI),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_billingPayPerRequest_onDemandThroughput(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceNameGSI := "aws_dynamodb_global_secondary_index.test"

	rNameTable := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameGSI := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughput(rNameTable, rNameGSI),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckInitialGSIExists(resourceNameGSI, &conf, &gsi),

					resource.TestCheckResourceAttr(resourceNameGSI, names.AttrName, rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "table", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key", rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceNameGSI, "read_capacity", "0"),
					resource.TestCheckResourceAttr(resourceNameGSI, "write_capacity", "0"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.#", "1"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.0.%", "2"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.0.max_read_request_units", "1"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.0.max_write_request_units", "1"),
				),
			},
			{
				ResourceName:                         resourceNameGSI,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateFunc(resourceNameGSI),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_billingPayPerRequest_capacityChange(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceNameGSI := "aws_dynamodb_global_secondary_index.test"

	rNameTable := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameGSI := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_basic_withCapacity(rNameTable, rNameGSI, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckInitialGSIExists(resourceNameGSI, &conf, &gsi),

					resource.TestCheckResourceAttr(resourceNameGSI, names.AttrName, rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "table", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key", rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceNameGSI, "read_capacity", "2"),
					resource.TestCheckResourceAttr(resourceNameGSI, "write_capacity", "2"),
				),
			},
			{
				ResourceName:                         resourceNameGSI,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateFunc(resourceNameGSI),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_basic_withCapacity(rNameTable, rNameGSI, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckInitialGSIExists(resourceNameGSI, &conf, &gsi),

					resource.TestCheckResourceAttr(resourceNameGSI, names.AttrName, rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "table", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key", rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceNameGSI, "read_capacity", "4"),
					resource.TestCheckResourceAttr(resourceNameGSI, "write_capacity", "4"),
				),
			},
			{
				ResourceName:                         resourceNameGSI,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateFunc(resourceNameGSI),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_billingPayPerRequest_capacityChange_ignoreChanges(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceNameGSI := "aws_dynamodb_global_secondary_index.test"

	rNameTable := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameGSI := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_basic_withCapacityAndIgnoreChanges(rNameTable, rNameGSI, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckInitialGSIExists(resourceNameGSI, &conf, &gsi),

					resource.TestCheckResourceAttr(resourceNameGSI, names.AttrName, rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "table", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key", rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceNameGSI, "read_capacity", "2"),
					resource.TestCheckResourceAttr(resourceNameGSI, "write_capacity", "2"),
				),
			},
			{
				ResourceName:                         resourceNameGSI,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateFunc(resourceNameGSI),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_basic_withCapacityAndIgnoreChanges(rNameTable, rNameGSI, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckInitialGSIExists(resourceNameGSI, &conf, &gsi),

					resource.TestCheckResourceAttr(resourceNameGSI, names.AttrName, rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "table", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key", rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceNameGSI, "read_capacity", "2"),
					resource.TestCheckResourceAttr(resourceNameGSI, "write_capacity", "2"),
				),
			},
			{
				ResourceName:                         resourceNameGSI,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateFunc(resourceNameGSI),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_billingPayPerRequest_onDemandThroughput_capacityChange(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceNameGSI := "aws_dynamodb_global_secondary_index.test"

	rNameTable := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameGSI := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughputWithCapacity(rNameTable, rNameGSI, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckInitialGSIExists(resourceNameGSI, &conf, &gsi),

					resource.TestCheckResourceAttr(resourceNameGSI, names.AttrName, rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "table", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key", rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceNameGSI, "read_capacity", "0"),
					resource.TestCheckResourceAttr(resourceNameGSI, "write_capacity", "0"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.#", "1"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.0.%", "2"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.0.max_read_request_units", "2"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.0.max_write_request_units", "2"),
				),
			},
			{
				ResourceName:                         resourceNameGSI,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateFunc(resourceNameGSI),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughputWithCapacity(rNameTable, rNameGSI, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckInitialGSIExists(resourceNameGSI, &conf, &gsi),

					resource.TestCheckResourceAttr(resourceNameGSI, names.AttrName, rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "table", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key", rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceNameGSI, "read_capacity", "0"),
					resource.TestCheckResourceAttr(resourceNameGSI, "write_capacity", "0"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.#", "1"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.0.%", "2"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.0.max_read_request_units", "4"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.0.max_write_request_units", "4"),
				),
			},
			{
				ResourceName:                         resourceNameGSI,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateFunc(resourceNameGSI),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_billingPayPerRequest_onDemandThroughput_capacityChange_ignoreChanges(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceNameGSI := "aws_dynamodb_global_secondary_index.test"

	rNameTable := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameGSI := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughputWithCapacityAndIgnoreChanges(rNameTable, rNameGSI, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckInitialGSIExists(resourceNameGSI, &conf, &gsi),

					resource.TestCheckResourceAttr(resourceNameGSI, names.AttrName, rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "table", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key", rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceNameGSI, "read_capacity", "0"),
					resource.TestCheckResourceAttr(resourceNameGSI, "write_capacity", "0"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.#", "1"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.0.%", "2"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.0.max_read_request_units", "2"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.0.max_write_request_units", "2"),
				),
			},
			{
				ResourceName:                         resourceNameGSI,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateFunc(resourceNameGSI),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughputWithCapacityAndIgnoreChanges(rNameTable, rNameGSI, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckInitialGSIExists(resourceNameGSI, &conf, &gsi),

					resource.TestCheckResourceAttr(resourceNameGSI, names.AttrName, rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "table", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key", rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceNameGSI, "read_capacity", "0"),
					resource.TestCheckResourceAttr(resourceNameGSI, "write_capacity", "0"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.#", "1"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.0.%", "2"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.0.max_read_request_units", "2"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.0.max_write_request_units", "2"),
				),
			},
			{
				ResourceName:                         resourceNameGSI,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateFunc(resourceNameGSI),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_billingPayPerRequest_warmThroughput(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceNameGSI := "aws_dynamodb_global_secondary_index.test"

	rNameTable := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameGSI := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_warmThroughput(rNameTable, rNameGSI),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckInitialGSIExists(resourceNameGSI, &conf, &gsi),

					resource.TestCheckResourceAttr(resourceNameGSI, names.AttrName, rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "table", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key", rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceNameGSI, "read_capacity", "0"),
					resource.TestCheckResourceAttr(resourceNameGSI, "write_capacity", "0"),
					resource.TestCheckResourceAttr(resourceNameGSI, "warm_throughput.#", "1"),
					resource.TestCheckResourceAttr(resourceNameGSI, "warm_throughput.0.%", "2"),
					resource.TestCheckResourceAttr(resourceNameGSI, "warm_throughput.0.read_units_per_second", "15000"),
					resource.TestCheckResourceAttr(resourceNameGSI, "warm_throughput.0.write_units_per_second", "5000"),
				),
			},
			{
				ResourceName:                         resourceNameGSI,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateFunc(resourceNameGSI),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_attributeValidation(t *testing.T) {
	ctx := acctest.Context(t)

	rNameTable := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameGSI := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccGlobalSecondaryIndexConfig_validateAttribute_missmatchedType(rNameTable, rNameGSI),
				ExpectError: regexache.MustCompile(`Changing already existing attribute`),
			},
			{
				Config:      testAccGlobalSecondaryIndexConfig_validateAttribute_unknownType(rNameTable, rNameGSI, "hash_key"),
				ExpectError: regexache.MustCompile(`"hash_key_type" must be set in this configuration`),
			},
			{
				Config:      testAccGlobalSecondaryIndexConfig_validateAttribute_unknownType(rNameTable, rNameGSI, "range_key"),
				ExpectError: regexache.MustCompile(`The argument "hash_key" is required, but no definition was found.`),
			},
			{
				Config:      testAccGlobalSecondaryIndexConfig_validateAttribute_unknownType(rNameTable, rNameGSI, fmt.Sprintf("hash_key = %q\nrange_key", rNameTable)),
				ExpectError: regexache.MustCompile(`"range_key_type" must be set in this configuration`),
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_payPerRequest_to_provisioned(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceNameGSI := "aws_dynamodb_global_secondary_index.test"

	rNameTable := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameGSI := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughputWithCapacity(rNameTable, rNameGSI, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckInitialGSIExists(resourceNameGSI, &conf, &gsi),

					resource.TestCheckResourceAttr(resourceNameGSI, names.AttrName, rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "table", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key", rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceNameGSI, "read_capacity", "0"),
					resource.TestCheckResourceAttr(resourceNameGSI, "write_capacity", "0"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.#", "1"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.0.%", "2"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.0.max_read_request_units", "2"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.0.max_write_request_units", "2"),
				),
			},
			{
				ResourceName:                         resourceNameGSI,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateFunc(resourceNameGSI),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_basic_withCapacity(rNameTable, rNameGSI, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckInitialGSIExists(resourceNameGSI, &conf, &gsi),

					resource.TestCheckResourceAttr(resourceNameGSI, names.AttrName, rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "table", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key", rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceNameGSI, "read_capacity", "2"),
					resource.TestCheckResourceAttr(resourceNameGSI, "write_capacity", "2"),
				),
			},
			{
				ResourceName:                         resourceNameGSI,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateFunc(resourceNameGSI),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_provisioned_to_payPerRequest(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceNameGSI := "aws_dynamodb_global_secondary_index.test"

	rNameTable := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameGSI := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_basic_withCapacity(rNameTable, rNameGSI, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckInitialGSIExists(resourceNameGSI, &conf, &gsi),

					resource.TestCheckResourceAttr(resourceNameGSI, names.AttrName, rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "table", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key", rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceNameGSI, "read_capacity", "2"),
					resource.TestCheckResourceAttr(resourceNameGSI, "write_capacity", "2"),
				),
			},
			{
				ResourceName:                         resourceNameGSI,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateFunc(resourceNameGSI),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughputWithCapacity(rNameTable, rNameGSI, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckInitialGSIExists(resourceNameGSI, &conf, &gsi),

					resource.TestCheckResourceAttr(resourceNameGSI, names.AttrName, rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "table", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key", rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceNameGSI, "read_capacity", "0"),
					resource.TestCheckResourceAttr(resourceNameGSI, "write_capacity", "0"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.#", "1"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.0.%", "2"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.0.max_read_request_units", "2"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.0.max_write_request_units", "2"),
				),
			},
			{
				ResourceName:                         resourceNameGSI,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateFunc(resourceNameGSI),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_differentKeys(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceNameGSI := "aws_dynamodb_global_secondary_index.test"

	rNameTable := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameGSI := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rHashKey := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rRangeKey := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_differentKeys(rNameTable, rNameGSI, rNameTable, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckInitialGSIExists(resourceNameGSI, &conf, &gsi),

					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key_type", "S"),
					func(s *terraform.State) error {
						return resource.TestCheckResourceAttr(resourceNameGSI, names.AttrARN, aws.ToString(gsi.IndexArn))(s)
					},
					resource.TestCheckResourceAttr(resourceNameGSI, names.AttrName, rNameGSI),
					resource.TestCheckNoResourceAttr(resourceNameGSI, "non_key_attributes"),
					resource.TestCheckResourceAttr(resourceNameGSI, "projection_type", "ALL"),
					resource.TestCheckNoResourceAttr(resourceNameGSI, "range_key"),
					resource.TestCheckNoResourceAttr(resourceNameGSI, "range_key_type"),
					resource.TestCheckResourceAttr(resourceNameGSI, "read_capacity", "1"),
					resource.TestCheckResourceAttr(resourceNameGSI, "table", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "write_capacity", "1"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.#", "0"),
					resource.TestCheckResourceAttr(resourceNameGSI, "warm_throughput.#", "0"),
				),
			},
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
				),
			},
			{
				ResourceName:                         resourceNameGSI,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateFunc(resourceNameGSI),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_differentKeys(rNameTable, rNameGSI, rHashKey, rRangeKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckInitialGSIExists(resourceNameGSI, &conf, &gsi),

					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key", rHashKey),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key_type", "S"),
					func(s *terraform.State) error {
						return resource.TestCheckResourceAttr(resourceNameGSI, names.AttrARN, aws.ToString(gsi.IndexArn))(s)
					},
					resource.TestCheckResourceAttr(resourceNameGSI, names.AttrName, rNameGSI),
					resource.TestCheckNoResourceAttr(resourceNameGSI, "non_key_attributes"),
					resource.TestCheckResourceAttr(resourceNameGSI, "projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key", rRangeKey),
					resource.TestCheckResourceAttr(resourceNameGSI, "range_key_type", "S"),
					resource.TestCheckResourceAttr(resourceNameGSI, "read_capacity", "1"),
					resource.TestCheckResourceAttr(resourceNameGSI, "table", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "write_capacity", "1"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.#", "0"),
					resource.TestCheckResourceAttr(resourceNameGSI, "warm_throughput.#", "0"),
				),
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
				),
			},
			{
				ResourceName:                         resourceNameGSI,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateFunc(resourceNameGSI),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccDynamoDBGlobalSecondaryIndex_nonKeyAttributes(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceNameGSI := "aws_dynamodb_global_secondary_index.test"

	rNameTable := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameGSI := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSecondaryIndexConfig_nonKeyAttributes(rNameTable, rNameGSI, []string{}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckInitialGSIExists(resourceNameGSI, &conf, &gsi),

					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key_type", "S"),
					func(s *terraform.State) error {
						return resource.TestCheckResourceAttr(resourceNameGSI, names.AttrARN, aws.ToString(gsi.IndexArn))(s)
					},
					resource.TestCheckResourceAttr(resourceNameGSI, names.AttrName, rNameGSI),
					resource.TestCheckNoResourceAttr(resourceNameGSI, "non_key_attributes"),
					resource.TestCheckResourceAttr(resourceNameGSI, "projection_type", "ALL"),
					resource.TestCheckNoResourceAttr(resourceNameGSI, "range_key"),
					resource.TestCheckNoResourceAttr(resourceNameGSI, "range_key_type"),
					resource.TestCheckResourceAttr(resourceNameGSI, "read_capacity", "1"),
					resource.TestCheckResourceAttr(resourceNameGSI, "table", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "write_capacity", "1"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.#", "0"),
					resource.TestCheckResourceAttr(resourceNameGSI, "warm_throughput.#", "0"),
				),
			},
			{
				ResourceName:                         resourceNameGSI,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateFunc(resourceNameGSI),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
			{
				Config: testAccGlobalSecondaryIndexConfig_nonKeyAttributes(rNameTable, rNameGSI, []string{"test1", "test2"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceNameTable, &conf),
					testAccCheckInitialGSIExists(resourceNameGSI, &conf, &gsi),

					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "hash_key_type", "S"),
					func(s *terraform.State) error {
						return resource.TestCheckResourceAttr(resourceNameGSI, names.AttrARN, aws.ToString(gsi.IndexArn))(s)
					},
					resource.TestCheckResourceAttr(resourceNameGSI, names.AttrName, rNameGSI),
					resource.TestCheckResourceAttr(resourceNameGSI, "non_key_attributes.#", "2"),
					resource.TestCheckResourceAttr(resourceNameGSI, "projection_type", "INCLUDE"),
					resource.TestCheckNoResourceAttr(resourceNameGSI, "range_key"),
					resource.TestCheckNoResourceAttr(resourceNameGSI, "range_key_type"),
					resource.TestCheckResourceAttr(resourceNameGSI, "read_capacity", "1"),
					resource.TestCheckResourceAttr(resourceNameGSI, "table", rNameTable),
					resource.TestCheckResourceAttr(resourceNameGSI, "write_capacity", "1"),
					resource.TestCheckResourceAttr(resourceNameGSI, "on_demand_throughput.#", "0"),
					resource.TestCheckResourceAttr(resourceNameGSI, "warm_throughput.#", "0"),
				),
			},
			{
				ResourceName:                         resourceNameGSI,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlobalSecondaryIndexImportStateFunc(resourceNameGSI),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func testAccGlobalSecondaryIndexImportStateFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		return rs.Primary.Attributes[names.AttrARN], nil
	}
}

func testAccCheckInitialGSIExists(n string, tbl *awstypes.TableDescription, gsi *awstypes.GlobalSecondaryIndexDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if tbl == nil {
			return errors.New("table config is empty")
		}

		for _, g := range tbl.GlobalSecondaryIndexes {
			if rs.Primary.Attributes[names.AttrName] == aws.ToString(g.IndexName) {
				*gsi = g

				break
			}
		}

		return nil
	}
}

func testAccGlobalSecondaryIndexConfig_basic(tableName, indexName string) string {
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

resource "aws_dynamodb_global_secondary_index" "test" {
  table           = aws_dynamodb_table.test.name
  name            = %[2]q
  hash_key        = %[1]q
  projection_type = "ALL"
}
`, tableName, indexName)
}

func testAccGlobalSecondaryIndexConfig_basic_withCapacity(tableName, indexName string, capacity int) string {
	return fmt.Sprintf(`
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

resource "aws_dynamodb_global_secondary_index" "test" {
  table           = aws_dynamodb_table.test.name
  name            = %[2]q
  hash_key        = %[1]q
  range_key       = %[2]q
  range_key_type  = "S"
  projection_type = "ALL"
  read_capacity   = %[3]d
  write_capacity  = %[3]d
}
`, tableName, indexName, capacity)
}

func testAccGlobalSecondaryIndexConfig_basic_withCapacityAndIgnoreChanges(tableName, indexName string, capacity int) string {
	return fmt.Sprintf(`
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

resource "aws_dynamodb_global_secondary_index" "test" {
  table           = aws_dynamodb_table.test.name
  name            = %[2]q
  hash_key        = %[1]q
  range_key       = %[2]q
  range_key_type  = "S"
  projection_type = "ALL"
  read_capacity   = %[3]d
  write_capacity  = %[3]d

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
resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  hash_key     = %[1]q
  billing_mode = "PAY_PER_REQUEST"

  attribute {
    name = %[1]q
    type = "S"
  }
}

resource "aws_dynamodb_global_secondary_index" "test" {
  table           = aws_dynamodb_table.test.name
  name            = %[2]q
  hash_key        = %[1]q
  range_key       = %[2]q
  range_key_type  = "S"
  projection_type = "ALL"
}
`, tableName, indexName)
}

func testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughput(tableName, indexName string) string {
	return fmt.Sprintf(`
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

resource "aws_dynamodb_global_secondary_index" "test" {
  table           = aws_dynamodb_table.test.name
  name            = %[2]q
  hash_key        = %[1]q
  range_key       = %[2]q
  range_key_type  = "S"
  projection_type = "ALL"

  on_demand_throughput {
    max_read_request_units  = 1
    max_write_request_units = 1
  }
}
`, tableName, indexName)
}

func testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughputWithCapacity(tableName, indexName string, capacity int) string {
	return fmt.Sprintf(`
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

resource "aws_dynamodb_global_secondary_index" "test" {
  table           = aws_dynamodb_table.test.name
  name            = %[2]q
  hash_key        = %[1]q
  range_key       = %[2]q
  range_key_type  = "S"
  projection_type = "ALL"

  on_demand_throughput {
    max_read_request_units  = %[3]d
    max_write_request_units = %[3]d
  }
}
`, tableName, indexName, capacity)
}

func testAccGlobalSecondaryIndexConfig_billingPayPerRequest_onDemandThroughputWithCapacityAndIgnoreChanges(tableName, indexName string, capacity int) string {
	return fmt.Sprintf(`
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

resource "aws_dynamodb_global_secondary_index" "test" {
  table           = aws_dynamodb_table.test.name
  name            = %[2]q
  hash_key        = %[1]q
  range_key       = %[2]q
  range_key_type  = "S"
  projection_type = "ALL"

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
`, tableName, indexName, capacity)
}

func testAccGlobalSecondaryIndexConfig_billingPayPerRequest_warmThroughput(tableName, indexName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  hash_key     = %[1]q
  billing_mode = "PAY_PER_REQUEST"

  attribute {
    name = %[1]q
    type = "S"
  }
}

resource "aws_dynamodb_global_secondary_index" "test" {
  table           = aws_dynamodb_table.test.name
  name            = %[2]q
  hash_key        = %[1]q
  range_key       = %[2]q
  range_key_type  = "S"
  projection_type = "ALL"

  warm_throughput {
    read_units_per_second  = 15000
    write_units_per_second = 5000
  }
}
`, tableName, indexName)
}

func testAccGlobalSecondaryIndexConfig_validateAttribute_missmatchedType(tableName, indexName string) string {
	return fmt.Sprintf(`
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

resource "aws_dynamodb_global_secondary_index" "test" {
  table           = aws_dynamodb_table.test.name
  name            = %[2]q
  hash_key        = %[1]q
  hash_key_type   = "N"
  range_key       = %[2]q
  range_key_type  = "S"
  projection_type = "ALL"

  on_demand_throughput {
    max_read_request_units  = 1
    max_write_request_units = 1
  }
}
`, tableName, indexName)
}

func testAccGlobalSecondaryIndexConfig_validateAttribute_unknownType(tableName, indexName, key string) string {
	return fmt.Sprintf(`
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

resource "aws_dynamodb_global_secondary_index" "test" {
  table           = aws_dynamodb_table.test.name
  name            = %[2]q
  %[3]s           = %[2]q
  projection_type = "ALL"

  on_demand_throughput {
    max_read_request_units  = 1
    max_write_request_units = 1
  }
}
`, tableName, indexName, key)
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

resource "aws_dynamodb_global_secondary_index" "test" {
  table           = aws_dynamodb_table.test.name
  name            = %[2]q
  hash_key        = %[1]q
  projection_type = %[3]q
  %[4]s
}
`, tableName, indexName, projectionType, nka)
}

func testAccGlobalSecondaryIndexConfig_differentKeys(tableName, indexName, pkGsi, skGsi string) string {
	if skGsi != "" {
		skGsi = fmt.Sprintf("range_key = %q\n  range_key_type = \"S\"\n", skGsi)
	}

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

resource "aws_dynamodb_global_secondary_index" "test" {
  table           = aws_dynamodb_table.test.name
  name            = %[2]q
  hash_key        = %[3]q
  hash_key_type   = "S"
  projection_type = "ALL"
  %[4]s
}
`, tableName, indexName, pkGsi, skGsi)
}
