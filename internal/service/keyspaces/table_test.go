// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package keyspaces_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/keyspaces"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkeyspaces "github.com/hashicorp/terraform-provider-aws/internal/service/keyspaces"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKeyspacesTable_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v keyspaces.GetTableOutput
	rName1 := "tf_acc_test_" + sdkacctest.RandString(20)
	rName2 := "tf_acc_test_" + sdkacctest.RandString(20)
	resourceName := "aws_keyspaces_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_basic(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "cassandra", fmt.Sprintf("/keyspace/%s/table/%s", rName1, rName2)),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.0.throughput_mode", "PAY_PER_REQUEST"),
					resource.TestCheckResourceAttr(resourceName, "client_side_timestamps.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "comment.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "comment.0.message", ""),
					resource.TestCheckResourceAttr(resourceName, "default_time_to_live", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "encryption_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_specification.0.type", "AWS_OWNED_KMS_KEY"),
					resource.TestCheckResourceAttr(resourceName, "keyspace_name", rName1),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.0.status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.clustering_key.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.column.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						names.AttrName: names.AttrMessage,
						names.AttrType: "ascii",
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.partition_key.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.partition_key.*", map[string]string{
						names.AttrName: names.AttrMessage,
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.static_column.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rName2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ttl.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccKeyspacesTable_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v keyspaces.GetTableOutput
	rName1 := "tf_acc_test_" + sdkacctest.RandString(20)
	rName2 := "tf_acc_test_" + sdkacctest.RandString(20)
	resourceName := "aws_keyspaces_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_basic(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfkeyspaces.ResourceTable(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKeyspacesTable_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v keyspaces.GetTableOutput
	rName1 := "tf_acc_test_" + sdkacctest.RandString(20)
	rName2 := "tf_acc_test_" + sdkacctest.RandString(20)
	resourceName := "aws_keyspaces_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_tags1(rName1, rName2, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableConfig_tags2(rName1, rName2, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccTableConfig_tags1(rName1, rName2, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccKeyspacesTable_clientSideTimestamps(t *testing.T) {
	ctx := acctest.Context(t)
	var v keyspaces.GetTableOutput
	rName1 := "tf_acc_test_" + sdkacctest.RandString(20)
	rName2 := "tf_acc_test_" + sdkacctest.RandString(20)
	resourceName := "aws_keyspaces_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_clientSideTimestamps(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "client_side_timestamps.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "client_side_timestamps.0.status", "ENABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccKeyspacesTable_multipleColumns(t *testing.T) {
	ctx := acctest.Context(t)
	var v keyspaces.GetTableOutput
	rName1 := "tf_acc_test_" + sdkacctest.RandString(20)
	rName2 := "tf_acc_test_" + sdkacctest.RandString(20)
	resourceName := "aws_keyspaces_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_multipleColumns(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.clustering_key.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.clustering_key.*", map[string]string{
						names.AttrName: "division",
						"order_by":     "ASC",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.clustering_key.*", map[string]string{
						names.AttrName: names.AttrRegion,
						"order_by":     "DESC",
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.column.#", "11"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						names.AttrName: names.AttrID,
						names.AttrType: "text",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						names.AttrName: "n",
						names.AttrType: "text",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						names.AttrName: names.AttrRegion,
						names.AttrType: "text",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						names.AttrName: "division",
						names.AttrType: "text",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						names.AttrName: "project",
						names.AttrType: "text",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						names.AttrName: names.AttrRole,
						names.AttrType: "text",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						names.AttrName: "pay_scale0",
						names.AttrType: "int",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						names.AttrName: "vacation_hrs",
						names.AttrType: "float",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						names.AttrName: "manager_id",
						names.AttrType: "text",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						names.AttrName: "nicknames",
						names.AttrType: "list<text>",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						names.AttrName: names.AttrTags,
						names.AttrType: "map<text, text>",
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.partition_key.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.partition_key.*", map[string]string{
						names.AttrName: names.AttrID,
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.static_column.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.static_column.*", map[string]string{
						names.AttrName: names.AttrRole,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.static_column.*", map[string]string{
						names.AttrName: "pay_scale0",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccKeyspacesTable_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 keyspaces.GetTableOutput
	rName1 := "tf_acc_test_" + sdkacctest.RandString(20)
	rName2 := "tf_acc_test_" + sdkacctest.RandString(20)
	resourceName := "aws_keyspaces_table.test"
	kmsKeyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_allAttributes(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &v1),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "cassandra", fmt.Sprintf("/keyspace/%s/table/%s", rName1, rName2)),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.0.read_capacity_units", "200"),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.0.throughput_mode", "PROVISIONED"),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.0.write_capacity_units", "100"),
					resource.TestCheckResourceAttr(resourceName, "comment.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "comment.0.message", "TESTING"),
					resource.TestCheckResourceAttr(resourceName, "default_time_to_live", "500000"),
					resource.TestCheckResourceAttr(resourceName, "encryption_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_specification.0.kms_key_identifier", kmsKeyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "encryption_specification.0.type", "CUSTOMER_MANAGED_KMS_KEY"),
					resource.TestCheckResourceAttr(resourceName, "keyspace_name", rName1),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.0.status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rName2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ttl.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ttl.0.status", "ENABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableConfig_allAttributesUpdated(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &v2),
					testAccCheckTableNotRecreated(&v1, &v2),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "cassandra", fmt.Sprintf("/keyspace/%s/table/%s", rName1, rName2)),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.0.read_capacity_units", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.0.throughput_mode", "PAY_PER_REQUEST"),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.0.write_capacity_units", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "comment.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "comment.0.message", "TESTING"),
					resource.TestCheckResourceAttr(resourceName, "default_time_to_live", "1500000"),
					resource.TestCheckResourceAttr(resourceName, "encryption_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_specification.0.kms_key_identifier", ""),
					resource.TestCheckResourceAttr(resourceName, "encryption_specification.0.type", "AWS_OWNED_KMS_KEY"),
					resource.TestCheckResourceAttr(resourceName, "keyspace_name", rName1),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.0.status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rName2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ttl.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ttl.0.status", "ENABLED"),
				),
			},
		},
	})
}

