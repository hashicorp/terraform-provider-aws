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

func TestAccWorkSpacesWebBrowserSettingsAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var browserSettings awstypes.BrowserSettings
	resourceName := "aws_workspacesweb_browser_settings_association.test"
	browserSettingsResourceName := "aws_workspacesweb_browser_settings.test"
	portalResourceName := "aws_workspacesweb_portal.test"
	browserPolicy1 := `{
		"chromePolicies":
		{
			"DefaultDownloadDirectory": {
				"value": "/home/as2-streaming-user/MyFiles/TemporaryFiles1"
			}
		}
	} `
	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrowserSettingsAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBrowserSettingsAssociationConfig_basic(browserPolicy1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserSettingsAssociationExists(ctx, t, resourceName, &browserSettings),
					resource.TestCheckResourceAttrPair(resourceName, "browser_settings_arn", browserSettingsResourceName, "browser_settings_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "portal_arn", portalResourceName, "portal_arn"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccBrowserSettingsAssociationImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "browser_settings_arn",
			},
			{
				ResourceName: resourceName,
				RefreshState: true,
			},
			{
				Config: testAccBrowserSettingsAssociationConfig_basic(browserPolicy1),
				Check: resource.ComposeAggregateTestCheckFunc(
					//The following checks are for the BrowserSettings Resource and the PortalResource (and not for the association resource).
					resource.TestCheckResourceAttr(browserSettingsResourceName, "associated_portal_arns.#", "1"),
					resource.TestCheckResourceAttrPair(browserSettingsResourceName, "associated_portal_arns.0", portalResourceName, "portal_arn"),
					resource.TestCheckResourceAttrPair(portalResourceName, "browser_settings_arn", browserSettingsResourceName, "browser_settings_arn"),
				),
			},
		},
	})
}

func TestAccWorkSpacesWebBrowserSettingsAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var browserSettings awstypes.BrowserSettings
	resourceName := "aws_workspacesweb_browser_settings_association.test"
	browserPolicy1 := `{
		"chromePolicies":
		{
			"DefaultDownloadDirectory": {
				"value": "/home/as2-streaming-user/MyFiles/TemporaryFiles1"
			}
		}
	} `
	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrowserSettingsAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBrowserSettingsAssociationConfig_basic(browserPolicy1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserSettingsAssociationExists(ctx, t, resourceName, &browserSettings),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfworkspacesweb.ResourceBrowserSettingsAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckBrowserSettingsAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workspacesweb_browser_settings_association" {
				continue
			}

			browserSettings, err := tfworkspacesweb.FindBrowserSettingsByARN(ctx, conn, rs.Primary.Attributes["browser_settings_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			// Check if the portal is still associated
			portalARN := rs.Primary.Attributes["portal_arn"]
			if slices.Contains(browserSettings.AssociatedPortalArns, portalARN) {
				return fmt.Errorf("WorkSpaces Web Browser Settings Association %s still exists", rs.Primary.Attributes["browser_settings_arn"])
			}
		}

		return nil
	}
}

func testAccCheckBrowserSettingsAssociationExists(ctx context.Context, t *testing.T, n string, v *awstypes.BrowserSettings) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		output, err := tfworkspacesweb.FindBrowserSettingsByARN(ctx, conn, rs.Primary.Attributes["browser_settings_arn"])

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

func testAccBrowserSettingsAssociationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["browser_settings_arn"], rs.Primary.Attributes["portal_arn"]), nil
	}
}

func testAccBrowserSettingsAssociationConfig_basic(browserPolicy string) string {
	return fmt.Sprintf(`
resource "aws_workspacesweb_portal" "test" {
  display_name = "test"
}

resource "aws_workspacesweb_browser_settings" "test" {
  browser_policy = %[1]q
}

resource "aws_workspacesweb_browser_settings_association" "test" {
  browser_settings_arn = aws_workspacesweb_browser_settings.test.browser_settings_arn
  portal_arn           = aws_workspacesweb_portal.test.portal_arn
}
`, browserPolicy)
}
