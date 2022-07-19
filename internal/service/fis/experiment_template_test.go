package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fis"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_fis_experiment_template", &resource.Sweeper{
		Name: "aws_fis_experiment_template",
		F:    testSweepFisExperimentTemplates,
	})
}

func testSweepFisExperimentTemplates(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).fisconn

	err = conn.ListExperimentTemplatesPages(&fis.ListExperimentTemplatesInput{}, func(page *fis.ListExperimentTemplatesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, experimentTemplate := range page.ExperimentTemplates {
			deleteExperimentTemplateInput := &fis.DeleteExperimentTemplateInput{Id: experimentTemplate.Id}

			log.Printf("[INFO] Deleting FIS Experiment Template: %s", aws.StringValue(experimentTemplate.Id))
			_, err = conn.DeleteExperimentTemplate(deleteExperimentTemplateInput)

			if err != nil {
				log.Printf("[ERROR] Error deleting FIS Experiment Template (%s): %s", aws.StringValue(experimentTemplate.Id), err)
			}

		}

		return !lastPage
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping FIS Experiment Template sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving FIS Experiment Templates: %s", err)
	}

	return nil
}

func TestAccAWSFisExperimentTemplate_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_fis_experiment_template.test"
	var conf fis.ExperimentTemplate

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, fis.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccAWSFisExperimentTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSFisExperimentTemplateConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSFisExperimentTemplateExists(resourceName, &conf),
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

func TestAccAWSFisExperimentTemplate_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_fis_experiment_template.test"
	var conf fis.ExperimentTemplate

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, fis.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccAWSFisExperimentTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSFisExperimentTemplateConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSFisExperimentTemplateExists(resourceName, &conf),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsFisExperimentTemplate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSFisExperimentTemplateExists(resourceName string, config *fis.ExperimentTemplate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).fisconn
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

func testAccAWSFisExperimentTemplateDestroy(s *terraform.State) error {
	meta := testAccProvider.Meta()
	conn := meta.(*AWSClient).fisconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_fis_experiment_template" {
			continue
		}

		_, err := conn.GetExperimentTemplate(&fis.GetExperimentTemplateInput{Id: aws.String(rs.Primary.ID)})
		if !isAWSErrRequestFailureStatusCode(err, 404) {
			return fmt.Errorf("Experiment Template '%s' was not deleted properly", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAWSFisExperimentTemplateConfigBasic(rName string) string {
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
