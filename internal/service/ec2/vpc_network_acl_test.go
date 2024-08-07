// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
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

func TestAccVPCNetworkACL_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.NetworkAcl
	resourceName := "aws_network_acl.test"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkACLConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`network-acl/acl-.+`)),
					resource.TestCheckResourceAttr(resourceName, "egress.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct0),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
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

func TestAccVPCNetworkACL_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.NetworkAcl
	resourceName := "aws_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkACLConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceNetworkACL(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCNetworkACL_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.NetworkAcl
	resourceName := "aws_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkACLConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(ctx, resourceName, &v),
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
				Config: testAccVPCNetworkACLConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccVPCNetworkACLConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccVPCNetworkACL_Egress_mode(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.NetworkAcl
	resourceName := "aws_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkACLConfig_egressModeBlocks(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "egress.#", acctest.Ct2),
				),
			},
			{
				Config: testAccVPCNetworkACLConfig_egressModeNoBlocks(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "egress.#", acctest.Ct2),
				),
			},
			{
				Config: testAccVPCNetworkACLConfig_egressModeZeroed(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "egress.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccVPCNetworkACL_Ingress_mode(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.NetworkAcl
	resourceName := "aws_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkACLConfig_ingressModeBlocks(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct2),
				),
			},
			{
				Config: testAccVPCNetworkACLConfig_ingressModeNoBlocks(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct2),
				),
			},
			{
				Config: testAccVPCNetworkACLConfig_ingressModeZeroed(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccVPCNetworkACL_egressAndIngressRules(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.NetworkAcl
	resourceName := "aws_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkACLConfig_egressNIngress(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(ctx, resourceName, &v),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						names.AttrProtocol:  "6",
						"rule_no":           acctest.Ct1,
						"from_port":         "80",
						"to_port":           "80",
						names.AttrAction:    "allow",
						names.AttrCIDRBlock: "10.3.0.0/18",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						names.AttrProtocol:  "6",
						"rule_no":           acctest.Ct2,
						"from_port":         "443",
						"to_port":           "443",
						names.AttrAction:    "allow",
						names.AttrCIDRBlock: "10.3.0.0/18",
					}),
				),
			},
		},
	})
}

func TestAccVPCNetworkACL_OnlyIngressRules_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.NetworkAcl
	resourceName := "aws_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkACLConfig_ingress(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(ctx, resourceName, &v),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						names.AttrProtocol:  "6",
						"rule_no":           acctest.Ct2,
						"from_port":         "443",
						"to_port":           "443",
						names.AttrAction:    "deny",
						names.AttrCIDRBlock: "10.2.0.0/18",
					}),
				),
			},
		},
	})
}

func TestAccVPCNetworkACL_OnlyIngressRules_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.NetworkAcl
	resourceName := "aws_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkACLConfig_ingress(resourceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						names.AttrProtocol: "6",
						"rule_no":          acctest.Ct1,
						"from_port":        acctest.Ct0,
						"to_port":          "22",
						names.AttrAction:   "deny",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						names.AttrCIDRBlock: "10.2.0.0/18",
						"from_port":         "443",
						"rule_no":           acctest.Ct2,
					}),
				),
			},
			{
				Config: testAccVPCNetworkACLConfig_ingressChange(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						names.AttrProtocol:  "6",
						"rule_no":           acctest.Ct1,
						"from_port":         acctest.Ct0,
						"to_port":           "22",
						names.AttrAction:    "deny",
						names.AttrCIDRBlock: "10.2.0.0/18",
					}),
				),
			},
		},
	})
}

func TestAccVPCNetworkACL_caseSensitivityNoChanges(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.NetworkAcl
	resourceName := "aws_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkACLConfig_caseSensitive(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(ctx, resourceName, &v),
				),
			},
		},
	})
}

func TestAccVPCNetworkACL_onlyEgressRules(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.NetworkAcl
	resourceName := "aws_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkACLConfig_egress(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLExists(ctx, resourceName, &v),
				),
			},
		},
	})
}

func TestAccVPCNetworkACL_subnetChange(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.NetworkAcl
	resourceName := "aws_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkACLConfig_subnet(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test1", names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCNetworkACLConfig_subnetChange(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test2", names.AttrID),
				),
			},
		},
	})
}

func TestAccVPCNetworkACL_subnets(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.NetworkAcl
	resourceName := "aws_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkACLConfig_subnetSubnetIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test1", names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test2", names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCNetworkACLConfig_subnetSubnetIDsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", acctest.Ct3),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test1", names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test3", names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test4", names.AttrID),
				),
			},
		},
	})
}

