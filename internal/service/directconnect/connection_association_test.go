package directconnect_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/directconnect"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccDirectConnectConnectionAssociation_basic(t *testing.T) {
	resourceName := "aws_dx_connection_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConnectionAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionAssociationExists(resourceName),
				),
			},
		},
	})
}

func TestAccDirectConnectConnectionAssociation_lagOnConnection(t *testing.T) {
	resourceName := "aws_dx_connection_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConnectionAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionAssociationConfig_lagOnConnection(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionAssociationExists(resourceName),
				),
			},
		},
	})
}

func TestAccDirectConnectConnectionAssociation_multiple(t *testing.T) {
	resourceName1 := "aws_dx_connection_association.test1"
	resourceName2 := "aws_dx_connection_association.test2"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConnectionAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionAssociationConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionAssociationExists(resourceName1),
					testAccCheckConnectionAssociationExists(resourceName2),
				),
			},
		},
	})
}

func testAccCheckConnectionAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DirectConnectConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dx_connection_association" {
			continue
		}

		err := tfdirectconnect.FindConnectionAssociationExists(conn, rs.Primary.ID, rs.Primary.Attributes["lag_id"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Direct Connect Connection (%s) LAG (%s) Association still exists", rs.Primary.ID, rs.Primary.Attributes["lag_id"])
	}

	return nil
}

func testAccCheckConnectionAssociationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DirectConnectConn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		err := tfdirectconnect.FindConnectionAssociationExists(conn, rs.Primary.ID, rs.Primary.Attributes["lag_id"])

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccConnectionAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

locals {
  location_code = tolist(data.aws_dx_locations.test.location_codes)[1]
}

resource "aws_dx_connection" "test" {
  name      = %[1]q
  bandwidth = "1Gbps"
  location  = local.location_code
}

resource "aws_dx_lag" "test" {
  name                  = %[1]q
  connections_bandwidth = "1Gbps"
  location              = local.location_code
  force_destroy         = true
}

resource "aws_dx_connection_association" "test" {
  connection_id = aws_dx_connection.test.id
  lag_id        = aws_dx_lag.test.id
}
`, rName)
}

func testAccConnectionAssociationConfig_lagOnConnection(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

locals {
  location_code = tolist(data.aws_dx_locations.test.location_codes)[1]
}

resource "aws_dx_connection" "test1" {
  name      = "%[1]s-1"
  bandwidth = "1Gbps"
  location  = local.location_code
}

resource "aws_dx_connection" "test2" {
  name      = "%[1]s-2"
  bandwidth = "1Gbps"
  location  = local.location_code
}

resource "aws_dx_lag" "test" {
  name                  = %[1]q
  connection_id         = aws_dx_connection.test1.id
  connections_bandwidth = "1Gbps"
  location              = local.location_code
}

resource "aws_dx_connection_association" "test" {
  connection_id = aws_dx_connection.test2.id
  lag_id        = aws_dx_lag.test.id
}
`, rName)
}

func testAccConnectionAssociationConfig_multiple(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

locals {
  location_code = tolist(data.aws_dx_locations.test.location_codes)[1]
}

resource "aws_dx_connection" "test1" {
  name      = "%[1]s-1"
  bandwidth = "1Gbps"
  location  = local.location_code
}

resource "aws_dx_connection" "test2" {
  name      = "%[1]s-2"
  bandwidth = "1Gbps"
  location  = local.location_code
}

resource "aws_dx_lag" "test" {
  name                  = %[1]q
  connections_bandwidth = "1Gbps"
  location              = local.location_code
  force_destroy         = true
}

resource "aws_dx_connection_association" "test1" {
  connection_id = aws_dx_connection.test1.id
  lag_id        = aws_dx_lag.test.id
}

resource "aws_dx_connection_association" "test2" {
  connection_id = aws_dx_connection.test2.id
  lag_id        = aws_dx_lag.test.id
}
`, rName)
}
