// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package keyspaces_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/keyspaces"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfkeyspaces "github.com/hashicorp/terraform-provider-aws/internal/service/keyspaces"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKeyspacesTable_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v keyspaces.GetTableOutput
	rName1 := "tf_acc_test_" + acctest.RandString(t, 20)
	rName2 := "tf_acc_test_" + acctest.RandString(t, 20)
	resourceName := "aws_keyspaces_table.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_basic(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "cassandra", fmt.Sprintf("/keyspace/%s/table/%s", rName1, rName2)),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.0.throughput_mode", "PAY_PER_REQUEST"),
					resource.TestCheckResourceAttr(resourceName, "client_side_timestamps.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "comment.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "comment.0.message", ""),
					resource.TestCheckResourceAttr(resourceName, "default_time_to_live", "0"),
					resource.TestCheckResourceAttr(resourceName, "encryption_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_specification.0.type", "AWS_OWNED_KMS_KEY"),
					resource.TestCheckResourceAttr(resourceName, "keyspace_name", rName1),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.0.status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.clustering_key.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.column.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						names.AttrName: names.AttrMessage,
						names.AttrType: "ascii",
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.partition_key.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.partition_key.*", map[string]string{
						names.AttrName: names.AttrMessage,
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.static_column.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rName2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "ttl.#", "0"),
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
	rName1 := "tf_acc_test_" + acctest.RandString(t, 20)
	rName2 := "tf_acc_test_" + acctest.RandString(t, 20)
	resourceName := "aws_keyspaces_table.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_basic(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfkeyspaces.ResourceTable(), resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccKeyspacesTable_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v keyspaces.GetTableOutput
	rName1 := "tf_acc_test_" + acctest.RandString(t, 20)
	rName2 := "tf_acc_test_" + acctest.RandString(t, 20)
	resourceName := "aws_keyspaces_table.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_tags1(rName1, rName2, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
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
					testAccCheckTableExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccTableConfig_tags1(rName1, rName2, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccKeyspacesTable_clientSideTimestamps(t *testing.T) {
	ctx := acctest.Context(t)
	var v keyspaces.GetTableOutput
	rName1 := "tf_acc_test_" + acctest.RandString(t, 20)
	rName2 := "tf_acc_test_" + acctest.RandString(t, 20)
	resourceName := "aws_keyspaces_table.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_clientSideTimestamps(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "client_side_timestamps.#", "1"),
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
	rName1 := "tf_acc_test_" + acctest.RandString(t, 20)
	rName2 := "tf_acc_test_" + acctest.RandString(t, 20)
	resourceName := "aws_keyspaces_table.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_multipleColumns(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.clustering_key.#", "2"),
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
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.partition_key.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.partition_key.*", map[string]string{
						names.AttrName: names.AttrID,
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.static_column.#", "2"),
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
	rName1 := "tf_acc_test_" + acctest.RandString(t, 20)
	rName2 := "tf_acc_test_" + acctest.RandString(t, 20)
	resourceName := "aws_keyspaces_table.test"
	kmsKeyResourceName := "aws_kms_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_allAttributes(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &v1),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "cassandra", fmt.Sprintf("/keyspace/%s/table/%s", rName1, rName2)),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.0.read_capacity_units", "200"),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.0.throughput_mode", "PROVISIONED"),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.0.write_capacity_units", "100"),
					resource.TestCheckResourceAttr(resourceName, "comment.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "comment.0.message", "TESTING"),
					resource.TestCheckResourceAttr(resourceName, "default_time_to_live", "500000"),
					resource.TestCheckResourceAttr(resourceName, "encryption_specification.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_specification.0.kms_key_identifier", kmsKeyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "encryption_specification.0.type", "CUSTOMER_MANAGED_KMS_KEY"),
					resource.TestCheckResourceAttr(resourceName, "keyspace_name", rName1),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.0.status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rName2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "ttl.#", "1"),
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
					testAccCheckTableExists(ctx, t, resourceName, &v2),
					testAccCheckTableNotRecreated(&v1, &v2),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "cassandra", fmt.Sprintf("/keyspace/%s/table/%s", rName1, rName2)),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.0.read_capacity_units", "0"),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.0.throughput_mode", "PAY_PER_REQUEST"),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.0.write_capacity_units", "0"),
					resource.TestCheckResourceAttr(resourceName, "comment.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "comment.0.message", "TESTING"),
					resource.TestCheckResourceAttr(resourceName, "default_time_to_live", "1500000"),
					resource.TestCheckResourceAttr(resourceName, "encryption_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_specification.0.kms_key_identifier", ""),
					resource.TestCheckResourceAttr(resourceName, "encryption_specification.0.type", "AWS_OWNED_KMS_KEY"),
					resource.TestCheckResourceAttr(resourceName, "keyspace_name", rName1),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.0.status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rName2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "ttl.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ttl.0.status", "ENABLED"),
				),
			},
		},
	})
}

