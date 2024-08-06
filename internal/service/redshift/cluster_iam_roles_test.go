// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/redshift"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftClusterIAMRoles_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v redshift.Cluster
	resourceName := "aws_redshift_cluster_iam_roles.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterIAMRolesConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "iam_role_arns.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterIAMRolesConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "iam_role_arns.#", acctest.Ct2),
				),
			},
			{
				Config: testAccClusterIAMRolesConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "iam_role_arns.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccRedshiftClusterIAMRoles_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v redshift.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterIAMRolesConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfredshift.ResourceCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccClusterIAMRolesConfigBase(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_iam_role" "ec2" {
  name = "%[1]s-ec2"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_role" "lambda" {
  name = "%[1]s-lambda"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "lambda.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

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
`, rName))
}

func testAccClusterIAMRolesConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccClusterIAMRolesConfigBase(rName), `
resource "aws_redshift_cluster_iam_roles" "test" {
  cluster_identifier = aws_redshift_cluster.test.cluster_identifier
  iam_role_arns      = [aws_iam_role.ec2.arn]
}
`)
}

func testAccClusterIAMRolesConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccClusterIAMRolesConfigBase(rName), `
resource "aws_redshift_cluster_iam_roles" "test" {
  cluster_identifier = aws_redshift_cluster.test.cluster_identifier
  iam_role_arns      = [aws_iam_role.ec2.arn, aws_iam_role.lambda.arn]
}
`)
}
