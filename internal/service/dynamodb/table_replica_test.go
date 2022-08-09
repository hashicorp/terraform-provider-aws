package dynamodb_test

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdynamodb "github.com/hashicorp/terraform-provider-aws/internal/service/dynamodb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDynamoDBTableReplica_basic(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "dynamodb", fmt.Sprintf("table/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "hash_key", rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
						"name": rName,
						"type": "S",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "table_class", ""),
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

func TestAccDynamoDBTableReplica_disappears(t *testing.T) {
	var table1 dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &table1),
					acctest.CheckResourceDisappears(acctest.Provider, tfdynamodb.ResourceTable(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDynamoDBTableReplica_Disappears_payPerRequestWithGSI(t *testing.T) {
	var table1, table2 dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_billingPayPerRequestGSI(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &table1),
					acctest.CheckResourceDisappears(acctest.Provider, tfdynamodb.ResourceTable(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccTableReplicaConfig_billingPayPerRequestGSI(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &table2),
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

func TestAccDynamoDBTableReplica_extended(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_initialState(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					testAccCheckInitialTableReplicaConf(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableReplicaConfig_addSecondaryGSI(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "hash_key", "TestTableHashKey"),
					resource.TestCheckResourceAttr(resourceName, "range_key", "TestTableRangeKey"),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", dynamodb.BillingModeProvisioned),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "attribute.#", "4"),
					resource.TestCheckResourceAttr(resourceName, "global_secondary_index.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "local_secondary_index.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
						"name": "TestTableHashKey",
						"type": "S",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
						"name": "TestTableRangeKey",
						"type": "S",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
						"name": "TestLSIRangeKey",
						"type": "N",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
						"name": "ReplacementGSIRangeKey",
						"type": "N",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"name":                 "ReplacementTestTableGSI",
						"hash_key":             "TestTableHashKey",
						"range_key":            "ReplacementGSIRangeKey",
						"write_capacity":       "5",
						"read_capacity":        "5",
						"projection_type":      "INCLUDE",
						"non_key_attributes.#": "1",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "global_secondary_index.*.non_key_attributes.*", "TestNonKeyAttribute"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "local_secondary_index.*", map[string]string{
						"name":            "TestTableLSI",
						"range_key":       "TestLSIRangeKey",
						"projection_type": "ALL",
					}),
				),
			},
		},
	})
}

func TestAccDynamoDBTableReplica_enablePITR(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_initialState(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					testAccCheckInitialTableReplicaConf(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableReplicaConfig_backup(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.0.enabled", "true"),
				),
			},
		},
	})
}

func TestAccDynamoDBTableReplica_BillingMode_payPerRequestToProvisioned(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_billingPayPerRequest(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", dynamodb.BillingModePayPerRequest),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableReplicaConfig_billingProvisioned(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", dynamodb.BillingModeProvisioned),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "5"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "5"),
				),
			},
		},
	})
}

func TestAccDynamoDBTableReplica_BillingMode_payPerRequestToProvisionedIgnoreChanges(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_billingPayPerRequest(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", dynamodb.BillingModePayPerRequest),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableReplicaConfig_billingProvisionedIgnoreChanges(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", dynamodb.BillingModeProvisioned),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "1"),
				),
			},
		},
	})
}

func TestAccDynamoDBTableReplica_BillingMode_provisionedToPayPerRequest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_billingProvisioned(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", dynamodb.BillingModeProvisioned),
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
				Config: testAccTableReplicaConfig_billingPayPerRequest(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", dynamodb.BillingModePayPerRequest),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "0"),
				),
			},
		},
	})
}

func TestAccDynamoDBTableReplica_BillingMode_provisionedToPayPerRequestIgnoreChanges(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_billingProvisioned(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", dynamodb.BillingModeProvisioned),
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
				Config: testAccTableReplicaConfig_billingPayPerRequestIgnoreChanges(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", dynamodb.BillingModePayPerRequest),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "0"),
				),
			},
		},
	})
}

