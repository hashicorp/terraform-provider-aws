package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccEC2NetworkACL_basic(t *testing.T) {
	resourceName := "aws_network_acl.test"
	var networkAcl ec2.NetworkAcl

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLEgressNIngressConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`network-acl/acl-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccEC2NetworkACL_tags(t *testing.T) {
	resourceName := "aws_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var networkAcl ec2.NetworkAcl

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNetworkACLTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccNetworkACLTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEC2NetworkACL_disappears(t *testing.T) {
	var networkAcl ec2.NetworkAcl
	resourceName := "aws_network_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLEgressNIngressConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceNetworkACL(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2NetworkACL_Egress_mode(t *testing.T) {
	var networkAcl1, networkAcl2, networkAcl3 ec2.NetworkAcl
	resourceName := "aws_network_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLEgressModeBlocksConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl1),
					testAccCheckNetworkACLEgressRuleLength(&networkAcl1, 2),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNetworkACLEgressModeNoBlocksConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl2),
					testAccCheckNetworkACLEgressRuleLength(&networkAcl2, 2),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNetworkACLEgressModeZeroedConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl3),
					testAccCheckNetworkACLEgressRuleLength(&networkAcl3, 0),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "0"),
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

func TestAccEC2NetworkACL_Ingress_mode(t *testing.T) {
	var networkAcl1, networkAcl2, networkAcl3 ec2.NetworkAcl
	resourceName := "aws_network_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLIngressModeBlocksConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl1),
					testIngressRuleLength(&networkAcl1, 2),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNetworkACLIngressModeNoBlocksConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl2),
					testIngressRuleLength(&networkAcl2, 2),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNetworkACLIngressModeZeroedConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl3),
					testIngressRuleLength(&networkAcl3, 0),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "0"),
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

func TestAccEC2NetworkACL_egressAndIngressRules(t *testing.T) {
	var networkAcl ec2.NetworkAcl
	resourceName := "aws_network_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLEgressNIngressConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"protocol":   "6",
						"rule_no":    "1",
						"from_port":  "80",
						"to_port":    "80",
						"action":     "allow",
						"cidr_block": "10.3.0.0/18",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"protocol":   "6",
						"rule_no":    "2",
						"from_port":  "443",
						"to_port":    "443",
						"action":     "allow",
						"cidr_block": "10.3.0.0/18",
					}),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
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

func TestAccEC2NetworkACL_OnlyIngressRules_basic(t *testing.T) {
	var networkAcl ec2.NetworkAcl
	resourceName := "aws_network_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLIngressConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"protocol":   "6",
						"rule_no":    "2",
						"from_port":  "443",
						"to_port":    "443",
						"action":     "deny",
						"cidr_block": "10.2.0.0/18",
					}),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
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

func TestAccEC2NetworkACL_OnlyIngressRules_update(t *testing.T) {
	var networkAcl ec2.NetworkAcl
	resourceName := "aws_network_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLIngressConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl),
					testIngressRuleLength(&networkAcl, 2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"protocol":  "6",
						"rule_no":   "1",
						"from_port": "0",
						"to_port":   "22",
						"action":    "deny",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_block": "10.2.0.0/18",
						"from_port":  "443",
						"rule_no":    "2",
					}),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNetworkACLIngressChangeConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl),
					testIngressRuleLength(&networkAcl, 1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"protocol":   "6",
						"rule_no":    "1",
						"from_port":  "0",
						"to_port":    "22",
						"action":     "deny",
						"cidr_block": "10.2.0.0/18",
					}),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
				),
			},
		},
	})
}

func TestAccEC2NetworkACL_caseSensitivityNoChanges(t *testing.T) {
	var networkAcl ec2.NetworkAcl
	resourceName := "aws_network_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLCaseSensitiveConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl),
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

func TestAccEC2NetworkACL_onlyEgressRules(t *testing.T) {
	var networkAcl ec2.NetworkAcl
	resourceName := "aws_network_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLEgressConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "tf-acc-acl-egress"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
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

func TestAccEC2NetworkACL_subnetChange(t *testing.T) {
	var networkAcl ec2.NetworkAcl
	resourceName := "aws_network_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLSubnetConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl),
					testAccCheckSubnetIsAssociatedWithAcl(resourceName, "aws_subnet.old"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNetworkACLSubnetChangeConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl),
					testAccCheckSubnetIsNotAssociatedWithAcl(resourceName, "aws_subnet.old"),
					testAccCheckSubnetIsAssociatedWithAcl(resourceName, "aws_subnet.new"),
				),
			},
		},
	})

}

func TestAccEC2NetworkACL_subnets(t *testing.T) {
	var networkAcl ec2.NetworkAcl
	resourceName := "aws_network_acl.test"

	checkACLSubnets := func(acl *ec2.NetworkAcl, count int) resource.TestCheckFunc {
		return func(*terraform.State) (err error) {
			if count != len(acl.Associations) {
				return fmt.Errorf("ACL association count does not match, expected %d, got %d", count, len(acl.Associations))
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLSubnet_SubnetIDs,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl),
					testAccCheckSubnetIsAssociatedWithAcl(resourceName, "aws_subnet.one"),
					testAccCheckSubnetIsAssociatedWithAcl(resourceName, "aws_subnet.two"),
					checkACLSubnets(&networkAcl, 2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNetworkACLSubnet_SubnetIdsUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl),
					testAccCheckSubnetIsAssociatedWithAcl(resourceName, "aws_subnet.one"),
					testAccCheckSubnetIsAssociatedWithAcl(resourceName, "aws_subnet.three"),
					testAccCheckSubnetIsAssociatedWithAcl(resourceName, "aws_subnet.four"),
					checkACLSubnets(&networkAcl, 3),
				),
			},
		},
	})
}

func TestAccEC2NetworkACL_subnetsDelete(t *testing.T) {
	var networkAcl ec2.NetworkAcl
	resourceName := "aws_network_acl.test"

	checkACLSubnets := func(acl *ec2.NetworkAcl, count int) resource.TestCheckFunc {
		return func(*terraform.State) (err error) {
			if count != len(acl.Associations) {
				return fmt.Errorf("ACL association count does not match, expected %d, got %d", count, len(acl.Associations))
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLSubnet_SubnetIDs,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl),
					testAccCheckSubnetIsAssociatedWithAcl(resourceName, "aws_subnet.one"),
					testAccCheckSubnetIsAssociatedWithAcl(resourceName, "aws_subnet.two"),
					checkACLSubnets(&networkAcl, 2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNetworkACLSubnet_SubnetIdsDeleteOne,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl),
					testAccCheckSubnetIsAssociatedWithAcl(resourceName, "aws_subnet.one"),
					checkACLSubnets(&networkAcl, 1),
				),
			},
		},
	})
}

func TestAccEC2NetworkACL_ipv6Rules(t *testing.T) {
	var networkAcl ec2.NetworkAcl
	resourceName := "aws_network_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLIPv6Config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl),
					resource.TestCheckResourceAttr(
						resourceName, "ingress.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"protocol":        "6",
						"rule_no":         "1",
						"from_port":       "0",
						"to_port":         "22",
						"action":          "allow",
						"ipv6_cidr_block": "::/0",
					}),
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

func TestAccEC2NetworkACL_ipv6ICMPRules(t *testing.T) {
	var networkAcl ec2.NetworkAcl
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_network_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLIPv6ICMPConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl),
				),
			},
		},
	})
}

func TestAccEC2NetworkACL_ipv6VPCRules(t *testing.T) {
	var networkAcl ec2.NetworkAcl
	resourceName := "aws_network_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLIPv6VPCConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl),
					resource.TestCheckResourceAttr(
						resourceName, "ingress.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"ipv6_cidr_block": "2600:1f16:d1e:9a00::/56",
					}),
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

func TestAccEC2NetworkACL_espProtocol(t *testing.T) {
	var networkAcl ec2.NetworkAcl
	resourceName := "aws_network_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLEsp,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl),
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

func testAccCheckNetworkACLDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_network" {
			continue
		}

		// Retrieve the network acl
		resp, err := conn.DescribeNetworkAcls(&ec2.DescribeNetworkAclsInput{
			NetworkAclIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err == nil {
			if len(resp.NetworkAcls) > 0 && *resp.NetworkAcls[0].NetworkAclId == rs.Primary.ID {
				return fmt.Errorf("Network Acl (%s) still exists.", rs.Primary.ID)
			}

			return nil
		}

		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		// Confirm error code is what we want
		if ec2err.Code() != "InvalidNetworkAclID.NotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckNetworkACLExists(n string, networkAcl *ec2.NetworkAcl) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set: %s", n)
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		resp, err := conn.DescribeNetworkAcls(&ec2.DescribeNetworkAclsInput{
			NetworkAclIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}

		if len(resp.NetworkAcls) > 0 && *resp.NetworkAcls[0].NetworkAclId == rs.Primary.ID {
			*networkAcl = *resp.NetworkAcls[0]
			return nil
		}

		return fmt.Errorf("Network Acls not found")
	}
}

func testAccCheckNetworkACLEgressRuleLength(networkAcl *ec2.NetworkAcl, length int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var entries []*ec2.NetworkAclEntry
		for _, entry := range networkAcl.Entries {
			if aws.BoolValue(entry.Egress) {
				entries = append(entries, entry)
			}
		}
		// There is always a default rule (ALL Traffic ... DENY)
		// so we have to increase the length by 1
		if len(entries) != length+1 {
			return fmt.Errorf("Invalid number of ingress entries found; count = %d", len(entries))
		}
		return nil
	}
}

func testIngressRuleLength(networkAcl *ec2.NetworkAcl, length int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var ingressEntries []*ec2.NetworkAclEntry
		for _, e := range networkAcl.Entries {
			if !*e.Egress {
				ingressEntries = append(ingressEntries, e)
			}
		}
		// There is always a default rule (ALL Traffic ... DENY)
		// so we have to increase the length by 1
		if len(ingressEntries) != length+1 {
			return fmt.Errorf("Invalid number of ingress entries found; count = %d", len(ingressEntries))
		}
		return nil
	}
}

func testAccCheckSubnetIsAssociatedWithAcl(acl string, sub string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		networkAcl := s.RootModule().Resources[acl]
		subnet := s.RootModule().Resources[sub]

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		resp, err := conn.DescribeNetworkAcls(&ec2.DescribeNetworkAclsInput{
			NetworkAclIds: []*string{aws.String(networkAcl.Primary.ID)},
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("association.subnet-id"),
					Values: []*string{aws.String(subnet.Primary.ID)},
				},
			},
		})
		if err != nil {
			return err
		}
		if len(resp.NetworkAcls) > 0 {
			return nil
		}

		return fmt.Errorf("Network Acl %s is not associated with subnet %s", acl, sub)
	}
}

func testAccCheckSubnetIsNotAssociatedWithAcl(acl string, subnet string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		networkAcl := s.RootModule().Resources[acl]
		subnet := s.RootModule().Resources[subnet]

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		resp, err := conn.DescribeNetworkAcls(&ec2.DescribeNetworkAclsInput{
			NetworkAclIds: []*string{aws.String(networkAcl.Primary.ID)},
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("association.subnet-id"),
					Values: []*string{aws.String(subnet.Primary.ID)},
				},
			},
		})

		if err != nil {
			return err
		}
		if len(resp.NetworkAcls) > 0 {
			return fmt.Errorf("Network Acl %s is still associated with subnet %s", acl, subnet)
		}
		return nil
	}
}

func testAccNetworkACLIPv6ICMPConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %q
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
    Name = %q
  }
}
`, rName, rName)
}

