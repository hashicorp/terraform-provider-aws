package keyspaces_test

import (
	"context"
	"fmt"
	"testing"

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
					testAccCheckTableExists(resourceName),
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
						"name": "message", // Keyspaces always changes the value to lowercase.
						"type": "ascii",   // Keyspaces always changes the value to lowercase.
					}),
					resource.TestCheckResourceAttr(resourceName, "schema_definition.0.partition_key.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "schema_definition.0.partition_key.*", map[string]string{
						"name": "message", // Keyspaces always changes the value to lowercase.
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
					testAccCheckTableExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfkeyspaces.ResourceTable(), resourceName),
				),
				ExpectNonEmptyPlan: true,
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

func testAccCheckTableExists(n string) resource.TestCheckFunc {
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

		_, err = tfkeyspaces.FindTableByTwoPartKey(context.Background(), conn, keyspaceName, tableName)

		if err != nil {
			return err
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
      name = "Message"
      type = "ASCII"
    }

    partition_key {
      name = "Message"
    }
  }
}
`, rName1, rName2)
}
