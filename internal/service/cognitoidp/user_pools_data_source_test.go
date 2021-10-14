package cognitoidp_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccDataSourceAwsCognitoUserPools_basic(t *testing.T) {
	rName := fmt.Sprintf("tf_acc_ds_cognito_user_pools_%s", sdkacctest.RandString(7))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck: acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_cognito_user_pools.selected", "ids.#", "2"),
					resource.TestCheckResourceAttr("data.aws_cognito_user_pools.selected", "arns.#", "2"),
				),
			},
			{
				Config:      testAccUserPoolsDataSourceConfig_notFound(rName),
				ExpectError: regexp.MustCompile(`No cognito user pool found with name:`),
			},
		},
	})
}

func testAccUserPoolsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "main" {
  count = 2
  name  = "%s"
}

data "aws_cognito_user_pools" "selected" {
  name = aws_cognito_user_pool.main.*.name[0]
}
`, rName)
}

func testAccUserPoolsDataSourceConfig_notFound(rName string) string {
	return fmt.Sprintf(`
data "aws_cognito_user_pools" "selected" {
  name = "%s-not-found"
}
`, rName)
}
