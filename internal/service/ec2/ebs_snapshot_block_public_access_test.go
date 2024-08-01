// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2EBSSnapshotBlockPublicAccess_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ebs_snapshot_block_public_access.this"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		WorkingDir:               "/tmp",
		CheckDestroy:             testAccEC2EBSCheckSnapshotBlockAccessDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				Config:       testAccEC2EBSSnapshotBlockPublicAccess_basic(string(types.SnapshotBlockPublicAccessStateBlockAllSharing)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "block-all-sharing"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName: resourceName,
				Config:       testAccEC2EBSSnapshotBlockPublicAccess_basic(string(types.SnapshotBlockPublicAccessStateBlockNewSharing)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "block-new-sharing"),
				),
			},
		},
	})
}

func testAccEC2EBSCheckSnapshotBlockAccessDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)
		response, err := conn.GetSnapshotBlockPublicAccessState(ctx, &ec2.GetSnapshotBlockPublicAccessStateInput{})
		if err != nil {
			return err
		}

		if response.State != types.SnapshotBlockPublicAccessStateUnblocked {
			return fmt.Errorf("EBS encryption by default is not in expected state (%s)", types.SnapshotBlockPublicAccessStateUnblocked)
		}
		return nil
	}
}

func testAccEC2EBSSnapshotBlockPublicAccess_basic(state string) string {
	return fmt.Sprintf(`
resource "aws_ebs_snapshot_block_public_access" "this" {
  state = "%[1]s"
}
`, state)
}
