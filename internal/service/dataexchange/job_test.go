package dataexchange_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/dataexchange"
	"github.com/aws/aws-sdk-go-v2/service/dataexchange/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfdataexchange "github.com/hashicorp/terraform-provider-aws/internal/service/dataexchange"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataExchangeJob_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var job dataexchange.GetJobOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dataexchange_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_importFromS3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "type", string(types.TypeImportAssetsFromS3)),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrPair(resourceName, "details.import_assets_from_s3.data_set_id", "aws_dataexchange_data_set.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "details.import_assets_from_s3.revision_id", "aws_dataexchange_revision.test", "id"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "dataexchange", regexache.MustCompile(`jobs/.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"state",
					"updated_at",
				},
			},
		},
	})
}

func TestAccDataExchangeJob_exportToS3(t *testing.T) {
	// Reuse same setup as basic test
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var job dataexchange.GetJobOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dataexchange_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { // Same as basic
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx), // Reuse same destroy check
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_exportToS3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job), // Reuse same existence check
					// Export-specific checks
					resource.TestCheckResourceAttr(resourceName, "type", string(types.TypeExportAssetsToS3)),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrPair(resourceName, "details.export_assets_to_s3.data_set_id", "aws_dataexchange_data_set.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "details.export_assets_to_s3.revision_id", "aws_dataexchange_revision.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "details.export_assets_to_s3.asset_destinations.#", "1"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "dataexchange", regexache.MustCompile(`jobs/.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"state",
					"updated_at",
				},
			},
		},
	})
}

func TestAccDataExchangeJob_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var job dataexchange.GetJobOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dataexchange_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_importFromS3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfdataexchange.ResourceJob, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckJobDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataExchangeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dataexchange_job" {
				continue
			}

			job, err := tfdataexchange.FindJobById(ctx, conn, rs.Primary.ID)

			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}

			// Jobs can't be deleted once completed, so we consider terminal states as "destroyed"
			if err == nil {
				switch job.State {
				case types.StateCompleted, types.StateError, types.StateCancelled:
					return nil
				}
				return create.Error(names.DataExchange, create.ErrActionCheckingDestroyed, tfdataexchange.ResNameJob, rs.Primary.ID, errors.New("job still exists and is not in a terminal state"))
			}

			return create.Error(names.DataExchange, create.ErrActionCheckingDestroyed, tfdataexchange.ResNameJob, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckJobExists(ctx context.Context, name string, job *dataexchange.GetJobOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DataExchange, create.ErrActionCheckingExistence, tfdataexchange.ResNameJob, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.DataExchange, create.ErrActionCheckingExistence, tfdataexchange.ResNameJob, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataExchangeClient(ctx)
		resp, err := tfdataexchange.FindJobById(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.DataExchange, create.ErrActionCheckingExistence, tfdataexchange.ResNameJob, rs.Primary.ID, err)
		}

		*job = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataExchangeClient(ctx)

	input := &dataexchange.ListJobsInput{}

	_, err := conn.ListJobs(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccJobConfig_importFromS3(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
    bucket = %[1]q
}

resource "aws_s3_object" "test" {
    bucket = aws_s3_bucket.test.id
    key    = "test-key"
    source = "test-fixtures/testfile.txt"
}

resource "aws_dataexchange_data_set" "test" {
    asset_type  = "S3_SNAPSHOT"
    description = "Test Data Set"
    name        = %[1]q
}

resource "aws_dataexchange_revision" "test" {
    data_set_id = aws_dataexchange_data_set.test.id
    comment     = "test revision"
}

resource "aws_dataexchange_job" "test" {
    type = "IMPORT_ASSETS_FROM_S3"
    details = {
        import_assets_from_s3 = {
            data_set_id = aws_dataexchange_data_set.test.id
            revision_id = aws_dataexchange_revision.test.id
            asset_sources = [{
                bucket = aws_s3_bucket.test.id
                key    = aws_s3_object.test.key
            }]
        }
        export_assets_to_s3 = null
    }
}
`, rName)
}

func testAccJobConfig_exportToS3(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
    bucket = %[1]q
}

resource "aws_dataexchange_data_set" "test" {
    asset_type  = "S3_SNAPSHOT"
    description = "Test Data Set"
    name        = %[1]q
}

resource "aws_dataexchange_revision" "test" {
    data_set_id = aws_dataexchange_data_set.test.id
    comment     = "test revision"
}

locals {
    # Create a valid alphanumeric asset ID
    asset_id = replace(substr(%[1]q, 0, 32), "-", "")
    # Get just the revision ID part after the colon
    revision_id = split(":", aws_dataexchange_revision.test.id)[1]
}

resource "aws_dataexchange_job" "test" {
    type = "EXPORT_ASSETS_TO_S3"
    details = {
        export_assets_to_s3 = {
            data_set_id = aws_dataexchange_data_set.test.id
            revision_id = local.revision_id
            asset_destinations = [{
                asset_id = local.asset_id
                bucket  = aws_s3_bucket.test.id
                key     = "exported/test-key"
            }]
        }
        import_assets_from_s3 = null
    }

    depends_on = [aws_dataexchange_revision.test]
}
`, rName)
}