func TestAccDynamoDBTableReplica_BillingModeGSI_payPerRequestToProvisioned(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_billingPayPerRequestGSI(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", dynamodb.BillingModePayPerRequest),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableReplicaConfig_billingProvisionedGSI(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", dynamodb.BillingModeProvisioned),
				),
			},
		},
	})
}

func TestAccDynamoDBTableReplica_BillingModeGSI_provisionedToPayPerRequest(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_billingProvisionedGSI(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", dynamodb.BillingModeProvisioned),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableReplicaConfig_billingPayPerRequestGSI(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", dynamodb.BillingModePayPerRequest),
				),
			},
		},
	})
}

func TestAccDynamoDBTableReplica_streamSpecification(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_streamSpecification(rName, true, "KEYS_ONLY"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stream_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stream_view_type", "KEYS_ONLY"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "stream_arn", "dynamodb", regexp.MustCompile(fmt.Sprintf("table/%s/stream", rName))),
					resource.TestCheckResourceAttrSet(resourceName, "stream_label"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableReplicaConfig_streamSpecification(rName, false, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stream_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "stream_view_type", ""),
					acctest.MatchResourceAttrRegionalARN(resourceName, "stream_arn", "dynamodb", regexp.MustCompile(fmt.Sprintf("table/%s/stream", rName))),
					resource.TestCheckResourceAttrSet(resourceName, "stream_label"),
				),
			},
		},
	})
}

func TestAccDynamoDBTableReplica_streamSpecificationValidation(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTableReplicaConfig_streamSpecification("anything", true, ""),
				ExpectError: regexp.MustCompile(`stream_view_type is required when stream_enabled = true`),
			},
		},
	})
}

func TestAccDynamoDBTableReplica_tags(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_tags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					testAccCheckInitialTableReplicaConf(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
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
func TestAccDynamoDBTableReplica_gsiUpdateCapacity(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_gsiUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "global_secondary_index.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"read_capacity":  "1",
						"write_capacity": "1",
						"name":           "att1-index",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"read_capacity":  "1",
						"write_capacity": "1",
						"name":           "att2-index",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"read_capacity":  "1",
						"write_capacity": "1",
						"name":           "att3-index",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableReplicaConfig_gsiUpdatedCapacity(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "global_secondary_index.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"read_capacity":  "2",
						"write_capacity": "2",
						"name":           "att1-index",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"read_capacity":  "2",
						"write_capacity": "2",
						"name":           "att2-index",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"read_capacity":  "2",
						"write_capacity": "2",
						"name":           "att3-index",
					}),
				),
			},
		},
	})
}

func TestAccDynamoDBTableReplica_gsiUpdateOtherAttributes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_gsiUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "global_secondary_index.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att3",
						"name":                 "att3-index",
						"non_key_attributes.#": "0",
						"projection_type":      "ALL",
						"range_key":            "",
						"read_capacity":        "1",
						"write_capacity":       "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att1",
						"name":                 "att1-index",
						"non_key_attributes.#": "0",
						"projection_type":      "ALL",
						"range_key":            "",
						"read_capacity":        "1",
						"write_capacity":       "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att2",
						"name":                 "att2-index",
						"non_key_attributes.#": "0",
						"projection_type":      "ALL",
						"range_key":            "",
						"read_capacity":        "1",
						"write_capacity":       "1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableReplicaConfig_gsiUpdatedOtherAttributes(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "global_secondary_index.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att4",
						"name":                 "att2-index",
						"non_key_attributes.#": "0",
						"projection_type":      "ALL",
						"range_key":            "att2",
						"read_capacity":        "1",
						"write_capacity":       "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att3",
						"name":                 "att3-index",
						"non_key_attributes.#": "1",
						"projection_type":      "INCLUDE",
						"range_key":            "att4",
						"read_capacity":        "1",
						"write_capacity":       "1",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "global_secondary_index.*.non_key_attributes.*", "RandomAttribute"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att1",
						"name":                 "att1-index",
						"non_key_attributes.#": "0",
						"projection_type":      "ALL",
						"range_key":            "",
						"read_capacity":        "1",
						"write_capacity":       "1",
					}),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/15115
