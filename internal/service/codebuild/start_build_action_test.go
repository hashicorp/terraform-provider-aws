// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package codebuild_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	awstypes "github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfcodebuild "github.com/hashicorp/terraform-provider-aws/internal/service/codebuild"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCodeBuildStartBuildAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: testAccCheckProjectDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStartBuildActionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBuildStarted(ctx, t, rName),
				),
			},
		},
	})
}

func TestAccCodeBuildStartBuildAction_withEnvironmentVariables(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: testAccCheckProjectDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStartBuildActionConfig_withEnvironmentVariables(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBuildStarted(ctx, t, rName),
				),
			},
		},
	})
}

func testAccCheckBuildStarted(ctx context.Context, t *testing.T, projectName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CodeBuildClient(ctx)

		// List builds for the project
		input := &codebuild.ListBuildsForProjectInput{
			ProjectName: &projectName,
		}

		timeout := time.After(5 * time.Minute)
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-timeout:
				return fmt.Errorf("timeout waiting for build to be started for project %s", projectName)
			case <-ticker.C:
				output, err := conn.ListBuildsForProject(ctx, input)
				if err != nil {
					continue
				}

				if len(output.Ids) == 0 {
					continue
				}

				// Get build details
				batchInput := &codebuild.BatchGetBuildsInput{
					Ids: output.Ids[:1], // Check most recent build
				}
				batchOutput, err := conn.BatchGetBuilds(ctx, batchInput)
				if err != nil {
					continue
				}

				if len(batchOutput.Builds) > 0 {
					build := batchOutput.Builds[0]
					// Verify build was started (any status other than not found)
					if build.BuildStatus != "" {
						return nil
					}
				}
			}
		}
	}
}

func testAccStartBuildActionConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "codebuild.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:${data.aws_partition.current.partition}:logs:*:*:*"
      }
    ]
  })
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "aws/codebuild/amazonlinux2-x86_64-standard:3.0"
    type         = "LINUX_CONTAINER"
  }

  source {
    type      = "NO_SOURCE"
    buildspec = "version: 0.2\nphases:\n  build:\n    commands:\n      - echo 'Hello World'"
  }
}

action "aws_codebuild_start_build" "test" {
  config {
    project_name = aws_codebuild_project.test.name
  }
}

resource "terraform_data" "trigger" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.aws_codebuild_start_build.test]
    }
  }

  depends_on = [aws_codebuild_project.test]
}
`, rName)
}

func testAccStartBuildActionConfig_withEnvironmentVariables(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "codebuild.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:${data.aws_partition.current.partition}:logs:*:*:*"
      }
    ]
  })
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "aws/codebuild/amazonlinux2-x86_64-standard:3.0"
    type         = "LINUX_CONTAINER"
  }

  source {
    type      = "NO_SOURCE"
    buildspec = "version: 0.2\nphases:\n  build:\n    commands:\n      - echo \"TEST_VAR is $TEST_VAR\""
  }
}

action "aws_codebuild_start_build" "test" {
  config {
    project_name = aws_codebuild_project.test.name

    environment_variables_override {
      name  = "TEST_VAR"
      value = "test_value"
      type  = "PLAINTEXT"
    }
  }
}

resource "terraform_data" "trigger" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.aws_codebuild_start_build.test]
    }
  }

  depends_on = [aws_codebuild_project.test]
}
`, rName)
}

// TestStartBuildActionHostKernelOverride verifies the start-build action model maps
// host_kernel_override to StartBuildInput.HostKernelOverride for every enum value,
// and leaves it empty when unset.
func TestStartBuildActionHostKernelOverride(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]struct {
		override fwtypes.StringEnum[awstypes.HostKernel]
		expected awstypes.HostKernel
	}{
		"LINUX_KERNEL_4":      {override: fwtypes.StringEnumValue(awstypes.HostKernelLinuxKernel4), expected: awstypes.HostKernelLinuxKernel4},
		"LINUX_KERNEL_6":      {override: fwtypes.StringEnumValue(awstypes.HostKernelLinuxKernel6), expected: awstypes.HostKernelLinuxKernel6},
		"LINUX_KERNEL_LATEST": {override: fwtypes.StringEnumValue(awstypes.HostKernelLinuxKernelLatest), expected: awstypes.HostKernelLinuxKernelLatest},
		"unset":               {override: fwtypes.StringEnumNull[awstypes.HostKernel](), expected: ""},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := tfcodebuild.StartBuildActionModel{
				ProjectName:                  types.StringValue("test-project"),
				EnvironmentVariablesOverride: fwtypes.NewListNestedObjectValueOfNull[tfcodebuild.EnvironmentVariableModel](ctx),
				HostKernelOverride:           tc.override,
			}

			var input codebuild.StartBuildInput
			if diags := fwflex.Expand(ctx, model, &input); diags.HasError() {
				t.Fatalf("fwflex.Expand returned errors: %v", diags)
			}

			if input.HostKernelOverride != tc.expected {
				t.Errorf("HostKernelOverride = %q, want %q", input.HostKernelOverride, tc.expected)
			}
		})
	}
}

func TestAccCodeBuildStartBuildAction_hostKernelOverride(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: testAccCheckProjectDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStartBuildActionConfig_hostKernelOverride(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBuildStarted(ctx, t, rName),
				),
			},
		},
	})
}

func testAccStartBuildActionConfig_hostKernelOverride(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "codebuild.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:${data.aws_partition.current.partition}:logs:*:*:*"
      }
    ]
  })
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "aws/codebuild/amazonlinux-x86_64-standard:6.0"
    type         = "LINUX_CONTAINER"
    host_kernel  = "LINUX_KERNEL_6"
  }

  source {
    type      = "NO_SOURCE"
    buildspec = "version: 0.2\nphases:\n  build:\n    commands:\n      - echo 'Hello World'"
  }
}

action "aws_codebuild_start_build" "test" {
  config {
    project_name         = aws_codebuild_project.test.name
    host_kernel_override = "LINUX_KERNEL_6"
  }
}

resource "terraform_data" "trigger" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.aws_codebuild_start_build.test]
    }
  }

  depends_on = [aws_codebuild_project.test]
}
`, rName)
}
