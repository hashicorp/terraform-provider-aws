// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDirectConnectHostedConnection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	connectionID := acctest.SkipIfEnvVarNotSet(t, "TEST_AWS_DX_CONNECTION_ID")
	ownerAccountID := acctest.SkipIfEnvVarNotSet(t, "TEST_AWS_DX_OWNER_ACCOUNT_ID")

	connectionName := fmt.Sprintf("tf-dx-%s", sdkacctest.RandString(5))
	resourceName := "aws_dx_hosted_connection.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHostedConnectionDestroy(ctx, func() *schema.Provider { return acctest.Provider }),
		Steps: []resource.TestStep{
			{
				Config: testAccHostedConnectionConfig_basic(connectionName, connectionID, ownerAccountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedConnectionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "bandwidth", "100Mbps"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionID),
					resource.TestCheckResourceAttrSet(resourceName, "connection_region"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, connectionName),
					resource.TestCheckResourceAttr(resourceName, names.AttrOwnerAccountID, ownerAccountID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRegion),
					resource.TestCheckResourceAttr(resourceName, "vlan", "4094"),
				),
			},
		},
	})
}

func testAccCheckHostedConnectionDestroy(ctx context.Context, providerFunc func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider := providerFunc()
		conn := provider.Meta().(*conns.AWSClient).DirectConnectClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dx_hosted_connection" {
				continue
			}

			// Get the parent connection ID from the resource attributes
			parentConnectionID := rs.Primary.Attributes[names.AttrConnectionID]
			_, err := tfdirectconnect.FindHostedConnectionByID(ctx, conn, rs.Primary.ID, parentConnectionID)

			if retry.NotFound(err) {
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

func testAccCheckHostedConnectionExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DirectConnectClient(ctx)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		// Get the parent connection ID from the resource state
		parentConnectionID := rs.Primary.Attributes[names.AttrConnectionID]

		_, err := tfdirectconnect.FindHostedConnectionByID(ctx, conn, rs.Primary.ID, parentConnectionID)

		return err
	}
}

func testAccHostedConnectionConfig_basic(name, connectionID, ownerAccountID string) string {
	return fmt.Sprintf(`
resource "aws_dx_hosted_connection" "test" {
  name             = %[1]q
  connection_id    = %[2]q
  owner_account_id = %[3]q
  bandwidth        = "100Mbps"
  vlan             = 4094
}
`, name, connectionID, ownerAccountID)
}
