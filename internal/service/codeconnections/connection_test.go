// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package codeconnections_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/codeconnections/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcodeconnections "github.com/hashicorp/terraform-provider-aws/internal/service/codeconnections"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCodeConnectionsConnection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.Connection
	resourceName := "aws_codeconnections_connection.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeConnectionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrID, "codeconnections", regexache.MustCompile("connection/.+")),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "codeconnections", regexache.MustCompile("connection/.+")),
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

func TestAccCodeConnectionsConnection_hostARN(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.Connection
	resourceName := "aws_codeconnections_connection.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeConnectionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_hostARN(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrID, "codeconnections", regexache.MustCompile("connection/.+")),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "codeconnections", regexache.MustCompile("connection/.+")),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "host_arn", "codeconnections", regexache.MustCompile("host/.+")),
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

func TestAccCodeConnectionsConnection_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.Connection
	resourceName := "aws_codeconnections_connection.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeConnectionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcodeconnections.ResourceConnection, resourceName),
				),
				ExpectNonEmptyPlan: true,
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

		conn := acctest.ProviderMeta(ctx, t).CodeConnectionsClient(ctx)

		output, err := tfcodeconnections.FindConnectionByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckConnectionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CodeConnectionsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codeconnections_connection" {
				continue
			}

			_, err := tfcodeconnections.FindConnectionByARN(ctx, conn, rs.Primary.ID)

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
resource "aws_codeconnections_connection" "test" {
  name          = %[1]q
  provider_type = "Bitbucket"
}
`, rName)
}

func testAccConnectionConfig_hostARN(rName string) string {
	return fmt.Sprintf(`
resource "aws_codeconnections_host" "test" {
  name              = %[1]q
  provider_endpoint = "https://example.com"
  provider_type     = "GitHubEnterpriseServer"
}

resource "aws_codeconnections_connection" "test" {
  name     = %[1]q
  host_arn = aws_codeconnections_host.test.arn
}
`, rName)
}
