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
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccAWSIdentityStoreGroupDataSource_DisplayName(t *testing.T) {
	dataSourceName := "data.aws_identitystore_group.test"
	name := os.Getenv("AWS_IDENTITY_STORE_GROUP_NAME")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckAWSSSOAdminInstances(t)
			testAccPreCheckAWSIdentityStoreGroupName(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, identitystore.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIdentityStoreGroupDataSourceConfigDisplayName(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "group_id"),
					resource.TestCheckResourceAttr(dataSourceName, "display_name", name),
				),
			},
		},
	})
}

func TestAccAWSIdentityStoreGroupDataSource_GroupID(t *testing.T) {
	dataSourceName := "data.aws_identitystore_group.test"
	name := os.Getenv("AWS_IDENTITY_STORE_GROUP_NAME")
	groupID := os.Getenv("AWS_IDENTITY_STORE_GROUP_ID")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckAWSSSOAdminInstances(t)
			testAccPreCheckAWSIdentityStoreGroupName(t)
			testAccPreCheckAWSIdentityStoreGroupID(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, identitystore.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIdentityStoreGroupDataSourceConfigGroupID(name, groupID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "group_id", groupID),
					resource.TestCheckResourceAttrSet(dataSourceName, "display_name"),
				),
			},
		},
	})
}

func TestAccAWSIdentityStoreGroupDataSource_NonExistent(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSSSOAdminInstances(t) },
		ErrorCheck:   acctest.ErrorCheck(t, identitystore.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSIdentityStoreGroupDataSourceConfigNonExistent,
				ExpectError: regexp.MustCompile(`no Identity Store Group found matching criteria`),
			},
		},
	})
}

func testAccPreCheckAWSIdentityStoreGroupName(t *testing.T) {
	if os.Getenv("AWS_IDENTITY_STORE_GROUP_NAME") == "" {
		t.Skip("AWS_IDENTITY_STORE_GROUP_NAME env var must be set for AWS Identity Store Group acceptance test. " +
			"This is required until ListGroups API returns results without filtering by name.")
	}
}

func testAccPreCheckAWSIdentityStoreGroupID(t *testing.T) {
	if os.Getenv("AWS_IDENTITY_STORE_GROUP_ID") == "" {
		t.Skip("AWS_IDENTITY_STORE_GROUP_ID env var must be set for AWS Identity Store Group acceptance test. " +
			"This is required until ListGroups API returns results without filtering by name.")
	}
}

func testAccAWSIdentityStoreGroupDataSourceConfigDisplayName(name string) string {
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

func testAccAWSIdentityStoreGroupDataSourceConfigGroupID(name, id string) string {
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

const testAccAWSIdentityStoreGroupDataSourceConfigNonExistent = `
data "aws_ssoadmin_instances" "test" {}

data "aws_identitystore_group" "test" {
  filter {
    attribute_path  = "DisplayName"
    attribute_value = "does-not-exist"
  }
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
}
`

func testAccPreCheckAWSSSOAdminInstances(t *testing.T) {
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
