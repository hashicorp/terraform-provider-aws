// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cognitoidp_test

import (
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCognitoIDPUserDataSource_selectors(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	email := acctest.RandomEmailAddress(acctest.RandomDomainName(t))
	emailDataSourceName := "data.aws_cognito_user.by_email"
	usernameDataSourceName := "data.aws_cognito_user.by_username"
	resourceName := "aws_cognito_user.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig_selectors(rName, email),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(emailDataSourceName, names.AttrUserPoolID, resourceName, names.AttrUserPoolID),
					resource.TestCheckResourceAttrPair(emailDataSourceName, names.AttrEmail, resourceName, "attributes.email"),
					resource.TestCheckResourceAttrPair(emailDataSourceName, names.AttrUsername, resourceName, names.AttrUsername),
					resource.TestCheckResourceAttrPair(emailDataSourceName, "attributes.email", resourceName, "attributes.email"),
					resource.TestCheckResourceAttr(emailDataSourceName, "attributes.email_verified", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(emailDataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(emailDataSourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttrSet(emailDataSourceName, "last_modified_date"),
					resource.TestCheckResourceAttrSet(emailDataSourceName, "sub"),
					resource.TestCheckResourceAttr(emailDataSourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(emailDataSourceName, "mfa_setting_list.#", "0"),
					resource.TestCheckResourceAttr(emailDataSourceName, names.AttrStatus, string(awstypes.UserStatusTypeForceChangePassword)),
					resource.TestCheckResourceAttrPair(usernameDataSourceName, names.AttrUsername, resourceName, names.AttrUsername),
					resource.TestCheckResourceAttrPair(usernameDataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(usernameDataSourceName, "attributes.email", resourceName, "attributes.email"),
					resource.TestCheckResourceAttr(usernameDataSourceName, names.AttrStatus, string(awstypes.UserStatusTypeForceChangePassword)),
				),
			},
		},
	})
}

func testAccUserDataSourceConfig_selectors(rName, email string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name             = %[1]q
  alias_attributes = ["email"]
}

resource "aws_cognito_user" "test" {
  user_pool_id   = aws_cognito_user_pool.test.id
  username       = %[1]q
  message_action = "SUPPRESS"

  attributes = {
    email          = %[2]q
    email_verified = "false"
  }
}

data "aws_cognito_user" "by_email" {
  user_pool_id = aws_cognito_user.test.user_pool_id
  email        = aws_cognito_user.test.attributes.email
}

data "aws_cognito_user" "by_username" {
  user_pool_id = aws_cognito_user.test.user_pool_id
  username     = aws_cognito_user.test.username
}
`, rName, email)
}
