// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDirectConnectConnectionConfirmation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	connectionID := acctest.SkipIfEnvVarNotSet(t, "TEST_AWS_DX_CONNECTION_ID")
	ownerAccountID := acctest.SkipIfEnvVarNotSet(t, "TEST_AWS_DX_OWNER_ACCOUNT_ID")

	var providers []*schema.Provider
	connectionName := fmt.Sprintf("tf-dx-%s", sdkacctest.RandString(5))
	resourceName := "aws_dx_connection_confirmation.test"
	providerFunc := testAccConnectionConfirmationProvider(&providers, 0)
	altProviderFunc := testAccConnectionConfirmationProvider(&providers, 1)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             testAccCheckHostedConnectionDestroy(ctx, altProviderFunc),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfirmationConfig_basic(connectionName, connectionID, ownerAccountID),
				Check:  testAccCheckConnectionConfirmationExists(ctx, t, resourceName, providerFunc),
			},
		},
	})
}

func testAccCheckConnectionConfirmationExists(ctx context.Context, _ *testing.T, n string, providerFunc func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		provider := providerFunc()
		conn := provider.Meta().(*conns.AWSClient).DirectConnectClient(ctx)

		connection, err := tfdirectconnect.FindConnectionByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if connection.ConnectionState != awstypes.ConnectionStateAvailable {
			return fmt.Errorf("Direct Connect Connection %s in unexpected state: %s", rs.Primary.ID, string(connection.ConnectionState))
		}

		return nil
	}
}

func testAccConnectionConfirmationConfig_basic(name, connectionID, ownerAccountID string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		fmt.Sprintf(`
resource "aws_dx_hosted_connection" "connection" {
  provider = "awsalternate"

  name             = %[1]q
  connection_id    = %[2]q
  owner_account_id = %[3]q
  bandwidth        = "100Mbps"
  vlan             = 4092
}

resource "aws_dx_connection_confirmation" "test" {
  connection_id = aws_dx_hosted_connection.connection.id
}
`, name, connectionID, ownerAccountID))
}

func testAccConnectionConfirmationProvider(providers *[]*schema.Provider, index int) func() *schema.Provider {
	return func() *schema.Provider {
		return (*providers)[index]
	}
}
