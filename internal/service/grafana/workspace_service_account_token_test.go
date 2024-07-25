// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/grafana"
	"github.com/aws/aws-sdk-go-v2/service/grafana/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfgrafana "github.com/hashicorp/terraform-provider-aws/internal/service/grafana"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGrafanaWorkspaceServiceAccountToken_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_grafana_workspace_service_account_token.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var v types.ServiceAccountTokenSummary

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GrafanaEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, grafana.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkspaceServiceAccountTokenDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceServiceAccountTokenConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceServiceAccountTokenExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(resourceName, "service_account_token_id"),
				),
			},
		},
	})
}

func TestAccGrafanaWorkspaceServiceAccountToken_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.ServiceAccountTokenSummary
	resourceName := "aws_grafana_workspace_service_account_token.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GrafanaEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GrafanaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkspaceServiceAccountTokenDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceServiceAccountTokenConfig_basic(resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceServiceAccountTokenExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfgrafana.ResourceWorkspaceServiceAccountToken, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckWorkspaceServiceAccountTokenExists(ctx context.Context, n string, v *types.ServiceAccountTokenSummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GrafanaClient(ctx)

		output, err := tfgrafana.FindWorkspaceServiceAccountTokenByThreePartKey(ctx, conn, rs.Primary.Attributes["workspace_id"], rs.Primary.Attributes["service_account_id"], rs.Primary.Attributes["service_account_token_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckWorkspaceServiceAccountTokenDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GrafanaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_grafana_workspace_service_account_token" {
				continue
			}

			_, err := tfgrafana.FindWorkspaceServiceAccountTokenByThreePartKey(ctx, conn, rs.Primary.Attributes["workspace_id"], rs.Primary.Attributes["service_account_id"], rs.Primary.Attributes["service_account_token_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Grafana Workspace Service Account Token %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccWorkspaceServiceAccountTokenConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccWorkspaceServiceAccountConfig_basic(rName), fmt.Sprintf(`
resource "aws_grafana_workspace_service_account_token" "test" {
  name               = %[1]q
  service_account_id = aws_grafana_workspace_service_account.test.service_account_id
  seconds_to_live    = 3600
  workspace_id       = aws_grafana_workspace.test.id
}
`, rName))
}
