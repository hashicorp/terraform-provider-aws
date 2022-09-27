package controltower_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/controltower"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccControlTowerControlsDataSource_id(t *testing.T) {
	dataSourceName := "data.aws_controltower_controls.test"
	ouName := "Security"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckControlTowerDeployed(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, controltower.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccControlsDataSourceConfig_id(ouName),
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrGlobalARN(dataSourceName, "target_identifier", "organizations", regexp.MustCompile(`ou/.+`)),
				),
			},
		},
	})
}

func testAccControlsDataSourceConfig_id(ouName string) string {
	return fmt.Sprintf(`
data "aws_organizations_organization" "test" {}

data "aws_organizations_organizational_units" "test" {
  parent_id = data.aws_organizations_organization.test.roots[0].id
}

data "aws_controltower_controls" "test" {

  target_identifier = [
    for x in data.aws_organizations_organizational_units.test.children :
    x.arn if x.name == "%[1]s"
  ][0]

}
`, ouName)
}
