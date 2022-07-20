package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccVPCDHCPOptions_basic(t *testing.T) {
	var d ec2.DhcpOptions
	resourceName := "aws_vpc_dhcp_options.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDHCPOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDHCPOptionsConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists(resourceName, &d),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`dhcp-options/dopt-.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain_name", ""),
					resource.TestCheckResourceAttr(resourceName, "domain_name_servers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "netbios_name_servers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "netbios_node_type", "1"),
					resource.TestCheckResourceAttr(resourceName, "ntp_servers.#", "0"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	var d ec2.DhcpOptions
	resourceName := "aws_vpc_dhcp_options.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDHCPOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDHCPOptionsConfig_full(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDHCPOptionsExists(resourceName, &d),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`dhcp-options/dopt-.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "domain_name_servers.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_servers.0", "127.0.0.1"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_servers.1", "10.0.0.2"),
					resource.TestCheckResourceAttr(resourceName, "netbios_name_servers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "netbios_name_servers.0", "127.0.0.1"),
					resource.TestCheckResourceAttr(resourceName, "netbios_node_type", "2"),
					resource.TestCheckResourceAttr(resourceName, "ntp_servers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ntp_servers.0", "127.0.0.1"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
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
	var d ec2.DhcpOptions
	resourceName := "aws_vpc_dhcp_options.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDHCPOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDHCPOptionsConfig_tags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists(resourceName, &d),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCDHCPOptionsConfig_tags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists(resourceName, &d),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccVPCDHCPOptionsConfig_tags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists(resourceName, &d),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccVPCDHCPOptions_disappears(t *testing.T) {
	var d ec2.DhcpOptions
	resourceName := "aws_vpc_dhcp_options.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDHCPOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDHCPOptionsConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists(resourceName, &d),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceVPCDHCPOptions(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDHCPOptionsDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_dhcp_options" {
			continue
		}

		_, err := tfec2.FindDHCPOptionsByID(conn, rs.Primary.ID)

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

func testAccCheckDHCPOptionsExists(n string, v *ec2.DhcpOptions) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 DHCP Options Set ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindDHCPOptionsByID(conn, rs.Primary.ID)

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
  domain_name          = %[2]q
  domain_name_servers  = ["127.0.0.1", "10.0.0.2"]
  ntp_servers          = ["127.0.0.1"]
  netbios_name_servers = ["127.0.0.1"]
  netbios_node_type    = "2"

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
