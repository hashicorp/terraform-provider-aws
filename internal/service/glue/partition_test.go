// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGluePartition_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	parValue := sdkacctest.RandString(10)
	resourceName := "aws_glue_partition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPartitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPartitionConfig_basic(rName, parValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPartitionExists(ctx, resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, rName),
					resource.TestCheckResourceAttr(resourceName, "partition_values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "partition_values.0", parValue),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
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

func TestAccGluePartition_multipleValues(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	parValue := sdkacctest.RandString(10)
	parValue2 := sdkacctest.RandString(11)
	resourceName := "aws_glue_partition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPartitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPartitionConfig_multiplePartValue(rName, parValue, parValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPartitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "partition_values.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "partition_values.0", parValue),
					resource.TestCheckResourceAttr(resourceName, "partition_values.1", parValue2),
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

func TestAccGluePartition_parameters(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	parValue := sdkacctest.RandString(10)
	resourceName := "aws_glue_partition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPartitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPartitionConfig_parameters1(rName, parValue, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPartitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.key1", acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPartitionConfig_parameters2(rName, parValue, acctest.CtKey1, "valueUpdated1", acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPartitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "parameters.key1", "valueUpdated1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.key2", acctest.CtValue2),
				),
			},
			{
				Config: testAccPartitionConfig_parameters1(rName, parValue, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPartitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.key2", acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccGluePartition_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	parValue := sdkacctest.RandString(10)
	resourceName := "aws_glue_partition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPartitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPartitionConfig_basic(rName, parValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPartitionExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfglue.ResourcePartition(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccGluePartition_Disappears_table(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	parValue := sdkacctest.RandString(10)
	resourceName := "aws_glue_partition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPartitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPartitionConfig_basic(rName, parValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPartitionExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfglue.ResourceCatalogTable(), "aws_glue_catalog_table.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPartitionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_glue_partition" {
				continue
			}

			if _, err := tfglue.FindPartitionByValues(ctx, conn, rs.Primary.ID); err != nil {
				if errs.IsA[*awstypes.EntityNotFoundException](err) {
					continue
				}

				return err
			}
			return fmt.Errorf("still exists")
		}
		return nil
	}
}

func testAccCheckPartitionExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)
		out, err := tfglue.FindPartitionByValues(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if out == nil {
			return fmt.Errorf("No Glue Partition Found")
		}

		return nil
	}
}

func testAccPartitionConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

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
    name    = "my_column_12"
    type    = "date"
    comment = "my_column_1_comment2"
  }
}
`, rName)
}

func testAccPartitionConfig_basic(rName, parValue string) string {
	return testAccPartitionConfigBase(rName) +
		fmt.Sprintf(`
resource "aws_glue_partition" "test" {
  database_name    = aws_glue_catalog_database.test.name
  table_name       = aws_glue_catalog_table.test.name
  partition_values = ["%[1]s"]
}
`, parValue)
}

func testAccPartitionConfig_parameters1(rName, parValue, key1, value1 string) string {
	return testAccPartitionConfigBase(rName) +
		fmt.Sprintf(`
resource "aws_glue_partition" "test" {
  database_name    = aws_glue_catalog_database.test.name
  table_name       = aws_glue_catalog_table.test.name
  partition_values = ["%[1]s"]

  parameters = {
    %[2]q = %[3]q
  }
}
`, parValue, key1, value1)
}

func testAccPartitionConfig_parameters2(rName, parValue, key1, value1, key2, value2 string) string {
	return testAccPartitionConfigBase(rName) +
		fmt.Sprintf(`
resource "aws_glue_partition" "test" {
  database_name    = aws_glue_catalog_database.test.name
  table_name       = aws_glue_catalog_table.test.name
  partition_values = ["%[1]s"]

  parameters = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, parValue, key1, value1, key2, value2)
}

func testAccPartitionConfig_multiplePartValue(rName, parValue, parValue2 string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

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
    name    = "my_column_12"
    type    = "date"
    comment = "my_column_1_comment2"
  }

  partition_keys {
    name    = "my_column_11"
    type    = "date"
    comment = "my_column_1_comment2"
  }
}

resource "aws_glue_partition" "test" {
  database_name    = aws_glue_catalog_database.test.name
  table_name       = aws_glue_catalog_table.test.name
  partition_values = ["%[2]s", "%[3]s"]
}
`, rName, parValue, parValue2)
}
