package servicecatalog_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfservicecatalog "github.com/hashicorp/terraform-provider-aws/internal/service/servicecatalog"
)

func TestAccServiceCatalogOrganizationsAccess_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_organizations_access.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationsEnabled(t)
			acctest.PreCheckOrganizationManagementAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationsAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationsAccessConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationsAccessExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
		},
	})
}

func testAccCheckOrganizationsAccessDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_organizations_access" {
			continue
		}

		output, err := tfservicecatalog.WaitOrganizationsAccessStable(conn, tfservicecatalog.OrganizationsAccessStableTimeout)

		if err != nil {
			return fmt.Errorf("error describing Service Catalog AWS Organizations Access (%s): %w", rs.Primary.ID, err)
		}

		if output == "" {
			return fmt.Errorf("error getting Service Catalog AWS Organizations Access (%s): empty response", rs.Primary.ID)
		}

		return nil
	}

	return nil
}

func testAccCheckOrganizationsAccessExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn

		output, err := tfservicecatalog.WaitOrganizationsAccessStable(conn, tfservicecatalog.OrganizationsAccessStableTimeout)

		if err != nil {
			return fmt.Errorf("error describing Service Catalog AWS Organizations Access (%s): %w", rs.Primary.ID, err)
		}

		if output == "" {
			return fmt.Errorf("error getting Service Catalog AWS Organizations Access (%s): empty response", rs.Primary.ID)
		}

		if output != servicecatalog.AccessStatusEnabled && rs.Primary.Attributes["enabled"] == "true" {
			return fmt.Errorf("error getting Service Catalog AWS Organizations Access (%s): wrong setting", rs.Primary.ID)
		}

		if output == servicecatalog.AccessStatusEnabled && rs.Primary.Attributes["enabled"] == "false" {
			return fmt.Errorf("error getting Service Catalog AWS Organizations Access (%s): wrong setting", rs.Primary.ID)
		}

		return nil
	}
}

func testAccOrganizationsAccessConfig_basic() string {
	return `
resource "aws_servicecatalog_organizations_access" "test" {
  enabled = "true"
}
`
}