func TestAccDynamoDBTableReplica_lsiNonKeyAttributes(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_lsiNonKeyAttributes(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "local_secondary_index.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "local_secondary_index.*", map[string]string{
						"name":                 "TestTableLSI",
						"non_key_attributes.#": "1",
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
func TestAccDynamoDBTableReplica_gsiUpdateNonKeyAttributes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_gsiUpdatedOtherAttributes(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "global_secondary_index.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att4",
						"name":                 "att2-index",
						"non_key_attributes.#": "0",
						"projection_type":      "ALL",
						"range_key":            "att2",
						"read_capacity":        "1",
						"write_capacity":       "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att3",
						"name":                 "att3-index",
						"non_key_attributes.#": "1",
						"projection_type":      "INCLUDE",
						"range_key":            "att4",
						"read_capacity":        "1",
						"write_capacity":       "1",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "global_secondary_index.*.non_key_attributes.*", "RandomAttribute"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att1",
						"name":                 "att1-index",
						"non_key_attributes.#": "0",
						"projection_type":      "ALL",
						"range_key":            "",
						"read_capacity":        "1",
						"write_capacity":       "1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableReplicaConfig_gsiUpdatedNonKeyAttributes(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att4",
						"name":                 "att2-index",
						"non_key_attributes.#": "0",
						"projection_type":      "ALL",
						"range_key":            "att2",
						"read_capacity":        "1",
						"write_capacity":       "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att3",
						"name":                 "att3-index",
						"non_key_attributes.#": "2",
						"projection_type":      "INCLUDE",
						"range_key":            "att4",
						"read_capacity":        "1",
						"write_capacity":       "1",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "global_secondary_index.*.non_key_attributes.*", "RandomAttribute"),
					resource.TestCheckTypeSetElemAttr(resourceName, "global_secondary_index.*.non_key_attributes.*", "AnotherAttribute"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att1",
						"name":                 "att1-index",
						"non_key_attributes.#": "0",
						"projection_type":      "ALL",
						"range_key":            "",
						"read_capacity":        "1",
						"write_capacity":       "1",
					}),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/671
func TestAccDynamoDBTableReplica_GsiUpdateNonKeyAttributes_emptyPlan(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	attributes := fmt.Sprintf("%q, %q", "AnotherAttribute", "RandomAttribute")
	reorderedAttributes := fmt.Sprintf("%q, %q", "RandomAttribute", "AnotherAttribute")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_gsiMultipleNonKeyAttributes(rName, attributes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att1",
						"name":                 "att1-index",
						"non_key_attributes.#": "2",
						"projection_type":      "INCLUDE",
						"range_key":            "att2",
						"read_capacity":        "1",
						"write_capacity":       "1",
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
				Config:             testAccTableReplicaConfig_gsiMultipleNonKeyAttributes(rName, reorderedAttributes),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TTL tests must be split since it can only be updated once per hour
// ValidationException: Time to live has been modified multiple times within a fixed interval
func TestAccDynamoDBTableReplica_TTL_enabled(t *testing.T) {
	var table dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_timeToLive(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "ttl.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ttl.0.enabled", "true"),
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

// TTL tests must be split since it can only be updated once per hour
// ValidationException: Time to live has been modified multiple times within a fixed interval
func TestAccDynamoDBTableReplica_TTL_disabled(t *testing.T) {
	var table dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_timeToLive(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "ttl.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ttl.0.enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableReplicaConfig_timeToLive(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "ttl.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ttl.0.enabled", "true"),
				),
			},
		},
	})
}