const testAccNetworkACLIPv6Config = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-network-acl-ipv6"
  }
}

resource "aws_subnet" "blob" {
  cidr_block              = "10.1.1.0/24"
  vpc_id                  = aws_vpc.foo.id
  map_public_ip_on_launch = true

  tags = {
    Name = "tf-acc-network-acl-ipv6"
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.foo.id

  ingress {
    protocol        = 6
    rule_no         = 1
    action          = "allow"
    ipv6_cidr_block = "::/0"
    from_port       = 0
    to_port         = 22
  }

  subnet_ids = [aws_subnet.blob.id]

  tags = {
    Name = "tf-acc-acl-ipv6"
  }
}
`

const testAccNetworkACLIPv6VPCConfig = `
resource "aws_vpc" "foo" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = "terraform-testacc-network-acl-ipv6-vpc-rules"
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.foo.id

  ingress {
    protocol        = 6
    rule_no         = 1
    action          = "allow"
    ipv6_cidr_block = "2600:1f16:d1e:9a00::/56"
    from_port       = 0
    to_port         = 22
  }

  tags = {
    Name = "tf-acc-acl-ipv6"
  }
}
`

const testAccNetworkACLIngressConfig = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-network-acl-ingress"
  }
}

resource "aws_subnet" "blob" {
  cidr_block              = "10.1.1.0/24"
  vpc_id                  = aws_vpc.foo.id
  map_public_ip_on_launch = true

  tags = {
    Name = "tf-acc-network-acl-ingress"
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.foo.id

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

  subnet_ids = [aws_subnet.blob.id]

  tags = {
    Name = "tf-acc-acl-ingress"
  }
}
`

