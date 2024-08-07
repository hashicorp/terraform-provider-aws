// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfdynamodb "github.com/hashicorp/terraform-provider-aws/internal/service/dynamodb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.DynamoDBServiceID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"Unsupported input parameter TableClass",
	)
}

func TestUpdateDiffGSI(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Old             []interface{}
		New             []interface{}
		ExpectedUpdates []awstypes.GlobalSecondaryIndexUpdate
	}{
		{ // No-op => no changes
			Old: []interface{}{
				map[string]interface{}{
					names.AttrName:    "att1-index",
					"hash_key":        "att1",
					"write_capacity":  10,
					"read_capacity":   10,
					"projection_type": "ALL",
				},
			},
			New: []interface{}{
				map[string]interface{}{
					names.AttrName:    "att1-index",
					"hash_key":        "att1",
					"write_capacity":  10,
					"read_capacity":   10,
					"projection_type": "ALL",
				},
			},
			ExpectedUpdates: nil,
		},
		{ // No-op => ignore ordering of non_key_attributes
			Old: []interface{}{
				map[string]interface{}{
					names.AttrName:       "att1-index",
					"hash_key":           "att1",
					"write_capacity":     10,
					"read_capacity":      10,
					"projection_type":    "INCLUDE",
					"non_key_attributes": schema.NewSet(schema.HashString, []interface{}{"attr3", "attr1", "attr2"}),
				},
			},
			New: []interface{}{
				map[string]interface{}{
					names.AttrName:       "att1-index",
					"hash_key":           "att1",
					"write_capacity":     10,
					"read_capacity":      10,
					"projection_type":    "INCLUDE",
					"non_key_attributes": schema.NewSet(schema.HashString, []interface{}{"attr1", "attr2", "attr3"}),
				},
			},
			ExpectedUpdates: nil,
		},

		{ // Creation
			Old: []interface{}{
				map[string]interface{}{
					names.AttrName:    "att1-index",
					"hash_key":        "att1",
					"write_capacity":  10,
					"read_capacity":   10,
					"projection_type": "ALL",
				},
			},
			New: []interface{}{
				map[string]interface{}{
					names.AttrName:    "att1-index",
					"hash_key":        "att1",
					"write_capacity":  10,
					"read_capacity":   10,
					"projection_type": "ALL",
				},
				map[string]interface{}{
					names.AttrName:    "att2-index",
					"hash_key":        "att2",
					"write_capacity":  12,
					"read_capacity":   11,
					"projection_type": "ALL",
				},
			},
			ExpectedUpdates: []awstypes.GlobalSecondaryIndexUpdate{
				{
					Create: &awstypes.CreateGlobalSecondaryIndexAction{
						IndexName: aws.String("att2-index"),
						KeySchema: []awstypes.KeySchemaElement{
							{
								AttributeName: aws.String("att2"),
								KeyType:       awstypes.KeyTypeHash,
							},
						},
						ProvisionedThroughput: &awstypes.ProvisionedThroughput{
							WriteCapacityUnits: aws.Int64(12),
							ReadCapacityUnits:  aws.Int64(11),
						},
						Projection: &awstypes.Projection{
							ProjectionType: awstypes.ProjectionTypeAll,
						},
					},
				},
			},
		},

		{ // Deletion
			Old: []interface{}{
				map[string]interface{}{
					names.AttrName:    "att1-index",
					"hash_key":        "att1",
					"write_capacity":  10,
					"read_capacity":   10,
					"projection_type": "ALL",
				},
				map[string]interface{}{
					names.AttrName:    "att2-index",
					"hash_key":        "att2",
					"write_capacity":  12,
					"read_capacity":   11,
					"projection_type": "ALL",
				},
			},
			New: []interface{}{
				map[string]interface{}{
					names.AttrName:    "att1-index",
					"hash_key":        "att1",
					"write_capacity":  10,
					"read_capacity":   10,
					"projection_type": "ALL",
				},
			},
			ExpectedUpdates: []awstypes.GlobalSecondaryIndexUpdate{
				{
					Delete: &awstypes.DeleteGlobalSecondaryIndexAction{
						IndexName: aws.String("att2-index"),
					},
				},
			},
		},

		{ // Update
			Old: []interface{}{
				map[string]interface{}{
					names.AttrName:    "att1-index",
					"hash_key":        "att1",
					"write_capacity":  10,
					"read_capacity":   10,
					"projection_type": "ALL",
				},
			},
			New: []interface{}{
				map[string]interface{}{
					names.AttrName:    "att1-index",
					"hash_key":        "att1",
					"write_capacity":  20,
					"read_capacity":   30,
					"projection_type": "ALL",
				},
			},
			ExpectedUpdates: []awstypes.GlobalSecondaryIndexUpdate{
				{
					Update: &awstypes.UpdateGlobalSecondaryIndexAction{
						IndexName: aws.String("att1-index"),
						ProvisionedThroughput: &awstypes.ProvisionedThroughput{
							WriteCapacityUnits: aws.Int64(20),
							ReadCapacityUnits:  aws.Int64(30),
						},
					},
				},
			},
		},

		{ // Update of non-capacity attributes
			Old: []interface{}{
				map[string]interface{}{
					names.AttrName:    "att1-index",
					"hash_key":        "att1",
					"write_capacity":  10,
					"read_capacity":   10,
					"projection_type": "ALL",
				},
			},
			New: []interface{}{
				map[string]interface{}{
					names.AttrName:       "att1-index",
					"hash_key":           "att-new",
					"range_key":          "new-range-key",
					"write_capacity":     10,
					"read_capacity":      10,
					"projection_type":    "KEYS_ONLY",
					"non_key_attributes": schema.NewSet(schema.HashString, []interface{}{"RandomAttribute"}),
				},
			},
			ExpectedUpdates: []awstypes.GlobalSecondaryIndexUpdate{
				{
					Delete: &awstypes.DeleteGlobalSecondaryIndexAction{
						IndexName: aws.String("att1-index"),
					},
				},
				{
					Create: &awstypes.CreateGlobalSecondaryIndexAction{
						IndexName: aws.String("att1-index"),
						KeySchema: []awstypes.KeySchemaElement{
							{
								AttributeName: aws.String("att-new"),
								KeyType:       awstypes.KeyTypeHash,
							},
							{
								AttributeName: aws.String("new-range-key"),
								KeyType:       awstypes.KeyTypeRange,
							},
						},
						ProvisionedThroughput: &awstypes.ProvisionedThroughput{
							WriteCapacityUnits: aws.Int64(10),
							ReadCapacityUnits:  aws.Int64(10),
						},
						Projection: &awstypes.Projection{
							ProjectionType:   awstypes.ProjectionTypeKeysOnly,
							NonKeyAttributes: []string{"RandomAttribute"},
						},
					},
				},
			},
		},

		{ // Update of all attributes
			Old: []interface{}{
				map[string]interface{}{
					names.AttrName:    "att1-index",
					"hash_key":        "att1",
					"write_capacity":  10,
					"read_capacity":   10,
					"projection_type": "ALL",
				},
			},
			New: []interface{}{
				map[string]interface{}{
					names.AttrName:       "att1-index",
					"hash_key":           "att-new",
					"range_key":          "new-range-key",
					"write_capacity":     12,
					"read_capacity":      12,
					"projection_type":    "INCLUDE",
					"non_key_attributes": schema.NewSet(schema.HashString, []interface{}{"RandomAttribute"}),
				},
			},
			ExpectedUpdates: []awstypes.GlobalSecondaryIndexUpdate{
				{
					Delete: &awstypes.DeleteGlobalSecondaryIndexAction{
						IndexName: aws.String("att1-index"),
					},
				},
				{
					Create: &awstypes.CreateGlobalSecondaryIndexAction{
						IndexName: aws.String("att1-index"),
						KeySchema: []awstypes.KeySchemaElement{
							{
								AttributeName: aws.String("att-new"),
								KeyType:       awstypes.KeyTypeHash,
							},
							{
								AttributeName: aws.String("new-range-key"),
								KeyType:       awstypes.KeyTypeRange,
							},
						},
						ProvisionedThroughput: &awstypes.ProvisionedThroughput{
							WriteCapacityUnits: aws.Int64(12),
							ReadCapacityUnits:  aws.Int64(12),
						},
						Projection: &awstypes.Projection{
							ProjectionType:   awstypes.ProjectionTypeInclude,
							NonKeyAttributes: []string{"RandomAttribute"},
						},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		ops, err := tfdynamodb.UpdateDiffGSI(tc.Old, tc.New, awstypes.BillingModeProvisioned)
		if err != nil {
			t.Fatal(err)
		}

		if got, want := ops, tc.ExpectedUpdates; !reflect.DeepEqual(got, want) {
			t.Errorf(
				"%d: Got:\n\n%#v\n\nExpected:\n\n%#v\n",
				i,
				got,
				want)
		}
	}
}

func TestAccDynamoDBTable_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "dynamodb", fmt.Sprintf("table/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
						names.AttrName: rName,
						names.AttrType: "S",
					}),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", "PROVISIONED"),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "global_secondary_index.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "hash_key", rName),
					resource.TestCheckResourceAttr(resourceName, "local_secondary_index.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.0.enabled", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(resourceName, "range_key"),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "restore_date_time"),
					resource.TestCheckNoResourceAttr(resourceName, "restore_source_name"),
					resource.TestCheckNoResourceAttr(resourceName, "restore_to_latest_time"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrStreamARN, ""),
					resource.TestCheckResourceAttr(resourceName, "stream_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "stream_label", ""),
					resource.TestCheckResourceAttr(resourceName, "stream_view_type", ""),
					resource.TestCheckResourceAttr(resourceName, "table_class", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", acctest.Ct1),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ttl"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"attribute_name":  knownvalue.StringExact(""),
							names.AttrEnabled: knownvalue.Bool(false),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDynamoDBTable_deletion_protection(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_enable_deletion_protection(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "dynamodb", fmt.Sprintf("table/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
						names.AttrName: rName,
						names.AttrType: "S",
					}),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// disable deletion protection for the sweeper to work
			{
				Config: testAccTableConfig_disable_deletion_protection(rName),
				Check:  resource.TestCheckResourceAttr(resourceName, "deletion_protection_enabled", acctest.CtFalse),
			},
		},
	})
}

