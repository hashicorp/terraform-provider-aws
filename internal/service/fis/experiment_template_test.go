package fis_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fis"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tffis "github.com/hashicorp/terraform-provider-aws/internal/service/fis"
)

func TestAccFISExperimentTemplate_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fis_experiment_template.test"
	var conf fis.ExperimentTemplate

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, fis.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccExperimentTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExperimentTemplateConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccExperimentTemplateExists(resourceName, &conf),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fis_experiment_template.test"
	var conf fis.ExperimentTemplate

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, fis.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccExperimentTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExperimentTemplateConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccExperimentTemplateExists(resourceName, &conf),
					acctest.CheckResourceDisappears(acctest.Provider, tffis.ResourceExperimentTemplate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccExperimentTemplateExists(resourceName string, config *fis.ExperimentTemplate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FISConn
		out, err := conn.GetExperimentTemplate(&fis.GetExperimentTemplateInput{Id: aws.String(rs.Primary.ID)})

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

func testAccExperimentTemplateDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).FISConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_fis_experiment_template" {
			continue
		}

		_, err := conn.GetExperimentTemplate(&fis.GetExperimentTemplateInput{Id: aws.String(rs.Primary.ID)})
		if !tfawserr.ErrCodeEquals(err, fis.ErrCodeResourceNotFoundException) {
			return fmt.Errorf("Experiment Template '%s' was not deleted properly", rs.Primary.ID)
		}
	}

	return nil
}

func testAccExperimentTemplateConfigBasic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "fis.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_fis_experiment_template" "test" {
  description = "An experiment template for testing"
  role_arn    = aws_iam_role.test.arn

  stop_condition {
    source = "none"
  }

  action {
    name      = "test-action-1"
    action_id = "aws:ec2:terminate-instances"

    target {
      key   = "Instances"
      value = "to-terminate-1"
    }
  }

  target {
    name           = "to-terminate-1"
    resource_type  = "aws:ec2:instance"
    selection_mode = "COUNT(1)"

    resource_tag {
      key   = "env"
      value = "test"
    }
  }
}
`, rName)
}
