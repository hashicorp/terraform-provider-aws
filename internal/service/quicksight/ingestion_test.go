// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"slices"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQuickSightIngestion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var ingestion awstypes.Ingestion
	dataSetName := "aws_quicksight_data_set.test"
	resourceName := "aws_quicksight_ingestion.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngestionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionConfig_basic(rId, rName, string(awstypes.IngestionTypeFullRefresh)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIngestionExists(ctx, t, resourceName, &ingestion),
					resource.TestCheckResourceAttr(resourceName, "ingestion_id", rId),
					resource.TestCheckResourceAttr(resourceName, "ingestion_type", string(awstypes.IngestionTypeFullRefresh)),
					resource.TestCheckResourceAttrPair(resourceName, "data_set_id", dataSetName, "data_set_id"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "quicksight", fmt.Sprintf("dataset/%[1]s/ingestion/%[1]s", rId)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"ingestion_status",
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
	var ingestion awstypes.Ingestion
	dataSetName := "aws_quicksight_data_set.test"
	resourceName := "aws_quicksight_ingestion.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngestionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionConfig_basic(rId, rName, string(awstypes.IngestionTypeFullRefresh)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIngestionExists(ctx, t, resourceName, &ingestion),
					acctest.CheckSDKResourceDisappears(ctx, t, tfquicksight.ResourceDataSet(), dataSetName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckIngestionExists(ctx context.Context, t *testing.T, n string, v *awstypes.Ingestion) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		output, err := tfquicksight.FindIngestionByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes["data_set_id"], rs.Primary.Attributes["ingestion_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckIngestionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_ingestion" {
				continue
			}

			output, err := tfquicksight.FindIngestionByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes["data_set_id"], rs.Primary.Attributes["ingestion_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			if isDestroyedStatus(output.IngestionStatus) {
				continue
			}

			return fmt.Errorf("QuickSight Ingestion (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func isDestroyedStatus(status awstypes.IngestionStatus) bool {
	return slices.Contains([]awstypes.IngestionStatus{
		awstypes.IngestionStatusCancelled,
		awstypes.IngestionStatusCompleted,
		awstypes.IngestionStatusFailed,
	}, status)
}

func testAccIngestionConfig_base(rId, rName string) string {
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
		testAccIngestionConfig_base(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_ingestion" "test" {
  data_set_id    = aws_quicksight_data_set.test.data_set_id
  ingestion_id   = %[1]q
  ingestion_type = %[3]q
}
`, rId, rName, ingestionType))
}
