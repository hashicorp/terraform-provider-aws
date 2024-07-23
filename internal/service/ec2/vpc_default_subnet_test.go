// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccPreCheckDefaultSubnetExists(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeSubnetsInput{
		Filters: tfec2.NewAttributeFilterList(
			map[string]string{
				"defaultForAz": acctest.CtTrue,
			},
		),
	}

	subnets, err := tfec2.FindSubnets(ctx, conn, input)

	if err != nil {
		t.Fatalf("error listing default subnets: %s", err)
	}

	if len(subnets) == 0 {
		t.Skip("skipping since no default subnet is available")
	}
}

func testAccPreCheckDefaultSubnetNotFound(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeSubnetsInput{
		Filters: tfec2.NewAttributeFilterList(
			map[string]string{
				"defaultForAz": acctest.CtTrue,
			},
		),
	}

	subnets, err := tfec2.FindSubnets(ctx, conn, input)

	if err != nil {
		t.Fatalf("error listing default subnets: %s", err)
	}

	for _, v := range subnets {
		subnetID := aws.ToString(v.SubnetId)

		t.Logf("Deleting existing default subnet: %s", subnetID)

		r := tfec2.ResourceSubnet()
		d := r.Data(nil)
		d.SetId(subnetID)

		err := acctest.DeleteResource(ctx, r, d, acctest.Provider.Meta())

		if err != nil {
			t.Fatalf("error deleting default subnet: %s", err)
		}
	}
}

func testAccDefaultSubnet_Existing_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Subnet
	resourceName := "aws_default_subnet.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegionNot(t, names.USWest2RegionID, names.USGovWest1RegionID)
			testAccPreCheckDefaultSubnetExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultSubnetDestroyExists(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultSubnetConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "assign_ipv6_address_on_creation", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCIDRBlock),
					resource.TestCheckResourceAttr(resourceName, "customer_owned_ipv4_pool", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_dns64", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_lni_at_device_index", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_aaaa_record_on_launch", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_a_record_on_launch", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "existing_default_subnet", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_native", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "map_customer_owned_ip_on_launch", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "map_public_ip_on_launch", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrSet(resourceName, "private_dns_hostname_type_on_launch"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVPCID),
				),
			},
		},
	})
}

func testAccDefaultSubnet_Existing_forceDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Subnet
	resourceName := "aws_default_subnet.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegionNot(t, endpoints.UsWest2RegionID, endpoints.UsGovWest1RegionID)
			testAccPreCheckDefaultSubnetExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultSubnetDestroyNotFound(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultSubnetConfig_forceDestroy(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "existing_default_subnet", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtTrue),
				),
			},
		},
	})
}

func testAccDefaultSubnet_Existing_ipv6(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Subnet
	resourceName := "aws_default_subnet.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegionNot(t, endpoints.UsWest2RegionID, endpoints.UsGovWest1RegionID)
			testAccPreCheckDefaultSubnetExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultSubnetDestroyNotFound(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultSubnetConfig_ipv6(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "assign_ipv6_address_on_creation", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCIDRBlock),
					resource.TestCheckResourceAttr(resourceName, "customer_owned_ipv4_pool", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_dns64", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_lni_at_device_index", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_aaaa_record_on_launch", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_a_record_on_launch", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "existing_default_subnet", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_cidr_block"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_native", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "map_customer_owned_ip_on_launch", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "map_public_ip_on_launch", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, "private_dns_hostname_type_on_launch", "ip-name"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVPCID),
				),
			},
		},
	})
}

func testAccDefaultSubnet_Existing_privateDNSNameOptionsOnLaunch(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Subnet
	resourceName := "aws_default_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegionNot(t, endpoints.UsWest2RegionID, endpoints.UsGovWest1RegionID)
			testAccPreCheckDefaultSubnetExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultSubnetDestroyExists(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultSubnetConfig_privateDNSNameOptionsOnLaunch(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "assign_ipv6_address_on_creation", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCIDRBlock),
					resource.TestCheckResourceAttr(resourceName, "customer_owned_ipv4_pool", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_dns64", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_lni_at_device_index", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_aaaa_record_on_launch", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_a_record_on_launch", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "existing_default_subnet", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_native", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "map_customer_owned_ip_on_launch", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "map_public_ip_on_launch", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, "private_dns_hostname_type_on_launch", "resource-name"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVPCID),
				),
			},
		},
	})
}

func testAccDefaultSubnet_NotFound_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Subnet
	resourceName := "aws_default_subnet.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegionNot(t, endpoints.UsWest2RegionID, endpoints.UsGovWest1RegionID)
			testAccPreCheckDefaultSubnetNotFound(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultSubnetDestroyExists(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultSubnetConfig_notFound(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "assign_ipv6_address_on_creation", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCIDRBlock),
					resource.TestCheckResourceAttr(resourceName, "customer_owned_ipv4_pool", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_dns64", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_lni_at_device_index", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_aaaa_record_on_launch", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_a_record_on_launch", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "existing_default_subnet", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_native", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "map_customer_owned_ip_on_launch", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "map_public_ip_on_launch", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrSet(resourceName, "private_dns_hostname_type_on_launch"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVPCID),
				),
			},
		},
	})
}