func TestAccKeyspacesTable_addColumns(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 keyspaces.GetTableOutput
	rName1 := "tf_acc_test_" + acctest.RandString(t, 20)
	rName2 := "tf_acc_test_" + acctest.RandString(t, 20)
	resourceName := "aws_keyspaces_table.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_basic(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.clustering_key.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.column.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						names.AttrName: names.AttrMessage,
						names.AttrType: "ascii",
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.partition_key.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.partition_key.*", map[string]string{
						names.AttrName: names.AttrMessage,
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.static_column.#", "0"),
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
					testAccCheckTableExists(ctx, t, resourceName, &v2),
					testAccCheckTableNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.clustering_key.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.column.#", "3"),
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
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.partition_key.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.partition_key.*", map[string]string{
						names.AttrName: names.AttrMessage,
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.static_column.#", "0"),
				),
			},
		},
	})
}

func TestAccKeyspacesTable_delColumns(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 keyspaces.GetTableOutput
	rName1 := "tf_acc_test_" + acctest.RandString(t, 20)
	rName2 := "tf_acc_test_" + acctest.RandString(t, 20)
	resourceName := "aws_keyspaces_table.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_newColumns(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.clustering_key.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.column.#", "3"),
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
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.partition_key.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.partition_key.*", map[string]string{
						names.AttrName: names.AttrMessage,
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.static_column.#", "0"),
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
					testAccCheckTableExists(ctx, t, resourceName, &v2),
					testAccCheckTableRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.clustering_key.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.column.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						names.AttrName: names.AttrMessage,
						names.AttrType: "ascii",
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.partition_key.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.partition_key.*", map[string]string{
						names.AttrName: names.AttrMessage,
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.static_column.#", "0"),
				),
			},
		},
	})
}

func TestAccKeyspacesTable_autoScaling_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v keyspaces.GetTableOutput
	rName1 := "tf_acc_test_" + acctest.RandString(t, 20)
	rName2 := "tf_acc_test_" + acctest.RandString(t, 20)
	resourceName := "aws_keyspaces_table.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_autoScaling(rName1, rName2, 5, 10, 70),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.0.throughput_mode", "PROVISIONED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_specification.0.read_capacity_auto_scaling.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_specification.0.read_capacity_auto_scaling.0.auto_scaling_disabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_specification.0.read_capacity_auto_scaling.0.minimum_units", "5"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_specification.0.read_capacity_auto_scaling.0.maximum_units", "10"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_specification.0.read_capacity_auto_scaling.0.target_tracking_scaling_policy_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_specification.0.read_capacity_auto_scaling.0.target_tracking_scaling_policy_configuration.0.target_value", "70"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_specification.0.write_capacity_auto_scaling.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_specification.0.write_capacity_auto_scaling.0.minimum_units", "5"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_specification.0.write_capacity_auto_scaling.0.maximum_units", "10"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_specification.0.write_capacity_auto_scaling.0.target_tracking_scaling_policy_configuration.0.target_value", "70"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"capacity_specification.0.read_capacity_units", "capacity_specification.0.write_capacity_units"},
			},
		},
	})
}

