package aws

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const SagemakerNotebookInstanceLifecycleConfigurationResourcePrefix = "tf-acc-test"

func init() {
	resource.AddTestSweepers("aws_sagemaker_notebook_instance_lifecycle_configuration", &resource.Sweeper{
		Name: "aws_sagemaker_notebook_instance_lifecycle_configuration",
		F:    testSweepSagemakerNotebookInstanceLifecycleConfiguration,
		Dependencies: []string{
			"aws_sagemaker_notebook_instance",
		},
	})
}

func testSweepSagemakerNotebookInstanceLifecycleConfiguration(region string) error {
	client, err := acctest.SharedRegionalSweeperClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).SageMakerConn

	input := &sagemaker.ListNotebookInstanceLifecycleConfigsInput{}
	err = conn.ListNotebookInstanceLifecycleConfigsPages(input, func(page *sagemaker.ListNotebookInstanceLifecycleConfigsOutput, lastPage bool) bool {
		if len(page.NotebookInstanceLifecycleConfigs) == 0 {
			log.Printf("[INFO] No SageMaker Notebook Instance Lifecycle Configuration to sweep")
			return false
		}
		for _, lifecycleConfig := range page.NotebookInstanceLifecycleConfigs {
			name := aws.StringValue(lifecycleConfig.NotebookInstanceLifecycleConfigName)
			if !strings.HasPrefix(name, SagemakerNotebookInstanceLifecycleConfigurationResourcePrefix) {
				log.Printf("[INFO] Skipping SageMaker Notebook Instance Lifecycle Configuration: %s", name)
				continue
			}

			log.Printf("[INFO] Deleting SageMaker Notebook Instance Lifecycle Configuration: %s", name)
			_, err := conn.DeleteNotebookInstanceLifecycleConfig(&sagemaker.DeleteNotebookInstanceLifecycleConfigInput{
				NotebookInstanceLifecycleConfigName: aws.String(name),
			})
			if err != nil {
				log.Printf("[ERROR] Failed to delete SageMaker Notebook Instance Lifecycle Configuration %s: %s", name, err)
			}
		}
		return !lastPage
	})
	if err != nil {
		if acctest.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SageMaker Notebook Instance Lifecycle Configuration sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving SageMaker Notebook Instance Lifecycle Configuration: %s", err)
	}

	return nil
}

func TestAccAWSSagemakerNotebookInstanceLifecycleConfiguration_basic(t *testing.T) {
	var lifecycleConfig sagemaker.DescribeNotebookInstanceLifecycleConfigOutput
	rName := sdkacctest.RandomWithPrefix(SagemakerNotebookInstanceLifecycleConfigurationResourcePrefix)
	resourceName := "aws_sagemaker_notebook_instance_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSSagemakerNotebookInstanceLifecycleConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerNotebookInstanceLifecycleConfigurationConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceLifecycleConfigurationExists(resourceName, &lifecycleConfig),

					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckNoResourceAttr(resourceName, "on_create"),
					resource.TestCheckNoResourceAttr(resourceName, "on_start"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("notebook-instance-lifecycle-config/%s", rName)),
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

func TestAccAWSSagemakerNotebookInstanceLifecycleConfiguration_Update(t *testing.T) {
	var lifecycleConfig sagemaker.DescribeNotebookInstanceLifecycleConfigOutput
	rName := sdkacctest.RandomWithPrefix(SagemakerNotebookInstanceLifecycleConfigurationResourcePrefix)
	resourceName := "aws_sagemaker_notebook_instance_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSSagemakerNotebookInstanceLifecycleConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerNotebookInstanceLifecycleConfigurationConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceLifecycleConfigurationExists(resourceName, &lifecycleConfig),

					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				Config: testAccSagemakerNotebookInstanceLifecycleConfigurationConfig_Update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceLifecycleConfigurationExists(resourceName, &lifecycleConfig),

					resource.TestCheckResourceAttr(resourceName, "on_create", verify.Base64Encode([]byte("echo bla"))),
					resource.TestCheckResourceAttr(resourceName, "on_start", verify.Base64Encode([]byte("echo blub"))),
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

func testAccCheckAWSSagemakerNotebookInstanceLifecycleConfigurationExists(resourceName string, lifecycleConfig *sagemaker.DescribeNotebookInstanceLifecycleConfigOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn
		output, err := conn.DescribeNotebookInstanceLifecycleConfig(&sagemaker.DescribeNotebookInstanceLifecycleConfigInput{
			NotebookInstanceLifecycleConfigName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("no SageMaker Notebook Instance Lifecycle Configuration")
		}

		*lifecycleConfig = *output

		return nil
	}
}

func testAccCheckAWSSagemakerNotebookInstanceLifecycleConfigurationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_notebook_instance_lifecycle_configuration" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn
		lifecycleConfig, err := conn.DescribeNotebookInstanceLifecycleConfig(&sagemaker.DescribeNotebookInstanceLifecycleConfigInput{
			NotebookInstanceLifecycleConfigName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if tfawserr.ErrMessageContains(err, "ValidationException", "") {
				continue
			}
			return err
		}

		if lifecycleConfig != nil && aws.StringValue(lifecycleConfig.NotebookInstanceLifecycleConfigName) == rs.Primary.ID {
			return fmt.Errorf("SageMaker Notebook Instance Lifecycle Configuration %s still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccSagemakerNotebookInstanceLifecycleConfigurationConfig_Basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance_lifecycle_configuration" "test" {
  name = %q
}
`, rName)
}

func testAccSagemakerNotebookInstanceLifecycleConfigurationConfig_Update(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance_lifecycle_configuration" "test" {
  name      = %q
  on_create = base64encode("echo bla")
  on_start  = base64encode("echo blub")
}
`, rName)
}
