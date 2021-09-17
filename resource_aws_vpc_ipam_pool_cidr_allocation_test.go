package aws

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSVpcIpamPoolAllocation_basic(t *testing.T) {
	var allocation ec2.IpamPoolAllocation
	resourceName := "aws_vpc_ipam_pool_cidr_allocation.test"
	cidr := "172.2.0.0/28"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsVpcIpamPoolAllocationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVpcIpamPoolAllocation(cidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsVpcIpamAllocationExists(resourceName, &allocation),
					resource.TestCheckResourceAttr(resourceName, "cidr", cidr),
					resource.TestMatchResourceAttr(resourceName, "id", regexp.MustCompile(`^alloc-([\da-f]{8})((-[\da-f]{4}){3})(-[\da-f]{12})_ipam-pool(-[\da-f]+)$`)),
					resource.TestMatchResourceAttr(resourceName, "allocation_id", regexp.MustCompile(`^alloc-([\da-f]{8})((-[\da-f]{4}){3})(-[\da-f]{12})$`)),
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

func TestAccAWSVpcIpamPoolAllocation_basicNetmask(t *testing.T) {
	var allocation ec2.IpamPoolAllocation
	resourceName := "aws_vpc_ipam_pool_cidr_allocation.test"
	netmask := "28"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsVpcIpamPoolAllocationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVpcIpamPoolAllocationNetmask(netmask),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsVpcIpamAllocationExists(resourceName, &allocation),
					testAccCheckVpcIpamCidrPrefix(&allocation, netmask),
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

func testAccCheckAwsVpcIpamAllocationExists(n string, allocation *ec2.IpamPoolAllocation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		id := rs.Primary.ID
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		cidr_allocation, _, err := findIpamPoolCidrAllocation(conn, id)

		if err != nil {
			return err
		}
		*allocation = *cidr_allocation

		return nil
	}
}

func testAccCheckVpcIpamCidrPrefix(allocation *ec2.IpamPoolAllocation, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if strings.Split(aws.StringValue(allocation.CidrBlock), "/")[1] != expected {
			return fmt.Errorf("Bad cidr prefix: %s", aws.StringValue(allocation.CidrBlock))
		}

		return nil
	}
}

func testAccCheckAwsVpcIpamPoolAllocationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_ipam_pool_cidr_allocation" {
			continue
		}

		id := rs.Primary.ID
		_, _, err := findIpamPoolCidrAllocation(conn, id)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, IpamPoolAllocationNotFound) || tfawserr.ErrCodeEquals(err, InvalidIpamPoolIdNotFound) {
				return nil
			}
			return err
		}

	}

	return nil
}

func testAccAwsVpcIpamPoolAllocation(cidr string) string {
	return testAccVpcIpamBase + fmt.Sprintf(`
resource "aws_vpc_ipam_pool_cidr_allocation" "test" {
	ipam_pool_id = aws_vpc_ipam_pool.test.id
	cidr         = %[1]q
	depends_on   = [
		aws_vpc_ipam_pool_cidr.test
	]
}
`, cidr)
}

func testAccAwsVpcIpamPoolAllocationNetmask(netmask string) string {
	return testAccVpcIpamBase + fmt.Sprintf(`
resource "aws_vpc_ipam_pool_cidr_allocation" "test" {
	ipam_pool_id   = aws_vpc_ipam_pool.test.id
	netmask_length = %[1]q
	depends_on     = [
		aws_vpc_ipam_pool_cidr.test
	]
}
`, netmask)
}
