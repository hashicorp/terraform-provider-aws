package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccPreCheckDefaultSubnetExists(t *testing.T) {
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

func testAccPreCheckDefaultSubnetNotFound(t *testing.T) {
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

	for _, v := range subnets {
		subnetID := aws.StringValue(v.SubnetId)

		t.Logf("Deleting existing default subnet: %s", subnetID)

		r := tfec2.ResourceSubnet()
		d := r.Data(nil)
		d.SetId(subnetID)

		err := acctest.DeleteResource(r, d, acctest.Provider.Meta())

		if err != nil {
			t.Fatalf("error deleting default subnet: %s", err)
		}
	}
}

func testAccEC2DefaultSubnet_Existing_basic(t *testing.T) {
	var v ec2.Subnet
	resourceName := "aws_default_subnet.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckRegionNot(t, endpoints.UsWest2RegionID, endpoints.UsGovWest1RegionID)
			testAccPreCheckDefaultSubnetExists(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDefaultSubnetDestroyExists,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultSubnetConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "assign_ipv6_address_on_creation", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrSet(resourceName, "cidr_block"),
					resource.TestCheckResourceAttr(resourceName, "customer_owned_ipv4_pool", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_dns64", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_aaaa_record_on_launch", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_a_record_on_launch", "false"),
					resource.TestCheckResourceAttr(resourceName, "existing_default_subnet", "true"),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "false"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_native", "false"),
					resource.TestCheckResourceAttr(resourceName, "map_customer_owned_ip_on_launch", "false"),
					resource.TestCheckResourceAttr(resourceName, "map_public_ip_on_launch", "true"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttrSet(resourceName, "private_dns_hostname_type_on_launch"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
				),
			},
		},
	})
}

func testAccEC2DefaultSubnet_Existing_forceDestroy(t *testing.T) {
	var v ec2.Subnet
	resourceName := "aws_default_subnet.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckRegionNot(t, endpoints.UsWest2RegionID, endpoints.UsGovWest1RegionID)
			testAccPreCheckDefaultSubnetExists(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDefaultSubnetDestroyNotFound,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultSubnetForceDestroyConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "existing_default_subnet", "true"),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "true"),
				),
			},
		},
	})
}

