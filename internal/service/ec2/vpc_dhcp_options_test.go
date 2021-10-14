package ec2_test

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)





func TestAccEC2VPCDHCPOptions_basic(t *testing.T) {
	var d ec2.DhcpOptions
	resourceName := "aws_vpc_dhcp_options.test"
	rName := sdkacctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDHCPOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPOptionsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists(resourceName, &d),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`dhcp-options/dopt-.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain_name", fmt.Sprintf("service.%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "domain_name_servers.0", "127.0.0.1"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_servers.1", "10.0.0.2"),
					resource.TestCheckResourceAttr(resourceName, "ntp_servers.0", "127.0.0.1"),
					resource.TestCheckResourceAttr(resourceName, "netbios_name_servers.0", "127.0.0.1"),
					resource.TestCheckResourceAttr(resourceName, "netbios_node_type", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
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

func TestAccEC2VPCDHCPOptions_deleteOptions(t *testing.T) {
	var d ec2.DhcpOptions
	resourceName := "aws_vpc_dhcp_options.test"
	rName := sdkacctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDHCPOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPOptionsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists(resourceName, &d),
					testAccCheckDHCPOptionsDelete(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2VPCDHCPOptions_tags(t *testing.T) {
	var d ec2.DhcpOptions
	resourceName := "aws_vpc_dhcp_options.test"
	rName := sdkacctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDHCPOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPOptionsConfigTags1(rName, "key1", "value1"),
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
				Config: testAccDHCPOptionsConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists(resourceName, &d),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDHCPOptionsConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists(resourceName, &d),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEC2VPCDHCPOptions_disappears(t *testing.T) {
	var d ec2.DhcpOptions
	resourceName := "aws_vpc_dhcp_options.test"
	rName := sdkacctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDHCPOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPOptionsConfig(rName),
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

		// Try to find the resource
		resp, err := conn.DescribeDhcpOptions(&ec2.DescribeDhcpOptionsInput{
			DhcpOptionsIds: []*string{
				aws.String(rs.Primary.ID),
			},
		})
		if tfawserr.ErrMessageContains(err, "InvalidDhcpOptionID.NotFound", "") {
			continue
		}
		if err == nil {
			if len(resp.DhcpOptions) > 0 {
				return fmt.Errorf("still exists")
			}

			return nil
		}

		if !tfawserr.ErrMessageContains(err, "InvalidDhcpOptionID.NotFound", "") {
			return err
		}
	}

	return nil
}

func testAccCheckDHCPOptionsExists(n string, d *ec2.DhcpOptions) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		resp, err := conn.DescribeDhcpOptions(&ec2.DescribeDhcpOptionsInput{
			DhcpOptionsIds: []*string{
				aws.String(rs.Primary.ID),
			},
		})
		if err != nil {
			return err
		}
		if len(resp.DhcpOptions) == 0 {
			return fmt.Errorf("DHCP Options not found")
		}

		*d = *resp.DhcpOptions[0]

		return nil
	}
}

func testAccCheckDHCPOptionsDelete(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		_, err := conn.DeleteDhcpOptions(&ec2.DeleteDhcpOptionsInput{
			DhcpOptionsId: aws.String(rs.Primary.ID),
		})

		return err
	}
}

func testAccDHCPOptionsConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc_dhcp_options" "test" {
  domain_name          = "service.%s"
  domain_name_servers  = ["127.0.0.1", "10.0.0.2"]
  ntp_servers          = ["127.0.0.1"]
  netbios_name_servers = ["127.0.0.1"]
  netbios_node_type    = 2
}
`, rName)
}

func testAccDHCPOptionsConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc_dhcp_options" "test" {
  domain_name          = "service.%[1]s"
  domain_name_servers  = ["127.0.0.1", "10.0.0.2"]
  ntp_servers          = ["127.0.0.1"]
  netbios_name_servers = ["127.0.0.1"]
  netbios_node_type    = 2

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDHCPOptionsConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc_dhcp_options" "test" {
  domain_name          = "service.%[1]s"
  domain_name_servers  = ["127.0.0.1", "10.0.0.2"]
  ntp_servers          = ["127.0.0.1"]
  netbios_name_servers = ["127.0.0.1"]
  netbios_node_type    = 2

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
