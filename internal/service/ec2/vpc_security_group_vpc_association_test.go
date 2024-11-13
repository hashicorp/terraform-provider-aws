// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCSecurityGroupVPCAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var securityGroupVpcAssociation types.SecurityGroupVpcAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_vpc_association.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCSecurityGroupVPCAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupVPCAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCSecurityGroupVPCAssociationExists(ctx, resourceName, &securityGroupVpcAssociation),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "state", string(types.SecurityGroupVpcAssociationStateAssociated)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateIdFunc:                    testAccVPCSecurityGroupVPCAssociationImportStateIDFunc(resourceName),
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrVPCID,
			},
		},
	})
}

func TestAccVPCSecurityGroupVPCAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var securityGroupVpcAssociation types.SecurityGroupVpcAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_vpc_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCSecurityGroupVPCAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupVPCAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCSecurityGroupVPCAssociationExists(ctx, resourceName, &securityGroupVpcAssociation),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfec2.ResourceSecurityGroupVPCAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVPCSecurityGroupVPCAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_security_group_vpc_association" {
				continue
			}

			_, err := tfec2.FindSecurityGroupVPCAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["security_group_id"], rs.Primary.Attributes[names.AttrVPCID])
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNameSecurityGroupVPCAssociation, rs.Primary.Attributes["security_group_id"], err)
			}

			return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNameSecurityGroupVPCAssociation, rs.Primary.Attributes["security_group_id"], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccVPCSecurityGroupVPCAssociationImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["security_group_id"], rs.Primary.Attributes[names.AttrVPCID]), nil
	}
}

func testAccCheckVPCSecurityGroupVPCAssociationExists(ctx context.Context, name string, VPCSecurityGroupassociation *types.SecurityGroupVpcAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameSecurityGroupVPCAssociation, name, errors.New("not found"))
		}

		if rs.Primary.Attributes["security_group_id"] == "" || rs.Primary.Attributes[names.AttrVPCID] == "" {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameSecurityGroupVPCAssociation, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		resp, err := tfec2.FindSecurityGroupVPCAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["security_group_id"], rs.Primary.Attributes[names.AttrVPCID])
		if err != nil {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameSecurityGroupVPCAssociation, rs.Primary.Attributes["security_group_id"], err)
		}

		*VPCSecurityGroupassociation = *resp

		return nil
	}
}

func testAccVPCSecurityGroupVPCAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block           = "10.6.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "bar" {
  cidr_block           = "10.7.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.foo.id
}

resource "aws_vpc_security_group_vpc_association" "test" {
  security_group_id = aws_security_group.test.id
  vpc_id            = aws_vpc.bar.id
}
`, rName)
}
