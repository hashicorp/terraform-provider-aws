// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfodb "github.com/hashicorp/terraform-provider-aws/internal/service/odb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type iamRoleAssociationDisassociationTest struct {
}

var iamRoleAssociationDisassociationTestEntity = iamRoleAssociationDisassociationTest{}

func TestAccODBAssociateDisassociateIAMRole_vmc(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var associatedisassociateiamrole odbtypes.IamRole
	//rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_odb_associate_disassociate_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			//testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociateDisassociateIAMRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: iamRoleAssociationDisassociationTestEntity.associateIAMRoleToCloudVMCluster(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssociateDisassociateIAMRoleExists(ctx, resourceName, &associatedisassociateiamrole),
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

func TestAccODBAssociateDisassociateIAMRole_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var iamRole odbtypes.IamRole
	//rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_odb_associate_disassociate_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			//testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociateDisassociateIAMRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: iamRoleAssociationDisassociationTestEntity.associateIAMRoleToCloudVMCluster(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssociateDisassociateIAMRoleExists(ctx, resourceName, &iamRole),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfodb.AssociateDisassociateIAMRole, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckAssociateDisassociateIAMRoleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_odb_associate_disassociate_iam_role" {
				continue
			}

			comboID := rs.Primary.Attributes["iam_role_resource_combined_arn"]
			fmt.Println(comboID)
			_, err := tfodb.FindAssociatedDisassociatedIAMRoleOracleDBResource(ctx, conn, nil, nil)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.ResNameAssociateDisassociateIAMRole, rs.Primary.ID, err)
			}

			return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.ResNameAssociateDisassociateIAMRole, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAssociateDisassociateIAMRoleExists(ctx context.Context, name string, associatedisassociateiamrole *odbtypes.IamRole) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameAssociateDisassociateIAMRole, name, errors.New("not found"))
		}

		comboID := rs.Primary.Attributes["iam_role_resource_combined_arn"]
		fmt.Println(comboID)

		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

		resp, err := tfodb.FindAssociatedDisassociatedIAMRoleOracleDBResource(ctx, conn, nil, nil)
		if err != nil {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameAssociateDisassociateIAMRole, rs.Primary.ID, err)
		}

		*associatedisassociateiamrole = *resp

		return nil
	}
}

/*func testAccCheckAssociateDisassociateIAMRoleNotRecreated(before, after *odb.DescribeAssociateDisassociateIAMRoleResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.AssociateDisassociateIAMRoleId), aws.ToString(after.AssociateDisassociateIAMRoleId); before != after {
			return create.Error(names.ODB, create.ErrActionCheckingNotRecreated, tfodb.ResNameAssociateDisassociateIAMRole, aws.ToString(before.AssociateDisassociateIAMRoleId), errors.New("recreated"))
		}

		return nil
	}
}*/

func (test iamRoleAssociationDisassociationTest) associateIAMRoleToCloudVMCluster() string {
	return fmt.Sprintf(`
data "aws_odb_cloud_vm_cluster" "test" {
  tags = {
    env = "tf-test"
  }
}

data "aws_iam_role" "test" {
  name = "my-role-name"
}

resource "aws_odb_associate_disassociate_iam_role" "test" {
  aws_integration = "KmsTde"
  composite_arn {
    iam_role_arn = data.aws_iam_role.arn
    resource_arn = aws_odb_cloud_vm_cluster.test.arn
  }
}
`)
}
