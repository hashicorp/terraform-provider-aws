// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshiftserverless "github.com/hashicorp/terraform-provider-aws/internal/service/redshiftserverless"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftServerlessSnapshot_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "namespace_name", "aws_redshiftserverless_namespace.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "snapshot_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrRetentionPeriod, "-1"),
					resource.TestCheckResourceAttr(resourceName, "admin_username", "admin"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_account"),
					resource.TestCheckResourceAttr(resourceName, "accounts_with_provisioned_restore_access.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "accounts_with_restore_access.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSnapshotConfig_retention(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "namespace_name", "aws_redshiftserverless_namespace.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "snapshot_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrRetentionPeriod, acctest.Ct10),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_account"),
				),
			},
		},
	})
}

func TestAccRedshiftServerlessSnapshot_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfredshiftserverless.ResourceSnapshot(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSnapshotDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftServerlessConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshiftserverless_snapshot" {
				continue
			}
			_, err := tfredshiftserverless.FindSnapshotByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Redshift Serverless Snapshot %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSnapshotExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Redshift Serverless Snapshot is not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftServerlessConn(ctx)

		_, err := tfredshiftserverless.FindSnapshotByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccSnapshotConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
}

resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name = %[1]q
}

resource "aws_redshiftserverless_snapshot" "test" {
  namespace_name = aws_redshiftserverless_workgroup.test.namespace_name
  snapshot_name  = %[1]q
}
`, rName)
}

func testAccSnapshotConfig_retention(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
}

resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name = %[1]q
}

resource "aws_redshiftserverless_snapshot" "test" {
  namespace_name   = aws_redshiftserverless_workgroup.test.namespace_name
  snapshot_name    = %[1]q
  retention_period = 10
}
`, rName)
}
