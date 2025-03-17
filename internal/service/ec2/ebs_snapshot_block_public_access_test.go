// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2EBSSnapshotBlockPublicAccess_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ebs_snapshot_block_public_access.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		WorkingDir:               "/tmp",
		CheckDestroy:             testAccCheckEBSSnapshotBlockAccessDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				Config:       testAccEBSSnapshotBlockPublicAccess_basic(string(types.SnapshotBlockPublicAccessStateBlockAllSharing)),
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
				Config:       testAccEBSSnapshotBlockPublicAccess_basic(string(types.SnapshotBlockPublicAccessStateBlockNewSharing)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "block-new-sharing"),
				),
			},
		},
	})
}

func testAccCheckEBSSnapshotBlockAccessDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)
		input := ec2.GetSnapshotBlockPublicAccessStateInput{}
		response, err := conn.GetSnapshotBlockPublicAccessState(ctx, &input)
		if err != nil {
			return err
		}

		if response.State != types.SnapshotBlockPublicAccessStateUnblocked {
			return fmt.Errorf("EBS encryption by default is not in expected state (%s)", types.SnapshotBlockPublicAccessStateUnblocked)
		}
		return nil
	}
}

func testAccEBSSnapshotBlockPublicAccess_basic(state string) string {
	return fmt.Sprintf(`
resource "aws_ebs_snapshot_block_public_access" "test" {
  state = %[1]q
}
`, state)
}