const testAccNetworkACLCaseSensitiveConfig = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-network-acl-ingress"
  }
}

resource "aws_subnet" "blob" {
  cidr_block              = "10.1.1.0/24"
  vpc_id                  = aws_vpc.foo.id
  map_public_ip_on_launch = true

  tags = {
    Name = "tf-acc-network-acl-ingress"
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.foo.id

  ingress {
    protocol   = 6
    rule_no    = 1
    action     = "Allow"
    cidr_block = "10.2.0.0/18"
    from_port  = 0
    to_port    = 22
  }

  subnet_ids = [aws_subnet.blob.id]

  tags = {
    Name = "tf-acc-acl-case-sensitive"
  }
}
`

const testAccNetworkACLIngressChangeConfig = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-network-acl-ingress"
  }
}

resource "aws_subnet" "blob" {
  cidr_block              = "10.1.1.0/24"
  vpc_id                  = aws_vpc.foo.id
  map_public_ip_on_launch = true

  tags = {
    Name = "tf-acc-network-acl-ingress"
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.foo.id

  ingress {
    protocol   = 6
    rule_no    = 1
    action     = "deny"
    cidr_block = "10.2.0.0/18"
    from_port  = 0
    to_port    = 22
  }

  subnet_ids = [aws_subnet.blob.id]

  tags = {
    Name = "tf-acc-acl-ingress"
  }
}
`

const testAccNetworkACLEgressConfig = `
resource "aws_vpc" "foo" {
  cidr_block = "10.2.0.0/16"

  tags = {
    Name = "terraform-testacc-network-acl-egress"
  }
}

resource "aws_subnet" "blob" {
  cidr_block              = "10.2.0.0/24"
  vpc_id                  = aws_vpc.foo.id
  map_public_ip_on_launch = true

  tags = {
    Name = "tf-acc-network-acl-egress"
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.foo.id

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
    foo  = "bar"
    Name = "tf-acc-acl-egress"
  }
}
`

const testAccNetworkACLEgressNIngressConfig = `
resource "aws_vpc" "foo" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = "terraform-testacc-network-acl-egress-and-ingress"
  }
}

resource "aws_subnet" "blob" {
  cidr_block              = "10.3.0.0/24"
  vpc_id                  = aws_vpc.foo.id
  map_public_ip_on_launch = true

  tags = {
    Name = "tf-acc-network-acl-egress-and-ingress"
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.foo.id

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
}
`

const testAccNetworkACLSubnetConfig = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-network-acl-subnet-change"
  }
}

