// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/grafana/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfgrafana "github.com/hashicorp/terraform-provider-aws/internal/service/grafana"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccLicenseAssociation_freeTrial(t *testing.T) {
	acctest.Skip(t, "ENTERPRISE_FREE_TRIAL has been deprecated and is no longer offered")

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_license_association.test"
	workspaceResourceName := "aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.GrafanaEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GrafanaServiceID),
		CheckDestroy:             testAccCheckLicenseAssociationDestroy(ctx),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLicenseAssociationConfig_basic(rName, string(awstypes.LicenseTypeEnterpriseFreeTrial)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLicenseAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "free_trial_expiration"),
					resource.TestCheckResourceAttr(resourceName, "license_type", string(awstypes.LicenseTypeEnterpriseFreeTrial)),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, names.AttrID),
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

func testAccLicenseAssociationConfig_basic(rName string, licenseType string) string {
	return acctest.ConfigCompose(testAccWorkspaceConfig_authenticationProvider(rName, "SAML"), fmt.Sprintf(`
resource "aws_grafana_license_association" "test" {
  workspace_id = aws_grafana_workspace.test.id
  license_type = %[1]q
}
`, licenseType))
}

func testAccCheckLicenseAssociationExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GrafanaClient(ctx)

		_, err := tfgrafana.FindLicensedWorkspaceByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckLicenseAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GrafanaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_grafana_license_association" {
				continue
			}

			_, err := tfgrafana.FindLicensedWorkspaceByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Grafana License Association %s still exists", rs.Primary.ID)
		}
		return nil
	}
}
