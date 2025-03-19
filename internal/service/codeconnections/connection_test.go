// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codeconnections_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/codeconnections/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodeconnections "github.com/hashicorp/terraform-provider-aws/internal/service/codeconnections"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCodeConnectionsConnection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.Connection
	resourceName := "aws_codeconnections_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeConnectionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, resourceName, &v),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeConnectionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_hostARN(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, resourceName, &v),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeConnectionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfcodeconnections.ResourceConnection, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCodeConnectionsConnection_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.Connection
	resourceName := "aws_codeconnections_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeConnectionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, resourceName, &v),
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
					testAccCheckConnectionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccConnectionConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckConnectionExists(ctx context.Context, n string, v *types.Connection) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeConnectionsClient(ctx)

		output, err := tfcodeconnections.FindConnectionByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckConnectionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeConnectionsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codeconnections_connection" {
				continue
			}

			_, err := tfcodeconnections.FindConnectionByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
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

func testAccConnectionConfig_tags1(rName string, tagKey1 string, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_codeconnections_connection" "test" {
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
resource "aws_codeconnections_connection" "test" {
  name          = %[1]q
  provider_type = "Bitbucket"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