resource "aws_subnet" "old" {
  cidr_block              = "10.1.111.0/24"
  vpc_id                  = aws_vpc.foo.id
  map_public_ip_on_launch = true

  tags = {
    Name = "tf-acc-network-acl-subnet-change-old"
  }
}

resource "aws_subnet" "new" {
  cidr_block              = "10.1.1.0/24"
  vpc_id                  = aws_vpc.foo.id
  map_public_ip_on_launch = true

  tags = {
    Name = "tf-acc-network-acl-subnet-change-new"
  }
}

resource "aws_network_acl" "roll" {
  vpc_id     = aws_vpc.foo.id
  subnet_ids = [aws_subnet.new.id]

  tags = {
    Name = "tf-acc-acl-subnet-change-roll"
  }
}

resource "aws_network_acl" "test" {
  vpc_id     = aws_vpc.foo.id
  subnet_ids = [aws_subnet.old.id]

  tags = {
    Name = "tf-acc-acl-subnet-change-test"
  }
}
`

const testAccNetworkACLSubnetChangeConfig = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-network-acl-subnet-change"
  }
}

resource "aws_subnet" "old" {
  cidr_block              = "10.1.111.0/24"
  vpc_id                  = aws_vpc.foo.id
  map_public_ip_on_launch = true

  tags = {
    Name = "tf-acc-network-acl-subnet-change-old"
  }
}

resource "aws_subnet" "new" {
  cidr_block              = "10.1.1.0/24"
  vpc_id                  = aws_vpc.foo.id
  map_public_ip_on_launch = true

  tags = {
    Name = "tf-acc-network-acl-subnet-change-new"
  }
}

resource "aws_network_acl" "test" {
  vpc_id     = aws_vpc.foo.id
  subnet_ids = [aws_subnet.new.id]

  tags = {
    Name = "tf-acc-acl-subnet-change-test"
  }
}
`

const testAccNetworkACLSubnet_SubnetIDs = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-network-acl-subnet-ids"
  }
}

