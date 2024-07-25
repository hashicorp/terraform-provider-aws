// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCDefaultVPCAndSubnet_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"VPC": {
			"existing.basic":                                testAccDefaultVPC_Existing_basic,
			"existing.assignGeneratedIPv6CIDRBlock":         testAccDefaultVPC_Existing_assignGeneratedIPv6CIDRBlock,
			"existing.forceDestroy":                         testAccDefaultVPC_Existing_forceDestroy,
			"notFound.basic":                                testAccDefaultVPC_NotFound_basic,
			"notFound.assignGeneratedIPv6CIDRBlock":         testAccDefaultVPC_NotFound_assignGeneratedIPv6CIDRBlock,
			"notFound.forceDestroy":                         testAccDefaultVPC_NotFound_forceDestroy,
			"notFound.assignGeneratedIPv6CIDRBlockAdoption": testAccDefaultVPC_NotFound_assignGeneratedIPv6CIDRBlockAdoption,
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

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccPreCheckDefaultVPCExists(ctx context.Context, t *testing.T) {
	if !hasDefaultVPC(ctx, t) {
		t.Skip("skipping since no default VPC exists")
	}
}

func testAccPreCheckDefaultVPCNotFound(ctx context.Context, t *testing.T) {
	if vpcID := defaultVPC(ctx, t); vpcID != "" {
		t.Logf("Deleting existing default VPC: %s", vpcID)

		err := testAccEmptyDefaultVPC(ctx, vpcID)

		if err != nil {
			t.Fatalf("error emptying default VPC: %s", err)
		}

		r := tfec2.ResourceVPC()
		d := r.Data(nil)
		d.SetId(vpcID)

		err = acctest.DeleteResource(ctx, r, d, acctest.Provider.Meta())

		if err != nil {
			t.Fatalf("error deleting default VPC: %s", err)
		}
	}
}

func testAccDefaultVPC_Existing_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Vpc
	resourceName := "aws_default_vpc.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegionNot(t, names.USWest2RegionID, names.USGovWest1RegionID)
			testAccPreCheckDefaultVPCExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultVPCDestroyExists(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultVPCConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrCIDRBlock, "172.31.0.0/16"),
					resource.TestCheckResourceAttrSet(resourceName, "default_network_acl_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_route_table_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "dhcp_options_id"),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_hostnames", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_support", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_network_address_usage_metrics", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "existing_default_vpc", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "instance_tenancy", "default"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block_network_border_group", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_ipam_pool_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_netmask_length", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "main_route_table_id"),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func testAccDefaultVPC_Existing_assignGeneratedIPv6CIDRBlock(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Vpc
	resourceName := "aws_default_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegionNot(t, names.USWest2RegionID, names.USGovWest1RegionID)
			testAccPreCheckDefaultVPCExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultVPCDestroyExists(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultVPCConfig_assignGeneratedIPv6CIDRBlock(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrCIDRBlock, "172.31.0.0/16"),
					resource.TestCheckResourceAttrSet(resourceName, "default_network_acl_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_route_table_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "dhcp_options_id"),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_hostnames", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_support", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_network_address_usage_metrics", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "existing_default_vpc", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "instance_tenancy", "default"),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_association_id"),
					resource.TestMatchResourceAttr(resourceName, "ipv6_cidr_block", regexache.MustCompile(`/56$`)),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block_network_border_group", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "ipv6_ipam_pool_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_netmask_length", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "main_route_table_id"),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func testAccDefaultVPC_Existing_forceDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Vpc
	resourceName := "aws_default_vpc.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegionNot(t, names.USWest2RegionID, names.USGovWest1RegionID)
			testAccPreCheckDefaultVPCExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultVPCDestroyNotFound(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultVPCConfig_forceDestroy,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "existing_default_vpc", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtTrue),
					testAccCheckDefaultVPCEmpty(ctx, &v),
				),
			},
		},
	})
}

