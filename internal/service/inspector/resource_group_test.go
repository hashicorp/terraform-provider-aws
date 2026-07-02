// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package inspector_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/inspector/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfinspector "github.com/hashicorp/terraform-provider-aws/internal/service/inspector"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccInspectorResourceGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ResourceGroup
	resourceName := "aws_inspector_resource_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.InspectorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGroupConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGroupExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "inspector", regexache.MustCompile(`resourcegroup/.+`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "foo"),
				),
			},
		},
	})
}

func testAccCheckResourceGroupExists(ctx context.Context, t *testing.T, n string, v *awstypes.ResourceGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).InspectorClient(ctx)

		output, err := tfinspector.FindResourceGroupByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/inspector.amazonaws.com")
}

var testAccResourceGroupConfig_basic = `
resource "aws_inspector_resource_group" "test" {
  tags = {
    Name = "foo"
  }
}
`
