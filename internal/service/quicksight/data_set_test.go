package quicksight_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
)

func TestAccQuickSightDataSet_basic(t *testing.T) {
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, quicksight.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQuickSightDataSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfig(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightDataSetExists(resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, "data_set_id", rId),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "quicksight", fmt.Sprintf("dataset/%s", rId)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "import_mode", "SPICE"),
					resource.TestCheckResourceAttr(resourceName, "physical_table_map.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "physical_table_map.0.s3_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "physical_table_map.0.s3_source.0.input_columns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "physical_table_map.0.s3_source.0.input_columns.0.name", "ColumnId-1"),
					resource.TestCheckResourceAttr(resourceName, "physical_table_map.0.s3_source.0.input_columns.0.type", "STRING"),
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

func TestAccQuickSightDataSet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, quicksight.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQuickSightDataSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfig(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightDataSetExists(resourceName, &dataSet),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfquicksight.ResourceDataSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// REQUIRES LOGICAL_TABLE_MAP
func TestAccQuickSightDataSet_column_groups(t *testing.T) {
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, quicksight.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQuickSightDataSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfigColumnGroups(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightDataSetExists(resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, "column_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "column_groups.0.geo_spatial_column_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "column_groups.0.geo_spatial_column_group.0.columns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "column_groups.0.geo_spatial_column_group.0.columns.0", "column1"),
					resource.TestCheckResourceAttr(resourceName, "column_groups.0.geo_spatial_column_group.0.country_code", "US"),
					resource.TestCheckResourceAttr(resourceName, "column_groups.0.geo_spatial_column_group.0.name", "test"),
				),
			},
		},
	})
}

func TestAccQuickSightDataSet_column_level_permission_rules(t *testing.T) {
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, quicksight.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQuickSightDataSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfigColumnLevelPermissionRules(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightDataSetExists(resourceName, &dataSet),
				),
			},
		},
	})
}

func TestAccQuickSightDataSet_data_set_usage_configuration(t *testing.T) {
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, quicksight.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQuickSightDataSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfigDataSetUsageConfiguration(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightDataSetExists(resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, "data_set_usage_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_set_usage_configuration.0.disable_use_as_direct_query_source", "false"),
					resource.TestCheckResourceAttr(resourceName, "data_set_usage_configuration.0.disable_use_as_imported_source", "false"),
				),
			},
		},
	})
}

func TestAccQuickSightDataSet_field_folders(t *testing.T) {
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, quicksight.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQuickSightDataSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfigFieldFolders(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightDataSetExists(resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, "field_folders.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "field_folders.0.columns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "field_folders.0.columns.0", "ColumnId-1"),
					resource.TestCheckResourceAttr(resourceName, "field_folders.0.description", "test"),
				),
			},
		},
	})
}

// ERROR: submitted a ticket to aws, they said you cannot create a ltm at this moment, instead just update an old one
func TestAccQuickSightDataSet_logical_table_map(t *testing.T) {
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, quicksight.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQuickSightDataSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfigLogicalTableMap(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightDataSetExists(resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, "logical_table_map.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logical_table_map.alias", "Group 1"),
					resource.TestCheckResourceAttr(resourceName, "logical_table_map.source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logical_table_map.source.physical_table_id", "s3PhysicalTable"),
				),
			},
		},
	})
}

// need help with this, don't know how to test
func TestAccQuickSightDataSet_permissions(t *testing.T) {
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, quicksight.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQuickSightDataSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfigPermissions(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightDataSetExists(resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, "permission.#", "1"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "permission.*", map[string]*regexp.Regexp{
						"principal": regexp.MustCompile(fmt.Sprintf(`user/default/%s`, rName)),
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "permission.*.actions.*", "quicksight:DescribeDataSet"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permission.*.actions.*", "quicksight:DescribeDataSetPermissions"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permission.*.actions.*", "quicksight:PassDataSet"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSetConfigUpdatePermissions(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightDataSetExists(resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, "permission.#", "1"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "permission.*", map[string]*regexp.Regexp{
						"principal": regexp.MustCompile(fmt.Sprintf(`user/default/%s`, rName)),
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "permission.*.actions.*", "quicksight:DescribeDataSet"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permission.*.actions.*", "quicksight:DescribeDataSetPermissions"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permission.*.actions.*", "quicksight:PassDataSet"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permission.*.actions.*", "quicksight:UpdateDataSet"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permission.*.actions.*", "quicksight:DeleteDataSet"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permission.*.actions.*", "quicksight:UpdateDataSetPermissions"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSetConfig(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightDataSetExists(resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, "permission.#", "0"),
				),
			},
		},
	})
}

// data set is created, but attribute gets deleted on creation
func TestAccQuickSightDataSet_row_level_permission_data_set(t *testing.T) {
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, quicksight.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQuickSightDataSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfigRowLevelPermissionDataSet(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightDataSetExists(resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, "row_level_permission_data_set.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "row_level_permission_data_set.arn", "this.arn"),
					resource.TestCheckResourceAttr(resourceName, "row_level_permission_data_set.permission_policy", "GRANT_ACCESS"),
					resource.TestCheckResourceAttr(resourceName, "row_level_permission_data_set.format_version", "VERSION_1"),
					resource.TestCheckResourceAttr(resourceName, "row_level_permission_data_set.namespace", "namespace"),
					resource.TestCheckResourceAttr(resourceName, "row_level_permission_data_set.status", "ENABLED"),
				),
			},
		},
	})
}

