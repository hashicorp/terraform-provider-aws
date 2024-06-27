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
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGrafanaWorkspaceServiceAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_grafana_workspace_service_account.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var v types.ServiceAccountSummary

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.Grafana) },
		ErrorCheck:               acctest.ErrorCheck(t, grafana.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceServiceAccountConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceServiceAccountExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "grafana_role"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "workspace_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccCheckWorkspaceServiceAccountImportStateIdFunc(resourceName),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Grafana)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GrafanaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceServiceAccountConfig_basic(resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceServiceAccountExists(ctx, resourceName, &v),
				),
			},
		},
	})
}

func testAccCheckWorkspaceServiceAccountExists(ctx context.Context, n string, v *types.ServiceAccountSummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GrafanaClient(ctx)
		output, err := tfgrafana.FindWorkspaceServiceAccount(ctx, conn, rs.Primary.ID, rs.Primary.Attributes[names.AttrWorkspaceID])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckWorkspaceServiceAccountImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s,%s", rs.Primary.Attributes[names.AttrID], rs.Primary.Attributes["grafana_role"], rs.Primary.Attributes[names.AttrWorkspaceID]), nil
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