func TestAccKeyspacesTable_addColumns(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 keyspaces.GetTableOutput
	rName1 := "tf_acc_test_" + sdkacctest.RandString(20)
	rName2 := "tf_acc_test_" + sdkacctest.RandString(20)
	resourceName := "aws_keyspaces_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_basic(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.clustering_key.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.column.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						names.AttrName: names.AttrMessage,
						names.AttrType: "ascii",
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.partition_key.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.partition_key.*", map[string]string{
						names.AttrName: names.AttrMessage,
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.static_column.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableConfig_newColumns(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &v2),
					testAccCheckTableNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.clustering_key.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.column.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						names.AttrName: names.AttrMessage,
						names.AttrType: "ascii",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						names.AttrName: "ts",
						names.AttrType: "timestamp",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						names.AttrName: "amount",
						names.AttrType: "decimal",
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.partition_key.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.partition_key.*", map[string]string{
						names.AttrName: names.AttrMessage,
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.static_column.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccKeyspacesTable_delColumns(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 keyspaces.GetTableOutput
	rName1 := "tf_acc_test_" + sdkacctest.RandString(20)
	rName2 := "tf_acc_test_" + sdkacctest.RandString(20)
	resourceName := "aws_keyspaces_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_newColumns(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.clustering_key.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.column.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						names.AttrName: names.AttrMessage,
						names.AttrType: "ascii",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						names.AttrName: "ts",
						names.AttrType: "timestamp",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						names.AttrName: "amount",
						names.AttrType: "decimal",
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.partition_key.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.partition_key.*", map[string]string{
						names.AttrName: names.AttrMessage,
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.static_column.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableConfig_basic(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &v2),
					testAccCheckTableRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.clustering_key.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.column.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						names.AttrName: names.AttrMessage,
						names.AttrType: "ascii",
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.partition_key.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.partition_key.*", map[string]string{
						names.AttrName: names.AttrMessage,
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.static_column.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccCheckTableDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KeyspacesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_keyspaces_table" {
				continue
			}

			keyspaceName, tableName, err := tfkeyspaces.TableParseResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfkeyspaces.FindTableByTwoPartKey(ctx, conn, keyspaceName, tableName)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Keyspaces Table %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTableExists(ctx context.Context, n string, v *keyspaces.GetTableOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Keyspaces Table ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KeyspacesClient(ctx)

		keyspaceName, tableName, err := tfkeyspaces.TableParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		output, err := tfkeyspaces.FindTableByTwoPartKey(ctx, conn, keyspaceName, tableName)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTableNotRecreated(i, j *keyspaces.GetTableOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.ToTime(i.CreationTimestamp).Equal(aws.ToTime(j.CreationTimestamp)) {
			return errors.New("Keyspaces Table was recreated")
		}

		return nil
	}
}

func testAccCheckTableRecreated(i, j *keyspaces.GetTableOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToTime(i.CreationTimestamp).Equal(aws.ToTime(j.CreationTimestamp)) {
			return errors.New("Keyspaces Table was not recreated")
		}

		return nil
	}
}

func testAccTableConfig_basic(rName1, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_keyspaces_keyspace" "test" {
  name = %[1]q
}

resource "aws_keyspaces_table" "test" {
  keyspace_name = aws_keyspaces_keyspace.test.name
  table_name    = %[2]q

  schema_definition {
    column {
      name = "message"
      type = "ascii"
    }

    partition_key {
      name = "message"
    }
  }
}
`, rName1, rName2)
}

func testAccTableConfig_tags1(rName1, rName2, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_keyspaces_keyspace" "test" {
  name = %[1]q
}

resource "aws_keyspaces_table" "test" {
  keyspace_name = aws_keyspaces_keyspace.test.name
  table_name    = %[2]q

  schema_definition {
    column {
      name = "message"
      type = "ascii"
    }

    partition_key {
      name = "message"
    }
  }

  tags = {
    %[3]q = %[4]q
  }
}
`, rName1, rName2, tagKey1, tagValue1)
}

