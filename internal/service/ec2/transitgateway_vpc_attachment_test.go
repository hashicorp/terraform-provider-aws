// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfplancheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/plancheck"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccTransitGatewayVPCAttachment_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGatewayVpcAttachment1 awstypes.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGatewayVPCAttachment(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayVPCAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment1),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "ec2", "transit-gateway-attachment/{id}"),
					resource.TestCheckResourceAttr(resourceName, "dns_support", string(awstypes.DnsSupportValueEnable)),
					resource.TestCheckResourceAttr(resourceName, "ipv6_support", string(awstypes.Ipv6SupportValueDisable)),
					resource.TestCheckResourceAttr(resourceName, "security_group_referencing_support", string(awstypes.SecurityGroupReferencingSupportValueEnable)),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayID, transitGatewayResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_owner_id"),
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

func testAccTransitGatewayVPCAttachment_disappears(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGatewayVpcAttachment1 awstypes.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGatewayVPCAttachment(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayVPCAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment1),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceTransitGatewayVPCAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTransitGatewayVPCAttachment_applianceModeSupport(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGatewayVpcAttachment1, transitGatewayVpcAttachment2 awstypes.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGatewayVPCAttachment(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayVPCAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentConfig_applianceModeSupport(rName, string(awstypes.ApplianceModeSupportValueDisable)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment1),
					resource.TestCheckResourceAttr(resourceName, "appliance_mode_support", string(awstypes.ApplianceModeSupportValueDisable)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayVPCAttachmentConfig_applianceModeSupport(rName, string(awstypes.ApplianceModeSupportValueEnable)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment2),
					testAccCheckTransitGatewayVPCAttachmentNotRecreated(&transitGatewayVpcAttachment1, &transitGatewayVpcAttachment2),
					resource.TestCheckResourceAttr(resourceName, "appliance_mode_support", string(awstypes.ApplianceModeSupportValueEnable)),
				),
			},
		},
	})
}

func testAccTransitGatewayVPCAttachment_dnsSupport(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGatewayVpcAttachment1, transitGatewayVpcAttachment2 awstypes.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGatewayVPCAttachment(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayVPCAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentConfig_dnsSupport(rName, string(awstypes.DnsSupportValueDisable)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment1),
					resource.TestCheckResourceAttr(resourceName, "dns_support", string(awstypes.DnsSupportValueDisable)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayVPCAttachmentConfig_dnsSupport(rName, string(awstypes.DnsSupportValueEnable)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment2),
					testAccCheckTransitGatewayVPCAttachmentNotRecreated(&transitGatewayVpcAttachment1, &transitGatewayVpcAttachment2),
					resource.TestCheckResourceAttr(resourceName, "dns_support", string(awstypes.DnsSupportValueEnable)),
				),
			},
		},
	})
}

func testAccTransitGatewayVPCAttachment_ipv6Support(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGatewayVpcAttachment1, transitGatewayVpcAttachment2 awstypes.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGatewayVPCAttachment(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayVPCAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentConfig_ipv6Support(rName, string(awstypes.Ipv6SupportValueEnable)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment1),
					resource.TestCheckResourceAttr(resourceName, "ipv6_support", string(awstypes.Ipv6SupportValueEnable)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayVPCAttachmentConfig_ipv6Support(rName, string(awstypes.Ipv6SupportValueDisable)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment2),
					testAccCheckTransitGatewayVPCAttachmentNotRecreated(&transitGatewayVpcAttachment1, &transitGatewayVpcAttachment2),
					resource.TestCheckResourceAttr(resourceName, "ipv6_support", string(awstypes.Ipv6SupportValueDisable)),
				),
			},
		},
	})
}

