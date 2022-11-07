package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccIPAMPoolCIDR_basic(t *testing.T) {
	var cidr ec2.IpamPoolCidr
	resourceName := "aws_vpc_ipam_pool_cidr.test"
	cidrBlock := "10.0.0.0/24"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMPoolCIDRDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolCIDRConfig_provisionedIPv4(cidrBlock),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMPoolCIDRExists(resourceName, &cidr),
					resource.TestCheckResourceAttr(resourceName, "cidr", cidrBlock),
					resource.TestCheckResourceAttrPair(resourceName, "ipam_pool_id", "aws_vpc_ipam_pool.test", "id"),
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

func TestAccIPAMPoolCIDR_disappears(t *testing.T) {
	var cidr ec2.IpamPoolCidr
	resourceName := "aws_vpc_ipam_pool_cidr.test"
	cidrBlock := "10.0.0.0/24"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMPoolCIDRDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolCIDRConfig_provisionedIPv4(cidrBlock),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMPoolCIDRExists(resourceName, &cidr),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceIPAMPoolCIDR(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIPAMPoolCIDR_Disappears_ipam(t *testing.T) {
	var cidr ec2.IpamPoolCidr
	resourceName := "aws_vpc_ipam_pool_cidr.test"
	ipamResourceName := "aws_vpc_ipam.test"
	cidrBlock := "10.0.0.0/24"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMPoolCIDRDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolCIDRConfig_provisionedIPv4(cidrBlock),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMPoolCIDRExists(resourceName, &cidr),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceIPAM(), ipamResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckIPAMPoolCIDRExists(n string, v *ec2.IpamPoolCidr) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IPAM Pool CIDR ID is set")
		}

		cidrBlock, poolID, err := tfec2.IPAMPoolCIDRParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindIPAMPoolCIDRByTwoPartKey(conn, cidrBlock, poolID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckIPAMPoolCIDRDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_ipam_pool_cidr" {
			continue
		}

		cidrBlock, poolID, err := tfec2.IPAMPoolCIDRParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfec2.FindIPAMPoolCIDRByTwoPartKey(conn, cidrBlock, poolID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("IPAM Pool CIDR still exists: %s", rs.Primary.ID)
	}

	return nil
}

const testAccIPAMPoolCIDRConfig_base = `
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  description = "test"

  operating_regions {
    region_name = data.aws_region.current.name
  }

  cascade = true
}
`

const testAccIPAMPoolCIDRConfig_privatePool = `
resource "aws_vpc_ipam_pool" "test" {
  address_family = "ipv4"
  ipam_scope_id  = aws_vpc_ipam.test.private_default_scope_id
  locale         = data.aws_region.current.name
}
`

func testAccIPAMPoolCIDRConfig_provisionedIPv4(cidr string) string {
	return acctest.ConfigCompose(testAccIPAMPoolCIDRConfig_base, testAccIPAMPoolCIDRConfig_privatePool, fmt.Sprintf(`
resource "aws_vpc_ipam_pool_cidr" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = %[1]q
}
`, cidr))
}
