package grafana_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/managedgrafana"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfgrafana "github.com/hashicorp/terraform-provider-aws/internal/service/grafana"
)

func TestAccGrafanaWorkspace_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_workspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(managedgrafana.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy: nil,
		Providers:    acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "account_access_type", managedgrafana.AccountAccessTypeCurrentAccount),
					resource.TestCheckResourceAttr(resourceName, "permission_type", managedgrafana.PermissionTypeCustomerManaged),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", managedgrafana.WorkspaceStatusActive),
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

func testAccWorkspaceConfigBasic(name string) string {
	return fmt.Sprintf(`
resource "aws_grafana_workspace" "test" {
	account_access_type = "CURRENT_ACCOUNT"
	authentication_providers = ["SAML"]
	permission_type = "CUSTOMER_MANAGED"
	name = %[1]q
	role_arn = aws_iam_role.test.arn
}

resource "aws_iam_role" "test" {
	name = %[1]q
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
`, name)
}

func testAccCheckWorkspaceExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Grafana Workspace ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GrafanaConn

		_, err := tfgrafana.FindWorkspaceById(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}
