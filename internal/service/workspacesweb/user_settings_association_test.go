// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package workspacesweb_test

import (
	"context"
	"fmt"
	"slices"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/workspacesweb/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfworkspacesweb "github.com/hashicorp/terraform-provider-aws/internal/service/workspacesweb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWorkSpacesWebUserSettingsAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var userSettings awstypes.UserSettings
	resourceName := "aws_workspacesweb_user_settings_association.test"
	userSettingsResourceName := "aws_workspacesweb_user_settings.test"
	portalResourceName := "aws_workspacesweb_portal.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserSettingsAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserSettingsAssociationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserSettingsAssociationExists(ctx, t, resourceName, &userSettings),
					resource.TestCheckResourceAttrPair(resourceName, "user_settings_arn", userSettingsResourceName, "user_settings_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "portal_arn", portalResourceName, "portal_arn"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccUserSettingsAssociationImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "user_settings_arn",
			},
			{
				ResourceName: resourceName,
				RefreshState: true,
			},
			{
				Config: testAccUserSettingsAssociationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					//The following checks are for the UserSettings Resource and the PortalResource (and not for the association resource).
					resource.TestCheckResourceAttr(userSettingsResourceName, "associated_portal_arns.#", "1"),
					resource.TestCheckResourceAttrPair(userSettingsResourceName, "associated_portal_arns.0", portalResourceName, "portal_arn"),
					resource.TestCheckResourceAttrPair(portalResourceName, "user_settings_arn", userSettingsResourceName, "user_settings_arn"),
				),
			},
		},
	})
}

func TestAccWorkSpacesWebUserSettingsAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var userSettings awstypes.UserSettings
	resourceName := "aws_workspacesweb_user_settings_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserSettingsAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserSettingsAssociationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserSettingsAssociationExists(ctx, t, resourceName, &userSettings),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfworkspacesweb.ResourceUserSettingsAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckUserSettingsAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workspacesweb_user_settings_association" {
				continue
			}

			userSettings, err := tfworkspacesweb.FindUserSettingsByARN(ctx, conn, rs.Primary.Attributes["user_settings_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			// Check if the portal is still associated
			portalARN := rs.Primary.Attributes["portal_arn"]
			if slices.Contains(userSettings.AssociatedPortalArns, portalARN) {
				return fmt.Errorf("WorkSpaces Web User Settings Association %s still exists", rs.Primary.Attributes["user_settings_arn"])
			}
		}

		return nil
	}
}

func testAccCheckUserSettingsAssociationExists(ctx context.Context, t *testing.T, n string, v *awstypes.UserSettings) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		output, err := tfworkspacesweb.FindUserSettingsByARN(ctx, conn, rs.Primary.Attributes["user_settings_arn"])

		if err != nil {
			return err
		}

		// Check if the portal is associated
		portalARN := rs.Primary.Attributes["portal_arn"]
		if !slices.Contains(output.AssociatedPortalArns, portalARN) {
			return fmt.Errorf("Association not found")
		}

		*v = *output

		return nil
	}
}

func testAccUserSettingsAssociationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["user_settings_arn"], rs.Primary.Attributes["portal_arn"]), nil
	}
}

func testAccUserSettingsAssociationConfig_basic() string {
	return `
resource "aws_workspacesweb_portal" "test" {
  display_name = "test"
}

resource "aws_workspacesweb_user_settings" "test" {
  copy_allowed     = "Enabled"
  download_allowed = "Enabled"
  paste_allowed    = "Enabled"
  print_allowed    = "Enabled"
  upload_allowed   = "Enabled"
}

resource "aws_workspacesweb_user_settings_association" "test" {
  user_settings_arn = aws_workspacesweb_user_settings.test.user_settings_arn
  portal_arn        = aws_workspacesweb_portal.test.portal_arn
}
`
}
