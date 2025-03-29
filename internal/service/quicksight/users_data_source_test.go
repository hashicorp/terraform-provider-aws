// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/quicksight"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccQuickSightUsersDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_quicksight_user.user"
	dataSourceName := "data.aws_quicksight_users.users"
	dataSourceNameFilterByActive := "data.aws_quicksight_users.filter_by_active"
	dataSourceNameFilterByEmail := "data.aws_quicksight_users.filter_by_email"
	dataSourceNameFilterByIdentityType := "data.aws_quicksight_users.filter_by_identity_type"
	dataSourceNameFilterByRole := "data.aws_quicksight_users.filter_by_role"
	dataSourceNameFilterByUser := "data.aws_quicksight_users.filter_by_user"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, quicksight.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUsersDataSourceBase(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "aws_account_id", resourceName, "aws_account_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "namespace", resourceName, "namespace"),
					resource.TestCheckResourceAttr(dataSourceNameFilterByActive, "users.0.active", "false"),
					resource.TestCheckResourceAttr(dataSourceNameFilterByIdentityType, "users.0.identity_type", "QUICKSIGHT"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceNameFilterByIdentityType, "users.*.identity_type", resourceName, "identity_type"),
					resource.TestCheckResourceAttr(dataSourceNameFilterByEmail, "users.0.email", acctest.DefaultEmailAddress),
					resource.TestCheckTypeSetElemAttrPair(dataSourceNameFilterByEmail, "users.*.email", resourceName, "email"),
					resource.TestCheckResourceAttr(dataSourceNameFilterByUser, "users.0.user_name", rName),
					resource.TestCheckTypeSetElemAttrPair(dataSourceNameFilterByUser, "users.*.user_name", resourceName, "user_name"),
					resource.TestCheckResourceAttr(dataSourceNameFilterByUser, "users.0.user_role", "ADMIN"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceNameFilterByRole, "users.*.user_role", resourceName, "user_role"),
				),
			},
		},
	})
}

func testAccUsersDataSourceBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_quicksight_user" "user" {
  user_name     = %[1]q
  email         = %[2]q
  identity_type = "QUICKSIGHT"
  user_role     = "ADMIN"
}

data "aws_quicksight_users" "users" {
	depends_on = [aws_quicksight_user.user]
}

data "aws_quicksight_users" "filter_by_active" {
	filter {
		user_name_regex = %[1]q
		email_regex = %[2]q
		active = false
	}
	depends_on = [aws_quicksight_user.user]
}

data "aws_quicksight_users" "filter_by_email" {
	filter {
		email_regex = %[2]q
	}
	depends_on = [aws_quicksight_user.user]
}

data "aws_quicksight_users" "filter_by_identity_type" {
	filter {
		identity_type = "QUICKSIGHT"
	}
	depends_on = [aws_quicksight_user.user]
}

data "aws_quicksight_users" "filter_by_role" {
	filter {
		user_name_regex = %[1]q
		email_regex = %[2]q
		user_role = "ADMIN"
	}
	depends_on = [aws_quicksight_user.user]
}

data "aws_quicksight_users" "filter_by_user" {
	filter {
		user_name_regex = %[1]q
	}
	depends_on = [aws_quicksight_user.user]
}

`, rName, acctest.DefaultEmailAddress)
}