func testAccDefaultVPC_NotFound_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Vpc
	resourceName := "aws_default_vpc.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegionNot(t, names.USWest2RegionID, names.USGovWest1RegionID)
			testAccPreCheckDefaultVPCNotFound(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultVPCDestroyExists(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultVPCConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrCIDRBlock, "172.31.0.0/16"),
					resource.TestCheckResourceAttrSet(resourceName, "default_network_acl_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_route_table_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "dhcp_options_id"),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_hostnames", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_support", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_network_address_usage_metrics", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "existing_default_vpc", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "instance_tenancy", "default"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block_network_border_group", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_ipam_pool_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_netmask_length", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "main_route_table_id"),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func testAccDefaultVPC_NotFound_assignGeneratedIPv6CIDRBlock(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Vpc
	resourceName := "aws_default_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegionNot(t, names.USWest2RegionID, names.USGovWest1RegionID)
			testAccPreCheckDefaultVPCNotFound(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultVPCDestroyExists(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultVPCConfig_assignGeneratedIPv6CIDRBlock(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrCIDRBlock, "172.31.0.0/16"),
					resource.TestCheckResourceAttrSet(resourceName, "default_network_acl_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_route_table_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "dhcp_options_id"),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_hostnames", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_support", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_network_address_usage_metrics", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "existing_default_vpc", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "instance_tenancy", "default"),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_association_id"),
					resource.TestMatchResourceAttr(resourceName, "ipv6_cidr_block", regexache.MustCompile(`/56$`)),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block_network_border_group", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "ipv6_ipam_pool_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_netmask_length", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "main_route_table_id"),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func testAccDefaultVPC_NotFound_forceDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Vpc
	resourceName := "aws_default_vpc.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegionNot(t, names.USWest2RegionID, names.USGovWest1RegionID)
			testAccPreCheckDefaultVPCNotFound(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultVPCDestroyNotFound(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultVPCConfig_forceDestroy,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "existing_default_vpc", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtTrue),
					testAccCheckDefaultVPCEmpty(ctx, &v),
				),
			},
		},
	})
}

func testAccDefaultVPC_NotFound_assignGeneratedIPv6CIDRBlockAdoption(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Vpc
	resourceName := "aws_default_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegionNot(t, names.USWest2RegionID, names.USGovWest1RegionID)
			testAccPreCheckDefaultVPCNotFound(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultVPCDestroyExists(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultVPCConfig_assignGeneratedIPv6CIDRBlockAdoptionStep1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrCIDRBlock, "172.31.0.0/16"),
					resource.TestCheckResourceAttrSet(resourceName, "default_network_acl_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_route_table_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "dhcp_options_id"),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_hostnames", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_support", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_network_address_usage_metrics", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "existing_default_vpc", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "instance_tenancy", "default"),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_association_id"),
					resource.TestMatchResourceAttr(resourceName, "ipv6_cidr_block", regexache.MustCompile(`/56$`)),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block_network_border_group", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "ipv6_ipam_pool_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_netmask_length", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "main_route_table_id"),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccVPCDefaultVPCConfig_assignGeneratedIPv6CIDRBlockAdoptionStep2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify that the default VPC is no longer in state.
					func(s *terraform.State) error {
						_, ok := s.RootModule().Resources[resourceName]
						if ok {
							return fmt.Errorf("Found: %s", resourceName)
						}

						return nil
					},
				),
			},
			{
				Config: testAccVPCDefaultVPCConfig_assignGeneratedIPv6CIDRBlockAdoptionStep3(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrCIDRBlock, "172.31.0.0/16"),
					resource.TestCheckResourceAttrSet(resourceName, "default_network_acl_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_route_table_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "dhcp_options_id"),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_hostnames", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_support", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_network_address_usage_metrics", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "existing_default_vpc", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "instance_tenancy", "default"),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_association_id"),
					resource.TestMatchResourceAttr(resourceName, "ipv6_cidr_block", regexache.MustCompile(`/56$`)),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block_network_border_group", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "ipv6_ipam_pool_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_netmask_length", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "main_route_table_id"),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

// testAccCheckDefaultVPCDestroyExists runs after all resources are destroyed.
// It verifies that the default VPC still exists.
func testAccCheckDefaultVPCDestroyExists(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_default_vpc" {
				continue
			}

			_, err := tfec2.FindVPCByID(ctx, conn, rs.Primary.ID)

			if err != nil {
				return err
			}
		}

		return nil
	}
}

