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

func TestAccQuickSightFolderMembership_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var folderMember awstypes.MemberIdArnPair
	resourceName := "aws_quicksight_folder_membership.test"
	folderResourceName := "aws_quicksight_folder.test"
	dataSetResourceName := "aws_quicksight_data_set.test"
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFolderMembershipDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFolderMembershipConfig_basic(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFolderMembershipExists(ctx, t, resourceName, &folderMember),
					resource.TestCheckResourceAttrPair(resourceName, "folder_id", folderResourceName, "folder_id"),
					resource.TestCheckResourceAttrPair(resourceName, "member_id", dataSetResourceName, "data_set_id"),
					resource.TestCheckResourceAttr(resourceName, "member_type", string(awstypes.MemberTypeDataset)),
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

func TestAccQuickSightFolderMembership_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var folderMember awstypes.MemberIdArnPair
	resourceName := "aws_quicksight_folder_membership.test"
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFolderMembershipDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFolderMembershipConfig_basic(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFolderMembershipExists(ctx, t, resourceName, &folderMember),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfquicksight.ResourceFolderMembership, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFolderMembershipExists(ctx context.Context, t *testing.T, n string, v *awstypes.MemberIdArnPair) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		output, err := tfquicksight.FindFolderMembershipByFourPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes["folder_id"], rs.Primary.Attributes["member_type"], rs.Primary.Attributes["member_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckFolderMembershipDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_folder_membership" {
				continue
			}

			_, err := tfquicksight.FindFolderMembershipByFourPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes["folder_id"], rs.Primary.Attributes["member_type"], rs.Primary.Attributes["member_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("QuickSight Folder Membership (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccFolderMembershipConfig_basic(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccDataSetConfigBasic(rId, rName),
		testAccFolderConfig_basic(rId, rName),
		`
resource "aws_quicksight_folder_membership" "test" {
  folder_id   = aws_quicksight_folder.test.folder_id
  member_type = "DATASET"
  member_id   = aws_quicksight_data_set.test.data_set_id
}
`)
}
