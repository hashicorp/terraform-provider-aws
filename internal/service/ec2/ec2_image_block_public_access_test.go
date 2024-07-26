// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2ImageBlockPublicAccess_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic: testAccImageBlockPublicAccess_basic,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccImageBlockPublicAccess_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_image_block_public_access.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccImageBlockPublicAccessConfig_basic("unblocked"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "unblocked"),
				),
			},
			{
				Config: testAccImageBlockPublicAccessConfig_basic("block-new-sharing"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "block-new-sharing"),
				),
			},
		},
	})
}

func testAccImageBlockPublicAccessConfig_basic(state string) string {
	return fmt.Sprintf(`
resource "aws_ec2_image_block_public_access" "test" {
  state = %[1]q
}
`, state)
}