func testAccDefaultSubnet_NotFound_ipv6Native(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Subnet
	resourceName := "aws_default_subnet.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegionNot(t, endpoints.UsWest2RegionID, endpoints.UsGovWest1RegionID)
			testAccPreCheckDefaultSubnetNotFound(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultSubnetDestroyNotFound(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultSubnetConfig_ipv6Native(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "assign_ipv6_address_on_creation", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrCIDRBlock, ""),
					resource.TestCheckResourceAttr(resourceName, "customer_owned_ipv4_pool", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_dns64", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_lni_at_device_index", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_aaaa_record_on_launch", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_a_record_on_launch", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "existing_default_subnet", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_cidr_block"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_native", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "map_customer_owned_ip_on_launch", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "map_public_ip_on_launch", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrSet(resourceName, "private_dns_hostname_type_on_launch"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVPCID),
				),
			},
		},
	})
}

// testAccCheckDefaultSubnetDestroyExists runs after all resources are destroyed.
// It verifies that the default subnet still exists.
// Any missing default subnets are then created.
func testAccCheckDefaultSubnetDestroyExists(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_subnet" {
				continue
			}

			_, err := tfec2.FindSubnetByID(ctx, conn, rs.Primary.ID)

			if err != nil {
				return err
			}
		}

		return testAccCreateMissingDefaultSubnets(ctx)
	}
}

// testAccCheckDefaultSubnetDestroyNotFound runs after all resources are destroyed.
// It verifies that the default subnet does not exist.
// Any missing default subnets are then created.
func testAccCheckDefaultSubnetDestroyNotFound(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_subnet" {
				continue
			}

			_, err := tfec2.FindSubnetByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Default Subnet %s still exists", rs.Primary.ID)
		}

		return testAccCreateMissingDefaultSubnets(ctx)
	}
}

func testAccCreateMissingDefaultSubnets(ctx context.Context) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	output, err := conn.DescribeAvailabilityZones(ctx, &ec2.DescribeAvailabilityZonesInput{
		Filters: tfec2.NewAttributeFilterList(
			map[string]string{
				"opt-in-status": "opt-in-not-required",
				names.AttrState: "available",
			},
		),
	})

	if err != nil {
		return err
	}

	for _, v := range output.AvailabilityZones {
		availabilityZone := aws.ToString(v.ZoneName)

		_, err := conn.CreateDefaultSubnet(ctx, &ec2.CreateDefaultSubnetInput{
			AvailabilityZone: aws.String(availabilityZone),
		})

		if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeDefaultSubnetAlreadyExistsInAvailabilityZone) {
			continue
		}

		if err != nil {
			return fmt.Errorf("creating new default subnet (%s): %w", availabilityZone, err)
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

func testAccVPCDefaultSubnetConfig_basic() string {
	return acctest.ConfigCompose(testAccDefaultSubnetConfigBaseExisting, `
resource "aws_default_subnet" "test" {
  availability_zone = data.aws_subnet.test.availability_zone
}
`)
}

func testAccVPCDefaultSubnetConfig_notFound() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
resource "aws_default_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
}
`)
}

func testAccVPCDefaultSubnetConfig_forceDestroy() string {
	return acctest.ConfigCompose(testAccDefaultSubnetConfigBaseExisting, `
resource "aws_default_subnet" "test" {
  availability_zone = data.aws_subnet.test.availability_zone
  force_destroy     = true
}
`)
}

func testAccVPCDefaultSubnetConfig_ipv6() string {
	return acctest.ConfigCompose(testAccDefaultSubnetConfigBaseExisting, `
resource "aws_default_vpc" "test" {
  assign_generated_ipv6_cidr_block = true
}

resource "aws_default_subnet" "test" {
  availability_zone = data.aws_subnet.test.availability_zone

  ipv6_cidr_block                 = cidrsubnet(aws_default_vpc.test.ipv6_cidr_block, 8, 1)
  assign_ipv6_address_on_creation = true
  enable_dns64                    = true

  private_dns_hostname_type_on_launch = "ip-name"

  # force_destroy so that the default VPC can have IPv6 disabled.
  force_destroy = true

  depends_on = [aws_default_vpc.test]
}
`)
}

func testAccVPCDefaultSubnetConfig_ipv6Native() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
resource "aws_default_vpc" "test" {
  assign_generated_ipv6_cidr_block = true
}

resource "aws_default_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]

  assign_ipv6_address_on_creation = true
  ipv6_native                     = true
  map_public_ip_on_launch         = false

  enable_resource_name_dns_aaaa_record_on_launch = true

  # force_destroy so that the default VPC can have IPv6 disabled.
  force_destroy = true

  depends_on = [aws_default_vpc.test]
}
`)
}

func testAccVPCDefaultSubnetConfig_privateDNSNameOptionsOnLaunch(rName string) string {
	return acctest.ConfigCompose(testAccDefaultSubnetConfigBaseExisting, fmt.Sprintf(`
resource "aws_default_subnet" "test" {
  availability_zone = data.aws_subnet.test.availability_zone

  map_public_ip_on_launch             = false
  private_dns_hostname_type_on_launch = "resource-name"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}