// testAccCheckDefaultVPCDestroyNotFound runs after all resources are destroyed.
// It verifies that the default VPC does not exist.
// A new default VPC is then created.
func testAccCheckDefaultVPCDestroyNotFound(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_default_vpc" {
				continue
			}

			_, err := tfec2.FindVPCByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Default VPC %s still exists", rs.Primary.ID)
		}

		_, err := conn.CreateDefaultVpc(ctx, &ec2.CreateDefaultVpcInput{})

		if err != nil {
			return fmt.Errorf("error creating new default VPC: %w", err)
		}

		return nil
	}
}

// testAccCheckDefaultVPCEmpty returns a TestCheckFunc that empties the specified default VPC.
func testAccCheckDefaultVPCEmpty(ctx context.Context, v *awstypes.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccEmptyDefaultVPC(ctx, aws.ToString(v.VpcId))
	}
}

// testAccEmptyDefaultVPC empties a default VPC so that it can be deleted.
func testAccEmptyDefaultVPC(ctx context.Context, vpcID string) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	// Delete the default IGW.
	igw, err := tfec2.FindInternetGateway(ctx, conn, &ec2.DescribeInternetGatewaysInput{
		Filters: tfec2.NewAttributeFilterList(
			map[string]string{
				"attachment.state":  "available",
				"attachment.vpc-id": vpcID,
			},
		),
	})

	if err == nil {
		r := tfec2.ResourceInternetGateway()
		d := r.Data(nil)
		d.SetId(aws.ToString(igw.InternetGatewayId))
		d.Set(names.AttrVPCID, vpcID)

		err := acctest.DeleteResource(ctx, r, d, acctest.Provider.Meta())

		if err != nil {
			return err
		}
	} else if !tfresource.NotFound(err) {
		return err
	}

	// Delete default subnets.
	subnets, err := tfec2.FindSubnets(ctx, conn, &ec2.DescribeSubnetsInput{
		Filters: tfec2.NewAttributeFilterList(
			map[string]string{
				"defaultForAz": acctest.CtTrue,
			},
		),
	})

	if err != nil {
		return err
	}

	for _, v := range subnets {
		r := tfec2.ResourceSubnet()
		d := r.Data(nil)
		d.SetId(aws.ToString(v.SubnetId))

		err := acctest.DeleteResource(ctx, r, d, acctest.Provider.Meta())

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

func testAccVPCDefaultVPCConfig_assignGeneratedIPv6CIDRBlockAdoptionStep1(rName string) string {
	// Create an IPV6 subnet in the default VPC.
	return fmt.Sprintf(`
resource "aws_default_vpc" "test" {
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

data "aws_vpc" "test" {
  default = true

  depends_on = [aws_default_vpc.test]
}

resource "aws_subnet" "test" {
  cidr_block                      = "172.31.96.0/20"
  vpc_id                          = data.aws_vpc.test.id
  ipv6_cidr_block                 = cidrsubnet(data.aws_vpc.test.ipv6_cidr_block, 8, 1)
  assign_ipv6_address_on_creation = true

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCDefaultVPCConfig_assignGeneratedIPv6CIDRBlockAdoptionStep2(rName string) string {
	// Remove the default VPC from state.
	return fmt.Sprintf(`
data "aws_vpc" "test" {
  default = true
}

resource "aws_subnet" "test" {
  cidr_block                      = "172.31.96.0/20"
  vpc_id                          = data.aws_vpc.test.id
  ipv6_cidr_block                 = cidrsubnet(data.aws_vpc.test.ipv6_cidr_block, 8, 1)
  assign_ipv6_address_on_creation = true

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCDefaultVPCConfig_assignGeneratedIPv6CIDRBlockAdoptionStep3(rName string) string {
	// Adopt the default VPC.
	return fmt.Sprintf(`
resource "aws_default_vpc" "test" {
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

data "aws_vpc" "test" {
  default = true
}

resource "aws_subnet" "test" {
  cidr_block                      = "172.31.96.0/20"
  vpc_id                          = data.aws_vpc.test.id
  ipv6_cidr_block                 = cidrsubnet(data.aws_vpc.test.ipv6_cidr_block, 8, 1)
  assign_ipv6_address_on_creation = true

  tags = {
    Name = %[1]q
  }
}
`, rName)
}
