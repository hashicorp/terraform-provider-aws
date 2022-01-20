package ec2_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func testAccPreCheckDefaultSubnetAvailable(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeSubnetsInput{
		Filters: tfec2.BuildAttributeFilterList(
			map[string]string{
				"defaultForAz": "true",
			},
		),
	}

	subnets, err := tfec2.FindSubnets(conn, input)

	if err != nil {
		t.Fatalf("error listing default subnets: %s", err)
	}

	if len(subnets) == 0 {
		t.Skip("skipping since no default subnet is available")
	}
}

func testAccEC2DefaultSubnet_basic(t *testing.T) {
	var v ec2.Subnet
	resourceName := "aws_default_subnet.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckDefaultSubnetAvailable(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDefaultSubnetDestroyExists,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultSubnetConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrSet(resourceName, "cidr_block"),
					resource.TestCheckResourceAttr(resourceName, "customer_owned_ipv4_pool", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_dns64", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_aaaa_record_on_launch", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_a_record_on_launch", "false"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_native", "false"),
					resource.TestCheckResourceAttr(resourceName, "map_customer_owned_ip_on_launch", "false"),
					resource.TestCheckResourceAttr(resourceName, "map_public_ip_on_launch", "true"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_hostname_type_on_launch", "ip-name"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
				),
			},
		},
	})
}

// testAccCheckDefaultSubnetDestroyExists runs after all resources are destroyed.
// It verifies that the default subnet still exists.
func testAccCheckDefaultSubnetDestroyExists(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_subnet" {
			continue
		}

		_, err := tfec2.FindSubnetByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}
	}

	return nil
}

const testAccDefaultSubnetConfigBaseExisting = `
data "aws_subnets" "test" {
  filter {
    name   = "defaultForAz"
    values = ["true"]
  }
}

data "aws_subnet" "test" {
  id = data.aws_subnets.test.ids[0]
}
`

func testAccDefaultSubnetConfig() string {
	return acctest.ConfigCompose(testAccDefaultSubnetConfigBaseExisting, `
resource "aws_default_subnet" "test" {
  availability_zone = data.aws_subnet.test.availability_zone
}
`)
}
