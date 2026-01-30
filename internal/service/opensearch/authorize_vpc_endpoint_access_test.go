// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearch_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearch/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfopensearch "github.com/hashicorp/terraform-provider-aws/internal/service/opensearch"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchAuthorizeVPCEndpointAccess_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var authorizevpcendpointaccess awstypes.AuthorizedPrincipal
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearch_authorize_vpc_endpoint_access.test"
	domainName := testAccRandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAuthorizeVPCEndpointAccessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizeVPCEndpointAccessConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizeVPCEndpointAccessExists(ctx, t, resourceName, &authorizevpcendpointaccess),
					resource.TestCheckResourceAttrSet(resourceName, "account"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDomainName),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: names.AttrDomainName,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", names.AttrDomainName, "account"),
			},
		},
	})
}

func TestAccOpenSearchAuthorizeVPCEndpointAccess_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var authorizevpcendpointaccess awstypes.AuthorizedPrincipal
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearch_authorize_vpc_endpoint_access.test"
	domainName := testAccRandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAuthorizeVPCEndpointAccessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizeVPCEndpointAccessConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizeVPCEndpointAccessExists(ctx, t, resourceName, &authorizevpcendpointaccess),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfopensearch.ResourceAuthorizeVPCEndpointAccess, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAuthorizeVPCEndpointAccessDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).OpenSearchClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_opensearch_authorize_vpc_endpoint_access" {
				continue
			}

			_, err := tfopensearch.FindAuthorizeVPCEndpointAccessByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrDomainName], rs.Primary.Attributes["account"])
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("OpenSearch Authorize VPC Endpoint Access %s still exists", rs.Primary.Attributes[names.AttrDomainName])
		}

		return nil
	}
}

func testAccCheckAuthorizeVPCEndpointAccessExists(ctx context.Context, t *testing.T, n string, v *awstypes.AuthorizedPrincipal) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).OpenSearchClient(ctx)

		output, err := tfopensearch.FindAuthorizeVPCEndpointAccessByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrDomainName], rs.Primary.Attributes["account"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAuthorizeVPCEndpointAccessConfig_basic(rName, domainName string) string {
	return acctest.ConfigCompose(testAccVPCEndpointConfig_base(rName, domainName), `
data "aws_caller_identity" "current" {}

resource "aws_opensearch_vpc_endpoint" "test" {
  domain_arn = aws_opensearch_domain.test.arn

  vpc_options {
    subnet_ids = aws_subnet.client[*].id
  }
}

resource "aws_opensearch_authorize_vpc_endpoint_access" "test" {
  domain_name = aws_opensearch_domain.test.domain_name
  account     = data.aws_caller_identity.current.account_id
}
`)
}
