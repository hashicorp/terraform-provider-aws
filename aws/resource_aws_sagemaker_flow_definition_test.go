package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/sagemaker/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_sagemaker_flow_definition", &resource.Sweeper{
		Name: "aws_sagemaker_flow_definition",
		F:    testSweepSagemakerFlowDefinitions,
	})
}

func testSweepSagemakerFlowDefinitions(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).sagemakerconn
	var sweeperErrs *multierror.Error

	err = conn.ListFlowDefinitionsPages(&sagemaker.ListFlowDefinitionsInput{}, func(page *sagemaker.ListFlowDefinitionsOutput, lastPage bool) bool {
		for _, flowDefinition := range page.FlowDefinitionSummaries {

			r := resourceAwsSagemakerFlowDefinition()
			d := r.Data(nil)
			d.SetId(aws.StringValue(flowDefinition.FlowDefinitionName))
			err := r.Delete(d, client)
			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker Flow Definition sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Sagemaker Flow Definitions: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func testAccAWSSagemakerFlowDefinition_basic(t *testing.T) {
	var flowDefinition sagemaker.DescribeFlowDefinitionOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_flow_definition.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerFlowDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerFlowDefinitionBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerFlowDefinitionExists(resourceName, &flowDefinition),
					resource.TestCheckResourceAttr(resourceName, "flow_definition_name", rName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("flow-definition/%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "human_loop_request_source.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "human_loop_activation_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "human_loop_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "human_loop_config.0.public_workforce_task_price.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "human_loop_config.0.human_task_ui_arn", "aws_sagemaker_human_task_ui.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "human_loop_config.0.task_availability_lifetime_in_seconds", "1"),
					resource.TestCheckResourceAttr(resourceName, "human_loop_config.0.task_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "human_loop_config.0.task_description", rName),
					resource.TestCheckResourceAttr(resourceName, "human_loop_config.0.task_title", rName),
					resource.TestCheckResourceAttrPair(resourceName, "human_loop_config.0.workteam_arn", "aws_sagemaker_workteam.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "output_config.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "output_config.0.s3_output_path"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func testAccAWSSagemakerFlowDefinition_humanLoopConfig_publicWorkforce(t *testing.T) {
	var flowDefinition sagemaker.DescribeFlowDefinitionOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_flow_definition.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerFlowDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerFlowDefinitionPublicWorkforceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerFlowDefinitionExists(resourceName, &flowDefinition),
					resource.TestCheckResourceAttr(resourceName, "flow_definition_name", rName),
					resource.TestCheckResourceAttr(resourceName, "human_loop_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "human_loop_config.0.public_workforce_task_price.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "human_loop_config.0.public_workforce_task_price.0.amount_in_usd.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "human_loop_config.0.public_workforce_task_price.0.amount_in_usd.0.cents", "1"),
					resource.TestCheckResourceAttr(resourceName, "human_loop_config.0.public_workforce_task_price.0.amount_in_usd.0.tenth_fractions_of_a_cent", "2"),
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

func testAccAWSSagemakerFlowDefinition_humanLoopRequestSource(t *testing.T) {
	var flowDefinition sagemaker.DescribeFlowDefinitionOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_flow_definition.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerFlowDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerFlowDefinitionHumanLoopRequestSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerFlowDefinitionExists(resourceName, &flowDefinition),
					resource.TestCheckResourceAttr(resourceName, "flow_definition_name", rName),
					resource.TestCheckResourceAttr(resourceName, "human_loop_request_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "human_loop_request_source.0.aws_managed_human_loop_request_source", "AWS/Textract/AnalyzeDocument/Forms/V1"),
					resource.TestCheckResourceAttr(resourceName, "human_loop_activation_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "human_loop_activation_config.0.human_loop_activation_conditions_config.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "human_loop_activation_config.0.human_loop_activation_conditions_config.0.human_loop_activation_conditions"),
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

func testAccAWSSagemakerFlowDefinition_tags(t *testing.T) {
	var flowDefinition sagemaker.DescribeFlowDefinitionOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_flow_definition.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerFlowDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerFlowDefinitionTagsConfig1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerFlowDefinitionExists(resourceName, &flowDefinition),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSagemakerFlowDefinitionTagsConfig2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerFlowDefinitionExists(resourceName, &flowDefinition),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSagemakerFlowDefinitionTagsConfig1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerFlowDefinitionExists(resourceName, &flowDefinition),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccAWSSagemakerFlowDefinition_disappears(t *testing.T) {
	var flowDefinition sagemaker.DescribeFlowDefinitionOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_flow_definition.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerFlowDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerFlowDefinitionBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerFlowDefinitionExists(resourceName, &flowDefinition),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSagemakerFlowDefinition(), resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSagemakerFlowDefinition(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSagemakerFlowDefinitionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_flow_definition" {
			continue
		}

		_, err := finder.FlowDefinitionByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("SageMaker Flow Definition %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSSagemakerFlowDefinitionExists(n string, flowDefinition *sagemaker.DescribeFlowDefinitionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SageMaker Flow Definition ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

		output, err := finder.FlowDefinitionByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*flowDefinition = *output

		return nil
	}
}

func testAccAWSSagemakerFlowDefinitionBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_human_task_ui" "test" {
  human_task_ui_name = %[1]q

  ui_template {
    content = file("test-fixtures/sagemaker-human-task-ui-tmpl.html")
  }
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = true
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:PutObject"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetBucketLocation"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}
`, rName)
}

func testAccAWSSagemakerFlowDefinitionBasicConfig(rName string) string {
	return composeConfig(testAccAWSSagemakerFlowDefinitionBaseConfig(rName),
		testAccAWSSagemakerWorkteamCognitoConfig(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_flow_definition" "test" {
  flow_definition_name = %[1]q
  role_arn             = aws_iam_role.test.arn

  human_loop_config {
    human_task_ui_arn                     = aws_sagemaker_human_task_ui.test.arn
    task_availability_lifetime_in_seconds = 1
    task_count                            = 1
    task_description                      = %[1]q
    task_title                            = %[1]q
    workteam_arn                          = aws_sagemaker_workteam.test.arn
  }

  output_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/"
  }
}
`, rName))
}

func testAccAWSSagemakerFlowDefinitionPublicWorkforceConfig(rName string) string {
	return composeConfig(testAccAWSSagemakerFlowDefinitionBaseConfig(rName),
		fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_sagemaker_flow_definition" "test" {
  flow_definition_name = %[1]q
  role_arn             = aws_iam_role.test.arn

  human_loop_config {
    human_task_ui_arn                     = aws_sagemaker_human_task_ui.test.arn
    task_availability_lifetime_in_seconds = 1
    task_count                            = 1
    task_description                      = %[1]q
    task_title                            = %[1]q
    workteam_arn                          = "arn:${data.aws_partition.current.partition}:sagemaker:${data.aws_region.current.name}:394669845002:workteam/public-crowd/default"

    public_workforce_task_price {
      amount_in_usd {
        cents                     = 1
        tenth_fractions_of_a_cent = 2
      }
    }
  }

  output_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/"
  }
}
`, rName))
}

func testAccAWSSagemakerFlowDefinitionHumanLoopRequestSourceConfig(rName string) string {
	return composeConfig(testAccAWSSagemakerFlowDefinitionBaseConfig(rName),
		testAccAWSSagemakerWorkteamCognitoConfig(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_flow_definition" "test" {
  flow_definition_name = %[1]q
  role_arn             = aws_iam_role.test.arn

  human_loop_config {
    human_task_ui_arn                     = aws_sagemaker_human_task_ui.test.arn
    task_availability_lifetime_in_seconds = 1
    task_count                            = 1
    task_description                      = %[1]q
    task_title                            = %[1]q
    workteam_arn                          = aws_sagemaker_workteam.test.arn
  }

  human_loop_request_source {
    aws_managed_human_loop_request_source = "AWS/Textract/AnalyzeDocument/Forms/V1"
  }

  human_loop_activation_config {
    human_loop_activation_conditions_config {
      human_loop_activation_conditions = <<EOF
        {
			"Conditions": [
			  {
				"ConditionType": "Sampling",
				"ConditionParameters": {
				  "RandomSamplingPercentage": 5
				}
			  }
			]
		}
        EOF
    }
  }

  output_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/"
  }
}
`, rName))
}

func testAccAWSSagemakerFlowDefinitionTagsConfig1(rName, tagKey1, tagValue1 string) string {
	return composeConfig(testAccAWSSagemakerFlowDefinitionBaseConfig(rName),
		testAccAWSSagemakerWorkteamCognitoConfig(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_flow_definition" "test" {
  flow_definition_name = %[1]q
  role_arn             = aws_iam_role.test.arn

  human_loop_config {
    human_task_ui_arn                     = aws_sagemaker_human_task_ui.test.arn
    task_availability_lifetime_in_seconds = 1
    task_count                            = 1
    task_description                      = %[1]q
    task_title                            = %[1]q
    workteam_arn                          = aws_sagemaker_workteam.test.arn
  }

  output_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccAWSSagemakerFlowDefinitionTagsConfig2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(testAccAWSSagemakerFlowDefinitionBaseConfig(rName),
		testAccAWSSagemakerWorkteamCognitoConfig(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_flow_definition" "test" {
  flow_definition_name = %[1]q
  role_arn             = aws_iam_role.test.arn

  human_loop_config {
    human_task_ui_arn                     = aws_sagemaker_human_task_ui.test.arn
    task_availability_lifetime_in_seconds = 1
    task_count                            = 1
    task_description                      = %[1]q
    task_title                            = %[1]q
    workteam_arn                          = aws_sagemaker_workteam.test.arn
  }

  output_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