func TestAccDynamoDBTableReplica_attributeUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_oneAttribute(rName, "firstKey", "firstKey", "S"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // Attribute type change
				Config: testAccTableReplicaConfig_oneAttribute(rName, "firstKey", "firstKey", "N"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
				),
			},
			{ // New attribute addition (index update)
				Config: testAccTableReplicaConfig_twoAttributes(rName, "firstKey", "secondKey", "firstKey", "N", "secondKey", "S"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
				),
			},
			{ // Attribute removal (index update)
				Config: testAccTableReplicaConfig_oneAttribute(rName, "firstKey", "firstKey", "S"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
				),
			},
		},
	})
}

func TestAccDynamoDBTableReplica_lsiUpdate(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_lsi(rName, "lsi-original"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // Change name of local secondary index
				Config: testAccTableReplicaConfig_lsi(rName, "lsi-changed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
				),
			},
		},
	})
}

func TestAccDynamoDBTableReplica_attributeUpdateValidation(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTableReplicaConfig_oneAttribute(rName, "firstKey", "unusedKey", "S"),
				ExpectError: regexp.MustCompile(`attributes must be indexed. Unused attributes: \["unusedKey"\]`),
			},
			{
				Config:      testAccTableReplicaConfig_twoAttributes(rName, "firstKey", "secondKey", "firstUnused", "N", "secondUnused", "S"),
				ExpectError: regexp.MustCompile(`attributes must be indexed. Unused attributes: \["firstUnused"\ \"secondUnused\"]`),
			},
			{
				Config:      testAccTableReplicaConfig_unmatchedIndexes(rName, "firstUnused", "secondUnused"),
				ExpectError: regexp.MustCompile(`indexes must match a defined attribute. Unmatched indexes:`),
			},
		},
	})
}

func TestAccDynamoDBTableReplica_encryption(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var confBYOK, confEncEnabled, confEncDisabled dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	kmsAliasDatasourceName := "data.aws_kms_alias.dynamodb"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_initialStateEncryptionBYOK(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &confBYOK),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "server_side_encryption.0.kms_key_arn", kmsKeyResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableReplicaConfig_initialStateEncryptionAmazonCMK(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &confEncDisabled),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "0"),
					func(s *terraform.State) error {
						if !confEncDisabled.Table.CreationDateTime.Equal(*confBYOK.Table.CreationDateTime) {
							return fmt.Errorf("DynamoDB table recreated when changing SSE")
						}
						return nil
					},
				),
			},
			{
				Config: testAccTableReplicaConfig_initialStateEncryptionAmazonCMK(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &confEncEnabled),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "server_side_encryption.0.kms_key_arn", kmsAliasDatasourceName, "target_key_arn"),
					func(s *terraform.State) error {
						if !confEncEnabled.Table.CreationDateTime.Equal(*confEncDisabled.Table.CreationDateTime) {
							return fmt.Errorf("DynamoDB table recreated when changing SSE")
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccDynamoDBTableReplica_Replica_multiple(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var table dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(t, 3),
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_replica2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "2"),
				),
			},
			{
				Config:            testAccTableReplicaConfig_replica2(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableReplicaConfig_replica0(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "0"),
				),
			},
			{
				Config: testAccTableReplicaConfig_replica2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "2"),
				),
			},
		},
	})
}

func TestAccDynamoDBTableReplica_Replica_single(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(t, 3), // 3 due to shared test configuration
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_replica1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "1"),
				),
			},
			{
				Config:            testAccTableReplicaConfig_replica1(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableReplicaConfig_replica0(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "0"),
				),
			},
			{
				Config: testAccTableReplicaConfig_replica1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "1"),
				),
			},
		},
	})
}

