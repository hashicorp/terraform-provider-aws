// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCDHCPOptions_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var d awstypes.DhcpOptions
	resourceName := "aws_vpc_dhcp_options.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDHCPOptionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDHCPOptionsConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists(ctx, resourceName, &d),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`dhcp-options/dopt-.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, ""),
					resource.TestCheckResourceAttr(resourceName, "domain_name_servers.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_preferred_lease_time", ""),
					resource.TestCheckResourceAttr(resourceName, "netbios_name_servers.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "netbios_node_type", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ntp_servers.#", acctest.Ct0),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccVPCDHCPOptions_full(t *testing.T) {
	ctx := acctest.Context(t)
	var d awstypes.DhcpOptions
	resourceName := "aws_vpc_dhcp_options.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDHCPOptionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDHCPOptionsConfig_full(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDHCPOptionsExists(ctx, resourceName, &d),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`dhcp-options/dopt-.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, "domain_name_servers.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "domain_name_servers.0", "127.0.0.1"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_servers.1", "10.0.0.2"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_preferred_lease_time", "1440"),
					resource.TestCheckResourceAttr(resourceName, "netbios_name_servers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "netbios_name_servers.0", "127.0.0.1"),
					resource.TestCheckResourceAttr(resourceName, "netbios_node_type", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "ntp_servers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ntp_servers.0", "127.0.0.1"),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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

func TestAccVPCDHCPOptions_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var d awstypes.DhcpOptions
	resourceName := "aws_vpc_dhcp_options.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDHCPOptionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDHCPOptionsConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists(ctx, resourceName, &d),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCDHCPOptionsConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists(ctx, resourceName, &d),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccVPCDHCPOptionsConfig_tags1(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists(ctx, resourceName, &d),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccVPCDHCPOptions_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var d awstypes.DhcpOptions
	resourceName := "aws_vpc_dhcp_options.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDHCPOptionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDHCPOptionsConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists(ctx, resourceName, &d),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPCDHCPOptions(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDHCPOptionsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_dhcp_options" {
				continue
			}

			_, err := tfec2.FindDHCPOptionsByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 DHCP Options Set %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDHCPOptionsExists(ctx context.Context, n string, v *awstypes.DhcpOptions) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 DHCP Options Set ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindDHCPOptionsByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

const testAccVPCDHCPOptionsConfig_basic = `
resource "aws_vpc_dhcp_options" "test" {
  netbios_node_type = 1
}
`

func testAccVPCDHCPOptionsConfig_full(rName, domainName string) string {
	return fmt.Sprintf(`
resource "aws_vpc_dhcp_options" "test" {
  domain_name                       = %[2]q
  domain_name_servers               = ["127.0.0.1", "10.0.0.2"]
  ipv6_address_preferred_lease_time = 1440
  ntp_servers                       = ["127.0.0.1"]
  netbios_name_servers              = ["127.0.0.1"]
  netbios_node_type                 = "2"

  tags = {
    Name = %[1]q
  }
}
`, rName, domainName)
}

func testAccVPCDHCPOptionsConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc_dhcp_options" "test" {
  netbios_node_type = 2

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccVPCDHCPOptionsConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc_dhcp_options" "test" {
  netbios_node_type = 2

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}
