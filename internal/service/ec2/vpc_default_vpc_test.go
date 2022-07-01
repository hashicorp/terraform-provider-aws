package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccVPCDefaultVPCAndSubnet_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"VPC": {
			"existing.basic":                        testAccDefaultVPC_Existing_basic,
			"existing.assignGeneratedIPv6CIDRBlock": testAccDefaultVPC_Existing_assignGeneratedIPv6CIDRBlock,
			"existing.forceDestroy":                 testAccDefaultVPC_Existing_forceDestroy,
			"notFound.basic":                        testAccDefaultVPC_NotFound_basic,
			"notFound.assignGeneratedIPv6CIDRBlock": testAccDefaultVPC_NotFound_assignGeneratedIPv6CIDRBlock,
			"notFound.forceDestroy":                 testAccDefaultVPC_NotFound_forceDestroy,
		},
		"Subnet": {
			"existing.basic":                         testAccDefaultSubnet_Existing_basic,
			"existing.forceDestroy":                  testAccDefaultSubnet_Existing_forceDestroy,
			"existing.ipv6":                          testAccDefaultSubnet_Existing_ipv6,
			"existing.privateDnsNameOptionsOnLaunch": testAccDefaultSubnet_Existing_privateDNSNameOptionsOnLaunch,
			"notFound.basic":                         testAccDefaultSubnet_NotFound_basic,
			"notFound.ipv6Native":                    testAccDefaultSubnet_NotFound_ipv6Native,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}

func testAccPreCheckDefaultVPCExists(t *testing.T) {
	if !hasDefaultVPC(t) {
		t.Skip("skipping since no default VPC exists")
	}
}

func testAccPreCheckDefaultVPCNotFound(t *testing.T) {
	if vpcID := defaultVPC(t); vpcID != "" {
		t.Logf("Deleting existing default VPC: %s", vpcID)

		err := testAccEmptyDefaultVPC(vpcID)

		if err != nil {
			t.Fatalf("error emptying default VPC: %s", err)
		}

		r := tfec2.ResourceVPC()
		d := r.Data(nil)
		d.SetId(vpcID)

		err = acctest.DeleteResource(r, d, acctest.Provider.Meta())

		if err != nil {
			t.Fatalf("error deleting default VPC: %s", err)
		}
	}
}

func testAccDefaultVPC_Existing_basic(t *testing.T) {
	var v ec2.Vpc
	resourceName := "aws_default_vpc.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckRegionNot(t, endpoints.UsWest2RegionID, endpoints.UsGovWest1RegionID)
			testAccPreCheckDefaultVPCExists(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDefaultVPCDestroyExists,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultVPCConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", "false"),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "172.31.0.0/16"),
					resource.TestCheckResourceAttrSet(resourceName, "default_network_acl_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_route_table_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "dhcp_options_id"),
					resource.TestCheckResourceAttr(resourceName, "enable_classiclink", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_classiclink_dns_support", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_hostnames", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_support", "true"),
					resource.TestCheckResourceAttr(resourceName, "existing_default_vpc", "true"),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "false"),
					resource.TestCheckResourceAttr(resourceName, "instance_tenancy", "default"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block_network_border_group", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_ipam_pool_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_netmask_length", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "main_route_table_id"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccDefaultVPC_Existing_assignGeneratedIPv6CIDRBlock(t *testing.T) {
	var v ec2.Vpc
	resourceName := "aws_default_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckRegionNot(t, endpoints.UsWest2RegionID, endpoints.UsGovWest1RegionID)
			testAccPreCheckDefaultVPCExists(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDefaultVPCDestroyExists,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultVPCConfig_assignGeneratedIPv6CIDRBlock(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", "true"),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "172.31.0.0/16"),
					resource.TestCheckResourceAttrSet(resourceName, "default_network_acl_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_route_table_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "dhcp_options_id"),
					resource.TestCheckResourceAttr(resourceName, "enable_classiclink", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_classiclink_dns_support", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_hostnames", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_support", "true"),
					resource.TestCheckResourceAttr(resourceName, "existing_default_vpc", "true"),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "false"),
					resource.TestCheckResourceAttr(resourceName, "instance_tenancy", "default"),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_association_id"),
					resource.TestMatchResourceAttr(resourceName, "ipv6_cidr_block", regexp.MustCompile(`/56$`)),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block_network_border_group", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "ipv6_ipam_pool_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_netmask_length", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "main_route_table_id"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func testAccDefaultVPC_Existing_forceDestroy(t *testing.T) {
	var v ec2.Vpc
	resourceName := "aws_default_vpc.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckRegionNot(t, endpoints.UsWest2RegionID, endpoints.UsGovWest1RegionID)
			testAccPreCheckDefaultVPCExists(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDefaultVPCDestroyNotFound,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultVPCConfig_forceDestroy,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "existing_default_vpc", "true"),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "true"),
					testAccCheckDefaultVPCEmpty(&v),
				),
			},
		},
	})
}

func testAccDefaultVPC_NotFound_basic(t *testing.T) {
	var v ec2.Vpc
	resourceName := "aws_default_vpc.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckRegionNot(t, endpoints.UsWest2RegionID, endpoints.UsGovWest1RegionID)
			testAccPreCheckDefaultVPCNotFound(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDefaultVPCDestroyExists,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultVPCConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", "false"),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "172.31.0.0/16"),
					resource.TestCheckResourceAttrSet(resourceName, "default_network_acl_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_route_table_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "dhcp_options_id"),
					resource.TestCheckResourceAttr(resourceName, "enable_classiclink", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_classiclink_dns_support", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_hostnames", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_support", "true"),
					resource.TestCheckResourceAttr(resourceName, "existing_default_vpc", "false"),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "false"),
					resource.TestCheckResourceAttr(resourceName, "instance_tenancy", "default"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block_network_border_group", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_ipam_pool_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_netmask_length", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "main_route_table_id"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccDefaultVPC_NotFound_assignGeneratedIPv6CIDRBlock(t *testing.T) {
	var v ec2.Vpc
	resourceName := "aws_default_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckRegionNot(t, endpoints.UsWest2RegionID, endpoints.UsGovWest1RegionID)
			testAccPreCheckDefaultVPCNotFound(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDefaultVPCDestroyExists,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultVPCConfig_assignGeneratedIPv6CIDRBlock(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", "true"),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "172.31.0.0/16"),
					resource.TestCheckResourceAttrSet(resourceName, "default_network_acl_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_route_table_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "dhcp_options_id"),
					resource.TestCheckResourceAttr(resourceName, "enable_classiclink", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_classiclink_dns_support", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_hostnames", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_support", "true"),
					resource.TestCheckResourceAttr(resourceName, "existing_default_vpc", "false"),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "false"),
					resource.TestCheckResourceAttr(resourceName, "instance_tenancy", "default"),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_association_id"),
					resource.TestMatchResourceAttr(resourceName, "ipv6_cidr_block", regexp.MustCompile(`/56$`)),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block_network_border_group", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "ipv6_ipam_pool_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_netmask_length", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "main_route_table_id"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func testAccDefaultVPC_NotFound_forceDestroy(t *testing.T) {
	var v ec2.Vpc
	resourceName := "aws_default_vpc.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckRegionNot(t, endpoints.UsWest2RegionID, endpoints.UsGovWest1RegionID)
			testAccPreCheckDefaultVPCNotFound(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDefaultVPCDestroyNotFound,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultVPCConfig_forceDestroy,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "existing_default_vpc", "false"),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "true"),
					testAccCheckDefaultVPCEmpty(&v),
				),
			},
		},
	})
}

// testAccCheckDefaultVPCDestroyExists runs after all resources are destroyed.
// It verifies that the default VPC still exists.
func testAccCheckDefaultVPCDestroyExists(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_default_vpc" {
			continue
		}

		_, err := tfec2.FindVPCByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}
	}

	return nil
}

