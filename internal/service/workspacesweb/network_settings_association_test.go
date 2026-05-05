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

func TestAccWorkSpacesWebNetworkSettingsAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var networkSettings awstypes.NetworkSettings
	resourceName := "aws_workspacesweb_network_settings_association.test"
	networkSettingsResourceName := "aws_workspacesweb_network_settings.test"
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
		CheckDestroy:             testAccCheckNetworkSettingsAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkSettingsAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkSettingsAssociationExists(ctx, t, resourceName, &networkSettings),
					resource.TestCheckResourceAttrPair(resourceName, "network_settings_arn", networkSettingsResourceName, "network_settings_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "portal_arn", portalResourceName, "portal_arn"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccNetworkSettingsAssociationImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "network_settings_arn",
			},
			{
				ResourceName: resourceName,
				RefreshState: true,
			},
			{
				Config: testAccNetworkSettingsAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					//The following checks are for the NetworkSettings Resource and the PortalResource (and not for the association resource).
					resource.TestCheckResourceAttr(networkSettingsResourceName, "associated_portal_arns.#", "1"),
					resource.TestCheckResourceAttrPair(networkSettingsResourceName, "associated_portal_arns.0", portalResourceName, "portal_arn"),
					resource.TestCheckResourceAttrPair(portalResourceName, "network_settings_arn", networkSettingsResourceName, "network_settings_arn"),
				),
			},
		},
	})
}

func TestAccWorkSpacesWebNetworkSettingsAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var networkSettings awstypes.NetworkSettings
	resourceName := "aws_workspacesweb_network_settings_association.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkSettingsAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkSettingsAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkSettingsAssociationExists(ctx, t, resourceName, &networkSettings),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfworkspacesweb.ResourceNetworkSettingsAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckNetworkSettingsAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workspacesweb_network_settings_association" {
				continue
			}

			networkSettings, err := tfworkspacesweb.FindNetworkSettingsByARN(ctx, conn, rs.Primary.Attributes["network_settings_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			// Check if the portal is still associated
			portalARN := rs.Primary.Attributes["portal_arn"]
			if slices.Contains(networkSettings.AssociatedPortalArns, portalARN) {
				return fmt.Errorf("WorkSpaces Web Network Settings Association %s still exists", rs.Primary.Attributes["network_settings_arn"])
			}
		}

		return nil
	}
}

func testAccCheckNetworkSettingsAssociationExists(ctx context.Context, t *testing.T, n string, v *awstypes.NetworkSettings) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		output, err := tfworkspacesweb.FindNetworkSettingsByARN(ctx, conn, rs.Primary.Attributes["network_settings_arn"])

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

func testAccNetworkSettingsAssociationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["network_settings_arn"], rs.Primary.Attributes["portal_arn"]), nil
	}
}

func testAccNetworkSettingsAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccNetworkSettingsConfig_base(rName), `
resource "aws_workspacesweb_portal" "test" {
  display_name = "test"
}

resource "aws_workspacesweb_network_settings" "test" {
  vpc_id             = aws_vpc.test.id
  subnet_ids         = [aws_subnet.test[0].id, aws_subnet.test[1].id]
  security_group_ids = [aws_security_group.test[0].id, aws_security_group.test[1].id]
}

resource "aws_workspacesweb_network_settings_association" "test" {
  network_settings_arn = aws_workspacesweb_network_settings.test.network_settings_arn
  portal_arn           = aws_workspacesweb_portal.test.portal_arn
}
`)
}
