// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bcmdataexports_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bcmdataexports"
	"github.com/aws/aws-sdk-go-v2/service/bcmdataexports/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfbcmdataexports "github.com/hashicorp/terraform-provider-aws/internal/service/bcmdataexports"
)

func TestAccBCMDataExportsExport_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var export bcmdataexports.DescribeExportResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bcmdataexports_export.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BCMDataExportsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BCMDataExportsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExportDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExportConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExportExists(ctx, resourceName, &export),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "bcmdataexports", regexache.MustCompile(`export:+.`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccBCMDataExportsExport_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var export bcmdataexports.DescribeExportResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bcmdataexports_export.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BCMDataExportsEndpointID)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BCMDataExportsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExportDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExportConfig_basic(rName, testAccExportVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExportExists(ctx, resourceName, &export),
					// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
					// but expects a new resource factory function as the third argument. To expose this
					// private function to the testing package, you may need to add a line like the following
					// to exports_test.go:
					//
					//   var ResourceExport = newResourceExport
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbcmdataexports.ResourceExport, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckExportDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BCMDataExportsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bcmdataexports_export" {
				continue
			}

			input := &bcmdataexports.DescribeExportInput{
				ExportId: aws.String(rs.Primary.ID),
			}
			_, err := conn.DescribeExport(ctx, &bcmdataexports.DescribeExportInput{
				ExportId: aws.String(rs.Primary.ID),
			})
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.BCMDataExports, create.ErrActionCheckingDestroyed, tfbcmdataexports.ResNameExport, rs.Primary.ID, err)
			}

			return create.Error(names.BCMDataExports, create.ErrActionCheckingDestroyed, tfbcmdataexports.ResNameExport, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckExportExists(ctx context.Context, name string, export *bcmdataexports.DescribeExportResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.BCMDataExports, create.ErrActionCheckingExistence, tfbcmdataexports.ResNameExport, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.BCMDataExports, create.ErrActionCheckingExistence, tfbcmdataexports.ResNameExport, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BCMDataExportsClient(ctx)
		resp, err := conn.DescribeExport(ctx, &bcmdataexports.DescribeExportInput{
			ExportId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.BCMDataExports, create.ErrActionCheckingExistence, tfbcmdataexports.ResNameExport, rs.Primary.ID, err)
		}

		*export = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BCMDataExportsClient(ctx)

	input := &bcmdataexports.ListExportsInput{}
	_, err := conn.ListExports(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckExportNotRecreated(before, after *bcmdataexports.DescribeExportResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.ExportId), aws.ToString(after.ExportId); before != after {
			return create.Error(names.BCMDataExports, create.ErrActionCheckingNotRecreated, tfbcmdataexports.ResNameExport, aws.ToString(before.ExportId), errors.New("recreated"))
		}

		return nil
	}
}

func testAccExportConfig_basic(rName, version string) string {
	return fmt.Sprintf(`

resource "aws_s3_bucket" "test" {
  bucket = "testing-bucket"
}

resource "aws_bcmdataexports_export" "test" {
  export {
    name = "exampleexportname"
    data_query {
      query_statement = "SELECT identity_line_item_id, identity_time_interval, line_item_product_code,line_item_unblended_cost FROM COST_AND_USAGE_REPORT"
      table_configurations {
        "COST_AND_USAGE_REPORT" {
          "TIME_GRANULARITY" = "DAILY"
        }
      }
    }

    "destination_configurations" {
      "s3_destination" {
          "s3_bucket" = aws_s3_bucket.test.bucket
          "s3_prefix" = aws_s3_bucket.test.bucket_prefix
          "s3_region" = aws_s3_bucket.test.region
          "s3_output_configurations" {
            "overwrite" = "OVERWRITE_REPORT"
            "format" = "TEXT_OR_CSV"
            "compression" = "GZIP"
            "output_type" = "CUSTOM"
          }
        }
    }

    "refresh_cadence" {
      "frequency" = "SYNCHRONOUS"
    }
  }
}
`, rName, version)
}
