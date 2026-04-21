// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package grafana_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/grafana"
	"github.com/aws/aws-sdk-go-v2/service/grafana/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfgrafana "github.com/hashicorp/terraform-provider-aws/internal/service/grafana"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGrafanaWorkspaceServiceAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_grafana_workspace_service_account.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	var v types.ServiceAccountSummary

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GrafanaEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, grafana.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkspaceServiceAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceServiceAccountConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceServiceAccountExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "service_account_id"),
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

func TestAccGrafanaWorkspaceServiceAccount_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.ServiceAccountSummary
	resourceName := "aws_grafana_workspace_service_account.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GrafanaEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GrafanaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkspaceServiceAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceServiceAccountConfig_basic(resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceServiceAccountExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfgrafana.ResourceWorkspaceServiceAccount, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckWorkspaceServiceAccountExists(ctx context.Context, t *testing.T, n string, v *types.ServiceAccountSummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).GrafanaClient(ctx)

		output, err := tfgrafana.FindWorkspaceServiceAccountByTwoPartKey(ctx, conn, rs.Primary.Attributes["workspace_id"], rs.Primary.Attributes["service_account_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckWorkspaceServiceAccountDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).GrafanaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_grafana_workspace_service_account" {
				continue
			}

			_, err := tfgrafana.FindWorkspaceServiceAccountByTwoPartKey(ctx, conn, rs.Primary.Attributes["workspace_id"], rs.Primary.Attributes["service_account_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Grafana Workspace Service Account %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccWorkspaceServiceAccountConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccWorkspaceConfig_authenticationProvider(rName, "AWS_SSO"), fmt.Sprintf(`
resource "aws_grafana_workspace_service_account" "test" {
  name         = %[1]q
  grafana_role = "ADMIN"
  workspace_id = aws_grafana_workspace.test.id
}
`, rName))
}