func TestAccDynamoDBTable_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdynamodb.ResourceTable(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDynamoDBTable_Disappears_payPerRequestWithGSI(t *testing.T) {
	ctx := acctest.Context(t)
	var table1, table2 awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_billingPayPerRequestGSI(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &table1),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdynamodb.ResourceTable(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccTableConfig_billingPayPerRequestGSI(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &table2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"global_secondary_index"},
			},
		},
	})
}

func TestAccDynamoDBTable_extended(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_initialState(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					testAccCheckInitialTableConf(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableConfig_addSecondaryGSI(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "hash_key", "TestTableHashKey"),
					resource.TestCheckResourceAttr(resourceName, "range_key", "TestTableRangeKey"),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", string(awstypes.BillingModeProvisioned)),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "attribute.#", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "global_secondary_index.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "local_secondary_index.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
						names.AttrName: "TestTableHashKey",
						names.AttrType: "S",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
						names.AttrName: "TestTableRangeKey",
						names.AttrType: "S",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
						names.AttrName: "TestLSIRangeKey",
						names.AttrType: "N",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
						names.AttrName: "ReplacementGSIRangeKey",
						names.AttrType: "N",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						names.AttrName:         "ReplacementTestTableGSI",
						"hash_key":             "TestTableHashKey",
						"range_key":            "ReplacementGSIRangeKey",
						"write_capacity":       "5",
						"read_capacity":        "5",
						"projection_type":      "INCLUDE",
						"non_key_attributes.#": acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "global_secondary_index.*.non_key_attributes.*", "TestNonKeyAttribute"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "local_secondary_index.*", map[string]string{
						names.AttrName:    "TestTableLSI",
						"range_key":       "TestLSIRangeKey",
						"projection_type": "ALL",
					}),
				),
			},
		},
	})
}

func TestAccDynamoDBTable_enablePITR(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_initialState(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					testAccCheckInitialTableConf(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableConfig_backup(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.0.enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccDynamoDBTable_BillingMode_payPerRequestToProvisioned(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_billingPayPerRequest(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", string(awstypes.BillingModePayPerRequest)),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableConfig_billingProvisioned(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", string(awstypes.BillingModeProvisioned)),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "5"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "5"),
				),
			},
		},
	})
}

func TestAccDynamoDBTable_BillingMode_payPerRequestToProvisionedIgnoreChanges(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_billingPayPerRequest(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", string(awstypes.BillingModePayPerRequest)),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableConfig_billingProvisionedIgnoreChanges(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", string(awstypes.BillingModeProvisioned)),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccDynamoDBTable_BillingMode_provisionedToPayPerRequest(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_billingProvisioned(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", string(awstypes.BillingModeProvisioned)),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "5"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableConfig_billingPayPerRequest(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", string(awstypes.BillingModePayPerRequest)),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccDynamoDBTable_BillingMode_provisionedToPayPerRequestIgnoreChanges(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_billingProvisioned(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", string(awstypes.BillingModeProvisioned)),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "5"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableConfig_billingPayPerRequestIgnoreChanges(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", string(awstypes.BillingModePayPerRequest)),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccDynamoDBTable_BillingModeGSI_payPerRequestToProvisioned(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_billingPayPerRequestGSI(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", string(awstypes.BillingModePayPerRequest)),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"global_secondary_index"},
			},
			{
				Config: testAccTableConfig_billingProvisionedGSI(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", string(awstypes.BillingModeProvisioned)),
				),
			},
		},
	})
}

func TestAccDynamoDBTable_BillingModeGSI_provisionedToPayPerRequest(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_billingProvisionedGSI(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", string(awstypes.BillingModeProvisioned)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableConfig_billingPayPerRequestGSI(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", string(awstypes.BillingModePayPerRequest)),
				),
			},
		},
	})
}

func TestAccDynamoDBTable_streamSpecification(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_streamSpecification(rName, true, "KEYS_ONLY"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stream_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "stream_view_type", "KEYS_ONLY"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrStreamARN, "dynamodb", regexache.MustCompile(fmt.Sprintf("table/%s/stream", rName))),
					resource.TestCheckResourceAttrSet(resourceName, "stream_label"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableConfig_streamSpecification(rName, false, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stream_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "stream_view_type", "KEYS_ONLY"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrStreamARN, "dynamodb", regexache.MustCompile(fmt.Sprintf("table/%s/stream", rName))),
					resource.TestCheckResourceAttrSet(resourceName, "stream_label"),
				),
			},
		},
	})
}

func TestAccDynamoDBTable_streamSpecificationDiffs(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_streamSpecification(rName, true, "KEYS_ONLY"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stream_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "stream_view_type", "KEYS_ONLY"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrStreamARN, "dynamodb", regexache.MustCompile(fmt.Sprintf("table/%s/stream", rName))),
					resource.TestCheckResourceAttrSet(resourceName, "stream_label"),
				),
			},
			{
				Config: testAccTableConfig_streamSpecification(rName, true, "NEW_IMAGE"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stream_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "stream_view_type", "NEW_IMAGE"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrStreamARN, "dynamodb", regexache.MustCompile(fmt.Sprintf("table/%s/stream", rName))),
					resource.TestCheckResourceAttrSet(resourceName, "stream_label"),
				),
			},
			{
				Config: testAccTableConfig_streamSpecification(rName, false, "NEW_IMAGE"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stream_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "stream_view_type", "NEW_IMAGE"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrStreamARN, "dynamodb", regexache.MustCompile(fmt.Sprintf("table/%s/stream", rName))),
					resource.TestCheckResourceAttrSet(resourceName, "stream_label"),
				),
			},
			{
				Config: testAccTableConfig_streamSpecification(rName, false, "KEYS_ONLY"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stream_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "stream_view_type", "KEYS_ONLY"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrStreamARN, "dynamodb", regexache.MustCompile(fmt.Sprintf("table/%s/stream", rName))),
					resource.TestCheckResourceAttrSet(resourceName, "stream_label"),
				),
			},
			{
				Config: testAccTableConfig_streamSpecification(rName, false, "null"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stream_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "stream_view_type", "KEYS_ONLY"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrStreamARN, "dynamodb", regexache.MustCompile(fmt.Sprintf("table/%s/stream", rName))),
					resource.TestCheckResourceAttrSet(resourceName, "stream_label"),
				),
			},
			{
				Config: testAccTableConfig_streamSpecification(rName, true, "KEYS_ONLY"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stream_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "stream_view_type", "KEYS_ONLY"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrStreamARN, "dynamodb", regexache.MustCompile(fmt.Sprintf("table/%s/stream", rName))),
					resource.TestCheckResourceAttrSet(resourceName, "stream_label"),
				),
			},
			{
				Config: testAccTableConfig_streamSpecification(rName, true, "KEYS_ONLY"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stream_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "stream_view_type", "KEYS_ONLY"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrStreamARN, "dynamodb", regexache.MustCompile(fmt.Sprintf("table/%s/stream", rName))),
					resource.TestCheckResourceAttrSet(resourceName, "stream_label"),
				),
			},
		},
	})
}

func TestAccDynamoDBTable_streamSpecificationValidation(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccTableConfig_streamSpecification("anything", true, ""),
				ExpectError: regexache.MustCompile(`stream_view_type is required when stream_enabled = true`),
			},
		},
	})
}

func TestAccDynamoDBTable_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_tags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					testAccCheckInitialTableConf(resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
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

// https://github.com/hashicorp/terraform/issues/13243
func TestAccDynamoDBTable_gsiUpdateCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_gsiUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "global_secondary_index.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"read_capacity":  acctest.Ct1,
						"write_capacity": acctest.Ct1,
						names.AttrName:   "att1-index",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"read_capacity":  acctest.Ct1,
						"write_capacity": acctest.Ct1,
						names.AttrName:   "att2-index",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"read_capacity":  acctest.Ct1,
						"write_capacity": acctest.Ct1,
						names.AttrName:   "att3-index",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableConfig_gsiUpdatedCapacity(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "global_secondary_index.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"read_capacity":  acctest.Ct2,
						"write_capacity": acctest.Ct2,
						names.AttrName:   "att1-index",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"read_capacity":  acctest.Ct2,
						"write_capacity": acctest.Ct2,
						names.AttrName:   "att2-index",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"read_capacity":  acctest.Ct2,
						"write_capacity": acctest.Ct2,
						names.AttrName:   "att3-index",
					}),
				),
			},
		},
	})
}

