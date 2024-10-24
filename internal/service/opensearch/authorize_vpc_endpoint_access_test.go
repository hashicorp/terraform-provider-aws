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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfopensearch "github.com/hashicorp/terraform-provider-aws/internal/service/opensearch"
)

func TestAccOpenSearchAuthorizeVpcEndpointAccess_basic(t *testing.T) {
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
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAuthorizeVpcEndpointAccessDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizeVpcEndpointAccessConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizeVpcEndpointAccessExists(ctx, resourceName, &authorizevpcendpointaccess),
					resource.TestCheckResourceAttrSet(resourceName, "account"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDomainName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOpenSearchAuthorizeVpcEndpointAccess_disappears(t *testing.T) {
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
		CheckDestroy:             testAccCheckAuthorizeVpcEndpointAccessDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizeVpcEndpointAccessConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizeVpcEndpointAccessExists(ctx, resourceName, &authorizevpcendpointaccess),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfopensearch.resourceAuthorizeVpcEndpointAccess, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAuthorizeVpcEndpointAccessDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_opensearch_authorize_vpc_endpoint_access" {
				continue
			}

			_, err := tfopensearch.FindAuthorizeVpcEndpointAccessByName(ctx, conn, rs.Primary.Attributes["domain_name"])

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

func testAccCheckAuthorizeVpcEndpointAccessExists(ctx context.Context, name string, authorizevpcendpointaccess *awstypes.AuthorizedPrincipal) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.OpenSearch, create.ErrActionCheckingExistence, tfopensearch.ResNameAuthorizeVpcEndpointAccess, name, errors.New("not found"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchClient(ctx)

		resp, err := tfopensearch.FindAuthorizeVpcEndpointAccessByName(ctx, conn, rs.Primary.Attributes["authorized_principal"])

		if err != nil {
			return create.Error(names.OpenSearch, create.ErrActionCheckingExistence, tfopensearch.ResNameAuthorizeVpcEndpointAccess, rs.Primary.ID, err)
		}

		*authorizevpcendpointaccess = *resp

		return nil
	}
}

// ``func testAccPreCheck(ctx context.Context, t *testing.T) {
// 	conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchClient(ctx)

// 	input := &opensearch.ListAuthorizeVpcEndpointAccesssInput{}
// 	_, err := conn.ListAuthorizeVpcEndpointAccesss(ctx, input)

// 	if acctest.PreCheckSkipError(err) {
// 		t.Skipf("skipping acceptance testing: %s", err)
// 	}
// 	if err != nil {
// 		t.Fatalf("unexpected PreCheck error: %s", err)
// 	}
// }

// func testAccCheckAuthorizeVpcEndpointAccessNotRecreated(before, after *opensearch.DescribeAuthorizeVpcEndpointAccessResponse) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		if before, after := aws.ToString(before.AuthorizeVpcEndpointAccessId), aws.ToString(after.AuthorizeVpcEndpointAccessId); before != after {
// 			return create.Error(names.OpenSearch, create.ErrActionCheckingNotRecreated, tfopensearch.ResNameAuthorizeVpcEndpointAccess, aws.ToString(before.AuthorizeVpcEndpointAccessId), errors.New("recreated"))
// 		}

// 		return nil
// 	}
// }

func testAccAuthorizeVpcEndpointAccessConfig_basic(rName, domainName string) string {
	return acctest.ConfigCompose(testAccVPCEndpointConfig_base(rName, domainName), `
data "aws_caller_identity" "current" {}

resource "aws_opensearch_vpc_endpoint" "test" {
  domain_arn = aws_opensearch_domain.test.arn

  vpc_options {
    subnet_ids = aws_subnet.client[*].id
  }
}

resource "aws_opensearch_authorize_vpc_endpoint_access" "test" {
  domain_name = aws_opensearch_domain.name
  account = data.aws_caller_identity.current.account_id

}
`)
}
