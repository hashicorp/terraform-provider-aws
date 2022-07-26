package directconnect_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
)

func TestAccDirectConnectConnectionConfirmation_basic(t *testing.T) {
	env, err := testAccCheckHostedConnectionEnv()
	if err != nil {
		acctest.Skip(t, err.Error())
	}

	var providers []*schema.Provider

	connectionName := fmt.Sprintf("tf-dx-%s", sdkacctest.RandString(5))
	resourceName := "aws_dx_connection_confirmation.test"
	providerFunc := testAccConnectionConfirmationProvider(&providers, 0)
	altProviderFunc := testAccConnectionConfirmationProvider(&providers, 1)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckHostedConnectionDestroy(altProviderFunc),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfirmationConfig_basic(connectionName, env.ConnectionId, env.OwnerAccountId),
				Check:  testAccCheckConnectionConfirmationExists(resourceName, providerFunc),
			},
		},
	})
}

func testAccCheckConnectionConfirmationExists(name string, providerFunc func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Direct Connect Connection ID is set")
		}

		provider := providerFunc()
		conn := provider.Meta().(*conns.AWSClient).DirectConnectConn

		connection, err := tfdirectconnect.FindConnectionByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if state := aws.StringValue(connection.ConnectionState); state != directconnect.ConnectionStateAvailable {
			return fmt.Errorf("Direct Connect Connection %s in unexpected state: %s", rs.Primary.ID, state)
		}

		return nil
	}
}

func testAccConnectionConfirmationConfig_basic(name, connectionId, ownerAccountId string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		fmt.Sprintf(`
resource "aws_dx_hosted_connection" "connection" {
  provider = "awsalternate"

  name             = "%s"
  connection_id    = "%s"
  owner_account_id = "%s"
  bandwidth        = "100Mbps"
  vlan             = 4092
}

resource "aws_dx_connection_confirmation" "test" {
  connection_id = aws_dx_hosted_connection.connection.id
}
`, name, connectionId, ownerAccountId))
}

func testAccConnectionConfirmationProvider(providers *[]*schema.Provider, index int) func() *schema.Provider {
	return func() *schema.Provider {
		return (*providers)[index]
	}
}
