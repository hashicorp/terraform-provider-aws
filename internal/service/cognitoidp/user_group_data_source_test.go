// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cognitoidp_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCognitoIDPUserGroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_cognito_user_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckIdentityProvider(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDescription, "test"),
				),
			},
		},
	})
}

func testAccUserGroupDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_group" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
  description  = "test"
}

data "aws_cognito_user_group" "test" {
  name         = aws_cognito_user_group.test.name
  user_pool_id = aws_cognito_user_group.test.user_pool_id
}
`, rName)
}
