package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/finder"
)

func init() {
	resource.AddTestSweepers("aws_sagemaker_feature_group", &resource.Sweeper{
		Name: "aws_sagemaker_feature_group",
		F:    testSweepSagemakerFeatureGroups,
	})
}

func testSweepSagemakerFeatureGroups(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).sagemakerconn

	err = conn.ListFeatureGroupsPages(&sagemaker.ListFeatureGroupsInput{}, func(page *sagemaker.ListFeatureGroupsOutput, lastPage bool) bool {
		for _, group := range page.FeatureGroupSummaries {
			name := aws.StringValue(group.FeatureGroupName)

			input := &sagemaker.DeleteFeatureGroupInput{
				FeatureGroupName: group.FeatureGroupName,
			}

			log.Printf("[INFO] Deleting SageMaker Feature Group: %s", name)
			if _, err := conn.DeleteFeatureGroup(input); err != nil {
				log.Printf("[ERROR] Error deleting SageMaker Feature Group (%s): %s", name, err)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker Feature Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving SageMaker Feature Groups: %w", err)
	}

	return nil
}

func TestAccAWSSagemakerFeatureGroup_basic(t *testing.T) {
	var notebook sagemaker.DescribeFeatureGroupOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_feature_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerFeatureGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerFeatureGroupBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerFeatureGroupExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "event_time_feature_name", rName),
					resource.TestCheckResourceAttr(resourceName, "record_identifier_feature_name", rName),
					resource.TestCheckResourceAttr(resourceName, "online_store_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "online_store_config.0.enable_online_store", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.0.feature_name", rName),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.0.feature_type", "String"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("feature-group/%s", rName)),
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

func TestAccAWSSagemakerFeatureGroup_disappears(t *testing.T) {
	var notebook sagemaker.DescribeFeatureGroupOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_feature_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerFeatureGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerFeatureGroupBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerFeatureGroupExists(resourceName, &notebook),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSagemakerFeatureGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSagemakerFeatureGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_feature_group" {
			continue
		}

		codeRepository, err := finder.FeatureGroupByName(conn, rs.Primary.ID)
		if err != nil {
			return nil
		}

		if aws.StringValue(codeRepository.FeatureGroupName) == rs.Primary.ID {
			return fmt.Errorf("Sagemaker Feature Group %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSSagemakerFeatureGroupExists(n string, codeRepo *sagemaker.DescribeFeatureGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker Feature Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn
		resp, err := finder.FeatureGroupByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*codeRepo = *resp

		return nil
	}
}

func testAccAWSSagemakerFeatureGroupBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}
`, rName)
}

func testAccAWSSagemakerFeatureGroupBasicConfig(rName string) string {
	return testAccAWSSagemakerFeatureGroupBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_feature_group" "test" {
  feature_group_name             = %[1]q
  record_identifier_feature_name = %[1]q
  event_time_feature_name        = %[1]q
  role_arn                       = aws_iam_role.test.arn

  feature_definition {
	feature_name = %[1]q
    feature_type = "String"
  }

  online_store_config {
	enable_online_store = true
  }  
}
`, rName)
}