func TestAccDynamoDBTable_gsiUpdateOtherAttributes(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_gsiUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "global_secondary_index.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att3",
						names.AttrName:         "att3-index",
						"non_key_attributes.#": acctest.Ct0,
						"projection_type":      "ALL",
						"range_key":            "",
						"read_capacity":        acctest.Ct1,
						"write_capacity":       acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att1",
						names.AttrName:         "att1-index",
						"non_key_attributes.#": acctest.Ct0,
						"projection_type":      "ALL",
						"range_key":            "",
						"read_capacity":        acctest.Ct1,
						"write_capacity":       acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att2",
						names.AttrName:         "att2-index",
						"non_key_attributes.#": acctest.Ct0,
						"projection_type":      "ALL",
						"range_key":            "",
						"read_capacity":        acctest.Ct1,
						"write_capacity":       acctest.Ct1,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableConfig_gsiUpdatedOtherAttributes(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "global_secondary_index.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att4",
						names.AttrName:         "att2-index",
						"non_key_attributes.#": acctest.Ct0,
						"projection_type":      "ALL",
						"range_key":            "att2",
						"read_capacity":        acctest.Ct1,
						"write_capacity":       acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att3",
						names.AttrName:         "att3-index",
						"non_key_attributes.#": acctest.Ct1,
						"projection_type":      "INCLUDE",
						"range_key":            "att4",
						"read_capacity":        acctest.Ct1,
						"write_capacity":       acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "global_secondary_index.*.non_key_attributes.*", "RandomAttribute"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att1",
						names.AttrName:         "att1-index",
						"non_key_attributes.#": acctest.Ct0,
						"projection_type":      "ALL",
						"range_key":            "",
						"read_capacity":        acctest.Ct1,
						"write_capacity":       acctest.Ct1,
					}),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/15115
func TestAccDynamoDBTable_lsiNonKeyAttributes(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_lsiNonKeyAttributes(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "local_secondary_index.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "local_secondary_index.*", map[string]string{
						names.AttrName:         "TestTableLSI",
						"non_key_attributes.#": acctest.Ct1,
						"non_key_attributes.0": "TestNonKeyAttribute",
						"projection_type":      "INCLUDE",
						"range_key":            "TestLSIRangeKey",
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

// https://github.com/hashicorp/terraform-provider-aws/issues/566
func TestAccDynamoDBTable_gsiUpdateNonKeyAttributes(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_gsiUpdatedOtherAttributes(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "global_secondary_index.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att4",
						names.AttrName:         "att2-index",
						"non_key_attributes.#": acctest.Ct0,
						"projection_type":      "ALL",
						"range_key":            "att2",
						"read_capacity":        acctest.Ct1,
						"write_capacity":       acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att3",
						names.AttrName:         "att3-index",
						"non_key_attributes.#": acctest.Ct1,
						"projection_type":      "INCLUDE",
						"range_key":            "att4",
						"read_capacity":        acctest.Ct1,
						"write_capacity":       acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "global_secondary_index.*.non_key_attributes.*", "RandomAttribute"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att1",
						names.AttrName:         "att1-index",
						"non_key_attributes.#": acctest.Ct0,
						"projection_type":      "ALL",
						"range_key":            "",
						"read_capacity":        acctest.Ct1,
						"write_capacity":       acctest.Ct1,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableConfig_gsiUpdatedNonKeyAttributes(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att4",
						names.AttrName:         "att2-index",
						"non_key_attributes.#": acctest.Ct0,
						"projection_type":      "ALL",
						"range_key":            "att2",
						"read_capacity":        acctest.Ct1,
						"write_capacity":       acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att3",
						names.AttrName:         "att3-index",
						"non_key_attributes.#": acctest.Ct2,
						"projection_type":      "INCLUDE",
						"range_key":            "att4",
						"read_capacity":        acctest.Ct1,
						"write_capacity":       acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "global_secondary_index.*.non_key_attributes.*", "RandomAttribute"),
					resource.TestCheckTypeSetElemAttr(resourceName, "global_secondary_index.*.non_key_attributes.*", "AnotherAttribute"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att1",
						names.AttrName:         "att1-index",
						"non_key_attributes.#": acctest.Ct0,
						"projection_type":      "ALL",
						"range_key":            "",
						"read_capacity":        acctest.Ct1,
						"write_capacity":       acctest.Ct1,
					}),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/671
func TestAccDynamoDBTable_GsiUpdateNonKeyAttributes_emptyPlan(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	attributes := fmt.Sprintf("%q, %q", "AnotherAttribute", "RandomAttribute")
	reorderedAttributes := fmt.Sprintf("%q, %q", "RandomAttribute", "AnotherAttribute")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_gsiMultipleNonKeyAttributes(rName, attributes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att1",
						names.AttrName:         "att1-index",
						"non_key_attributes.#": acctest.Ct2,
						"projection_type":      "INCLUDE",
						"range_key":            "att2",
						"read_capacity":        acctest.Ct1,
						"write_capacity":       acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "global_secondary_index.*.non_key_attributes.*", "AnotherAttribute"),
					resource.TestCheckTypeSetElemAttr(resourceName, "global_secondary_index.*.non_key_attributes.*", "RandomAttribute"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:             testAccTableConfig_gsiMultipleNonKeyAttributes(rName, reorderedAttributes),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TTL tests must be split since it can only be updated once per hour
// ValidationException: Time to live has been modified multiple times within a fixed interval
func TestAccDynamoDBTable_TTL_enabled(t *testing.T) {
	ctx := acctest.Context(t)
	var table awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_timeToLive(rName, rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &table),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ttl"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"attribute_name":  knownvalue.StringExact(rName),
							names.AttrEnabled: knownvalue.Bool(true),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:             testAccTableConfig_timeToLive_unset(rName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TTL tests must be split since it can only be updated once per hour
// ValidationException: Time to live has been modified multiple times within a fixed interval
func TestAccDynamoDBTable_TTL_disabled(t *testing.T) {
	ctx := acctest.Context(t)
	var table awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_timeToLive(rName, "", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &table),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ttl"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"attribute_name":  knownvalue.StringExact(""),
							names.AttrEnabled: knownvalue.Bool(false),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:             testAccTableConfig_timeToLive_unset(rName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TTL tests must be split since it can only be updated once per hour
// ValidationException: Time to live has been modified multiple times within a fixed interval
func TestAccDynamoDBTable_TTL_update(t *testing.T) {
	ctx := acctest.Context(t)
	var table awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_timeToLive(rName, "", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &table),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ttl"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"attribute_name":  knownvalue.StringExact(""),
							names.AttrEnabled: knownvalue.Bool(false),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableConfig_timeToLive(rName, rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &table),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ttl"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"attribute_name":  knownvalue.StringExact(rName),
							names.AttrEnabled: knownvalue.Bool(true),
						}),
					})),
				},
			},
		},
	})
}

func TestAccDynamoDBTable_TTL_validate(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccTableConfig_timeToLive(rName, "TestTTL", false),
				ExpectError: regexache.MustCompile(regexp.QuoteMeta(`Attribute "ttl[0].attribute_name" cannot be specified when "ttl[0].enabled" is "false"`)),
			},
			{
				Config:      testAccTableConfig_timeToLive(rName, "", true),
				ExpectError: regexache.MustCompile(regexp.QuoteMeta(`Attribute "ttl[0].attribute_name" cannot have an empty value`)),
			},
			{
				Config:      testAccTableConfig_TTL_missingAttributeName(rName, true),
				ExpectError: regexache.MustCompile(regexp.QuoteMeta(`Attribute "ttl[0].attribute_name" cannot have an empty value`)),
			},
		},
	})
}

func TestAccDynamoDBTable_attributeUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_oneAttribute(rName, "firstKey", "firstKey", "S"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // Attribute type change
				Config: testAccTableConfig_oneAttribute(rName, "firstKey", "firstKey", "N"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
				),
			},
			{ // New attribute addition (index update)
				Config: testAccTableConfig_twoAttributes(rName, "firstKey", "secondKey", "firstKey", "N", "secondKey", "S"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
				),
			},
			{ // Attribute removal (index update)
				Config: testAccTableConfig_oneAttribute(rName, "firstKey", "firstKey", "S"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
				),
			},
		},
	})
}

func TestAccDynamoDBTable_lsiUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_lsi(rName, "lsi-original"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // Change name of local secondary index
				Config: testAccTableConfig_lsi(rName, "lsi-changed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
				),
			},
		},
	})
}

func TestAccDynamoDBTable_attributeUpdateValidation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccTableConfig_oneAttribute(rName, "firstKey", "unusedKey", "S"),
				ExpectError: regexache.MustCompile(`attributes must be indexed. Unused attributes: \["unusedKey"\]`),
			},
			{
				Config:      testAccTableConfig_twoAttributes(rName, "firstKey", "secondKey", "firstUnused", "N", "secondUnused", "S"),
				ExpectError: regexache.MustCompile(`attributes must be indexed. Unused attributes: \["firstUnused"\ \"secondUnused\"]`),
			},
			{
				Config:      testAccTableConfig_unmatchedIndexes(rName, "firstUnused", "secondUnused"),
				ExpectError: regexache.MustCompile(`indexes must match a defined attribute. Unmatched indexes:`),
			},
		},
	})
}

