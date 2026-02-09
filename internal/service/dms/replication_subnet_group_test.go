// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dms_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdms "github.com/hashicorp/terraform-provider-aws/internal/service/dms"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDMSReplicationSubnetGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_replication_subnet_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationSubnetGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationSubnetGroupConfig_basic(rName, "desc1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationSubnetGroupExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "replication_subnet_group_arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_subnet_group_description", "desc1"),
					resource.TestCheckResourceAttr(resourceName, "replication_subnet_group_id", rName),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVPCID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReplicationSubnetGroupConfig_basic(rName, "desc2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSubnetGroupExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "replication_subnet_group_description", "desc2"),
				),
			},
		},
	})
}

func TestAccDMSReplicationSubnetGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_replication_subnet_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationSubnetGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationSubnetGroupConfig_basic(rName, "desc1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSubnetGroupExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdms.ResourceReplicationSubnetGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckReplicationSubnetGroupExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DMSClient(ctx)

		_, err := tfdms.FindReplicationSubnetGroupByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckReplicationSubnetGroupDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DMSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dms_replication_subnet_group" {
				continue
			}

			_, err := tfdms.FindReplicationSubnetGroupByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DMS Replication Subnet Group %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccReplicationSubnetGroupConfig_basic(rName, description string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 3), fmt.Sprintf(`
resource "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_id          = %[1]q
  replication_subnet_group_description = %[2]q
  subnet_ids                           = aws_subnet.test[*].id
}
`, rName, description))
}
