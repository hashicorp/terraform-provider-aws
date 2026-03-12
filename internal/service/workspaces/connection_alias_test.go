// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package workspaces_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workspaces/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfworkspaces "github.com/hashicorp/terraform-provider-aws/internal/service/workspaces"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWorkSpacesConnectionAlias_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var connectionalias awstypes.ConnectionAlias
	rName := acctest.RandomFQDomainName()
	resourceName := "aws_workspaces_connection_alias.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(workspaces.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionAliasConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionAliasExists(ctx, t, resourceName, &connectionalias),
					resource.TestCheckResourceAttr(resourceName, "connection_string", rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrState),
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

func TestAccWorkSpacesConnectionAlias_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var connectionalias awstypes.ConnectionAlias
	rName := acctest.RandomFQDomainName()
	resourceName := "aws_workspaces_connection_alias.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(workspaces.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionAliasConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionAliasExists(ctx, t, resourceName, &connectionalias),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfworkspaces.ResourceConnectionAlias, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWorkSpacesConnectionAlias_tags(t *testing.T) {
	ctx := acctest.Context(t)

	var connectionalias awstypes.ConnectionAlias
	rName := acctest.RandomFQDomainName()
	resourceName := "aws_workspaces_connection_alias.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(workspaces.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionAliasConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionAliasExists(ctx, t, resourceName, &connectionalias),
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
				Config: testAccConnectionAliasConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionAliasExists(ctx, t, resourceName, &connectionalias),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccConnectionAliasConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionAliasExists(ctx, t, resourceName, &connectionalias),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckConnectionAliasDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WorkSpacesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workspaces_connection_alias" {
				continue
			}

			_, err := tfworkspaces.FindConnectionAliasByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				return nil
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WorkSpaces Connection Alias %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckConnectionAliasExists(ctx context.Context, t *testing.T, n string, v *awstypes.ConnectionAlias) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).WorkSpacesClient(ctx)

		output, err := tfworkspaces.FindConnectionAliasByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).WorkSpacesClient(ctx)

	input := &workspaces.DescribeConnectionAliasesInput{}
	_, err := conn.DescribeConnectionAliases(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccConnectionAliasConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_workspaces_connection_alias" "test" {
  connection_string = %[1]q
}
`, rName)
}

func testAccConnectionAliasConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_workspaces_connection_alias" "test" {
  connection_string = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccConnectionAliasConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_workspaces_connection_alias" "test" {
  connection_string = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
