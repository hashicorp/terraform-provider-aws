// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workspaces/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfworkspaces "github.com/hashicorp/terraform-provider-aws/internal/service/workspaces"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWorkSpacesConnectionAlias_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var connectionalias awstypes.ConnectionAlias
	rName := acctest.RandomFQDomainName()
	resourceName := "aws_workspaces_connection_alias.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(workspaces.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionAliasConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionAliasExists(ctx, resourceName, &connectionalias),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(workspaces.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionAliasConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionAliasExists(ctx, resourceName, &connectionalias),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfworkspaces.ResourceConnectionAlias, resourceName),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(workspaces.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionAliasConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionAliasExists(ctx, resourceName, &connectionalias),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
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
					testAccCheckConnectionAliasExists(ctx, resourceName, &connectionalias),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccConnectionAliasConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionAliasExists(ctx, resourceName, &connectionalias),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckConnectionAliasDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkSpacesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workspaces_connection_alias" {
				continue
			}

			_, err := tfworkspaces.FindConnectionAliasByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				return nil
			}

			if err != nil {
				return err
			}

			return create.Error(names.WorkSpaces, create.ErrActionCheckingDestroyed, tfworkspaces.ResNameConnectionAlias, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckConnectionAliasExists(ctx context.Context, name string, connectionalias *awstypes.ConnectionAlias) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.WorkSpaces, create.ErrActionCheckingExistence, tfworkspaces.ResNameConnectionAlias, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.WorkSpaces, create.ErrActionCheckingExistence, tfworkspaces.ResNameConnectionAlias, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkSpacesClient(ctx)
		out, err := tfworkspaces.FindConnectionAliasByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.WorkSpaces, create.ErrActionCheckingExistence, tfworkspaces.ResNameConnectionAlias, rs.Primary.ID, err)
		}

		*connectionalias = *out

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).WorkSpacesClient(ctx)

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
