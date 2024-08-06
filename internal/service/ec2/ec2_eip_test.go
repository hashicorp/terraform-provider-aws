// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2EIP_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.Address
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomain, "vpc"),
					resource.TestCheckResourceAttr(resourceName, "ptr_record", ""),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
					testAccCheckEIPPublicDNS(ctx, resourceName),
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

func TestAccEC2EIP_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.Address
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceEIP(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2EIP_migrateVPCToDomain(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.Address
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy: testAccCheckEIPDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "4.67.0",
					},
				},
				Config: testAccEIPConfig_vpc,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomain, "vpc"),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
					testAccCheckEIPPublicDNS(ctx, resourceName),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccEIPConfig_basic,
				PlanOnly:                 true,
			},
		},
	})
}

func TestAccEC2EIP_noVPC(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.Address
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_noVPC,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomain, "vpc"),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
					testAccCheckEIPPublicDNS(ctx, resourceName),
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

func TestAccEC2EIP_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.Address
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEIPConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccEIPConfig_tags1(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccEC2EIP_instance(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.Address
	instanceResourceName := "aws_instance.test"
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_instance(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "instance", instanceResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
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

// Regression test for https://github.com/hashicorp/terraform/issues/3429 (now
// https://github.com/hashicorp/terraform-provider-aws/issues/42)
func TestAccEC2EIP_Instance_reassociate(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.Address
	instanceResourceName := "aws_instance.test"
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_instanceReassociate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "instance", instanceResourceName, names.AttrID),
				),
			},
			{
				Config: testAccEIPConfig_instanceReassociate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "instance", instanceResourceName, names.AttrID),
				),
				Taint: []string{resourceName},
			},
		},
	})
}

// This test is an expansion of TestAccEC2EIP_Instance_associatedUserPrivateIP, by testing the
// associated Private EIPs of two instances
func TestAccEC2EIP_Instance_associatedUserPrivateIP(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.Address
	instance1ResourceName := "aws_instance.test.1"
	instance2ResourceName := "aws_instance.test.0"
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_instanceAssociated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAssociationID),
					resource.TestCheckResourceAttrPair(resourceName, "instance", instance1ResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"associate_with_private_ip"},
			},
			{
				Config: testAccEIPConfig_instanceAssociatedSwitch(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAssociationID),
					resource.TestCheckResourceAttrPair(resourceName, "instance", instance2ResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
				),
			},
		},
	})
}

func TestAccEC2EIP_Instance_notAssociated(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.Address
	instanceResourceName := "aws_instance.test"
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_instanceAssociateNotAssociated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrAssociationID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance", ""),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEIPConfig_instance(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAssociationID),
					resource.TestCheckResourceAttrPair(resourceName, "instance", instanceResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
				),
			},
		},
	})
}

func TestAccEC2EIP_networkInterface(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.Address
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_networkInterface(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(ctx, resourceName, &conf),
					testAccCheckEIPPrivateDNS(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "allocation_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAssociationID),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
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

func TestAccEC2EIP_NetworkInterface_twoEIPsOneInterface(t *testing.T) {
	ctx := acctest.Context(t)
	var one, two types.Address
	resource1Name := "aws_eip.test.0"
	resource2Name := "aws_eip.test.1"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_multiNetworkInterface(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(ctx, resource1Name, &one),
					resource.TestCheckResourceAttrSet(resource1Name, names.AttrAssociationID),
					resource.TestCheckResourceAttrSet(resource1Name, "public_ip"),

					testAccCheckEIPExists(ctx, resource2Name, &two),
					resource.TestCheckResourceAttrSet(resource2Name, names.AttrAssociationID),
					resource.TestCheckResourceAttrSet(resource2Name, "public_ip"),
				),
			},
		},
	})
}

func TestAccEC2EIP_association(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.Address
	instanceResourceName := "aws_instance.test"
	eniResourceName := "aws_network_interface.test"
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_associationNone(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrAssociationID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface", ""),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
				),
			},
			{
				Config: testAccEIPConfig_associationENI(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAssociationID),
					resource.TestCheckResourceAttr(resourceName, "instance", ""),
					resource.TestCheckResourceAttrPair(resourceName, "network_interface", eniResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
				),
			},
			{
				Config: testAccEIPConfig_associationInstance(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAssociationID),
					resource.TestCheckResourceAttrPair(resourceName, "instance", instanceResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "network_interface", instanceResourceName, "primary_network_interface_id"),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
				),
			},
		},
	})
}

