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

func TestAccVPCSecurityGroupAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var securityGroupVpcAssociation types.SecurityGroupVpcAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCSecurityGroupAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCSecurityGroupAssociationExists(ctx, resourceName, &securityGroupVpcAssociation),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateIdFunc:                    testAccVPCSecurityGroupAssociationImportStateIDFunc(resourceName),
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrVPCID,
			},
		},
	})
}

func TestAccVPCSecurityGroupAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var securityGroupVpcAssociation types.SecurityGroupVpcAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCSecurityGroupAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCSecurityGroupAssociationExists(ctx, resourceName, &securityGroupVpcAssociation),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPCSecurityGroupAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVPCSecurityGroupAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_security_group_association" {
				continue
			}

			_, err := tfec2.FindVPCSecurityGroupAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrSecurityGroupID], rs.Primary.Attributes[names.AttrVPCID])
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNameVPCSecurityGroupAssociation, rs.Primary.Attributes[names.AttrSecurityGroupID], err)
			}

			return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNameVPCSecurityGroupAssociation, rs.Primary.Attributes[names.AttrSecurityGroupID], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccVPCSecurityGroupAssociationImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		return fmt.Sprintf("%s:%s", rs.Primary.Attributes[names.AttrSecurityGroupID], rs.Primary.Attributes[names.AttrVPCID]), nil
	}
}

func testAccCheckVPCSecurityGroupAssociationExists(ctx context.Context, name string, VPCSecurityGroupassociation *types.SecurityGroupVpcAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVPCSecurityGroupAssociation, name, errors.New("not found"))
		}

		if rs.Primary.Attributes[names.AttrSecurityGroupID] == "" || rs.Primary.Attributes[names.AttrVPCID] == "" {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVPCSecurityGroupAssociation, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		resp, err := tfec2.FindVPCSecurityGroupAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrSecurityGroupID], rs.Primary.Attributes[names.AttrVPCID])
		if err != nil {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVPCSecurityGroupAssociation, rs.Primary.Attributes[names.AttrSecurityGroupID], err)
		}

		*VPCSecurityGroupassociation = *resp

		return nil
	}
}

func testAccVPCSecurityGroupAssociationConfig_basic(rName string) string {
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

resource "aws_security_group" "foo" {
  name   = %[1]q
  vpc_id = aws_vpc.foo.id
}

resource "aws_vpc_security_group_association" "test" {
  security_group_id = aws_security_group.foo.id
  vpc_id            = aws_vpc.bar.id
}
`, rName)
}
