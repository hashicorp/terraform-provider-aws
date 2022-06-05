package directconnect_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/directconnect"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

type testAccDxHostedConnectionEnv struct {
	ConnectionId   string
	OwnerAccountId string
}

func TestAccDirectConnectHostedConnection_basic(t *testing.T) {
	env, err := testAccCheckHostedConnectionEnv()
	if err != nil {
		acctest.Skip(t, err.Error())
	}

	connectionName := fmt.Sprintf("tf-dx-%s", sdkacctest.RandString(5))
	resourceName := "aws_dx_hosted_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckHostedConnectionDestroy(testAccHostedConnectionProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccHostedConnectionConfig_basic(connectionName, env.ConnectionId, env.OwnerAccountId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", connectionName),
					resource.TestCheckResourceAttr(resourceName, "connection_id", env.ConnectionId),
					resource.TestCheckResourceAttr(resourceName, "owner_account_id", env.OwnerAccountId),
					resource.TestCheckResourceAttr(resourceName, "bandwidth", "100Mbps"),
					resource.TestCheckResourceAttr(resourceName, "vlan", "4094"),
				),
			},
		},
	})
}

func testAccCheckHostedConnectionEnv() (*testAccDxHostedConnectionEnv, error) {
	result := &testAccDxHostedConnectionEnv{
		ConnectionId:   os.Getenv("TEST_AWS_DX_CONNECTION_ID"),
		OwnerAccountId: os.Getenv("TEST_AWS_DX_OWNER_ACCOUNT_ID"),
	}

	if result.ConnectionId == "" || result.OwnerAccountId == "" {
		return nil, errors.New("TEST_AWS_DX_CONNECTION_ID and TEST_AWS_DX_OWNER_ACCOUNT_ID must be set for tests involving hosted connections")
	}

	return result, nil
}

func testAccCheckHostedConnectionDestroy(providerFunc func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider := providerFunc()
		conn := provider.Meta().(*conns.AWSClient).DirectConnectConn

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dx_hosted_connection" {
				continue
			}

			_, err := tfdirectconnect.FindHostedConnectionByID(conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Direct Connect Hosted Connection %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckHostedConnectionExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DirectConnectConn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Direct Connect Hosted Connection ID is set")
		}

		_, err := tfdirectconnect.FindHostedConnectionByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccHostedConnectionConfig_basic(name, connectionId, ownerAccountId string) string {
	return fmt.Sprintf(`
resource "aws_dx_hosted_connection" "test" {
  name             = "%s"
  connection_id    = "%s"
  owner_account_id = "%s"
  bandwidth        = "100Mbps"
  vlan             = 4094
}
`, name, connectionId, ownerAccountId)
}

func testAccHostedConnectionProvider() *schema.Provider {
	return acctest.Provider
}
