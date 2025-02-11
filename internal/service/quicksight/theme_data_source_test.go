// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQuickSightThemeDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_quicksight_theme.test"
	dataSourceName := "data.aws_quicksight_theme.test"
	themeId := "MIDNIGHT"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThemeDataSourceConfig_basic(rId, rName, themeId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "configuration.0.data_color_palette.0.colors.0", "#FFFFFF"),
					resource.TestCheckResourceAttr(dataSourceName, "configuration.0.data_color_palette.0.empty_fill_color", "#FFFFFF"),
					resource.TestCheckResourceAttr(dataSourceName, "configuration.0.data_color_palette.0.min_max_gradient.0", "#FFFFFF"),
					resource.TestCheckNoResourceAttr(dataSourceName, "configuration.0.sheet.0"),
					resource.TestCheckNoResourceAttr(dataSourceName, "configuration.0.typography.0"),
					resource.TestCheckNoResourceAttr(dataSourceName, "configuration.0.ui_color_palette.0"),
				),
			},
		},
	})
}

func TestAccQuickSightThemeDataSource_fullConfig(t *testing.T) {
	ctx := acctest.Context(t)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_quicksight_theme.test"
	dataSourceName := "data.aws_quicksight_theme.test"
	themeId := "MIDNIGHT"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccThemeDataSourceConfig_fullConfig(rId, rName, themeId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "configuration.0.data_color_palette.0.colors.0", "#FFFFFF"),
					resource.TestCheckResourceAttr(dataSourceName, "configuration.0.data_color_palette.0.empty_fill_color", "#FFFFFF"),
					resource.TestCheckResourceAttr(dataSourceName, "configuration.0.data_color_palette.0.min_max_gradient.0", "#FFFFFF"),
					resource.TestCheckResourceAttr(dataSourceName, "configuration.0.sheet.0.tile.0.border.0.show", acctest.CtFalse),
					resource.TestCheckResourceAttr(dataSourceName, "configuration.0.sheet.0.tile_layout.0.gutter.0.show", acctest.CtFalse),
					resource.TestCheckResourceAttr(dataSourceName, "configuration.0.sheet.0.tile_layout.0.margin.0.show", acctest.CtFalse),
					resource.TestCheckResourceAttr(dataSourceName, "configuration.0.typography.0.font_families.0.font_family", "monospace"),
					resource.TestCheckResourceAttr(dataSourceName, "configuration.0.typography.0.font_families.1.font_family", "Roboto"),
					resource.TestCheckResourceAttr(dataSourceName, "configuration.0.ui_color_palette.0.accent", "#202020"),
					resource.TestCheckResourceAttr(dataSourceName, "configuration.0.ui_color_palette.0.accent_foreground", "#FFFFFF"),
				),
			},
		},
	})
}

func testAccThemeDataSourceConfig_basic(rId, rName, baseThemId string) string {
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

data "aws_quicksight_theme" "test" {
  theme_id = aws_quicksight_theme.test.theme_id
}
`, rId, rName, baseThemId))
}

func testAccThemeDataSourceConfig_fullConfig(rId, rName, baseThemId string) string {
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

data "aws_quicksight_theme" "test" {
  theme_id = aws_quicksight_theme.test.theme_id
}
`, rId, rName, baseThemId))
}
