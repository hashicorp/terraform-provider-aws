package aws

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsDxConnection_Basic(t *testing.T) {
	key := "DX_LOCATION"
	location := os.Getenv(key)
	if location == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

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
				Config: testAccDataSourceAwsDxConnectionConfig_Name(rName, location),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "location", resourceName, "location"),
					resource.TestCheckResourceAttrPair(datasourceName, "jumbo_frame_capable", resourceName, "jumbo_frame_capable"),
					resource.TestCheckResourceAttrPair(datasourceName, "bandwidth", resourceName, "bandwidth"),
					resource.TestCheckResourceAttr(datasourceName, "state", "requested"),
				),
			},
		},
	})
}

func testAccDataSourceAwsDxConnectionConfig_Name(rName, location string) string {
	return fmt.Sprintf(`
resource "aws_dx_connection" "wrong" {
	name            = "%s-wrong"
	bandwidth       = "1Gbps"
	location        = "%s"
}
resource "aws_dx_connection" "test" {
	name            = "%s"
	bandwidth       = "1Gbps"
	location        = "%s"
}

data "aws_dx_connection" "test" {
  name = "${aws_dx_connection.test.name}"
}
`, rName, location, rName, location)
}

const testAccDataSourceAwsDxConnectionConfig_NonExistent = `
data "aws_dx_connection" "test" {
  name = "tf-acc-test-does-not-exist"
}
`
