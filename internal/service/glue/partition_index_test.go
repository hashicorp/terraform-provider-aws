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

func TestAccGluePartitionIndex_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_partition_index.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPartitionIndexDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccPartitionIndexConfig_basic(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPartitionIndexExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTableName, "aws_glue_catalog_table.test", names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDatabaseName, "aws_glue_catalog_database.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "partition_index.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "partition_index.0.index_name", rName),
					resource.TestCheckResourceAttr(resourceName, "partition_index.0.keys.#", acctest.Ct2),
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

func TestAccGluePartitionIndex_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_partition_index.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPartitionIndexDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccPartitionIndexConfig_basic(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPartitionIndexExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfglue.ResourcePartitionIndex(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfglue.ResourcePartitionIndex(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccGluePartitionIndex_Disappears_table(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_partition_index.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPartitionIndexDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccPartitionIndexConfig_basic(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPartitionIndexExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfglue.ResourceCatalogTable(), "aws_glue_catalog_table.test"),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfglue.ResourcePartitionIndex(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccGluePartitionIndex_Disappears_database(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_partition_index.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPartitionIndexDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccPartitionIndexConfig_basic(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPartitionIndexExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfglue.ResourceCatalogDatabase(), "aws_glue_catalog_database.test"),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfglue.ResourcePartitionIndex(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccPartitionIndexConfig_basic(rName string) string {
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
}

resource "aws_glue_partition_index" "test" {
  database_name = aws_glue_catalog_database.test.name
  table_name    = aws_glue_catalog_table.test.name

  partition_index {
    index_name = %[1]q
    keys       = ["my_column_1", "my_column_2"]
  }
}
`, rName)
}

func testAccCheckPartitionIndexDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_glue_partition_index" {
				continue
			}

			if _, err := tfglue.FindPartitionIndexByName(ctx, conn, rs.Primary.ID); err != nil {
				//Verify the error is what we want
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

func testAccCheckPartitionIndexExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)
		out, err := tfglue.FindPartitionIndexByName(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if out == nil {
			return fmt.Errorf("No Glue Partition Index Found")
		}

		return nil
	}
}
