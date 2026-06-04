// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package interconnect_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/interconnect/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfinterconnect "github.com/hashicorp/terraform-provider-aws/internal/service/interconnect"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccInterconnectConnectionProposalAcceptor_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var connection awstypes.Connection
	resourceName := "aws_interconnect_connection_proposal_acceptor.test"
	activationKey := acctest.SkipIfEnvVarNotSet(t, "INTERCONNECT_ACTIVATION_KEY")
	directConnectGatewayID := acctest.SkipIfEnvVarNotSet(t, "INTERCONNECT_DIRECT_CONNECT_GATEWAY_ID")

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InterconnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionProposalAcceptorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionProposalAcceptorConfig_basic(activationKey, directConnectGatewayID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionProposalAcceptorExists(ctx, t, resourceName, &connection),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "interconnect", "connection/{id}"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrState),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				// activation_key is a write-only input not returned by the API.
				// billing_tier is assigned by AWS as the connection provisions, so it
				// can change between create and import.
				ImportStateVerifyIgnore: []string{"activation_key", "billing_tier"},
			},
		},
	})
}

func testAccCheckConnectionProposalAcceptorDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).InterconnectClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_interconnect_connection_proposal_acceptor" {
				continue
			}

			_, err := tfinterconnect.FindConnectionByID(ctx, conn, rs.Primary.Attributes[names.AttrID])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Interconnect Connection %s still exists", rs.Primary.Attributes[names.AttrID])
		}

		return nil
	}
}

func testAccCheckConnectionProposalAcceptorExists(ctx context.Context, t *testing.T, n string, v *awstypes.Connection) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).InterconnectClient(ctx)

		output, err := tfinterconnect.FindConnectionByID(ctx, conn, rs.Primary.Attributes[names.AttrID])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccConnectionProposalAcceptorConfig_basic(activationKey, directConnectGatewayID string) string {
	return fmt.Sprintf(`
resource "aws_interconnect_connection_proposal_acceptor" "test" {
  activation_key = %[1]q

  attach_point {
    direct_connect_gateway = %[2]q
  }
}
`, activationKey, directConnectGatewayID)
}
