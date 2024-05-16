// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/waf/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwaf "github.com/hashicorp/terraform-provider-aws/internal/service/waf"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFSQLInjectionMatchSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SqlInjectionMatchSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_sql_injection_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSQLInjectionMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSQLInjectionMatchSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "sql_injection_match_tuples.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sql_injection_match_tuples.*", map[string]string{
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

func TestAccWAFSQLInjectionMatchSet_changeNameForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.SqlInjectionMatchSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameNew := sdkacctest.RandomWithPrefix("tf-acc-test-new")
	resourceName := "aws_waf_sql_injection_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSQLInjectionMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSQLInjectionMatchSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "sql_injection_match_tuples.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSQLInjectionMatchSetConfig_changeName(rNameNew),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameNew),
					resource.TestCheckResourceAttr(resourceName, "sql_injection_match_tuples.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccWAFSQLInjectionMatchSet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SqlInjectionMatchSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_sql_injection_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSQLInjectionMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSQLInjectionMatchSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfwaf.ResourceSQLInjectionMatchSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFSQLInjectionMatchSet_changeTuples(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.SqlInjectionMatchSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_sql_injection_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSQLInjectionMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSQLInjectionMatchSetConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "sql_injection_match_tuples.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sql_injection_match_tuples.*", map[string]string{
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
			{
				Config: testAccSQLInjectionMatchSetConfig_changeTuples(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "sql_injection_match_tuples.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccWAFSQLInjectionMatchSet_noTuples(t *testing.T) {
	ctx := acctest.Context(t)
	var sqlSet awstypes.SqlInjectionMatchSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_sql_injection_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSQLInjectionMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSQLInjectionMatchSetConfig_noTuples(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(ctx, resourceName, &sqlSet),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "sql_injection_match_tuples.#", acctest.Ct0),
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFClient(ctx)

		output, err := tfwaf.FindSQLInjectionMatchSetByID(ctx, conn, rs.Primary.ID)

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
			if rs.Type != "aws_waf_sql_injection_match_set" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).WAFClient(ctx)

			_, err := tfwaf.FindSQLInjectionMatchSetByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WAF SqlInjectionMatchSet %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccSQLInjectionMatchSetConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_sql_injection_match_set" "test" {
  name = %[1]q

  sql_injection_match_tuples {
    text_transformation = "URL_DECODE"

    field_to_match {
      type = "QUERY_STRING"
    }
  }
}
`, name)
}

func testAccSQLInjectionMatchSetConfig_changeName(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_sql_injection_match_set" "test" {
  name = %[1]q

  sql_injection_match_tuples {
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
resource "aws_waf_sql_injection_match_set" "test" {
  name = %[1]q

  sql_injection_match_tuples {
    text_transformation = "NONE"

    field_to_match {
      type = "METHOD"
    }
  }
}
`, name)
}

func testAccSQLInjectionMatchSetConfig_noTuples(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_sql_injection_match_set" "test" {
  name = %[1]q
}
`, name)
}
