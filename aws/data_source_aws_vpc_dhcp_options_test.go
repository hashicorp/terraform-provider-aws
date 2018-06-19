package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsVpcDhcpOptions_basic(t *testing.T) {
	resourceName := "aws_vpc_dhcp_options.test"
	datasourceName := "data.aws_vpc_dhcp_options.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcDhcpOptionsConfig_DhcpOptionsID,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "dhcp_options_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "domain_name", resourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "domain_name_servers.#", resourceName, "domain_name_servers.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "domain_name_servers.0", resourceName, "domain_name_servers.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "domain_name_servers.1", resourceName, "domain_name_servers.1"),
					resource.TestCheckResourceAttrPair(datasourceName, "netbios_name_servers.#", resourceName, "netbios_name_servers.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "netbios_name_servers.0", resourceName, "netbios_name_servers.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "netbios_node_type", resourceName, "netbios_node_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "ntp_servers.#", resourceName, "ntp_servers.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "ntp_servers.0", resourceName, "ntp_servers.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Name", resourceName, "tags.Name"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcDhcpOptions_Filter(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_vpc_dhcp_options.test"
	datasourceName := "data.aws_vpc_dhcp_options.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcDhcpOptionsConfig_Filter(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "dhcp_options_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "domain_name", resourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "domain_name_servers.#", resourceName, "domain_name_servers.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "domain_name_servers.0", resourceName, "domain_name_servers.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "domain_name_servers.1", resourceName, "domain_name_servers.1"),
					resource.TestCheckResourceAttrPair(datasourceName, "netbios_name_servers.#", resourceName, "netbios_name_servers.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "netbios_name_servers.0", resourceName, "netbios_name_servers.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "netbios_node_type", resourceName, "netbios_node_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "ntp_servers.#", resourceName, "ntp_servers.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "ntp_servers.0", resourceName, "ntp_servers.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Name", resourceName, "tags.Name"),
				),
			},
		},
	})
}

const testAccDataSourceAwsVpcDhcpOptionsConfig_DhcpOptionsID = `
resource "aws_vpc_dhcp_options" "test" {
  domain_name          = "service.consul"
  domain_name_servers  = ["127.0.0.1", "10.0.0.2"]
  netbios_name_servers = ["127.0.0.1"]
  netbios_node_type    = 2
  ntp_servers          = ["127.0.0.1"]

  tags {
    Name = "tf-acc-test"
  }
}

data "aws_vpc_dhcp_options" "test" {
  dhcp_options_id = "${aws_vpc_dhcp_options.test.id}"
}
`

func testAccDataSourceAwsVpcDhcpOptionsConfig_Filter(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc_dhcp_options" "test" {
  domain_name          = "tf-acc-test-%d.example.com"
  domain_name_servers  = ["127.0.0.1", "10.0.0.2"]
  netbios_name_servers = ["127.0.0.1"]
  netbios_node_type    = 2
  ntp_servers          = ["127.0.0.1"]

  tags {
    Name = "tf-acc-test-%d"
  }
}

data "aws_vpc_dhcp_options" "test" {
  filter {
    name   = "key"
    values = ["domain-name"]
  }

  filter {
    name   = "value"
    values = ["${aws_vpc_dhcp_options.test.domain_name}"]
  }
}
`, rInt, rInt)
}
