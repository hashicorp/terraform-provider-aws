// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package interconnect_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/interconnect/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfinterconnect "github.com/hashicorp/terraform-provider-aws/internal/service/interconnect"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Interconnect connections share a low per-provider quota (for example, GCP
// multicloud connections are limited to 2 per account), so the connection
// acceptance tests are serialized to avoid exceeding it when run in parallel.
func TestAccInterconnectConnection_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Connection": {
			acctest.CtBasic:      testAccInterconnectConnection_basic,
			acctest.CtDisappears: testAccInterconnectConnection_disappears,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccInterconnectConnection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var connection awstypes.Connection
	resourceName := "aws_interconnect_connection.test"
	environmentID := acctest.SkipIfEnvVarNotSet(t, "INTERCONNECT_ENVIRONMENT_ID")
	directConnectGatewayID := acctest.SkipIfEnvVarNotSet(t, "INTERCONNECT_DIRECT_CONNECT_GATEWAY_ID")
	remoteAccountIdentifier := acctest.SkipIfEnvVarNotSet(t, "INTERCONNECT_REMOTE_ACCOUNT_IDENTIFIER")

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InterconnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_basic(environmentID, directConnectGatewayID, remoteAccountIdentifier, "1Gbps"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &connection),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "interconnect", "connection/{id}"),
					resource.TestCheckResourceAttr(resourceName, "bandwidth", "1Gbps"),
					resource.TestCheckResourceAttr(resourceName, "environment_id", environmentID),
					resource.TestCheckResourceAttr(resourceName, "remote_account.0.identifier", remoteAccountIdentifier),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.ConnectionStateRequested)),
					resource.TestCheckResourceAttrSet(resourceName, "activation_key"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateVerifyIgnore:              []string{"remote_account"},
			},
		},
	})
}

func testAccInterconnectConnection_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var connection awstypes.Connection
	resourceName := "aws_interconnect_connection.test"
	environmentID := acctest.SkipIfEnvVarNotSet(t, "INTERCONNECT_ENVIRONMENT_ID")
	directConnectGatewayID := acctest.SkipIfEnvVarNotSet(t, "INTERCONNECT_DIRECT_CONNECT_GATEWAY_ID")
	remoteAccountIdentifier := acctest.SkipIfEnvVarNotSet(t, "INTERCONNECT_REMOTE_ACCOUNT_IDENTIFIER")

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InterconnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_basic(environmentID, directConnectGatewayID, remoteAccountIdentifier, "1Gbps"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &connection),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfinterconnect.ResourceConnection, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckConnectionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).InterconnectClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_interconnect_connection" {
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

func testAccCheckConnectionExists(ctx context.Context, t *testing.T, n string, v *awstypes.Connection) resource.TestCheckFunc {
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

func testAccConnectionConfig_basic(environmentID, directConnectGatewayID, remoteAccountIdentifier, bandwidth string) string {
	return fmt.Sprintf(`
resource "aws_interconnect_connection" "test" {
  bandwidth      = %[4]q
  environment_id = %[1]q

  attach_point {
    direct_connect_gateway = %[2]q
  }

  remote_account {
    identifier = %[3]q
  }
}
`, environmentID, directConnectGatewayID, remoteAccountIdentifier, bandwidth)
}
