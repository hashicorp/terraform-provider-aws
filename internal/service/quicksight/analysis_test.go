// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQuickSightAnalysis_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var analysis awstypes.Analysis
	resourceName := "aws_quicksight_analysis.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnalysisDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAnalysisConfig_basic(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalysisExists(ctx, t, resourceName, &analysis),
					resource.TestCheckResourceAttr(resourceName, "analysis_id", rId),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.ResourceStatusCreationSuccessful)),
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

func TestAccQuickSightAnalysis_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var analysis awstypes.Analysis
	resourceName := "aws_quicksight_analysis.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnalysisDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAnalysisConfig_basic(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalysisExists(ctx, t, resourceName, &analysis),
					acctest.CheckSDKResourceDisappears(ctx, t, tfquicksight.ResourceAnalysis(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccQuickSightAnalysis_sourceEntity(t *testing.T) {
	ctx := acctest.Context(t)
	var analysis awstypes.Analysis
	resourceName := "aws_quicksight_analysis.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	sourceName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	sourceId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnalysisDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAnalysisConfig_TemplateSourceEntity(rId, rName, sourceId, sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalysisExists(ctx, t, resourceName, &analysis),
					resource.TestCheckResourceAttr(resourceName, "analysis_id", rId),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.ResourceStatusCreationSuccessful)),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, "source_entity.0.source_template.0.arn", "quicksight", fmt.Sprintf("template/%s", sourceId)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source_entity"},
			},
		},
	})
}

func TestAccQuickSightAnalysis_update(t *testing.T) {
	ctx := acctest.Context(t)
	var analysis awstypes.Analysis
	resourceName := "aws_quicksight_analysis.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnalysisDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAnalysisConfig_basic(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalysisExists(ctx, t, resourceName, &analysis),
					resource.TestCheckResourceAttr(resourceName, "analysis_id", rId),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.ResourceStatusCreationSuccessful)),
				),
			},
			{
				Config: testAccAnalysisConfig_basic(rId, rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalysisExists(ctx, t, resourceName, &analysis),
					resource.TestCheckResourceAttr(resourceName, "analysis_id", rId),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.ResourceStatusUpdateSuccessful)),
				),
			},
		},
	})
}

func TestAccQuickSightAnalysis_parametersConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var analysis awstypes.Analysis
	resourceName := "aws_quicksight_analysis.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnalysisDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAnalysisConfig_ParametersConfig(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalysisExists(ctx, t, resourceName, &analysis),
					resource.TestCheckResourceAttr(resourceName, "analysis_id", rId),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.ResourceStatusCreationSuccessful)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrParameters},
			},
		},
	})
}

func TestAccQuickSightAnalysis_forceDelete(t *testing.T) {
	ctx := acctest.Context(t)
	var analysis awstypes.Analysis
	resourceName := "aws_quicksight_analysis.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnalysisDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAnalysisConfig_ForceDelete(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalysisExists(ctx, t, resourceName, &analysis),
					resource.TestCheckResourceAttr(resourceName, "analysis_id", rId),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.ResourceStatusCreationSuccessful)),
				),
			},
		},
	})
}