// data set is created, but attribute gets deleted on creation
func TestAccQuickSightDataSet_row_level_permission_tag_configuration(t *testing.T) {
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, quicksight.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQuickSightDataSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfigRowLevelPermissionTagConfiguration(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightDataSetExists(resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, "row_level_permission_tag_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "row_level_permission_tag_configuration.0.tag_rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "row_level_permission_tag_configuration.0.tag_rules.0.column_name", "columnname"),
					resource.TestCheckResourceAttr(resourceName, "row_level_permission_tag_configuration.0.tag_rules.0.tag_key", "uniquetagkey"),
					resource.TestCheckResourceAttr(resourceName, "row_level_permission_tag_configuration.0.tag_rules.0.match_all_value", "*"),
					resource.TestCheckResourceAttr(resourceName, "row_level_permission_tag_configuration.0.tag_rules.0.tag_multi_value_delimiter", ","),
				),
			},
		},
	})
}

func TestAccQuickSightDataSet_tags(t *testing.T) {
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, quicksight.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQuickSightDataSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTags1DataSetConfig(rId, rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightDataSetExists(resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func testAccCheckQuickSightDataSetExists(resourceName string, dataSet *quicksight.DataSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		awsAccountID, dataSetId, err := tfquicksight.ParseDataSetID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn()

		input := &quicksight.DescribeDataSetInput{
			AwsAccountId: aws.String(awsAccountID),
			DataSetId:    aws.String(dataSetId),
		}

		output, err := conn.DescribeDataSet(input)

		if err != nil {
			return err
		}

		if output == nil || output.DataSet == nil {
			return fmt.Errorf("QuickSight Data Set (%s) not found", rs.Primary.ID)
		}

		*dataSet = *output.DataSet

		return nil
	}
}

func testAccCheckQuickSightDataSetDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_quicksight_data_set" {
			continue
		}

		awsAccountID, dataSetId, err := tfquicksight.ParseDataSetID(rs.Primary.ID)
		if err != nil {
			return err
		}

		output, err := conn.DescribeDataSet(&quicksight.DescribeDataSetInput{
			AwsAccountId: aws.String(awsAccountID),
			DataSetId:    aws.String(dataSetId),
		})

		if tfawserr.ErrMessageContains(err, quicksight.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil && output.DataSet != nil {
			return fmt.Errorf("QuickSight Data Set (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccBaseDataSetConfig(rId string, rName string) string {
	return acctest.ConfigCompose(
		testAccBaseDataSourceConfig(rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_source" "test" {
  data_source_id = %[1]q
  name           = %[2]q

  parameters {
    s3 {
      manifest_file_location {
        bucket = aws_s3_bucket.test.bucket
        key    = aws_s3_object.test.key
      }
    }
  }

  type = "S3"
}
`, rId, rName))
}

func testAccDataSetConfig(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccBaseDataSetConfig(rId, rName),
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
				name = "ColumnId-1"
				type = "STRING"
			}
		}
	}
}
`, rId, rName))
}

func testAccDataSetConfigColumnGroups(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccBaseDataSetConfig(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_set" "test" {
	data_set_id = %[1]q
	name        = %[2]q
	import_mode = "SPICE"

	physical_table_map {
		physical_table_map_id = "unique_id"
		s3_source {
			data_source_arn = aws_quicksight_data_source.test.arn
			input_columns {
				name = "ColumnId-1"
				type = "STRING"
			}
		}
	}
	column_groups {
		geo_spatial_column_group {
			columns = ["column1"]
			country_code = "US"
			name = "test"
		}
	}
}
`, rId, rName))
}

func testAccDataSetConfigColumnLevelPermissionRules(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccBaseDataSetConfig(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_set" "test" {
	data_set_id = %[1]q
	name        = %[2]q
	import_mode = "SPICE"

	physical_table_map {
		physical_table_map_id = "unique_id"
		s3_source {
			data_source_arn = aws_quicksight_data_source.test.arn
			input_columns {
				name = "ColumnId-1"
				type = "STRING"
			}
		}
	}
	column_level_permission_rules {
		column_names = ["ColumnId-1"]
	}
}
`, rId, rName))
}

