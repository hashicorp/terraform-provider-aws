// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53recoveryreadiness_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53recoveryreadiness"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53recoveryreadiness "github.com/hashicorp/terraform-provider-aws/internal/service/route53recoveryreadiness"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53RecoveryReadinessCell_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53recoveryreadiness_cell.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53RecoveryReadinessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCellDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCellConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCellExists(ctx, resourceName),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "route53-recovery-readiness", regexache.MustCompile(`cell/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cells.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "parent_readiness_scopes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccRoute53RecoveryReadinessCell_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53recoveryreadiness_cell.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53RecoveryReadinessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCellDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCellConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCellExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfroute53recoveryreadiness.ResourceCell(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53RecoveryReadinessCell_nestedCell(t *testing.T) {
	ctx := acctest.Context(t)
	rNameParent := sdkacctest.RandomWithPrefix("tf-acc-test-parent")
	rNameChild := sdkacctest.RandomWithPrefix("tf-acc-test-child")
	resourceNameParent := "aws_route53recoveryreadiness_cell.test_parent"
	resourceNameChild := "aws_route53recoveryreadiness_cell.test_child"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53RecoveryReadinessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCellDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCellConfig_child(rNameChild),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCellExists(ctx, resourceNameChild),
					acctest.MatchResourceAttrGlobalARN(resourceNameChild, names.AttrARN, "route53-recovery-readiness", regexache.MustCompile(`cell/.+`)),
				),
			},
			{
				Config: testAccCellConfig_parent(rNameChild, rNameParent),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCellExists(ctx, resourceNameParent),
					acctest.MatchResourceAttrGlobalARN(resourceNameParent, names.AttrARN, "route53-recovery-readiness", regexache.MustCompile(`cell/.+`)),
					resource.TestCheckResourceAttr(resourceNameParent, "cells.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceNameParent, "parent_readiness_scopes.#", acctest.Ct0),
					testAccCheckCellExists(ctx, resourceNameChild),
					acctest.MatchResourceAttrGlobalARN(resourceNameChild, names.AttrARN, "route53-recovery-readiness", regexache.MustCompile(`cell/.+`)),
					resource.TestCheckResourceAttr(resourceNameChild, "cells.#", acctest.Ct0),
				),
			},
			{
				Config: testAccCellConfig_parent(rNameChild, rNameParent),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceNameChild, "parent_readiness_scopes.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceNameParent,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      resourceNameChild,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRoute53RecoveryReadinessCell_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53recoveryreadiness_cell.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53RecoveryReadinessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCellDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCellConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCellExists(ctx, resourceName),
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
				Config: testAccCellConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCellExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccCellConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCellExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccRoute53RecoveryReadinessCell_timeout(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53recoveryreadiness_cell.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53RecoveryReadinessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCellDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCellConfig_timeout(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCellExists(ctx, resourceName),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "route53-recovery-readiness", regexache.MustCompile(`cell/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cells.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "parent_readiness_scopes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func testAccCheckCellDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53RecoveryReadinessConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53recoveryreadiness_cell" {
				continue
			}

			input := &route53recoveryreadiness.GetCellInput{
				CellName: aws.String(rs.Primary.ID),
			}

			_, err := conn.GetCellWithContext(ctx, input)
			if err == nil {
				return fmt.Errorf("Route53RecoveryReadiness Channel (%s) not deleted", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckCellExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53RecoveryReadinessConn(ctx)

		input := &route53recoveryreadiness.GetCellInput{
			CellName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetCellWithContext(ctx, input)

		return err
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53RecoveryReadinessConn(ctx)

	input := &route53recoveryreadiness.ListCellsInput{}

	_, err := conn.ListCellsWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCellConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_cell" "test" {
  cell_name = %q
}
`, rName)
}

func testAccCellConfig_child(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_cell" "test_child" {
  cell_name = %q
}
`, rName)
}

func testAccCellConfig_parent(rName, rName2 string) string {
	return acctest.ConfigCompose(testAccCellConfig_child(rName), fmt.Sprintf(`
resource "aws_route53recoveryreadiness_cell" "test_parent" {
  cell_name = %q
  cells     = [aws_route53recoveryreadiness_cell.test_child.arn]
}
`, rName2))
}

func testAccCellConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_cell" "test" {
  cell_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccCellConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_cell" "test" {
  cell_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccCellConfig_timeout(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_cell" "test" {
  cell_name = %q

  timeouts {
    delete = "10m"
  }
}
`, rName)
}
