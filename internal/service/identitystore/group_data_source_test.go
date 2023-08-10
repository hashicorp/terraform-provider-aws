// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identitystore_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/identitystore"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccIdentityStoreGroupDataSource_uniqueAttributeDisplayName(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_identitystore_group.test"
	dataSourceName := "data.aws_identitystore_group.test"
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
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

func TestAccIdentityStoreGroupDataSource_externalIDConflictsWithUniqueAttribute(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, identitystore.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccGroupDataSourceConfig_externalIDConflictsWithUniqueAttribute,
				ExpectError: regexp.MustCompile(`Invalid combination of arguments`),
			},
		},
	})
}

func TestAccIdentityStoreGroupDataSource_groupIDConflictsWithUniqueAttribute(t *testing.T) {
	ctx := acctest.Context(t)
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, identitystore.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccGroupDataSourceConfig_groupIDConflictsWithUniqueAttribute(name),
				ExpectError: regexp.MustCompile(`Invalid combination of arguments`),
			},
		},
	})
}

func TestAccIdentityStoreGroupDataSource_groupIDConflictsWithExternalID(t *testing.T) {
	ctx := acctest.Context(t)
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, identitystore.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccGroupDataSourceConfig_groupIDConflictsWithExternalID(name),
				ExpectError: regexp.MustCompile(`Invalid combination of arguments`),
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

const testAccGroupDataSourceConfig_externalIDConflictsWithUniqueAttribute = `
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

func testAccGroupDataSourceConfig_groupIDConflictsWithUniqueAttribute(name string) string {
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

func testAccGroupDataSourceConfig_groupIDConflictsWithExternalID(name string) string {
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
	conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminConn(ctx)

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
