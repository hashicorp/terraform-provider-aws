package ec2_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccIPAMPoolCIDRAllocation_ipv4Basic(t *testing.T) {
	var allocation ec2.IpamPoolAllocation
	resourceName := "aws_vpc_ipam_pool_cidr_allocation.test"
	cidr := "172.2.0.0/28"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMPoolAllocationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolCIDRAllocationConfig_ipv4(cidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMPoolCIDRAllocationExists(resourceName, &allocation),
					resource.TestCheckResourceAttr(resourceName, "cidr", cidr),
					resource.TestMatchResourceAttr(resourceName, "id", regexp.MustCompile(`^ipam-pool-alloc-[\da-f]+_ipam-pool(-[\da-f]+)$`)),
					resource.TestMatchResourceAttr(resourceName, "ipam_pool_allocation_id", regexp.MustCompile(`^ipam-pool-alloc-[\da-f]+$`)),
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

func TestAccIPAMPoolCIDRAllocation_disappears(t *testing.T) {
	var allocation ec2.IpamPoolAllocation
	resourceName := "aws_vpc_ipam_pool_cidr_allocation.test"
	cidr := "172.2.0.0/28"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMPoolAllocationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolCIDRAllocationConfig_ipv4(cidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMPoolCIDRAllocationExists(resourceName, &allocation),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceIPAMPoolCIDRAllocation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIPAMPoolCIDRAllocation_ipv4BasicNetmask(t *testing.T) {
	var allocation ec2.IpamPoolAllocation
	resourceName := "aws_vpc_ipam_pool_cidr_allocation.test"
	netmask := "28"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMPoolAllocationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolCIDRAllocationConfig_ipv4Netmask(netmask),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMPoolCIDRAllocationExists(resourceName, &allocation),
					testAccCheckIPAMCIDRPrefix(&allocation, netmask),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"netmask_length"},
			},
		},
	})
}

func TestAccIPAMPoolCIDRAllocation_ipv4DisallowedCIDR(t *testing.T) {
	var allocation ec2.IpamPoolAllocation
	resourceName := "aws_vpc_ipam_pool_cidr_allocation.test"
	disallowedCidr := "172.2.0.0/28"
	netmaskLength := "28"
	expectedCidr := "172.2.0.16/28"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolCIDRAllocationConfig_ipv4Disallowed(netmaskLength, disallowedCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMPoolCIDRAllocationExists(resourceName, &allocation),
					resource.TestCheckResourceAttr(resourceName, "cidr", expectedCidr),
					resource.TestCheckResourceAttr(resourceName, "disallowed_cidrs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "disallowed_cidrs.0", disallowedCidr),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "ipam_pool_id", "aws_vpc_ipam_pool.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "netmask_length", netmaskLength),
				),
			},
		},
	})
}

