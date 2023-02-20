package fms_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/fms"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tffms "github.com/hashicorp/terraform-provider-aws/internal/service/fms"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccFMSProtocolList_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fms_protocol.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckAdmin(t)
			acctest.PreCheckOrganizationsEnabled(t)
			acctest.PreCheckOrganizationManagementAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, fms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProtocolListDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProtocolListConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtocolListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "protocols.0", "IPv4"),
					resource.TestCheckResourceAttr(resourceName, "protocols.1", "IPv6"),
					resource.TestCheckResourceAttr(resourceName, "protocols.2", "ICMP"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"protocol_update_token"},
			},
		},
	})
}

func TestAccFMSProtocolList_update(t *testing.T) {
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fms_protocol.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckAdmin(t)
			acctest.PreCheckOrganizationsEnabled(t)
			acctest.PreCheckOrganizationManagementAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, fms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProtocolListDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProtocolListConfig_basic(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtocolListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "protocols.0", "IPv4"),
					resource.TestCheckResourceAttr(resourceName, "protocols.1", "IPv6"),
					resource.TestCheckResourceAttr(resourceName, "protocols.2", "ICMP"),
				),
			},
			{
				Config: testAccProtocolListConfig_basic(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtocolListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "protocols.0", "IPv4"),
					resource.TestCheckResourceAttr(resourceName, "protocols.1", "IPv6"),
					resource.TestCheckResourceAttr(resourceName, "protocols.2", "ICMP"),
				),
			},
			{
				Config: testAccProtocolListConfig_basic2(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtocolListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "protocols.0", "IPv4"),
					resource.TestCheckResourceAttr(resourceName, "protocols.1", "UDP"),
				),
			},
		},
	})
}

func TestAccFMSProtocolList_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fms_protocol.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckAdmin(t)
			acctest.PreCheckOrganizationsEnabled(t)
			acctest.PreCheckOrganizationManagementAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, fms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProtocolListDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProtocolListConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccProtocolListConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtocolListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckProtocolListExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No FMS Protocol List ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FMSConn

		_, err := tffms.FindProtocolListByID(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckProtocolListDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).FMSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_fms_protocol" {
			continue
		}

		_, err := tffms.FindProtocolListByID(context.Background(), conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("FMS Protocol List %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccProtocolListConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_fms_protocol" "test" {
  name      = %q
  protocols = ["IPv4", "IPv6", "ICMP"]
}
`, rName)
}

func testAccProtocolListConfig_basic2(rName string) string {
	return fmt.Sprintf(`
resource "aws_fms_protocol" "test" {
  name      = %q
  protocols = ["IPv4", "UDP"]
}
`, rName)
}

func testAccProtocolListConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_fms_protocol" "test" {
  name      = %q
  protocols = ["IPv4", "IPv6", "ICMP"]
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccProtocolListConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_fms_protocol" "test" {
  name      = %q
  protocols = ["IPv4", "IPv6", "ICMP"]
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
