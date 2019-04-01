package aws

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/service/sagemaker"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform/helper/resource"
)

const SagemakerLifecycleConfigurationResourcePrefix = "tf-acc-test"

func init() {
	resource.AddTestSweepers("aws_sagemaker_lifecycle_config", &resource.Sweeper{
		Name: "aws_sagemaker_lifecycle_config",
		F:    testSweepSagemakerLifecycleConfiguration,
	})
}

func testSweepSagemakerLifecycleConfiguration(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).sagemakerconn

	input := &sagemaker.ListNotebookInstanceLifecycleConfigsInput{}
	err = conn.ListNotebookInstanceLifecycleConfigsPages(input, func(page *sagemaker.ListNotebookInstanceLifecycleConfigsOutput, lastPage bool) bool {
		if len(page.NotebookInstanceLifecycleConfigs) == 0 {
			log.Printf("[INFO] No SageMaker notebook instance lifecycle configuration to sweep")
			return false
		}
		for _, lifecycleConfig := range page.NotebookInstanceLifecycleConfigs {
			name := aws.StringValue(lifecycleConfig.NotebookInstanceLifecycleConfigName)
			if !strings.HasPrefix(name, SagemakerLifecycleConfigurationResourcePrefix) {
				log.Printf("[INFO] Skipping SageMaker notebook instance lifecycle configuration: %s", name)
				continue
			}

			log.Printf("[INFO] Deleting SageMaker notebook instance lifecycle configuration: %s", name)
			_, err := conn.DeleteNotebookInstanceLifecycleConfig(&sagemaker.DeleteNotebookInstanceLifecycleConfigInput{
				NotebookInstanceLifecycleConfigName: aws.String(name),
			})
			if err != nil {
				log.Printf("[ERROR] Failed to delete SageMaker notebook instance lifecycle configuration %s: %s", name, err)
			}
		}
		return !lastPage
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping SageMaker notebook instance lifecycle configuration sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving SageMaker notebook instance lifecycle configuration: %s", err)
	}

	return nil
}

func TestAccAWSSagemakerLifecycleConfiguration_Basic(t *testing.T) {
	var lifecycleConfig sagemaker.DescribeNotebookInstanceLifecycleConfigOutput
	rName := acctest.RandomWithPrefix(SagemakerLifecycleConfigurationResourcePrefix)
	resourceName := "aws_sagemaker_lifecycle_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerLifecycleConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerLifecycleConfigurationConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerLifecycleConfigurationExists(resourceName, &lifecycleConfig),

					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "on_create", base64Encode([]byte("echo foo"))),
					resource.TestCheckResourceAttr(resourceName, "on_start", base64Encode([]byte("echo bar"))),
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

func TestAccAWSSagemakerLifecycleConfiguration_Update(t *testing.T) {
	var lifecycleConfig sagemaker.DescribeNotebookInstanceLifecycleConfigOutput
	rName := acctest.RandomWithPrefix(SagemakerLifecycleConfigurationResourcePrefix)
	resourceName := "aws_sagemaker_lifecycle_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerLifecycleConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerLifecycleConfigurationConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerLifecycleConfigurationExists(resourceName, &lifecycleConfig),

					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "on_create", base64Encode([]byte("echo foo"))),
					resource.TestCheckResourceAttr(resourceName, "on_start", base64Encode([]byte("echo bar"))),
				),
			},
			{
				Config: testAccSagemakerLifecycleConfigurationConfig_Update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerLifecycleConfigurationExists(resourceName, &lifecycleConfig),

					resource.TestCheckResourceAttr(resourceName, "on_create", base64Encode([]byte("echo bla"))),
					resource.TestCheckResourceAttr(resourceName, "on_start", base64Encode([]byte("echo blub"))),
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

func testAccCheckAWSSagemakerLifecycleConfigurationExists(resourceName string, lifecycleConfig *sagemaker.DescribeNotebookInstanceLifecycleConfigOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn
		output, err := conn.DescribeNotebookInstanceLifecycleConfig(&sagemaker.DescribeNotebookInstanceLifecycleConfigInput{
			NotebookInstanceLifecycleConfigName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("no SageMaker notebook instance lifecycle configuration")
		}

		*lifecycleConfig = *output

		return nil
	}
}

func testAccCheckAWSSagemakerLifecycleConfigurationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_lifecycle_config" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn
		lifecycleConfig, err := conn.DescribeNotebookInstanceLifecycleConfig(&sagemaker.DescribeNotebookInstanceLifecycleConfigInput{
			NotebookInstanceLifecycleConfigName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if isAWSErr(err, "ValidationException", "") {
				return nil
			}
			return err
		}

		if lifecycleConfig != nil && aws.StringValue(lifecycleConfig.NotebookInstanceLifecycleConfigName) == rs.Primary.ID {
			return fmt.Errorf("SageMaker notebook instance lifecycle configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}

	return nil
}

func testAccSagemakerLifecycleConfigurationConfig_Basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_lifecycle_config" "test" {
  name = %q
  on_create = "${base64encode("echo foo")}"
  on_start = "${base64encode("echo bar")}"
}
`, rName)
}

func testAccSagemakerLifecycleConfigurationConfig_Update(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_lifecycle_config" "test" {
  name = %q
  on_create = "${base64encode("echo bla")}"
  on_start = "${base64encode("echo blub")}"
}
`, rName)
}
