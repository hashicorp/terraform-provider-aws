// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQuickSightTheme_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var theme quicksight.Theme
	resourceName := "aws_quicksight_theme.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	themeId := "MIDNIGHT"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckThemeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccThemeConfig_basic(rId, rName, themeId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThemeExists(ctx, resourceName, &theme),
					resource.TestCheckResourceAttr(resourceName, "theme_id", rId),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, quicksight.ResourceStatusCreationSuccessful),
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

func TestAccQuickSightTheme_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var theme quicksight.Theme
	resourceName := "aws_quicksight_theme.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	themeId := "MIDNIGHT"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckThemeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccThemeConfig_basic(rId, rName, themeId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThemeExists(ctx, resourceName, &theme),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfquicksight.ResourceTheme(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
func TestAccQuickSightTheme_fullConfig(t *testing.T) {
	ctx := acctest.Context(t)

	var theme quicksight.Theme
	resourceName := "aws_quicksight_theme.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	themeId := "MIDNIGHT"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckThemeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccThemeConfig_fullConfig(rId, rName, themeId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThemeExists(ctx, resourceName, &theme),
					resource.TestCheckResourceAttr(resourceName, "theme_id", rId),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, quicksight.ResourceStatusCreationSuccessful),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.ui_color_palette.0.measure_foreground", "#FFFFFF"),
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

func TestAccQuickSightTheme_update(t *testing.T) {
	ctx := acctest.Context(t)

	var theme quicksight.Theme
	resourceName := "aws_quicksight_theme.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	themeId := "MIDNIGHT"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckThemeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccThemeConfig_update(rId, rName, themeId, "#FFFFFF"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThemeExists(ctx, resourceName, &theme),
					resource.TestCheckResourceAttr(resourceName, "theme_id", rId),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, quicksight.ResourceStatusCreationSuccessful),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.data_color_palette.0.empty_fill_color", "#FFFFFF"),
					resource.TestCheckResourceAttr(resourceName, "version_number", acctest.Ct1),
				),
			},
			{
				Config: testAccThemeConfig_update(rId, rNameUpdated, themeId, "#000000"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThemeExists(ctx, resourceName, &theme),
					resource.TestCheckResourceAttr(resourceName, "theme_id", rId),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, quicksight.ResourceStatusCreationSuccessful),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.data_color_palette.0.empty_fill_color", "#000000"),
					resource.TestCheckResourceAttr(resourceName, "version_number", acctest.Ct2),
				),
			},
		},
	})
}

func testAccCheckThemeDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_theme" {
				continue
			}

			output, err := tfquicksight.FindThemeByID(ctx, conn, rs.Primary.ID)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
					return nil
				}
				return err
			}

			if output != nil {
				return fmt.Errorf("QuickSight Theme (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckThemeExists(ctx context.Context, name string, theme *quicksight.Theme) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.QuickSight, create.ErrActionCheckingExistence, tfquicksight.ResNameTheme, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.QuickSight, create.ErrActionCheckingExistence, tfquicksight.ResNameTheme, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn(ctx)
		output, err := tfquicksight.FindThemeByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.QuickSight, create.ErrActionCheckingExistence, tfquicksight.ResNameTheme, rs.Primary.ID, err)
		}

		*theme = *output

		return nil
	}
}

func testAccThemeConfig_basic(rId, rName, baseThemId string) string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
resource "aws_quicksight_theme" "test" {
  theme_id = %[1]q
  name     = %[2]q

  base_theme_id = %[3]q

  configuration {
    data_color_palette {
      colors = [
        "#FFFFFF",
        "#111111",
        "#222222",
        "#333333",
        "#444444",
        "#555555",
        "#666666",
        "#777777",
        "#888888",
        "#999999"
      ]
      empty_fill_color = "#FFFFFF"
      min_max_gradient = [
        "#FFFFFF",
        "#111111",
      ]
    }
  }
}
`, rId, rName, baseThemId))
}

func testAccThemeConfig_fullConfig(rId, rName, baseThemId string) string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
resource "aws_quicksight_theme" "test" {
  theme_id = %[1]q
  name     = %[2]q

  base_theme_id = %[3]q

  configuration {
    data_color_palette {
      colors = [
        "#FFFFFF",
        "#111111",
        "#222222",
        "#333333",
        "#444444",
        "#555555",
        "#666666",
        "#777777",
        "#888888",
        "#999999"
      ]
      empty_fill_color = "#FFFFFF"
      min_max_gradient = [
        "#FFFFFF",
        "#111111",
      ]
    }
    sheet {
      tile {
        border {
          show = false
        }
      }
      tile_layout {
        gutter {
          show = false
        }
        margin {
          show = false
        }
      }
    }
    typography {
      font_families {
        font_family = "monospace"
      }
      font_families {
        font_family = "Roboto"
      }
    }
    ui_color_palette {
      accent               = "#202020"
      accent_foreground    = "#FFFFFF"
      danger               = "#202020"
      danger_foreground    = "#FFFFFF"
      dimension            = "#202020"
      dimension_foreground = "#FFFFFF"
      measure              = "#202020"
      measure_foreground   = "#FFFFFF"
      primary_background   = "#202020"
      primary_foreground   = "#FFFFFF"
      secondary_background = "#202020"
      secondary_foreground = "#FFFFFF"
      success              = "#202020"
      success_foreground   = "#FFFFFF"
      warning              = "#202020"
      warning_foreground   = "#FFFFFF"
    }
  }
}
`, rId, rName, baseThemId))
}

func testAccThemeConfig_update(rId, rName, baseThemId, emptyFillColor string) string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
resource "aws_quicksight_theme" "test" {
  theme_id = %[1]q
  name     = %[2]q

  base_theme_id = %[3]q

  configuration {
    data_color_palette {
      colors = [
        "#FFFFFF",
        "#111111",
        "#222222",
        "#333333",
        "#444444",
        "#555555",
        "#666666",
        "#777777",
        "#888888",
        "#999999"
      ]
      empty_fill_color = %[4]q
      min_max_gradient = [
        "#FFFFFF",
        "#111111",
      ]
    }
  }
}
`, rId, rName, baseThemId, emptyFillColor))
}
