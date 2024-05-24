// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQuickSightDataSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfigBasic(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(ctx, resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, "data_set_id", rId),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "quicksight", fmt.Sprintf("dataset/%s", rId)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "import_mode", "SPICE"),
					resource.TestCheckResourceAttr(resourceName, "physical_table_map.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "physical_table_map.0.s3_source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "physical_table_map.0.s3_source.0.input_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "physical_table_map.0.s3_source.0.input_columns.0.name", "Column1"),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfigBasic(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(ctx, resourceName, &dataSet),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfquicksight.ResourceDataSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccQuickSightDataSet_columnGroups(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfigColumnGroups(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(ctx, resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, "column_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "column_groups.0.geo_spatial_column_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "column_groups.0.geo_spatial_column_group.0.columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "column_groups.0.geo_spatial_column_group.0.columns.0", "Column1"),
					resource.TestCheckResourceAttr(resourceName, "column_groups.0.geo_spatial_column_group.0.country_code", "US"),
					resource.TestCheckResourceAttr(resourceName, "column_groups.0.geo_spatial_column_group.0.name", "test"),
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

func TestAccQuickSightDataSet_columnLevelPermissionRules(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfigColumnLevelPermissionRules(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(ctx, resourceName, &dataSet),
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

func TestAccQuickSightDataSet_dataSetUsageConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfigDataSetUsageConfiguration(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(ctx, resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, "data_set_usage_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_set_usage_configuration.0.disable_use_as_direct_query_source", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "data_set_usage_configuration.0.disable_use_as_imported_source", acctest.CtFalse),
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

func TestAccQuickSightDataSet_fieldFolders(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfigFieldFolders(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(ctx, resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, "field_folders.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "field_folders.0.columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "field_folders.0.columns.0", "Column1"),
					resource.TestCheckResourceAttr(resourceName, "field_folders.0.description", "test"),
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

func TestAccQuickSightDataSet_logicalTableMap(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfigLogicalTableMap(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(ctx, resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, "logical_table_map.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logical_table_map.0.alias", "Group1"),
					resource.TestCheckResourceAttr(resourceName, "logical_table_map.0.source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logical_table_map.0.source.0.physical_table_id", rId),
					resource.TestCheckResourceAttr(resourceName, "output_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "output_columns.0.name", "Column1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSetConfigUpdateLogicalTableMap(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(ctx, resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, "logical_table_map.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logical_table_map.0.alias", "Group1"),
					resource.TestCheckResourceAttr(resourceName, "logical_table_map.0.source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logical_table_map.0.source.0.physical_table_id", rId),
					resource.TestCheckResourceAttr(resourceName, "output_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "output_columns.0.name", "Column1"),
					resource.TestCheckResourceAttr(resourceName, "output_columns.1.name", "Column2"),
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

func TestAccQuickSightDataSet_permissions(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfigPermissions(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(ctx, resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct1),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "permissions.*", map[string]*regexp.Regexp{
						names.AttrPrincipal: regexache.MustCompile(fmt.Sprintf(`user/default/%s`, rName)),
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:DescribeDataSet"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:DescribeDataSetPermissions"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:PassDataSet"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:DescribeIngestion"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:ListIngestions"),
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
					testAccCheckDataSetExists(ctx, resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct1),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "permissions.*", map[string]*regexp.Regexp{
						names.AttrPrincipal: regexache.MustCompile(fmt.Sprintf(`user/default/%s`, rName)),
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:DescribeDataSet"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:DescribeDataSetPermissions"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:PassDataSet"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:DescribeIngestion"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:ListIngestions"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:UpdateDataSet"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:DeleteDataSet"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:CreateIngestion"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:CancelIngestion"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:UpdateDataSetPermissions"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSetConfigBasic(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(ctx, resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, "permission.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccQuickSightDataSet_permissionsMultiple(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfigPermissionsMultiple(rId, rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(ctx, resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", acctest.Ct3),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "permissions.*", map[string]*regexp.Regexp{
						names.AttrPrincipal: regexache.MustCompile(fmt.Sprintf(`user/default/%s`, rName)),
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:DescribeDataSet"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:DescribeDataSetPermissions"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:PassDataSet"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:DescribeIngestion"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permissions.*.actions.*", "quicksight:ListIngestions"),
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

func TestAccQuickSightDataSet_rowLevelPermissionTagConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfigRowLevelPermissionTagConfiguration(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(ctx, resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, "row_level_permission_tag_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "row_level_permission_tag_configuration.0.tag_rules.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "row_level_permission_tag_configuration.0.tag_rules.0.column_name", "Column1"),
					resource.TestCheckResourceAttr(resourceName, "row_level_permission_tag_configuration.0.tag_rules.0.tag_key", "uniquetagkey"),
					resource.TestCheckResourceAttr(resourceName, "row_level_permission_tag_configuration.0.tag_rules.0.match_all_value", "*"),
					resource.TestCheckResourceAttr(resourceName, "row_level_permission_tag_configuration.0.tag_rules.0.tag_multi_value_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "row_level_permission_tag_configuration.0.status", quicksight.StatusEnabled),
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

func TestAccQuickSightDataSet_refreshProperties(t *testing.T) {
	ctx := acctest.Context(t)
	// This test requires additional configuration of the QuickSight service role. Ensure
	// the role has the AWS managed "AmazonS3ReadOnlyAccess" and "AWSQuicksightAthenaAccess"
	// identity policies attached. The default LakeFormation data catalog settings may also
	// need adjusted, depending on the current configuration.
	//
	// Ref: https://docs.aws.amazon.com/lake-formation/latest/dg/change-settings.html
	if os.Getenv("QUICKSIGHT_ATHENA_TESTING_ENABLED") == "" {
		t.Skip("Environment variable QUICKSIGHT_ATHENA_TESTING_ENABLED is not set")
	}

	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfigRefreshProperties(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(ctx, resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, "refresh_properties.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "refresh_properties.0.refresh_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "refresh_properties.0.refresh_configuration.0.incremental_refresh.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "refresh_properties.0.refresh_configuration.0.incremental_refresh.0.lookback_window.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "refresh_properties.0.refresh_configuration.0.incremental_refresh.0.lookback_window.0.column_name", "column1"),
					resource.TestCheckResourceAttr(resourceName, "refresh_properties.0.refresh_configuration.0.incremental_refresh.0.lookback_window.0.size", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "refresh_properties.0.refresh_configuration.0.incremental_refresh.0.lookback_window.0.size_unit", "DAY"),
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

func TestAccQuickSightDataSet_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfigTags1(rId, rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(ctx, resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSetConfigTags2(rId, rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(ctx, resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDataSetConfigTags1(rId, rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(ctx, resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccQuickSightDataSet_noPhysicalTableMap(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfigNoPhysicalTableMap(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(ctx, resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, "data_set_id", rId),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "quicksight", fmt.Sprintf("dataset/%s", rId)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "import_mode", "SPICE"),
					resource.TestCheckResourceAttr(resourceName, "physical_table_map.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "logical_table_map.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "logical_table_map.0.logical_table_map_id", "joined"),
					resource.TestCheckResourceAttr(resourceName, "logical_table_map.0.alias", "j"),
					resource.TestCheckResourceAttr(resourceName, "logical_table_map.0.source.0.join_instruction.0.right_operand", "right"),
					resource.TestCheckResourceAttr(resourceName, "logical_table_map.0.source.0.join_instruction.0.left_operand", "left"),
					resource.TestCheckResourceAttr(resourceName, "logical_table_map.0.source.0.join_instruction.0.type", "INNER"),
					resource.TestCheckResourceAttr(resourceName, "logical_table_map.0.source.0.join_instruction.0.on_clause", "Column1 = Column2"),
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

func testAccCheckDataSetExists(ctx context.Context, resourceName string, dataSet *quicksight.DataSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		awsAccountID, dataSetId, err := tfquicksight.ParseDataSetID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn(ctx)

		input := &quicksight.DescribeDataSetInput{
			AwsAccountId: aws.String(awsAccountID),
			DataSetId:    aws.String(dataSetId),
		}

		output, err := conn.DescribeDataSetWithContext(ctx, input)
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

func testAccCheckDataSetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_data_set" {
				continue
			}

			awsAccountID, dataSetId, err := tfquicksight.ParseDataSetID(rs.Primary.ID)
			if err != nil {
				return err
			}

			output, err := conn.DescribeDataSetWithContext(ctx, &quicksight.DescribeDataSetInput{
				AwsAccountId: aws.String(awsAccountID),
				DataSetId:    aws.String(dataSetId),
			})
			if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
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
}

func testAccDataSetConfigBase(rId string, rName string) string {
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

func testAccDataSetConfigBasic(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccDataSetConfigBase(rId, rName),
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
      upload_settings {
        format = "JSON"
      }
    }
  }
}
`, rId, rName))
}

func testAccDataSetConfigColumnGroups(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccDataSetConfigBase(rId, rName),
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
      upload_settings {
        format = "JSON"
      }
    }
  }
  logical_table_map {
    logical_table_map_id = %[1]q
    alias                = "Group1"
    source {
      physical_table_id = %[1]q
    }
    data_transforms {
      tag_column_operation {
        column_name = "Column1"
        tags {
          column_geographic_role = "STATE"
        }
      }
    }
  }
  column_groups {
    geo_spatial_column_group {
      columns      = ["Column1"]
      country_code = "US"
      name         = "test"
    }
  }
}
`, rId, rName))
}

func testAccDataSetConfigColumnLevelPermissionRules(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccDataSetConfigBase(rId, rName),
		testAccDataSource_UserConfig(rName),
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
      upload_settings {
        format = "JSON"
      }
    }
  }
  column_level_permission_rules {
    column_names = ["Column1"]
    principals   = [aws_quicksight_user.test.arn]
  }
}
`, rId, rName))
}

func testAccDataSetConfigDataSetUsageConfiguration(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccDataSetConfigBase(rId, rName),
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
      upload_settings {
        format = "JSON"
      }
    }
  }
  data_set_usage_configuration {
    disable_use_as_direct_query_source = false
    disable_use_as_imported_source     = false
  }
}
`, rId, rName))
}

func testAccDataSetConfigFieldFolders(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccDataSetConfigBase(rId, rName),
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
      upload_settings {
        format = "JSON"
      }
    }
  }
  field_folders {
    field_folders_id = %[1]q
    columns          = ["Column1"]
    description      = "test"
  }
}
`, rId, rName))
}

func testAccDataSetConfigLogicalTableMap(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccDataSetConfigBase(rId, rName),
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
      upload_settings {}
    }
  }
  logical_table_map {
    logical_table_map_id = %[1]q
    alias                = "Group1"
    source {
      physical_table_id = %[1]q
    }
  }
}
`, rId, rName))
}

func testAccDataSetConfigUpdateLogicalTableMap(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccDataSetConfigBase(rId, rName),
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
      create_columns_operation {
        columns {
          column_id   = "Column2"
          column_name = "Column2"
          expression  = "Column1"
        }
      }
    }
  }
}
`, rId, rName))
}

func testAccDataSetConfigPermissions(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccDataSetConfigBase(rId, rName),
		testAccDataSource_UserConfig(rName),
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
      upload_settings {
        format = "JSON"
      }
    }
  }
  permissions {
    actions = [
      "quicksight:DescribeDataSet",
      "quicksight:DescribeDataSetPermissions",
      "quicksight:PassDataSet",
      "quicksight:DescribeIngestion",
      "quicksight:ListIngestions",
    ]
    principal = aws_quicksight_user.test.arn
  }
}
`, rId, rName))
}

func testAccDataSetConfigUpdatePermissions(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccDataSetConfigBase(rId, rName),
		testAccDataSource_UserConfig(rName),
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
      upload_settings {
        format = "JSON"
      }
    }
  }
  permissions {
    actions = [
      "quicksight:DescribeDataSet",
      "quicksight:DescribeDataSetPermissions",
      "quicksight:PassDataSet",
      "quicksight:DescribeIngestion",
      "quicksight:ListIngestions",
      "quicksight:UpdateDataSet",
      "quicksight:DeleteDataSet",
      "quicksight:CreateIngestion",
      "quicksight:CancelIngestion",
      "quicksight:UpdateDataSetPermissions",
    ]
    principal = aws_quicksight_user.test.arn
  }
}
`, rId, rName))
}

func testAccDataSetConfigPermissionsMultiple(rId, rName string, count int) string {
	return acctest.ConfigCompose(
		testAccDataSetConfigBase(rId, rName),
		testAccDataSource_UserConfigMultiple(rName, count),
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
      upload_settings {
        format = "JSON"
      }
    }
  }

  dynamic "permissions" {
    for_each = aws_quicksight_user.test
    content {
      actions = [
        "quicksight:DescribeDataSet",
        "quicksight:DescribeDataSetPermissions",
        "quicksight:PassDataSet",
        "quicksight:DescribeIngestion",
        "quicksight:ListIngestions",
      ]
      principal = aws_quicksight_user.test[permissions.key].arn
    }
  }
}
`, rId, rName))
}

func testAccDataSetConfigRowLevelPermissionTagConfiguration(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccDataSetConfigBase(rId, rName),
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
      upload_settings {
        format = "JSON"
      }
    }
  }
  row_level_permission_tag_configuration {
    status = "ENABLED"
    tag_rules {
      column_name               = "Column1"
      tag_key                   = "uniquetagkey"
      match_all_value           = "*"
      tag_multi_value_delimiter = ","
    }
  }
}
`, rId, rName))
}

func testAccDataSetConfigRefreshProperties(rId, rName string) string {
	// NOTE: Must use Athena data source here as incremental refresh is not supported by S3
	return acctest.ConfigCompose(
		testAccBaseDataSourceConfig(rName),
		fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[2]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[2]q
  database_name = aws_glue_catalog_database.test.name
  table_type    = "EXTERNAL_TABLE"

  parameters = {
    EXTERNAL       = "TRUE"
    classification = "json"
  }

  storage_descriptor {
    location      = "s3://${aws_s3_bucket.test.id}/data/"
    input_format  = "org.apache.hadoop.mapred.TextInputFormat"
    output_format = "org.apache.hadoop.hive.ql.io.HiveIgnoreKeyTextOutputFormat"

    ser_de_info {
      name                  = "jsonserde"
      serialization_library = "org.openx.data.jsonserde.JsonSerDe"
      parameters = {
        "serialization.format" = "1"
      }
    }
    columns {
      name = "column1"
      type = "date"
    }
  }
}

resource "aws_quicksight_data_source" "test" {
  data_source_id = %[1]q
  name           = %[2]q
  type           = "ATHENA"
  parameters {
    athena {
      work_group = "primary"
    }
  }
  ssl_properties {
    disable_ssl = false
  }
}

resource "aws_quicksight_data_set" "test" {
  data_set_id = %[1]q
  name        = %[2]q
  import_mode = "SPICE"

  physical_table_map {
    physical_table_map_id = %[1]q
    relational_table {
      data_source_arn = aws_quicksight_data_source.test.arn
      catalog         = "AwsDataCatalog"
      schema          = aws_glue_catalog_database.test.name
      name            = aws_glue_catalog_table.test.name
      input_columns {
        name = "column1"
        type = "DATETIME"
      }
    }
  }
  refresh_properties {
    refresh_configuration {
      incremental_refresh {
        lookback_window {
          column_name = "column1"
          size        = 1
          size_unit   = "DAY"
        }
      }
    }
  }
}
`, rId, rName))
}

