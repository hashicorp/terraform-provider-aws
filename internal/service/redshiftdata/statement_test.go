// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftdata_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/redshiftdata"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshiftdata "github.com/hashicorp/terraform-provider-aws/internal/service/redshiftdata"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftDataStatement_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v redshiftdata.DescribeStatementOutput
	resourceName := "aws_redshiftdata_statement.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftDataServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccStatementConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStatementExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClusterIdentifier, "aws_redshift_cluster.test", names.AttrClusterIdentifier),
					resource.TestCheckResourceAttr(resourceName, "parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sql", "CREATE GROUP group_name;"),
					resource.TestCheckResourceAttr(resourceName, "workgroup_name", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrDatabase, "db_user"},
			},
		},
	})
}

func TestAccRedshiftDataStatement_workgroup(t *testing.T) {
	ctx := acctest.Context(t)
	var v redshiftdata.DescribeStatementOutput
	resourceName := "aws_redshiftdata_statement.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftDataServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccStatementConfig_workgroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStatementExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrClusterIdentifier, ""),
					resource.TestCheckResourceAttr(resourceName, "parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sql", "CREATE GROUP group_name;"),
					resource.TestCheckResourceAttrPair(resourceName, "workgroup_name", "aws_redshiftserverless_workgroup.test", "workgroup_name"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrDatabase, "db_user"},
			},
		},
	})
}

func testAccCheckStatementExists(ctx context.Context, n string, v *redshiftdata.DescribeStatementOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftDataClient(ctx)

		output, err := tfredshiftdata.FindStatementByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccStatementConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
}

resource "aws_redshiftdata_statement" "test" {
  cluster_identifier = aws_redshift_cluster.test.cluster_identifier
  database           = aws_redshift_cluster.test.database_name
  db_user            = aws_redshift_cluster.test.master_username
  sql                = "CREATE GROUP group_name;"
}
`, rName))
}

func testAccStatementConfig_workgroup(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
}

resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name = %[1]q
}

resource "aws_redshiftdata_statement" "test" {
  workgroup_name = aws_redshiftserverless_workgroup.test.workgroup_name
  database       = "dev"
  sql            = "CREATE GROUP group_name;"
}
`, rName)
}
