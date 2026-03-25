// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfredshiftserverless "github.com/hashicorp/terraform-provider-aws/internal/service/redshiftserverless"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftServerlessSnapshot_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_snapshot.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "namespace_name", "aws_redshiftserverless_namespace.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "snapshot_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrRetentionPeriod, "-1"),
					resource.TestCheckResourceAttr(resourceName, "admin_username", "admin"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "owner_account"),
					resource.TestCheckResourceAttr(resourceName, "accounts_with_provisioned_restore_access.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "accounts_with_restore_access.#", "0"),
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
					testAccCheckSnapshotExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "namespace_name", "aws_redshiftserverless_namespace.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "snapshot_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrRetentionPeriod, "10"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "owner_account"),
				),
			},
		},
	})
}

func TestAccRedshiftServerlessSnapshot_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_snapshot.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfredshiftserverless.ResourceSnapshot(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSnapshotDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).RedshiftServerlessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshiftserverless_snapshot" {
				continue
			}
			_, err := tfredshiftserverless.FindSnapshotByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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

func testAccCheckSnapshotExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Redshift Serverless Snapshot is not set")
		}

		conn := acctest.ProviderMeta(ctx, t).RedshiftServerlessClient(ctx)

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
