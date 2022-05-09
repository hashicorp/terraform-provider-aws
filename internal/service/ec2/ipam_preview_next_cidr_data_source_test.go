package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccIPAMPreviewNextCIDRDataSource_ipv4Basic(t *testing.T) {
	datasourceName := "data.aws_vpc_ipam_preview_next_cidr.test"
	netmaskLength := "28"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccIPAMPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVPCIpamPreviewNextCidrIpv4Basic(netmaskLength),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceName, "cidr"),
					resource.TestCheckResourceAttrPair(datasourceName, "ipam_pool_id", "aws_vpc_ipam_pool.test", "id"),
					resource.TestCheckResourceAttr(datasourceName, "netmask_length", netmaskLength),
				),
			},
		},
	})
}

func TestAccIPAMPreviewNextCIDRDataSource_ipv4Allocated(t *testing.T) {
	datasourceName := "data.aws_vpc_ipam_preview_next_cidr.test"
	netmaskLength := "28"
	allocatedCidr := "172.2.0.0/28"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccIPAMPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVPCIpamPreviewNextCidrIpv4Basic(netmaskLength),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "cidr", allocatedCidr),
					resource.TestCheckResourceAttrPair(datasourceName, "ipam_pool_id", "aws_vpc_ipam_pool.test", "id"),
					resource.TestCheckResourceAttr(datasourceName, "netmask_length", netmaskLength),
				),
			},
			{
				Config: testAccDataSourceVPCIpamPreviewNextCidrIpv4Allocated(netmaskLength),
				Check: resource.ComposeTestCheckFunc(
					// cidr should not change even after allocation
					resource.TestCheckResourceAttr(datasourceName, "cidr", allocatedCidr),
					resource.TestCheckResourceAttrPair(datasourceName, "ipam_pool_id", "aws_vpc_ipam_pool.test", "id"),
					resource.TestCheckResourceAttr(datasourceName, "netmask_length", netmaskLength),
				),
			},
		},
	})
}

func TestAccIPAMPreviewNextCIDRDataSource_ipv4DisallowedCidr(t *testing.T) {
	datasourceName := "data.aws_vpc_ipam_preview_next_cidr.test"
	disallowedCidr := "172.2.0.0/28"
	netmaskLength := "28"
	expectedCidr := "172.2.0.16/28"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccIPAMPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVPCIpamPreviewNextCidrIpv4DisallowedCidr(netmaskLength, disallowedCidr),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "cidr", expectedCidr),
					resource.TestCheckResourceAttr(datasourceName, "disallowed_cidrs.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "disallowed_cidrs.0", disallowedCidr),
					resource.TestCheckResourceAttrPair(datasourceName, "ipam_pool_id", "aws_vpc_ipam_pool.test", "id"),
					resource.TestCheckResourceAttr(datasourceName, "netmask_length", netmaskLength),
				),
			},
		},
	})
}

const testAccDataSourceVPCIpamPreviewNextCidrIpv4Base = `
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  description = "test"
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

func testAccDataSourceVPCIpamPreviewNextCidrIpv4Basic(netmaskLength string) string {
	return acctest.ConfigCompose(
		testAccDataSourceVPCIpamPreviewNextCidrIpv4Base,
		fmt.Sprintf(`
data "aws_vpc_ipam_preview_next_cidr" "test" {
  ipam_pool_id   = aws_vpc_ipam_pool.test.id
  netmask_length = %[1]q

  depends_on = [
    aws_vpc_ipam_pool_cidr.test
  ]
}
`, netmaskLength))
}

func testAccDataSourceVPCIpamPreviewNextCidrIpv4Allocated(netmaskLength string) string {
	return acctest.ConfigCompose(
		testAccDataSourceVPCIpamPreviewNextCidrIpv4Base,
		fmt.Sprintf(`
data "aws_vpc_ipam_preview_next_cidr" "test" {
  ipam_pool_id   = aws_vpc_ipam_pool.test.id
  netmask_length = %[1]q

  depends_on = [
    aws_vpc_ipam_pool_cidr.test
  ]
}

resource "aws_vpc_ipam_pool_cidr_allocation" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = data.aws_vpc_ipam_preview_next_cidr.test.cidr

  lifecycle {
    ignore_changes = [cidr]
  }
}
`, netmaskLength))
}

func testAccDataSourceVPCIpamPreviewNextCidrIpv4DisallowedCidr(netmaskLength, disallowedCidr string) string {
	return testAccDataSourceVPCIpamPreviewNextCidrIpv4Base + fmt.Sprintf(`
data "aws_vpc_ipam_preview_next_cidr" "test" {
  ipam_pool_id   = aws_vpc_ipam_pool.test.id
  netmask_length = %[1]q

  disallowed_cidrs = [
    %[2]q
  ]

  depends_on = [
    aws_vpc_ipam_pool_cidr.test
  ]
}
`, netmaskLength, disallowedCidr)
}
