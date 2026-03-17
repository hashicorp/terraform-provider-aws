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

func TestAccWorkSpacesWebUserAccessLoggingSettingsAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var userAccessLoggingSettings awstypes.UserAccessLoggingSettings
	resourceName := "aws_workspacesweb_user_access_logging_settings_association.test"
	userAccessLoggingSettingsResourceName := "aws_workspacesweb_user_access_logging_settings.test"
	portalResourceName := "aws_workspacesweb_portal.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserAccessLoggingSettingsAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserAccessLoggingSettingsAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserAccessLoggingSettingsAssociationExists(ctx, t, resourceName, &userAccessLoggingSettings),
					resource.TestCheckResourceAttrPair(resourceName, "user_access_logging_settings_arn", userAccessLoggingSettingsResourceName, "user_access_logging_settings_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "portal_arn", portalResourceName, "portal_arn"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccUserAccessLoggingSettingsAssociationImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "user_access_logging_settings_arn",
			},
			{
				ResourceName: resourceName,
				RefreshState: true,
			},
			{
				Config: testAccUserAccessLoggingSettingsAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					//The following checks are for the UserAccessLoggingSettings Resource and the PortalResource (and not for the association resource).
					resource.TestCheckResourceAttr(userAccessLoggingSettingsResourceName, "associated_portal_arns.#", "1"),
					resource.TestCheckResourceAttrPair(userAccessLoggingSettingsResourceName, "associated_portal_arns.0", portalResourceName, "portal_arn"),
					resource.TestCheckResourceAttrPair(portalResourceName, "user_access_logging_settings_arn", userAccessLoggingSettingsResourceName, "user_access_logging_settings_arn"),
				),
			},
		},
	})
}

func TestAccWorkSpacesWebUserAccessLoggingSettingsAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var userAccessLoggingSettings awstypes.UserAccessLoggingSettings
	resourceName := "aws_workspacesweb_user_access_logging_settings_association.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserAccessLoggingSettingsAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserAccessLoggingSettingsAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserAccessLoggingSettingsAssociationExists(ctx, t, resourceName, &userAccessLoggingSettings),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfworkspacesweb.ResourceUserAccessLoggingSettingsAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckUserAccessLoggingSettingsAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workspacesweb_user_access_logging_settings_association" {
				continue
			}

			userAccessLoggingSettings, err := tfworkspacesweb.FindUserAccessLoggingSettingsByARN(ctx, conn, rs.Primary.Attributes["user_access_logging_settings_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			// Check if the portal is still associated
			portalARN := rs.Primary.Attributes["portal_arn"]
			if slices.Contains(userAccessLoggingSettings.AssociatedPortalArns, portalARN) {
				return fmt.Errorf("WorkSpaces Web User Access Logging Settings Association %s still exists", rs.Primary.Attributes["user_access_logging_settings_arn"])
			}
		}

		return nil
	}
}

func testAccCheckUserAccessLoggingSettingsAssociationExists(ctx context.Context, t *testing.T, n string, v *awstypes.UserAccessLoggingSettings) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		output, err := tfworkspacesweb.FindUserAccessLoggingSettingsByARN(ctx, conn, rs.Primary.Attributes["user_access_logging_settings_arn"])

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

func testAccUserAccessLoggingSettingsAssociationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["user_access_logging_settings_arn"], rs.Primary.Attributes["portal_arn"]), nil
	}
}

func testAccUserAccessLoggingSettingsAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_workspacesweb_portal" "test" {
  display_name = "test"
}

resource "aws_kinesis_stream" "test" {
  name        = "amazon-workspaces-web-%[1]s"
  shard_count = 1
}

resource "aws_workspacesweb_user_access_logging_settings" "test" {
  kinesis_stream_arn = aws_kinesis_stream.test.arn
}

resource "aws_workspacesweb_user_access_logging_settings_association" "test" {
  user_access_logging_settings_arn = aws_workspacesweb_user_access_logging_settings.test.user_access_logging_settings_arn
  portal_arn                       = aws_workspacesweb_portal.test.portal_arn
}
`, rName)
}