func TestAccDynamoDBTable_encryption(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var confBYOK, confEncEnabled, confEncDisabled awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_initialStateEncryptionBYOK(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &confBYOK),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "server_side_encryption.0.kms_key_arn", kmsKeyResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableConfig_initialStateEncryptionAmazonCMK(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &confEncDisabled),
					testAccCheckTableNotRecreated(&confEncDisabled, &confBYOK),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", acctest.Ct0),
				),
			},
			{
				Config: testAccTableConfig_initialStateEncryptionAmazonCMK(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &confEncEnabled),
					testAccCheckTableNotRecreated(&confEncEnabled, &confEncDisabled),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.kms_key_arn", ""),
				),
			},
		},
	})
}

func TestAccDynamoDBTable_Replica_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var table awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3),
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_replica2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct2),
				),
			},
			{
				Config:            testAccTableConfig_replica2(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableConfig_replica0(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct0),
				),
			},
			{
				Config: testAccTableConfig_replica2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccDynamoDBTable_Replica_single(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3), // 3 due to shared test configuration
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_replica1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct1),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "dynamodb", fmt.Sprintf("table/%s", rName)),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]*regexp.Regexp{
						names.AttrARN: regexache.MustCompile(fmt.Sprintf(`:dynamodb:%s:`, acctest.AlternateRegion())),
					}),
				),
			},
			{
				Config:            testAccTableConfig_replica1(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableConfig_replica0(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct0),
				),
			},
			{
				Config: testAccTableConfig_replica1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct1),
				),
			},
			{
				Config:   testAccTableConfig_replica1(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccDynamoDBTable_Replica_singleStreamSpecification(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3), // 3 due to shared test configuration
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_replicaStreamSpecification(rName, true, "KEYS_ONLY"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct1),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "dynamodb", fmt.Sprintf("table/%s", rName)),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]*regexp.Regexp{
						names.AttrARN:       regexache.MustCompile(fmt.Sprintf(`:dynamodb:%s:.*table/%s`, acctest.AlternateRegion(), rName)),
						names.AttrStreamARN: regexache.MustCompile(fmt.Sprintf(`:dynamodb:%s:.*table/%s/stream`, acctest.AlternateRegion(), rName)),
						"stream_label":      regexache.MustCompile(`[0-9]+.*:[0-9]+`),
					}),
				),
			},
		},
	})
}

func TestAccDynamoDBTable_Replica_singleDefaultKeyEncrypted(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3), // 3 due to shared test configuration
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_replicaEncryptedDefault(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccTableConfig_replicaEncryptedDefault(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", acctest.CtTrue),
				),
			},
			{
				Config:   testAccTableConfig_replicaEncryptedDefault(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccDynamoDBTable_Replica_singleCMK(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	kmsKeyResourceName := "aws_kms_key.test"
	kmsKeyReplicaResourceName := "aws_kms_key.awsalternate"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3), // 3 due to shared test configuration
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_replicaCMK(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "replica.0.kms_key_arn", kmsKeyReplicaResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "server_side_encryption.0.kms_key_arn", kmsKeyResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccDynamoDBTable_Replica_doubleAddCMK(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	kmsKeyResourceName := "aws_kms_key.test"
	kmsKey1Replica1ResourceName := "aws_kms_key.awsalternate1"
	kmsKey2Replica1ResourceName := "aws_kms_key.awsalternate2"
	kmsKey1Replica2ResourceName := "aws_kms_key.awsthird1"
	kmsKey2Replica2ResourceName := "aws_kms_key.awsthird2"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3), // 3 due to shared test configuration
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_replicaAmazonManagedKey(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "replica.0.kms_key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "replica.1.kms_key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.kms_key_arn", ""),
				),
			},
			{
				Config: testAccTableConfig_replicaCMKUpdate(rName, "awsalternate1", "awsthird1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct2),
					resource.TestCheckResourceAttrPair(resourceName, "replica.0.kms_key_arn", kmsKey1Replica1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "replica.1.kms_key_arn", kmsKey1Replica2ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "server_side_encryption.0.kms_key_arn", kmsKeyResourceName, names.AttrARN),
				),
			},
			{
				Config: testAccTableConfig_replicaCMKUpdate(rName, "awsalternate2", "awsthird2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct2),
					resource.TestCheckResourceAttrPair(resourceName, "replica.0.kms_key_arn", kmsKey2Replica1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "replica.1.kms_key_arn", kmsKey2Replica2ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "server_side_encryption.0.kms_key_arn", kmsKeyResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccDynamoDBTable_Replica_pitr(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf, replica1, replica2, replica3, replica4 awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3), // 3 due to shared test configuration
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_replicaPITR(rName, false, true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					testAccCheckReplicaExists(ctx, resourceName, acctest.AlternateRegion(), &replica1),
					testAccCheckReplicaExists(ctx, resourceName, acctest.ThirdRegion(), &replica2),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"point_in_time_recovery": acctest.CtTrue,
						"region_name":            acctest.AlternateRegion(),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"point_in_time_recovery": acctest.CtFalse,
						"region_name":            acctest.ThirdRegion(),
					}),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.0.enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccTableConfig_replicaPITR(rName, true, false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					testAccCheckReplicaExists(ctx, resourceName, acctest.AlternateRegion(), &replica3),
					testAccCheckReplicaExists(ctx, resourceName, acctest.ThirdRegion(), &replica4),
					testAccCheckTableNotRecreated(&replica1, &replica3),
					testAccCheckTableNotRecreated(&replica2, &replica4),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"point_in_time_recovery": acctest.CtFalse,
						"region_name":            acctest.AlternateRegion(),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"point_in_time_recovery": acctest.CtTrue,
						"region_name":            acctest.ThirdRegion(),
					}),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.0.enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccDynamoDBTable_Replica_pitrKMS(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf, replica1, replica2, replica3, replica4 awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3), // 3 due to shared test configuration
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_replicaPITRKMS(rName, false, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					testAccCheckReplicaExists(ctx, resourceName, acctest.AlternateRegion(), &replica1),
					testAccCheckReplicaExists(ctx, resourceName, acctest.ThirdRegion(), &replica2),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"point_in_time_recovery": acctest.CtFalse,
						"region_name":            acctest.AlternateRegion(),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"point_in_time_recovery": acctest.CtFalse,
						"region_name":            acctest.ThirdRegion(),
					}),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.0.enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccTableConfig_replicaPITRKMS(rName, false, true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					testAccCheckReplicaExists(ctx, resourceName, acctest.AlternateRegion(), &replica3),
					testAccCheckReplicaExists(ctx, resourceName, acctest.ThirdRegion(), &replica4),
					testAccCheckTableNotRecreated(&replica1, &replica3),
					testAccCheckTableNotRecreated(&replica2, &replica4),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"point_in_time_recovery": acctest.CtTrue,
						"region_name":            acctest.AlternateRegion(),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"point_in_time_recovery": acctest.CtFalse,
						"region_name":            acctest.ThirdRegion(),
					}),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.0.enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccTableConfig_replicaPITRKMS(rName, false, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					testAccCheckReplicaExists(ctx, resourceName, acctest.AlternateRegion(), &replica1),
					testAccCheckReplicaExists(ctx, resourceName, acctest.ThirdRegion(), &replica2),
					testAccCheckTableNotRecreated(&replica1, &replica3),
					testAccCheckTableNotRecreated(&replica2, &replica4),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"point_in_time_recovery": acctest.CtTrue,
						"region_name":            acctest.AlternateRegion(),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"point_in_time_recovery": acctest.CtTrue,
						"region_name":            acctest.ThirdRegion(),
					}),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.0.enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccTableConfig_replicaPITRKMS(rName, true, false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					testAccCheckReplicaExists(ctx, resourceName, acctest.AlternateRegion(), &replica3),
					testAccCheckReplicaExists(ctx, resourceName, acctest.ThirdRegion(), &replica4),
					testAccCheckTableNotRecreated(&replica1, &replica3),
					testAccCheckTableNotRecreated(&replica2, &replica4),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"point_in_time_recovery": acctest.CtFalse,
						"region_name":            acctest.AlternateRegion(),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"point_in_time_recovery": acctest.CtTrue,
						"region_name":            acctest.ThirdRegion(),
					}),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.0.enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccTableConfig_replicaPITRKMS(rName, false, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					testAccCheckReplicaExists(ctx, resourceName, acctest.AlternateRegion(), &replica1),
					testAccCheckReplicaExists(ctx, resourceName, acctest.ThirdRegion(), &replica2),
					testAccCheckTableNotRecreated(&replica1, &replica3),
					testAccCheckTableNotRecreated(&replica2, &replica4),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"point_in_time_recovery": acctest.CtFalse,
						"region_name":            acctest.AlternateRegion(),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"point_in_time_recovery": acctest.CtFalse,
						"region_name":            acctest.ThirdRegion(),
					}),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.0.enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccDynamoDBTable_Replica_tagsOneOfTwo(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3),
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_replicaTags(rName, "benny", "smiles", true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					testAccCheckReplicaHasTags(ctx, resourceName, acctest.AlternateRegion(), 3),
					testAccCheckReplicaHasTags(ctx, resourceName, acctest.ThirdRegion(), 0),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"region_name":           acctest.AlternateRegion(),
						names.AttrPropagateTags: acctest.CtTrue,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"region_name":           acctest.ThirdRegion(),
						names.AttrPropagateTags: acctest.CtFalse,
					}),
				),
			},
		},
	})
}

