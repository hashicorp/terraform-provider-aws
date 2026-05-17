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

func TestAccWorkSpacesWebDataProtectionSettingsAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var dataProtectionSettings awstypes.DataProtectionSettings
	resourceName := "aws_workspacesweb_data_protection_settings_association.test"
	dataProtectionSettingsResourceName := "aws_workspacesweb_data_protection_settings.test"
	portalResourceName := "aws_workspacesweb_portal.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProtectionSettingsAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProtectionSettingsAssociationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProtectionSettingsAssociationExists(ctx, t, resourceName, &dataProtectionSettings),
					resource.TestCheckResourceAttrPair(resourceName, "data_protection_settings_arn", dataProtectionSettingsResourceName, "data_protection_settings_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "portal_arn", portalResourceName, "portal_arn"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccDataProtectionSettingsAssociationImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "data_protection_settings_arn",
			},
			{
				ResourceName: resourceName,
				RefreshState: true,
			},
			{
				Config: testAccDataProtectionSettingsAssociationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					//The following checks are for the DataProtectionSettings Resource and the PortalResource (and not for the association resource).
					resource.TestCheckResourceAttr(dataProtectionSettingsResourceName, "associated_portal_arns.#", "1"),
					resource.TestCheckResourceAttrPair(dataProtectionSettingsResourceName, "associated_portal_arns.0", portalResourceName, "portal_arn"),
					resource.TestCheckResourceAttrPair(portalResourceName, "data_protection_settings_arn", dataProtectionSettingsResourceName, "data_protection_settings_arn"),
				),
			},
		},
	})
}

func TestAccWorkSpacesWebDataProtectionSettingsAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var dataProtectionSettings awstypes.DataProtectionSettings
	resourceName := "aws_workspacesweb_data_protection_settings_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProtectionSettingsAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProtectionSettingsAssociationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProtectionSettingsAssociationExists(ctx, t, resourceName, &dataProtectionSettings),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfworkspacesweb.ResourceDataProtectionSettingsAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDataProtectionSettingsAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workspacesweb_data_protection_settings_association" {
				continue
			}

			dataProtectionSettings, err := tfworkspacesweb.FindDataProtectionSettingsByARN(ctx, conn, rs.Primary.Attributes["data_protection_settings_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			// Check if the portal is still associated
			portalARN := rs.Primary.Attributes["portal_arn"]
			if slices.Contains(dataProtectionSettings.AssociatedPortalArns, portalARN) {
				return fmt.Errorf("WorkSpaces Web Data Protection Settings Association %s still exists", rs.Primary.Attributes["data_protection_settings_arn"])
			}
		}

		return nil
	}
}

func testAccCheckDataProtectionSettingsAssociationExists(ctx context.Context, t *testing.T, n string, v *awstypes.DataProtectionSettings) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		output, err := tfworkspacesweb.FindDataProtectionSettingsByARN(ctx, conn, rs.Primary.Attributes["data_protection_settings_arn"])

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

func testAccDataProtectionSettingsAssociationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["data_protection_settings_arn"], rs.Primary.Attributes["portal_arn"]), nil
	}
}

func testAccDataProtectionSettingsAssociationConfig_basic() string {
	return `
resource "aws_workspacesweb_portal" "test" {
  display_name = "test"
}

resource "aws_workspacesweb_data_protection_settings" "test" {
  display_name = "test"
}

resource "aws_workspacesweb_data_protection_settings_association" "test" {
  data_protection_settings_arn = aws_workspacesweb_data_protection_settings.test.data_protection_settings_arn
  portal_arn                   = aws_workspacesweb_portal.test.portal_arn
}
`
}
