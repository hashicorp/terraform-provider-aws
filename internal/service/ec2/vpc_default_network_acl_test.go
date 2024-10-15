// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCDefaultNetworkACL_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.NetworkAcl
	resourceName := "aws_default_network_acl.test"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultNetworkACLConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultNetworkACLExists(ctx, resourceName, &v),
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

func TestAccVPCDefaultNetworkACL_basicIPv6VPC(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.NetworkAcl
	resourceName := "aws_default_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultNetworkACLConfig_ipv6(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultNetworkACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "egress.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct0),
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

func TestAccVPCDefaultNetworkACL_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.NetworkAcl
	resourceName := "aws_default_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultNetworkACLConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultNetworkACLExists(ctx, resourceName, &v),
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
				Config: testAccVPCDefaultNetworkACLConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultNetworkACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccVPCDefaultNetworkACLConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultNetworkACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccVPCDefaultNetworkACL_Deny_ingress(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.NetworkAcl
	resourceName := "aws_default_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultNetworkACLConfig_denyIngress(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultNetworkACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "egress.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						names.AttrProtocol:  "-1",
						"rule_no":           "100",
						"from_port":         acctest.Ct0,
						"to_port":           acctest.Ct0,
						names.AttrAction:    "allow",
						names.AttrCIDRBlock: "0.0.0.0/0",
					}),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccVPCDefaultNetworkACL_withIPv6Ingress(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.NetworkAcl
	resourceName := "aws_default_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultNetworkACLConfig_includingIPv6Rule(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultNetworkACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "egress.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						names.AttrProtocol: "-1",
						"rule_no":          "101",
						"from_port":        acctest.Ct0,
						"to_port":          acctest.Ct0,
						names.AttrAction:   "allow",
						"ipv6_cidr_block":  "::/0",
					}),
				),
			},
		},
	})
}

func TestAccVPCDefaultNetworkACL_subnetRemoval(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.NetworkAcl
	resourceName := "aws_default_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultNetworkACLConfig_subnets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultNetworkACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test1", names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test2", names.AttrID),
				),
			},
			// Here the Subnets have been removed from the Default Network ACL Config,
			// but have not been reassigned. The result is that the Subnets are still
			// there, and we have a non-empty plan
			{
				Config: testAccVPCDefaultNetworkACLConfig_subnetsRemove(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultNetworkACLExists(ctx, resourceName, &v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCDefaultNetworkACL_subnetReassign(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.NetworkAcl
	resourceName := "aws_default_network_acl.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultNetworkACLConfig_subnets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultNetworkACLExists(ctx, resourceName, &v),
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
			// Here we've reassigned the subnets to a different ACL.
			// Without any otherwise association between the `aws_network_acl` and
			// `aws_default_network_acl` resources, we cannot guarantee that the
			// reassignment of the two subnets to the `aws_network_acl` will happen
			// before the update/read on the `aws_default_network_acl` resource.
			// Because of this, there could be a non-empty plan if a READ is done on
			// the default before the reassignment occurs on the other resource.
			//
			// For the sake of testing, here we introduce a depends_on attribute from
			// the default resource to the other acl resource, to ensure the latter's
			// update occurs first, and the former's READ will correctly read zero
			// subnets
			{
				Config: testAccVPCDefaultNetworkACLConfig_subnetsMove(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultNetworkACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", acctest.Ct0),
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

func testAccCheckDefaultNetworkACLDestroy(s *terraform.State) error {
	// The default NACL is not deleted.
	return nil
}

func testAccCheckDefaultNetworkACLExists(ctx context.Context, n string, v *types.NetworkAcl) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Default Network ACL ID is set: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindNetworkACLByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if !aws.ToBool(output.IsDefault) {
			return fmt.Errorf("EC2 Network ACL %s is not a default NACL", rs.Primary.ID)
		}

		*v = *output

		return nil
	}
}

func testAccVPCDefaultNetworkACLConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_network_acl" "test" {
  default_network_acl_id = aws_vpc.test.default_network_acl_id
}
`, rName)
}

func testAccVPCDefaultNetworkACLConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_network_acl" "test" {
  default_network_acl_id = aws_vpc.test.default_network_acl_id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccVPCDefaultNetworkACLConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_network_acl" "test" {
  default_network_acl_id = aws_vpc.test.default_network_acl_id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccVPCDefaultNetworkACLConfig_includingIPv6Rule(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_network_acl" "test" {
  default_network_acl_id = aws_vpc.test.default_network_acl_id

  ingress {
    protocol        = -1
    rule_no         = 101
    action          = "allow"
    ipv6_cidr_block = "::/0"
    from_port       = 0
    to_port         = 0
  }
}
`, rName)
}

func testAccVPCDefaultNetworkACLConfig_denyIngress(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_network_acl" "test" {
  default_network_acl_id = aws_vpc.test.default_network_acl_id

  egress {
    protocol   = -1
    rule_no    = 100
    action     = "allow"
    cidr_block = "0.0.0.0/0"
    from_port  = 0
    to_port    = 0
  }
}
`, rName)
}

func testAccDefaultNetworkACLSubnetsBaseConfig(rName string) string {
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
`, rName)
}

func testAccVPCDefaultNetworkACLConfig_subnets(rName string) string {
	return acctest.ConfigCompose(testAccDefaultNetworkACLSubnetsBaseConfig(rName), fmt.Sprintf(`
resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_network_acl" "test" {
  default_network_acl_id = aws_vpc.test.default_network_acl_id

  subnet_ids = [aws_subnet.test1.id, aws_subnet.test2.id]
}
`, rName))
}

func testAccVPCDefaultNetworkACLConfig_subnetsRemove(rName string) string {
	return acctest.ConfigCompose(testAccDefaultNetworkACLSubnetsBaseConfig(rName), fmt.Sprintf(`
resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_network_acl" "test" {
  default_network_acl_id = aws_vpc.test.default_network_acl_id

  depends_on = [aws_network_acl.test]
}
`, rName))
}

func testAccVPCDefaultNetworkACLConfig_subnetsMove(rName string) string {
	return acctest.ConfigCompose(testAccDefaultNetworkACLSubnetsBaseConfig(rName), fmt.Sprintf(`
resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  subnet_ids = [aws_subnet.test1.id, aws_subnet.test2.id]

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_network_acl" "test" {
  default_network_acl_id = aws_vpc.test.default_network_acl_id

  depends_on = [aws_network_acl.test]
}
`, rName))
}

func testAccVPCDefaultNetworkACLConfig_ipv6(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_network_acl" "test" {
  default_network_acl_id = aws_vpc.test.default_network_acl_id
}
`, rName)
}