func TestAccVPCNetworkACL_subnetsDelete(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.NetworkAcl
	resourceName := "aws_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkACLConfig_subnetSubnetIDs(resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test1", names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test2", names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCNetworkACLConfig_subnetSubnetIDsDeleteOne(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test1", names.AttrID),
				),
			},
		},
	})
}

func TestAccVPCNetworkACL_ipv6Rules(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.NetworkAcl
	resourceName := "aws_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkACLConfig_ipv6(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "ingress.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						names.AttrProtocol: "6",
						"rule_no":          acctest.Ct1,
						"from_port":        acctest.Ct0,
						"to_port":          "22",
						names.AttrAction:   "allow",
						"ipv6_cidr_block":  "::/0",
					}),
				),
			},
		},
	})
}

func TestAccVPCNetworkACL_ipv6ICMPRules(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.NetworkAcl
	resourceName := "aws_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkACLConfig_ipv6ICMP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLExists(ctx, resourceName, &v),
				),
			},
		},
	})
}

func TestAccVPCNetworkACL_ipv6VPCRules(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.NetworkAcl
	resourceName := "aws_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkACLConfig_ipv6VPC(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"ipv6_cidr_block": "2600:1f16:d1e:9a00::/56",
					}),
				),
			},
		},
	})
}

func TestAccVPCNetworkACL_espProtocol(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.NetworkAcl
	resourceName := "aws_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkACLConfig_esp(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLExists(ctx, resourceName, &v),
				),
			},
		},
	})
}

func testAccCheckNetworkACLDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_network_acl" {
				continue
			}

			_, err := tfec2.FindNetworkACLByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Network ACL %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckNetworkACLExists(ctx context.Context, n string, v *awstypes.NetworkAcl) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Network ACL ID is set: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindNetworkACLByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVPCNetworkACLConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id
}
`, rName)
}

func testAccVPCNetworkACLConfig_ipv6ICMP(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  ingress {
    action          = "allow"
    from_port       = 0
    icmp_code       = -1
    icmp_type       = -1
    ipv6_cidr_block = "::/0"
    protocol        = 58
    rule_no         = 1
    to_port         = 0
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCNetworkACLConfig_ipv6(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block              = "10.1.1.0/24"
  vpc_id                  = aws_vpc.test.id
  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  ingress {
    protocol        = 6
    rule_no         = 1
    action          = "allow"
    ipv6_cidr_block = "::/0"
    from_port       = 0
    to_port         = 22
  }

  subnet_ids = [aws_subnet.test.id]

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCNetworkACLConfig_ipv6VPC(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  ingress {
    protocol        = 6
    rule_no         = 1
    action          = "allow"
    ipv6_cidr_block = "2600:1f16:d1e:9a00::/56"
    from_port       = 0
    to_port         = 22
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCNetworkACLConfig_ingress(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block              = "10.1.1.0/24"
  vpc_id                  = aws_vpc.test.id
  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  ingress {
    protocol   = 6
    rule_no    = 1
    action     = "deny"
    cidr_block = "10.2.0.0/18"
    from_port  = 0
    to_port    = 22
  }

  ingress {
    protocol   = 6
    rule_no    = 2
    action     = "deny"
    cidr_block = "10.2.0.0/18"
    from_port  = 443
    to_port    = 443
  }

  subnet_ids = [aws_subnet.test.id]

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCNetworkACLConfig_caseSensitive(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block              = "10.1.1.0/24"
  vpc_id                  = aws_vpc.test.id
  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  ingress {
    protocol   = 6
    rule_no    = 1
    action     = "Allow"
    cidr_block = "10.2.0.0/18"
    from_port  = 0
    to_port    = 22
  }

  subnet_ids = [aws_subnet.test.id]

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCNetworkACLConfig_ingressChange(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block              = "10.1.1.0/24"
  vpc_id                  = aws_vpc.test.id
  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  ingress {
    protocol   = 6
    rule_no    = 1
    action     = "deny"
    cidr_block = "10.2.0.0/18"
    from_port  = 0
    to_port    = 22
  }

  subnet_ids = [aws_subnet.test.id]

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCNetworkACLConfig_egress(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.2.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block              = "10.2.0.0/24"
  vpc_id                  = aws_vpc.test.id
  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  egress {
    protocol   = 6
    rule_no    = 2
    action     = "allow"
    cidr_block = "10.2.0.0/18"
    from_port  = 443
    to_port    = 443
  }

  egress {
    protocol   = "-1"
    rule_no    = 4
    action     = "allow"
    cidr_block = "0.0.0.0/0"
    from_port  = 0
    to_port    = 0
  }

  egress {
    protocol   = 6
    rule_no    = 1
    action     = "allow"
    cidr_block = "10.2.0.0/18"
    from_port  = 80
    to_port    = 80
  }

  egress {
    protocol   = 6
    rule_no    = 3
    action     = "allow"
    cidr_block = "10.2.0.0/18"
    from_port  = 22
    to_port    = 22
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCNetworkACLConfig_egressNIngress(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block              = "10.3.0.0/24"
  vpc_id                  = aws_vpc.test.id
  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  egress {
    protocol   = 6
    rule_no    = 2
    action     = "allow"
    cidr_block = "10.3.0.0/18"
    from_port  = 443
    to_port    = 443
  }

  ingress {
    protocol   = 6
    rule_no    = 1
    action     = "allow"
    cidr_block = "10.3.0.0/18"
    from_port  = 80
    to_port    = 80
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCNetworkACLConfig_subnet(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test1" {
  cidr_block              = "10.1.111.0/24"
  vpc_id                  = aws_vpc.test.id
  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  cidr_block              = "10.1.1.0/24"
  vpc_id                  = aws_vpc.test.id
  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test1" {
  vpc_id     = aws_vpc.test.id
  subnet_ids = [aws_subnet.test2.id]

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id     = aws_vpc.test.id
  subnet_ids = [aws_subnet.test1.id]

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCNetworkACLConfig_subnetChange(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test1" {
  cidr_block              = "10.1.111.0/24"
  vpc_id                  = aws_vpc.test.id
  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  cidr_block              = "10.1.1.0/24"
  vpc_id                  = aws_vpc.test.id
  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id     = aws_vpc.test.id
  subnet_ids = [aws_subnet.test2.id]

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCNetworkACLConfig_subnetSubnetIDs(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test1" {
  cidr_block = "10.1.111.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id     = aws_vpc.test.id
  subnet_ids = [aws_subnet.test1.id, aws_subnet.test2.id]

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCNetworkACLConfig_subnetSubnetIDsUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test1" {
  cidr_block = "10.1.111.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test3" {
  cidr_block = "10.1.222.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test4" {
  cidr_block = "10.1.4.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id
  subnet_ids = [
    aws_subnet.test1.id,
    aws_subnet.test3.id,
    aws_subnet.test4.id,
  ]

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCNetworkACLConfig_subnetSubnetIDsDeleteOne(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test1" {
  cidr_block = "10.1.111.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id     = aws_vpc.test.id
  subnet_ids = [aws_subnet.test1.id]

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCNetworkACLConfig_esp(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  egress {
    protocol   = "esp"
    rule_no    = 5
    action     = "allow"
    cidr_block = "10.3.0.0/18"
    from_port  = 0
    to_port    = 0
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCNetworkACLConfig_egressModeBlocks(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  tags = {
    Name = %[1]q
  }

  vpc_id = aws_vpc.test.id

  egress {
    action     = "allow"
    cidr_block = aws_vpc.test.cidr_block
    from_port  = 0
    protocol   = 6
    rule_no    = 1
    to_port    = 0
  }

  egress {
    action     = "allow"
    cidr_block = aws_vpc.test.cidr_block
    from_port  = 0
    protocol   = "udp"
    rule_no    = 2
    to_port    = 0
  }
}
`, rName)
}

func testAccVPCNetworkACLConfig_egressModeNoBlocks(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  tags = {
    Name = %[1]q
  }

  vpc_id = aws_vpc.test.id
}
`, rName)
}

func testAccVPCNetworkACLConfig_egressModeZeroed(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  egress = []

  tags = {
    Name = %[1]q
  }

  vpc_id = aws_vpc.test.id
}
`, rName)
}

func testAccVPCNetworkACLConfig_ingressModeBlocks(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  tags = {
    Name = %[1]q
  }

  vpc_id = aws_vpc.test.id

  ingress {
    action     = "allow"
    cidr_block = aws_vpc.test.cidr_block
    from_port  = 0
    protocol   = 6
    rule_no    = 1
    to_port    = 0
  }

  ingress {
    action     = "allow"
    cidr_block = aws_vpc.test.cidr_block
    from_port  = 0
    protocol   = "udp"
    rule_no    = 2
    to_port    = 0
  }
}
`, rName)
}

func testAccVPCNetworkACLConfig_ingressModeNoBlocks(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  tags = {
    Name = %[1]q
  }

  vpc_id = aws_vpc.test.id
}
`, rName)
}

func testAccVPCNetworkACLConfig_ingressModeZeroed(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  ingress = []

  tags = {
    Name = %[1]q
  }

  vpc_id = aws_vpc.test.id
}
`, rName)
}

func testAccVPCNetworkACLConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccVPCNetworkACLConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
