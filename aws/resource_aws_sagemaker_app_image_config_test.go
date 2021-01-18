package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/finder"
)

// func init() {
// 	resource.AddTestSweepers("aws_sagemaker_app_image_config", &resource.Sweeper{
// 		Name: "aws_sagemaker_app_image_config",
// 		F:    testSweepSagemakerAppImageConfigs,
// 	})
// }

// func testSweepSagemakerAppImageConfigs(region string) error {
// 	client, err := sharedClientForRegion(region)
// 	if err != nil {
// 		return fmt.Errorf("error getting client: %s", err)
// 	}
// 	conn := client.(*AWSClient).sagemakerconn

// 	err = conn.ListAppImageConfigs(&sagemaker.ListAppImageConfigsInput{}, func(page *sagemaker.ListAppImageConfigsOutput) bool {
// 		for _, instance := range page.AppImageConfigs {
// 			name := aws.StringValue(instance.AppImageConfigName)

// 			input := &sagemaker.DeleteAppImageConfigInput{
// 				AppImageConfigName: instance.AppImageConfigName,
// 			}

// 			log.Printf("[INFO] Deleting SageMaker App Image Config: %s", name)
// 			if _, err := conn.DeleteAppImageConfig(input); err != nil {
// 				log.Printf("[ERROR] Error deleting SageMaker App Image Config (%s): %s", name, err)
// 				continue
// 			}
// 		}
// 	})

// 	if testSweepSkipSweepError(err) {
// 		log.Printf("[WARN] Skipping SageMaker App Image Config sweep for %s: %s", region, err)
// 		return nil
// 	}

// 	if err != nil {
// 		return fmt.Errorf("Error retrieving SageMaker App Image Configs: %w", err)
// 	}

// 	return nil
// }

func TestAccAWSSagemakerAppImageConfig_basic(t *testing.T) {
	var notebook sagemaker.DescribeAppImageConfigOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_app_image_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerAppImageConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerAppImageConfigBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerAppImageConfigExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("app-image-config/%s", rName)),
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

func TestAccAWSSagemakerAppImageConfig_gitConfig_branch(t *testing.T) {
	var notebook sagemaker.DescribeAppImageConfigOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_app_image_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerAppImageConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerAppImageConfigGitConfigBranchConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerAppImageConfigExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("code-repository/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "git_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "git_config.0.repository_url", "https://github.com/hashicorp/terraform-provider-aws.git"),
					resource.TestCheckResourceAttr(resourceName, "git_config.0.branch", "master"),
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

func TestAccAWSSagemakerAppImageConfig_disappears(t *testing.T) {
	var notebook sagemaker.DescribeAppImageConfigOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_app_image_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerAppImageConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerAppImageConfigBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerAppImageConfigExists(resourceName, &notebook),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSagemakerAppImageConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSagemakerAppImageConfigDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_app_image_config" {
			continue
		}

		codeRepository, err := finder.AppImageConfigByName(conn, rs.Primary.ID)
		if err != nil {
			return nil
		}

		if aws.StringValue(codeRepository.AppImageConfigName) == rs.Primary.ID {
			return fmt.Errorf("Sagemaker App Image Config %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSSagemakerAppImageConfigExists(n string, codeRepo *sagemaker.DescribeAppImageConfigOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker App Image Config ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn
		resp, err := finder.AppImageConfigByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*codeRepo = *resp

		return nil
	}
}

func testAccAWSSagemakerAppImageConfigBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q
}
`, rName)
}

func testAccAWSSagemakerAppImageConfigGitConfigBranchConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q

  git_config {
    repository_url = "https://github.com/hashicorp/terraform-provider-aws.git"
    branch         = "master"
  }
}
`, rName)
}

func testAccAWSSagemakerAppImageConfigGitConfigSecretConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = jsonencode({ username = "example", password = "example" })
}

resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q

  git_config {
    repository_url = "https://github.com/hashicorp/terraform-provider-aws.git"
    secret_arn     = aws_secretsmanager_secret.test.arn
  }

  depends_on = [aws_secretsmanager_secret_version.test]
}
`, rName)
}

func testAccAWSSagemakerAppImageConfigGitConfigSecretUpdatedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test2" {
  name = "%[1]s-2"
}

resource "aws_secretsmanager_secret_version" "test2" {
  secret_id     = aws_secretsmanager_secret.test2.id
  secret_string = jsonencode({ username = "example", password = "example" })
}

resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q

  git_config {
    repository_url = "https://github.com/hashicorp/terraform-provider-aws.git"
    secret_arn     = aws_secretsmanager_secret.test2.arn
  }

  depends_on = [aws_secretsmanager_secret_version.test2]
}
`, rName)
}
