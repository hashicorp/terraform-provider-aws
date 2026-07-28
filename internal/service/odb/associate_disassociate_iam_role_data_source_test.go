// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type iamRoleAssociationDisassociationDSTest struct {
}

var iamRoleAssociationDisassociationDSTestEntity = iamRoleAssociationDisassociationDSTest{}

func TestAccODBAssociateDisassociateIAMRoleDataSource_basic(t *testing.T) {
	fixtures := testAccIAMRoleAssociationAVMCFixtures(t)
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	dataSourceName := "data.aws_odb_iam_role_association.test"
	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociateDisassociateIAMRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: iamRoleAssociationDisassociationDSTestEntity.testAccAssociateDisassociateIAMRoleDataSourceAutonomousCloudVmClusterConfig(fixtures),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrStatus, "CONNECTED"),
				),
			},
		},
	})
}

func TestAccODBAssociateDisassociateIAMRoleDataSource_avmc(t *testing.T) {
	fixtures := testAccIAMRoleAssociationAVMCFixtures(t)
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	dataSourceName := "data.aws_odb_iam_role_association.test"
	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociateDisassociateIAMRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: iamRoleAssociationDisassociationDSTestEntity.testAccAssociateDisassociateIAMRoleDataSourceAutonomousCloudVmClusterConfig(fixtures),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrStatus, "CONNECTED"),
				),
			},
		},
	})
}

func TestAccODBAssociateDisassociateIAMRoleDataSource_vmc(t *testing.T) {
	fixtures := testAccIAMRoleAssociationVMCFixtures(t)
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	dataSourceName := "data.aws_odb_iam_role_association.test"
	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociateDisassociateIAMRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: iamRoleAssociationDisassociationDSTestEntity.testAccAssociateDisassociateIAMRoleDataSourceCloudVmClusterConfig(fixtures),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrStatus, "CONNECTED"),
				),
			},
		},
	})
}

func (iamRoleAssociationDisassociationDSTest) testAccAssociateDisassociateIAMRoleDataSourceAutonomousCloudVmClusterConfig(fixtures iamRoleAssociationTestFixtures) string {
	return fmt.Sprintf(`
data "aws_odb_cloud_autonomous_vm_cluster" "test" {
  id = %[1]q
}

resource "aws_odb_iam_role_association" "fixture" {
  aws_integration = "KmsTde"
  iam_role_arn   = %[2]q
  resource_arn   = data.aws_odb_cloud_autonomous_vm_cluster.test.arn
}

data "aws_odb_iam_role_association" "test" {
  iam_role_arn = aws_odb_iam_role_association.fixture.iam_role_arn
  resource_arn = aws_odb_iam_role_association.fixture.resource_arn
}
`, fixtures.resourceID, fixtures.iamRoleARN)
}

func (iamRoleAssociationDisassociationDSTest) testAccAssociateDisassociateIAMRoleDataSourceCloudVmClusterConfig(fixtures iamRoleAssociationTestFixtures) string {
	return fmt.Sprintf(`
data "aws_odb_cloud_vm_cluster" "test" {
  id = %[1]q
}

resource "aws_odb_iam_role_association" "fixture" {
  aws_integration = "KmsTde"
  iam_role_arn   = %[2]q
  resource_arn   = data.aws_odb_cloud_vm_cluster.test.arn
}

data "aws_odb_iam_role_association" "test" {
  iam_role_arn = aws_odb_iam_role_association.fixture.iam_role_arn
  resource_arn = aws_odb_iam_role_association.fixture.resource_arn
}
`, fixtures.resourceID, fixtures.iamRoleARN)
}
