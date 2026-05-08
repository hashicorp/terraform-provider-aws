// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package codestarconnections_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/codestarconnections/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcodestarconnections "github.com/hashicorp/terraform-provider-aws/internal/service/codestarconnections"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCodeStarConnectionsConnection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.Connection
	resourceName := "aws_codestarconnections_connection.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CodeStarConnectionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeStarConnectionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrID, "codestar-connections", regexache.MustCompile("connection/.+")),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "codestar-connections", regexache.MustCompile("connection/.+")),
					resource.TestCheckResourceAttr(resourceName, "provider_type", string(types.ProviderTypeBitbucket)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "connection_status", string(types.ConnectionStatusPending)),
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

func TestAccCodeStarConnectionsConnection_hostARN(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.Connection
	resourceName := "aws_codestarconnections_connection.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CodeStarConnectionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeStarConnectionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_hostARN(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrID, "codestar-connections", regexache.MustCompile("connection/.+")),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "codestar-connections", regexache.MustCompile("connection/.+")),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "host_arn", "codestar-connections", regexache.MustCompile("host/.+")),
					resource.TestCheckResourceAttr(resourceName, "provider_type", string(types.ProviderTypeGithubEnterpriseServer)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "connection_status", string(types.ConnectionStatusPending)),
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

func TestAccCodeStarConnectionsConnection_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.Connection
	resourceName := "aws_codestarconnections_connection.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CodeStarConnectionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeStarConnectionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfcodestarconnections.ResourceConnection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCodeStarConnectionsConnection_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.Connection
	resourceName := "aws_codestarconnections_connection.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CodeStarConnectionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeStarConnectionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConnectionConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccConnectionConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckConnectionExists(ctx context.Context, t *testing.T, n string, v *types.Connection) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CodeStarConnectionsClient(ctx)

		output, err := tfcodestarconnections.FindConnectionByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckConnectionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CodeStarConnectionsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codestarconnections_connection" {
				continue
			}

			_, err := tfcodestarconnections.FindConnectionByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CodeStar Connections Connection %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccConnectionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "Bitbucket"
}
`, rName)
}

func testAccConnectionConfig_hostARN(rName string) string {
	return fmt.Sprintf(`
resource "aws_codestarconnections_host" "test" {
  name              = %[1]q
  provider_endpoint = "https://example.com"
  provider_type     = "GitHubEnterpriseServer"
}

resource "aws_codestarconnections_connection" "test" {
  name     = %[1]q
  host_arn = aws_codestarconnections_host.test.arn
}
`, rName)
}

func testAccConnectionConfig_tags1(rName string, tagKey1 string, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "Bitbucket"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccConnectionConfig_tags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "Bitbucket"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
