package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccVPCIpamPreviewNextCidr_ipv4Basic(t *testing.T) {
	resourceName := "aws_vpc_ipam_preview_next_cidr.test"
	netmaskLength := "28"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccIPAMPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIpamPreviewNextCidrIpv4Basic(netmaskLength),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "cidr"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "ipam_pool_id", "aws_vpc_ipam_pool.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "netmask_length", netmaskLength),
				),
			},
		},
	})
}

func TestAccVPCIpamPreviewNextCidr_ipv4Allocated(t *testing.T) {
	resourceName := "aws_vpc_ipam_preview_next_cidr.test"
	netmaskLength := "28"
	allocatedCidr := "172.2.0.0/28"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccIPAMPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIpamPreviewNextCidrIpv4Basic(netmaskLength),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "cidr", allocatedCidr),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "ipam_pool_id", "aws_vpc_ipam_pool.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "netmask_length", netmaskLength),
				),
			},
			{
				Config: testAccVPCIpamPreviewNextCidrIpv4Allocated(netmaskLength),
				Check: resource.ComposeTestCheckFunc(
					// cidr should not change even after allocation
					resource.TestCheckResourceAttr(resourceName, "cidr", allocatedCidr),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "ipam_pool_id", "aws_vpc_ipam_pool.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "netmask_length", netmaskLength),
				),
			},
		},
	})
}

// // temp comment out till bug is resolved
// func TestAccVPCIpamPreviewNextCidr_ipv4DisallowedCidr(t *testing.T) {
// 	resourceName := "aws_vpc_ipam_preview_next_cidr.test"
// 	disallowedCidr := "172.2.0.0/32"
// 	netmaskLength := "28"

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck:     func() { acctest.PreCheck(t); testAccIPAMPreCheck(t) },
// 		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
// 		Providers:    acctest.Providers,
// 		CheckDestroy: nil,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccVPCIpamPreviewNextCidrIpv4DisallowedCidr(netmaskLength, disallowedCidr),
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttrSet(resourceName, "cidr"),
// 					resource.TestCheckResourceAttr(resourceName, "disallowed_cidrs.#", "1"),
// 					resource.TestCheckResourceAttr(resourceName, "disallowed_cidrs.0", disallowedCidr),
// 					resource.TestCheckResourceAttrSet(resourceName, "id"),
// 					resource.TestCheckResourceAttrPair(resourceName, "ipam_pool_id", "aws_vpc_ipam_pool.test", "id"),
// 					resource.TestCheckResourceAttr(resourceName, "netmask_length", netmaskLength),
// 				),
// 			},
// 		},
// 	})
// }

const testAccVPCIpamPreviewNextCidrIpv4Base = `
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

func testAccVPCIpamPreviewNextCidrIpv4Basic(netmaskLength string) string {
	return acctest.ConfigCompose(
		testAccVPCIpamPreviewNextCidrIpv4Base,
		fmt.Sprintf(`
resource "aws_vpc_ipam_preview_next_cidr" "test" {
  ipam_pool_id   = aws_vpc_ipam_pool.test.id
  netmask_length = %[1]q

  depends_on = [
    aws_vpc_ipam_pool_cidr.test
  ]
}
`, netmaskLength))
}

func testAccVPCIpamPreviewNextCidrIpv4Allocated(netmaskLength string) string {
	return acctest.ConfigCompose(
		testAccVPCIpamPreviewNextCidrIpv4Base,
		fmt.Sprintf(`
resource "aws_vpc_ipam_preview_next_cidr" "test" {
  ipam_pool_id   = aws_vpc_ipam_pool.test.id
  netmask_length = %[1]q

  depends_on = [
    aws_vpc_ipam_pool_cidr.test
  ]
}

resource "aws_vpc_ipam_pool_cidr_allocation" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = aws_vpc_ipam_preview_next_cidr.test.cidr
}
`, netmaskLength))
}

// // temp comment out till bug is resolved
// func testAccVPCIpamPreviewNextCidrIpv4DisallowedCidr(netmaskLength, disallowedCidr string) string {
// 	return testAccVPCIpamPreviewNextCidrIpv4Base + fmt.Sprintf(`
// resource "aws_vpc_ipam_preview_next_cidr" "test" {
//   ipam_pool_id   = aws_vpc_ipam_pool.test.id
//   netmask_length = %[1]q

//   disallowed_cidrs = [
//     %[2]q
//   ]

//   depends_on = [
//     aws_vpc_ipam_pool_cidr.test
//   ]
// }
// `, netmaskLength, disallowedCidr)
// }
