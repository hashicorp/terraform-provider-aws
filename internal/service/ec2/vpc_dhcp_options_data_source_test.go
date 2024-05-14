// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCDHCPOptionsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_dhcp_options.test"
	datasourceName := "data.aws_vpc_dhcp_options.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccVPCDHCPOptionsDataSourceConfig_missing,
				ExpectError: regexache.MustCompile(`no matching EC2 DHCP Options Set found`),
			},
			{
				Config: testAccVPCDHCPOptionsDataSourceConfig_id,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "dhcp_options_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDomainName, resourceName, names.AttrDomainName),
					resource.TestCheckResourceAttrPair(datasourceName, "domain_name_servers.#", resourceName, "domain_name_servers.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "domain_name_servers.0", resourceName, "domain_name_servers.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "domain_name_servers.1", resourceName, "domain_name_servers.1"),
					resource.TestCheckResourceAttrPair(datasourceName, "ipv6_address_preferred_lease_time", resourceName, "ipv6_address_preferred_lease_time"),
					resource.TestCheckResourceAttrPair(datasourceName, "netbios_name_servers.#", resourceName, "netbios_name_servers.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "netbios_name_servers.0", resourceName, "netbios_name_servers.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "netbios_node_type", resourceName, "netbios_node_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "ntp_servers.#", resourceName, "ntp_servers.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "ntp_servers.0", resourceName, "ntp_servers.0"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrOwnerID, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccVPCDHCPOptionsDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	rInt := sdkacctest.RandInt()
	resourceName := "aws_vpc_dhcp_options.test.0"
	datasourceName := "data.aws_vpc_dhcp_options.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDHCPOptionsDataSourceConfig_filter(rInt, 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "dhcp_options_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDomainName, resourceName, names.AttrDomainName),
					resource.TestCheckResourceAttrPair(datasourceName, "domain_name_servers.#", resourceName, "domain_name_servers.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "domain_name_servers.0", resourceName, "domain_name_servers.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "domain_name_servers.1", resourceName, "domain_name_servers.1"),
					resource.TestCheckResourceAttrPair(datasourceName, "ipv6_address_preferred_lease_time", resourceName, "ipv6_address_preferred_lease_time"),
					resource.TestCheckResourceAttrPair(datasourceName, "netbios_name_servers.#", resourceName, "netbios_name_servers.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "netbios_name_servers.0", resourceName, "netbios_name_servers.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "netbios_node_type", resourceName, "netbios_node_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "ntp_servers.#", resourceName, "ntp_servers.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "ntp_servers.0", resourceName, "ntp_servers.0"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrOwnerID, resourceName, names.AttrOwnerID),
				),
			},
			{
				Config:      testAccVPCDHCPOptionsDataSourceConfig_filter(rInt, 2),
				ExpectError: regexache.MustCompile(`multiple EC2 DHCP Options Sets matched`),
			},
			{
				// We have one last empty step here because otherwise we'll leave the
				// test case with resources in the state and an erroneous config, and
				// thus the automatic destroy step will fail. This ensures we end with
				// both an empty state and a valid config.
				Config: testAccVPCDHCPOptionsDataSourceConfig_blank(),
			},
		},
	})
}

const testAccVPCDHCPOptionsDataSourceConfig_missing = `
data "aws_vpc_dhcp_options" "test" {
  dhcp_options_id = "does-not-exist"
}
`

const testAccVPCDHCPOptionsDataSourceConfig_id = `
resource "aws_vpc_dhcp_options" "incorrect" {
  domain_name = "tf-acc-test-incorrect.example.com"
}

resource "aws_vpc_dhcp_options" "test" {
  domain_name                       = "service.consul"
  domain_name_servers               = ["127.0.0.1", "10.0.0.2"]
  ipv6_address_preferred_lease_time = 3600
  netbios_name_servers              = ["127.0.0.1"]
  netbios_node_type                 = 2
  ntp_servers                       = ["127.0.0.1"]

  tags = {
    Name = "tf-acc-test"
  }
}

data "aws_vpc_dhcp_options" "test" {
  dhcp_options_id = aws_vpc_dhcp_options.test.id
}
`

func testAccVPCDHCPOptionsDataSourceConfig_filter(rInt, count int) string {
	return fmt.Sprintf(`
resource "aws_vpc_dhcp_options" "incorrect" {
  domain_name = "tf-acc-test-incorrect.example.com"
}

resource "aws_vpc_dhcp_options" "test" {
  count = %[2]d

  domain_name                       = "tf-acc-test-%[1]d.example.com"
  domain_name_servers               = ["127.0.0.1", "10.0.0.2"]
  ipv6_address_preferred_lease_time = 3600
  netbios_name_servers              = ["127.0.0.1"]
  netbios_node_type                 = 2
  ntp_servers                       = ["127.0.0.1"]

  tags = {
    Name = "tf-acc-test-%[1]d"
  }
}

data "aws_vpc_dhcp_options" "test" {
  filter {
    name   = "key"
    values = ["domain-name"]
  }

  filter {
    name   = "value"
    values = [aws_vpc_dhcp_options.test[0].domain_name]
  }
}
`, rInt, count)
}

func testAccVPCDHCPOptionsDataSourceConfig_blank() string {
	return `/* this config intentionally left blank */`
}
