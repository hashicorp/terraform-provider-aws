package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSVpcIpamPoolCidr_basic(t *testing.T) {
	var cidr ec2.IpamPoolCidr
	resourceName := "aws_vpc_ipam_pool_cidr.test"
	cidr_range := "10.0.0.0/24"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsVpcIpamProvisionedPoolCidrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVpcIpamProvisionedPoolCidr(cidr_range),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsVpcIpamCidrExists(resourceName, &cidr),
					resource.TestCheckResourceAttr(resourceName, "cidr", cidr_range),
					resource.TestCheckResourceAttrPair(resourceName, "ipam_pool_id", "aws_vpc_ipam_pool.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "state", "provisioned"),
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

func testAccCheckAwsVpcIpamCidrExists(n string, cidr *ec2.IpamPoolCidr) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		id := rs.Primary.ID
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		found_cidr, _, err := findIpamPoolCidr(conn, id)

		if err != nil {
			return err
		}
		*cidr = *found_cidr

		return nil
	}
}

func testAccCheckAwsVpcIpamProvisionedPoolCidrDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_ipam_pool_cidr" {
			continue
		}

		id := rs.Primary.ID

		_, pool_id, err := DecodeIpamPoolCidrID(id)
		if err != nil {
			return fmt.Errorf("error decoding ID (%s): %w", id, err)
		}

		if _, err = waitIpamPoolDeleted(conn, pool_id, IpamPoolDeleteTimeout); err != nil {
			if isResourceNotFoundError(err) {
				return nil
			}
			return fmt.Errorf("error waiting for IPAM Pool (%s) to be deleted: %w", id, err)
		}
	}

	return nil
}

const testAccAwsVpcIpamPrivatePool = `
resource "aws_vpc_ipam_pool" "test" {
    address_family = "ipv4"
    ipam_scope_id  = aws_vpc_ipam.test.private_default_scope_id
	locale         = data.aws_region.current.name
}
`

func testAccAwsVpcIpamProvisionedPoolCidr(cidr string) string {
	return testAccAwsVpcIpam + testAccAwsVpcIpamPrivatePool + fmt.Sprintf(`
resource "aws_vpc_ipam_pool_cidr" "test" {
	ipam_pool_id = aws_vpc_ipam_pool.test.id
	cidr         = %[1]q
}
`, cidr)
}
