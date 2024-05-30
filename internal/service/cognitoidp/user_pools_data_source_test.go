// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCognitoIDPUserPoolsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_cognito_user_pools.test", "arns.#", acctest.Ct2),
					resource.TestCheckResourceAttr("data.aws_cognito_user_pools.test", "ids.#", acctest.Ct2),
					resource.TestCheckResourceAttr("data.aws_cognito_user_pools.empty", "arns.#", acctest.Ct0),
					resource.TestCheckResourceAttr("data.aws_cognito_user_pools.empty", "ids.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccUserPoolsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  count = 2
  name  = %[1]q
}

data "aws_cognito_user_pools" "test" {
  name = %[1]q

  depends_on = [aws_cognito_user_pool.test[0], aws_cognito_user_pool.test[1]]
}

data "aws_cognito_user_pools" "empty" {
  name = "not.%[1]s"

  depends_on = [aws_cognito_user_pool.test[0], aws_cognito_user_pool.test[1]]
}
`, rName)
}
