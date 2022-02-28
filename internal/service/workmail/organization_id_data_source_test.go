package workmail_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/workmail"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccWorkMailOrganizationDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_workmail_organization.test"
	datasourceName := "data.aws_workmail_organization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, workmail.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config:      testAccOrganizationDataSourceConfig_NonExistent,
				ExpectError: regexp.MustCompile(`not found`),
			},
			{
				Config: testAccOrganizationDataSourceConfig_Custom(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationCheckDataSource(datasourceName, resourceName),
				),
			},
		},
	})
}

func testAccOrganizationCheckDataSource(datasourceName, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource, ok := s.RootModule().Resources[datasourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", datasourceName)
		}

		dataSource, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", resourceName)
		}

		attrNames := []string{
			"alias",
			"arn",
		}

		for _, attrName := range attrNames {
			if resource.Primary.Attributes[attrName] != dataSource.Primary.Attributes[attrName] {
				return fmt.Errorf(
					"%s is %s; want %s",
					attrName,
					resource.Primary.Attributes[attrName],
					dataSource.Primary.Attributes[attrName],
				)
			}
		}

		return nil
	}
}

const testAccOrganizationDataSourceConfig_NonExistent = `
data "aws_workmail_organization" "test" {
  organization_id = "blah-test-does-not-exist"
}
`

func testAccOrganizationDataSourceConfig_VersionID(rName string) string {
	return fmt.Sprintf(`
resource "aws_workmail_organization" "test" {
  alias = "%[1]s"
}

data "aws_workmail_organization" "test" {
  secret_id  = aws_workmail_organization.test.id
  version_id = aws_workmail_organization.test.version_id
}
`, rName)
}

func testAccOrganizationDataSourceConfig_Custom(rName string) string {
	return fmt.Sprintf(`
resource "aws_workmail_organization" "test" {
  name = "%[1]s"
}

resource "aws_workmail_organization" "test" {
  alias  = "test-alias"
}

data "aws_workmail_organization" "test" {
  organization_id     = aws_workmail_organization.test.organization_id
}
`, rName)
}
