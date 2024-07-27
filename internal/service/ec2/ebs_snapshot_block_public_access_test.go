// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSnapshotBlockPublicAccess_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ebs_snapshot_block_public_access.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotBlockAccessDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotBlockPublicAccess_basic(string(types.SnapshotBlockPublicAccessStateBlockAllSharing)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotBlockPublicAccess(ctx, resourceName, false),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, "false"),
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

func testAccSnapshotBlockPublicAccess_basic(status string) string {
	return fmt.Sprintf(`
resource "aws_ebs_snapshot_block_public_access" "test" {
  status = %[1]s
}
`, status)
}
