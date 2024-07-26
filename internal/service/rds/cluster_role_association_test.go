// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSClusterRoleAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var dbClusterRole types.DBClusterRole
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbClusterResourceName := "aws_rds_cluster.test"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_rds_cluster_role_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterRoleAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterRoleAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterRoleAssociationExists(ctx, resourceName, &dbClusterRole),
					resource.TestCheckResourceAttrPair(resourceName, "db_cluster_identifier", dbClusterResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "feature_name", "s3Import"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, iamRoleResourceName, names.AttrARN),
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

func TestAccRDSClusterRoleAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var dbClusterRole types.DBClusterRole
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster_role_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterRoleAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterRoleAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterRoleAssociationExists(ctx, resourceName, &dbClusterRole),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfrds.ResourceClusterRoleAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRDSClusterRoleAssociation_Disappears_cluster(t *testing.T) {
	ctx := acctest.Context(t)
	var dbClusterRole types.DBClusterRole
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster_role_association.test"
	clusterResourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterRoleAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterRoleAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterRoleAssociationExists(ctx, resourceName, &dbClusterRole),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfrds.ResourceCluster(), clusterResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRDSClusterRoleAssociation_Disappears_role(t *testing.T) {
	ctx := acctest.Context(t)
	var dbClusterRole types.DBClusterRole
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster_role_association.test"
	roleResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterRoleAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterRoleAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterRoleAssociationExists(ctx, resourceName, &dbClusterRole),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiam.ResourceRole(), roleResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckClusterRoleAssociationExists(ctx context.Context, resourceName string, v *types.DBClusterRole) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		output, err := tfrds.FindDBClusterRoleByTwoPartKey(ctx, conn, rs.Primary.Attributes["db_cluster_identifier"], rs.Primary.Attributes[names.AttrRoleARN])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckClusterRoleAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_db_cluster_role_association" {
				continue
			}

			_, err := tfrds.FindDBClusterRoleByTwoPartKey(ctx, conn, rs.Primary.Attributes["db_cluster_identifier"], rs.Primary.Attributes[names.AttrRoleARN])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS DB Cluster IAM Role Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccClusterRoleAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_rds_cluster_role_association" "test" {
  db_cluster_identifier = aws_rds_cluster.test.id
  feature_name          = "s3Import"
  role_arn              = aws_iam_role.test.arn
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = "aurora-postgresql"
  availability_zones  = [data.aws_availability_zones.available.names[0], data.aws_availability_zones.available.names[1], data.aws_availability_zones.available.names[2]]
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "foobarfoobarfoobar"
  skip_final_snapshot = true
}

resource "aws_iam_role" "test" {
  assume_role_policy = data.aws_iam_policy_document.rds_assume_role_policy.json
  name               = %[1]q

  # ensure IAM role is created just before association to exercise IAM eventual consistency
  depends_on = [aws_rds_cluster.test]
}

data "aws_iam_policy_document" "rds_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"

    principals {
      identifiers = ["rds.amazonaws.com"]
      type        = "Service"
    }
  }
}
`, rName))
}
