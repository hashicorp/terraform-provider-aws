// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/wafregional/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwafregional "github.com/hashicorp/terraform-provider-aws/internal/service/wafregional"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFRegionalSQLInjectionMatchSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SqlInjectionMatchSet
	resourceName := "aws_wafregional_sql_injection_match_set.sql_injection_match_set"
	sqlInjectionMatchSet := fmt.Sprintf("sqlInjectionMatchSet-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSQLInjectionMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSQLInjectionMatchSetConfig_basic(sqlInjectionMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, names.AttrName, sqlInjectionMatchSet),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sql_injection_match_tuple.*", map[string]string{
						"field_to_match.#":      acctest.Ct1,
						"field_to_match.0.data": "",
						"field_to_match.0.type": "QUERY_STRING",
						"text_transformation":   "URL_DECODE",
					}),
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

func TestAccWAFRegionalSQLInjectionMatchSet_changeNameForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.SqlInjectionMatchSet
	resourceName := "aws_wafregional_sql_injection_match_set.sql_injection_match_set"
	sqlInjectionMatchSet := fmt.Sprintf("sqlInjectionMatchSet-%s", sdkacctest.RandString(5))
	sqlInjectionMatchSetNewName := fmt.Sprintf("sqlInjectionMatchSetNewName-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSQLInjectionMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSQLInjectionMatchSetConfig_basic(sqlInjectionMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(
						resourceName, names.AttrName, sqlInjectionMatchSet),
					resource.TestCheckResourceAttr(
						resourceName, "sql_injection_match_tuple.#", acctest.Ct1),
				),
			},
			{
				Config: testAccSQLInjectionMatchSetConfig_basic(sqlInjectionMatchSetNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(
						resourceName, names.AttrName, sqlInjectionMatchSetNewName),
					resource.TestCheckResourceAttr(
						resourceName, "sql_injection_match_tuple.#", acctest.Ct1),
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

func TestAccWAFRegionalSQLInjectionMatchSet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SqlInjectionMatchSet
	resourceName := "aws_wafregional_sql_injection_match_set.sql_injection_match_set"
	sqlInjectionMatchSet := fmt.Sprintf("sqlInjectionMatchSet-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSQLInjectionMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSQLInjectionMatchSetConfig_basic(sqlInjectionMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfwafregional.ResourceSQLInjectionMatchSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFRegionalSQLInjectionMatchSet_changeTuples(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.SqlInjectionMatchSet
	resourceName := "aws_wafregional_sql_injection_match_set.sql_injection_match_set"
	setName := fmt.Sprintf("sqlInjectionMatchSet-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSQLInjectionMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSQLInjectionMatchSetConfig_basic(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(
						resourceName, names.AttrName, setName),
					resource.TestCheckResourceAttr(
						resourceName, "sql_injection_match_tuple.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sql_injection_match_tuple.*", map[string]string{
						"field_to_match.#":      acctest.Ct1,
						"field_to_match.0.data": "",
						"field_to_match.0.type": "QUERY_STRING",
						"text_transformation":   "URL_DECODE",
					}),
				),
			},
			{
				Config: testAccSQLInjectionMatchSetConfig_changeTuples(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(
						resourceName, names.AttrName, setName),
					resource.TestCheckResourceAttr(
						resourceName, "sql_injection_match_tuple.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sql_injection_match_tuple.*", map[string]string{
						"field_to_match.#":      acctest.Ct1,
						"field_to_match.0.data": "user-agent",
						"field_to_match.0.type": "HEADER",
						"text_transformation":   "NONE",
					}),
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

func TestAccWAFRegionalSQLInjectionMatchSet_noTuples(t *testing.T) {
	ctx := acctest.Context(t)
	var ipset awstypes.SqlInjectionMatchSet
	resourceName := "aws_wafregional_sql_injection_match_set.sql_injection_match_set"
	setName := fmt.Sprintf("sqlInjectionMatchSet-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSQLInjectionMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSQLInjectionMatchSetConfig_noTuples(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(ctx, resourceName, &ipset),
					resource.TestCheckResourceAttr(
						resourceName, names.AttrName, setName),
					resource.TestCheckResourceAttr(
						resourceName, "sql_injection_match_tuple.#", acctest.Ct0),
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

func testAccCheckSQLInjectionMatchSetExists(ctx context.Context, n string, v *awstypes.SqlInjectionMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFRegionalClient(ctx)

		output, err := tfwafregional.FindSQLInjectionMatchSetByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckSQLInjectionMatchSetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_wafregional_sql_injection_match_set" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).WAFRegionalClient(ctx)

			_, err := tfwafregional.FindSQLInjectionMatchSetByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WAF Regional SQL Injection Match Set %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccSQLInjectionMatchSetConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_sql_injection_match_set" "sql_injection_match_set" {
  name = %[1]q

  sql_injection_match_tuple {
    text_transformation = "URL_DECODE"

    field_to_match {
      type = "QUERY_STRING"
    }
  }
}
`, name)
}

func testAccSQLInjectionMatchSetConfig_changeTuples(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_sql_injection_match_set" "sql_injection_match_set" {
  name = %[1]q

  sql_injection_match_tuple {
    text_transformation = "NONE"

    field_to_match {
      type = "HEADER"
      data = "User-Agent"
    }
  }
}
`, name)
}

func testAccSQLInjectionMatchSetConfig_noTuples(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_sql_injection_match_set" "sql_injection_match_set" {
  name = %[1]q
}
`, name)
}
