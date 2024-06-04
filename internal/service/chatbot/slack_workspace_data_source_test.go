// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chatbot_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccChatbotSlackWorkspaceDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	// The slack workspace must be created via the AWS Console. It cannot be created via APIs or Terraform.
	// Once it is created, export the name of the workspace in the env variable  for this test
	key := "CHATBOT_SLACK_WORKSPACE_NAME"
	workspace_name := os.Getenv(key)
	if workspace_name == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	dataSourceName := "data.aws_chatbot_slack_workspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChatbotServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSlackWorkspaceDataSourceConfig_basic(workspace_name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "slack_team_name", workspace_name),
					resource.TestCheckResourceAttrSet(dataSourceName, "slack_team_id"),
				),
			},
		},
	})
}

func testAccSlackWorkspaceDataSourceConfig_basic(workspace_name string) string {
	return fmt.Sprintf(`
data "aws_chatbot_slack_workspace" "test" {
  slack_team_name = "%[1]s"
}
`, workspace_name)
}
