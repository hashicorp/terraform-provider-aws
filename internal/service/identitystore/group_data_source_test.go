package identitystore_test

import (
	"context"
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

func TestAccIdentityStoreGroupDataSource_filterDisplayName(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_identitystore_group.test"
	dataSourceName := "data.aws_identitystore_group.test"
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, identitystore.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfig_filterDisplayName(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "display_name", resourceName, "display_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "group_id", resourceName, "group_id"),
					resource.TestCheckResourceAttr(dataSourceName, "external_ids.#", "0"),
				),
			},
		},
	})
}

func TestAccIdentityStoreGroupDataSource_uniqueAttributeDisplayName(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_identitystore_group.test"
	dataSourceName := "data.aws_identitystore_group.test"
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, identitystore.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfig_uniqueAttributeDisplayName(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "display_name", resourceName, "display_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "group_id", resourceName, "group_id"),
					resource.TestCheckResourceAttr(dataSourceName, "external_ids.#", "0"),
				),
			},
		},
	})
}

func TestAccIdentityStoreGroupDataSource_filterDisplayNameAndGroupId(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_identitystore_group.test"
	dataSourceName := "data.aws_identitystore_group.test"
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, identitystore.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfig_filterDisplayNameAndGroupId(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "display_name", resourceName, "display_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "group_id", resourceName, "group_id"),
					resource.TestCheckResourceAttr(dataSourceName, "external_ids.#", "0"),
				),
			},
		},
	})
}

func TestAccIdentityStoreGroupDataSource_nonExistent(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, identitystore.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccGroupDataSourceConfig_nonExistent,
				ExpectError: regexp.MustCompile(`no Identity Store Group found matching criteria`),
			},
		},
	})
}

func TestAccIdentityStoreGroupDataSource_groupIdFilterMismatch(t *testing.T) {
	ctx := acctest.Context(t)
	name1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	name2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, identitystore.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccGroupDataSourceConfig_groupIdFilterMismatch(name1, name2),
				ExpectError: regexp.MustCompile(`no Identity Store Group found matching criteria`),
			},
		},
	})
}

func TestAccIdentityStoreGroupDataSource_externalIdConflictsWithUniqueAttribute(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, identitystore.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccGroupDataSourceConfig_externalIdConflictsWithUniqueAttribute,
				ExpectError: regexp.MustCompile(`Invalid combination of arguments`),
			},
		},
	})
}

func TestAccIdentityStoreGroupDataSource_filterConflictsWithUniqueAttribute(t *testing.T) {
	ctx := acctest.Context(t)
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, identitystore.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccGroupDataSourceConfig_filterConflictsWithUniqueAttribute(name),
				ExpectError: regexp.MustCompile(`Conflicting configuration arguments`),
			},
		},
	})
}

func TestAccIdentityStoreGroupDataSource_groupIdConflictsWithUniqueAttribute(t *testing.T) {
	ctx := acctest.Context(t)
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, identitystore.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccGroupDataSourceConfig_groupIdConflictsWithUniqueAttribute(name),
				ExpectError: regexp.MustCompile(`Conflicting configuration arguments`),
			},
		},
	})
}

func TestAccIdentityStoreGroupDataSource_filterConflictsWithExternalId(t *testing.T) {
	ctx := acctest.Context(t)
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, identitystore.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccGroupDataSourceConfig_filterConflictsWithExternalId(name),
				ExpectError: regexp.MustCompile(`Conflicting configuration arguments`),
			},
		},
	})
}

func TestAccIdentityStoreGroupDataSource_groupIdConflictsWithExternalId(t *testing.T) {
	ctx := acctest.Context(t)
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, identitystore.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccGroupDataSourceConfig_groupIdConflictsWithExternalId(name),
				ExpectError: regexp.MustCompile(`Conflicting configuration arguments`),
			},
		},
	})
}

func testAccGroupDataSourceConfig_base(name string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_group" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  display_name      = %[1]q
  description       = "Acceptance Test"
}
`, name)
}

func testAccGroupDataSourceConfig_filterDisplayName(name string) string {
	return acctest.ConfigCompose(
		testAccGroupDataSourceConfig_base(name),
		`
