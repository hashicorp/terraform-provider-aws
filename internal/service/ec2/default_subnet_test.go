package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2DefaultSubnet_basic(t *testing.T) {
	var v ec2.Subnet

	resourceName := "aws_default_subnet.foo"
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDefaultSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultSubnetBasicConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(
						resourceName, "availability_zone", availabilityZonesDataSourceName, "names.0"),
					resource.TestCheckResourceAttrSet(
						resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttr(
						resourceName, "assign_ipv6_address_on_creation", "false"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.Name", fmt.Sprintf("terraform-testacc-default-subnet-%d", rInt)),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
				),
			},
		},
	})
}

func TestAccEC2DefaultSubnet_publicIP(t *testing.T) {
	var v ec2.Subnet

	resourceName := "aws_default_subnet.foo"
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDefaultSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultSubnetPublicIPConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(
						resourceName, "availability_zone", availabilityZonesDataSourceName, "names.1"),
					resource.TestCheckResourceAttr(
						resourceName, "map_public_ip_on_launch", "true"),
					resource.TestCheckResourceAttr(
						resourceName, "assign_ipv6_address_on_creation", "false"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.Name", fmt.Sprintf("terraform-testacc-default-subnet-%d", rInt)),
				),
			},
			{
				Config: testAccDefaultSubnetNoPublicIPConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(
						resourceName, "availability_zone", availabilityZonesDataSourceName, "names.1"),
					resource.TestCheckResourceAttr(
						resourceName, "map_public_ip_on_launch", "false"),
					resource.TestCheckResourceAttr(
						resourceName, "assign_ipv6_address_on_creation", "false"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.Name", fmt.Sprintf("terraform-testacc-default-subnet-%d", rInt)),
				),
			},
		},
	})
}

func testAccCheckDefaultSubnetDestroy(s *terraform.State) error {
	// We expect subnet to still exist
	return nil
}

func testAccDefaultSubnetBasicConfig(rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_default_subnet" "foo" {
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "terraform-testacc-default-subnet-%d"
  }
}
`, rInt))
}

func testAccDefaultSubnetPublicIPConfig(rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_default_subnet" "foo" {
  availability_zone       = data.aws_availability_zones.available.names[1]
  map_public_ip_on_launch = true

  tags = {
    Name = "terraform-testacc-default-subnet-%d"
  }
}
`, rInt))
}

func testAccDefaultSubnetNoPublicIPConfig(rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_default_subnet" "foo" {
  availability_zone       = data.aws_availability_zones.available.names[1]
  map_public_ip_on_launch = false

  tags = {
    Name = "terraform-testacc-default-subnet-%d"
  }
}
`, rInt))
}
