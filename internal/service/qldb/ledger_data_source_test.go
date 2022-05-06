package qldb_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/qldb"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccQLDBLedgerDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qldb_ledger.test"
	datasourceName := "data.aws_qldb_ledger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(qldb.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, qldb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLedgerDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "deletion_protection", resourceName, "deletion_protection"),
					resource.TestCheckResourceAttrPair(datasourceName, "kms_key", resourceName, "kms_key"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "permissions_mode", resourceName, "permissions_mode"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccLedgerDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_qldb_ledger" "test" {
  name                = %[1]q
  permissions_mode    = "STANDARD"
  deletion_protection = false

  tags = {
    Env = "test"
  }
}

data "aws_qldb_ledger" "test" {
  name = aws_qldb_ledger.test.id
}
`, rName)
}