func TestAccKeyspacesTable_autoScaling_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 keyspaces.GetTableOutput
	rName1 := "tf_acc_test_" + acctest.RandString(t, 20)
	rName2 := "tf_acc_test_" + acctest.RandString(t, 20)
	resourceName := "aws_keyspaces_table.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_autoScaling(rName1, rName2, 5, 10, 70),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_specification.0.read_capacity_auto_scaling.0.maximum_units", "10"),
				),
			},
			{
				Config: testAccTableConfig_autoScaling(rName1, rName2, 5, 20, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &v2),
					testAccCheckTableNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_specification.0.read_capacity_auto_scaling.0.maximum_units", "20"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_specification.0.read_capacity_auto_scaling.0.target_tracking_scaling_policy_configuration.0.target_value", "60"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_specification.0.write_capacity_auto_scaling.0.maximum_units", "20"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_specification.0.write_capacity_auto_scaling.0.target_tracking_scaling_policy_configuration.0.target_value", "60"),
				),
			},
		},
	})
}

func TestAccKeyspacesTable_autoScaling_disabled(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 keyspaces.GetTableOutput
	rName1 := "tf_acc_test_" + acctest.RandString(t, 20)
	rName2 := "tf_acc_test_" + acctest.RandString(t, 20)
	resourceName := "aws_keyspaces_table.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_autoScaling(rName1, rName2, 5, 10, 70),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_specification.0.read_capacity_auto_scaling.0.auto_scaling_disabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccTableConfig_autoScalingDisabled(rName1, rName2, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &v2),
					testAccCheckTableNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_specification.0.read_capacity_auto_scaling.0.auto_scaling_disabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_specification.0.write_capacity_auto_scaling.0.auto_scaling_disabled", acctest.CtTrue),
					// Confirm against the API, not just state, that both dimensions are disabled.
					testAccCheckTableAutoScalingDisabled(ctx, t, resourceName),
				),
			},
		},
	})
}

func TestAccKeyspacesTable_autoScaling_payPerRequestConflict(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := "tf_acc_test_" + acctest.RandString(t, 20)
	rName2 := "tf_acc_test_" + acctest.RandString(t, 20)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccTableConfig_autoScalingPayPerRequest(rName1, rName2),
				ExpectError: regexache.MustCompile(`auto_scaling_specification requires capacity_specification.throughput_mode to be`),
			},
		},
	})
}

func TestAccKeyspacesTable_autoScaling_omittedCapacitySpecificationConflict(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := "tf_acc_test_" + acctest.RandString(t, 20)
	rName2 := "tf_acc_test_" + acctest.RandString(t, 20)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				// capacity_specification is entirely omitted here, which defaults to
				// PAY_PER_REQUEST; this must be rejected the same way an explicit
				// PAY_PER_REQUEST is.
				Config:      testAccTableConfig_autoScalingOmittedCapacitySpecification(rName1, rName2),
				ExpectError: regexache.MustCompile(`auto_scaling_specification requires capacity_specification.throughput_mode to be`),
			},
		},
	})
}

func TestAccKeyspacesTable_autoScaling_invertedRangeConflict(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := "tf_acc_test_" + acctest.RandString(t, 20)
	rName2 := "tf_acc_test_" + acctest.RandString(t, 20)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				// minimum_units > maximum_units must be rejected at plan time rather than
				// left for the AWS API to reject at apply time.
				Config:      testAccTableConfig_autoScaling(rName1, rName2, 10, 5, 70),
				ExpectError: regexache.MustCompile(`minimum_units \(10\) cannot be greater than maximum_units \(5\)`),
			},
		},
	})
}

func TestAccKeyspacesTable_autoScaling_removal(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 keyspaces.GetTableOutput
	rName1 := "tf_acc_test_" + acctest.RandString(t, 20)
	rName2 := "tf_acc_test_" + acctest.RandString(t, 20)
	resourceName := "aws_keyspaces_table.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_autoScaling(rName1, rName2, 5, 10, 70),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &v1),
				),
			},
			{
				// Removing the block must disable auto scaling against the API
				Config: testAccTableConfig_provisionedCapacity(rName1, rName2, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &v2),
					testAccCheckTableNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.0.throughput_mode", "PROVISIONED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_specification.#", "0"),
					testAccCheckTableAutoScalingDisabled(ctx, t, resourceName),
				),
			},
		},
	})
}