func TestAccDynamoDBTable_Replica_tagsTwoOfTwo(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3),
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_replicaTags(rName, "Structure", "Adobe", true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					testAccCheckReplicaHasTags(ctx, resourceName, acctest.AlternateRegion(), 3),
					testAccCheckReplicaHasTags(ctx, resourceName, acctest.ThirdRegion(), 3),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"region_name":           acctest.AlternateRegion(),
						names.AttrPropagateTags: acctest.CtTrue,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"region_name":           acctest.ThirdRegion(),
						names.AttrPropagateTags: acctest.CtTrue,
					}),
				),
			},
		},
	})
}

func TestAccDynamoDBTable_Replica_tagsNext(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3),
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_replicaTagsNext1(rName, acctest.AlternateRegion(), true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					testAccCheckReplicaHasTags(ctx, resourceName, acctest.AlternateRegion(), 2),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"region_name":           acctest.AlternateRegion(),
						names.AttrPropagateTags: acctest.CtTrue,
					}),
				),
			},
			{
				Config: testAccTableConfig_replicaTagsNext2(rName, acctest.AlternateRegion(), true, acctest.ThirdRegion(), true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					testAccCheckReplicaHasTags(ctx, resourceName, acctest.AlternateRegion(), 2),
					testAccCheckReplicaHasTags(ctx, resourceName, acctest.ThirdRegion(), 2),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"region_name":           acctest.AlternateRegion(),
						names.AttrPropagateTags: acctest.CtTrue,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"region_name":           acctest.ThirdRegion(),
						names.AttrPropagateTags: acctest.CtTrue,
					}),
				),
			},
			{
				Config: testAccTableConfig_replicaTagsNext1(rName, acctest.AlternateRegion(), true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					testAccCheckReplicaHasTags(ctx, resourceName, acctest.AlternateRegion(), 2),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"region_name":           acctest.AlternateRegion(),
						names.AttrPropagateTags: acctest.CtTrue,
					}),
				),
			},
			{
				Config: testAccTableConfig_replicaTagsNext2(rName, acctest.AlternateRegion(), true, acctest.ThirdRegion(), false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					testAccCheckReplicaHasTags(ctx, resourceName, acctest.AlternateRegion(), 2),
					testAccCheckReplicaHasTags(ctx, resourceName, acctest.ThirdRegion(), 0),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"region_name":           acctest.AlternateRegion(),
						names.AttrPropagateTags: acctest.CtTrue,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"region_name":           acctest.ThirdRegion(),
						names.AttrPropagateTags: acctest.CtFalse,
					}),
				),
			},
			{
				Config: testAccTableConfig_replicaTagsNext2(rName, acctest.AlternateRegion(), false, acctest.ThirdRegion(), false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					testAccCheckReplicaHasTags(ctx, resourceName, acctest.AlternateRegion(), 2),
					testAccCheckReplicaHasTags(ctx, resourceName, acctest.ThirdRegion(), 0),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"region_name":           acctest.AlternateRegion(),
						names.AttrPropagateTags: acctest.CtFalse,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"region_name":           acctest.ThirdRegion(),
						names.AttrPropagateTags: acctest.CtFalse,
					}),
				),
			},
		},
	})
}

func TestAccDynamoDBTable_Replica_tagsUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3),
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_replicaTagsUpdate1(rName, acctest.AlternateRegion()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					testAccCheckReplicaHasTags(ctx, resourceName, acctest.AlternateRegion(), 2),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"region_name":           acctest.AlternateRegion(),
						names.AttrPropagateTags: acctest.CtTrue,
					}),
				),
			},
			{
				Config: testAccTableConfig_replicaTagsUpdate2(rName, acctest.AlternateRegion()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					testAccCheckReplicaHasTags(ctx, resourceName, acctest.AlternateRegion(), 4),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"region_name":           acctest.AlternateRegion(),
						names.AttrPropagateTags: acctest.CtTrue,
					}),
				),
			},
			{
				Config: testAccTableConfig_replicaTagsUpdate3(rName, acctest.AlternateRegion(), acctest.ThirdRegion()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					testAccCheckReplicaHasTags(ctx, resourceName, acctest.AlternateRegion(), 4),
					testAccCheckReplicaHasTags(ctx, resourceName, acctest.ThirdRegion(), 4),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"region_name":           acctest.AlternateRegion(),
						names.AttrPropagateTags: acctest.CtTrue,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"region_name":           acctest.ThirdRegion(),
						names.AttrPropagateTags: acctest.CtTrue,
					}),
				),
			},
			{
				Config: testAccTableConfig_replicaTagsUpdate4(rName, acctest.AlternateRegion(), acctest.ThirdRegion()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					testAccCheckReplicaHasTags(ctx, resourceName, acctest.AlternateRegion(), 6),
					testAccCheckReplicaHasTags(ctx, resourceName, acctest.ThirdRegion(), 6),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"region_name":           acctest.AlternateRegion(),
						names.AttrPropagateTags: acctest.CtTrue,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"region_name":           acctest.ThirdRegion(),
						names.AttrPropagateTags: acctest.CtTrue,
					}),
				),
			},
			{
				Config: testAccTableConfig_replicaTagsUpdate5(rName, acctest.AlternateRegion(), acctest.ThirdRegion()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					testAccCheckReplicaHasTags(ctx, resourceName, acctest.AlternateRegion(), 1),
					testAccCheckReplicaHasTags(ctx, resourceName, acctest.ThirdRegion(), 1),
					resource.TestCheckResourceAttr(resourceName, "replica.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"region_name":           acctest.AlternateRegion(),
						names.AttrPropagateTags: acctest.CtTrue,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"region_name":           acctest.ThirdRegion(),
						names.AttrPropagateTags: acctest.CtTrue,
					}),
				),
			},
		},
	})
}

func TestAccDynamoDBTable_tableClassInfrequentAccess(t *testing.T) {
	ctx := acctest.Context(t)
	var table awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_class(rName, "STANDARD_INFREQUENT_ACCESS"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "table_class", "STANDARD_INFREQUENT_ACCESS"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableConfig_class(rName, "STANDARD"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "table_class", "STANDARD"),
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

func TestAccDynamoDBTable_tableClassExplicitDefault(t *testing.T) {
	ctx := acctest.Context(t)
	var table awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "table_class", "STANDARD"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:   testAccTableConfig_class(rName, "STANDARD"),
				PlanOnly: true,
			},
		},
	})
}

func TestAccDynamoDBTable_tableClass_ConcurrentModification(t *testing.T) {
	ctx := acctest.Context(t)
	var table awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_classConcurrent(rName, "STANDARD_INFREQUENT_ACCESS", 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "table_class", "STANDARD_INFREQUENT_ACCESS"),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableConfig_classConcurrent(rName, "STANDARD", 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "table_class", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "5"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "5"),
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

func TestAccDynamoDBTable_tableClass_migrate(t *testing.T) {
	ctx := acctest.Context(t)
	var table awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.DynamoDBServiceID),
		CheckDestroy: testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "4.57.0",
					},
				},
				Config: testAccTableConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "table_class", ""),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccTableConfig_basic(rName),
				PlanOnly:                 true,
			},
		},
	})
}

func TestAccDynamoDBTable_backupEncryption(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var confBYOK awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_backupInitialStateEncryption(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &confBYOK),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "server_side_encryption.0.kms_key_arn", kmsKeyResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"restore_to_latest_time",
					"restore_date_time",
					"restore_source_name",
				},
			},
		},
	})
}

func TestAccDynamoDBTable_backup_overrideEncryption(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var confBYOK awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_backupInitialStateOverrideEncryption(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &confBYOK),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "server_side_encryption.0.kms_key_arn", kmsKeyResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"restore_to_latest_time",
					"restore_date_time",
					"restore_source_name",
				},
			},
		},
	})
}

// lintignore:AT002
func TestAccDynamoDBTable_importTable(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf awstypes.TableDescription
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_import(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "dynamodb", fmt.Sprintf("table/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "hash_key", rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
						names.AttrName: rName,
						names.AttrType: "S",
					}),
					resource.TestCheckResourceAttr(resourceName, "table_class", "STANDARD"),
				),
			},
		},
	})
}

func testAccCheckTableDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dynamodb_table" {
				continue
			}

			_, err := tfdynamodb.FindTableByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DynamoDB Table %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckInitialTableExists(ctx context.Context, n string, v *awstypes.TableDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBClient(ctx)

		output, err := tfdynamodb.FindTableByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTableNotRecreated(i, j *awstypes.TableDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !i.CreationDateTime.Equal(aws.ToTime(j.CreationDateTime)) {
			return errors.New("DynamoDB Table was recreated")
		}

		return nil
	}
}

