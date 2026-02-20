// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// The default quota is 5 secondary networks per region. Serialize at the
// resource test level to ensure the total number of networks will not exceed
// the quota, even when run in parallel with other resource tests.
func TestAccEC2SecondarySubnet_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccEC2SecondarySubnet_basic,
		acctest.CtDisappears: testAccEC2SecondarySubnet_disappears,
		"tags":               testAccEC2SecondarySubnet_tags,
		"availabilityZoneID": testAccEC2SecondarySubnet_availabilityZoneID,
		"Identity":           testAccEC2SecondarySubnet_identitySerial,
		"List":               testAccEC2SecondarySubnet_listSerial,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccEC2SecondarySubnet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_secondary_subnet.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSecondaryNetwork(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecondarySubnetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecondarySubnetConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecondarySubnetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ipv4_cidr_block", "10.0.0.0/24"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ec2", regexache.MustCompile(`secondary-subnet/ss-.+`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrSet(resourceName, "secondary_network_id"),
					resource.TestCheckResourceAttrSet(resourceName, "secondary_network_type"),
					resource.TestCheckResourceAttrSet(resourceName, "secondary_subnet_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrState),
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

func testAccEC2SecondarySubnet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_secondary_subnet.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSecondaryNetwork(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecondarySubnetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecondarySubnetConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecondarySubnetExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfec2.ResourceSecondarySubnet, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccEC2SecondarySubnet_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_secondary_subnet.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSecondaryNetwork(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecondarySubnetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecondarySubnetConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecondarySubnetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSecondarySubnetConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecondarySubnetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccSecondarySubnetConfig_tags1(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecondarySubnetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccEC2SecondarySubnet_availabilityZoneID(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_secondary_subnet.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSecondaryNetwork(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecondarySubnetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecondarySubnetConfig_availabilityZoneID(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecondarySubnetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ipv4_cidr_block", "10.0.0.0/24"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
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

func testAccCheckSecondarySubnetExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		_, err := tfec2.FindSecondarySubnetByID(ctx, conn, rs.Primary.ID)
		return err
	}
}

func testAccCheckSecondarySubnetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_secondary_subnet" {
				continue
			}

			_, err := tfec2.FindSecondarySubnetByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Secondary Subnet %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccSecondarySubnetConfig_base() string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		`
resource "aws_ec2_secondary_network" "test" {
  ipv4_cidr_block = "10.0.0.0/16"
  network_type    = "rdma"
}
`)
}

func testAccSecondarySubnetConfig_basic() string {
	return acctest.ConfigCompose(
		testAccSecondarySubnetConfig_base(),
		`
resource "aws_ec2_secondary_subnet" "test" {
  secondary_network_id = aws_ec2_secondary_network.test.id
  ipv4_cidr_block      = "10.0.0.0/24"
  availability_zone    = data.aws_availability_zones.available.names[0]
}
`)
}

func testAccSecondarySubnetConfig_tags1(tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccSecondarySubnetConfig_base(),
		fmt.Sprintf(`
resource "aws_ec2_secondary_subnet" "test" {
  secondary_network_id = aws_ec2_secondary_network.test.id
  ipv4_cidr_block      = "10.0.0.0/24"
  availability_zone    = data.aws_availability_zones.available.names[0]

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccSecondarySubnetConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccSecondarySubnetConfig_base(),
		fmt.Sprintf(`
resource "aws_ec2_secondary_subnet" "test" {
  secondary_network_id = aws_ec2_secondary_network.test.id
  ipv4_cidr_block      = "10.0.0.0/24"
  availability_zone    = data.aws_availability_zones.available.names[0]

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccSecondarySubnetConfig_availabilityZoneID() string {
	return acctest.ConfigCompose(
		testAccSecondarySubnetConfig_base(),
		`
resource "aws_ec2_secondary_subnet" "test" {
  secondary_network_id = aws_ec2_secondary_network.test.id
  ipv4_cidr_block      = "10.0.0.0/24"
  availability_zone_id = data.aws_availability_zones.available.zone_ids[0]
}
`)
}
