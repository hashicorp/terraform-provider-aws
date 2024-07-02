// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccCognitoIDPUserGroupsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_cognito_user_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckIdentityProvider(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, cognitoidentityprovider.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "groups.#", acctest.Ct2),
				),
			},
		},
	})
}

func testAccUserGroupsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %q
}

resource "aws_cognito_user_group" "test_1" {
  name         = "%s-1"
  user_pool_id = aws_cognito_user_pool.test.id
  description  = "test 1"
}
resource "aws_cognito_user_group" "test_2" {
  name         = "%s-2"
  user_pool_id = aws_cognito_user_pool.test.id
  description  = "test 2"
}

data "aws_cognito_user_groups" "test" {
  user_pool_id = aws_cognito_user_group.test_1.user_pool_id
}
`, rName, rName, rName)
}
