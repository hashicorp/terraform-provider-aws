package fis_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fis"
	"github.com/aws/aws-sdk-go-v2/service/fis/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tffis "github.com/hashicorp/terraform-provider-aws/internal/service/fis"
)

func TestAccFISExperimentTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fis_experiment_template.test"
	var conf types.ExperimentTemplate

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, fis.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExperimentTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExperimentTemplateConfig_basic(rName, "An experiment template for testing", "test-action-1", "", "aws:ec2:terminate-instances", "Instances", "to-terminate-1", "aws:ec2:instance", "COUNT(1)", "env", "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccExperimentTemplateExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "An experiment template for testing"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "stop_condition.0.source", "none"),
					resource.TestCheckResourceAttr(resourceName, "stop_condition.0.value", ""),
					resource.TestCheckResourceAttr(resourceName, "stop_condition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.name", "test-action-1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.description", ""),
					resource.TestCheckResourceAttr(resourceName, "action.0.action_id", "aws:ec2:terminate-instances"),
					resource.TestCheckResourceAttr(resourceName, "action.0.parameter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.start_after.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.0.key", "Instances"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.0.value", "to-terminate-1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target.0.name", "to-terminate-1"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_type", "aws:ec2:instance"),
					resource.TestCheckResourceAttr(resourceName, "target.0.selection_mode", "COUNT(1)"),
					resource.TestCheckResourceAttr(resourceName, "target.0.filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_arns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.0.key", "env"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.0.value", "test"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target.#", "1"),
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

func TestAccFISExperimentTemplate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fis_experiment_template.test"
	var conf types.ExperimentTemplate

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, fis.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExperimentTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExperimentTemplateConfig_basic(rName, "An experiment template for testing", "test-action-1", "", "aws:ec2:terminate-instances", "Instances", "to-terminate-1", "aws:ec2:instance", "COUNT(1)", "env", "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccExperimentTemplateExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tffis.ResourceExperimentTemplate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFISExperimentTemplate_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fis_experiment_template.test"
	var conf types.ExperimentTemplate

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, fis.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExperimentTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExperimentTemplateConfig_basic(rName, "An experiment template for testing", "test-action-1", "", "aws:ec2:terminate-instances", "Instances", "to-terminate-1", "aws:ec2:instance", "COUNT(1)", "env", "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccExperimentTemplateExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "An experiment template for testing"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "stop_condition.0.source", "none"),
					resource.TestCheckResourceAttr(resourceName, "stop_condition.0.value", ""),
					resource.TestCheckResourceAttr(resourceName, "stop_condition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.name", "test-action-1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.description", ""),
					resource.TestCheckResourceAttr(resourceName, "action.0.action_id", "aws:ec2:terminate-instances"),
					resource.TestCheckResourceAttr(resourceName, "action.0.parameter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.start_after.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.0.key", "Instances"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.0.value", "to-terminate-1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target.0.name", "to-terminate-1"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_type", "aws:ec2:instance"),
					resource.TestCheckResourceAttr(resourceName, "target.0.selection_mode", "COUNT(1)"),
					resource.TestCheckResourceAttr(resourceName, "target.0.filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_arns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.0.key", "env"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.0.value", "test"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target.#", "1"),
				),
			},
			{
				Config: testAccExperimentTemplateConfig_basic(rName, "Artic Lake", "test-action-2", "Lane 8", "aws:ec2:stop-instances", "Instances", "to-stop-1", "aws:ec2:instance", "ALL", "env2", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccExperimentTemplateExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "Artic Lake"),
					resource.TestCheckResourceAttr(resourceName, "action.0.name", "test-action-2"),
					resource.TestCheckResourceAttr(resourceName, "action.0.description", "Lane 8"),
					resource.TestCheckResourceAttr(resourceName, "action.0.action_id", "aws:ec2:stop-instances"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.0.key", "Instances"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.0.value", "to-stop-1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target.0.name", "to-stop-1"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_type", "aws:ec2:instance"),
					resource.TestCheckResourceAttr(resourceName, "target.0.selection_mode", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.0.key", "env2"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.0.value", "test2"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.#", "1"),
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

func TestAccFISExperimentTemplate_spot(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fis_experiment_template.test"
	var conf types.ExperimentTemplate

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, fis.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExperimentTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExperimentTemplateConfig_actionParameter(rName, "Send Spot Instance Interruptions", "Send-Spot-Instance-Interruptions", "Send Spot Instance Interruptions", "aws:ec2:send-spot-instance-interruptions", "SpotInstances", "send-spot-instance-interruptions-target", "durationBeforeInterruption", "PT2M", "aws:ec2:spot-instance", "PERCENT(25)", "env", "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccExperimentTemplateExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "Send Spot Instance Interruptions"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "stop_condition.0.source", "none"),
					resource.TestCheckResourceAttr(resourceName, "stop_condition.0.value", ""),
					resource.TestCheckResourceAttr(resourceName, "stop_condition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.name", "Send-Spot-Instance-Interruptions"),
					resource.TestCheckResourceAttr(resourceName, "action.0.description", "Send Spot Instance Interruptions"),
					resource.TestCheckResourceAttr(resourceName, "action.0.action_id", "aws:ec2:send-spot-instance-interruptions"),
					resource.TestCheckResourceAttr(resourceName, "action.0.parameter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.parameter.0.key", "durationBeforeInterruption"),
					resource.TestCheckResourceAttr(resourceName, "action.0.parameter.0.value", "PT2M"),
					resource.TestCheckResourceAttr(resourceName, "action.0.start_after.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.0.key", "SpotInstances"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.0.value", "send-spot-instance-interruptions-target"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target.0.name", "send-spot-instance-interruptions-target"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_type", "aws:ec2:spot-instance"),
					resource.TestCheckResourceAttr(resourceName, "target.0.selection_mode", "PERCENT(25)"),
					resource.TestCheckResourceAttr(resourceName, "target.0.filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_arns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.0.key", "env"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.0.value", "test"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target.#", "1"),
				),
			},
		},
	})
}

func testAccExperimentTemplateExists(ctx context.Context, resourceName string, config *types.ExperimentTemplate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FISClient()
		out, err := conn.GetExperimentTemplate(ctx, &fis.GetExperimentTemplateInput{Id: aws.String(rs.Primary.ID)})

		if err != nil {
			return fmt.Errorf("Describe Experiment Template error: %v", err)
		}

		if out.ExperimentTemplate == nil {
			return fmt.Errorf("No Experiment Template returned %v in %v", out.ExperimentTemplate, out)
		}

		*out.ExperimentTemplate = *config

		return nil
	}
}

func testAccCheckExperimentTemplateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).FISClient()
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_fis_experiment_template" {
				continue
			}

			_, err := conn.GetExperimentTemplate(ctx, &fis.GetExperimentTemplateInput{Id: aws.String(rs.Primary.ID)})

			var nf *types.ResourceNotFoundException
			if !tfawserr.ErrStatusCodeEquals(err, tffis.ErrCodeNotFound) && !errors.As(err, &nf) {
				return fmt.Errorf("Experiment Template '%s' was not deleted properly", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccExperimentTemplateConfig_basic(rName, desc, actionName, actionDesc, actionID, actionTargetK, actionTargetV, targetResType, targetSelectMode, targetResTagK, targetResTagV string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = [
          "fis.${data.aws_partition.current.dns_suffix}",
        ]
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_fis_experiment_template" "test" {
  description = %[2]q
  role_arn    = aws_iam_role.test.arn

  stop_condition {
    source = "none"
  }

  action {
    name        = %[3]q
    description = %[4]q
    action_id   = %[5]q

    target {
      key   = %[6]q
      value = %[7]q
    }
  }

  target {
    name           = %[7]q
    resource_type  = %[8]q
    selection_mode = %[9]q

    resource_tag {
      key   = %[10]q
      value = %[11]q
    }
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, desc, actionName, actionDesc, actionID, actionTargetK, actionTargetV, targetResType, targetSelectMode, targetResTagK, targetResTagV)
}

func testAccExperimentTemplateConfig_actionParameter(rName, desc, actionName, actionDesc, actionID, actionTargetK, actionTargetV, paramK, paramV, targetResType, targetSelectMode, targetResTagK, targetResTagV string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = [
          "fis.${data.aws_partition.current.dns_suffix}",
        ]
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_fis_experiment_template" "test" {
  description = %[2]q
  role_arn    = aws_iam_role.test.arn

  stop_condition {
    source = "none"
  }

  action {
    name        = %[3]q
    description = %[4]q
    action_id   = %[5]q

    target {
      key   = %[6]q
      value = %[7]q
    }

    parameter {
      key   = %[8]q
      value = %[9]q
    }
  }

  target {
    name           = %[7]q
    resource_type  = %[10]q
    selection_mode = %[11]q

    resource_tag {
      key   = %[12]q
      value = %[13]q
    }
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, desc, actionName, actionDesc, actionID, actionTargetK, actionTargetV, paramK, paramV, targetResType, targetSelectMode, targetResTagK, targetResTagV)
}
