// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package redshift_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftSnapshotCopyGrant_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_snapshot_copy_grant.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotCopyGrantDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotCopyGrantConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotCopyGrantExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrKMSKeyID),
					resource.TestCheckResourceAttr(resourceName, "snapshot_copy_grant_name", rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func TestAccRedshiftSnapshotCopyGrant_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_snapshot_copy_grant.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotCopyGrantDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotCopyGrantConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotCopyGrantExists(ctx, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfredshift.ResourceSnapshotCopyGrant(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSnapshotCopyGrantDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshift_snapshot_copy_grant" {
				continue
			}

			_, err := tfredshift.FindSnapshotCopyGrantByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Redshift Snapshot Copy Grant %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSnapshotCopyGrantExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftClient(ctx)

		_, err := tfredshift.FindSnapshotCopyGrantByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccSnapshotCopyGrantConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshift_snapshot_copy_grant" "test" {
  snapshot_copy_grant_name = %[1]q
}
`, rName)
}
