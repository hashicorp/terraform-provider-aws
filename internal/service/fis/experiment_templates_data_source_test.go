// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package fis_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccFISExperimentTemplatesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FISServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccExperimentTemplatesDataSourceConfig_basic(rName),
			},
			{
				Config: testAccExperimentTemplatesDataSourceConfig_dataSource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_fis_experiment_templates.selected", "ids.#", "1"),
					resource.TestCheckResourceAttr("data.aws_fis_experiment_templates.tier_1", "ids.#", "3"),
					acctest.CheckResourceAttrGreaterThanValue("data.aws_fis_experiment_templates.all", "ids.#", 0),
					resource.TestCheckResourceAttr("data.aws_fis_experiment_templates.none", "ids.#", "0"),
				),
			},
		},
	})
}

func testAccExperimentTemplatesDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_iam_policy_document" "assume" {
  version = "2012-10-17"
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["fis.amazonaws.com"]
    }
  }
}

data "aws_iam_policy" "fis_ec2_access" {
  name = "AWSFaultInjectionSimulatorEC2Access"
}

resource "aws_iam_role" "fis" {
  name                  = %[1]q
  assume_role_policy    = data.aws_iam_policy_document.assume.json
  force_detach_policies = true
  managed_policy_arns   = [data.aws_iam_policy.fis_ec2_access.arn]
}

resource "aws_fis_experiment_template" "fis_a" {
  description = "Stop EC2 instances"
  role_arn    = aws_iam_role.fis.arn

  stop_condition {
    source = "none"
  }

  action {
    name        = "stop-instances"
    action_id   = "aws:ec2:stop-instances"
    description = "Stop EC2 instances"

    target {
      key   = "Instances"
      value = "test-provider-1"
    }

    parameter {
      key   = "startInstancesAfterDuration"
      value = "PT2M"
    }
  }

  target {
    name           = "test-provider-1"
    resource_type  = "aws:ec2:instance"
    selection_mode = "ALL"

    resource_tag {
      key   = "Name"
      value = "fis"
    }
  }

  tags = {
    Name = %[1]q
    Tier = 1
  }
}

resource "aws_fis_experiment_template" "fis_b" {
  description = "Stop EC2 instances"
  role_arn    = aws_iam_role.fis.arn

  stop_condition {
    source = "none"
  }

  action {
    name        = "stop-instances"
    action_id   = "aws:ec2:stop-instances"
    description = "Stop EC2 instances"

    target {
      key   = "Instances"
      value = "test-provider-2"
    }

    parameter {
      key   = "startInstancesAfterDuration"
      value = "PT2M"
    }
  }

  target {
    name           = "test-provider-2"
    resource_type  = "aws:ec2:instance"
    selection_mode = "ALL"

    resource_tag {
      key   = "Name"
      value = "fis"
    }
  }

  tags = {
    Name = %[1]q
    Tier = 2
  }
}

resource "aws_fis_experiment_template" "fis_c" {
  description = "Stop EC2 instances"
  role_arn    = aws_iam_role.fis.arn

  stop_condition {
    source = "none"
  }

  action {
    name        = "stop-instances"
    action_id   = "aws:ec2:stop-instances"
    description = "Stop EC2 instances"

    target {
      key   = "Instances"
      value = "test-provider-3"
    }

    parameter {
      key   = "startInstancesAfterDuration"
      value = "PT2M"
    }
  }

  target {
    name           = "test-provider-3"
    resource_type  = "aws:ec2:instance"
    selection_mode = "ALL"

    resource_tag {
      key   = "Name"
      value = "fis"
    }
  }

  tags = {
    Name = %[1]q
    Tier = 1
  }
}

resource "aws_fis_experiment_template" "fis_d" {
  description = "Stop EC2 instances"
  role_arn    = aws_iam_role.fis.arn

  stop_condition {
    source = "none"
  }

  action {
    name        = "stop-instances"
    action_id   = "aws:ec2:stop-instances"
    description = "Stop EC2 instances"

    target {
      key   = "Instances"
      value = "test-provider-4"
    }

    parameter {
      key   = "startInstancesAfterDuration"
      value = "PT2M"
    }
  }

  target {
    name           = "test-provider-4"
    resource_type  = "aws:ec2:instance"
    selection_mode = "ALL"

    resource_tag {
      key   = "Name"
      value = "fis"
    }
  }

  tags = {
    Name = "selected"
    Tier = 1
  }
}
`, rName))
}

func testAccExperimentTemplatesDataSourceConfig_dataSource(rName string) string {
	return acctest.ConfigCompose(testAccExperimentTemplatesDataSourceConfig_basic(rName), `
data "aws_fis_experiment_templates" "selected" {
  tags = {
    Name = "selected"
  }
}

data "aws_fis_experiment_templates" "all" {}

data "aws_fis_experiment_templates" "tier_1" {
  tags = {
    Tier = 1
  }
}

data "aws_fis_experiment_templates" "none" {
  tags = {
    Name = "none"
  }
}
`)
}
