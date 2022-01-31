package quicksight_test

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
)

func TestAccQuickSightDataSet_basic(t *testing.T) {
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.dset"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		ErrorCheck:        acctest.ErrorCheck(t, quicksight.EndpointsID),
		CheckDestroy:      testAccCheckQuickSightDataSetDestroy,
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
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.dset"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		ErrorCheck:        acctest.ErrorCheck(t, quicksight.EndpointsID),
		CheckDestroy:      testAccCheckQuickSightDataSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfig(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightDataSetExists(resourceName, &dataSet),
					acctest.CheckResourceDisappears(acctest.Provider, tfquicksight.ResourceDataSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccQuickSightDataSet_column_groups(t *testing.T) {
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.dset"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		ErrorCheck:        acctest.ErrorCheck(t, quicksight.EndpointsID),
		CheckDestroy:      testAccCheckQuickSightDataSetDestroy,
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
	resourceName := "aws_quicksight_data_set.dset"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		ErrorCheck:        acctest.ErrorCheck(t, quicksight.EndpointsID),
		CheckDestroy:      testAccCheckQuickSightDataSetDestroy,
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
	resourceName := "aws_quicksight_data_set.dset"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		ErrorCheck:        acctest.ErrorCheck(t, quicksight.EndpointsID),
		CheckDestroy:      testAccCheckQuickSightDataSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfigFieldFolders(rId, rName),
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
	resourceName := "aws_quicksight_data_set.dset"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		ErrorCheck:        acctest.ErrorCheck(t, quicksight.EndpointsID),
		CheckDestroy:      testAccCheckQuickSightDataSetDestroy,
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

func TestAccQuickSightDataSet_logical_table_map(t *testing.T) {
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.dset"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		ErrorCheck:        acctest.ErrorCheck(t, quicksight.EndpointsID),
		CheckDestroy:      testAccCheckQuickSightDataSetDestroy,
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

func TestQuickSightDataSetPermissionsDiff(t *testing.T) {
	testCases := []struct {
		name            string
		oldPermissions  []interface{}
		newPermissions  []interface{}
		expectedGrants  []*quicksight.ResourcePermission
		expectedRevokes []*quicksight.ResourcePermission
	}{
		{
			name:            "no changes;empty",
			oldPermissions:  []interface{}{},
			newPermissions:  []interface{}{},
			expectedGrants:  nil,
			expectedRevokes: nil,
		},
		{
			name: "no changes;same",
			oldPermissions: []interface{}{
				map[string]interface{}{
					"principal": "principal1",
					"actions": schema.NewSet(schema.HashString, []interface{}{
						"action1",
						"action2",
					}),
				},
			},
			newPermissions: []interface{}{
				map[string]interface{}{
					"principal": "principal1",
					"actions": schema.NewSet(schema.HashString, []interface{}{
						"action1",
						"action2",
					}),
				}},

			expectedGrants:  nil,
			expectedRevokes: nil,
		},
		{
			name:           "grant only",
			oldPermissions: []interface{}{},
			newPermissions: []interface{}{
				map[string]interface{}{
					"principal": "principal1",
					"actions": schema.NewSet(schema.HashString, []interface{}{
						"action1",
						"action2",
					}),
				},
			},
			expectedGrants: []*quicksight.ResourcePermission{
				{
					Actions:   aws.StringSlice([]string{"action1", "action2"}),
					Principal: aws.String("principal1"),
				},
			},
			expectedRevokes: nil,
		},
		{
			name: "revoke only",
			oldPermissions: []interface{}{
				map[string]interface{}{
					"principal": "principal1",
					"actions": schema.NewSet(schema.HashString, []interface{}{
						"action1",
						"action2",
					}),
				},
			},
			newPermissions: []interface{}{},
			expectedGrants: nil,
			expectedRevokes: []*quicksight.ResourcePermission{
				{
					Actions:   aws.StringSlice([]string{"action1", "action2"}),
					Principal: aws.String("principal1"),
				},
			},
		},
		{
			name: "grant new action",
			oldPermissions: []interface{}{
				map[string]interface{}{
					"principal": "principal1",
					"actions": schema.NewSet(schema.HashString, []interface{}{
						"action1",
					}),
				},
			},
			newPermissions: []interface{}{
				map[string]interface{}{
					"principal": "principal1",
					"actions": schema.NewSet(schema.HashString, []interface{}{
						"action1",
						"action2",
					}),
				},
			},
			expectedGrants: []*quicksight.ResourcePermission{
				{
					Actions:   aws.StringSlice([]string{"action2"}),
					Principal: aws.String("principal1"),
				},
			},
			expectedRevokes: nil,
		},
		{
			name: "revoke old action",
			oldPermissions: []interface{}{
				map[string]interface{}{
					"principal": "principal1",
					"actions": schema.NewSet(schema.HashString, []interface{}{
						"oldAction",
						"onlyOldAction",
					}),
				},
			},
			newPermissions: []interface{}{
				map[string]interface{}{
					"principal": "principal1",
					"actions": schema.NewSet(schema.HashString, []interface{}{
						"oldAction",
					}),
				},
			},
			expectedGrants: nil,
			expectedRevokes: []*quicksight.ResourcePermission{
				{
					Actions:   aws.StringSlice([]string{"onlyOldAction"}),
					Principal: aws.String("principal1"),
				},
			},
		},
		{
			name: "multiple permissions",
			oldPermissions: []interface{}{
				map[string]interface{}{
					"principal": "principal1",
					"actions": schema.NewSet(schema.HashString, []interface{}{
						"action1",
						"action2",
					}),
				},
				map[string]interface{}{
					"principal": "principal2",
					"actions": schema.NewSet(schema.HashString, []interface{}{
						"action1",
						"action3",
						"action4",
					}),
				},
				map[string]interface{}{
					"principal": "principal3",
					"actions": schema.NewSet(schema.HashString, []interface{}{
						"action5",
					}),
				},
			},
			newPermissions: []interface{}{
				map[string]interface{}{
					"principal": "principal1",
					"actions": schema.NewSet(schema.HashString, []interface{}{
						"action1",
						"action2",
					}),
				},
				map[string]interface{}{
					"principal": "principal2",
					"actions": schema.NewSet(schema.HashString, []interface{}{
						"action3",
						"action5",
					}),
				},
			},
			expectedGrants: []*quicksight.ResourcePermission{
				{
					Actions:   aws.StringSlice([]string{"action5"}),
					Principal: aws.String("principal2"),
				},
			},
			expectedRevokes: []*quicksight.ResourcePermission{
				{
					Actions:   aws.StringSlice([]string{"action1", "action4"}),
					Principal: aws.String("principal2"),
				},
				{
					Actions:   aws.StringSlice([]string{"action5"}),
					Principal: aws.String("principal3"),
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			toGrant, toRevoke := tfquicksight.DiffPermissions(testCase.oldPermissions, testCase.newPermissions)
			if !reflect.DeepEqual(toGrant, testCase.expectedGrants) {
				t.Fatalf("Expected: %v, got: %v", testCase.expectedGrants, toGrant)
			}

			if !reflect.DeepEqual(toRevoke, testCase.expectedRevokes) {
				t.Fatalf("Expected: %v, got: %v", testCase.expectedRevokes, toRevoke)
			}
		})
	}
}

func TestAccQuickSightDataSet_permissions(t *testing.T) {
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		ErrorCheck:        acctest.ErrorCheck(t, quicksight.EndpointsID),
		CheckDestroy:      testAccCheckQuickSightDataSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfig_Permissions(rId, rName),
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
				Config: testAccDataSetConfig_UpdatePermissions(rId, rName),
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

func TestAccQuickSightDataSet_row_level_permission_data_set(t *testing.T) {
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.dset"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		ErrorCheck:        acctest.ErrorCheck(t, quicksight.EndpointsID),
		CheckDestroy:      testAccCheckQuickSightDataSetDestroy,
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

func TestAccQuickSightDataSet_row_level_permission_tag_configuration(t *testing.T) {
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.dset"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		ErrorCheck:        acctest.ErrorCheck(t, quicksight.EndpointsID),
		CheckDestroy:      testAccCheckQuickSightDataSetDestroy,
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

// func TestAccQuickSightDataSet_attribute(t *testing.T) {
// 	var dataSet quicksight.DataSet
// 	resourceName := "aws_quicksight_data_set.dset"
// 	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
// 	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
//
// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck:          func() { acctest.PreCheck(t) },
// 		ProviderFactories: acctest.ProviderFactories,
// 		ErrorCheck:        acctest.ErrorCheck(t, quicksight.EndpointsID),
// 		CheckDestroy:      testAccCheckQuickSightDataSetDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccDataSetConfigAttribute(rId, rName),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckQuickSightDataSetExists(resourceName, &dataSet),
// 				),
// 			},
// 		},
// 	})
// }
//
// func TestAccQuickSightDataSet_attribute(t *testing.T) {
// 	var dataSet quicksight.DataSet
// 	resourceName := "aws_quicksight_data_set.dset"
// 	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
// 	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
//
// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck:          func() { acctest.PreCheck(t) },
// 		ProviderFactories: acctest.ProviderFactories,
// 		ErrorCheck:        acctest.ErrorCheck(t, quicksight.EndpointsID),
// 		CheckDestroy:      testAccCheckQuickSightDataSetDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccDataSetConfigAttribute(rId, rName),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckQuickSightDataSetExists(resourceName, &dataSet),
// 				),
// 			},
// 		},
// 	})
// }

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

		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn

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
	conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn
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
data "aws_partition" "partition" {}

resource "aws_s3_bucket" "s3bucket" {
	acl           = "public-read"
	bucket        = %[1]q
	force_destroy = true
}
		
resource "aws_s3_bucket_object" "object" {
	bucket  = aws_s3_bucket.test.bucket
	key     = %[1]q
	content = <<EOF
	{
		"fileLocations": [
			{
				"URIs": [
					"https://${aws_s3_bucket.test.bucket}.s3.${data.aws_partition.current.dns_suffix}/%[1]s"
				]
			}
		],
		"globalUploadSettings": {
			"format": "JSON"
		}
	}
	EOF
	acl     = "public-read"
}
		
resource "aws_quicksight_data_source" "dsource" {
	data_source_id = %[1]q
	name           = %[2]q
		
	parameters {
		s3 {
			manifest_file_location {
				bucket = aws_s3_bucket.test.bucket
				key    = aws_s3_bucket_object.test.key
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
resource "aws_quicksight_data_set" "dset" {
	
	data_set_id = %[1]q
	name        = %[2]q
	import_mode = "SPICE"
	physical_table_map {
		s3_source {
			data_source_arn = aws_quicksight_data_source.dsource.arn
			input_columns {
				name = "ColumnId-1"
				type = "STRING"
			}
		}
	}
}
`, rId, rName))
}

//REQUIRES LOGICAL_TABLE_MAP
func testAccDataSetConfigColumnGroups(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccBaseDataSetConfig(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_set" "dset" {
	
	data_set_id = %[1]q
	name        = %[2]q
	import_mode = "SPICE"
	physical_table_map {
		s3_source {
			data_source_arn = aws_quicksight_data_source.dsource.arn
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

// ERROR: ColumnLevelPermissionRules is only for enterprise accounts. This account 123456789 is not an enterprise account.
// Should this be tested?
func testAccDataSetConfigColumnLevelPermissionRules(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccBaseDataSetConfig(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_set" "dset" {
	
	data_set_id = %[1]q
	name        = %[2]q
	import_mode = "SPICE"
	physical_table_map {
		s3_source {
			data_source_arn = aws_quicksight_data_source.dsource.arn
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
resource "aws_quicksight_data_set" "dset" {
	
	data_set_id = %[1]q
	name        = %[2]q
	import_mode = "SPICE"
	physical_table_map {
		s3_source {
			data_source_arn = aws_quicksight_data_source.dsource.arn
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
resource "aws_quicksight_data_set" "dset" {
	
	data_set_id = %[1]q
	name        = %[2]q
	import_mode = "SPICE"
	physical_table_map {
		s3_source {
			data_source_arn = aws_quicksight_data_source.dsource.arn
			input_columns {
				name = "ColumnId-1"
				type = "STRING"
			}
		}
	}
	field_folders {
		columns = ["ColumnId-1"]
		description = "test"
	}
}
`, rId, rName))
}

// DOES NOT WORK -- AWS TICKET IN PROGRESS
func testAccDataSetConfigLogicalTableMap(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccBaseDataSetConfig(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_set" "dset" {
	
	data_set_id = %[1]q
	name        = %[2]q
	import_mode = "SPICE"
	physical_table_map {
		s3_source {
			data_source_arn = aws_quicksight_data_source.dsource.arn
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

// NEED ACCOUNT ACCESS
func testAccDataSetConfig_Permissions(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccBaseDataSetConfig(rId, rName),
		testAccDataSource_UserConfig(rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_set" "test" {
	data_set_id = %[1]q
	name        = %[2]q
	import_mode = "SPICE"
	physical_table_map {
		s3_source {
			data_source_arn = aws_quicksight_data_source.dsource.arn
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
      		"quicksight:PassDataSet"
   		]
    	principal = aws_quicksight_user.test.arn
  	}
}
`, rId, rName))
}

// NEED ACCOUNT ACCESS
func testAccDataSetConfig_UpdatePermissions(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccBaseDataSetConfig(rId, rName),
		testAccDataSource_UserConfig(rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_set" "test" {
  data_set_id = %[1]q
  name        = %[2]q
  import_mode = "SPICE"
  physical_table_map {
	  s3_source {
		  data_source_arn = aws_quicksight_data_source.dsource.arn
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

// NEED TO KNOW HOW TO REFERENCE CURRENT DATA SET
func testAccDataSetConfigRowLevelPermissionDataSet(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccBaseDataSetConfig(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_set" "dset" {
	
	data_set_id = %[1]q
	name        = %[2]q
	import_mode = "SPICE"
	physical_table_map {
		s3_source {
			data_source_arn = aws_quicksight_data_source.dsource.arn
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
resource "aws_quicksight_data_set" "dset" {
	
	data_set_id = %[1]q
	name        = %[2]q
	import_mode = "SPICE"
	physical_table_map {
		s3_source {
			data_source_arn = aws_quicksight_data_source.dsource.arn
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
