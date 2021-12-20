package cognitoidp_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccCognitoIDPUserPoolClientsDataSource_basic(t *testing.T) {
	testName := fmt.Sprintf("tf_acc_ds_cognito_user_pools_%s", sdkacctest.RandString(7))
	resourceName := "aws_cognito_user_pool.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientsDataSource_basic(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, nil),
					resource.TestCheckResourceAttr("data.aws_cognito_user_pool_clients.test", "client_ids.#", "3"),
				),
			},
		},
	})
}

func testAccUserPoolClientsDataSource_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = "%s"
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
