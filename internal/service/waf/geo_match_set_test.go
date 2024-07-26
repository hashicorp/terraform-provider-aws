// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
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

func TestAccWAFGeoMatchSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.GeoMatchSet
	geoMatchSet := fmt.Sprintf("geoMatchSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_geo_match_set.geo_match_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGeoMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGeoMatchSetConfig_basic(geoMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeoMatchSetExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "waf", regexache.MustCompile(`geomatchset/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, geoMatchSet),
					resource.TestCheckResourceAttr(resourceName, "geo_match_constraint.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "geo_match_constraint.*", map[string]string{
						names.AttrType:  "Country",
						names.AttrValue: "US",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "geo_match_constraint.*", map[string]string{
						names.AttrType:  "Country",
						names.AttrValue: "CA",
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

func TestAccWAFGeoMatchSet_changeNameForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.GeoMatchSet
	geoMatchSet := fmt.Sprintf("geoMatchSet-%s", sdkacctest.RandString(5))
	geoMatchSetNewName := fmt.Sprintf("geoMatchSetNewName-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_geo_match_set.geo_match_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGeoMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGeoMatchSetConfig_basic(geoMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeoMatchSetExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, geoMatchSet),
					resource.TestCheckResourceAttr(resourceName, "geo_match_constraint.#", acctest.Ct2),
				),
			},
			{
				Config: testAccGeoMatchSetConfig_changeName(geoMatchSetNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeoMatchSetExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, geoMatchSetNewName),
					resource.TestCheckResourceAttr(resourceName, "geo_match_constraint.#", acctest.Ct2),
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

func TestAccWAFGeoMatchSet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.GeoMatchSet
	geoMatchSet := fmt.Sprintf("geoMatchSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_geo_match_set.geo_match_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGeoMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGeoMatchSetConfig_basic(geoMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeoMatchSetExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfwaf.ResourceGeoMatchSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFGeoMatchSet_changeConstraints(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.GeoMatchSet
	setName := fmt.Sprintf("geoMatchSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_geo_match_set.geo_match_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGeoMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGeoMatchSetConfig_basic(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGeoMatchSetExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, setName),
					resource.TestCheckResourceAttr(resourceName, "geo_match_constraint.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "geo_match_constraint.*", map[string]string{
						names.AttrType:  "Country",
						names.AttrValue: "US",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "geo_match_constraint.*", map[string]string{
						names.AttrType:  "Country",
						names.AttrValue: "CA",
					}),
				),
			},
			{
				Config: testAccGeoMatchSetConfig_changeConstraints(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGeoMatchSetExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, setName),
					resource.TestCheckResourceAttr(resourceName, "geo_match_constraint.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "geo_match_constraint.*", map[string]string{
						names.AttrType:  "Country",
						names.AttrValue: "RU",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "geo_match_constraint.*", map[string]string{
						names.AttrType:  "Country",
						names.AttrValue: "CN",
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

func TestAccWAFGeoMatchSet_noConstraints(t *testing.T) {
	ctx := acctest.Context(t)
	var ipset awstypes.GeoMatchSet
	setName := fmt.Sprintf("geoMatchSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_geo_match_set.geo_match_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGeoMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGeoMatchSetConfig_noConstraints(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGeoMatchSetExists(ctx, resourceName, &ipset),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, setName),
					resource.TestCheckResourceAttr(resourceName, "geo_match_constraint.#", acctest.Ct0),
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

func testAccCheckGeoMatchSetExists(ctx context.Context, n string, v *awstypes.GeoMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFClient(ctx)

		output, err := tfwaf.FindGeoMatchSetByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckGeoMatchSetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_waf_geo_match_set" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).WAFClient(ctx)

			_, err := tfwaf.FindGeoMatchSetByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WAF GeoMatchSet %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccGeoMatchSetConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_geo_match_set" "geo_match_set" {
  name = %[1]q

  geo_match_constraint {
    type  = "Country"
    value = "US"
  }

  geo_match_constraint {
    type  = "Country"
    value = "CA"
  }
}
`, name)
}

func testAccGeoMatchSetConfig_changeName(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_geo_match_set" "geo_match_set" {
  name = %[1]q

  geo_match_constraint {
    type  = "Country"
    value = "US"
  }

  geo_match_constraint {
    type  = "Country"
    value = "CA"
  }
}
`, name)
}

func testAccGeoMatchSetConfig_changeConstraints(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_geo_match_set" "geo_match_set" {
  name = %[1]q

  geo_match_constraint {
    type  = "Country"
    value = "RU"
  }

  geo_match_constraint {
    type  = "Country"
    value = "CN"
  }
}
`, name)
}

func testAccGeoMatchSetConfig_noConstraints(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_geo_match_set" "geo_match_set" {
  name = %[1]q
}
`, name)
}
