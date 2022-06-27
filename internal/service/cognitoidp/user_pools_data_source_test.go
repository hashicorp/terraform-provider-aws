package cognitoidp_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccCognitoIDPUserPoolsDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_cognito_user_pools.test", "arns.#", "2"),
					resource.TestCheckResourceAttr("data.aws_cognito_user_pools.test", "ids.#", "2"),
					resource.TestCheckResourceAttr("data.aws_cognito_user_pools.empty", "arns.#", "0"),
					resource.TestCheckResourceAttr("data.aws_cognito_user_pools.empty", "ids.#", "0"),
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