func testAccDataSetConfigDataSetUsageConfiguration(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccBaseDataSetConfig(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_set" "test" {
	data_set_id = %[1]q
	name        = %[2]q
	import_mode = "SPICE"

	physical_table_map {
		physical_table_map_id = "unique_id"
		s3_source {
			data_source_arn = aws_quicksight_data_source.test.arn
			input_columns {
				name = "ColumnId-1"
				type = "STRING"
			}
		}
	}
	data_set_usage_configuration {
		disable_use_as_direct_query_source = false
		disable_use_as_imported_source = false
	  }
}
`, rId, rName))
}

func testAccDataSetConfigFieldFolders(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccBaseDataSetConfig(rId, rName),
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
				name = "ColumnId-1"
				type = "STRING"
			}
		}
	}
	field_folders {
		field_folders_id = %[1]q
		columns = ["ColumnId-1"]
		description = "test"
	}
}
`, rId, rName))
}

func testAccDataSetConfigLogicalTableMap(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccBaseDataSetConfig(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_set" "test" {
	data_set_id = %[1]q
	name        = %[2]q
	import_mode = "SPICE"

	physical_table_map {
		physical_table_map_id = "unique_id"
		s3_source {
			data_source_arn = aws_quicksight_data_source.test.arn
			input_columns {
				name = "ColumnId-1"
				type = "STRING"
			}
		}
	}
	logical_table_map {
		alias = "Group 1"
		source {
			physical_table_id = "s3PhysicalTable"
		}
	}
}
`, rId, rName))
}

func testAccDataSetConfigPermissions(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccBaseDataSetConfig(rId, rName),
		testAccDataSource_UserConfig(rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_set" "test" {
	data_set_id = %[1]q
	name        = %[2]q
	import_mode = "SPICE"

	physical_table_map {
		physical_table_map_id = "unique_id"
		s3_source {
			data_source_arn = aws_quicksight_data_source.test.arn
			input_columns {
				name = "ColumnId-1"
				type = "STRING"
			}
		}
	}
	permissions {
    	actions = [
      		"quicksight:DescribeDataSet",
      		"quicksight:DescribeDataSetPermissions",
      		"quicksight:PassDataSet"
   		]
    	principal = aws_quicksight_user.test.arn
  	}
}
`, rId, rName))
}

func testAccDataSetConfigUpdatePermissions(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccBaseDataSetConfig(rId, rName),
		testAccDataSource_UserConfig(rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_set" "test" {
  data_set_id = %[1]q
  name        = %[2]q
  import_mode = "SPICE"

  physical_table_map {
	physical_table_map_id = "unique_id"
	  s3_source {
		  data_source_arn = aws_quicksight_data_source.test.arn
		  input_columns {
			  name = "ColumnId-1"
			  type = "STRING"
		  }
	  }
  }
  permission {
    actions = [
      "quicksight:DescribeDataSet",
      "quicksight:DescribeDataSetPermissions",
      "quicksight:PassDataSet",
      "quicksight:UpdateDataSet",
      "quicksight:DeleteDataSet",
      "quicksight:UpdateDataSetPermissions"
    ]

    principal = aws_quicksight_user.test.arn
  }
}
`, rId, rName))
}

func testAccDataSetConfigRowLevelPermissionDataSet(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccBaseDataSetConfig(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_set" "test" {
	data_set_id = %[1]q
	name        = %[2]q
	import_mode = "SPICE"

	physical_table_map {
		physical_table_map_id = "unique_id"
		s3_source {
			data_source_arn = aws_quicksight_data_source.test.arn
			input_columns {
				name = "ColumnId-1"
				type = "STRING"
			}
		}
	}
	row_level_permission_data_set {
		arn = "this.arn"
		permission_policy = "GRANT_ACCESS"
		format_version = "VERSION_1"
		namespace = "namespace"
		status = "ENABLED"
	  }
}
`, rId, rName))
}

func testAccDataSetConfigRowLevelPermissionTagConfiguration(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccBaseDataSetConfig(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_set" "test" {
	data_set_id = %[1]q
	name        = %[2]q
	import_mode = "SPICE"

	physical_table_map {
		physical_table_map_id = "unique_id"
		s3_source {
			data_source_arn = aws_quicksight_data_source.test.arn
			input_columns {
				name = "ColumnId-1"
				type = "STRING"
			}
		}
	}
	row_level_permission_tag_configuration {
		tag_rules {
			column_name = "columnname"
			tag_key = "uniquetagkey"
			match_all_value = "*"
			tag_multi_value_delimiter = ","
		}
	}
}
`, rId, rName))
}

func testAccTags1DataSetConfig(rId, rName, key, value string) string {
	return acctest.ConfigCompose(
		testAccBaseDataSetConfig(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_set" "test" {
	data_set_id = %[1]q
	name        = %[2]q
	import_mode = "SPICE"

	physical_table_map {
		physical_table_map_id = "unique_id"
		s3_source {
			data_source_arn = aws_quicksight_data_source.test.arn
			input_columns {
				name = "ColumnId-1"
				type = "STRING"
			}
		}
	}
	tags = {
		%[3]q = %[4]q
	}
}
`, rId, rName, key, value))
}