func TestAccDynamoDBTableReplica_Replica_tagMagic(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(t, 3), // 3 due to shared test configuration
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_replicaMoreTags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "1"),
				),
			},
			{
				Config:            testAccTableReplicaConfig_replicaMoreTags(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDynamoDBTableReplica_Replica_singleWithCMK(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	kmsKeyResourceName := "aws_kms_key.test"
	// kmsAliasDatasourceName := "data.aws_kms_alias.master"
	kmsKeyReplicaResourceName := "aws_kms_key.alt_test"
	// kmsAliasReplicaDatasourceName := "data.aws_kms_alias.replica"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(t, 3), // 3 due to shared test configuration
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_replicaCMK(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replica.0.kms_key_arn", kmsKeyReplicaResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "server_side_encryption.0.kms_key_arn", kmsKeyResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccDynamoDBTableReplica_Replica_pitr(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(t, 3), // 3 due to shared test configuration
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_replicaPITR(rName, false, true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"point_in_time_recovery": "true",
						"region_name":            acctest.AlternateRegion(),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"point_in_time_recovery": "false",
						"region_name":            acctest.ThirdRegion(),
					}),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.0.enabled", "false"),
				),
			},
			{
				Config: testAccTableReplicaConfig_replicaPITR(rName, true, false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"point_in_time_recovery": "false",
						"region_name":            acctest.AlternateRegion(),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"point_in_time_recovery": "true",
						"region_name":            acctest.ThirdRegion(),
					}),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.0.enabled", "true"),
				),
			},
		},
	})
}

func TestAccDynamoDBTableReplica_Replica_tagsOneOfTwo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(t, 3),
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_replicaTags(rName, "benny", "smiles", true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					testAccCheckTableReplicaHasTags(resourceName, acctest.AlternateRegion(), true),
					testAccCheckTableReplicaHasTags(resourceName, acctest.ThirdRegion(), false),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"region_name":    acctest.AlternateRegion(),
						"propagate_tags": "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"region_name":    acctest.ThirdRegion(),
						"propagate_tags": "false",
					}),
				),
			},
		},
	})
}

func TestAccDynamoDBTableReplica_Replica_tagsTwoOfTwo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(t, 3),
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_replicaTags(rName, "Structure", "Adobe", true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &conf),
					testAccCheckTableReplicaHasTags(resourceName, acctest.AlternateRegion(), true),
					testAccCheckTableReplicaHasTags(resourceName, acctest.ThirdRegion(), true),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"region_name":    acctest.AlternateRegion(),
						"propagate_tags": "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replica.*", map[string]string{
						"region_name":    acctest.ThirdRegion(),
						"propagate_tags": "true",
					}),
				),
			},
		},
	})
}

func TestAccDynamoDBTableReplica_tableClassInfrequentAccess(t *testing.T) {
	var table dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_class(rName, "STANDARD_INFREQUENT_ACCESS"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "table_class", "STANDARD_INFREQUENT_ACCESS"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableReplicaConfig_class(rName, "STANDARD"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &table),
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

func TestAccDynamoDBTableReplica_backupEncryption(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var confBYOK dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_backupInitialStateEncryption(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &confBYOK),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "server_side_encryption.0.kms_key_arn", kmsKeyResourceName, "arn"),
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

func TestAccDynamoDBTableReplica_backup_overrideEncryption(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var confBYOK dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicaConfig_backupInitialStateOverrideEncryption(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableReplicaExists(resourceName, &confBYOK),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "server_side_encryption.0.kms_key_arn", kmsKeyResourceName, "arn"),
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

func testAccCheckTableReplicaDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dynamodb_table" {
			continue
		}

		log.Printf("[DEBUG] Checking if DynamoDB table %s exists", rs.Primary.ID)
		// Check if queue exists by checking for its attributes
		params := &dynamodb.DescribeTableInput{
			TableName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeTable(params)
		if err == nil {
			return fmt.Errorf("DynamoDB table %s still exists. Failing!", rs.Primary.ID)
		}

		if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceNotFoundException) {
			continue
		}

		return err
	}

	return nil
}

func testAccCheckInitialTableReplicaExists(n string, table *dynamodb.DescribeTableOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		log.Printf("[DEBUG] Trying to create initial table state!")
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DynamoDB table name specified!")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBConn

		params := &dynamodb.DescribeTableInput{
			TableName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeTable(params)

		if err != nil {
			return fmt.Errorf("Problem describing table '%s': %s", rs.Primary.ID, err)
		}

		*table = *resp

		return nil
	}
}

