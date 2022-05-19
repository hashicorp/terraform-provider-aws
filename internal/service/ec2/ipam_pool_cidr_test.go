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

func TestAccIPAMPool_cidrIPv4Basic(t *testing.T) {
	var cidr ec2.IpamPoolCidr
	resourceName := "aws_vpc_ipam_pool_cidr.test"
	cidr_range := "10.0.0.0/24"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccIPAMPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIPAMProvisionedPoolCIDRDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMConfig_provisionedPoolCIDRIPv4(cidr_range),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMCIDRExists(resourceName, &cidr),
					resource.TestCheckResourceAttr(resourceName, "cidr", cidr_range),
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

func testAccCheckIPAMCIDRExists(n string, cidr *ec2.IpamPoolCidr) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		id := rs.Primary.ID
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		found_cidr, _, err := tfec2.FindIPAMPoolCIDR(conn, id)

		if err != nil {
			return err
		}
		*cidr = *found_cidr

		return nil
	}
}

func testAccCheckIPAMProvisionedPoolCIDRDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_ipam_pool_cidr" {
			continue
		}

		id := rs.Primary.ID

		_, pool_id, err := tfec2.DecodeIPAMPoolCIDRID(id)
		if err != nil {
			return fmt.Errorf("error decoding ID (%s): %w", id, err)
		}

		if _, err = tfec2.WaitIPAMPoolDeleted(conn, pool_id, tfec2.IPAMPoolDeleteTimeout); err != nil {
			if tfresource.NotFound(err) {
				return nil
			}
			return fmt.Errorf("error waiting for IPAM Pool (%s) to be deleted: %w", id, err)
		}
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
}
`

const testAccIPAMPoolCIDRConfig_privatePool = `
resource "aws_vpc_ipam_pool" "test" {
  address_family = "ipv4"
  ipam_scope_id  = aws_vpc_ipam.test.private_default_scope_id
  locale         = data.aws_region.current.name
}
`

func testAccIPAMConfig_provisionedPoolCIDRIPv4(cidr string) string {
	return testAccIPAMPoolCIDRConfig_base + testAccIPAMPoolCIDRConfig_privatePool + fmt.Sprintf(`
resource "aws_vpc_ipam_pool_cidr" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = %[1]q
}
`, cidr)
}
