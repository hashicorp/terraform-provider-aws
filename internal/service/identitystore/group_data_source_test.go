package identitystore_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/identitystore"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccIdentityStoreGroupDataSource_displayName(t *testing.T) {
	dataSourceName := "data.aws_identitystore_group.test"
	name := os.Getenv("AWS_IDENTITY_STORE_GROUP_NAME")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckSSOAdminInstances(t)
			testAccPreCheckGroupName(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, identitystore.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDisplayNameDataSourceConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "group_id"),
					resource.TestCheckResourceAttr(dataSourceName, "display_name", name),
				),
			},
		},
	})
}

func TestAccIdentityStoreGroupDataSource_groupID(t *testing.T) {
	dataSourceName := "data.aws_identitystore_group.test"
	name := os.Getenv("AWS_IDENTITY_STORE_GROUP_NAME")
	groupID := os.Getenv("AWS_IDENTITY_STORE_GROUP_ID")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckSSOAdminInstances(t)
			testAccPreCheckGroupName(t)
			testAccPreCheckGroupID(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, identitystore.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupGroupIDDataSourceConfig(name, groupID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "group_id", groupID),
					resource.TestCheckResourceAttrSet(dataSourceName, "display_name"),
				),
			},
		},
	})
}

func TestAccIdentityStoreGroupDataSource_nonExistent(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSSOAdminInstances(t) },
		ErrorCheck:        acctest.ErrorCheck(t, identitystore.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config:      testAccGroupNonExistentDataSourceConfig,
				ExpectError: regexp.MustCompile(`no Identity Store Group found matching criteria`),
			},
		},
	})
}

func testAccPreCheckGroupName(t *testing.T) {
	if os.Getenv("AWS_IDENTITY_STORE_GROUP_NAME") == "" {
		t.Skip("AWS_IDENTITY_STORE_GROUP_NAME env var must be set for AWS Identity Store Group acceptance test. " +
			"This is required until ListGroups API returns results without filtering by name.")
	}
}

func testAccPreCheckGroupID(t *testing.T) {
	if os.Getenv("AWS_IDENTITY_STORE_GROUP_ID") == "" {
		t.Skip("AWS_IDENTITY_STORE_GROUP_ID env var must be set for AWS Identity Store Group acceptance test. " +
			"This is required until ListGroups API returns results without filtering by name.")
	}
}

func testAccGroupDisplayNameDataSourceConfig(name string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

data "aws_identitystore_group" "test" {
  filter {
    attribute_path  = "DisplayName"
    attribute_value = %q
  }
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
}
`, name)
}

func testAccGroupGroupIDDataSourceConfig(name, id string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

data "aws_identitystore_group" "test" {
  filter {
    attribute_path  = "DisplayName"
    attribute_value = %q
  }

  group_id = %q

  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
}
`, name, id)
}

const testAccGroupNonExistentDataSourceConfig = `
data "aws_ssoadmin_instances" "test" {}

data "aws_identitystore_group" "test" {
  filter {
    attribute_path  = "DisplayName"
    attribute_value = "does-not-exist"
  }
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
}
`

func testAccPreCheckSSOAdminInstances(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminConn

	var instances []*ssoadmin.InstanceMetadata
	err := conn.ListInstancesPages(&ssoadmin.ListInstancesInput{}, func(page *ssoadmin.ListInstancesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		instances = append(instances, page.Instances...)

		return !lastPage
	})

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if len(instances) == 0 {
		t.Skip("skipping acceptance testing: No SSO Instance found.")
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
