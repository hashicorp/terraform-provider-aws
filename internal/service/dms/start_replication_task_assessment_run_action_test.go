// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dms_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	dmssdk "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDMSStartReplicationTaskAssessmentRunAction_includeOnlySuccess(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_task.test"
	assessmentRunName := fmt.Sprintf("%s-assessment", rName)
	var v awstypes.ReplicationTask

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: testAccCheckReplicationTaskDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStartReplicationTaskAssessmentRunActionConfig_includeOnly(rName, assessmentRunName, "mysql-check-binlog-format"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, t, resourceName, &v),
					testAccCheckReplicationTaskAssessmentRunExists(ctx, t, resourceName, "aws_s3_bucket.test", "aws_iam_role.test", assessmentRunName),
				),
			},
		},
	})
}

func TestAccDMSStartReplicationTaskAssessmentRunAction_includeOnlyFailure(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	assessmentRunName := fmt.Sprintf("%s-assessment", rName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: testAccCheckReplicationTaskDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccStartReplicationTaskAssessmentRunActionConfig_includeOnly(rName, assessmentRunName, "mysql-check-target-network-parameter"),
				ExpectError: regexache.MustCompile(`DMS Assessment Run Failed: DMS assessment run\s+.* reached failure status failed\.`),
			},
		},
	})
}

func testAccCheckReplicationTaskAssessmentRunExists(ctx context.Context, t *testing.T, replicationTaskResourceName, bucketResourceName, roleResourceName, assessmentRunName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		taskRS, ok := s.RootModule().Resources[replicationTaskResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", replicationTaskResourceName)
		}

		bucketRS, ok := s.RootModule().Resources[bucketResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", bucketResourceName)
		}

		roleRS, ok := s.RootModule().Resources[roleResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", roleResourceName)
		}

		conn := acctest.ProviderMeta(ctx, t).DMSClient(ctx)
		replicationTaskARN := taskRS.Primary.Attributes["replication_task_arn"]
		resultBucket := bucketRS.Primary.Attributes[names.AttrBucket]
		serviceAccessRoleARN := roleRS.Primary.Attributes[names.AttrARN]

		timeout := time.After(10 * time.Minute)
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-timeout:
				return fmt.Errorf("timeout waiting for DMS assessment run %q for replication task %s", assessmentRunName, replicationTaskARN)
			case <-ticker.C:
				input := dmssdk.DescribeReplicationTaskAssessmentRunsInput{
					Filters: []awstypes.Filter{
						{
							Name:   aws.String("replication-task-arn"),
							Values: []string{replicationTaskARN},
						},
					},
				}
				output, err := conn.DescribeReplicationTaskAssessmentRuns(ctx, &input)
				if err != nil {
					continue
				}

				for _, run := range output.ReplicationTaskAssessmentRuns {
					if aws.ToString(run.AssessmentRunName) != assessmentRunName {
						continue
					}

					if aws.ToString(run.ReplicationTaskAssessmentRunArn) == "" {
						return fmt.Errorf("assessment run %q returned an empty ARN", assessmentRunName)
					}

					if aws.ToString(run.ReplicationTaskArn) != replicationTaskARN {
						return fmt.Errorf("expected replication task ARN %s, got %s", replicationTaskARN, aws.ToString(run.ReplicationTaskArn))
					}

					if aws.ToString(run.ResultLocationBucket) != resultBucket {
						return fmt.Errorf("expected result bucket %s, got %s", resultBucket, aws.ToString(run.ResultLocationBucket))
					}

					if aws.ToString(run.ServiceAccessRoleArn) != serviceAccessRoleARN {
						return fmt.Errorf("expected service access role ARN %s, got %s", serviceAccessRoleARN, aws.ToString(run.ServiceAccessRoleArn))
					}

					switch status := aws.ToString(run.Status); status {
					case "passed", "warning":
						return nil
					case "starting", "provisioning", "running":
						continue
					default:
						return fmt.Errorf("assessment run %q reached unexpected status %q: %s", assessmentRunName, status, aws.ToString(run.LastFailureMessage))
					}
				}
			}
		}
	}
}

func testAccStartReplicationTaskAssessmentRunActionConfig_includeOnly(rName, assessmentRunName, includeOnly string) string {
	return acctest.ConfigCompose(testAccReplicationTaskConfig_start(rName, false, "1"), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "dms.${data.aws_partition.current.dns_suffix}"
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "s3:GetBucketLocation",
        "s3:ListBucket",
        "s3:GetObject",
        "s3:PutObject",
        "s3:DeleteObject"
      ]
      Resource = [
        aws_s3_bucket.test.arn,
        "${aws_s3_bucket.test.arn}/*"
      ]
    }]
  })
}

action "aws_dms_start_replication_task_assessment_run" "test" {
  config {
    assessment_run_name     = %[2]q
    replication_task_arn    = aws_dms_replication_task.test.replication_task_arn
    include_only            = [%[3]q]
    result_location_bucket  = aws_s3_bucket.test.bucket
    result_location_folder  = "assessment-results"
    service_access_role_arn = aws_iam_role.test.arn
    timeout                 = 1800
  }
}

resource "terraform_data" "trigger" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.aws_dms_start_replication_task_assessment_run.test]
    }
  }

  depends_on = [
    aws_dms_replication_task.test,
    aws_iam_role_policy.test
  ]
}
`, rName, assessmentRunName, includeOnly))
}