func TestAccKeyspacesTable_autoScaling_enableWithModeSwitch(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 keyspaces.GetTableOutput
	rName1 := "tf_acc_test_" + acctest.RandString(t, 20)
	rName2 := "tf_acc_test_" + acctest.RandString(t, 20)
	resourceName := "aws_keyspaces_table.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_basic(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.0.throughput_mode", "PAY_PER_REQUEST"),
				),
			},
			{
				// Step 1 of the two-step transition: switch PAY_PER_REQUEST -> PROVISIONED
				// with explicit capacity units before auto scaling is involved. Enabling auto
				// scaling in the same apply is intentionally rejected when the units are held
				// under ignore_changes (see _modeSwitchWithIgnoredUnitsConflict).
				Config: testAccTableConfig_provisionedCapacity(rName1, rName2, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &v2),
					testAccCheckTableNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.0.throughput_mode", "PROVISIONED"),
				),
			},
			{
				// Step 2: enable auto scaling now that the table is already PROVISIONED. The
				// capacity units (5) are established in prior state, so ignore_changes keeps
				// them in the plan and the transition succeeds.
				Config: testAccTableConfig_autoScaling(rName1, rName2, 5, 10, 70),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &v2),
					testAccCheckTableNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.0.throughput_mode", "PROVISIONED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_specification.#", "1"),
				),
			},
		},
	})
}

func TestAccKeyspacesTable_autoScaling_modeSwitchWithIgnoredUnitsConflict(t *testing.T) {
	ctx := acctest.Context(t)
	var v1 keyspaces.GetTableOutput
	rName1 := "tf_acc_test_" + acctest.RandString(t, 20)
	rName2 := "tf_acc_test_" + acctest.RandString(t, 20)
	resourceName := "aws_keyspaces_table.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_basic(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &v1),
				),
			},
			{
				// Switching PAY_PER_REQUEST -> PROVISIONED while the capacity units are held
				// under ignore_changes must be rejected at plan time: the required units are
				// not part of the planned change, and recovering them from raw configuration
				// would bypass the plan. The transition must be done in two steps instead.
				Config:      testAccTableConfig_autoScaling(rName1, rName2, 5, 10, 70),
				ExpectError: regexache.MustCompile(`requires read_capacity_units and write_capacity_units`),
			},
		},
	})
}

func TestAccKeyspacesTable_autoScaling_disableWithModeSwitch(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 keyspaces.GetTableOutput
	rName1 := "tf_acc_test_" + acctest.RandString(t, 20)
	rName2 := "tf_acc_test_" + acctest.RandString(t, 20)
	resourceName := "aws_keyspaces_table.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_autoScaling(rName1, rName2, 5, 10, 70),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &v1),
				),
			},
			{
				// Disabling auto scaling and switching PROVISIONED -> PAY_PER_REQUEST in one apply:
				// the provider must issue the auto-scaling-disable call before the capacity-mode
				// call, since AWS rejects PAY_PER_REQUEST while provisioned auto scaling is still active.
				Config: testAccTableConfig_payPerRequestExplicit(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceName, &v2),
					testAccCheckTableNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.0.throughput_mode", "PAY_PER_REQUEST"),
				),
			},
		},
	})
}

func testAccCheckTableDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).KeyspacesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_keyspaces_table" {
				continue
			}

			keyspaceName, tableName, err := tfkeyspaces.TableParseResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfkeyspaces.FindTableByTwoPartKey(ctx, conn, keyspaceName, tableName)

			if retry.NotFound(err) {
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

func testAccCheckTableExists(ctx context.Context, t *testing.T, n string, v *keyspaces.GetTableOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Keyspaces Table ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).KeyspacesClient(ctx)

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

// testAccCheckTableAutoScalingDisabled asserts against the API that both capacity
// dimensions report auto scaling as disabled, i.e. that removing the block actually
// turned auto scaling off rather than leaving it silently active.
func testAccCheckTableAutoScalingDisabled(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).KeyspacesClient(ctx)

		keyspaceName, tableName, err := tfkeyspaces.TableParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		output, err := tfkeyspaces.FindTableAutoScalingSettingsByTwoPartKey(ctx, conn, keyspaceName, tableName)

		// AWS may drop the scalable targets entirely once auto scaling is disabled,
		// which surfaces as a not-found error; that is a valid "disabled" outcome.
		if retry.NotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}

		spec := output.AutoScalingSpecification
		if spec == nil {
			return nil
		}

		if v := spec.ReadCapacityAutoScaling; v != nil && !v.AutoScalingDisabled {
			return fmt.Errorf("read capacity auto scaling is still enabled for %s", n)
		}
		if v := spec.WriteCapacityAutoScaling; v != nil && !v.AutoScalingDisabled {
			return fmt.Errorf("write capacity auto scaling is still enabled for %s", n)
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
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
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
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
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

func testAccTableConfig_autoScaling(rName1, rName2 string, minimumUnits, maximumUnits, targetValue int) string {
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

  capacity_specification {
    throughput_mode      = "PROVISIONED"
    read_capacity_units  = %[3]d
    write_capacity_units = %[3]d
  }

  auto_scaling_specification {
    read_capacity_auto_scaling {
      minimum_units = %[3]d
      maximum_units = %[4]d
      target_tracking_scaling_policy_configuration {
        target_value = %[5]d
      }
    }
    write_capacity_auto_scaling {
      minimum_units = %[3]d
      maximum_units = %[4]d
      target_tracking_scaling_policy_configuration {
        target_value = %[5]d
      }
    }
  }

  lifecycle {
    ignore_changes = [
      capacity_specification[0].read_capacity_units,
      capacity_specification[0].write_capacity_units,
    ]
  }
}
`, rName1, rName2, minimumUnits, maximumUnits, targetValue)
}

func testAccTableConfig_autoScalingDisabled(rName1, rName2 string, capacityUnits int) string {
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

  capacity_specification {
    throughput_mode      = "PROVISIONED"
    read_capacity_units  = %[3]d
    write_capacity_units = %[3]d
  }

  auto_scaling_specification {
    read_capacity_auto_scaling {
      auto_scaling_disabled = true
    }
    write_capacity_auto_scaling {
      auto_scaling_disabled = true
    }
  }

  lifecycle {
    ignore_changes = [
      capacity_specification[0].read_capacity_units,
      capacity_specification[0].write_capacity_units,
    ]
  }
}
`, rName1, rName2, capacityUnits)
}

func testAccTableConfig_autoScalingPayPerRequest(rName1, rName2 string) string {
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

  capacity_specification {
    throughput_mode = "PAY_PER_REQUEST"
  }

  auto_scaling_specification {
    read_capacity_auto_scaling {
      minimum_units = 5
      maximum_units = 10
      target_tracking_scaling_policy_configuration {
        target_value = 70
      }
    }
  }
}
`, rName1, rName2)
}

func testAccTableConfig_autoScalingOmittedCapacitySpecification(rName1, rName2 string) string {
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

  auto_scaling_specification {
    read_capacity_auto_scaling {
      minimum_units = 5
      maximum_units = 10
      target_tracking_scaling_policy_configuration {
        target_value = 70
      }
    }
  }
}
`, rName1, rName2)
}

func testAccTableConfig_provisionedCapacity(rName1, rName2 string, units int) string {
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

  capacity_specification {
    throughput_mode      = "PROVISIONED"
    read_capacity_units  = %[3]d
    write_capacity_units = %[3]d
  }
}
`, rName1, rName2, units)
}

func testAccTableConfig_payPerRequestExplicit(rName1, rName2 string) string {
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

  capacity_specification {
    throughput_mode = "PAY_PER_REQUEST"
  }
}
`, rName1, rName2)
}
