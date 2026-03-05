// Copyright (c) HashiCorp, Inc.
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
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	//var associateDisassociateiamrole odbtypes.IamRole
	//rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_odb_associate_disassociate_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		//CheckDestroy:             testAccCheckAssociateDisassociateIAMRoleDestroy(ctx),
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

/*func testAccCheckAssociateDisassociateIAMRoleExists(ctx context.Context, name string, associatedisassociateiamrole *odbtypes.IamRole) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.AssociateDisassociateIAMRoleDataSource, name, errors.New("not found"))
		}

		comboID := rs.Primary.Attributes["iam_role_resource_combined_arn"]
		fmt.Println(comboID)

		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

		resp, err := tfodb.FindAssociatedDisassociatedIAMRoleOracleDBDataSource(ctx, conn, nil, nil)
		if err != nil {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.AssociateDisassociateIAMRoleDataSource, rs.Primary.ID, err)
		}

		*associatedisassociateiamrole = *resp

		return nil
	}
}*/

func (iamRoleAssociationDisassociationDSTest) testAccAssociateDisassociateIAMRoleDataSourceAutonomousCloudVmClusterConfig() string {
	return fmt.Sprintf(`

data "aws_odb_associate_disassociate_iam_role" "test" {
  combined_arn{
 	iam_role_arn = "arn:aws:iam::711387093194:role/OracleDBKMS_avmc_wxdhmnurzo"
 	resource_arn = "arn:aws:odb:us-east-1:711387093194:cloud-autonomous-vm-cluster/avmc_wxdhmnurzo"
  }
}

`)
}
