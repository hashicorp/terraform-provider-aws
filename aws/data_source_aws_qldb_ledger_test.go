package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDataSourceAwsQLDBLedger(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsQLDBLedgerConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsQLDBLedgerCheck("data.aws_qldb_ledger.by_name"),
				),
			},
		},
	})
}

func testAccDataSourceAwsQLDBLedgerCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		ledgerRs, ok := s.RootModule().Resources["aws_qldb_ledger.tf_test"]
		if !ok {
			return fmt.Errorf("can't find aws_qldb_ledger.tf_test in state")
		}

		attr := rs.Primary.Attributes

		if attr["name"] != ledgerRs.Primary.Attributes["name"] {
			return fmt.Errorf(
				"name is %s; want %s",
				attr["name"],
				ledgerRs.Primary.Attributes["name"],
			)
		}

		return nil
	}
}

const testAccDataSourceAwsQLDBLedgerConfig = `
resource "aws_qldb_ledger" "tf_wrong1" {
  name = "wrong1"
  deletion_protection = false
}
resource "aws_qldb_ledger" "tf_test" {
  name = "tf-test"
  deletion_protection = false
}
resource "aws_qldb_ledger" "tf_wrong2" {
  name = "wrong2"
  deletion_protection = false
}

data "aws_qldb_ledger" "by_name" {
  name = "${aws_qldb_ledger.tf_test.name}"
}
`
