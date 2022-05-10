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
	resourceName := "aws_keyspaces_keyspace.test"

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
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "cassandra", "xyzzy"),
					resource.TestCheckResourceAttr(resourceName, "keyspace_name", rName1),
					resource.TestCheckResourceAttr(resourceName, "table_name", rName1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	resourceName := "aws_keyspaces_keyspace.test"

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
}
`, rName1, rName2)
}
