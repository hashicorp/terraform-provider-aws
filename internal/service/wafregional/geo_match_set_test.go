// Copyright IBM Corp. 2014, 2026
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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfwafregional "github.com/hashicorp/terraform-provider-aws/internal/service/wafregional"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFRegionalGeoMatchSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.GeoMatchSet
	resourceName := "aws_wafregional_geo_match_set.test"
	geoMatchSet := fmt.Sprintf("tfacc-%s", sdkacctest.RandString(5))

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGeoMatchSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGeoMatchSetConfig_basic(geoMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeoMatchSetExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, names.AttrName, geoMatchSet),
					resource.TestCheckResourceAttr(
						resourceName, "geo_match_constraint.#", "2"),
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

func TestAccWAFRegionalGeoMatchSet_changeNameForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.GeoMatchSet
	resourceName := "aws_wafregional_geo_match_set.test"
	geoMatchSet := fmt.Sprintf("tfacc-%s", sdkacctest.RandString(5))
	geoMatchSetNewName := fmt.Sprintf("geoMatchSetNewName-%s", sdkacctest.RandString(5))

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGeoMatchSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGeoMatchSetConfig_basic(geoMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeoMatchSetExists(ctx, t, resourceName, &before),
					resource.TestCheckResourceAttr(
						resourceName, names.AttrName, geoMatchSet),
					resource.TestCheckResourceAttr(
						resourceName, "geo_match_constraint.#", "2"),
				),
			},
			{
				Config: testAccGeoMatchSetConfig_changeName(geoMatchSetNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeoMatchSetExists(ctx, t, resourceName, &after),
					testAccCheckGeoMatchSetIdDiffers(&before, &after),
					resource.TestCheckResourceAttr(
						resourceName, names.AttrName, geoMatchSetNewName),
					resource.TestCheckResourceAttr(
						resourceName, "geo_match_constraint.#", "2"),
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

func TestAccWAFRegionalGeoMatchSet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.GeoMatchSet
	resourceName := "aws_wafregional_geo_match_set.test"
	geoMatchSet := fmt.Sprintf("tfacc-%s", sdkacctest.RandString(5))

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGeoMatchSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGeoMatchSetConfig_basic(geoMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeoMatchSetExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfwafregional.ResourceGeoMatchSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFRegionalGeoMatchSet_changeConstraints(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.GeoMatchSet
	resourceName := "aws_wafregional_geo_match_set.test"
	setName := fmt.Sprintf("tfacc-%s", sdkacctest.RandString(5))

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGeoMatchSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGeoMatchSetConfig_basic(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGeoMatchSetExists(ctx, t, resourceName, &before),
					resource.TestCheckResourceAttr(
						resourceName, names.AttrName, setName),
					resource.TestCheckResourceAttr(
						resourceName, "geo_match_constraint.#", "2"),
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
					resource.TestCheckResourceAttr(
						resourceName, names.AttrName, setName),
					resource.TestCheckResourceAttr(
						resourceName, "geo_match_constraint.#", "2"),
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

func TestAccWAFRegionalGeoMatchSet_noConstraints(t *testing.T) {
	ctx := acctest.Context(t)
	var ipset awstypes.GeoMatchSet
	resourceName := "aws_wafregional_geo_match_set.test"
	setName := fmt.Sprintf("tfacc-%s", sdkacctest.RandString(5))

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGeoMatchSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGeoMatchSetConfig_noConstraints(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGeoMatchSetExists(ctx, t, resourceName, &ipset),
					resource.TestCheckResourceAttr(
						resourceName, names.AttrName, setName),
					resource.TestCheckResourceAttr(
						resourceName, "geo_match_constraint.#", "0"),
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

func testAccCheckGeoMatchSetIdDiffers(before, after *awstypes.GeoMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.GeoMatchSetId == *after.GeoMatchSetId {
			return fmt.Errorf("Expected different IDs, given %q for both sets", *before.GeoMatchSetId)
		}
		return nil
	}
}

func testAccCheckGeoMatchSetExists(ctx context.Context, t *testing.T, n string, v *awstypes.GeoMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF Regional Geo Match Set ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).WAFRegionalClient(ctx)

		output, err := tfwafregional.FindGeoMatchSetByID(ctx, conn, rs.Primary.ID)

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
			if rs.Type != "aws_wafregional_geo_match_set" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).WAFRegionalClient(ctx)

			_, err := tfwafregional.FindGeoMatchSetByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WAF Regional Geo Match Set %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccGeoMatchSetConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_geo_match_set" "test" {
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
resource "aws_wafregional_geo_match_set" "test" {
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
resource "aws_wafregional_geo_match_set" "test" {
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
resource "aws_wafregional_geo_match_set" "test" {
  name = %[1]q
}
`, name)
}
