// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"context"
	"errors"
	"testing"

	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfodb "github.com/hashicorp/terraform-provider-aws/internal/service/odb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type iamRoleAssociationDisassociationTest struct {
}

var iamRoleAssociationDisassociationTestEntity = iamRoleAssociationDisassociationTest{}

func TestAccODBAssociateDisassociateIAMRole_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var associateDisassociateIAMRole odbtypes.IamRole
	resourceName := "aws_odb_associate_disassociate_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			autonomousVMClusterResourceTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociateDisassociateIAMRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: iamRoleAssociationDisassociationTestEntity.associateIAMRoleToAutonomousCloudVMCluster(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssociateDisassociateIAMRoleExists(ctx, resourceName, &associateDisassociateIAMRole),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "composite_arn",
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					rs, ok := state.RootModule().Resources[resourceName]
					if !ok {
						return "", errors.New("resource not found in state")
					}

					iamRoleARN, ok := rs.Primary.Attributes["composite_arn.0.iam_role_arn"]
					if !ok || iamRoleARN == "" {
						return "", errors.New("missing composite_arn.0.iam_role_arn in state")
					}

					resourceARN, ok := rs.Primary.Attributes["composite_arn.0.resource_arn"]
					if !ok || resourceARN == "" {
						return "", errors.New("missing composite_arn.0.resource_arn in state")
					}

					return "iam_role_arn=" + iamRoleARN + ",resource_arn=" + resourceARN, nil
				},
			},
		},
	})
}

func TestAccODBAssociateDisassociateIAMRole_avmc(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var associateDisassociateIAMRole odbtypes.IamRole
	resourceName := "aws_odb_associate_disassociate_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			autonomousVMClusterResourceTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociateDisassociateIAMRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: iamRoleAssociationDisassociationTestEntity.associateIAMRoleToAutonomousCloudVMCluster(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssociateDisassociateIAMRoleExists(ctx, resourceName, &associateDisassociateIAMRole),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "composite_arn",
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					rs, ok := state.RootModule().Resources[resourceName]
					if !ok {
						return "", errors.New("resource not found in state")
					}

					iamRoleARN, ok := rs.Primary.Attributes["composite_arn.0.iam_role_arn"]
					if !ok || iamRoleARN == "" {
						return "", errors.New("missing composite_arn.0.iam_role_arn in state")
					}

					resourceARN, ok := rs.Primary.Attributes["composite_arn.0.resource_arn"]
					if !ok || resourceARN == "" {
						return "", errors.New("missing composite_arn.0.resource_arn in state")
					}

					return "iam_role_arn=" + iamRoleARN + ",resource_arn=" + resourceARN, nil
				},
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
	resourceName := "aws_odb_associate_disassociate_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			autonomousVMClusterResourceTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociateDisassociateIAMRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: iamRoleAssociationDisassociationTestEntity.associateIAMRoleToAutonomousCloudVMCluster(),
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

			resourceARN, ok := rs.Primary.Attributes["composite_arn.0.resource_arn"]
			if !ok || resourceARN == "" {
				return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.ResNameAssociateDisassociateIAMRole, rs.Primary.ID, errors.New("resource ARN not found in state"))
			}

			iamRoleARN, ok := rs.Primary.Attributes["composite_arn.0.iam_role_arn"]
			if !ok || iamRoleARN == "" {
				return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.ResNameAssociateDisassociateIAMRole, rs.Primary.ID, errors.New("IAM role ARN not found in state"))
			}

			_, err := tfodb.FindAssociatedDisassociatedIAMRoleOracleDBResource(ctx, conn, resourceARN, iamRoleARN)
			if retry.NotFound(err) {
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

func testAccCheckAssociateDisassociateIAMRoleExists(ctx context.Context, name string, associateDisassociateIAMRole *odbtypes.IamRole) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameAssociateDisassociateIAMRole, name, errors.New("not found"))
		}

		resourceARN, ok := rs.Primary.Attributes["composite_arn.0.resource_arn"]
		if !ok || resourceARN == "" {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameAssociateDisassociateIAMRole, rs.Primary.ID, errors.New("resource ARN not found in state"))
		}

		iamRoleARN, ok := rs.Primary.Attributes["composite_arn.0.iam_role_arn"]
		if !ok || iamRoleARN == "" {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameAssociateDisassociateIAMRole, rs.Primary.ID, errors.New("IAM role ARN not found in state"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

		resp, err := tfodb.FindAssociatedDisassociatedIAMRoleOracleDBResource(ctx, conn, resourceARN, iamRoleARN)
		if err != nil {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameAssociateDisassociateIAMRole, rs.Primary.ID, err)
		}

		*associateDisassociateIAMRole = *resp

		return nil
	}
}

func (test iamRoleAssociationDisassociationTest) associateIAMRoleToAutonomousCloudVMCluster() string {
	return `
data "aws_iam_role" "test" {
  name = "OracleDBKMS_avmc_hvlokll3j2"
}

data "aws_odb_cloud_autonomous_vm_cluster" "test" {
  id = "avmc_hvlokll3j2"
}

resource "aws_odb_associate_disassociate_iam_role" "test" {
  aws_integration = "KmsTde"
  composite_arn {
    iam_role_arn = data.aws_iam_role.test.arn
    resource_arn = data.aws_odb_cloud_autonomous_vm_cluster.test.arn
  }
}
`
}

func (test iamRoleAssociationDisassociationTest) associateIAMRoleToCloudVMCluster() string {
	return `
data "aws_iam_role" "test" {
  name = "OracleDBKMS_vmc_fh3d42fmeu"
}

data "aws_odb_cloud_vm_cluster" "test" {
  id = "vmc_fh3d42fmeu"
}

resource "aws_odb_associate_disassociate_iam_role" "test" {
  aws_integration = "KmsTde"
  composite_arn {
    iam_role_arn = data.aws_iam_role.test.arn
    resource_arn = data.aws_odb_cloud_vm_cluster.test.arn
  }
}
`
}
