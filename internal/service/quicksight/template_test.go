// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQuickSightTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.Template
	resourceName := "aws_quicksight_template.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfig_basic(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "template_id", rId),
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

func TestAccQuickSightTemplate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.Template
	resourceName := "aws_quicksight_template.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfig_basic(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &template),
					acctest.CheckSDKResourceDisappears(ctx, t, tfquicksight.ResourceTemplate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccQuickSightTemplate_barChart(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.Template
	resourceName := "aws_quicksight_template.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfig_BarChart(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "template_id", rId),
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

func TestAccQuickSightTemplate_table(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.Template
	resourceName := "aws_quicksight_template.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfig_Table(rId, rName, "ASC", "START"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "template_id", rId),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.ResourceStatusCreationSuccessful)),
					resource.TestCheckResourceAttr(resourceName, "definition.0.sheets.0.visuals.0.table_visual.0.chart_configuration.0.sort_configuration.0.row_sort.0.field_sort.0.direction", "ASC"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.sheets.0.visuals.0.table_visual.0.chart_configuration.0.total_options.0.placement", "START"),
				),
			},
			{
				Config: testAccTemplateConfig_Table(rId, rName, "DESC", "END"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &v2),
					testAccCheckTemplateNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "template_id", rId),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.ResourceStatusCreationSuccessful)),
					resource.TestCheckResourceAttr(resourceName, "definition.0.sheets.0.visuals.0.table_visual.0.chart_configuration.0.sort_configuration.0.row_sort.0.field_sort.0.direction", "DESC"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.sheets.0.visuals.0.table_visual.0.chart_configuration.0.total_options.0.placement", "END"),
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

func TestAccQuickSightTemplate_sourceEntity(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.Template
	resourceName := "aws_quicksight_template.copy"
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
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfig_TemplateSourceEntity(rId, rName, sourceId, sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "template_id", rId),
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

func TestAccQuickSightTemplate_update(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.Template
	resourceName := "aws_quicksight_template.copy"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	sourceName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	sourceId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfig_TemplateSourceEntity(rId, rName, sourceId, sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "template_id", rId),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.ResourceStatusCreationSuccessful)),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, "source_entity.0.source_template.0.arn", "quicksight", fmt.Sprintf("template/%s", sourceId)),
					resource.TestCheckResourceAttr(resourceName, "version_number", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source_entity"},
			},

			{
				Config: testAccTemplateConfig_TemplateSourceEntity(rId, rNameUpdated, sourceId, sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "template_id", rId),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.ResourceStatusCreationSuccessful)),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, "source_entity.0.source_template.0.arn", "quicksight", fmt.Sprintf("template/%s", sourceId)),
					resource.TestCheckResourceAttr(resourceName, "version_number", "2"),
				),
			},
		},
	})
}

func testAccCheckTemplateDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_template" {
				continue
			}

			_, err := tfquicksight.FindTemplateByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes["template_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("QuickSight Template (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTemplateExists(ctx context.Context, t *testing.T, n string, v *awstypes.Template) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		output, err := tfquicksight.FindTemplateByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes["template_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTemplateNotRecreated(before, after *awstypes.Template) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if creationTimeBefore, creationTimeAfter := aws.ToTime(before.CreatedTime), aws.ToTime(after.CreatedTime); creationTimeBefore != creationTimeAfter {
			return errors.New("QuickSight Template not recreated")
		}

		return nil
	}
}

func testAccTemplateConfig_base(rId string, rName string) string {
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

func testAccTemplateConfig_basic(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccTemplateConfig_base(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_template" "test" {
  template_id         = %[1]q
  name                = %[2]q
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
`, rId, rName))
}

func testAccTemplateConfig_BarChart(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccTemplateConfig_base(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_template" "test" {
  template_id         = %[1]q
  name                = %[2]q
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
        bar_chart_visual {
          visual_id = "BarChart"
          chart_configuration {
            field_wells {
              bar_chart_aggregated_field_wells {
                category {
                  categorical_dimension_field {
                    field_id = "1"
                    column {
                      column_name         = "Column1"
                      data_set_identifier = "1"
                    }
                  }
                }
                values {
                  numerical_measure_field {
                    field_id = "2"
                    column {
                      column_name         = "Column2"
                      data_set_identifier = "1"
                    }
                    aggregation_function {
                      simple_numerical_aggregation = "SUM"
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
}
`, rId, rName))
}

func testAccTemplateConfig_Table(rId, rName, sortDirection, totalPlacement string) string {
	return acctest.ConfigCompose(
		testAccTemplateConfig_base(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_template" "test" {
  template_id         = %[1]q
  name                = %[2]q
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
        table_visual {
          visual_id = "Table"
          chart_configuration {
            field_wells {
              table_unaggregated_field_wells {
                values {
                  field_id = "1"
                  column {
                    column_name         = "Column1"
                    data_set_identifier = "1"
                  }
                }
                values {
                  field_id = "2"
                  column {
                    column_name         = "Column2"
                    data_set_identifier = "1"
                  }
                }
              }
            }
            table_options {
              header_style {
                background_color = "#99CCFF"
                height           = 20
                font_configuration {
                  font_color = "#212121"
                  font_size {
                    relative = "LARGE"
                  }
                }
              }
            }
            sort_configuration {
              row_sort {
                field_sort {
                  field_id  = "1"
                  direction = %[3]q
                }
              }
            }
            total_options {
              custom_label      = "Total"
              placement         = %[4]q
              totals_visibility = "VISIBLE"
            }
          }
        }
      }
    }
  }
}
`, rId, rName, sortDirection, totalPlacement))
}

func testAccTemplateConfig_TemplateSourceEntity(rId, rName, sourceId, sourceName string) string {
	return acctest.ConfigCompose(
		testAccTemplateConfig_BarChart(sourceId, sourceName),
		fmt.Sprintf(`
resource "aws_quicksight_template" "copy" {
  template_id         = %[1]q
  name                = %[2]q
  version_description = "test"
  source_entity {
    source_template {
      arn = aws_quicksight_template.test.arn
    }
  }
}
`, rId, rName))
}