func TestAccIPAMPoolCIDRAllocation_multiple(t *testing.T) {
	var allocation1, allocation2 ec2.IpamPoolAllocation
	resourceName := "aws_vpc_ipam_pool_cidr_allocation.test1"
	resourceName2 := "aws_vpc_ipam_pool_cidr_allocation.test2"
	cidr1 := "172.2.0.0/28"
	cidr2 := "10.1.0.0/28"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMPoolAllocationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolCIDRAllocationConfig_multiple(cidr1, cidr2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMPoolCIDRAllocationExists(resourceName, &allocation1),
					testAccCheckIPAMPoolCIDRAllocationExists(resourceName2, &allocation2),
					resource.TestCheckResourceAttr(resourceName, "cidr", cidr1),
					resource.TestMatchResourceAttr(resourceName, "id", regexp.MustCompile(`^ipam-pool-alloc-[\da-f]+_ipam-pool(-[\da-f]+)$`)),
					resource.TestMatchResourceAttr(resourceName, "ipam_pool_allocation_id", regexp.MustCompile(`^ipam-pool-alloc-[\da-f]+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "ipam_pool_id", "aws_vpc_ipam_pool.test", "id"),
					resource.TestCheckResourceAttr(resourceName2, "cidr", cidr2),
					resource.TestMatchResourceAttr(resourceName2, "id", regexp.MustCompile(`^ipam-pool-alloc-[\da-f]+_ipam-pool(-[\da-f]+)$`)),
					resource.TestMatchResourceAttr(resourceName2, "ipam_pool_allocation_id", regexp.MustCompile(`^ipam-pool-alloc-[\da-f]+$`)),
					resource.TestCheckResourceAttrPair(resourceName2, "ipam_pool_id", "aws_vpc_ipam_pool.test", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      resourceName2,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckIPAMCIDRPrefix(allocation *ec2.IpamPoolAllocation, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if strings.Split(aws.StringValue(allocation.Cidr), "/")[1] != expected {
			return fmt.Errorf("Bad cidr prefix: %s", aws.StringValue(allocation.Cidr))
		}

		return nil
	}
}

func testAccCheckIPAMPoolCIDRAllocationExists(n string, v *ec2.IpamPoolAllocation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IPAM Pool CIDR Allocation ID is set")
		}

		allocationID, poolID, err := tfec2.IPAMPoolCIDRAllocationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindIPAMPoolAllocationByTwoPartKey(conn, allocationID, poolID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckIPAMPoolAllocationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_ipam_pool_cidr_allocation" {
			continue
		}

		allocationID, poolID, err := tfec2.IPAMPoolCIDRAllocationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfec2.FindIPAMPoolAllocationByTwoPartKey(conn, allocationID, poolID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("IPAM Pool CIDR Allocation still exists: %s", rs.Primary.ID)
	}

	return nil
}

const testAccIPAMPoolCIDRAllocationConfig_base = `
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  operating_regions {
    region_name = data.aws_region.current.name
  }
}

resource "aws_vpc_ipam_pool" "test" {
  address_family = "ipv4"
  ipam_scope_id  = aws_vpc_ipam.test.private_default_scope_id
  locale         = data.aws_region.current.name
}

resource "aws_vpc_ipam_pool_cidr" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = "172.2.0.0/24"
}
`

func testAccIPAMPoolCIDRAllocationConfig_ipv4(cidr string) string {
	return acctest.ConfigCompose(testAccIPAMPoolCIDRAllocationConfig_base, fmt.Sprintf(`
resource "aws_vpc_ipam_pool_cidr_allocation" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = %[1]q

  depends_on = [
    aws_vpc_ipam_pool_cidr.test
  ]
}
`, cidr))
}

func testAccIPAMPoolCIDRAllocationConfig_ipv4Netmask(netmask string) string {
	return acctest.ConfigCompose(testAccIPAMPoolCIDRAllocationConfig_base, fmt.Sprintf(`
resource "aws_vpc_ipam_pool_cidr_allocation" "test" {
  ipam_pool_id   = aws_vpc_ipam_pool.test.id
  netmask_length = %[1]q

  depends_on = [
    aws_vpc_ipam_pool_cidr.test
  ]
}
`, netmask))
}

func testAccIPAMPoolCIDRAllocationConfig_ipv4Disallowed(netmaskLength, disallowedCidr string) string {
	return acctest.ConfigCompose(testAccIPAMPoolCIDRAllocationConfig_base, fmt.Sprintf(`
resource "aws_vpc_ipam_pool_cidr_allocation" "test" {
  ipam_pool_id   = aws_vpc_ipam_pool.test.id
  netmask_length = %[1]q

  disallowed_cidrs = [
    %[2]q
  ]

  depends_on = [
    aws_vpc_ipam_pool_cidr.test
  ]
}
`, netmaskLength, disallowedCidr))
}

func testAccIPAMPoolCIDRAllocationConfig_multiple(cidr1, cidr2 string) string {
	return acctest.ConfigCompose(testAccIPAMPoolCIDRAllocationConfig_base, fmt.Sprintf(`
resource "aws_vpc_ipam_pool_cidr_allocation" "test1" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = %[1]q

  depends_on = [
    aws_vpc_ipam_pool_cidr.test
  ]
}

resource "aws_vpc_ipam_pool_cidr" "test2" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = "10.1.0.0/24"
}

resource "aws_vpc_ipam_pool_cidr_allocation" "test2" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = %[2]q
  depends_on = [
    aws_vpc_ipam_pool_cidr.test2
  ]
}
`, cidr1, cidr2))
}
