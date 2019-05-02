package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsCognitoUserPoolClient_basic(t *testing.T) {
	rPoolName := fmt.Sprintf("tf_acc_ds_cognito_user_pools_%s", acctest.RandString(7))
	rClientName := fmt.Sprintf("tf_acc_ds_cognito_user_pools_%s", acctest.RandString(7))
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsCognitoUserPoolClientConfig_basic(rPoolName, rClientName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_cognito_user_pools.selected", "ids.#", "2"),
					resource.TestCheckResourceAttr("data.aws_cognito_user_pools.selected", "arns.#", "2"),
					// check that client_id is set
					// check that client_secret is set
				),
			},
			{
				Config:      testAccDataSourceAwsCognitoUserPoolClientConfig_notFound(rPoolName, rClientName),
				ExpectError: regexp.MustCompile(`No cognito user pool client found with name:`),
			},
		},
	})
}

func testAccDataSourceAwsCognitoUserPoolClientConfig_basic(rPoolName string, rClientName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "main" {
	name  = "%s"
}

resource "aws_cognito_user_pool_client" "main" {
	name         = "%s"
  user_pool_id = "${aws_cognito_user_pool.pool.id}"
}

data "aws_cognito_user_pool_client" "selected" {
	user_pool_id = "${aws_cognito_user_pool.main.id}"
	name         = "%s"
}
`, rPoolName, rClientName, rClientName)
}

func testAccDataSourceAwsCognitoUserPoolClientConfig_notFound(rPoolName string, rClientName string) string {
	return fmt.Sprintf(`
data "aws_cognito_user_pool_client" "selected" {
	name = "%s-not-found"
}
`, rPoolName)
}
