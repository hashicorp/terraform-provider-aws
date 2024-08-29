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

func TestAccCognitoIDPUserPoolClientsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	datasourceName := "data.aws_cognito_user_pool_clients.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientsDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "client_ids.#", acctest.Ct3),
					resource.TestCheckResourceAttr(datasourceName, "client_names.#", acctest.Ct3),
				),
			},
		},
	})
}

func testAccUserPoolClientsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool_client" "test" {
  count        = 3
  name         = "client${count.index}"
  user_pool_id = aws_cognito_user_pool.test.id
}

data "aws_cognito_user_pool_clients" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
  depends_on   = [aws_cognito_user_pool_client.test]
}
 `, rName)
}
