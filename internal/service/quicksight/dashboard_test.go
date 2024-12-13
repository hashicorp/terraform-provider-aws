// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQuickSightDashboard_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var dashboard awstypes.Dashboard
	resourceName := "aws_quicksight_dashboard.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDashboardDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDashboardConfig_basic(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDashboardExists(ctx, resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "dashboard_id", rId),
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

func TestAccQuickSightDashboard_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var dashboard awstypes.Dashboard
	resourceName := "aws_quicksight_dashboard.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDashboardDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDashboardConfig_basic(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDashboardExists(ctx, resourceName, &dashboard),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfquicksight.ResourceDashboard(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccQuickSightDashboard_sourceEntity(t *testing.T) {
	ctx := acctest.Context(t)
	var dashboard awstypes.Dashboard
	resourceName := "aws_quicksight_dashboard.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDashboardDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDashboardConfig_TemplateSourceEntity(rId, rName, sourceId, sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDashboardExists(ctx, resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "dashboard_id", rId),
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

func TestAccQuickSightDashboard_updateVersionNumber(t *testing.T) {
	ctx := acctest.Context(t)
	var dashboard awstypes.Dashboard
	var dashboardV1 awstypes.Dashboard
	resourceName := "aws_quicksight_dashboard.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDashboardDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDashboardConfig_basic(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDashboardExists(ctx, resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "dashboard_id", rId),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.ResourceStatusCreationSuccessful)),
					resource.TestCheckResourceAttr(resourceName, "version_number", "1"),
				),
			},
			{
				Config: testAccDashboardConfig_basic(rId, rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDashboardExists(ctx, resourceName, &dashboardV1),
					resource.TestCheckResourceAttr(resourceName, "dashboard_id", rId),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.ResourceStatusCreationSuccessful)),
					resource.TestCheckResourceAttr(resourceName, "version_number", "2"),
					testAccCheckDashboardVersionExists(ctx, resourceName, 1, &dashboardV1),
				),
			},
		},
	})
}

func TestAccQuickSightDashboard_dashboardSpecificConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var dashboard awstypes.Dashboard
	resourceName := "aws_quicksight_dashboard.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDashboardDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDashboardConfig_DashboardSpecificConfig(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDashboardExists(ctx, resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "dashboard_id", rId),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "dashboard_publish_options.0.ad_hoc_filtering_option.0.availability_status", string(awstypes.StatusDisabled)),
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

func testAccCheckDashboardDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_dashboard" {
				continue
			}

			_, err := tfquicksight.FindDashboardByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes["dashboard_id"], tfquicksight.DashboardLatestVersion)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("QuickSight Dashboard (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDashboardExists(ctx context.Context, n string, v *awstypes.Dashboard) resource.TestCheckFunc {
	return testAccCheckDashboardVersionExists(ctx, n, tfquicksight.DashboardLatestVersion, v)
}

func testAccCheckDashboardVersionExists(ctx context.Context, n string, version int64, v *awstypes.Dashboard) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightClient(ctx)

		output, err := tfquicksight.FindDashboardByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes["dashboard_id"], version)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccDashboardConfig_base(rId string, rName string) string {
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

func testAccDashboardConfig_basic(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccDashboardConfig_base(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_dashboard" "test" {
  dashboard_id        = %[1]q
  name                = %[2]q
  version_description = "test"
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

func testAccDashboardConfig_TemplateSourceEntity(rId, rName, sourceId, sourceName string) string {
	return acctest.ConfigCompose(
		testAccDashboardConfig_base(rId, rName),
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

resource "aws_quicksight_dashboard" "test" {
  dashboard_id        = %[1]q
  name                = %[2]q
  version_description = "test"
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

func testAccDashboardConfig_DashboardSpecificConfig(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccDashboardConfig_base(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_dashboard" "test" {
  dashboard_id        = %[1]q
  name                = %[2]q
  version_description = "test"
  parameters {
    string_parameters {
      name   = "test"
      values = ["value"]
    }
  }
  dashboard_publish_options {
    ad_hoc_filtering_option {
      availability_status = "DISABLED"
    }
    data_point_drill_up_down_option {
      availability_status = "ENABLED"
    }
    data_point_menu_label_option {
      availability_status = "ENABLED"
    }
    data_point_tooltip_option {
      availability_status = "ENABLED"
    }
    export_to_csv_option {
      availability_status = "ENABLED"
    }
    export_with_hidden_fields_option {
      availability_status = "DISABLED"
    }
    sheet_controls_option {
      visibility_state = "COLLAPSED"
    }
    sheet_layout_element_maximization_option {
      availability_status = "ENABLED"
    }
    visual_axis_sort_option {
      availability_status = "ENABLED"
    }
    visual_menu_option {
      availability_status = "ENABLED"
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