resource "aws_subnet" "one" {
  cidr_block = "10.1.111.0/24"
  vpc_id     = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-network-acl-subnet-ids-one"
  }
}

resource "aws_subnet" "two" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-network-acl-subnet-ids-two"
  }
}

resource "aws_network_acl" "test" {
  vpc_id     = aws_vpc.foo.id
  subnet_ids = [aws_subnet.one.id, aws_subnet.two.id]

  tags = {
    Name = "tf-acc-acl-subnet-ids"
  }
}
`

const testAccNetworkACLSubnet_SubnetIdsUpdate = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-network-acl-subnet-ids"
  }
}

resource "aws_subnet" "one" {
  cidr_block = "10.1.111.0/24"
  vpc_id     = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-network-acl-subnet-ids-one"
  }
}

resource "aws_subnet" "two" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-network-acl-subnet-ids-two"
  }
}

resource "aws_subnet" "three" {
  cidr_block = "10.1.222.0/24"
  vpc_id     = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-network-acl-subnet-ids-three"
  }
}

resource "aws_subnet" "four" {
  cidr_block = "10.1.4.0/24"
  vpc_id     = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-network-acl-subnet-ids-four"
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.foo.id
  subnet_ids = [
    aws_subnet.one.id,
    aws_subnet.three.id,
    aws_subnet.four.id,
  ]

  tags = {
    Name = "tf-acc-acl-subnet-ids"
  }
}
`

const testAccNetworkACLSubnet_SubnetIdsDeleteOne = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-network-acl-subnet-ids"
  }
}

resource "aws_subnet" "one" {
  cidr_block = "10.1.111.0/24"
  vpc_id     = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-network-acl-subnet-ids-one"
  }
}

resource "aws_network_acl" "test" {
  vpc_id     = aws_vpc.foo.id
  subnet_ids = [aws_subnet.one.id]

  tags = {
    Name = "tf-acc-acl-subnet-ids"
  }
}
`

const testAccNetworkACLEsp = `
resource "aws_vpc" "testvpc" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-network-acl-esp"
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.testvpc.id

  egress {
    protocol   = "esp"
    rule_no    = 5
    action     = "allow"
    cidr_block = "10.3.0.0/18"
    from_port  = 0
    to_port    = 0
  }

  tags = {
    Name = "tf-acc-acl-esp"
  }
}
`

func testAccNetworkACLEgressModeBlocksConfig() string {
	return `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-network-acl-egress-computed-attribute-mode"
  }
}

resource "aws_network_acl" "test" {
  tags = {
    Name = "terraform-testacc-network-acl-egress-computed-attribute-mode"
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
`
}

func testAccNetworkACLEgressModeNoBlocksConfig() string {
	return `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-network-acl-egress-computed-attribute-mode"
  }
}

resource "aws_network_acl" "test" {
  tags = {
    Name = "terraform-testacc-network-acl-egress-computed-attribute-mode"
  }

  vpc_id = aws_vpc.test.id
}
`
}

func testAccNetworkACLEgressModeZeroedConfig() string {
	return `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-network-acl-egress-computed-attribute-mode"
  }
}

resource "aws_network_acl" "test" {
  egress = []

  tags = {
    Name = "terraform-testacc-network-acl-egress-computed-attribute-mode"
  }

  vpc_id = aws_vpc.test.id
}
`
}

func testAccNetworkACLIngressModeBlocksConfig() string {
	return `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-network-acl-ingress-computed-attribute-mode"
  }
}

resource "aws_network_acl" "test" {
  tags = {
    Name = "terraform-testacc-network-acl-ingress-computed-attribute-mode"
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
`
}

func testAccNetworkACLIngressModeNoBlocksConfig() string {
	return `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-network-acl-ingress-computed-attribute-mode"
  }
}

resource "aws_network_acl" "test" {
  tags = {
    Name = "terraform-testacc-network-acl-ingress-computed-attribute-mode"
  }

  vpc_id = aws_vpc.test.id
}
`
}

func testAccNetworkACLIngressModeZeroedConfig() string {
	return `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-network-acl-ingress-computed-attribute-mode"
  }
}

resource "aws_network_acl" "test" {
  ingress = []

  tags = {
    Name = "terraform-testacc-network-acl-ingress-computed-attribute-mode"
  }

  vpc_id = aws_vpc.test.id
}
`
}

func testAccNetworkACLTags1Config(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id  = aws_vpc.test.id
  ingress = []

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccNetworkACLTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id  = aws_vpc.test.id
  ingress = []

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
