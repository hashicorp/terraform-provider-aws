// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQuickSightIngestion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var ingestion quicksight.Ingestion
	dataSetName := "aws_quicksight_data_set.test"
	resourceName := "aws_quicksight_ingestion.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngestionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionConfig_basic(rId, rName, quicksight.IngestionTypeFullRefresh),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIngestionExists(ctx, resourceName, &ingestion),
					resource.TestCheckResourceAttr(resourceName, "ingestion_id", rId),
					resource.TestCheckResourceAttr(resourceName, "ingestion_type", quicksight.IngestionTypeFullRefresh),
					resource.TestCheckResourceAttrPair(resourceName, "data_set_id", dataSetName, "data_set_id"),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "quicksight", fmt.Sprintf("dataset/%[1]s/ingestion/%[1]s", rId)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"ingestion_type",
				},
			},
		},
	})
}

// NOTE: There is no base _disappears test for this resource. Ingestions
// persist for the life of the parent data set, even if cancelled, so
// disappearance of this upstream resource is tested instead.
func TestAccQuickSightIngestion_disappears_dataSet(t *testing.T) {
	ctx := acctest.Context(t)
	var ingestion quicksight.Ingestion
	dataSetName := "aws_quicksight_data_set.test"
	resourceName := "aws_quicksight_ingestion.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngestionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionConfig_basic(rId, rName, quicksight.IngestionTypeFullRefresh),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIngestionExists(ctx, resourceName, &ingestion),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfquicksight.ResourceDataSet(), dataSetName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckIngestionExists(ctx context.Context, resourceName string, ingestion *quicksight.Ingestion) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn(ctx)
		output, err := tfquicksight.FindIngestionByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.QuickSight, create.ErrActionCheckingExistence, tfquicksight.ResNameIngestion, rs.Primary.ID, err)
		}

		*ingestion = *output

		return nil
	}
}

func testAccCheckIngestionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_ingestion" {
				continue
			}

			output, err := tfquicksight.FindIngestionByID(ctx, conn, rs.Primary.ID)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
					return nil
				}
				return err
			}

			if output != nil && !isDestroyedStatus(aws.StringValue(output.IngestionStatus)) {
				return create.Error(names.QuickSight, create.ErrActionCheckingDestroyed, tfquicksight.ResNameIngestion, rs.Primary.ID, err)
			}
		}

		return nil
	}
}

func isDestroyedStatus(status string) bool {
	targetStatuses := []string{
		quicksight.IngestionStatusCancelled,
		quicksight.IngestionStatusCompleted,
		quicksight.IngestionStatusFailed,
	}
	for _, target := range targetStatuses {
		if status == target {
			return true
		}
	}
	return false
}

func testAccIngestionConfigBase(rId, rName string) string {
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

func testAccIngestionConfig_basic(rId, rName, ingestionType string) string {
	return acctest.ConfigCompose(
		testAccIngestionConfigBase(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_ingestion" "test" {
  data_set_id    = aws_quicksight_data_set.test.data_set_id
  ingestion_id   = %[1]q
  ingestion_type = %[3]q
}
`, rId, rName, ingestionType))
}
