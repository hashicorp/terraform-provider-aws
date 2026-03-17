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

func TestAccWorkSpacesWebIPAccessSettingsAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var ipAccessSettings awstypes.IpAccessSettings
	resourceName := "aws_workspacesweb_ip_access_settings_association.test"
	ipAccessSettingsResourceName := "aws_workspacesweb_ip_access_settings.test"
	portalResourceName := "aws_workspacesweb_portal.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAccessSettingsAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAccessSettingsAssociationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPAccessSettingsAssociationExists(ctx, t, resourceName, &ipAccessSettings),
					resource.TestCheckResourceAttrPair(resourceName, "ip_access_settings_arn", ipAccessSettingsResourceName, "ip_access_settings_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "portal_arn", portalResourceName, "portal_arn"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccIPAccessSettingsAssociationImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "ip_access_settings_arn",
			},
			{
				ResourceName: resourceName,
				RefreshState: true,
			},
			{
				Config: testAccIPAccessSettingsAssociationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					//The following checks are for the IPAccessSettings Resource and the PortalResource (and not for the association resource).
					resource.TestCheckResourceAttr(ipAccessSettingsResourceName, "associated_portal_arns.#", "1"),
					resource.TestCheckResourceAttrPair(ipAccessSettingsResourceName, "associated_portal_arns.0", portalResourceName, "portal_arn"),
					resource.TestCheckResourceAttrPair(portalResourceName, "ip_access_settings_arn", ipAccessSettingsResourceName, "ip_access_settings_arn"),
				),
			},
		},
	})
}

func TestAccWorkSpacesWebIPAccessSettingsAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var ipAccessSettings awstypes.IpAccessSettings
	resourceName := "aws_workspacesweb_ip_access_settings_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAccessSettingsAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAccessSettingsAssociationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPAccessSettingsAssociationExists(ctx, t, resourceName, &ipAccessSettings),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfworkspacesweb.ResourceIPAccessSettingsAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckIPAccessSettingsAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workspacesweb_ip_access_settings_association" {
				continue
			}

			ipAccessSettings, err := tfworkspacesweb.FindIPAccessSettingsByARN(ctx, conn, rs.Primary.Attributes["ip_access_settings_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			// Check if the portal is still associated
			portalARN := rs.Primary.Attributes["portal_arn"]
			if slices.Contains(ipAccessSettings.AssociatedPortalArns, portalARN) {
				return fmt.Errorf("WorkSpaces Web IP Access Settings Association %s still exists", rs.Primary.Attributes["ip_access_settings_arn"])
			}
		}

		return nil
	}
}

func testAccCheckIPAccessSettingsAssociationExists(ctx context.Context, t *testing.T, n string, v *awstypes.IpAccessSettings) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		output, err := tfworkspacesweb.FindIPAccessSettingsByARN(ctx, conn, rs.Primary.Attributes["ip_access_settings_arn"])

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

func testAccIPAccessSettingsAssociationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["ip_access_settings_arn"], rs.Primary.Attributes["portal_arn"]), nil
	}
}

func testAccIPAccessSettingsAssociationConfig_basic() string {
	return `
resource "aws_workspacesweb_portal" "test" {
  display_name = "test"
}

resource "aws_workspacesweb_ip_access_settings" "test" {
  display_name = "test"

  ip_rule {
    ip_range = "10.0.0.0/16"
  }
}

resource "aws_workspacesweb_ip_access_settings_association" "test" {
  ip_access_settings_arn = aws_workspacesweb_ip_access_settings.test.ip_access_settings_arn
  portal_arn             = aws_workspacesweb_portal.test.portal_arn
}
`
}