func testAccTransitGatewayVPCAttachment_securityGroupReferencingSupport(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGatewayVpcAttachment1, transitGatewayVpcAttachment2 awstypes.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGatewayVPCAttachment(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayVPCAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentConfig_securityGroupReferencingSupport(rName, string(awstypes.SecurityGroupReferencingSupportValueDisable)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment1),
					resource.TestCheckResourceAttr(resourceName, "security_group_referencing_support", string(awstypes.SecurityGroupReferencingSupportValueDisable)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayVPCAttachmentConfig_securityGroupReferencingSupport(rName, string(awstypes.SecurityGroupReferencingSupportValueEnable)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment2),
					testAccCheckTransitGatewayVPCAttachmentNotRecreated(&transitGatewayVpcAttachment1, &transitGatewayVpcAttachment2),
					resource.TestCheckResourceAttr(resourceName, "security_group_referencing_support", string(awstypes.SecurityGroupReferencingSupportValueEnable)),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/39518.
// Resources created at <= v5.58.0 show drift after upgrade to v5.69.0.
func testAccTransitGatewayVPCAttachment_securityGroupReferencingSupportV5690Diff(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGatewayVpcAttachment1 awstypes.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGatewayVPCAttachment(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy: testAccCheckTransitGatewayVPCAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.68.0",
					},
				},
				Config: testAccTransitGatewayVPCAttachmentConfig_basic(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoValue(resourceName, tfjsonpath.New("security_group_referencing_support")),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment1),
				),
			},
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.69.0",
					},
				},
				Config: testAccTransitGatewayVPCAttachmentConfig_basic(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						tfplancheck.ExpectKnownValueChange(resourceName, tfjsonpath.New("security_group_referencing_support"), knownvalue.StringExact(string(awstypes.SecurityGroupReferencingSupportValueEnable)), knownvalue.StringExact(string(awstypes.SecurityGroupReferencingSupportValueDisable))),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("security_group_referencing_support"), knownvalue.StringExact(string(awstypes.SecurityGroupReferencingSupportValueDisable))),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment1),
				),
			},
		},
	})
}

func testAccTransitGatewayVPCAttachment_securityGroupReferencingSupportExistingResource(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGatewayVpcAttachment1 awstypes.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGatewayVPCAttachment(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy: testAccCheckTransitGatewayVPCAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.68.0",
					},
				},
				Config: testAccTransitGatewayVPCAttachmentConfig_basic(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoValue(resourceName, tfjsonpath.New("security_group_referencing_support")),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment1),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccTransitGatewayVPCAttachmentConfig_basic(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("security_group_referencing_support"), knownvalue.StringExact(string(awstypes.SecurityGroupReferencingSupportValueEnable))),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment1),
				),
			},
		},
	})
}

func testAccTransitGatewayVPCAttachment_sharedTransitGateway(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGatewayVpcAttachment1 awstypes.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			testAccPreCheckTransitGatewayVPCAttachment(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckTransitGatewayVPCAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentConfig_sharedTransitGateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment1),
				),
			},
			{
				Config:            testAccTransitGatewayVPCAttachmentConfig_sharedTransitGateway(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTransitGatewayVPCAttachment_subnetIDs(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGatewayVpcAttachment1, transitGatewayVpcAttachment2, transitGatewayVpcAttachment3 awstypes.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGatewayVPCAttachment(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayVPCAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentConfig_subnetIDs2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment1),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayVPCAttachmentConfig_subnetIDs1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment2),
					testAccCheckTransitGatewayVPCAttachmentNotRecreated(&transitGatewayVpcAttachment1, &transitGatewayVpcAttachment2),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
				),
			},
			{
				Config: testAccTransitGatewayVPCAttachmentConfig_subnetIDs2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment3),
					testAccCheckTransitGatewayVPCAttachmentNotRecreated(&transitGatewayVpcAttachment2, &transitGatewayVpcAttachment3),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
				),
			},
		},
	})
}

