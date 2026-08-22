// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package interconnect_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/interconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/interconnect/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// TestAccInterconnectConnectionProposalAcceptor_basic accepts a Connection proposal.
//
// An Activation Key is generated on the Interconnect partner's portal when the remote
// side requests a connection, so there is no AWS API that can produce one for a test.
// The key must therefore be supplied out of band, and is single use: accepting it
// consumes the proposal. Set the following to run this test:
//
//	INTERCONNECT_ACTIVATION_KEY            Activation Key from the partner's portal
//	INTERCONNECT_DIRECT_CONNECT_GATEWAY_ID Direct Connect Gateway to attach to
//	INTERCONNECT_REGION                    (optional) Region holding the proposal
func TestAccInterconnectConnectionProposalAcceptor_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var connection awstypes.Connection
	resourceName := "aws_interconnect_connection_proposal_acceptor.test"
	activationKey := acctest.SkipIfEnvVarNotSet(t, "INTERCONNECT_ACTIVATION_KEY")
	directConnectGatewayID := acctest.SkipIfEnvVarNotSet(t, "INTERCONNECT_DIRECT_CONNECT_GATEWAY_ID")
	region := testAccRegion()

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InterconnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionProposalAcceptorDestroy(ctx, t, region),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionProposalAcceptorConfig_basic(activationKey, directConnectGatewayID, region),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionProposalAcceptorExists(ctx, t, resourceName, region, &connection),
					// The ARN is matched with a regular expression rather than using
					// acctest.CheckResourceAttrRegionalARNFormat, which builds the expected
					// ARN from the provider's Region and so cannot describe a resource that
					// sets the region argument.
					resource.TestMatchResourceAttr(resourceName, names.AttrARN,
						regexache.MustCompile(fmt.Sprintf(`^arn:[^:]+:interconnect:%s:\d{12}:connection/.+$`, region))),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrState),
					resource.TestCheckResourceAttrSet(resourceName, "bandwidth"),
					resource.TestCheckResourceAttrSet(resourceName, "environment_id"),
					resource.TestCheckResourceAttrSet(resourceName, "interconnect_provider"),
					resource.TestCheckResourceAttr(resourceName, "attach_point.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "attach_point.0.direct_connect_gateway", directConnectGatewayID),
					resource.TestCheckResourceAttr(resourceName, names.AttrRegion, region),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				// activation_key is a write-only input that the API does not return.
				// billing_tier is assigned by AWS as the connection provisions, so it
				// can change between create and import.
				ImportStateVerifyIgnore: []string{"activation_key", "billing_tier"},
			},
		},
	})
}

// testAccFindConnectionByID looks up a Connection in the given Region, mirroring the
// behaviour of the package's own finder. The exported finder uses the client's configured
// Region, whereas this test targets the Region holding the proposal.
func testAccFindConnectionByID(ctx context.Context, t *testing.T, region, id string) (*awstypes.Connection, error) {
	t.Helper()

	conn := acctest.ProviderMeta(ctx, t).InterconnectClient(ctx)

	input := interconnect.GetConnectionInput{
		Identifier: aws.String(id),
	}
	out, err := conn.GetConnection(ctx, &input, func(o *interconnect.Options) {
		o.Region = region
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{LastError: err}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Connection == nil {
		return nil, &retry.NotFoundError{}
	}

	// A deleted Connection lingers in the API in the "deleted" state rather than
	// returning a not-found error. Treat it as not found.
	if out.Connection.State == awstypes.ConnectionStateDeleted {
		return nil, &retry.NotFoundError{
			LastError: fmt.Errorf("Interconnect Connection (%s) in state %q", id, out.Connection.State),
		}
	}

	return out.Connection, nil
}

func testAccCheckConnectionProposalAcceptorDestroy(ctx context.Context, t *testing.T, region string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_interconnect_connection_proposal_acceptor" {
				continue
			}

			_, err := testAccFindConnectionByID(ctx, t, region, rs.Primary.Attributes[names.AttrID])

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

func testAccCheckConnectionProposalAcceptorExists(ctx context.Context, t *testing.T, n, region string, v *awstypes.Connection) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		output, err := testAccFindConnectionByID(ctx, t, region, rs.Primary.Attributes[names.AttrID])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccConnectionProposalAcceptorConfig_basic(activationKey, directConnectGatewayID, region string) string {
	return fmt.Sprintf(`
resource "aws_interconnect_connection_proposal_acceptor" "test" {
  activation_key = %[1]q
  region         = %[3]q

  attach_point {
    direct_connect_gateway = %[2]q
  }
}
`, activationKey, directConnectGatewayID, region)
}
