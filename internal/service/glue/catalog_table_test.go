package glue_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(glue.EndpointsID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"AccessDeniedException: Operation not allowed",
	)
}

func TestAccGlueCatalogTable_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccCatalogTableConfig_basic(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableExists(resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("table/%s/%s", rName, rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_table.#", "0"),
					acctest.CheckResourceAttrAccountID(resourceName, "catalog_id"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "partition_index.#", "0"),
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

func TestAccGlueCatalogTable_columnParameters(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccCatalogTableConfig_columnParameters(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.name", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.parameters.param2", "param2_val"),
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

func TestAccGlueCatalogTable_full(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "A test table from terraform"
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccCatalogTableConfig_full(rName, description),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "owner", "my_owner"),
					resource.TestCheckResourceAttr(resourceName, "retention", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.name", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.type", "int"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.comment", "my_column1_comment"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.1.name", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.1.type", "string"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.1.comment", "my_column2_comment"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.location", "my_location"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.input_format", "SequenceFileInputFormat"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.output_format", "SequenceFileInputFormat"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.compressed", "false"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.number_of_buckets", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.name", "ser_de_name"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.parameters.param1", "param_val_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.serialization_library", "org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.bucket_columns.0", "bucket_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.sort_columns.0.column", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.sort_columns.0.sort_order", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.parameters.param1", "param1_val"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_value_location_maps.my_column_1", "my_column_1_val_loc_map"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_values.0", "skewed_val_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.stored_as_sub_directories", "false"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.0.name", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.0.type", "int"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.0.comment", "my_column_1_comment"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.1.name", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.1.type", "string"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.1.comment", "my_column_2_comment"),
					resource.TestCheckResourceAttr(resourceName, "view_original_text", "view_original_text_1"),
					resource.TestCheckResourceAttr(resourceName, "view_expanded_text", "view_expanded_text_1"),
					resource.TestCheckResourceAttr(resourceName, "table_type", "VIRTUAL_VIEW"),
					resource.TestCheckResourceAttr(resourceName, "parameters.param1", "param1_val"),
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

func TestAccGlueCatalogTable_Update_addValues(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "A test table from terraform"
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccCatalogTableConfig_basic(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:  testAccCatalogTableConfig_full(rName, description),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "owner", "my_owner"),
					resource.TestCheckResourceAttr(resourceName, "retention", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.name", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.type", "int"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.comment", "my_column1_comment"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.1.name", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.1.type", "string"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.1.comment", "my_column2_comment"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.location", "my_location"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.input_format", "SequenceFileInputFormat"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.output_format", "SequenceFileInputFormat"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.compressed", "false"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.number_of_buckets", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.name", "ser_de_name"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.parameters.param1", "param_val_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.serialization_library", "org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.bucket_columns.0", "bucket_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.sort_columns.0.column", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.sort_columns.0.sort_order", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.parameters.param1", "param1_val"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_value_location_maps.my_column_1", "my_column_1_val_loc_map"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_values.0", "skewed_val_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.stored_as_sub_directories", "false"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.0.name", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.0.type", "int"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.0.comment", "my_column_1_comment"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.1.name", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.1.type", "string"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.1.comment", "my_column_2_comment"),
					resource.TestCheckResourceAttr(resourceName, "view_original_text", "view_original_text_1"),
					resource.TestCheckResourceAttr(resourceName, "view_expanded_text", "view_expanded_text_1"),
					resource.TestCheckResourceAttr(resourceName, "table_type", "VIRTUAL_VIEW"),
					resource.TestCheckResourceAttr(resourceName, "parameters.param1", "param1_val"),
				),
			},
		},
	})
}

func TestAccGlueCatalogTable_Update_replaceValues(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "A test table from terraform"
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccCatalogTableConfig_full(rName, description),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "owner", "my_owner"),
					resource.TestCheckResourceAttr(resourceName, "retention", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.name", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.type", "int"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.comment", "my_column1_comment"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.1.name", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.1.type", "string"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.1.comment", "my_column2_comment"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.location", "my_location"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.input_format", "SequenceFileInputFormat"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.output_format", "SequenceFileInputFormat"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.compressed", "false"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.number_of_buckets", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.name", "ser_de_name"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.parameters.param1", "param_val_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.serialization_library", "org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.bucket_columns.0", "bucket_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.sort_columns.0.column", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.sort_columns.0.sort_order", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.parameters.param1", "param1_val"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_value_location_maps.my_column_1", "my_column_1_val_loc_map"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_values.0", "skewed_val_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.stored_as_sub_directories", "false"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.0.name", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.0.type", "int"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.0.comment", "my_column_1_comment"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.1.name", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.1.type", "string"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.1.comment", "my_column_2_comment"),
					resource.TestCheckResourceAttr(resourceName, "view_original_text", "view_original_text_1"),
					resource.TestCheckResourceAttr(resourceName, "view_expanded_text", "view_expanded_text_1"),
					resource.TestCheckResourceAttr(resourceName, "table_type", "VIRTUAL_VIEW"),
					resource.TestCheckResourceAttr(resourceName, "parameters.param1", "param1_val"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:  testAccCatalogTableConfig_fullReplacedValues(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "A test table from terraform2"),
					resource.TestCheckResourceAttr(resourceName, "owner", "my_owner2"),
					resource.TestCheckResourceAttr(resourceName, "retention", "2"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.name", "my_column_12"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.type", "date"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.comment", "my_column1_comment2"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.1.name", "my_column_22"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.1.type", "timestamp"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.1.comment", "my_column2_comment2"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.location", "my_location2"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.input_format", "TextInputFormat"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.output_format", "TextInputFormat"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.compressed", "true"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.number_of_buckets", "12"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.name", "ser_de_name2"),
					resource.TestCheckNoResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.parameters.param1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.parameters.param2", "param_val_12"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.serialization_library", "org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe2"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.bucket_columns.0", "bucket_column_12"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.bucket_columns.1", "bucket_column_2"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.sort_columns.0.column", "my_column_12"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.sort_columns.0.sort_order", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "storage_descriptor.0.parameters.param1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.parameters.param12", "param1_val2"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_names.0", "my_column_12"),
					resource.TestCheckNoResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_value_location_maps.my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_value_location_maps.my_column_12", "my_column_1_val_loc_map2"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_values.0", "skewed_val_12"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_values.1", "skewed_val_2"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.stored_as_sub_directories", "true"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.0.name", "my_column_12"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.0.type", "date"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.0.comment", "my_column_1_comment2"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.1.name", "my_column_22"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.1.type", "timestamp"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.1.comment", "my_column_2_comment2"),
					resource.TestCheckResourceAttr(resourceName, "view_original_text", "view_original_text_12"),
					resource.TestCheckResourceAttr(resourceName, "view_expanded_text", "view_expanded_text_12"),
					resource.TestCheckResourceAttr(resourceName, "table_type", "EXTERNAL_TABLE"),
					//resource.TestCheckResourceAttr(resourceName, "parameters.param1", "param1_val"),
					resource.TestCheckResourceAttr(resourceName, "parameters.param2", "param1_val2"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11784
func TestAccGlueCatalogTable_StorageDescriptor_emptyBlock(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogTableConfig_storageDescriptorEmptyConfigurationBlock(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableExists(resourceName),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11784
func TestAccGlueCatalogTable_StorageDescriptorSerDeInfo_emptyBlock(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogTableConfig_storageDescriptorSerDeInfoEmptyConfigurationBlock(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableExists(resourceName),
				),
				// Expect non-empty instead of panic
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccGlueCatalogTable_StorageDescriptorSerDeInfo_updateValues(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccCatalogTableConfig_storageDescriptorSerDeInfo(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.name", "ser_de_name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:  testAccCatalogTableConfig_storageDescriptorSerDeInfoUpdate(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.parameters.param1", "param_val_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.serialization_library", "org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11784
func TestAccGlueCatalogTable_StorageDescriptorSkewedInfo_emptyBlock(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogTableConfig_storageDescriptorSkewedInfoEmptyConfigurationBlock(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableExists(resourceName),
				),
				// Expect non-empty instead of panic
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccGlueCatalogTable_StorageDescriptor_schemaReference(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogTableConfig_storageDescriptorSchemaReference(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.schema_reference.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.schema_reference.0.schema_version_number", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.schema_reference.0.schema_id.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_descriptor.0.schema_reference.0.schema_id.0.schema_name", "aws_glue_schema.test", "schema_name"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_descriptor.0.schema_reference.0.schema_id.0.registry_name", "aws_glue_schema.test", "registry_name"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCatalogTableConfig_storageDescriptorSchemaReferenceARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.schema_reference.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.schema_reference.0.schema_version_number", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.schema_reference.0.schema_id.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_descriptor.0.schema_reference.0.schema_id.0.schema_arn", "aws_glue_schema.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.#", "2"),
				),
			},
		},
	})
}

func TestAccGlueCatalogTable_StorageDescriptor_schemaReferenceARN(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogTableConfig_storageDescriptorSchemaReferenceARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.schema_reference.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.schema_reference.0.schema_version_number", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.schema_reference.0.schema_id.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_descriptor.0.schema_reference.0.schema_id.0.schema_arn", "aws_glue_schema.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.#", "2"),
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

func TestAccGlueCatalogTable_partitionIndexesSingle(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccCatalogTableConfig_partitionIndexesSingle(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableExists(resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("table/%s/%s", rName, rName)),
					resource.TestCheckResourceAttr(resourceName, "partition_index.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "partition_index.0.index_name", rName),
					resource.TestCheckResourceAttr(resourceName, "partition_index.0.index_status", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "partition_index.0.keys.#", "2"),
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

func TestAccGlueCatalogTable_partitionIndexesMultiple(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccCatalogTableConfig_partitionIndexesMultiple(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableExists(resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("table/%s/%s", rName, rName)),
					resource.TestCheckResourceAttr(resourceName, "partition_index.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "partition_index.0.index_name", rName),
					resource.TestCheckResourceAttr(resourceName, "partition_index.0.index_status", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "partition_index.0.keys.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "partition_index.1.index_name", fmt.Sprintf("%s-2", rName)),
					resource.TestCheckResourceAttr(resourceName, "partition_index.1.index_status", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "partition_index.1.keys.#", "1"),
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

func TestAccGlueCatalogTable_Disappears_database(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccCatalogTableConfig_basic(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfglue.ResourceCatalogDatabase(), "aws_glue_catalog_database.test"),
					acctest.CheckResourceDisappears(acctest.Provider, tfglue.ResourceCatalogTable(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccGlueCatalogTable_targetTable(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccCatalogTableConfig_target(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_table.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "target_table.0.catalog_id", "aws_glue_catalog_table.test2", "catalog_id"),
					resource.TestCheckResourceAttrPair(resourceName, "target_table.0.database_name", "aws_glue_catalog_table.test2", "database_name"),
					resource.TestCheckResourceAttrPair(resourceName, "target_table.0.name", "aws_glue_catalog_table.test2", "name"),
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

func TestAccGlueCatalogTable_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccCatalogTableConfig_basic(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfglue.ResourceCatalogTable(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfglue.ResourceCatalogTable(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCatalogTableConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name
}
`, rName)
}

func testAccCatalogTableConfig_full(rName, desc string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name               = %[1]q
  database_name      = aws_glue_catalog_database.test.name
  description        = %[2]q
  owner              = "my_owner"
  retention          = 1
  table_type         = "VIRTUAL_VIEW"
  view_expanded_text = "view_expanded_text_1"
  view_original_text = "view_original_text_1"

  storage_descriptor {
    bucket_columns            = ["bucket_column_1"]
    compressed                = false
    input_format              = "SequenceFileInputFormat"
    location                  = "my_location"
    number_of_buckets         = 1
    output_format             = "SequenceFileInputFormat"
    stored_as_sub_directories = false

    parameters = {
      param1 = "param1_val"
    }

    columns {
      name    = "my_column_1"
      type    = "int"
      comment = "my_column1_comment"
    }

    columns {
      name    = "my_column_2"
      type    = "string"
      comment = "my_column2_comment"
    }

    ser_de_info {
      name = "ser_de_name"

      parameters = {
        param1 = "param_val_1"
      }

      serialization_library = "org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe"
    }

    sort_columns {
      column     = "my_column_1"
      sort_order = 1
    }

    skewed_info {
      skewed_column_names = [
        "my_column_1",
      ]

      skewed_column_value_location_maps = {
        my_column_1 = "my_column_1_val_loc_map"
      }

      skewed_column_values = [
        "skewed_val_1",
      ]
    }
  }

  partition_keys {
    name    = "my_column_1"
    type    = "int"
    comment = "my_column_1_comment"
  }

  partition_keys {
    name    = "my_column_2"
    type    = "string"
    comment = "my_column_2_comment"
  }

  parameters = {
    param1 = "param1_val"
  }
}
`, rName, desc)
}

func testAccCatalogTableConfig_fullReplacedValues(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name               = %[1]q
  database_name      = aws_glue_catalog_database.test.name
  description        = "A test table from terraform2"
  owner              = "my_owner2"
  retention          = 2
  table_type         = "EXTERNAL_TABLE"
  view_expanded_text = "view_expanded_text_12"
  view_original_text = "view_original_text_12"

  storage_descriptor {
    bucket_columns = [
      "bucket_column_12",
      "bucket_column_2",
    ]

    compressed                = true
    input_format              = "TextInputFormat"
    location                  = "my_location2"
    number_of_buckets         = 12
    output_format             = "TextInputFormat"
    stored_as_sub_directories = true

    parameters = {
      param12 = "param1_val2"
    }

    columns {
      name    = "my_column_12"
      type    = "date"
      comment = "my_column1_comment2"
    }

    columns {
      name    = "my_column_22"
      type    = "timestamp"
      comment = "my_column2_comment2"
    }

    ser_de_info {
      name = "ser_de_name2"

      parameters = {
        param2 = "param_val_12"
      }

      serialization_library = "org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe2"
    }

    skewed_info {
      skewed_column_names = [
        "my_column_12",
      ]

      skewed_column_value_location_maps = {
        my_column_12 = "my_column_1_val_loc_map2"
      }

      skewed_column_values = [
        "skewed_val_12",
        "skewed_val_2",
      ]
    }

    sort_columns {
      column     = "my_column_12"
      sort_order = 0
    }
  }

  partition_keys {
    name    = "my_column_12"
    type    = "date"
    comment = "my_column_1_comment2"
  }

  partition_keys {
    name    = "my_column_22"
    type    = "timestamp"
    comment = "my_column_2_comment2"
  }

  parameters = {
    param2 = "param1_val2"
  }
}
`, rName)
}

func testAccCatalogTableConfig_storageDescriptorEmptyConfigurationBlock(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q

  storage_descriptor {}
}
`, rName)
}

func testAccCatalogTableConfig_storageDescriptorSerDeInfoEmptyConfigurationBlock(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q

  storage_descriptor {
    ser_de_info {}
  }
}
`, rName)
}

func testAccCatalogTableConfig_storageDescriptorSerDeInfo(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q

  storage_descriptor {
    ser_de_info {
      name = "ser_de_name"
    }
  }
}
`, rName)
}

func testAccCatalogTableConfig_storageDescriptorSerDeInfoUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q

  storage_descriptor {
    ser_de_info {
      parameters = {
        param1 = "param_val_1"
      }
      serialization_library = "org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe"
    }
  }
}
`, rName)
}

func testAccCatalogTableConfig_storageDescriptorSkewedInfoEmptyConfigurationBlock(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q

  storage_descriptor {
    skewed_info {
      skewed_column_names = []

      skewed_column_value_location_maps = {}
      skewed_column_values              = []
    }
  }
}
`, rName)
}

func testAccCatalogTableConfig_storageDescriptorSchemaReference(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_registry" "test" {
  registry_name = %[1]q
}

resource "aws_glue_schema" "test" {
  schema_name       = %[1]q
  registry_arn      = aws_glue_registry.test.arn
  data_format       = "AVRO"
  compatibility     = "NONE"
  schema_definition = "{\"type\": \"record\", \"name\": \"r1\", \"fields\": [ {\"name\": \"f1\", \"type\": \"int\"}, {\"name\": \"f2\", \"type\": \"string\"} ]}"
}

resource "aws_glue_catalog_table" "test" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q

  storage_descriptor {
    schema_reference {
      schema_id {
        schema_name   = aws_glue_schema.test.schema_name
        registry_name = aws_glue_schema.test.registry_name
      }

      schema_version_number = aws_glue_schema.test.latest_schema_version
    }
  }
}
`, rName)
}

func testAccCatalogTableConfig_storageDescriptorSchemaReferenceARN(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_registry" "test" {
  registry_name = %[1]q
}

resource "aws_glue_schema" "test" {
  schema_name       = %[1]q
  registry_arn      = aws_glue_registry.test.arn
  data_format       = "AVRO"
  compatibility     = "NONE"
  schema_definition = "{\"type\": \"record\", \"name\": \"r1\", \"fields\": [ {\"name\": \"f1\", \"type\": \"int\"}, {\"name\": \"f2\", \"type\": \"string\"} ]}"
}

resource "aws_glue_catalog_table" "test" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q

  storage_descriptor {
    schema_reference {
      schema_id {
        schema_arn = aws_glue_schema.test.arn
      }

      schema_version_number = aws_glue_schema.test.latest_schema_version
    }
  }
}
`, rName)
}

func testAccCatalogTableConfig_columnParameters(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name               = %[1]q
  database_name      = aws_glue_catalog_database.test.name
  owner              = "my_owner"
  retention          = 1
  table_type         = "VIRTUAL_VIEW"
  view_expanded_text = "view_expanded_text_1"
  view_original_text = "view_original_text_1"

  storage_descriptor {
    bucket_columns            = ["bucket_column_1"]
    compressed                = false
    input_format              = "SequenceFileInputFormat"
    location                  = "my_location"
    number_of_buckets         = 1
    output_format             = "SequenceFileInputFormat"
    stored_as_sub_directories = false

    parameters = {
      param1 = "param1_val"
    }

    columns {
      name    = "my_column_1"
      type    = "int"
      comment = "my_column1_comment"

      parameters = {
        param2 = "param2_val"
      }
    }

    ser_de_info {
      name = "ser_de_name"

      parameters = {
        param1 = "param_val_1"
      }

      serialization_library = "org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe"
    }

    sort_columns {
      column     = "my_column_1"
      sort_order = 1
    }

    skewed_info {
      skewed_column_names = [
        "my_column_1",
      ]

      skewed_column_value_location_maps = {
        my_column_1 = "my_column_1_val_loc_map"
      }

      skewed_column_values = [
        "skewed_val_1",
      ]
    }
  }

  partition_keys {
    name    = "my_column_1"
    type    = "int"
    comment = "my_column_1_comment"
  }

  parameters = {
    param1 = "param1_val"
  }
}
`, rName)
}

func testAccCheckTableDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_catalog_table" {
			continue
		}

		catalogId, dbName, resourceName, err := tfglue.ReadTableID(rs.Primary.ID)
		if err != nil {
			return err
		}

		if _, err := tfglue.FindTableByName(conn, catalogId, dbName, resourceName); err != nil {
			//Verify the error is what we want
			if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
				continue
			}

			return err
		}
		return fmt.Errorf("still exists")
	}
	return nil
}

func testAccCheckCatalogTableExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		catalogId, dbName, resourceName, err := tfglue.ReadTableID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn
		out, err := tfglue.FindTableByName(conn, catalogId, dbName, resourceName)
		if err != nil {
			return err
		}

		if out.Table == nil {
			return fmt.Errorf("No Glue Table Found")
		}

		if aws.StringValue(out.Table.Name) != resourceName {
			return fmt.Errorf("Glue Table Mismatch - existing: %q, state: %q",
				aws.StringValue(out.Table.Name), resourceName)
		}

		return nil
	}
}

func testAccCatalogTableConfig_partitionIndexesSingle(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name               = %[1]q
  database_name      = aws_glue_catalog_database.test.name
  owner              = "my_owner"
  retention          = 1
  table_type         = "VIRTUAL_VIEW"
  view_expanded_text = "view_expanded_text_1"
  view_original_text = "view_original_text_1"

  storage_descriptor {
    bucket_columns            = ["bucket_column_1"]
    compressed                = false
    input_format              = "SequenceFileInputFormat"
    location                  = "my_location"
    number_of_buckets         = 1
    output_format             = "SequenceFileInputFormat"
    stored_as_sub_directories = false

    parameters = {
      param1 = "param1_val"
    }

    columns {
      name    = "my_column_1"
      type    = "int"
      comment = "my_column1_comment"
    }

    columns {
      name    = "my_column_2"
      type    = "string"
      comment = "my_column2_comment"
    }

    ser_de_info {
      name = "ser_de_name"

      parameters = {
        param1 = "param_val_1"
      }

      serialization_library = "org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe"
    }

    sort_columns {
      column     = "my_column_1"
      sort_order = 1
    }

    skewed_info {
      skewed_column_names = [
        "my_column_1",
      ]

      skewed_column_value_location_maps = {
        my_column_1 = "my_column_1_val_loc_map"
      }

      skewed_column_values = [
        "skewed_val_1",
      ]
    }
  }

  partition_keys {
    name    = "my_column_1"
    type    = "int"
    comment = "my_column_1_comment"
  }

  partition_keys {
    name    = "my_column_2"
    type    = "string"
    comment = "my_column_2_comment"
  }

  parameters = {
    param1 = "param1_val"
  }

  partition_index {
    index_name = %[1]q
    keys       = ["my_column_1", "my_column_2"]
  }
}
`, rName)
}

func testAccCatalogTableConfig_partitionIndexesMultiple(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name               = %[1]q
  database_name      = aws_glue_catalog_database.test.name
  owner              = "my_owner"
  retention          = 1
  table_type         = "VIRTUAL_VIEW"
  view_expanded_text = "view_expanded_text_1"
  view_original_text = "view_original_text_1"

  storage_descriptor {
    bucket_columns            = ["bucket_column_1"]
    compressed                = false
    input_format              = "SequenceFileInputFormat"
    location                  = "my_location"
    number_of_buckets         = 1
    output_format             = "SequenceFileInputFormat"
    stored_as_sub_directories = false

    parameters = {
      param1 = "param1_val"
    }

    columns {
      name    = "my_column_1"
      type    = "int"
      comment = "my_column1_comment"
    }

    columns {
      name    = "my_column_2"
      type    = "string"
      comment = "my_column2_comment"
    }

    ser_de_info {
      name = "ser_de_name"

      parameters = {
        param1 = "param_val_1"
      }

      serialization_library = "org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe"
    }

    sort_columns {
      column     = "my_column_1"
      sort_order = 1
    }

    skewed_info {
      skewed_column_names = [
        "my_column_1",
      ]

      skewed_column_value_location_maps = {
        my_column_1 = "my_column_1_val_loc_map"
      }

      skewed_column_values = [
        "skewed_val_1",
      ]
    }
  }

  partition_keys {
    name    = "my_column_1"
    type    = "int"
    comment = "my_column_1_comment"
  }

  partition_keys {
    name    = "my_column_2"
    type    = "string"
    comment = "my_column_2_comment"
  }

  parameters = {
    param1 = "param1_val"
  }

  partition_index {
    index_name = %[1]q
    keys       = ["my_column_1", "my_column_2"]
  }

  partition_index {
    index_name = "%[1]s-2"
    keys       = ["my_column_1"]
  }
}
`, rName)
}

func testAccCatalogTableConfig_target(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  target_table {
    catalog_id    = aws_glue_catalog_table.test2.catalog_id
    database_name = aws_glue_catalog_table.test2.database_name
    name          = aws_glue_catalog_table.test2.name
  }
}

resource "aws_glue_catalog_database" "test2" {
  name = "%[1]s-2"
}

resource "aws_glue_catalog_table" "test2" {
  name          = "%[1]s-2"
  database_name = aws_glue_catalog_database.test2.name
}
`, rName)
}