func testAccEC2DefaultSubnet_Existing_ipv6(t *testing.T) {
	var v ec2.Subnet
	resourceName := "aws_default_subnet.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckRegionNot(t, endpoints.UsWest2RegionID, endpoints.UsGovWest1RegionID)
			testAccPreCheckDefaultSubnetExists(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDefaultSubnetDestroyNotFound,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultSubnetIPv6Config(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "assign_ipv6_address_on_creation", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrSet(resourceName, "cidr_block"),
					resource.TestCheckResourceAttr(resourceName, "customer_owned_ipv4_pool", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_dns64", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_aaaa_record_on_launch", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_a_record_on_launch", "false"),
					resource.TestCheckResourceAttr(resourceName, "existing_default_subnet", "true"),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_cidr_block"),
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

func testAccEC2DefaultSubnet_Existing_privateDnsNameOptionsOnLaunch(t *testing.T) {
	var v ec2.Subnet
	resourceName := "aws_default_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckRegionNot(t, endpoints.UsWest2RegionID, endpoints.UsGovWest1RegionID)
			testAccPreCheckDefaultSubnetExists(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDefaultSubnetDestroyExists,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultSubnetPrivateDnsNameOptionsOnLaunchConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "assign_ipv6_address_on_creation", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrSet(resourceName, "cidr_block"),
					resource.TestCheckResourceAttr(resourceName, "customer_owned_ipv4_pool", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_dns64", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_aaaa_record_on_launch", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_a_record_on_launch", "false"),
					resource.TestCheckResourceAttr(resourceName, "existing_default_subnet", "true"),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "false"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_native", "false"),
					resource.TestCheckResourceAttr(resourceName, "map_customer_owned_ip_on_launch", "false"),
					resource.TestCheckResourceAttr(resourceName, "map_public_ip_on_launch", "false"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_hostname_type_on_launch", "resource-name"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
				),
			},
		},
	})
}

func testAccEC2DefaultSubnet_NotFound_basic(t *testing.T) {
	var v ec2.Subnet
	resourceName := "aws_default_subnet.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckRegionNot(t, endpoints.UsWest2RegionID, endpoints.UsGovWest1RegionID)
			testAccPreCheckDefaultSubnetNotFound(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDefaultSubnetDestroyExists,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultSubnetNotFoundConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "assign_ipv6_address_on_creation", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrSet(resourceName, "cidr_block"),
					resource.TestCheckResourceAttr(resourceName, "customer_owned_ipv4_pool", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_dns64", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_aaaa_record_on_launch", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_a_record_on_launch", "false"),
					resource.TestCheckResourceAttr(resourceName, "existing_default_subnet", "false"),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "false"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_native", "false"),
					resource.TestCheckResourceAttr(resourceName, "map_customer_owned_ip_on_launch", "false"),
					resource.TestCheckResourceAttr(resourceName, "map_public_ip_on_launch", "true"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttrSet(resourceName, "private_dns_hostname_type_on_launch"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
				),
			},
		},
	})
}

func testAccEC2DefaultSubnet_NotFound_ipv6Native(t *testing.T) {
	var v ec2.Subnet
	resourceName := "aws_default_subnet.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckRegionNot(t, endpoints.UsWest2RegionID, endpoints.UsGovWest1RegionID)
			testAccPreCheckDefaultSubnetNotFound(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDefaultSubnetDestroyNotFound,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultSubnetIPv6NativeConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "assign_ipv6_address_on_creation", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "customer_owned_ipv4_pool", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_dns64", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_aaaa_record_on_launch", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_a_record_on_launch", "false"),
					resource.TestCheckResourceAttr(resourceName, "existing_default_subnet", "false"),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_cidr_block"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_native", "true"),
					resource.TestCheckResourceAttr(resourceName, "map_customer_owned_ip_on_launch", "false"),
					resource.TestCheckResourceAttr(resourceName, "map_public_ip_on_launch", "false"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttrSet(resourceName, "private_dns_hostname_type_on_launch"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
				),
			},
		},
	})
}

// testAccCheckDefaultSubnetDestroyExists runs after all resources are destroyed.
// It verifies that the default subnet still exists.
// Any missing default subnets are then created.
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

	err := testAccCreateMissingDefaultSubnets()

	if err != nil {
		return err
	}

	return nil
}

// testAccCheckDefaultSubnetDestroyNotFound runs after all resources are destroyed.
// It verifies that the default subnet does not exist.
// Any missing default subnets are then created.
func testAccCheckDefaultSubnetDestroyNotFound(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_subnet" {
			continue
		}

		_, err := tfec2.FindSubnetByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 Default Subnet %s still exists", rs.Primary.ID)
	}

	err := testAccCreateMissingDefaultSubnets()

	if err != nil {
		return err
	}

	return nil
}

func testAccCreateMissingDefaultSubnets() error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	output, err := conn.DescribeAvailabilityZones(&ec2.DescribeAvailabilityZonesInput{
		Filters: tfec2.BuildAttributeFilterList(
			map[string]string{
				"opt-in-status": "opt-in-not-required",
				"state":         "available",
			},
		),
	})

	if err != nil {
		return err
	}

	for _, v := range output.AvailabilityZones {
		availabilityZone := aws.StringValue(v.ZoneName)

		_, err := conn.CreateDefaultSubnet(&ec2.CreateDefaultSubnetInput{
			AvailabilityZone: aws.String(availabilityZone),
		})

		if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeDefaultSubnetAlreadyExistsInAvailabilityZone) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error creating new default subnet (%s): %w", availabilityZone, err)
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

func testAccDefaultSubnetNotFoundConfig() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
resource "aws_default_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
}
`)
}

func testAccDefaultSubnetForceDestroyConfig() string {
	return acctest.ConfigCompose(testAccDefaultSubnetConfigBaseExisting, `
resource "aws_default_subnet" "test" {
  availability_zone = data.aws_subnet.test.availability_zone
  force_destroy     = true
}
`)
}

func testAccDefaultSubnetIPv6Config() string {
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

func testAccDefaultSubnetIPv6NativeConfig() string {
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

func testAccDefaultSubnetPrivateDnsNameOptionsOnLaunchConfig(rName string) string {
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
