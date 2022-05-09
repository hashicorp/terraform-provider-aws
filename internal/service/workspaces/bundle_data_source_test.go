package workspaces_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccWorkspaceBundleDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_workspaces_bundle.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, workspaces.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceBundleDataSourceConfig("wsb-b0s22j3d7"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "bundle_id", "wsb-b0s22j3d7"),
					resource.TestCheckResourceAttr(dataSourceName, "compute_type.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "compute_type.0.name", "PERFORMANCE"),
					resource.TestCheckResourceAttrSet(dataSourceName, "description"),
					resource.TestCheckResourceAttr(dataSourceName, "name", "Performance with Windows 7"),
					resource.TestCheckResourceAttr(dataSourceName, "owner", "Amazon"),
					resource.TestCheckResourceAttr(dataSourceName, "root_storage.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "root_storage.0.capacity", "80"),
					resource.TestCheckResourceAttr(dataSourceName, "user_storage.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "user_storage.0.capacity", "100"),
				),
			},
		},
	})
}

func testAccWorkspaceBundleDataSource_byOwnerName(t *testing.T) {
	dataSourceName := "data.aws_workspaces_bundle.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, workspaces.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceBundleDataSourceConfig_byOwnerName("AMAZON", "Value with Windows 10 and Office 2016"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "bundle_id", "wsb-df76rqys9"),
					resource.TestCheckResourceAttr(dataSourceName, "compute_type.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "compute_type.0.name", "VALUE"),
					resource.TestCheckResourceAttrSet(dataSourceName, "description"),
					resource.TestCheckResourceAttr(dataSourceName, "name", "Value with Windows 10 and Office 2016"),
					resource.TestCheckResourceAttr(dataSourceName, "owner", "Amazon"),
					resource.TestCheckResourceAttr(dataSourceName, "root_storage.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "root_storage.0.capacity", "80"),
					resource.TestCheckResourceAttr(dataSourceName, "user_storage.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "user_storage.0.capacity", "10"),
				),
			},
		},
	})
}

func testAccWorkspaceBundleDataSource_bundleIDAndNameConflict(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, workspaces.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccWorkspaceBundleDataSourceConfig_bundleIDAndOwnerNameConflict("wsb-df76rqys9", "AMAZON", "Value with Windows 10 and Office 2016"),
				ExpectError: regexp.MustCompile("\"bundle_id\": conflicts with owner"),
			},
		},
	})
}

func testAccWorkspaceBundleDataSource_privateOwner(t *testing.T) {
	dataSourceName := "data.aws_workspaces_bundle.test"
	bundleName := os.Getenv("AWS_WORKSPACES_BUNDLE_NAME")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccBundlePreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, workspaces.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceBundleDataSourceConfig_privateOwner(bundleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", bundleName),
				),
			},
		},
	})
}

func testAccBundlePreCheck(t *testing.T) {
	if os.Getenv("AWS_WORKSPACES_BUNDLE_NAME") == "" {
		t.Skip("AWS_WORKSPACES_BUNDLE_NAME env var must be set for AWS WorkSpaces private bundle acceptance test. This is required until AWS provides bundle creation API.")
	}
}

func testAccWorkspaceBundleDataSourceConfig(bundleID string) string {
	return fmt.Sprintf(`
data "aws_workspaces_bundle" "test" {
  bundle_id = %q
}
`, bundleID)
}

func testAccWorkspaceBundleDataSourceConfig_byOwnerName(owner, name string) string {
	return fmt.Sprintf(`
data "aws_workspaces_bundle" "test" {
  owner = %q
  name  = %q
}
`, owner, name)
}

func testAccWorkspaceBundleDataSourceConfig_bundleIDAndOwnerNameConflict(bundleID, owner, name string) string {
	return fmt.Sprintf(`
data "aws_workspaces_bundle" "test" {
  bundle_id = %q
  owner     = %q
  name      = %q
}
`, bundleID, owner, name)
}

func testAccWorkspaceBundleDataSourceConfig_privateOwner(name string) string {
	return fmt.Sprintf(`
data "aws_workspaces_bundle" "test" {
  name = %q
}
`, name)
}
