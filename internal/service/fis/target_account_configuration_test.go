// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package fis_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fis"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fis/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tffis "github.com/hashicorp/terraform-provider-aws/internal/service/fis"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccFISTargetAccountConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var targetaccountconfiguration awstypes.TargetAccountConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fis_target_account_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FISServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetAccountConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetAccountConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetAccountConfigurationExists(ctx, resourceName, &targetaccountconfiguration),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAccountID),
					resource.TestCheckResourceAttrSet(resourceName, "experiment_template_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, fmt.Sprintf("%s target account configuration", rName)),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrRoleARN, "iam", regexache.MustCompile(`role/.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", names.AttrAccountID, "experiment_template_id"),
				ImportStateVerifyIdentifierAttribute: names.AttrAccountID,
			},
		},
	})
}

func TestAccFISTargetAccountConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)

	var before, after awstypes.TargetAccountConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fis_target_account_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FISServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetAccountConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetAccountConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetAccountConfigurationExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, fmt.Sprintf("%s target account configuration", rName)),
				),
			},
			{
				Config: testAccTargetAccountConfigurationConfig_update(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetAccountConfigurationExists(ctx, resourceName, &after),
					testAccCheckTargetAccountConfigurationNotRecreated(&before, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, fmt.Sprintf("%s target account configuration updated", rName)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", names.AttrAccountID, "experiment_template_id"),
				ImportStateVerifyIdentifierAttribute: names.AttrAccountID,
			},
		},
	})
}

func TestAccFISTargetAccountConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var targetaccountconfiguration awstypes.TargetAccountConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fis_target_account_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FISServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetAccountConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetAccountConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetAccountConfigurationExists(ctx, resourceName, &targetaccountconfiguration),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tffis.ResourceTargetAccountConfiguration, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTargetAccountConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).FISClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_fis_target_account_configuration" {
				continue
			}

			accountId := aws.String(rs.Primary.Attributes[names.AttrAccountID])
			experimentId := aws.String(rs.Primary.Attributes["experiment_template_id"])
			_, err := tffis.FindTargetAccountConfigurationByID(ctx, conn, accountId, experimentId)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.FIS, create.ErrActionCheckingDestroyed, tffis.ResNameTargetAccountConfiguration, *experimentId, err)
			}

			return create.Error(names.FIS, create.ErrActionCheckingDestroyed, tffis.ResNameTargetAccountConfiguration, *experimentId, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckTargetAccountConfigurationExists(ctx context.Context, name string, targetaccountconfiguration *awstypes.TargetAccountConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.FIS, create.ErrActionCheckingExistence, tffis.ResNameTargetAccountConfiguration, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.FIS, create.ErrActionCheckingExistence, tffis.ResNameTargetAccountConfiguration, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FISClient(ctx)

		accountId := aws.String(rs.Primary.Attributes[names.AttrAccountID])
		experimentTemplateId := aws.String(rs.Primary.Attributes["experiment_template_id"])

		resp, err := tffis.FindTargetAccountConfigurationByID(ctx, conn, accountId, experimentTemplateId)
		if err != nil {
			return create.Error(names.FIS, create.ErrActionCheckingExistence, tffis.ResNameTargetAccountConfiguration, rs.Primary.ID, err)
		}

		*targetaccountconfiguration = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).FISClient(ctx)

	input := &fis.ListExperimentTemplatesInput{}

	_, err := conn.ListExperimentTemplates(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckTargetAccountConfigurationNotRecreated(before, after *awstypes.TargetAccountConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.AccountId), aws.ToString(after.AccountId); before != after {
			return create.Error(names.FIS, create.ErrActionCheckingNotRecreated, tffis.ResNameTargetAccountConfiguration, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccTargetAccountConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = "%[1]s-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "fis.${data.aws_partition.current.dns_suffix}"
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSFaultInjectionSimulatorEC2Access"
}

resource "aws_fis_experiment_template" "test" {
  description = "%[1]s experiment template"
  role_arn    = aws_iam_role.test.arn

  experiment_options {
    account_targeting            = "multi-account"
    empty_target_resolution_mode = "fail"
  }

  action {
    name      = "stop-instances"
    action_id = "aws:ec2:stop-instances"

    parameter {
      key   = "startInstancesAfterDuration"
      value = "PT10M"
    }

    target {
      key   = "Instances"
      value = "test-instances"
    }
  }

  target {
    name           = "test-instances"
    resource_type  = "aws:ec2:instance"
    selection_mode = "PERCENT(50)"

    resource_tag {
      key   = "Environment"
      value = "test"
    }
  }

  stop_condition {
    source = "none"
  }

  tags = {
    Name = "%[1]s-experiment-template"
  }
}

resource "aws_fis_target_account_configuration" "test" {
  experiment_template_id = aws_fis_experiment_template.test.id
  account_id             = data.aws_caller_identity.current.account_id
  role_arn               = aws_iam_role.test.arn
  description            = "%[1]s target account configuration"
}
`, rName)
}

func testAccTargetAccountConfigurationConfig_update(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = "%[1]s-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "fis.${data.aws_partition.current.dns_suffix}"
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSFaultInjectionSimulatorEC2Access"
}

resource "aws_fis_experiment_template" "test" {
  description = "%[1]s experiment template"
  role_arn    = aws_iam_role.test.arn

  experiment_options {
    account_targeting            = "multi-account"
    empty_target_resolution_mode = "fail"
  }

  action {
    name      = "stop-instances"
    action_id = "aws:ec2:stop-instances"

    parameter {
      key   = "startInstancesAfterDuration"
      value = "PT10M"
    }

    target {
      key   = "Instances"
      value = "test-instances"
    }
  }

  target {
    name           = "test-instances"
    resource_type  = "aws:ec2:instance"
    selection_mode = "PERCENT(50)"

    resource_tag {
      key   = "Environment"
      value = "test"
    }
  }

  stop_condition {
    source = "none"
  }

  tags = {
    Name = "%[1]s-experiment-template"
  }
}

resource "aws_fis_target_account_configuration" "test" {
  experiment_template_id = aws_fis_experiment_template.test.id
  account_id             = data.aws_caller_identity.current.account_id
  role_arn               = aws_iam_role.test.arn
  description            = "%[1]s target account configuration updated"
}
`, rName)
}