func testAccCheckReplicaExists(ctx context.Context, n string, region string, v *awstypes.TableDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBClient(ctx)

		output, err := tfdynamodb.FindTableByName(ctx, conn, rs.Primary.ID, func(o *dynamodb.Options) {
			o.Region = region
		})

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckReplicaHasTags(ctx context.Context, n string, region string, count int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBClient(ctx)

		newARN, err := tfdynamodb.ARNForNewRegion(rs.Primary.Attributes[names.AttrARN], region)

		if err != nil {
			return err
		}

		tags, err := tfdynamodb.ListTags(ctx, conn, newARN, func(o *dynamodb.Options) {
			o.Region = region
		})

		if err != nil && !tfawserr.ErrMessageContains(err, "UnknownOperationException", "Tagging is not currently supported in DynamoDB Local.") {
			return create.Error(names.DynamoDB, create.ErrActionChecking, "Table", rs.Primary.Attributes[names.AttrARN], err)
		}

		if len(tags.Keys()) != count {
			return create.Error(names.DynamoDB, create.ErrActionChecking, "Table", rs.Primary.Attributes[names.AttrARN], fmt.Errorf("expected %d tags on replica, found %d", count, len(tags.Keys())))
		}

		return nil
	}
}

func testAccCheckInitialTableConf(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "hash_key", "TestTableHashKey"),
		resource.TestCheckResourceAttr(resourceName, "range_key", "TestTableRangeKey"),
		resource.TestCheckResourceAttr(resourceName, "billing_mode", string(awstypes.BillingModeProvisioned)),
		resource.TestCheckResourceAttr(resourceName, "write_capacity", acctest.Ct2),
		resource.TestCheckResourceAttr(resourceName, "read_capacity", acctest.Ct1),
		resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", acctest.Ct0),
		resource.TestCheckResourceAttr(resourceName, "attribute.#", acctest.Ct4),
		resource.TestCheckResourceAttr(resourceName, "global_secondary_index.#", acctest.Ct1),
		resource.TestCheckResourceAttr(resourceName, "local_secondary_index.#", acctest.Ct1),
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
			names.AttrName: "TestTableHashKey",
			names.AttrType: "S",
		}),
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
			names.AttrName: "TestTableRangeKey",
			names.AttrType: "S",
		}),
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
			names.AttrName: "TestLSIRangeKey",
			names.AttrType: "N",
		}),
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
			names.AttrName: "TestGSIRangeKey",
			names.AttrType: "S",
		}),
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
			names.AttrName:    "InitialTestTableGSI",
			"hash_key":        "TestTableHashKey",
			"range_key":       "TestGSIRangeKey",
			"write_capacity":  acctest.Ct1,
			"read_capacity":   acctest.Ct1,
			"projection_type": "KEYS_ONLY",
		}),
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "local_secondary_index.*", map[string]string{
			names.AttrName:    "TestTableLSI",
			"range_key":       "TestLSIRangeKey",
			"projection_type": "ALL",
		}),
	)
}

func testAccTableConfig_basic(rName string) string {
	return fmt.Sprintf(`
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
`, rName)
}

func testAccTableConfig_enable_deletion_protection(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name                        = %[1]q
  read_capacity               = 1
  write_capacity              = 1
  hash_key                    = %[1]q
  deletion_protection_enabled = true

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, rName)
}

func testAccTableConfig_disable_deletion_protection(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name                        = %[1]q
  read_capacity               = 1
  write_capacity              = 1
  hash_key                    = %[1]q
  deletion_protection_enabled = false
  attribute {
    name = %[1]q
    type = "S"
  }
}
`, rName)
}

func testAccTableConfig_backup(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  point_in_time_recovery {
    enabled = true
  }
}
`, rName)
}

func testAccTableConfig_billingPayPerRequest(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }
}
`, rName)
}

func testAccTableConfig_billingPayPerRequestIgnoreChanges(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  lifecycle {
    ignore_changes = [read_capacity, write_capacity]
  }
}
`, rName)
}

func testAccTableConfig_billingProvisioned(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  billing_mode = "PROVISIONED"
  hash_key     = "TestTableHashKey"

  read_capacity  = 5
  write_capacity = 5

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }
}
`, rName)
}

func testAccTableConfig_billingProvisionedIgnoreChanges(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  billing_mode = "PROVISIONED"
  hash_key     = "TestTableHashKey"

  read_capacity  = 5
  write_capacity = 5

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  lifecycle {
    ignore_changes = [read_capacity, write_capacity]
  }
}
`, rName)
}

func testAccTableConfig_billingPayPerRequestGSI(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  attribute {
    name = "TestTableGSIKey"
    type = "S"
  }

  global_secondary_index {
    name            = "TestTableGSI"
    hash_key        = "TestTableGSIKey"
    projection_type = "KEYS_ONLY"
  }
}
`, rName)
}

func testAccTableConfig_billingProvisionedGSI(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  billing_mode   = "PROVISIONED"
  hash_key       = "TestTableHashKey"
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  attribute {
    name = "TestTableGSIKey"
    type = "S"
  }

  global_secondary_index {
    hash_key        = "TestTableGSIKey"
    name            = "TestTableGSI"
    projection_type = "KEYS_ONLY"
    read_capacity   = 1
    write_capacity  = 1
  }
}
`, rName)
}

func testAccTableConfig_initialState(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 2
  hash_key       = "TestTableHashKey"
  range_key      = "TestTableRangeKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  attribute {
    name = "TestTableRangeKey"
    type = "S"
  }

  attribute {
    name = "TestLSIRangeKey"
    type = "N"
  }

  attribute {
    name = "TestGSIRangeKey"
    type = "S"
  }

  local_secondary_index {
    name            = "TestTableLSI"
    range_key       = "TestLSIRangeKey"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "InitialTestTableGSI"
    hash_key        = "TestTableHashKey"
    range_key       = "TestGSIRangeKey"
    write_capacity  = 1
    read_capacity   = 1
    projection_type = "KEYS_ONLY"
  }
}
`, rName)
}

func testAccTableConfig_initialStateEncryptionAmazonCMK(rName string, enabled bool) string {
	return fmt.Sprintf(`
data "aws_kms_alias" "dynamodb" {
  name = "alias/aws/dynamodb"
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  server_side_encryption {
    enabled = %[2]t
  }
}
`, rName, enabled)
}

func testAccTableConfig_initialStateEncryptionBYOK(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 2
  write_capacity = 2
  hash_key       = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  server_side_encryption {
    enabled     = true
    kms_key_arn = aws_kms_key.test.arn
  }
}
`, rName)
}

func testAccTableConfig_addSecondaryGSI(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 2
  write_capacity = 2
  hash_key       = "TestTableHashKey"
  range_key      = "TestTableRangeKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  attribute {
    name = "TestTableRangeKey"
    type = "S"
  }

  attribute {
    name = "TestLSIRangeKey"
    type = "N"
  }

  attribute {
    name = "ReplacementGSIRangeKey"
    type = "N"
  }

  local_secondary_index {
    name            = "TestTableLSI"
    range_key       = "TestLSIRangeKey"
    projection_type = "ALL"
  }

  global_secondary_index {
    name               = "ReplacementTestTableGSI"
    hash_key           = "TestTableHashKey"
    range_key          = "ReplacementGSIRangeKey"
    write_capacity     = 5
    read_capacity      = 5
    projection_type    = "INCLUDE"
    non_key_attributes = ["TestNonKeyAttribute"]
  }
}
`, rName)
}

func testAccTableConfig_streamSpecification(rName string, enabled bool, viewType string) string {
	if viewType != "null" {
		viewType = fmt.Sprintf(`"%s"`, viewType)
	}
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 2
  hash_key       = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  stream_enabled   = %[2]t
  stream_view_type = %[3]s
}
`, rName, enabled, viewType)
}

func testAccTableConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 2
  hash_key       = "TestTableHashKey"
  range_key      = "TestTableRangeKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  attribute {
    name = "TestTableRangeKey"
    type = "S"
  }

  attribute {
    name = "TestLSIRangeKey"
    type = "N"
  }

  attribute {
    name = "TestGSIRangeKey"
    type = "S"
  }

  local_secondary_index {
    name            = "TestTableLSI"
    range_key       = "TestLSIRangeKey"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "InitialTestTableGSI"
    hash_key        = "TestTableHashKey"
    range_key       = "TestGSIRangeKey"
    write_capacity  = 1
    read_capacity   = 1
    projection_type = "KEYS_ONLY"
  }

  tags = {
    Name    = %[1]q
    AccTest = "yes"
    Testing = "absolutely"
  }
}
`, rName)
}

func testAccTableConfig_gsiUpdate(rName string) string {
	return fmt.Sprintf(`
variable "capacity" {
  default = 1
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = var.capacity
  write_capacity = var.capacity
  hash_key       = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "att1"
    type = "S"
  }

  attribute {
    name = "att2"
    type = "S"
  }

  attribute {
    name = "att3"
    type = "S"
  }

  global_secondary_index {
    name            = "att1-index"
    hash_key        = "att1"
    write_capacity  = var.capacity
    read_capacity   = var.capacity
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att2-index"
    hash_key        = "att2"
    write_capacity  = var.capacity
    read_capacity   = var.capacity
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att3-index"
    hash_key        = "att3"
    write_capacity  = var.capacity
    read_capacity   = var.capacity
    projection_type = "ALL"
  }
}
`, rName)
}

func testAccTableConfig_gsiUpdatedCapacity(rName string) string {
	return fmt.Sprintf(`
