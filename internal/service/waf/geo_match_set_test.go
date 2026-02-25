// Copyright IBM Corp. 2014, 2026
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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfwaf "github.com/hashicorp/terraform-provider-aws/internal/service/waf"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFGeoMatchSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.GeoMatchSet
	geoMatchSet := fmt.Sprintf("geoMatchSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_geo_match_set.geo_match_set"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGeoMatchSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGeoMatchSetConfig_basic(geoMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeoMatchSetExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "waf", regexache.MustCompile(`geomatchset/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, geoMatchSet),
					resource.TestCheckResourceAttr(resourceName, "geo_match_constraint.#", "2"),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGeoMatchSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGeoMatchSetConfig_basic(geoMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeoMatchSetExists(ctx, t, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, geoMatchSet),
					resource.TestCheckResourceAttr(resourceName, "geo_match_constraint.#", "2"),
				),
			},
			{
				Config: testAccGeoMatchSetConfig_changeName(geoMatchSetNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeoMatchSetExists(ctx, t, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, geoMatchSetNewName),
					resource.TestCheckResourceAttr(resourceName, "geo_match_constraint.#", "2"),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGeoMatchSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGeoMatchSetConfig_basic(geoMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeoMatchSetExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfwaf.ResourceGeoMatchSet(), resourceName),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGeoMatchSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGeoMatchSetConfig_basic(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGeoMatchSetExists(ctx, t, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, setName),
					resource.TestCheckResourceAttr(resourceName, "geo_match_constraint.#", "2"),
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
					testAccCheckGeoMatchSetExists(ctx, t, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, setName),
					resource.TestCheckResourceAttr(resourceName, "geo_match_constraint.#", "2"),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGeoMatchSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGeoMatchSetConfig_noConstraints(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGeoMatchSetExists(ctx, t, resourceName, &ipset),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, setName),
					resource.TestCheckResourceAttr(resourceName, "geo_match_constraint.#", "0"),
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

func testAccCheckGeoMatchSetExists(ctx context.Context, t *testing.T, n string, v *awstypes.GeoMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).WAFClient(ctx)

		output, err := tfwaf.FindGeoMatchSetByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckGeoMatchSetDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_waf_geo_match_set" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).WAFClient(ctx)

			_, err := tfwaf.FindGeoMatchSetByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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
