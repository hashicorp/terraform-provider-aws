package keyspaces_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/keyspaces"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkeyspaces "github.com/hashicorp/terraform-provider-aws/internal/service/keyspaces"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccKeyspacesTable_basic(t *testing.T) {
	var v keyspaces.GetTableOutput
	rName1 := "tf_acc_test_" + sdkacctest.RandString(20)
	rName2 := "tf_acc_test_" + sdkacctest.RandString(20)
	resourceName := "aws_keyspaces_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, keyspaces.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "cassandra", fmt.Sprintf("/keyspace/%s/table/%s", rName1, rName2)),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.0.throughput_mode", "PAY_PER_REQUEST"),
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
						"name": "message",
						"type": "ascii",
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.partition_key.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.partition_key.*", map[string]string{
						"name": "message",
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.static_column.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "table_name", rName2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	var v keyspaces.GetTableOutput
	rName1 := "tf_acc_test_" + sdkacctest.RandString(20)
	rName2 := "tf_acc_test_" + sdkacctest.RandString(20)
	resourceName := "aws_keyspaces_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, keyspaces.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfkeyspaces.ResourceTable(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKeyspacesTable_tags(t *testing.T) {
	var v keyspaces.GetTableOutput
	rName1 := "tf_acc_test_" + sdkacctest.RandString(20)
	rName2 := "tf_acc_test_" + sdkacctest.RandString(20)
	resourceName := "aws_keyspaces_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, keyspaces.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfigTags1(rName1, rName2, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableConfigTags2(rName1, rName2, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccTableConfigTags1(rName1, rName2, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccKeyspacesTable_multipleColumns(t *testing.T) {
	var v keyspaces.GetTableOutput
	rName1 := "tf_acc_test_" + sdkacctest.RandString(20)
	rName2 := "tf_acc_test_" + sdkacctest.RandString(20)
	resourceName := "aws_keyspaces_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, keyspaces.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableMultipleColumnsConfig(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.clustering_key.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.clustering_key.*", map[string]string{
						"name":     "division",
						"order_by": "ASC",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.clustering_key.*", map[string]string{
						"name":     "region",
						"order_by": "DESC",
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.column.#", "9"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						"name": "id",
						"type": "text",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						"name": "name",
						"type": "text",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						"name": "region",
						"type": "text",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						"name": "division",
						"type": "text",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						"name": "project",
						"type": "text",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						"name": "role",
						"type": "text",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						"name": "pay_scale",
						"type": "int",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						"name": "vacation_hrs",
						"type": "float",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						"name": "manager_id",
						"type": "text",
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.partition_key.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.partition_key.*", map[string]string{
						"name": "id",
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.static_column.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.static_column.*", map[string]string{
						"name": "role",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.static_column.*", map[string]string{
						"name": "pay_scale",
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
	var v1, v2 keyspaces.GetTableOutput
	rName1 := "tf_acc_test_" + sdkacctest.RandString(20)
	rName2 := "tf_acc_test_" + sdkacctest.RandString(20)
	resourceName := "aws_keyspaces_table.test"
	kmsKeyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, keyspaces.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableAllAttributesConfig(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(resourceName, &v1),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "cassandra", fmt.Sprintf("/keyspace/%s/table/%s", rName1, rName2)),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.0.read_capacity_units", "200"),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.0.throughput_mode", "PROVISIONED"),
					resource.TestCheckResourceAttr(resourceName, "capacity_specification.0.write_capacity_units", "100"),
					resource.TestCheckResourceAttr(resourceName, "comment.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "comment.0.message", "TESTING"),
					resource.TestCheckResourceAttr(resourceName, "default_time_to_live", "500000"),
					resource.TestCheckResourceAttr(resourceName, "encryption_specification.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_specification.0.kms_key_identifier", kmsKeyResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "encryption_specification.0.type", "CUSTOMER_MANAGED_KMS_KEY"),
					resource.TestCheckResourceAttr(resourceName, "keyspace_name", rName1),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.0.status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "table_name", rName2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
				Config: testAccTableAllAttributesUpdatedConfig(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(resourceName, &v2),
					testAccCheckTableNotRecreated(&v1, &v2),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "cassandra", fmt.Sprintf("/keyspace/%s/table/%s", rName1, rName2)),
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
					resource.TestCheckResourceAttr(resourceName, "table_name", rName2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "ttl.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ttl.0.status", "ENABLED"),
				),
			},
		},
	})
}

