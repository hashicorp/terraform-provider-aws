package grafana_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/managedgrafana"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccWorkspaceApiKey_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace_api_key.test"
	workspaceResourceName := "aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(managedgrafana.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy:      nil,
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceApiKeyProvider_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "key_name"),
					resource.TestCheckResourceAttrSet(resourceName, "key_role"),
					resource.TestCheckResourceAttrSet(resourceName, "seconds_to_live"),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, "id"),
				),
			},
		},
	})
}

func testAccWorkspaceApiKeyProvider_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
	name = "test_iam"

	assume_role_policy = jsonencode({
		Version = "2012-10-17"
		Statement = [
		  {
			Action = "sts:AssumeRole"
			Effect = "Allow"
			Sid    = ""
			Principal = {
			  Service = "grafana.amazonaws.com"
			}
		  },
		]
	})
}
resource "aws_grafana_workspace" "test" {
	account_access_type      = "CURRENT_ACCOUNT"
	authentication_providers = ["SAML"]
	permission_type          = "SERVICE_MANAGED"
	role_arn                 = aws_iam_role.test.arn
}	
resource "aws_grafana_workspace_api_key" "test" {
  key_name			 = "test-key-1"
  key_role			 = "EDITOR"
  seconds_to_live 	 = 3600
  workspace_id       = aws_grafana_workspace.test.id
}
`)
}