func TestAccQuickSightAnalysis_Definition_calculatedFields(t *testing.T) {
	ctx := acctest.Context(t)
	var analysis awstypes.Analysis
	resourceName := "aws_quicksight_analysis.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnalysisDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAnalysisConfig_Definition_calculatedFields(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalysisExists(ctx, t, resourceName, &analysis),
					resource.TestCheckResourceAttr(resourceName, "analysis_id", rId),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.ResourceStatusCreationSuccessful)),
					resource.TestCheckResourceAttr(resourceName, "definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.calculated_fields.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "definition.0.calculated_fields.*", map[string]string{
						"data_set_identifier": "1",
						names.AttrExpression:  "1",
						names.AttrName:        "test1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "definition.0.calculated_fields.*", map[string]string{
						"data_set_identifier": "1",
						names.AttrExpression:  "2",
						names.AttrName:        "test2",
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

func TestAccQuickSightAnalysis_theme(t *testing.T) {
	ctx := acctest.Context(t)
	var analysis awstypes.Analysis
	resourceName := "aws_quicksight_analysis.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	themeArn := "arn:aws:quicksight::aws:theme/MIDNIGHT" //lintignore:AWSAT005

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnalysisDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAnalysisConfig_theme(rId, rName, themeArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalysisExists(ctx, t, resourceName, &analysis),
					resource.TestCheckResourceAttr(resourceName, "theme_arn", themeArn),
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

func TestAccQuickSightAnalysis_pieChartVisualArcThickness(t *testing.T) {
	ctx := acctest.Context(t)

	var analysis awstypes.Analysis
	resourceName := "aws_quicksight_analysis.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnalysisDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAnalysisConfig_pieChartVisualArcThickness(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalysisExists(ctx, t, resourceName, &analysis),
					resource.TestCheckResourceAttr(resourceName, "analysis_id", rId),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.ResourceStatusCreationSuccessful)),
					resource.TestCheckResourceAttr(resourceName, "definition.0.sheets.0.visuals.0.pie_chart_visual.0.chart_configuration.0.donut_options.0.arc_options.0.arc_thickness", string(awstypes.ArcThicknessWhole)),
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

func testAccCheckAnalysisDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_analysis" {
				continue
			}

			_, err := tfquicksight.FindAnalysisByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes["analysis_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("QuickSight Analysis (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAnalysisExists(ctx context.Context, t *testing.T, n string, v *awstypes.Analysis) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		output, err := tfquicksight.FindAnalysisByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes["analysis_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAnalysisConfig_base(rId string, rName string) string {
	return acctest.ConfigCompose(
		testAccDataSetConfig_base(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_set" "test" {
  data_set_id = %[1]q
  name        = %[2]q
  import_mode = "SPICE"

  physical_table_map {
    physical_table_map_id = %[1]q
    s3_source {
      data_source_arn = aws_quicksight_data_source.test.arn
      input_columns {
        name = "Column1"
        type = "STRING"
      }
      input_columns {
        name = "Column2"
        type = "STRING"
      }
      upload_settings {}
    }
  }
  logical_table_map {
    logical_table_map_id = %[1]q
    alias                = "Group1"
    source {
      physical_table_id = %[1]q
    }
    data_transforms {
      cast_column_type_operation {
        column_name     = "Column2"
        new_column_type = "INTEGER"
      }
    }
  }

  lifecycle {
    ignore_changes = [
      physical_table_map
    ]
  }
}
`, rId, rName))
}

func testAccAnalysisConfig_basic(rId, rName string) string {
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
}
`, rId, rName))
}

func testAccAnalysisConfig_TemplateSourceEntity(rId, rName, sourceId, sourceName string) string {
	return acctest.ConfigCompose(
		testAccAnalysisConfig_base(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_template" "test" {
  template_id         = %[3]q
  name                = %[4]q
  version_description = "test"
  definition {
    data_set_configuration {
      data_set_schema {
        column_schema_list {
          name      = "Column1"
          data_type = "STRING"
        }
        column_schema_list {
          name      = "Column2"
          data_type = "INTEGER"
        }
      }
      placeholder = "1"
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
}

resource "aws_quicksight_analysis" "test" {
  analysis_id = %[1]q
  name        = %[2]q
  source_entity {
    source_template {
      arn = aws_quicksight_template.test.arn
      data_set_references {
        data_set_arn         = aws_quicksight_data_set.test.arn
        data_set_placeholder = "1"
      }
    }
  }
}
`, rId, rName, sourceId, sourceName))
}

func testAccAnalysisConfig_ParametersConfig(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccAnalysisConfig_base(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_analysis" "test" {
  analysis_id = %[1]q
  name        = %[2]q
  parameters {
    string_parameters {
      name   = "test"
      values = ["value"]
    }
  }
  definition {
    data_set_identifiers_declarations {
      data_set_arn = aws_quicksight_data_set.test.arn
      identifier   = "1"
    }
    parameter_declarations {
      string_parameter_declaration {
        name                 = "test"
        parameter_value_type = "SINGLE_VALUED"
        default_values {
          static_values = ["value"]
        }
        values_when_unset {
          value_when_unset_option = "NULL"
        }
      }
    }
    sheets {
      title    = "Example"
      sheet_id = "Example1"
      visuals {
        line_chart_visual {
          visual_id = "LineChart"
          title {
            format_text {
              plain_text = "Line Chart Example"
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
}
`, rId, rName))
}

func testAccAnalysisConfig_ForceDelete(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccAnalysisConfig_base(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_analysis" "test" {
  analysis_id = %[1]q
  name        = %[2]q

  recovery_window_in_days = 0

  definition {
    data_set_identifiers_declarations {
      data_set_arn = aws_quicksight_data_set.test.arn
      identifier   = "1"
    }
    sheets {
      title    = "Example"
      sheet_id = "Example1"
      visuals {
        line_chart_visual {
          visual_id = "LineChart"
          title {
            format_text {
              plain_text = "Line Chart Example"
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
}
`, rId, rName))
}

func testAccAnalysisConfig_Definition_calculatedFields(rId, rName string) string {
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
    calculated_fields {
      data_set_identifier = "1"
      expression          = "1"
      name                = "test1"
    }
    calculated_fields {
      data_set_identifier = "1"
      expression          = "2"
      name                = "test2"
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
}
`, rId, rName))
}

func testAccAnalysisConfig_theme(rId, rName, themeArn string) string {
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
`, rId, rName, themeArn))
}

func testAccAnalysisConfig_pieChartVisualArcThickness(rId, rName string) string {
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
        pie_chart_visual {
          visual_id = "PieChart"
          title {
            format_text {
              plain_text = "Pie Chart Test"
            }
          }
          chart_configuration {
            field_wells {
              pie_chart_aggregated_field_wells {}
            }
            category_label_options {
              sort_icon_visibility = "HIDDEN"
              visibility           = "HIDDEN"
            }
            data_labels {
              category_label_visibility = "VISIBLE"
              label_color               = null
              label_content             = null
              measure_label_visibility  = "VISIBLE"
              overlap                   = "DISABLE_OVERLAP"
              position                  = null
              visibility                = "VISIBLE"
            }
            donut_options {
              arc_options {
                arc_thickness = "WHOLE"
              }
            }
          }
        }
      }
    }
  }
}
`, rId, rName))
}
