package ec2_test

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccVPCDefaultVPCDHCPOptions_basic(t *testing.T) {
	var d ec2.DhcpOptions
	resourceName := "aws_default_vpc_dhcp_options.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDefaultVPCDHCPOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultVPCDHCPOptionsBasicConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists(resourceName, &d),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`dhcp-options/dopt-.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain_name", tfec2.RegionalPrivateDNSSuffix(acctest.Region())),
					resource.TestCheckResourceAttr(resourceName, "domain_name_servers", "AmazonProvidedDNS"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Default DHCP Option Set"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
				),
			},
		},
	})
}

func TestAccVPCDefaultVPCDHCPOptions_owner(t *testing.T) {
	var d ec2.DhcpOptions
	resourceName := "aws_default_vpc_dhcp_options.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDefaultVPCDHCPOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultVPCDHCPOptionsOwnerConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists(resourceName, &d),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`dhcp-options/dopt-.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain_name", tfec2.RegionalPrivateDNSSuffix(acctest.Region())),
					resource.TestCheckResourceAttr(resourceName, "domain_name_servers", "AmazonProvidedDNS"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Default DHCP Option Set"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
				),
			},
		},
	})
}

func testAccCheckDefaultVPCDHCPOptionsDestroy(s *terraform.State) error {
	// We expect DHCP Options Set to still exist
	return nil
}

const testAccDefaultVPCDHCPOptionsBasicConfig = `
resource "aws_default_vpc_dhcp_options" "test" {
  tags = {
    Name = "Default DHCP Option Set"
  }
}
`

const testAccDefaultVPCDHCPOptionsOwnerConfig = `
data "aws_caller_identity" "current" {}

resource "aws_default_vpc_dhcp_options" "test" {
  owner_id = data.aws_caller_identity.current.account_id

  tags = {
    Name = "Default DHCP Option Set"
  }
}
`
