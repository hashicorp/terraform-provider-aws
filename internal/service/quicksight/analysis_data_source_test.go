// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQuickSightAnalysisDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_quicksight_analysis.test"
	dataSourceName := "data.aws_quicksight_analysis.test"
	themeArn := "arn:aws:quicksight::aws:theme/MIDNIGHT" //lintignore:AWSAT005

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAnalysisDataSourceConfig_basic(rId, rName, themeArn),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "theme_arn", themeArn),
					resource.TestCheckResourceAttr(dataSourceName, "definition.0.data_set_identifiers_declarations.0.identifier", "1"),
				),
			},
		},
	})
}

func testAccAnalysisDataSourceConfig_basic(rId, rName, themeArn string) string {
	return acctest.ConfigCompose(
		testAccAnalysisConfig_base(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_analysis" "test" {
  analysis_id = %[1]q
  name        = %[2]q

  definition {
    data_set_identifiers_declarations {
      data_set_arn = aws_quicksight_data_set.test.arn
      identifier   = "1"
    }

    sheets {
      title    = "Test"
      sheet_id = "Test1"

      visuals {
        custom_content_visual {
          data_set_identifier = "1"
          title {
            format_text {
              plain_text = "Test"
            }
          }
          visual_id = "Test1"
        }
      }

      visuals {
        line_chart_visual {
          visual_id = "LineChart"
          title {
            format_text {
              plain_text = "Line Chart Test"
            }
          }
          chart_configuration {
            field_wells {
              line_chart_aggregated_field_wells {
                category {
                  categorical_dimension_field {
                    field_id = "1"
                    column {
                      data_set_identifier = "1"
                      column_name         = "Column1"
                    }
                  }
                }
                values {
                  categorical_measure_field {
                    field_id = "2"
                    column {
                      data_set_identifier = "1"
                      column_name         = "Column1"
                    }
                    aggregation_function = "COUNT"
                  }
                }
              }
            }
          }
        }
      }
    }
  }

  theme_arn = %[3]q
}

data "aws_quicksight_analysis" "test" {
  analysis_id = aws_quicksight_analysis.test.analysis_id
}
`, rId, rName, themeArn))
}