variable "capacity" {
  default = 2
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = var.capacity
  write_capacity = var.capacity
  hash_key       = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "att1"
    type = "S"
  }

  attribute {
    name = "att2"
    type = "S"
  }

  attribute {
    name = "att3"
    type = "S"
  }

  global_secondary_index {
    name            = "att1-index"
    hash_key        = "att1"
    write_capacity  = var.capacity
    read_capacity   = var.capacity
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att2-index"
    hash_key        = "att2"
    write_capacity  = var.capacity
    read_capacity   = var.capacity
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att3-index"
    hash_key        = "att3"
    write_capacity  = var.capacity
    read_capacity   = var.capacity
    projection_type = "ALL"
  }
}
`, rName)
}

func testAccTableConfig_gsiUpdatedOtherAttributes(rName string) string {
	return fmt.Sprintf(`
variable "capacity" {
  default = 1
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = var.capacity
  write_capacity = var.capacity
  hash_key       = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "att1"
    type = "S"
  }

  attribute {
    name = "att2"
    type = "S"
  }

  attribute {
    name = "att3"
    type = "S"
  }

  attribute {
    name = "att4"
    type = "S"
  }

  global_secondary_index {
    name            = "att1-index"
    hash_key        = "att1"
    write_capacity  = var.capacity
    read_capacity   = var.capacity
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att2-index"
    hash_key        = "att4"
    range_key       = "att2"
    write_capacity  = var.capacity
    read_capacity   = var.capacity
    projection_type = "ALL"
  }

  global_secondary_index {
    name               = "att3-index"
    hash_key           = "att3"
    range_key          = "att4"
    write_capacity     = var.capacity
    read_capacity      = var.capacity
    projection_type    = "INCLUDE"
    non_key_attributes = ["RandomAttribute"]
  }
}
`, rName)
}

func testAccTableConfig_gsiUpdatedNonKeyAttributes(rName string) string {
	return fmt.Sprintf(`
variable "capacity" {
  default = 1
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = var.capacity
  write_capacity = var.capacity
  hash_key       = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "att1"
    type = "S"
  }

  attribute {
    name = "att2"
    type = "S"
  }

  attribute {
    name = "att3"
    type = "S"
  }

  attribute {
    name = "att4"
    type = "S"
  }

  global_secondary_index {
    name            = "att1-index"
    hash_key        = "att1"
    write_capacity  = var.capacity
    read_capacity   = var.capacity
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att2-index"
    hash_key        = "att4"
    range_key       = "att2"
    write_capacity  = var.capacity
    read_capacity   = var.capacity
    projection_type = "ALL"
  }

  global_secondary_index {
    name               = "att3-index"
    hash_key           = "att3"
    range_key          = "att4"
    write_capacity     = var.capacity
    read_capacity      = var.capacity
    projection_type    = "INCLUDE"
    non_key_attributes = ["RandomAttribute", "AnotherAttribute"]
  }
}
`, rName)
}

func testAccTableConfig_gsiMultipleNonKeyAttributes(rName, attributes string) string {
	return fmt.Sprintf(`
variable "capacity" {
  default = 1
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = var.capacity
  write_capacity = var.capacity
  hash_key       = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "att1"
    type = "S"
  }

  attribute {
    name = "att2"
    type = "S"
  }

  global_secondary_index {
    name               = "att1-index"
    hash_key           = "att1"
    range_key          = "att2"
    write_capacity     = var.capacity
    read_capacity      = var.capacity
    projection_type    = "INCLUDE"
    non_key_attributes = [%s]
  }
}
`, rName, attributes)
}

func testAccTableConfig_lsiNonKeyAttributes(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  hash_key       = "TestTableHashKey"
  range_key      = "TestTableRangeKey"
  write_capacity = 1
  read_capacity  = 1

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  attribute {
    name = "TestTableRangeKey"
    type = "S"
  }

  attribute {
    name = "TestLSIRangeKey"
    type = "N"
  }

  local_secondary_index {
    name               = "TestTableLSI"
    range_key          = "TestLSIRangeKey"
    projection_type    = "INCLUDE"
    non_key_attributes = ["TestNonKeyAttribute"]
  }
}
`, rName)
}

func testAccTableConfig_timeToLive(rName, ttlAttribute string, ttlEnabled bool) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  hash_key       = "TestTableHashKey"
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  ttl {
    attribute_name = %[2]q
    enabled        = %[3]t
  }
}
`, rName, ttlAttribute, ttlEnabled)
}

func testAccTableConfig_timeToLive_unset(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  hash_key       = "TestTableHashKey"
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }
}
`, rName)
}

func testAccTableConfig_TTL_missingAttributeName(rName string, ttlEnabled bool) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  hash_key       = "TestTableHashKey"
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  ttl {
    attribute_name = ""
    enabled        = %[2]t
  }
}
`, rName, ttlEnabled)
}

func testAccTableConfig_oneAttribute(rName, hashKey, attrName, attrType string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 10
  write_capacity = 10
  hash_key       = "staticHashKey"

  attribute {
    name = "staticHashKey"
    type = "S"
  }

  attribute {
    name = %[3]q
    type = %[4]q
  }

  global_secondary_index {
    name            = "gsiName"
    hash_key        = %[2]q
    write_capacity  = 10
    read_capacity   = 10
    projection_type = "KEYS_ONLY"
  }
}
`, rName, hashKey, attrName, attrType)
}

func testAccTableConfig_twoAttributes(rName, hashKey, rangeKey, attrName1, attrType1, attrName2, attrType2 string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 10
  write_capacity = 10
  hash_key       = "staticHashKey"

  attribute {
    name = "staticHashKey"
    type = "S"
  }

  attribute {
    name = %[4]q
    type = %[5]q
  }

  attribute {
    name = %[6]q
    type = %[7]q
  }

  global_secondary_index {
    name            = "gsiName"
    hash_key        = %[2]q
    range_key       = %[3]q
    write_capacity  = 10
    read_capacity   = 10
    projection_type = "KEYS_ONLY"
  }
}
`, rName, hashKey, rangeKey, attrName1, attrType1, attrName2, attrType2)
}

func testAccTableConfig_unmatchedIndexes(rName, attr1, attr2 string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 10
  write_capacity = 10
  hash_key       = "staticHashKey"
  range_key      = %[2]q

  attribute {
    name = "staticHashKey"
    type = "S"
  }

  local_secondary_index {
    name            = "lsiName"
    range_key       = %[3]q
    projection_type = "KEYS_ONLY"
  }
}
`, rName, attr1, attr2)
}

func testAccTableConfig_replica0(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3), // Prevent "Provider configuration not present" errors
		fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name             = %[1]q
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }
}
`, rName))
}

func testAccTableConfig_replica1(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3), // Prevent "Provider configuration not present" errors
		fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = "awsalternate"
}

resource "aws_dynamodb_table" "test" {
  name             = %[1]q
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  replica {
    region_name = data.aws_region.alternate.name
  }
}
`, rName))
}

func testAccTableConfig_replicaEncryptedDefault(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3), // Prevent "Provider configuration not present" errors
		fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = "awsalternate"
}

resource "aws_dynamodb_table" "test" {
  name             = %[1]q
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  server_side_encryption {
    enabled = true
  }

  replica {
    region_name = data.aws_region.alternate.name
  }
}
`, rName))
}

func testAccTableConfig_replicaCMK(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3), // Prevent "Provider configuration not present" errors
		fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = "awsalternate"
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_kms_key" "awsalternate" {
  provider                = "awsalternate"
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_dynamodb_table" "test" {
  name             = %[1]q
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  replica {
    region_name = data.aws_region.alternate.name
    kms_key_arn = aws_kms_key.awsalternate.arn
  }

  server_side_encryption {
    enabled     = true
    kms_key_arn = aws_kms_key.test.arn
  }

  timeouts {
    create = "20m"
    update = "20m"
    delete = "20m"
  }
}
`, rName))
}

func testAccTableConfig_replicaAmazonManagedKey(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3), // Prevent "Provider configuration not present" errors
		fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = "awsalternate"
}

data "aws_region" "third" {
  provider = "awsthird"
}

