// Copyright (c) HashiCorp, Inc.
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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfgrafana "github.com/hashicorp/terraform-provider-aws/internal/service/grafana"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWorkspaceServiceAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_grafana_workspace_service_account.test"
	var v types.ServiceAccountSummary

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, grafana.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkspaceServiceAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceServiceAccountConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceServiceAccountExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrAlias, ""),
					resource.TestCheckResourceAttrSet(resourceName, "service_account_role"),
					resource.TestCheckResourceAttrSet(resourceName, "service_account_name"),
					resource.TestCheckResourceAttrSet(resourceName, "workspace_id"),
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

func TestAccWorkspaceServiceAccount_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.ServiceAccountSummary
	resourceName := "aws_grafana_workspace_service_account.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AMPEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkspaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceServiceAccountConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceServiceAccountExists(ctx, resourceName, &v),
					// acctest.CheckResourceDisappears(ctx, acctest.Provider, tfgrafana.ResourceWorkspaceServiceAccount(), resourceName),
				),
				ExpectNonEmptyPlan: true,
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

		output, err := tfgrafana.FindWorkspaceServiceAccountByID(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["workspace_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckWorkspaceServiceAccountDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GrafanaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_grafana_workspace_service_account" {
				continue
			}

			_, err := tfgrafana.FindWorkspaceServiceAccountByID(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["workspace_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Prometheus Workspace %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

// func testAccCheckWorkspaceRecreated(i, j *types.WorkspaceDescription) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		if before, after := aws.ToString(i.WorkspaceId), aws.ToString(j.WorkspaceId); before == after {
// 			return fmt.Errorf("Prometheus Workspace (%s) not recreated", before)
// 		}

// 		return nil
// 	}
// }

// func testAccCheckWorkspaceNotRecreated(i, j *types.WorkspaceDescription) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		if before, after := aws.ToString(i.WorkspaceId), aws.ToString(j.WorkspaceId); before != after {
// 			return fmt.Errorf("Prometheus Workspace (%s) recreated", before)
// 		}

// 		return nil
// 	}
// }

func testAccWorkspaceServiceAccountConfig_basic() string {
	return `
resource "aws_grafana_workspace_service_account" "test" {}
`
}

func testAccWorkspaceConfig_alias(rName string) string {
	return fmt.Sprintf(`
resource "aws_prometheus_workspace" "test" {
  alias = %[1]q
}
`, rName)
}

func testAccWorkspaceConfig_loggingConfiguration(rName string, idx int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  count = 2

  name = "%[1]s-${count.index}"
}

resource "aws_prometheus_workspace" "test" {
  logging_configuration {
    log_group_arn = "${aws_cloudwatch_log_group.test[%[2]d].arn}:*"
  }
}
`, rName, idx)
}

func testAccWorkspaceConfig_kms(rName string) string {
	return fmt.Sprintf(`
resource "aws_prometheus_workspace" "test" {
  alias       = %[1]q
  kms_key_arn = aws_kms_key.test.arn
}

resource "aws_kms_key" "test" {
  description             = "Test"
  deletion_window_in_days = 7
}
`, rName)
}
