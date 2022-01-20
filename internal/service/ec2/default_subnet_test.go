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
		CheckDestroy: testAccCheckDefaultSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultSubnetConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "assign_ipv6_address_on_creation", "false"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckDefaultSubnetDestroy(s *terraform.State) error {
	// We expect subnet to still exist
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