func testAccDataSetConfigTags1(rId, rName, key1, value1 string) string {
	return acctest.ConfigCompose(
		testAccDataSetConfigBase(rId, rName),
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
      upload_settings {
        format = "JSON"
      }
    }
  }
  tags = {
    %[3]q = %[4]q
  }
}
`, rId, rName, key1, value1))
}

func testAccDataSetConfigTags2(rId, rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(
		testAccDataSetConfigBase(rId, rName),
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
      upload_settings {
        format = "JSON"
      }
    }
  }
  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rId, rName, key1, value1, key2, value2))
}

func testAccDataSetConfigNoPhysicalTableMap(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccDataSetConfigBase(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_set" "left" {
  data_set_id = "%[1]s-left"
  name        = "%[2]s-left"
  import_mode = "SPICE"

  physical_table_map {
    physical_table_map_id = "%[1]s-left"
    s3_source {
      data_source_arn = aws_quicksight_data_source.test.arn
      input_columns {
        name = "Column1"
        type = "STRING"
      }
      upload_settings {
        format = "JSON"
      }
    }
  }
}

resource "aws_quicksight_data_set" "right" {
  data_set_id = "%[1]s-right"
  name        = "%[2]s-right"
  import_mode = "SPICE"

  physical_table_map {
    physical_table_map_id = "%[1]s-right"
    s3_source {
      data_source_arn = aws_quicksight_data_source.test.arn
      input_columns {
        name = "Column2"
        type = "STRING"
      }
      upload_settings {
        format = "JSON"
      }
    }
  }
}

resource "aws_quicksight_data_set" "test" {
  data_set_id = %[1]q
  name        = %[2]q
  import_mode = "SPICE"

  logical_table_map {
    logical_table_map_id = "right"
    alias                = "r"
    source {
      data_set_arn = aws_quicksight_data_set.right.arn
    }
  }

  logical_table_map {
    logical_table_map_id = "left"
    alias                = "l"
    source {
      data_set_arn = aws_quicksight_data_set.left.arn
    }
  }

  logical_table_map {
    logical_table_map_id = "joined"
    alias                = "j"
    source {
      join_instruction {
        left_operand  = "left"
        right_operand = "right"
        type          = "INNER"
        on_clause     = "Column1 = Column2"
      }
    }
  }
}
`, rId, rName))
}
