// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identitystore_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIdentityStoreGroupDataSource_filterDisplayName(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_identitystore_group.test"
	dataSourceName := "data.aws_identitystore_group.test"
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfig_filterDisplayName(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDisplayName, resourceName, names.AttrDisplayName),
					resource.TestCheckResourceAttrPair(dataSourceName, "group_id", resourceName, "group_id"),
					resource.TestCheckResourceAttr(dataSourceName, "external_ids.#", acctest.Ct0),
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
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfig_uniqueAttributeDisplayName(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDisplayName, resourceName, names.AttrDisplayName),
					resource.TestCheckResourceAttrPair(dataSourceName, "group_id", resourceName, "group_id"),
					resource.TestCheckResourceAttr(dataSourceName, "external_ids.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccIdentityStoreGroupDataSource_filterDisplayNameAndGroupID(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_identitystore_group.test"
	dataSourceName := "data.aws_identitystore_group.test"
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfig_filterDisplayNameAndGroupID(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDisplayName, resourceName, names.AttrDisplayName),
					resource.TestCheckResourceAttrPair(dataSourceName, "group_id", resourceName, "group_id"),
					resource.TestCheckResourceAttr(dataSourceName, "external_ids.#", acctest.Ct0),
				),
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
	return acctest.ConfigCompose(testAccGroupDataSourceConfig_base(name), `
data "aws_identitystore_group" "test" {
  filter {
    attribute_path  = "DisplayName"
    attribute_value = aws_identitystore_group.test.display_name
  }

  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
}
`)
}

func testAccGroupDataSourceConfig_uniqueAttributeDisplayName(name string) string {
	return acctest.ConfigCompose(testAccGroupDataSourceConfig_base(name), `
data "aws_identitystore_group" "test" {
  alternate_identifier {
    unique_attribute {
      attribute_path  = "DisplayName"
      attribute_value = aws_identitystore_group.test.display_name
    }
  }

  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
}
`)
}

func testAccGroupDataSourceConfig_filterDisplayNameAndGroupID(name string) string {
	return acctest.ConfigCompose(testAccGroupDataSourceConfig_base(name), `
data "aws_identitystore_group" "test" {
  filter {
    attribute_path  = "DisplayName"
    attribute_value = aws_identitystore_group.test.display_name
  }

  group_id          = aws_identitystore_group.test.group_id
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
}
`)
}