resource "aws_dynamodb_table" "test" {
  name             = %[1]q
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  replica {
    region_name = data.aws_region.alternate.name
  }

  replica {
    region_name = data.aws_region.third.name
  }

  server_side_encryption {
    enabled = true
  }

  timeouts {
    create = "20m"
    update = "20m"
    delete = "20m"
  }
}
`, rName))
}

func testAccTableConfig_replicaCMKUpdate(rName, keyReplica1, keyReplica2 string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3), // Prevent "Provider configuration not present" errors
		fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = "awsalternate"
}

data "aws_region" "third" {
  provider = "awsthird"
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_kms_key" "awsalternate1" {
  provider                = "awsalternate"
  description             = "%[1]s-1"
  deletion_window_in_days = 7
}

resource "aws_kms_key" "awsalternate2" {
  provider                = "awsalternate"
  description             = "%[1]s-2"
  deletion_window_in_days = 7
}

resource "aws_kms_key" "awsthird1" {
  provider                = "awsthird"
  description             = "%[1]s-1"
  deletion_window_in_days = 7
}

resource "aws_kms_key" "awsthird2" {
  provider                = "awsthird"
  description             = "%[1]s-2"
  deletion_window_in_days = 7
}

resource "aws_dynamodb_table" "test" {
  name             = %[1]q
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  replica {
    region_name = data.aws_region.alternate.name
    kms_key_arn = aws_kms_key.%[2]s.arn
  }

  replica {
    region_name = data.aws_region.third.name
    kms_key_arn = aws_kms_key.%[3]s.arn
  }

  server_side_encryption {
    enabled     = true
    kms_key_arn = aws_kms_key.test.arn
  }

  timeouts {
    create = "20m"
    update = "20m"
    delete = "20m"
  }
}
`, rName, keyReplica1, keyReplica2))
}

func testAccTableConfig_replicaPITR(rName string, mainPITR, replica1, replica2 bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3), // Prevent "Provider configuration not present" errors
		fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = "awsalternate"
}

data "aws_region" "third" {
  provider = "awsthird"
}

resource "aws_dynamodb_table" "test" {
  name             = %[1]q
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  point_in_time_recovery {
    enabled = %[2]t
  }

  replica {
    region_name            = data.aws_region.alternate.name
    point_in_time_recovery = %[3]t
  }

  replica {
    region_name            = data.aws_region.third.name
    point_in_time_recovery = %[4]t
  }
}
`, rName, mainPITR, replica1, replica2))
}

func testAccTableConfig_replicaPITRKMS(rName string, mainPITR, replica1, replica2 bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = awsalternate
}

data "aws_region" "third" {
  provider = awsthird
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_kms_key" "alternate" {
  provider                = awsalternate
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_kms_key" "third" {
  provider                = awsthird
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_dynamodb_table" "test" {
  name             = %[1]q
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  point_in_time_recovery {
    enabled = %[2]t
  }

  server_side_encryption {
    enabled     = true
    kms_key_arn = aws_kms_key.test.arn
  }

  replica {
    region_name            = data.aws_region.alternate.name
    point_in_time_recovery = %[3]t
    kms_key_arn            = aws_kms_key.alternate.arn
  }

  replica {
    region_name            = data.aws_region.third.name
    point_in_time_recovery = %[4]t
    kms_key_arn            = aws_kms_key.third.arn
  }
}
`, rName, mainPITR, replica1, replica2))
}

func testAccTableConfig_replicaTags(rName, key, value string, propagate1, propagate2 bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = "awsalternate"
}

data "aws_region" "third" {
  provider = "awsthird"
}

resource "aws_dynamodb_table" "test" {
  name             = %[1]q
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  replica {
    region_name    = data.aws_region.alternate.name
    propagate_tags = %[4]t
  }

  replica {
    region_name    = data.aws_region.third.name
    propagate_tags = %[5]t
  }

  tags = {
    Name  = %[1]q
    Pozo  = "Amargo"
    %[2]s = %[3]q
  }
}
`, rName, key, value, propagate1, propagate2))
}

func testAccTableConfig_replica2(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = "awsalternate"
}

data "aws_region" "third" {
  provider = "awsthird"
}

resource "aws_dynamodb_table" "test" {
  name             = %[1]q
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  replica {
    region_name = data.aws_region.alternate.name
  }

  replica {
    region_name = data.aws_region.third.name
  }
}
`, rName))
}

func testAccTableConfig_replicaTagsNext1(rName string, region1 string, propagate1 bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name             = %[1]q
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  replica {
    region_name    = %[2]q
    propagate_tags = %[3]t
  }

  tags = {
    Name = %[1]q
    Pozo = "Amargo"
  }
}
`, rName, region1, propagate1))
}

func testAccTableConfig_replicaTagsNext2(rName, region1 string, propagate1 bool, region2 string, propagate2 bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name             = %[1]q
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  replica {
    region_name    = %[2]q
    propagate_tags = %[3]t
  }

  replica {
    region_name    = %[4]q
    propagate_tags = %[5]t
  }

  tags = {
    Name = %[1]q
    Pozo = "Amargo"
  }
}
`, rName, region1, propagate1, region2, propagate2))
}

func testAccTableConfig_replicaTagsUpdate1(rName, region1 string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name             = %[1]q
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  replica {
    region_name    = %[2]q
    propagate_tags = true
  }

  tags = {
    Name = %[1]q
    Pozo = "Amargo"
  }
}
`, rName, region1))
}

func testAccTableConfig_replicaTagsUpdate2(rName, region1 string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name             = %[1]q
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  replica {
    region_name    = %[2]q
    propagate_tags = true
  }

  tags = {
    Name   = %[1]q
    Pozo   = "Amargo"
    tyDi   = "Lullaby"
    Thrill = "Seekers"
  }
}
`, rName, region1))
}

func testAccTableConfig_replicaTagsUpdate3(rName, region1 string, region2 string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name             = %[1]q
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  replica {
    region_name    = %[2]q
    propagate_tags = true
  }

  replica {
    region_name    = %[3]q
    propagate_tags = true
  }

  tags = {
    Name   = %[1]q
    Pozo   = "Amargo"
    tyDi   = "Lullaby"
    Thrill = "Seekers"
  }
}
`, rName, region1, region2))
}

func testAccTableConfig_replicaTagsUpdate4(rName, region1 string, region2 string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name             = %[1]q
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  replica {
    region_name    = %[2]q
    propagate_tags = true
  }

  replica {
    region_name    = %[3]q
    propagate_tags = true
  }

  tags = {
    Name    = %[1]q
    Pozo    = "Amargo"
    tyDi    = "Lullaby"
    Thrill  = "Seekers"
    Tristan = "Joe"
    Humming = "bird"
  }
}
`, rName, region1, region2))
}

func testAccTableConfig_replicaTagsUpdate5(rName, region1 string, region2 string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name             = %[1]q
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  replica {
    region_name    = %[2]q
    propagate_tags = true
  }

  replica {
    region_name    = %[3]q
    propagate_tags = true
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, region1, region2))
}

func testAccTableConfig_replicaStreamSpecification(rName string, streamEnabled bool, viewType string) string {
	if viewType != "null" {
		viewType = fmt.Sprintf(`"%s"`, viewType)
	}
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = "awsalternate"
}

resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  hash_key     = "TestTableHashKey"
  billing_mode = "PAY_PER_REQUEST"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  replica {
    region_name = data.aws_region.alternate.name
  }

  stream_enabled   = %[2]t
  stream_view_type = %[3]s
}
`, rName, streamEnabled, viewType))
}

func testAccTableConfig_lsi(rName, lsiName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 10
  write_capacity = 10
  hash_key       = "staticHashKey"
  range_key      = "staticRangeKey"

  attribute {
    name = "staticHashKey"
    type = "S"
  }

  attribute {
    name = "staticRangeKey"
    type = "S"
  }

  attribute {
    name = "staticLSIRangeKey"
    type = "S"
  }

  local_secondary_index {
    name            = %[2]q
    range_key       = "staticLSIRangeKey"
    projection_type = "KEYS_ONLY"
  }
}
`, rName, lsiName)
}

func testAccTableConfig_class(rName, tableClass string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 1
  hash_key       = %[1]q

  table_class = %[2]q

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, rName, tableClass)
}

func testAccTableConfig_classConcurrent(rName, tableClass string, capacity int) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  hash_key       = "TestTableHashKey"
  name           = %[1]q
  read_capacity  = %[2]d
  write_capacity = %[2]d
  table_class    = %[3]q

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }
}
`, rName, capacity, tableClass)
}

func testAccTableConfig_backupInitialStateOverrideEncryption(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "source" {
  name           = "%[1]s-source"
  read_capacity  = 2
  write_capacity = 2
  hash_key       = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  point_in_time_recovery {
    enabled = true
  }

  server_side_encryption {
    enabled = false
  }
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_dynamodb_table" "test" {
  name                   = "%[1]s-target"
  restore_source_name    = aws_dynamodb_table.source.name
  restore_to_latest_time = true

  server_side_encryption {
    enabled     = true
    kms_key_arn = aws_kms_key.test.arn
  }
}
`, rName)
}

func testAccTableConfig_backupInitialStateEncryption(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "source" {
  name           = "%[1]s-source"
  read_capacity  = 2
  write_capacity = 2
  hash_key       = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  point_in_time_recovery {
    enabled = true
  }

  server_side_encryption {
    enabled     = true
    kms_key_arn = aws_kms_key.test.arn
  }
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_dynamodb_table" "test" {
  name                   = "%[1]s-target"
  restore_source_name    = aws_dynamodb_table.source.name
  restore_to_latest_time = true
}
`, rName)
}

func testAccTableConfig_import(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_object" "test" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "data/somedoc.json"
  content = "{\"Item\":{\"%[1]s\":{\"S\":\"test\"},\"field\":{\"S\":\"test\"}}}"
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

  import_table {
    input_compression_type = "NONE"
    input_format           = "DYNAMODB_JSON"
    s3_bucket_source {
      bucket     = aws_s3_bucket.test.bucket
      key_prefix = "data"
    }
  }
}
`, rName)
}
