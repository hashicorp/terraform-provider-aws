package directconnect_test

import (
	"fmt"
	"regexp"
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

func TestAccDirectConnectConnection_basic(t *testing.T) {
	var connection directconnect.Connection
	resourceName := "aws_dx_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxConnectionConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName, &connection),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "directconnect", regexp.MustCompile(`dxcon/.+`)),
					resource.TestCheckResourceAttr(resourceName, "bandwidth", "1Gbps"),
					resource.TestCheckResourceAttrSet(resourceName, "location"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "provider_name", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			// Test import.
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDirectConnectConnection_disappears(t *testing.T) {
	var connection directconnect.Connection
	resourceName := "aws_dx_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxConnectionConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName, &connection),
					acctest.CheckResourceDisappears(acctest.Provider, tfdirectconnect.ResourceConnection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDirectConnectConnection_providerName(t *testing.T) {
	var connection directconnect.Connection
	resourceName := "aws_dx_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxConnectionConfigProviderName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName, &connection),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "directconnect", regexp.MustCompile(`dxcon/.+`)),
					resource.TestCheckResourceAttr(resourceName, "bandwidth", "1Gbps"),
					resource.TestCheckResourceAttrSet(resourceName, "location"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "provider_name"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			// Test import.
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDirectConnectConnection_tags(t *testing.T) {
	var connection directconnect.Connection
	resourceName := "aws_dx_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxConnectionConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			// Test import.
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDxConnectionConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDxConnectionConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckConnectionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DirectConnectConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dx_connection" {
			continue
		}

		_, err := tfdirectconnect.FindConnectionByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Direct Connect Connection %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckConnectionExists(name string, v *directconnect.Connection) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DirectConnectConn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Direct Connect Connection ID is set")
		}

		connection, err := tfdirectconnect.FindConnectionByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *connection

		return nil
	}
}

func testAccDxConnectionConfigBasic(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

locals {
  location_codes = tolist(data.aws_dx_locations.test.location_codes)
  idx            = min(2, length(local.location_codes) - 1)
}

resource "aws_dx_connection" "test" {
  name      = %[1]q
  bandwidth = "1Gbps"
  location  = local.location_codes[local.idx]
}
`, rName)
}

func testAccDxConnectionConfigProviderName(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

locals {
  location_codes = tolist(data.aws_dx_locations.test.location_codes)
  idx            = min(2, length(local.location_codes) - 1)
}

data "aws_dx_location" "test" {
  location_code = local.location_codes[local.idx]
}

resource "aws_dx_connection" "test" {
  name      = %[1]q
  bandwidth = "1Gbps"
  location  = data.aws_dx_location.test.location_code

  provider_name = data.aws_dx_location.test.available_providers[0]
}
`, rName)
}

func testAccDxConnectionConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

locals {
  location_codes = tolist(data.aws_dx_locations.test.location_codes)
  idx            = min(2, length(local.location_codes) - 1)
}

resource "aws_dx_connection" "test" {
  name      = %[1]q
  bandwidth = "1Gbps"
  location  = local.location_codes[local.idx]

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDxConnectionConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

locals {
  location_codes = tolist(data.aws_dx_locations.test.location_codes)
  idx            = min(2, length(local.location_codes) - 1)
}

resource "aws_dx_connection" "test" {
  name      = %[1]q
  bandwidth = "1Gbps"
  location  = local.location_codes[local.idx]

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