func testAccCheckTableReplicaHasTags(n string, region string, should bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DynamoDB table name specified!")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBConn
		terraformVersion := acctest.Provider.Meta().(*conns.AWSClient).TerraformVersion

		if aws.StringValue(conn.Config.Region) != region {
			session, err := conns.NewSessionForRegion(&conn.Config, region, terraformVersion)
			if err != nil {
				return names.Error(names.DynamoDB, names.ErrActionChecking, "Table", rs.Primary.ID, err)
			}

			conn = dynamodb.New(session)
		}

		newARN, err := tfdynamodb.ARNForNewRegion(rs.Primary.Attributes["arn"], region)

		if err != nil {
			return names.Error(names.DynamoDB, names.ErrActionChecking, "Table", rs.Primary.ID, err)
		}

		tags, err := tfdynamodb.ListTags(conn, newARN)

		if err != nil && !tfawserr.ErrMessageContains(err, "UnknownOperationException", "Tagging is not currently supported in DynamoDB Local.") {
			return names.Error(names.DynamoDB, names.ErrActionChecking, "Table", rs.Primary.Attributes["arn"], err)
		}

		if len(tags.Keys()) > 0 && !should {
			return names.Error(names.DynamoDB, names.ErrActionChecking, "Table", rs.Primary.Attributes["arn"], errors.New("replica should not have tags but does"))
		}

		if len(tags.Keys()) == 0 && should {
			return names.Error(names.DynamoDB, names.ErrActionChecking, "Table", rs.Primary.Attributes["arn"], errors.New("replica should have tags but does not"))
		}

		return nil
	}
}

func testAccCheckInitialTableReplicaConf(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "hash_key", "TestTableHashKey"),
		resource.TestCheckResourceAttr(resourceName, "range_key", "TestTableRangeKey"),
		resource.TestCheckResourceAttr(resourceName, "billing_mode", dynamodb.BillingModeProvisioned),
		resource.TestCheckResourceAttr(resourceName, "write_capacity", "2"),
		resource.TestCheckResourceAttr(resourceName, "read_capacity", "1"),
		resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "0"),
		resource.TestCheckResourceAttr(resourceName, "attribute.#", "4"),
		resource.TestCheckResourceAttr(resourceName, "global_secondary_index.#", "1"),
		resource.TestCheckResourceAttr(resourceName, "local_secondary_index.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
			"name": "TestTableHashKey",
			"type": "S",
		}),
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
			"name": "TestTableRangeKey",
			"type": "S",
		}),
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
			"name": "TestLSIRangeKey",
			"type": "N",
		}),
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
			"name": "TestGSIRangeKey",
			"type": "S",
		}),
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
			"name":            "InitialTestTableGSI",
			"hash_key":        "TestTableHashKey",
			"range_key":       "TestGSIRangeKey",
			"write_capacity":  "1",
			"read_capacity":   "1",
			"projection_type": "KEYS_ONLY",
		}),
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "local_secondary_index.*", map[string]string{
			"name":            "TestTableLSI",
			"range_key":       "TestLSIRangeKey",
			"projection_type": "ALL",
		}),
	)
}