data "aws_identitystore_group" "test" {
  filter {
    attribute_path  = "DisplayName"
    attribute_value = aws_identitystore_group.test.display_name
  }

  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
}
`,
	)
}

func testAccGroupDataSourceConfig_uniqueAttributeDisplayName(name string) string {
	return acctest.ConfigCompose(
		testAccGroupDataSourceConfig_base(name),
		`
data "aws_identitystore_group" "test" {
  alternate_identifier {
    unique_attribute {
      attribute_path  = "DisplayName"
      attribute_value = aws_identitystore_group.test.display_name
    }
  }

  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
}
`,
	)
}

func testAccGroupDataSourceConfig_filterDisplayNameAndGroupId(name string) string {
	return acctest.ConfigCompose(
		testAccGroupDataSourceConfig_base(name),
		`
data "aws_identitystore_group" "test" {
  filter {
    attribute_path  = "DisplayName"
    attribute_value = aws_identitystore_group.test.display_name
  }

  group_id = aws_identitystore_group.test.group_id

  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
}
`,
	)
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

func testAccGroupDataSourceConfig_groupIdFilterMismatch(name1, name2 string) string {
	return acctest.ConfigCompose(
		testAccGroupDataSourceConfig_base(name1),
		fmt.Sprintf(`
resource "aws_identitystore_group" "test2" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  display_name      = %[1]q
  description       = "Acceptance Test"
}

data "aws_identitystore_group" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  filter {
    attribute_path  = "DisplayName"
    attribute_value = aws_identitystore_group.test.display_name
  }

  group_id = aws_identitystore_group.test2.group_id
}
`, name2),
	)
}

const testAccGroupDataSourceConfig_externalIdConflictsWithUniqueAttribute = `
data "aws_ssoadmin_instances" "test" {}

data "aws_identitystore_group" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  alternate_identifier {
    external_id {
      id     = "test"
      issuer = "test"
    }

    unique_attribute {
      attribute_path  = "DisplayName"
      attribute_value = "does-not-exist"
    }
  }
}
`

func testAccGroupDataSourceConfig_filterConflictsWithUniqueAttribute(name string) string {
	return acctest.ConfigCompose(
		testAccGroupDataSourceConfig_base(name),
		`
data "aws_identitystore_group" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  alternate_identifier {
    unique_attribute {
      attribute_path  = "DisplayName"
      attribute_value = aws_identitystore_group.test.display_name
    }
  }

  filter {
    attribute_path  = "DisplayName"
    attribute_value = aws_identitystore_group.test.display_name
  }
}
`,
	)
}

func testAccGroupDataSourceConfig_groupIdConflictsWithUniqueAttribute(name string) string {
	return acctest.ConfigCompose(
		testAccGroupDataSourceConfig_base(name),
		`
data "aws_identitystore_group" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  alternate_identifier {
    unique_attribute {
      attribute_path  = "DisplayName"
      attribute_value = aws_identitystore_group.test.display_name
    }
  }

  group_id = aws_identitystore_group.test.group_id
}
`,
	)
}

func testAccGroupDataSourceConfig_filterConflictsWithExternalId(name string) string {
	return acctest.ConfigCompose(
		testAccGroupDataSourceConfig_base(name),
		`
data "aws_identitystore_group" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  alternate_identifier {
    external_id {
      id     = "test"
      issuer = "test"
    }
  }

  filter {
    attribute_path  = "DisplayName"
    attribute_value = aws_identitystore_group.test.display_name
  }
}
`,
	)
}

func testAccGroupDataSourceConfig_groupIdConflictsWithExternalId(name string) string {
	return acctest.ConfigCompose(
		testAccGroupDataSourceConfig_base(name),
		`
data "aws_identitystore_group" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  alternate_identifier {
    external_id {
      id     = "test"
      issuer = "test"
    }
  }

  group_id = aws_identitystore_group.test.group_id
}
`,
	)
}

func testAccPreCheckSSOAdminInstances(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminConn()

	var instances []*ssoadmin.InstanceMetadata
	err := conn.ListInstancesPagesWithContext(ctx, &ssoadmin.ListInstancesInput{}, func(page *ssoadmin.ListInstancesOutput, lastPage bool) bool {
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
