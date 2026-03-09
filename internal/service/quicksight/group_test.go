// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQuickSightGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.Group
	resourceName := "aws_quicksight_group.default"
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, names.AttrGroupName, rName1),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "quicksight", fmt.Sprintf("group/default/%s", rName1)),
				),
			},
			{
				Config: testAccGroupConfig_basic(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, names.AttrGroupName, rName2),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "quicksight", fmt.Sprintf("group/default/%s", rName2)),
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

func TestAccQuickSightGroup_withDescription(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.Group
	resourceName := "aws_quicksight_group.default"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_description(rName, "Description 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Description 1"),
				),
			},
			{
				Config: testAccGroupConfig_description(rName, "Description 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Description 2"),
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

func TestAccQuickSightGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.Group
	resourceName := "aws_quicksight_group.default"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &group),
					acctest.CheckSDKResourceDisappears(ctx, t, tfquicksight.ResourceGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckGroupExists(ctx context.Context, t *testing.T, n string, v *awstypes.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		output, err := tfquicksight.FindGroupByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes[names.AttrNamespace], rs.Primary.Attributes[names.AttrGroupName])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckGroupDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_group" {
				continue
			}

			_, err := tfquicksight.FindGroupByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes[names.AttrNamespace], rs.Primary.Attributes[names.AttrGroupName])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("QuickSight Group (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_quicksight_group" "default" {
  group_name = %[1]q
}
`, rName)
}

func testAccGroupConfig_description(rName, description string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_quicksight_group" "default" {
  aws_account_id = data.aws_caller_identity.current.account_id
  group_name     = %[1]q
  description    = %[2]q
}
`, rName, description)
}