func testAccTransitGatewayVPCAttachment_tags(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGatewayVpcAttachment1, transitGatewayVpcAttachment2, transitGatewayVpcAttachment3 awstypes.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGatewayVPCAttachment(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayVPCAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment1),
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
				Config: testAccTransitGatewayVPCAttachmentConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment2),
					testAccCheckTransitGatewayVPCAttachmentNotRecreated(&transitGatewayVpcAttachment1, &transitGatewayVpcAttachment2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccTransitGatewayVPCAttachmentConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment3),
					testAccCheckTransitGatewayVPCAttachmentNotRecreated(&transitGatewayVpcAttachment2, &transitGatewayVpcAttachment3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccTransitGatewayVPCAttachment_transitGatewayDefaultRouteTableAssociationAndPropagationDisabled(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGateway1 awstypes.TransitGateway
	var transitGatewayVpcAttachment1 awstypes.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGatewayVPCAttachment(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayVPCAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentConfig_defaultRouteTableAssociationAndPropagationDisabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, transitGatewayResourceName, &transitGateway1),
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment1),
					testAccCheckTransitGatewayAssociationDefaultRouteTableAttachmentNotAssociated(ctx, &transitGateway1, &transitGatewayVpcAttachment1),
					testAccCheckTransitGatewayPropagationDefaultRouteTableAttachmentNotPropagated(ctx, &transitGateway1, &transitGatewayVpcAttachment1),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", acctest.CtFalse),
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

func testAccTransitGatewayVPCAttachment_transitGatewayDefaultRouteTableAssociation(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGateway1, transitGateway2, transitGateway3 awstypes.TransitGateway
	var transitGatewayVpcAttachment1, transitGatewayVpcAttachment2, transitGatewayVpcAttachment3 awstypes.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGatewayVPCAttachment(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayVPCAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentConfig_defaultRouteTableAssociation(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, transitGatewayResourceName, &transitGateway1),
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment1),
					testAccCheckTransitGatewayAssociationDefaultRouteTableAttachmentNotAssociated(ctx, &transitGateway1, &transitGatewayVpcAttachment1),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayVPCAttachmentConfig_defaultRouteTableAssociation(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, transitGatewayResourceName, &transitGateway2),
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment2),
					testAccCheckTransitGatewayVPCAttachmentNotRecreated(&transitGatewayVpcAttachment1, &transitGatewayVpcAttachment2),
					testAccCheckTransitGatewayAssociationDefaultRouteTableAttachmentAssociated(ctx, &transitGateway2, &transitGatewayVpcAttachment2),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", acctest.CtTrue),
				),
			},
			{
				Config: testAccTransitGatewayVPCAttachmentConfig_defaultRouteTableAssociation(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, transitGatewayResourceName, &transitGateway3),
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment3),
					testAccCheckTransitGatewayVPCAttachmentNotRecreated(&transitGatewayVpcAttachment2, &transitGatewayVpcAttachment3),
					testAccCheckTransitGatewayAssociationDefaultRouteTableAttachmentNotAssociated(ctx, &transitGateway3, &transitGatewayVpcAttachment3),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccTransitGatewayVPCAttachment_transitGatewayDefaultRouteTablePropagation(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGateway1, transitGateway2, transitGateway3 awstypes.TransitGateway
	var transitGatewayVpcAttachment1, transitGatewayVpcAttachment2, transitGatewayVpcAttachment3 awstypes.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGatewayVPCAttachment(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayVPCAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentConfig_defaultRouteTablePropagation(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, transitGatewayResourceName, &transitGateway1),
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment1),
					testAccCheckTransitGatewayPropagationDefaultRouteTableAttachmentNotPropagated(ctx, &transitGateway1, &transitGatewayVpcAttachment1),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayVPCAttachmentConfig_defaultRouteTablePropagation(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, transitGatewayResourceName, &transitGateway2),
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment2),
					testAccCheckTransitGatewayVPCAttachmentNotRecreated(&transitGatewayVpcAttachment1, &transitGatewayVpcAttachment2),
					testAccCheckTransitGatewayPropagationDefaultRouteTableAttachmentPropagated(ctx, &transitGateway2, &transitGatewayVpcAttachment2),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", acctest.CtTrue),
				),
			},
			{
				Config: testAccTransitGatewayVPCAttachmentConfig_defaultRouteTablePropagation(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, transitGatewayResourceName, &transitGateway3),
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment3),
					testAccCheckTransitGatewayVPCAttachmentNotRecreated(&transitGatewayVpcAttachment2, &transitGatewayVpcAttachment3),
					testAccCheckTransitGatewayPropagationDefaultRouteTableAttachmentNotPropagated(ctx, &transitGateway3, &transitGatewayVpcAttachment3),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccCheckTransitGatewayVPCAttachmentExists(ctx context.Context, n string, v *awstypes.TransitGatewayVpcAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Transit Gateway VPC Attachment ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindTransitGatewayVPCAttachmentByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTransitGatewayVPCAttachmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_transit_gateway_vpc_attachment" {
				continue
			}

			_, err := tfec2.FindTransitGatewayVPCAttachmentByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Transit Gateway VPC Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTransitGatewayVPCAttachmentNotRecreated(i, j *awstypes.TransitGatewayVpcAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.TransitGatewayAttachmentId) != aws.ToString(j.TransitGatewayAttachmentId) {
			return errors.New("EC2 Transit Gateway VPC Attachment was recreated")
		}

		return nil
	}
}

func testAccTransitGatewayVPCAttachmentConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccTransitGatewayVPCAttachmentConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccTransitGatewayVPCAttachmentConfig_base(rName), `
resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = aws_subnet.test[*].id
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}
`)
}

func testAccTransitGatewayVPCAttachmentConfig_applianceModeSupport(rName, appModeSupport string) string {
	return acctest.ConfigCompose(testAccTransitGatewayVPCAttachmentConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  appliance_mode_support = %[2]q
  subnet_ids             = aws_subnet.test[*].id
  transit_gateway_id     = aws_ec2_transit_gateway.test.id
  vpc_id                 = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, appModeSupport))
}

func testAccTransitGatewayVPCAttachmentConfig_dnsSupport(rName, dnsSupport string) string {
	return acctest.ConfigCompose(testAccTransitGatewayVPCAttachmentConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  dns_support        = %[2]q
  subnet_ids         = aws_subnet.test[*].id
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, dnsSupport))
}

func testAccTransitGatewayVPCAttachmentConfig_securityGroupReferencingSupport(rName, securityGroupReferencingSupport string) string {
	return acctest.ConfigCompose(testAccTransitGatewayVPCAttachmentConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  security_group_referencing_support = %[2]q
  subnet_ids                         = aws_subnet.test[*].id
  transit_gateway_id                 = aws_ec2_transit_gateway.test.id
  vpc_id                             = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, securityGroupReferencingSupport))
}

func testAccTransitGatewayVPCAttachmentConfig_ipv6Support(rName, ipv6Support string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnetsIPv6(rName, 1), fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  ipv6_support       = %[2]q
  subnet_ids         = aws_subnet.test[*].id
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, ipv6Support))
}

func testAccTransitGatewayVPCAttachmentConfig_sharedTransitGateway(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
data "aws_organizations_organization" "test" {}

resource "aws_ec2_transit_gateway" "test" {
  provider = "awsalternate"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ram_resource_share" "test" {
  provider = "awsalternate"

  name = %[1]q
}

resource "aws_ram_resource_association" "test" {
  provider = "awsalternate"

  resource_arn       = aws_ec2_transit_gateway.test.arn
  resource_share_arn = aws_ram_resource_share.test.id
}

resource "aws_ram_principal_association" "test" {
  provider = "awsalternate"

  principal          = data.aws_organizations_organization.test.arn
  resource_share_arn = aws_ram_resource_share.test.id
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  depends_on = [aws_ram_principal_association.test, aws_ram_resource_association.test]

  subnet_ids         = aws_subnet.test[*].id
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccTransitGatewayVPCAttachmentConfig_subnetIDs1(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test[0].id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccTransitGatewayVPCAttachmentConfig_subnetIDs2(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = aws_subnet.test[*].id
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccTransitGatewayVPCAttachmentConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccTransitGatewayVPCAttachmentConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = aws_subnet.test[*].id
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccTransitGatewayVPCAttachmentConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccTransitGatewayVPCAttachmentConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = aws_subnet.test[*].id
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccTransitGatewayVPCAttachmentConfig_defaultRouteTableAssociationAndPropagationDisabled(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  default_route_table_association = "disable"
  default_route_table_propagation = "disable"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids                                      = aws_subnet.test[*].id
  transit_gateway_default_route_table_association = false
  transit_gateway_default_route_table_propagation = false
  transit_gateway_id                              = aws_ec2_transit_gateway.test.id
  vpc_id                                          = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccTransitGatewayVPCAttachmentConfig_defaultRouteTableAssociation(rName string, transitGatewayDefaultRouteTableAssociation bool) string {
	return acctest.ConfigCompose(testAccTransitGatewayVPCAttachmentConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids                                      = aws_subnet.test[*].id
  transit_gateway_default_route_table_association = %[2]t
  transit_gateway_id                              = aws_ec2_transit_gateway.test.id
  vpc_id                                          = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, transitGatewayDefaultRouteTableAssociation))
}

func testAccTransitGatewayVPCAttachmentConfig_defaultRouteTablePropagation(rName string, transitGatewayDefaultRouteTablePropagation bool) string {
	return acctest.ConfigCompose(testAccTransitGatewayVPCAttachmentConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids                                      = aws_subnet.test[*].id
  transit_gateway_default_route_table_propagation = %[2]t
  transit_gateway_id                              = aws_ec2_transit_gateway.test.id
  vpc_id                                          = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, transitGatewayDefaultRouteTablePropagation))
}