func testAccTableConfig_tags2(rName1, rName2, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_keyspaces_keyspace" "test" {
  name = %[1]q
}

resource "aws_keyspaces_table" "test" {
  keyspace_name = aws_keyspaces_keyspace.test.name
  table_name    = %[2]q

  schema_definition {
    column {
      name = "message"
      type = "ascii"
    }

    partition_key {
      name = "message"
    }
  }

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName1, rName2, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccTableConfig_clientSideTimestamps(rName1, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_keyspaces_keyspace" "test" {
  name = %[1]q
}

resource "aws_keyspaces_table" "test" {
  keyspace_name = aws_keyspaces_keyspace.test.name
  table_name    = %[2]q

  schema_definition {
    column {
      name = "message"
      type = "ascii"
    }

    partition_key {
      name = "message"
    }
  }

  client_side_timestamps {
    status = "ENABLED"
  }
}
`, rName1, rName2)
}

func testAccTableConfig_multipleColumns(rName1, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_keyspaces_keyspace" "test" {
  name = %[1]q
}

resource "aws_keyspaces_table" "test" {
  keyspace_name = aws_keyspaces_keyspace.test.name
  table_name    = %[2]q

  schema_definition {
    column {
      name = "id"
      type = "text"
    }

    column {
      name = "n"
      type = "text"
    }

    column {
      name = "region"
      type = "text"
    }

    column {
      name = "division"
      type = "text"
    }

    column {
      name = "project"
      type = "text"
    }

    column {
      name = "role"
      type = "text"
    }

    column {
      name = "pay_scale0"
      type = "int"
    }

    column {
      name = "vacation_hrs"
      type = "float"
    }

    column {
      name = "manager_id"
      type = "text"
    }

    column {
      name = "nicknames"
      type = "list<text>"
    }

    column {
      name = "tags"
      type = "map<text, text>"
    }

    partition_key {
      name = "id"
    }

    clustering_key {
      name     = "division"
      order_by = "ASC"
    }

    clustering_key {
      name     = "region"
      order_by = "DESC"
    }

    static_column {
      name = "role"
    }

    static_column {
      name = "pay_scale0"
    }
  }
}
`, rName1, rName2)
}

func testAccTableConfig_allAttributes(rName1, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q
}

resource "aws_keyspaces_keyspace" "test" {
  name = %[1]q
}

resource "aws_keyspaces_table" "test" {
  keyspace_name = aws_keyspaces_keyspace.test.name
  table_name    = %[2]q

  schema_definition {
    column {
      name = "message"
      type = "ascii"
    }

    partition_key {
      name = "message"
    }
  }

  capacity_specification {
    read_capacity_units  = 200
    throughput_mode      = "PROVISIONED"
    write_capacity_units = 100
  }

  comment {
    message = "TESTING"
  }

  default_time_to_live = 500000

  encryption_specification {
    kms_key_identifier = aws_kms_key.test.arn
    type               = "CUSTOMER_MANAGED_KMS_KEY"
  }

  point_in_time_recovery {
    status = "ENABLED"
  }

  ttl {
    status = "ENABLED"
  }
}
`, rName1, rName2)
}

func testAccTableConfig_allAttributesUpdated(rName1, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q
}

resource "aws_keyspaces_keyspace" "test" {
  name = %[1]q
}

resource "aws_keyspaces_table" "test" {
  keyspace_name = aws_keyspaces_keyspace.test.name
  table_name    = %[2]q

  schema_definition {
    column {
      name = "message"
      type = "ascii"
    }

    partition_key {
      name = "message"
    }
  }

  capacity_specification {
    throughput_mode = "PAY_PER_REQUEST"
  }

  comment {
    message = "TESTING"
  }

  default_time_to_live = 1500000

  encryption_specification {
    type = "AWS_OWNED_KMS_KEY"
  }

  point_in_time_recovery {
    status = "DISABLED"
  }

  ttl {
    status = "ENABLED"
  }
}
`, rName1, rName2)
}

func testAccTableConfig_newColumns(rName1, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_keyspaces_keyspace" "test" {
  name = %[1]q
}

resource "aws_keyspaces_table" "test" {
  keyspace_name = aws_keyspaces_keyspace.test.name
  table_name    = %[2]q

  schema_definition {
    column {
      name = "message"
      type = "ascii"
    }

    column {
      name = "ts"
      type = "timestamp"
    }

    column {
      name = "amount"
      type = "decimal"
    }

    partition_key {
      name = "message"
    }
  }
}
`, rName1, rName2)
}
