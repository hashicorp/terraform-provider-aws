package controltower_test

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/controltower"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccControlsDataSource_id(t *testing.T) {
	dataSourceName := "data.aws_controltower_controls.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckControlTower(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, controltower.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccControlsDataSourceConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrGlobalARN(dataSourceName, "target_identifier", "organizations", regexp.MustCompile(`ou/.+`)),
				),
			},
		},
	})
}

func testAccControlsDataSourceConfig_id() string {
	return `

	data "aws_organizations_organization" "test" {}

	data "aws_organizations_organizational_units" "test" {
	  parent_id = data.aws_organizations_organization.test.roots[0].id
	}
	
	data "aws_controltower_controls" "test" {
	
	  target_identifier = [
		for x in data.aws_organizations_organizational_units.test.children :
		x.arn if x.name == "Security"
	  ][0]
	
	}

`
}
