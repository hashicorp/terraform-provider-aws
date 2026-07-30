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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfodb "github.com/hashicorp/terraform-provider-aws/internal/service/odb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type iamRoleAssociationDisassociationTest struct {
}

var iamRoleAssociationDisassociationTestEntity = iamRoleAssociationDisassociationTest{}

const (
	iamRoleAssociationAVMCIDEnvVar      = "TF_AWS_ODB_IAM_ROLE_ASSOCIATION_AVMC_ID"
	iamRoleAssociationAVMCRoleARNEnvVar = "TF_AWS_ODB_IAM_ROLE_ASSOCIATION_AVMC_ROLE_ARN"
	iamRoleAssociationVMCIDEnvVar       = "TF_AWS_ODB_IAM_ROLE_ASSOCIATION_VMC_ID"
	iamRoleAssociationVMCRoleARNEnvVar  = "TF_AWS_ODB_IAM_ROLE_ASSOCIATION_VMC_ROLE_ARN"
)

type iamRoleAssociationTestFixtures struct {
	iamRoleARN string
	resourceID string
}

func testAccIAMRoleAssociationAVMCFixtures(t *testing.T) iamRoleAssociationTestFixtures {
	t.Helper()

	return iamRoleAssociationTestFixtures{
		resourceID: acctest.SkipIfEnvVarNotSet(t, iamRoleAssociationAVMCIDEnvVar),
		iamRoleARN: acctest.SkipIfEnvVarNotSet(t, iamRoleAssociationAVMCRoleARNEnvVar),
	}
}

func testAccIAMRoleAssociationVMCFixtures(t *testing.T) iamRoleAssociationTestFixtures {
	t.Helper()

	return iamRoleAssociationTestFixtures{
		resourceID: acctest.SkipIfEnvVarNotSet(t, iamRoleAssociationVMCIDEnvVar),
		iamRoleARN: acctest.SkipIfEnvVarNotSet(t, iamRoleAssociationVMCRoleARNEnvVar),
	}
}

func TestAccODBAssociateDisassociateIAMRole_vmc(t *testing.T) {
	fixtures := testAccIAMRoleAssociationVMCFixtures(t)
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var associateDisassociateIAMRole odbtypes.IamRole
	resourceName := "aws_odb_iam_role_association.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			autonomousVMClusterResourceTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociateDisassociateIAMRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: iamRoleAssociationDisassociationTestEntity.associateIAMRoleToCloudVMCluster(fixtures),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssociateDisassociateIAMRoleExists(ctx, resourceName, &associateDisassociateIAMRole),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrResourceARN,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					rs, ok := state.RootModule().Resources[resourceName]
					if !ok {
						return "", errors.New("resource not found in state")
					}

					iamRoleARN, ok := rs.Primary.Attributes[names.AttrIAMRoleARN]
					if !ok || iamRoleARN == "" {
						return "", errors.New("missing iam_role_arn in state")
					}

					resourceARN, ok := rs.Primary.Attributes[names.AttrResourceARN]
					if !ok || resourceARN == "" {
						return "", errors.New("missing resource_arn in state")
					}

					return iamRoleARN + "," + resourceARN, nil
				},
			},
		},
	})
}

func TestAccODBAssociateDisassociateIAMRole_avmc(t *testing.T) {
	fixtures := testAccIAMRoleAssociationAVMCFixtures(t)
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var associateDisassociateIAMRole odbtypes.IamRole
	resourceName := "aws_odb_iam_role_association.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			autonomousVMClusterResourceTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociateDisassociateIAMRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: iamRoleAssociationDisassociationTestEntity.associateIAMRoleToAutonomousCloudVMCluster(fixtures),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssociateDisassociateIAMRoleExists(ctx, resourceName, &associateDisassociateIAMRole),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrResourceARN,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					rs, ok := state.RootModule().Resources[resourceName]
					if !ok {
						return "", errors.New("resource not found in state")
					}

					iamRoleARN, ok := rs.Primary.Attributes[names.AttrIAMRoleARN]
					if !ok || iamRoleARN == "" {
						return "", errors.New("missing iam_role_arn in state")
					}

					resourceARN, ok := rs.Primary.Attributes[names.AttrResourceARN]
					if !ok || resourceARN == "" {
						return "", errors.New("missing resource_arn in state")
					}

					return iamRoleARN + "," + resourceARN, nil
				},
			},
		},
	})
}

func TestAccODBAssociateDisassociateIAMRole_disappears(t *testing.T) {
	fixtures := testAccIAMRoleAssociationAVMCFixtures(t)
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var iamRole odbtypes.IamRole
	resourceName := "aws_odb_iam_role_association.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			autonomousVMClusterResourceTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociateDisassociateIAMRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: iamRoleAssociationDisassociationTestEntity.associateIAMRoleToAutonomousCloudVMCluster(fixtures),
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
			if rs.Type != "aws_odb_iam_role_association" {
				continue
			}

			resourceARN, ok := rs.Primary.Attributes[names.AttrResourceARN]
			if !ok || resourceARN == "" {
				return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.ResNameAssociateDisassociateIAMRole, rs.Primary.ID, errors.New("resource ARN not found in state"))
			}

			iamRoleARN, ok := rs.Primary.Attributes[names.AttrIAMRoleARN]
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

		resourceARN, ok := rs.Primary.Attributes[names.AttrResourceARN]
		if !ok || resourceARN == "" {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameAssociateDisassociateIAMRole, rs.Primary.ID, errors.New("resource ARN not found in state"))
		}

		iamRoleARN, ok := rs.Primary.Attributes[names.AttrIAMRoleARN]
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

func (test iamRoleAssociationDisassociationTest) associateIAMRoleToAutonomousCloudVMCluster(fixtures iamRoleAssociationTestFixtures) string {
	return fmt.Sprintf(`
data "aws_odb_cloud_autonomous_vm_cluster" "test" {
  id = %[1]q
}

resource "aws_odb_iam_role_association" "test" {
  aws_integration = "KmsTde"
  iam_role_arn    = %[2]q
  resource_arn    = data.aws_odb_cloud_autonomous_vm_cluster.test.arn
}
`, fixtures.resourceID, fixtures.iamRoleARN)
}

func (test iamRoleAssociationDisassociationTest) associateIAMRoleToCloudVMCluster(fixtures iamRoleAssociationTestFixtures) string {
	return fmt.Sprintf(`
data "aws_odb_cloud_vm_cluster" "test" {
  id = %[1]q
}

resource "aws_odb_iam_role_association" "test" {
  aws_integration = "KmsTde"
  iam_role_arn    = %[2]q
  resource_arn    = data.aws_odb_cloud_vm_cluster.test.arn
}
`, fixtures.resourceID, fixtures.iamRoleARN)
}
