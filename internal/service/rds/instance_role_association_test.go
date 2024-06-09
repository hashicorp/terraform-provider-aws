// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSInstanceRoleAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstanceRole1 rds.DBInstanceRole
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbInstanceResourceName := "aws_db_instance.test"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_db_instance_role_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceRoleAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceRoleAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceRoleAssociationExists(ctx, resourceName, &dbInstanceRole1),
					resource.TestCheckResourceAttrPair(resourceName, "db_instance_identifier", dbInstanceResourceName, names.AttrIdentifier),
					resource.TestCheckResourceAttr(resourceName, "feature_name", "S3_INTEGRATION"),
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

func TestAccRDSInstanceRoleAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance1 rds.DBInstance
	var dbInstanceRole1 rds.DBInstanceRole
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbInstanceResourceName := "aws_db_instance.test"
	resourceName := "aws_db_instance_role_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceRoleAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceRoleAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, dbInstanceResourceName, &dbInstance1),
					testAccCheckInstanceRoleAssociationExists(ctx, resourceName, &dbInstanceRole1),
					testAccCheckInstanceRoleAssociationDisappears(ctx, &dbInstance1, &dbInstanceRole1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckInstanceRoleAssociationExists(ctx context.Context, resourceName string, dbInstanceRole *rds.DBInstanceRole) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		dbInstanceIdentifier, roleArn, err := tfrds.InstanceRoleAssociationDecodeID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error reading resource ID: %s", err)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

		role, err := tfrds.DescribeDBInstanceRole(ctx, conn, dbInstanceIdentifier, roleArn)
		if tfresource.NotFound(err) {
			return fmt.Errorf("RDS DB Instance IAM Role Association not found")
		}
		if err != nil {
			return err
		}

		if aws.StringValue(role.Status) != "ACTIVE" {
			return fmt.Errorf("RDS DB Instance (%s) IAM Role (%s) association exists in non-ACTIVE (%s) state", dbInstanceIdentifier, roleArn, aws.StringValue(role.Status))
		}

		*dbInstanceRole = *role

		return nil
	}
}

func testAccCheckInstanceRoleAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_db_instance_role_association" {
				continue
			}

			dbInstanceIdentifier, roleArn, err := tfrds.InstanceRoleAssociationDecodeID(rs.Primary.ID)
			if err != nil {
				return fmt.Errorf("error reading resource ID: %s", err)
			}

			dbInstanceRole, err := tfrds.DescribeDBInstanceRole(ctx, conn, dbInstanceIdentifier, roleArn)
			if tfresource.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("RDS DB Instance (%s) IAM Role (%s) association still exists in non-deleted (%s) state", dbInstanceIdentifier, roleArn, aws.StringValue(dbInstanceRole.Status))
		}

		return nil
	}
}

func testAccCheckInstanceRoleAssociationDisappears(ctx context.Context, dbInstance *rds.DBInstance, dbInstanceRole *rds.DBInstanceRole) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

		input := &rds.RemoveRoleFromDBInstanceInput{
			DBInstanceIdentifier: dbInstance.DBInstanceIdentifier,
			FeatureName:          dbInstanceRole.FeatureName,
			RoleArn:              dbInstanceRole.RoleArn,
		}

		_, err := conn.RemoveRoleFromDBInstanceWithContext(ctx, input)
		if err != nil {
			return err
		}

		return tfrds.WaitForDBInstanceRoleDisassociation(ctx, conn, aws.StringValue(dbInstance.DBInstanceIdentifier), aws.StringValue(dbInstanceRole.RoleArn))
	}
}

func testAccInstanceRoleAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_db_instance_role_association" "test" {
  db_instance_identifier = aws_db_instance.test.identifier
  feature_name           = "S3_INTEGRATION"
  role_arn               = aws_iam_role.test.arn
}

resource "aws_db_instance" "test" {
  allocated_storage   = 10
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  license_model       = data.aws_rds_orderable_db_instance.test.license_model
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine        = "oracle-se2"
  license_model = "bring-your-own-license"
  storage_type  = "standard"

  preferred_instance_classes = ["db.m5.large", "db.m4.large", "db.r4.large"]
}

resource "aws_iam_role" "test" {
  assume_role_policy = data.aws_iam_policy_document.rds_assume_role_policy.json
  name               = %[1]q

  # ensure IAM role is created just before association to exercise IAM eventual consistency
  depends_on = [aws_db_instance.test]
}

data "aws_iam_policy_document" "rds_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"

    principals {
      identifiers = ["rds.${data.aws_partition.current.dns_suffix}"]
      type        = "Service"
    }
  }
}

data "aws_partition" "current" {}
`, rName)
}
