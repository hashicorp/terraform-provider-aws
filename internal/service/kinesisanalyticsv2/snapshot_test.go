// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kinesisanalyticsv2_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfkinesisanalyticsv2 "github.com/hashicorp/terraform-provider-aws/internal/service/kinesisanalyticsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKinesisAnalyticsV2Snapshot_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SnapshotInfoDetails
	resourceName := "aws_kinesisanalyticsv2_snapshot.test"
	applicationResourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSnapshotExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "application_name", applicationResourceName, names.AttrName),
					resource.TestCheckResourceAttrSet(resourceName, "application_version_id"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "kinesisanalytics", "snapshot/"+rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_20"),
					resource.TestCheckResourceAttrSet(resourceName, "size_in_bytes"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccKinesisAnalyticsV2Snapshot_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SnapshotInfoDetails
	resourceName := "aws_kinesisanalyticsv2_snapshot.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSnapshotExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfkinesisanalyticsv2.ResourceSnapshot, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccKinesisAnalyticsV2Snapshot_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SnapshotInfoDetails
	resourceName := "aws_kinesisanalyticsv2_snapshot.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSnapshotExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccSnapshotConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1Updated),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSnapshotExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
				),
			},
		},
	})
}

// TestAccKinesisAnalyticsV2Snapshot_restoreToApplication starts a second application from another application's snapshot.
func TestAccKinesisAnalyticsV2Snapshot_restoreToApplication(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SnapshotInfoDetails
	var application awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_snapshot.test"
	targetResourceName := "aws_kinesisanalyticsv2_application.target"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_restoreToApplication(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSnapshotExists(ctx, t, resourceName, &v),
					testAccCheckApplicationExists(ctx, t, targetResourceName, &application),
					resource.TestCheckResourceAttr(targetResourceName, names.AttrStatus, "RUNNING"),
					resource.TestCheckResourceAttr(targetResourceName, "application_configuration.0.application_snapshot_configuration.0.snapshot_mode", "ENHANCED_SNAPSHOTS"),
					resource.TestCheckResourceAttr(targetResourceName, "application_configuration.0.run_configuration.0.application_restore_configuration.0.application_restore_type", "RESTORE_FROM_CUSTOM_SNAPSHOT"),
					resource.TestCheckResourceAttrPair(targetResourceName, "application_configuration.0.run_configuration.0.application_restore_configuration.0.snapshot_arn", resourceName, names.AttrARN),
				),
			},
		},
	})
}

func testAccCheckSnapshotDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).KinesisAnalyticsV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kinesisanalyticsv2_snapshot" && rs.Type != "aws_kinesisanalyticsv2_snapshot_import" {
				continue
			}

			_, err := tfkinesisanalyticsv2.FindEnhancedSnapshotByName(ctx, conn, rs.Primary.Attributes[names.AttrName])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Kinesis Analytics v2 Snapshot %s still exists", rs.Primary.Attributes[names.AttrName])
		}

		return nil
	}
}

func testAccCheckSnapshotExists(ctx context.Context, t *testing.T, n string, v *awstypes.SnapshotInfoDetails) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).KinesisAnalyticsV2Client(ctx)

		output, err := tfkinesisanalyticsv2.FindEnhancedSnapshotByName(ctx, conn, rs.Primary.Attributes[names.AttrName])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

// testAccSnapshotConfig_enhancedFlinkApplication returns a running Apache Flink application in enhanced snapshot mode.
// restoreConfiguration is the body of its application_restore_configuration block.
func testAccSnapshotConfig_enhancedFlinkApplication(resourceName, applicationName, restoreConfiguration string) string {
	return fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" %[1]q {
  name                   = %[2]q
  runtime_environment    = "FLINK-1_20"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn     = aws_s3_bucket.test.arn
          file_key       = aws_s3_object.test[0].key
          object_version = aws_s3_object.test[0].version_id
        }
      }

      code_content_type = "ZIPFILE"
    }

    application_snapshot_configuration {
      snapshots_enabled = true
      snapshot_mode     = "ENHANCED_SNAPSHOTS"
    }

    environment_properties {
      property_group {
        property_group_id = "InputStream0"

        property_map = {
          "aws.region"  = data.aws_region.current.region
          "stream.name" = aws_kinesis_stream.input.name
        }
      }

      property_group {
        property_group_id = "OutputStream0"

        property_map = {
          "aws.region"  = data.aws_region.current.region
          "stream.name" = aws_kinesis_stream.output.name
        }
      }
    }

    run_configuration {
      application_restore_configuration {
        %[3]s
      }
    }
  }

  start_application = true
}
`, resourceName, applicationName, restoreConfiguration)
}

func testAccSnapshotConfig_base(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		testAccApplicationConfig_baseFlinkApplication(rName),
		testAccSnapshotConfig_enhancedFlinkApplication("test", rName, `application_restore_type = "SKIP_RESTORE_FROM_SNAPSHOT"`),
	)
}

func testAccSnapshotConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccSnapshotConfig_base(rName), fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_snapshot" "test" {
  application_name = aws_kinesisanalyticsv2_application.test.name
  name             = %[1]q
}
`, rName))
}

func testAccSnapshotConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccSnapshotConfig_base(rName), fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_snapshot" "test" {
  application_name = aws_kinesisanalyticsv2_application.test.name
  name             = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccSnapshotConfig_restoreToApplication(rName string) string {
	return acctest.ConfigCompose(
		testAccSnapshotConfig_basic(rName),
		testAccSnapshotConfig_enhancedFlinkApplication(names.AttrTarget, rName+"-target", `
        application_restore_type = "RESTORE_FROM_CUSTOM_SNAPSHOT"
        snapshot_arn             = aws_kinesisanalyticsv2_snapshot.test.arn
`),
	)
}
