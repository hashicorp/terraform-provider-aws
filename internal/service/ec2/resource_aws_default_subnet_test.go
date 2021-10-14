package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccAWSDefaultSubnet_basic(t *testing.T) {
	var v ec2.Subnet

	resourceName := "aws_default_subnet.foo"
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSDefaultSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDefaultSubnetConfigBasic(rInt),
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

func TestAccAWSDefaultSubnet_publicIp(t *testing.T) {
	var v ec2.Subnet

	resourceName := "aws_default_subnet.foo"
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSDefaultSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDefaultSubnetConfigPublicIp(rInt),
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
				Config: testAccAWSDefaultSubnetConfigNoPublicIp(rInt),
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

func testAccCheckAWSDefaultSubnetDestroy(s *terraform.State) error {
	// We expect subnet to still exist
	return nil
}

func testAccAWSDefaultSubnetConfigBasic(rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_default_subnet" "foo" {
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "terraform-testacc-default-subnet-%d"
  }
}
`, rInt))
}

func testAccAWSDefaultSubnetConfigPublicIp(rInt int) string {
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

func testAccAWSDefaultSubnetConfigNoPublicIp(rInt int) string {
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
