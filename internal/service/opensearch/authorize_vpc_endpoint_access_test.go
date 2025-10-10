// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearch/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfopensearch "github.com/hashicorp/terraform-provider-aws/internal/service/opensearch"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchAuthorizeVPCEndpointAccess_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var authorizevpcendpointaccess awstypes.AuthorizedPrincipal
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_authorize_vpc_endpoint_access.test"
	domainName := testAccRandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAuthorizeVPCEndpointAccessDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizeVPCEndpointAccessConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizeVPCEndpointAccessExists(ctx, resourceName, &authorizevpcendpointaccess),
					resource.TestCheckResourceAttrSet(resourceName, "account"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDomainName),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateId:                        domainName,
				ImportStateVerifyIdentifierAttribute: names.AttrDomainName,
				ImportStateIdFunc:                    testAccAuthorizeVPCEndpointAccessImportStateIDFunc(resourceName),
			},
		},
	})
}

func TestAccOpenSearchAuthorizeVPCEndpointAccess_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var authorizevpcendpointaccess awstypes.AuthorizedPrincipal
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_authorize_vpc_endpoint_access.test"
	domainName := testAccRandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAuthorizeVPCEndpointAccessDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizeVPCEndpointAccessConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizeVPCEndpointAccessExists(ctx, resourceName, &authorizevpcendpointaccess),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfopensearch.ResourceAuthorizeVPCEndpointAccess, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAuthorizeVPCEndpointAccessDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_opensearch_authorize_vpc_endpoint_access" {
				continue
			}

			_, err := tfopensearch.FindAuthorizeVPCEndpointAccessByName(ctx, conn, rs.Primary.Attributes[names.AttrDomainName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Elastic Beanstalk Application Version %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAuthorizeVPCEndpointAccessExists(ctx context.Context, name string, authorizevpcendpointaccess *awstypes.AuthorizedPrincipal) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.OpenSearch, create.ErrActionCheckingExistence, tfopensearch.ResNameAuthorizeVPCEndpointAccess, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Route53Profiles, create.ErrActionCheckingExistence, tfopensearch.ResNameAuthorizeVPCEndpointAccess, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchClient(ctx)

		resp, err := tfopensearch.FindAuthorizeVPCEndpointAccessByName(ctx, conn, rs.Primary.Attributes[names.AttrDomainName])
		if err != nil {
			return create.Error(names.OpenSearch, create.ErrActionCheckingExistence, tfopensearch.ResNameAuthorizeVPCEndpointAccess, rs.Primary.ID, err)
		}

		*authorizevpcendpointaccess = *resp

		return nil
	}
}

func testAccAuthorizeVPCEndpointAccessImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes[names.AttrDomainName], nil
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