func TestAccEC2EIP_PublicIPv4Pool_default(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.Address
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_publicIPv4PoolDefault(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
					resource.TestCheckResourceAttr(resourceName, "public_ipv4_pool", "amazon"),
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

func TestAccEC2EIP_PublicIPv4Pool_custom(t *testing.T) {
	ctx := acctest.Context(t)
	key := "AWS_EC2_EIP_PUBLIC_IPV4_POOL"
	poolName := os.Getenv(key)
	if poolName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var conf types.Address
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_publicIPv4PoolCustom(rName, poolName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
					resource.TestCheckResourceAttr(resourceName, "public_ipv4_pool", poolName),
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

func TestAccEC2EIP_customerOwnedIPv4Pool(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.Address
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_customerOwnedIPv4Pool(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(ctx, resourceName, &conf),
					resource.TestMatchResourceAttr(resourceName, "customer_owned_ipv4_pool", regexache.MustCompile(`^ipv4pool-coip-.+$`)),
					resource.TestMatchResourceAttr(resourceName, "customer_owned_ip", regexache.MustCompile(`\d+\.\d+\.\d+\.\d+`)),
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

func TestAccEC2EIP_networkBorderGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.Address
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_networkBorderGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "network_border_group", acctest.Region()),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
					resource.TestCheckResourceAttr(resourceName, "public_ipv4_pool", "amazon"),
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

func TestAccEC2EIP_carrierIP(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.Address
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckWavelengthZoneAvailable(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_carrierIP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "carrier_ip"),
					resource.TestCheckResourceAttrSet(resourceName, "network_border_group"),
					resource.TestCheckResourceAttr(resourceName, "public_ip", ""),
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

func TestAccEC2EIP_BYOIPAddress_default(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.Address
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_byoipAddressCustomDefault(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
				),
			},
		},
	})
}

func TestAccEC2EIP_BYOIPAddress_custom(t *testing.T) {
	ctx := acctest.Context(t)
	key := "AWS_EC2_EIP_BYOIP_ADDRESS"
	address := os.Getenv(key)
	if address == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var conf types.Address
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_byoipAddressCustom(rName, address),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "public_ip", address),
				),
			},
		},
	})
}

func TestAccEC2EIP_BYOIPAddress_customWithPublicIPv4Pool(t *testing.T) {
	ctx := acctest.Context(t)
	key := "AWS_EC2_EIP_BYOIP_ADDRESS"
	address := os.Getenv(key)
	if address == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	key = "AWS_EC2_EIP_PUBLIC_IPV4_POOL"
	poolName := os.Getenv(key)
	if poolName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var conf types.Address
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_byoipAddressCustomPublicIPv4Pool(rName, address, poolName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "public_ip", address),
					resource.TestCheckResourceAttr(resourceName, "public_ipv4_pool", poolName),
				),
			},
		},
	})
}

func testAccCheckEIPExists(ctx context.Context, n string, v *types.Address) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindEIPByAllocationID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckEIPDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_eip" {
				continue
			}

			_, err := tfec2.FindEIPByAllocationID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 EIP %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckEIPPrivateDNS(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		privateDNS := rs.Primary.Attributes["private_dns"]
		expectedPrivateDNS := acctest.Provider.Meta().(*conns.AWSClient).EC2PrivateDNSNameForIP(ctx, rs.Primary.Attributes["private_ip"])

		if privateDNS != expectedPrivateDNS {
			return fmt.Errorf("expected private_dns value (%s), received: %s", expectedPrivateDNS, privateDNS)
		}

		return nil
	}
}

func testAccCheckEIPPublicDNS(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		publicDNS := rs.Primary.Attributes["public_dns"]
		expectedPublicDNS := acctest.Provider.Meta().(*conns.AWSClient).EC2PublicDNSNameForIP(ctx, rs.Primary.Attributes["public_ip"])

		if publicDNS != expectedPublicDNS {
			return fmt.Errorf("expected public_dns value (%s), received: %s", expectedPublicDNS, publicDNS)
		}

		return nil
	}
}

const testAccEIPConfig_basic = `
resource "aws_eip" "test" {
  domain = "vpc"
}
`

const testAccEIPConfig_vpc = `
resource "aws_eip" "test" {
  vpc = true
}
`

const testAccEIPConfig_noVPC = `
resource "aws_eip" "test" {
}
`

func testAccEIPConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  vpc = true

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccEIPConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  domain = "vpc"

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccEIPConfig_baseInstance(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigVPCWithSubnets(rName, 1),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEIPConfig_instance(rName string) string {
	return acctest.ConfigCompose(testAccEIPConfig_baseInstance(rName), fmt.Sprintf(`
resource "aws_eip" "test" {
  instance = aws_instance.test.id
  domain   = "vpc"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEIPConfig_instanceReassociate(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigVPCWithSubnets(rName, 1),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_eip" "test" {
  instance = aws_instance.test.id
  domain   = "vpc"

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  associate_public_ip_address = true
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id                   = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  subnet_id      = aws_subnet.test[0].id
  route_table_id = aws_route_table.test.id
}
`, rName))
}

func testAccEIPConfig_baseInstanceAssociated(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id                  = aws_vpc.test.id
  availability_zone       = data.aws_availability_zones.available.names[0]
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = true

  depends_on = [aws_internet_gateway.test]

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  count = 2

  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  private_ip = "10.0.0.1${count.index}"
  subnet_id  = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEIPConfig_instanceAssociated(rName string) string {
	return acctest.ConfigCompose(testAccEIPConfig_baseInstanceAssociated(rName), fmt.Sprintf(`
resource "aws_eip" "test" {
  domain = "vpc"

  instance                  = aws_instance.test[1].id
  associate_with_private_ip = aws_instance.test[1].private_ip

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEIPConfig_instanceAssociatedSwitch(rName string) string {
	return acctest.ConfigCompose(testAccEIPConfig_baseInstanceAssociated(rName), fmt.Sprintf(`
resource "aws_eip" "test" {
  domain = "vpc"

  instance                  = aws_instance.test[0].id
  associate_with_private_ip = aws_instance.test[0].private_ip

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEIPConfig_instanceAssociateNotAssociated(rName string) string {
	return acctest.ConfigCompose(testAccEIPConfig_baseInstance(rName), fmt.Sprintf(`
resource "aws_eip" "test" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEIPConfig_networkInterface(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test[0].id
  private_ips     = ["10.0.0.10"]
  security_groups = [aws_vpc.test.default_security_group_id]

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip" "test" {
  domain            = "vpc"
  network_interface = aws_network_interface.test.id

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}
`, rName))
}

func testAccEIPConfig_multiNetworkInterface(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test[0].id
  private_ips     = ["10.0.0.10", "10.0.0.11"]
  security_groups = [aws_vpc.test.default_security_group_id]

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip" "test" {
  count = 2

  domain                    = "vpc"
  network_interface         = aws_network_interface.test.id
  associate_with_private_ip = "10.0.0.1${count.index}"

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}
`, rName))
}

func testAccEIPConfig_baseAssociation(rName string) string {
	return acctest.ConfigCompose(testAccEIPConfig_baseInstance(rName), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test[0].id
  private_ips     = ["10.0.0.10"]
  security_groups = [aws_vpc.test.default_security_group_id]

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}
`, rName))
}

func testAccEIPConfig_associationNone(rName string) string {
	return acctest.ConfigCompose(testAccEIPConfig_baseAssociation(rName), fmt.Sprintf(`
resource "aws_eip" "test" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_network_interface.test]
}
`, rName))
}

func testAccEIPConfig_associationENI(rName string) string {
	return acctest.ConfigCompose(testAccEIPConfig_baseAssociation(rName), fmt.Sprintf(`
resource "aws_eip" "test" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }

  network_interface = aws_network_interface.test.id
}
`, rName))
}

func testAccEIPConfig_associationInstance(rName string) string {
	return acctest.ConfigCompose(testAccEIPConfig_baseAssociation(rName), fmt.Sprintf(`
resource "aws_eip" "test" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }

  instance = aws_instance.test.id

  depends_on = [aws_network_interface.test]
}
`, rName))
}

func testAccEIPConfig_publicIPv4PoolDefault(rName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccEIPConfig_publicIPv4PoolCustom(rName, poolName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  domain           = "vpc"
  public_ipv4_pool = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, poolName)
}

func testAccEIPConfig_customerOwnedIPv4Pool(rName string) string {
	return fmt.Sprintf(`
data "aws_ec2_coip_pools" "test" {}

resource "aws_eip" "test" {
  customer_owned_ipv4_pool = tolist(data.aws_ec2_coip_pools.test.pool_ids)[0]
  domain                   = "vpc"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccEIPConfig_networkBorderGroup(rName string) string {
	return fmt.Sprintf(`
data "aws_region" current {}

resource "aws_eip" "test" {
  domain               = "vpc"
  network_border_group = data.aws_region.current.name

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccEIPConfig_carrierIP(rName string) string {
	return acctest.ConfigCompose(
		testAccAvailableAZsWavelengthZonesDefaultExcludeConfig(),
		fmt.Sprintf(`
data "aws_availability_zone" "available" {
  name = data.aws_availability_zones.available.names[0]
}

resource "aws_eip" "test" {
  domain               = "vpc"
  network_border_group = data.aws_availability_zone.available.network_border_group

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEIPConfig_byoipAddressCustomDefault(rName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccEIPConfig_byoipAddressCustom(rName, address string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  domain  = "vpc"
  address = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, address)
}

func testAccEIPConfig_byoipAddressCustomPublicIPv4Pool(rName, address, poolName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  domain           = "vpc"
  address          = %[2]q
  public_ipv4_pool = %[3]q

  tags = {
    Name = %[1]q
  }
}
`, rName, address, poolName)
}