// testAccCheckDefaultVPCDestroyNotFound runs after all resources are destroyed.
// It verifies that the default VPC does not exist.
// A new default VPC is then created.
func testAccCheckDefaultVPCDestroyNotFound(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_default_vpc" {
			continue
		}

		_, err := tfec2.FindVPCByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 Default VPC %s still exists", rs.Primary.ID)
	}

	_, err := conn.CreateDefaultVpc(&ec2.CreateDefaultVpcInput{})

	if err != nil {
		return fmt.Errorf("error creating new default VPC: %w", err)
	}

	return nil
}

// testAccCheckDefaultVPCEmpty returns a TestCheckFunc that empties the specified default VPC.
func testAccCheckDefaultVPCEmpty(v *ec2.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccEmptyDefaultVPC(aws.StringValue(v.VpcId))
	}
}

// testAccEmptyDefaultVPC empties a default VPC so that it can be deleted.
func testAccEmptyDefaultVPC(vpcID string) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	// Delete the default IGW.
	igw, err := tfec2.FindInternetGateway(conn, &ec2.DescribeInternetGatewaysInput{
		Filters: tfec2.BuildAttributeFilterList(
			map[string]string{
				"attachment.state":  "available",
				"attachment.vpc-id": vpcID,
			},
		),
	})

	if err == nil {
		r := tfec2.ResourceInternetGateway()
		d := r.Data(nil)
		d.SetId(aws.StringValue(igw.InternetGatewayId))
		d.Set("vpc_id", vpcID)

		err := acctest.DeleteResource(r, d, acctest.Provider.Meta())

		if err != nil {
			return err
		}
	} else if !tfresource.NotFound(err) {
		return err
	}

	// Delete default subnets.
	subnets, err := tfec2.FindSubnets(conn, &ec2.DescribeSubnetsInput{
		Filters: tfec2.BuildAttributeFilterList(
			map[string]string{
				"defaultForAz": "true",
			},
		),
	})

	if err != nil {
		return err
	}

	for _, v := range subnets {
		r := tfec2.ResourceSubnet()
		d := r.Data(nil)
		d.SetId(aws.StringValue(v.SubnetId))

		err := acctest.DeleteResource(r, d, acctest.Provider.Meta())

		if err != nil {
			return err
		}
	}

	return nil
}

const testAccVPCDefaultVPCConfig_basic = `
resource "aws_default_vpc" "test" {}
`

const testAccVPCDefaultVPCConfig_forceDestroy = `
resource "aws_default_vpc" "test" {
  force_destroy = true
}
`

func testAccVPCDefaultVPCConfig_assignGeneratedIPv6CIDRBlock(rName string) string {
	return fmt.Sprintf(`
resource "aws_default_vpc" "test" {
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}
`, rName)
}
