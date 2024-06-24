// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccUserDataSource_userID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	domain := acctest.RandomDomainName()
	email := acctest.RandomEmailAddress(domain)
	resourceName := "aws_connect_user.test"
	datasourceName := "data.aws_connect_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig_id(rName, rName2, rName3, rName4, rName5, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "directory_user_id", resourceName, "directory_user_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_group_id", resourceName, "hierarchy_group_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "identity_info.#", resourceName, "identity_info.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "identity_info.0.email", resourceName, "identity_info.0.email"),
					resource.TestCheckResourceAttrPair(datasourceName, "identity_info.0.first_name", resourceName, "identity_info.0.first_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "identity_info.0.last_name", resourceName, "identity_info.0.last_name"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceID, resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, "phone_config.#", resourceName, "phone_config.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "phone_config.0.after_contact_work_time_limit", resourceName, "phone_config.0.after_contact_work_time_limit"),
					resource.TestCheckResourceAttrPair(datasourceName, "phone_config.0.auto_accept", resourceName, "phone_config.0.auto_accept"),
					resource.TestCheckResourceAttrPair(datasourceName, "phone_config.0.desk_phone_number", resourceName, "phone_config.0.desk_phone_number"),
					resource.TestCheckResourceAttrPair(datasourceName, "phone_config.0.phone_type", resourceName, "phone_config.0.phone_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "routing_profile_id", resourceName, "routing_profile_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "security_profile_ids.#", resourceName, "security_profile_ids.#"),
					resource.TestCheckTypeSetElemAttrPair(datasourceName, "security_profile_ids.*", resourceName, "security_profile_ids.0"),
					resource.TestCheckTypeSetElemAttrPair(datasourceName, "security_profile_ids.*", resourceName, "security_profile_ids.1"),
					resource.TestCheckResourceAttrPair(datasourceName, "user_id", resourceName, "user_id"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Key1", resourceName, "tags.Key1"),
				),
			},
		},
	})
}

func testAccUserDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	domain := acctest.RandomDomainName()
	email := acctest.RandomEmailAddress(domain)
	resourceName := "aws_connect_user.test"
	datasourceName := "data.aws_connect_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig_name(rName, rName2, rName3, rName4, rName5, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "directory_user_id", resourceName, "directory_user_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_group_id", resourceName, "hierarchy_group_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "identity_info.#", resourceName, "identity_info.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "identity_info.0.email", resourceName, "identity_info.0.email"),
					resource.TestCheckResourceAttrPair(datasourceName, "identity_info.0.first_name", resourceName, "identity_info.0.first_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "identity_info.0.last_name", resourceName, "identity_info.0.last_name"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceID, resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, "phone_config.#", resourceName, "phone_config.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "phone_config.0.after_contact_work_time_limit", resourceName, "phone_config.0.after_contact_work_time_limit"),
					resource.TestCheckResourceAttrPair(datasourceName, "phone_config.0.auto_accept", resourceName, "phone_config.0.auto_accept"),
					resource.TestCheckResourceAttrPair(datasourceName, "phone_config.0.desk_phone_number", resourceName, "phone_config.0.desk_phone_number"),
					resource.TestCheckResourceAttrPair(datasourceName, "phone_config.0.phone_type", resourceName, "phone_config.0.phone_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "routing_profile_id", resourceName, "routing_profile_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "security_profile_ids.#", resourceName, "security_profile_ids.#"),
					resource.TestCheckTypeSetElemAttrPair(datasourceName, "security_profile_ids.*", resourceName, "security_profile_ids.0"),
					resource.TestCheckTypeSetElemAttrPair(datasourceName, "security_profile_ids.*", resourceName, "security_profile_ids.1"),
					resource.TestCheckResourceAttrPair(datasourceName, "user_id", resourceName, "user_id"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Key1", resourceName, "tags.Key1"),
				),
			},
		},
	})
}

func testAccUserBaseDataSourceConfig(rName, rName2, rName3, rName4, rName5, email string) string {
	return acctest.ConfigCompose(
		testAccUserConfig_base(rName, rName2, rName3, rName4),
		fmt.Sprintf(`
resource "aws_connect_user" "test" {
  instance_id        = aws_connect_instance.test.id
  name               = %[1]q
  password           = "Password123"
  routing_profile_id = data.aws_connect_routing_profile.test.routing_profile_id
  hierarchy_group_id = aws_connect_user_hierarchy_group.parent.hierarchy_group_id

  security_profile_ids = [
    data.aws_connect_security_profile.agent.security_profile_id,
    data.aws_connect_security_profile.call_center_manager.security_profile_id,
  ]

  identity_info {
    email      = %[2]q
    first_name = "example"
    last_name  = "example2"
  }

  phone_config {
    after_contact_work_time_limit = 0
    auto_accept                   = true
    desk_phone_number             = "+112345678913"
    phone_type                    = "DESK_PHONE"
  }

  tags = {
    "Key1" = "Value1",
  }
}
`, rName5, email))
}

func testAccUserDataSourceConfig_id(rName, rName2, rName3, rName4, rName5, email string) string {
	return acctest.ConfigCompose(
		testAccUserBaseDataSourceConfig(rName, rName2, rName3, rName4, rName5, email),
		`
data "aws_connect_user" "test" {
  instance_id = aws_connect_instance.test.id
  user_id     = aws_connect_user.test.user_id
}
`)
}

func testAccUserDataSourceConfig_name(rName, rName2, rName3, rName4, rName5, email string) string {
	return acctest.ConfigCompose(
		testAccUserBaseDataSourceConfig(rName, rName2, rName3, rName4, rName5, email),
		`
data "aws_connect_user" "test" {
  instance_id = aws_connect_instance.test.id
  name        = aws_connect_user.test.name
}
`)
}
