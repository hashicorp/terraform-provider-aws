// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
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

func TestAccGlueMlTransform_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var transform glue.GetMLTransformOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_ml_transform.test"
	roleResourceName := "aws_iam_role.test"
	tableResourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMLTransformDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMLTransformConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMLTransformExists(ctx, resourceName, &transform),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", regexache.MustCompile(`mlTransform/tfm-.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, roleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "input_record_tables.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "input_record_tables.0.database_name", tableResourceName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName, "input_record_tables.0.table_name", tableResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.transform_type", "FIND_MATCHES"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.primary_key_column_name", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.accuracy_cost_trade_off", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.precision_recall_trade_off", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.enforce_provided_labels", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrTimeout, "2880"),
					resource.TestCheckResourceAttr(resourceName, "schema.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "schema.0.data_type", "int"),
					resource.TestCheckResourceAttr(resourceName, "schema.0.name", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "schema.1.data_type", "string"),
					resource.TestCheckResourceAttr(resourceName, "schema.1.name", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "label_count", acctest.Ct0),
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

func TestAccGlueMlTransform_typeFindMatchesFull(t *testing.T) {
	ctx := acctest.Context(t)
	var transform glue.GetMLTransformOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_ml_transform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMLTransformDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMLTransformConfig_typeFindMatchesFull(rName, true, 0.5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMLTransformExists(ctx, resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, "parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.transform_type", "FIND_MATCHES"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.primary_key_column_name", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.accuracy_cost_trade_off", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.precision_recall_trade_off", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.enforce_provided_labels", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMLTransformConfig_typeFindMatchesFull(rName, false, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMLTransformExists(ctx, resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, "parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.transform_type", "FIND_MATCHES"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.primary_key_column_name", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.accuracy_cost_trade_off", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.precision_recall_trade_off", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.enforce_provided_labels", acctest.CtFalse),
				),
			},
			{
				Config: testAccMLTransformConfig_typeFindMatchesFull(rName, true, 0.5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMLTransformExists(ctx, resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, "parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.transform_type", "FIND_MATCHES"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.primary_key_column_name", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.accuracy_cost_trade_off", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.precision_recall_trade_off", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.enforce_provided_labels", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccGlueMlTransform_description(t *testing.T) {
	ctx := acctest.Context(t)
	var transform glue.GetMLTransformOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_ml_transform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMLTransformDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMLTransformConfig_description(rName, "First Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMLTransformExists(ctx, resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "First Description"),
				),
			},
			{
				Config: testAccMLTransformConfig_description(rName, "Second Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMLTransformExists(ctx, resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Second Description"),
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

func TestAccGlueMlTransform_glueVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var transform glue.GetMLTransformOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_ml_transform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMLTransformDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMLTransformConfig_version(rName, "0.9"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMLTransformExists(ctx, resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, "glue_version", "0.9"),
				),
			},
			{
				Config: testAccMLTransformConfig_version(rName, "1.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMLTransformExists(ctx, resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, "glue_version", "1.0"),
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

func TestAccGlueMlTransform_maxRetries(t *testing.T) {
	ctx := acctest.Context(t)
	var transform glue.GetMLTransformOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_ml_transform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMLTransformDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccMLTransformConfig_maxRetries(rName, 11),
				ExpectError: regexache.MustCompile(`expected max_retries to be in the range`),
			},
			{
				Config: testAccMLTransformConfig_maxRetries(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMLTransformExists(ctx, resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, "max_retries", acctest.Ct0),
				),
			},
			{
				Config: testAccMLTransformConfig_maxRetries(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMLTransformExists(ctx, resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, "max_retries", acctest.Ct10),
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

func TestAccGlueMlTransform_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var transform1, transform2, transform3 glue.GetMLTransformOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_ml_transform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMLTransformDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMLTransformConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMLTransformExists(ctx, resourceName, &transform1),
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
				Config: testAccMLTransformConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMLTransformExists(ctx, resourceName, &transform2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccMLTransformConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMLTransformExists(ctx, resourceName, &transform3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccGlueMlTransform_timeout(t *testing.T) {
	ctx := acctest.Context(t)
	var transform glue.GetMLTransformOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_ml_transform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMLTransformDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMLTransformConfig_timeout(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMLTransformExists(ctx, resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, names.AttrTimeout, acctest.Ct1),
				),
			},
			{
				Config: testAccMLTransformConfig_timeout(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMLTransformExists(ctx, resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, names.AttrTimeout, acctest.Ct2),
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

func TestAccGlueMlTransform_workerType(t *testing.T) {
	ctx := acctest.Context(t)
	var transform glue.GetMLTransformOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_ml_transform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMLTransformDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMLTransformConfig_workerType(rName, "Standard", 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMLTransformExists(ctx, resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, "worker_type", "Standard"),
					resource.TestCheckResourceAttr(resourceName, "number_of_workers", acctest.Ct1),
				),
			},
			{
				Config: testAccMLTransformConfig_workerType(rName, "G.1X", 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMLTransformExists(ctx, resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, "worker_type", "G.1X"),
					resource.TestCheckResourceAttr(resourceName, "number_of_workers", acctest.Ct2),
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

func TestAccGlueMlTransform_maxCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var transform glue.GetMLTransformOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_ml_transform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMLTransformDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMLTransformConfig_maxCapacity(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMLTransformExists(ctx, resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, names.AttrMaxCapacity, acctest.Ct10),
				),
			},
			{
				Config: testAccMLTransformConfig_maxCapacity(rName, 15),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMLTransformExists(ctx, resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, names.AttrMaxCapacity, "15"),
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

func TestAccGlueMlTransform_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var transform glue.GetMLTransformOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_ml_transform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMLTransformDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMLTransformConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMLTransformExists(ctx, resourceName, &transform),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfglue.ResourceMLTransform(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckMLTransformExists(ctx context.Context, resourceName string, mlTransform *glue.GetMLTransformOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Glue Job ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

		output, err := conn.GetMLTransform(ctx, &glue.GetMLTransformInput{
			TransformId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Glue ML Transform (%s) not found", rs.Primary.ID)
		}

		if aws.ToString(output.TransformId) == rs.Primary.ID {
			*mlTransform = *output
			return nil
		}

		return fmt.Errorf("Glue ML Transform (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckMLTransformDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_glue_ml_transform" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

			output, err := conn.GetMLTransform(ctx, &glue.GetMLTransformInput{
				TransformId: aws.String(rs.Primary.ID),
			})

			if err != nil {
				if errs.IsA[*awstypes.EntityNotFoundException](err) {
					return nil
				}
			}

			if output != nil && aws.ToString(output.TransformId) == rs.Primary.ID {
				return fmt.Errorf("Glue ML Transform %s still exists", rs.Primary.ID)
			}

			return err
		}

		return nil
	}
}

func testAccMLTransformBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy" "AWSGlueServiceRole" {
  arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSGlueServiceRole"
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "glue.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = data.aws_iam_policy.AWSGlueServiceRole.arn
  role       = aws_iam_role.test.name
}

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
`, rName)
}

func testAccMLTransformConfig_basic(rName string) string {
	return testAccMLTransformBaseConfig(rName) + fmt.Sprintf(`
resource "aws_glue_ml_transform" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  input_record_tables {
    database_name = aws_glue_catalog_table.test.database_name
    table_name    = aws_glue_catalog_table.test.name
  }

  parameters {
    transform_type = "FIND_MATCHES"

    find_matches_parameters {
      primary_key_column_name = "my_column_1"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccMLTransformConfig_typeFindMatchesFull(rName string, enforce bool, tradeOff float64) string {
	return testAccMLTransformBaseConfig(rName) + fmt.Sprintf(`
resource "aws_glue_ml_transform" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  input_record_tables {
    database_name = aws_glue_catalog_table.test.database_name
    table_name    = aws_glue_catalog_table.test.name
  }

  parameters {
    transform_type = "FIND_MATCHES"

    find_matches_parameters {
      primary_key_column_name    = "my_column_1"
      enforce_provided_labels    = %[2]t
      precision_recall_trade_off = %[3]g
      accuracy_cost_trade_off    = %[3]g
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, enforce, tradeOff)
}

func testAccMLTransformConfig_description(rName, description string) string {
	return testAccMLTransformBaseConfig(rName) + fmt.Sprintf(`
resource "aws_glue_ml_transform" "test" {
  name        = %[1]q
  description = %[2]q
  role_arn    = aws_iam_role.test.arn

  input_record_tables {
    database_name = aws_glue_catalog_table.test.database_name
    table_name    = aws_glue_catalog_table.test.name
  }

  parameters {
    transform_type = "FIND_MATCHES"

    find_matches_parameters {
      primary_key_column_name = "my_column_1"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, description)
}

func testAccMLTransformConfig_version(rName, glueVersion string) string {
	return testAccMLTransformBaseConfig(rName) + fmt.Sprintf(`
resource "aws_glue_ml_transform" "test" {
  name         = %[1]q
  glue_version = %[2]q
  role_arn     = aws_iam_role.test.arn

  input_record_tables {
    database_name = aws_glue_catalog_table.test.database_name
    table_name    = aws_glue_catalog_table.test.name
  }

  parameters {
    transform_type = "FIND_MATCHES"

    find_matches_parameters {
      primary_key_column_name = "my_column_1"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, glueVersion)
}

func testAccMLTransformConfig_maxRetries(rName string, maxRetries int) string {
	return testAccMLTransformBaseConfig(rName) + fmt.Sprintf(`
resource "aws_glue_ml_transform" "test" {
  name        = %[1]q
  max_retries = %[2]d
  role_arn    = aws_iam_role.test.arn

  input_record_tables {
    database_name = aws_glue_catalog_table.test.database_name
    table_name    = aws_glue_catalog_table.test.name
  }

  parameters {
    transform_type = "FIND_MATCHES"

    find_matches_parameters {
      primary_key_column_name = "my_column_1"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, maxRetries)
}

func testAccMLTransformConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return testAccMLTransformBaseConfig(rName) + fmt.Sprintf(`
resource "aws_glue_ml_transform" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  input_record_tables {
    database_name = aws_glue_catalog_table.test.database_name
    table_name    = aws_glue_catalog_table.test.name
  }

  parameters {
    transform_type = "FIND_MATCHES"

    find_matches_parameters {
      primary_key_column_name = "my_column_1"
    }
  }

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, tagKey1, tagValue1)
}

func testAccMLTransformConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccMLTransformBaseConfig(rName) + fmt.Sprintf(`
resource "aws_glue_ml_transform" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  input_record_tables {
    database_name = aws_glue_catalog_table.test.database_name
    table_name    = aws_glue_catalog_table.test.name
  }

  parameters {
    transform_type = "FIND_MATCHES"

    find_matches_parameters {
      primary_key_column_name = "my_column_1"
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccMLTransformConfig_timeout(rName string, timeout int) string {
	return testAccMLTransformBaseConfig(rName) + fmt.Sprintf(`
resource "aws_glue_ml_transform" "test" {
  name     = %[1]q
  timeout  = %[2]d
  role_arn = aws_iam_role.test.arn

  input_record_tables {
    database_name = aws_glue_catalog_table.test.database_name
    table_name    = aws_glue_catalog_table.test.name
  }

  parameters {
    transform_type = "FIND_MATCHES"

    find_matches_parameters {
      primary_key_column_name = "my_column_1"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, timeout)
}

func testAccMLTransformConfig_workerType(rName, workerType string, numOfWorkers int) string {
	return testAccMLTransformBaseConfig(rName) + fmt.Sprintf(`
resource "aws_glue_ml_transform" "test" {
  name              = %[1]q
  worker_type       = %[2]q
  number_of_workers = %[3]d
  role_arn          = aws_iam_role.test.arn

  input_record_tables {
    database_name = aws_glue_catalog_table.test.database_name
    table_name    = aws_glue_catalog_table.test.name
  }

  parameters {
    transform_type = "FIND_MATCHES"

    find_matches_parameters {
      primary_key_column_name = "my_column_1"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, workerType, numOfWorkers)
}

func testAccMLTransformConfig_maxCapacity(rName string, maxCapacity float64) string {
	return testAccMLTransformBaseConfig(rName) + fmt.Sprintf(`
resource "aws_glue_ml_transform" "test" {
  name         = %[1]q
  max_capacity = %[2]g
  role_arn     = aws_iam_role.test.arn

  input_record_tables {
    database_name = aws_glue_catalog_table.test.database_name
    table_name    = aws_glue_catalog_table.test.name
  }

  parameters {
    transform_type = "FIND_MATCHES"

    find_matches_parameters {
      primary_key_column_name = "my_column_1"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, maxCapacity)
}
