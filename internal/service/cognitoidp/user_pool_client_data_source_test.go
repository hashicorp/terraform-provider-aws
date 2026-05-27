// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cognitoidp_test

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCognitoIDPUserPoolClientDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.UserPoolClientType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "data.aws_cognito_user_pool_client.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, t, resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "explicit_auth_flows.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "explicit_auth_flows.*", "ADMIN_NO_SRP_AUTH"),
					resource.TestCheckResourceAttr(resourceName, "token_validity_units.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "analytics_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "refresh_token_rotation.#", "0"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClientDataSource_refreshTokenRotation(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.UserPoolClientType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "data.aws_cognito_user_pool_client.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientDataSourceConfig_refreshTokenRotation(rName, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, t, resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "explicit_auth_flows.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "explicit_auth_flows.*", "ADMIN_NO_SRP_AUTH"),
					resource.TestCheckResourceAttr(resourceName, "token_validity_units.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "analytics_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "refresh_token_rotation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "refresh_token_rotation.0.feature", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "refresh_token_rotation.0.retry_grace_period_seconds", "10"),
				),
			},
		},
	})
}

func testAccUserPoolClientDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccUserPoolClientConfig_basic(rName), `
data "aws_cognito_user_pool_client" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
  client_id    = aws_cognito_user_pool_client.test.id
}
`)
}

func testAccUserPoolClientDataSourceConfig_refreshTokenRotation(rName string, retryGracePeriodSeconds int32) string {
	return acctest.ConfigCompose(testAccUserPoolClientConfig_refreshTokenRotation(rName, retryGracePeriodSeconds), `
data "aws_cognito_user_pool_client" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
  client_id    = aws_cognito_user_pool_client.test.id
}
`)
}
