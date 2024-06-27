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

func TestAccGrafanaWorkspaceServiceAccountToken_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_grafana_workspace_service_account_token.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var v types.ServiceAccountTokenSummary

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.Grafana) },
		ErrorCheck:               acctest.ErrorCheck(t, grafana.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceServiceAccountTokenConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceServiceAccountTokenExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrKey),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrName),
					resource.TestCheckResourceAttrSet(resourceName, "service_account_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
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
			acctest.PreCheckPartitionHasService(t, names.Grafana)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GrafanaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceServiceAccountTokenConfig_basic(resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceServiceAccountTokenExists(ctx, resourceName, &v),
				),
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
		output, err := tfgrafana.FindWorkspaceServiceAccountToken(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["service_account_id"], rs.Primary.Attributes[names.AttrWorkspaceID])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccWorkspaceServiceAccountTokenConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccWorkspaceServiceAccountConfig_basic(rName), fmt.Sprintf(`
resource "aws_grafana_workspace_service_account_token" "test" {
  name               = %[1]q
  service_account_id = aws_grafana_workspace_service_account.test.id
  seconds_to_live    = 3600
  workspace_id       = aws_grafana_workspace.test.id
}
`, rName))
}