func testAccTableReplicaConfig_basic(rName string) string {
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

func testAccTableReplicaConfig_backup(rName string) string {
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

func testAccTableReplicaConfig_billingPayPerRequest(rName string) string {
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

func testAccTableReplicaConfig_billingPayPerRequestIgnoreChanges(rName string) string {
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

func testAccTableReplicaConfig_billingProvisioned(rName string) string {
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

func testAccTableReplicaConfig_billingProvisionedIgnoreChanges(rName string) string {
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

func testAccTableReplicaConfig_billingPayPerRequestGSI(rName string) string {
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

func testAccTableReplicaConfig_billingProvisionedGSI(rName string) string {
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

func testAccTableReplicaConfig_initialState(rName string) string {
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

func testAccTableReplicaConfig_initialStateEncryptionAmazonCMK(rName string, enabled bool) string {
	return fmt.Sprintf(`
data "aws_kms_alias" "dynamodb" {
  name = "alias/aws/dynamodb"
}

resource "aws_kms_key" "test" {
  description = %[1]q
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

func testAccTableReplicaConfig_initialStateEncryptionBYOK(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q
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

func testAccTableReplicaConfig_addSecondaryGSI(rName string) string {
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

func testAccTableReplicaConfig_streamSpecification(rName string, enabled bool, viewType string) string {
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
  stream_view_type = %[3]q
}
`, rName, enabled, viewType)
}

func testAccTableReplicaConfig_tags(rName string) string {
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

func testAccTableReplicaConfig_gsiUpdate(rName string) string {
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

func testAccTableReplicaConfig_gsiUpdatedCapacity(rName string) string {
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

func testAccTableReplicaConfig_gsiUpdatedOtherAttributes(rName string) string {
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

func testAccTableReplicaConfig_gsiUpdatedNonKeyAttributes(rName string) string {
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

func testAccTableReplicaConfig_gsiMultipleNonKeyAttributes(rName, attributes string) string {
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

func testAccTableReplicaConfig_lsiNonKeyAttributes(rName string) string {
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

func testAccTableReplicaConfig_timeToLive(rName string, ttlEnabled bool) string {
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
    attribute_name = %[2]t ? "TestTTL" : ""
    enabled        = %[2]t
  }
}
`, rName, ttlEnabled)
}

func testAccTableReplicaConfig_oneAttribute(rName, hashKey, attrName, attrType string) string {
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

func testAccTableReplicaConfig_twoAttributes(rName, hashKey, rangeKey, attrName1, attrType1, attrName2, attrType2 string) string {
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

func testAccTableReplicaConfig_unmatchedIndexes(rName, attr1, attr2 string) string {
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

func testAccTableReplicaConfig_replica0(rName string) string {
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

func testAccTableReplicaConfig_replica1(rName string) string {
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

func testAccTableReplicaConfig_replicaMoreTags(rName string) string {
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
    region_name    = data.aws_region.alternate.name
    propagate_tags = true
    
    tags = {
      ReplicaTag = "Bob"
    }
  }

  tags = {
    Name = %[1]q
    Pozo = "Amargo"
  }
}
`, rName))
}

func testAccTableReplicaConfig_replicaCMK(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3), // Prevent "Provider configuration not present" errors
		fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = "awsalternate"
}

resource "aws_kms_key" "test" {
  description = %[1]q
}

resource "aws_kms_key" "alt_test" {
  provider    = "awsalternate"
  description = "%[1]s-2"
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
    kms_key_arn = aws_kms_key.alt_test.arn
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

func testAccTableReplicaConfig_replicaPITR(rName string, mainPITR, replica1, replica2 bool) string {
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

func testAccTableReplicaConfig_replica2(rName string) string {
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

func testAccTableReplicaConfig_replicaTags(rName, key, value string, propagate1, propagate2 bool) string {
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

func testAccTableReplicaConfig_lsi(rName, lsiName string) string {
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

func testAccTableReplicaConfig_class(rName, tableClass string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  hash_key       = "TestTableHashKey"
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 1
  table_class    = %[2]q

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }
}
`, rName, tableClass)
}

func testAccTableReplicaConfig_backupInitialStateOverrideEncryption(rName string) string {
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
  description = %[1]q
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

func testAccTableReplicaConfig_backupInitialStateEncryption(rName string) string {
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
  description = %[1]q
}

resource "aws_dynamodb_table" "test" {
  name                   = "%[1]s-target"
  restore_source_name    = aws_dynamodb_table.source.name
  restore_to_latest_time = true
}
`, rName)
}