func TestAccKeyspacesTable_addColumns(t *testing.T) {
	var v1, v2 keyspaces.GetTableOutput
	rName1 := "tf_acc_test_" + sdkacctest.RandString(20)
	rName2 := "tf_acc_test_" + sdkacctest.RandString(20)
	resourceName := "aws_keyspaces_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, keyspaces.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.clustering_key.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.column.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						"name": "message",
						"type": "ascii",
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.partition_key.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.partition_key.*", map[string]string{
						"name": "message",
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
				Config: testAccTableNewColumnsConfig(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(resourceName, &v2),
					testAccCheckTableNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.clustering_key.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.column.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						"name": "message",
						"type": "ascii",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						"name": "ts",
						"type": "timestamp",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						"name": "amount",
						"type": "decimal",
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.partition_key.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.partition_key.*", map[string]string{
						"name": "message",
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.static_column.#", "0"),
				),
			},
		},
	})
}

func TestAccKeyspacesTable_delColumns(t *testing.T) {
	var v1, v2 keyspaces.GetTableOutput
	rName1 := "tf_acc_test_" + sdkacctest.RandString(20)
	rName2 := "tf_acc_test_" + sdkacctest.RandString(20)
	resourceName := "aws_keyspaces_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, keyspaces.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableNewColumnsConfig(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.clustering_key.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.column.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						"name": "message",
						"type": "ascii",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						"name": "ts",
						"type": "timestamp",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						"name": "amount",
						"type": "decimal",
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.partition_key.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.partition_key.*", map[string]string{
						"name": "message",
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
				Config: testAccTableConfig(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(resourceName, &v2),
					testAccCheckTableRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.clustering_key.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.column.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.column.*", map[string]string{
						"name": "message",
						"type": "ascii",
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.partition_key.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.partition_key.*", map[string]string{
						"name": "message",
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.static_column.#", "0"),
				),
			},
		},
	})
}

func testAccCheckTableDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KeyspacesConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_keyspaces_table" {
			continue
		}

		keyspaceName, tableName, err := tfkeyspaces.TableParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfkeyspaces.FindTableByTwoPartKey(context.Background(), conn, keyspaceName, tableName)

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

func testAccCheckTableExists(n string, v *keyspaces.GetTableOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Keyspaces Table ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KeyspacesConn

		keyspaceName, tableName, err := tfkeyspaces.TableParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		output, err := tfkeyspaces.FindTableByTwoPartKey(context.Background(), conn, keyspaceName, tableName)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTableNotRecreated(i, j *keyspaces.GetTableOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.CreationTimestamp).Equal(aws.TimeValue(j.CreationTimestamp)) {
			return errors.New("Keyspaces Table was recreated")
		}

		return nil
	}
}

func testAccCheckTableRecreated(i, j *keyspaces.GetTableOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreationTimestamp).Equal(aws.TimeValue(j.CreationTimestamp)) {
			return errors.New("Keyspaces Table was not recreated")
		}

		return nil
	}
}

func testAccTableConfig(rName1, rName2 string) string {
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

func testAccTableConfigTags1(rName1, rName2, tagKey1, tagValue1 string) string {
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

func testAccTableConfigTags2(rName1, rName2, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

func testAccTableMultipleColumnsConfig(rName1, rName2 string) string {
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
      name = "name"
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
      name = "pay_scale"
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
      name = "pay_scale"
    }
  }
}
`, rName1, rName2)
}

func testAccTableAllAttributesConfig(rName1, rName2 string) string {
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

func testAccTableAllAttributesUpdatedConfig(rName1, rName2 string) string {
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

func testAccTableNewColumnsConfig(rName1, rName2 string) string {
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
