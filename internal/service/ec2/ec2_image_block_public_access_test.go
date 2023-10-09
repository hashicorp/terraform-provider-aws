// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

// Note: these tests don't run in parallel because they conflict with each other
// This resource is once-per-region, so running multiple tests at the same time leads to tests stepping on
// each other's toes

func TestAccEC2ImageBlockPublicAccess_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_image_block_public_access.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             checkEC2ImageBlockPublicAccessDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEC2ImageBlockPublicAccessConfig_basic(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
			{
				Config: testAccEC2ImageBlockPublicAccessConfig_basic(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
		},
	},
	)
}

func TestAccEC2ImageBlockPublicAccess_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_image_block_public_access.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             checkEC2ImageBlockPublicAccessDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEC2ImageBlockPublicAccessConfig_basic(true),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceEC2ImageBlockPublicAccess(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	},
	)
}

func testAccEC2ImageBlockPublicAccessConfig_basic(enabled bool) string {
	return fmt.Sprintf(`
resource "aws_ec2_image_block_public_access" "test" {
  enabled = %[1]t
}
`, enabled)
}

func checkEC2ImageBlockPublicAccessDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		input := ec2.GetImageBlockPublicAccessStateInput{}

		output, err := conn.GetImageBlockPublicAccessState(ctx, &input)

		if err != nil {
			return err
		}

		if *output.ImageBlockPublicAccessState == string(ec2types.ImageBlockPublicAccessEnabledStateBlockNewSharing) {
			return fmt.Errorf("EC2 image public access block is still active")
		}

		return nil
	}
}
