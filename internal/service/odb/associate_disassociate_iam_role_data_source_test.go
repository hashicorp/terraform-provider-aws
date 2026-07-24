// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type iamRoleAssociationDisassociationDSTest struct {
}

var iamRoleAssociationDisassociationDSTestEntity = iamRoleAssociationDisassociationDSTest{}

func TestAccODBAssociateDisassociateIAMRoleDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	dataSourceName := "data.aws_odb_associate_disassociate_iam_role.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: iamRoleAssociationDisassociationDSTestEntity.testAccAssociateDisassociateIAMRoleDataSourceAutonomousCloudVmClusterConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrStatus, "CONNECTED"),
				),
			},
		},
	})
}

func TestAccODBAssociateDisassociateIAMRoleDataSource_avmc(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	dataSourceName := "data.aws_odb_associate_disassociate_iam_role.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: iamRoleAssociationDisassociationDSTestEntity.testAccAssociateDisassociateIAMRoleDataSourceAutonomousCloudVmClusterConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrStatus, "CONNECTED"),
				),
			},
		},
	})
}

func TestAccODBAssociateDisassociateIAMRoleDataSource_vmc(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	dataSourceName := "data.aws_odb_associate_disassociate_iam_role.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: iamRoleAssociationDisassociationDSTestEntity.testAccAssociateDisassociateIAMRoleDataSourceCloudVmClusterConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrStatus, "CONNECTED"),
				),
			},
		},
	})
}

func (iamRoleAssociationDisassociationDSTest) testAccAssociateDisassociateIAMRoleDataSourceAutonomousCloudVmClusterConfig() string {
	return `
data "aws_iam_role" "test" {
  name = "OracleDBKMS_avmc_r1a9jpx43r"
}

data "aws_odb_cloud_autonomous_vm_cluster" "test" {
  id = "avmc_r1a9jpx43r"
}


data "aws_odb_associate_disassociate_iam_role" "test" {
  composite_arn {
    iam_role_arn = data.aws_iam_role.test.arn
    resource_arn = data.aws_odb_cloud_autonomous_vm_cluster.test.arn
  }
}


`
}

func (iamRoleAssociationDisassociationDSTest) testAccAssociateDisassociateIAMRoleDataSourceCloudVmClusterConfig() string {
	return `
data "aws_iam_role" "test" {
  name = "OracleDBKMS_vmc_fh3d42fmeu"
}

data "aws_odb_cloud_vm_cluster" "test" {
  id = "vmc_fh3d42fmeu"
}


data "aws_odb_associate_disassociate_iam_role" "test" {
  composite_arn {
    iam_role_arn = data.aws_iam_role.test.arn
    resource_arn = data.aws_odb_cloud_vm_cluster.test.arn
  }
}
`
}
