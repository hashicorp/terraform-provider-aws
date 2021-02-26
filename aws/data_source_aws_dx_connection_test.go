package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsDxConnection_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_dx_connection.test"
	datasourceName := "data.aws_dx_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsDxConnectionConfig_NonExistent,
				ExpectError: regexp.MustCompile(`Direct Connect Connection not found`),
			},
			{
				Config: testAccDataSourceAwsDxConnectionConfig_Name(rName, "1Gbps", "EqDC2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "owner_account_id", resourceName, "owner_account_id"),
				),
			},
		},
	})
}

func testAccDataSourceAwsDxConnectionConfig_Name(rName, rBandwidth, rLocation string) string {
	return fmt.Sprintf(`
resource "aws_dx_connection" "wrong" {
  name            = "%s-wrong"
	bandwidth				= "%s"
	location				= "%s"
}

resource "aws_dx_connection" "test" {
  name            = "%s"
	bandwidth				= "%s"
	location				= "%s"
}

data "aws_dx_connection" "test" {
  name = aws_dx_connection.test.name
}
`, rName, rBandwidth, rLocation, rName, rBandwidth, rLocation)
}

const testAccDataSourceAwsDxConnectionConfig_NonExistent = `
data "aws_dx_connection" "test" {
  name = "tf-acc-test-does-not-exist"
}
`
