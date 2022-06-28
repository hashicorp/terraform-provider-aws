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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccLicenseAssociation_freeTrial(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_grafana_license_association.test"
	workspaceResourceName := "aws_grafana_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(managedgrafana.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, managedgrafana.EndpointsID),
		CheckDestroy:      testAccCheckLicenseAssociationDestroy,
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLicenseAssociationConfig_basic(rName, managedgrafana.LicenseTypeEnterpriseFreeTrial),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLicenseAssociationExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "free_trial_expiration"),
					resource.TestCheckResourceAttr(resourceName, "license_type", managedgrafana.LicenseTypeEnterpriseFreeTrial),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, "id"),
				),
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

func testAccCheckLicenseAssociationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Grafana Workspace ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GrafanaConn

		_, err := tfgrafana.FindLicensedWorkspaceByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckLicenseAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GrafanaConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_grafana_license_association" {
			continue
		}

		_, err := tfgrafana.FindLicensedWorkspaceByID(conn, rs.Primary.ID)

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
