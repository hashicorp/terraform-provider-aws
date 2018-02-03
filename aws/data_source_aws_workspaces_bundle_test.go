package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsWorkspaceBundle_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDataSourceAwsWorkspaceBundleConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_workspaces_bundle.test", "bundle_id", "wsb-b0s22j3d7"),
					resource.TestCheckResourceAttrSet(
						"data.aws_workspaces_bundle.test", "name"),
					resource.TestCheckResourceAttrSet(
						"data.aws_workspaces_bundle.test", "owner"),
				),
			},
		},
	})
}

const testAccDataSourceAwsWorkspaceBundleConfig = `
data "aws_workspaces_bundle" "test" {
  bundle_id = "wsb-b0s22j3d7"
}
`
