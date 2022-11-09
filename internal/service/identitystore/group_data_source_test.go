package identitystore_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/identitystore"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccIdentityStoreGroupDataSource_displayName(t *testing.T) {
	resourceName := "aws_identitystore_group.test"
	dataSourceName := "data.aws_identitystore_group.test"
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckSSOAdminInstances(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, identitystore.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfig_displayName(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "display_name", resourceName, "display_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "group_id", resourceName, "group_id"),
				),
			},
		},
	})
}

func TestAccIdentityStoreGroupDataSource_groupID(t *testing.T) {
	resourceName := "aws_identitystore_group.test"
	dataSourceName := "data.aws_identitystore_group.test"
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckSSOAdminInstances(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, identitystore.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfig_id(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "display_name", resourceName, "display_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "group_id", resourceName, "group_id"),
				),
			},
		},
	})
}

func TestAccIdentityStoreGroupDataSource_nonExistent(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckSSOAdminInstances(t) },
		ErrorCheck:               acctest.ErrorCheck(t, identitystore.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccGroupDataSourceConfig_nonExistent,
				ExpectError: regexp.MustCompile(`no Identity Store Group found matching criteria`),
			},
		},
	})
}

func testAccGroupDataSourceConfig_displayName(name string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_group" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  display_name      = %[1]q
}

data "aws_identitystore_group" "test" {
  filter {
    attribute_path  = "DisplayName"
    attribute_value = aws_identitystore_group.test.display_name
  }

  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
}
`, name)
}

func testAccGroupDataSourceConfig_id(name string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_group" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  display_name      = %[1]q
}

data "aws_identitystore_group" "test" {
  filter {
    attribute_path  = "DisplayName"
    attribute_value = aws_identitystore_group.test.display_name
  }

  group_id = aws_identitystore_group.test.group_id

  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
}
`, name)
}

const testAccGroupDataSourceConfig_nonExistent = `
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
