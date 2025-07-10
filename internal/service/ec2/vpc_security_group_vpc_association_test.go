// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCSecurityGroupVPCAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var assoc awstypes.SecurityGroupVpcAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_vpc_association.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupVPCAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupVPCAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupVPCAssociationExists(ctx, resourceName, &assoc),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.SecurityGroupVpcAssociationStateAssociated)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, "aws_vpc.target", names.AttrID),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateIdFunc:                    testAccSecurityGroupVPCAssociationImportStateIDFunc(resourceName),
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrVPCID,
			},
		},
	})
}

func TestAccVPCSecurityGroupVPCAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var assoc awstypes.SecurityGroupVpcAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_vpc_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupVPCAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupVPCAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupVPCAssociationExists(ctx, resourceName, &assoc),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfec2.ResourceSecurityGroupVPCAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupVPCAssociation_disappears_SecurityGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var assoc awstypes.SecurityGroupVpcAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_vpc_association.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupVPCAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupVPCAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupVPCAssociationExists(ctx, resourceName, &assoc),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfec2.ResourceSecurityGroupVPCAssociation, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceSecurityGroup(), sgResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupVPCAssociation_disappears_VPC(t *testing.T) {
	ctx := acctest.Context(t)
	var assoc awstypes.SecurityGroupVpcAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_vpc_association.test"
	vpcResourceName := "aws_vpc.target"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupVPCAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupVPCAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupVPCAssociationExists(ctx, resourceName, &assoc),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfec2.ResourceSecurityGroupVPCAssociation, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPC(), vpcResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Resource Identity was added in v6.1
func TestAccVPCSecurityGroupVPCAssociation_Identity_ExistingResource(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.SecurityGroupVpcAssociation
	resourceName := "aws_vpc_security_group_vpc_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy: testAccCheckSecurityGroupVPCAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "6.0.0",
					},
				},
				Config: testAccVPCSecurityGroupVPCAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupVPCAssociationExists(ctx, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoIdentity(resourceName),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccVPCSecurityGroupVPCAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupVPCAssociationExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrVPCID:     knownvalue.NotNull(),
						"security_group_id": knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrVPCID)),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New("security_group_id")),
				},
			},
		},
	})
}

func testAccCheckSecurityGroupVPCAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_security_group_vpc_association" {
				continue
			}

			_, err := tfec2.FindSecurityGroupVPCAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["security_group_id"], rs.Primary.Attributes[names.AttrVPCID])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Security Group (%s) VPC (%s) Association still exists", rs.Primary.Attributes["security_group_id"], rs.Primary.Attributes[names.AttrVPCID])
		}

		return nil
	}
}

func testAccCheckSecurityGroupVPCAssociationExists(ctx context.Context, n string, v *awstypes.SecurityGroupVpcAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindSecurityGroupVPCAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["security_group_id"], rs.Primary.Attributes[names.AttrVPCID])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccSecurityGroupVPCAssociationImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["security_group_id"], rs.Primary.Attributes[names.AttrVPCID]), nil
	}
}

func testAccVPCSecurityGroupVPCAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "source" {
  cidr_block = "10.6.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "target" {
  cidr_block = "10.7.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.source.id
}

resource "aws_vpc_security_group_vpc_association" "test" {
  security_group_id = aws_security_group.test.id
  vpc_id            = aws_vpc.target.id
}
`, rName)
}
