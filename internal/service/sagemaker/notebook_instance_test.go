package sagemaker_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
)

func TestAccSageMakerNotebookInstance_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "ml.t2.medium"),
					resource.TestCheckResourceAttr(resourceName, "platform_identifier", "notebook-al1-v1"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "direct_internet_access", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "root_access", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "volume_size", "5"),
					resource.TestCheckResourceAttr(resourceName, "default_code_repository", ""),
					resource.TestCheckResourceAttr(resourceName, "additional_code_repositories.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
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

func TestAccSageMakerNotebookInstance_update(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "ml.t2.medium"),
				),
			},

			{
				Config: testAccNotebookInstanceUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "ml.m4.xlarge"),
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

func TestAccSageMakerNotebookInstance_volumeSize(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook1, notebook2, notebook3 sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var resourceName = "aws_sagemaker_notebook_instance.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook1),
					resource.TestCheckResourceAttr(resourceName, "volume_size", "5"),
				),
			},
			{
				Config: testAccNotebookInstanceVolumeConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook2),
					resource.TestCheckResourceAttr(resourceName, "volume_size", "8"),
					testAccCheckNotebookInstanceNotRecreated(&notebook1, &notebook2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotebookInstanceBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook3),
					resource.TestCheckResourceAttr(resourceName, "volume_size", "5"),
					testAccCheckNotebookInstanceRecreated(&notebook2, &notebook3),
				),
			},
		},
	})
}

func TestAccSageMakerNotebookInstance_lifecycleName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_notebook_instance.test"
	sagemakerLifecycleConfigResourceName := "aws_sagemaker_notebook_instance_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceLifecycleNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttrPair(resourceName, "lifecycle_config_name", sagemakerLifecycleConfigResourceName, "name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotebookInstanceBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_config_name", ""),
				),
			},
			{
				Config: testAccNotebookInstanceLifecycleNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttrPair(resourceName, "lifecycle_config_name", sagemakerLifecycleConfigResourceName, "name"),
				),
			},
		},
	})
}

func TestAccSageMakerNotebookInstance_tags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook),
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
				Config: testAccNotebookInstanceTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccNotebookInstanceTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccSageMakerNotebookInstance_kms(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceKMSConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", "aws_kms_key.test", "id"),
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

func TestAccSageMakerNotebookInstance_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook),
					acctest.CheckResourceDisappears(acctest.Provider, tfsagemaker.ResourceNotebookInstance(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckNotebookInstanceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_notebook_instance" {
			continue
		}

		describeNotebookInput := &sagemaker.DescribeNotebookInstanceInput{
			NotebookInstanceName: aws.String(rs.Primary.ID),
		}
		notebookInstance, err := conn.DescribeNotebookInstance(describeNotebookInput)

		if tfawserr.ErrMessageContains(err, tfsagemaker.ErrCodeValidationException, "RecordNotFound") {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading SageMaker Notebook Instance (%s): %w", rs.Primary.ID, err)
		}

		if aws.StringValue(notebookInstance.NotebookInstanceName) == rs.Primary.ID {
			return fmt.Errorf("sagemaker notebook instance %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckNotebookInstanceExists(n string, notebook *sagemaker.DescribeNotebookInstanceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker Notebook Instance ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn
		opts := &sagemaker.DescribeNotebookInstanceInput{
			NotebookInstanceName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeNotebookInstance(opts)
		if err != nil {
			return err
		}

		*notebook = *resp

		return nil
	}
}

func testAccCheckNotebookInstanceNotRecreated(i, j *sagemaker.DescribeNotebookInstanceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.CreationTime).Equal(aws.TimeValue(j.CreationTime)) {
			return errors.New("SageMaker Notebook Instance was recreated")
		}

		return nil
	}
}

func testAccCheckNotebookInstanceRecreated(i, j *sagemaker.DescribeNotebookInstanceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreationTime).Equal(aws.TimeValue(j.CreationTime)) {
			return errors.New("SageMaker Notebook Instance was not recreated")
		}

		return nil
	}
}

func TestAccSageMakerNotebookInstance_Root_access(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceRootAccessConfig(rName, "Disabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "root_access", "Disabled"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotebookInstanceRootAccessConfig(rName, "Enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "root_access", "Enabled"),
				),
			},
		},
	})
}

func TestAccSageMakerNotebookInstance_Platform_identifier(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstancePlatformIdentifierConfig(rName, "notebook-al2-v1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "platform_identifier", "notebook-al2-v1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotebookInstancePlatformIdentifierConfig(rName, "notebook-al1-v1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "platform_identifier", "notebook-al1-v1"),
				),
			},
		},
	})
}

func TestAccSageMakerNotebookInstance_DirectInternet_access(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceDirectInternetAccessConfig(rName, "Disabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "direct_internet_access", "Disabled"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", "aws_subnet.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "network_interface_id", regexp.MustCompile("eni-.*")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotebookInstanceDirectInternetAccessConfig(rName, "Enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "direct_internet_access", "Enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", "aws_subnet.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "network_interface_id", regexp.MustCompile("eni-.*")),
				),
			},
		},
	})
}

func TestAccSageMakerNotebookInstance_DefaultCode_repository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var resourceName = "aws_sagemaker_notebook_instance.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceDefaultCodeRepositoryConfig(rName, "https://github.com/hashicorp/terraform-provider-aws.git"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "default_code_repository", "https://github.com/hashicorp/terraform-provider-aws.git"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotebookInstanceBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "default_code_repository", ""),
				),
			},
			{
				Config: testAccNotebookInstanceDefaultCodeRepositoryConfig(rName, "https://github.com/hashicorp/terraform-provider-aws.git"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "default_code_repository", "https://github.com/hashicorp/terraform-provider-aws.git"),
				),
			},
		},
	})
}

func TestAccSageMakerNotebookInstance_AdditionalCode_repositories(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var resourceName = "aws_sagemaker_notebook_instance.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceAdditionalCodeRepository1Config(rName, "https://github.com/hashicorp/terraform-provider-aws.git"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "additional_code_repositories.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "additional_code_repositories.*", "https://github.com/hashicorp/terraform-provider-aws.git"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotebookInstanceBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "additional_code_repositories.#", "0"),
				),
			},
			{
				Config: testAccNotebookInstanceAdditionalCodeRepository2Config(rName, "https://github.com/hashicorp/terraform-provider-aws.git", "https://github.com/hashicorp/terraform.git"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "additional_code_repositories.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "additional_code_repositories.*", "https://github.com/hashicorp/terraform-provider-aws.git"),
					resource.TestCheckTypeSetElemAttr(resourceName, "additional_code_repositories.*", "https://github.com/hashicorp/terraform.git"),
				),
			},
			{
				Config: testAccNotebookInstanceAdditionalCodeRepository1Config(rName, "https://github.com/hashicorp/terraform-provider-aws.git"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "additional_code_repositories.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "additional_code_repositories.*", "https://github.com/hashicorp/terraform-provider-aws.git"),
				),
			},
		},
	})
}

func TestAccSageMakerNotebookInstance_DefaultCodeRepository_sageMakerRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var resourceName = "aws_sagemaker_notebook_instance.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceDefaultCodeRepositoryRepoConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttrPair(resourceName, "default_code_repository", "aws_sagemaker_code_repository.test", "code_repository_name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotebookInstanceBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "default_code_repository", ""),
				),
			},
			{
				Config: testAccNotebookInstanceDefaultCodeRepositoryRepoConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttrPair(resourceName, "default_code_repository", "aws_sagemaker_code_repository.test", "code_repository_name")),
			},
		},
	})
}

func testAccNotebookInstanceBaseConfig(rName string) string {
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

func testAccNotebookInstanceBasicConfig(rName string) string {
	return testAccNotebookInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name          = %[1]q
  role_arn      = aws_iam_role.test.arn
  instance_type = "ml.t2.medium"
}
`, rName)
}

func testAccNotebookInstanceUpdateConfig(rName string) string {
	return testAccNotebookInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name          = %[1]q
  role_arn      = aws_iam_role.test.arn
  instance_type = "ml.m4.xlarge"
}
`, rName)
}

func testAccNotebookInstanceLifecycleNameConfig(rName string) string {
	return testAccNotebookInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance_lifecycle_configuration" "test" {
  name = %[1]q
}

resource "aws_sagemaker_notebook_instance" "test" {
  instance_type         = "ml.t2.medium"
  lifecycle_config_name = aws_sagemaker_notebook_instance_lifecycle_configuration.test.name
  name                  = %[1]q
  role_arn              = aws_iam_role.test.arn
}
`, rName)
}

func testAccNotebookInstanceTags1Config(rName, tagKey1, tagValue1 string) string {
	return testAccNotebookInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name          = %[1]q
  role_arn      = aws_iam_role.test.arn
  instance_type = "ml.t2.medium"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccNotebookInstanceTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccNotebookInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name          = %[1]q
  role_arn      = aws_iam_role.test.arn
  instance_type = "ml.t2.medium"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccNotebookInstanceRootAccessConfig(rName string, rootAccess string) string {
	return testAccNotebookInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name          = %[1]q
  role_arn      = aws_iam_role.test.arn
  instance_type = "ml.t2.medium"
  root_access   = %[2]q
}
`, rName, rootAccess)
}

func testAccNotebookInstancePlatformIdentifierConfig(rName string, platformIdentifier string) string {
	return testAccNotebookInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name                = %[1]q
  role_arn            = aws_iam_role.test.arn
  instance_type       = "ml.t2.medium"
  platform_identifier = %[2]q
}
`, rName, platformIdentifier)
}
func testAccNotebookInstanceDirectInternetAccessConfig(rName string, directInternetAccess string) string {
	return testAccNotebookInstanceBaseConfig(rName) +
		fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name                   = %[1]q
  role_arn               = aws_iam_role.test.arn
  instance_type          = "ml.t2.medium"
  security_groups        = [aws_security_group.test.id]
  subnet_id              = aws_subnet.test.id
  direct_internet_access = %[2]q
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.0.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, directInternetAccess)
}

func testAccNotebookInstanceVolumeConfig(rName string) string {
	return testAccNotebookInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name          = %[1]q
  role_arn      = aws_iam_role.test.arn
  instance_type = "ml.t2.medium"
  volume_size   = 8
}
  `, rName)
}

func testAccNotebookInstanceDefaultCodeRepositoryConfig(rName string, defaultCodeRepository string) string {
	return testAccNotebookInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name                    = %[1]q
  role_arn                = aws_iam_role.test.arn
  instance_type           = "ml.t2.medium"
  default_code_repository = %[2]q
}
`, rName, defaultCodeRepository)
}

func testAccNotebookInstanceAdditionalCodeRepository1Config(rName, repo1 string) string {
	return testAccNotebookInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name                         = %[1]q
  role_arn                     = aws_iam_role.test.arn
  instance_type                = "ml.t2.medium"
  additional_code_repositories = ["%[2]s"]
}
`, rName, repo1)
}

func testAccNotebookInstanceAdditionalCodeRepository2Config(rName, repo1, repo2 string) string {
	return testAccNotebookInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name                         = %[1]q
  role_arn                     = aws_iam_role.test.arn
  instance_type                = "ml.t2.medium"
  additional_code_repositories = ["%[2]s", "%[3]s"]
}
`, rName, repo1, repo2)
}

func testAccNotebookInstanceDefaultCodeRepositoryRepoConfig(rName string) string {
	return testAccNotebookInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_code_repository" "test" {
  code_repository_name = %[1]q

  git_config {
    repository_url = "https://github.com/hashicorp/terraform-provider-aws.git"
  }
}

resource "aws_sagemaker_notebook_instance" "test" {
  name                    = %[1]q
  role_arn                = aws_iam_role.test.arn
  instance_type           = "ml.t2.medium"
  default_code_repository = aws_sagemaker_code_repository.test.code_repository_name
}
`, rName)
}

func testAccNotebookInstanceKMSConfig(rName string) string {
	return testAccNotebookInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = "Terraform acc test %[1]s"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_sagemaker_notebook_instance" "test" {
  name          = %[1]q
  role_arn      = aws_iam_role.test.arn
  instance_type = "ml.t2.medium"
  kms_key_id    = aws_kms_key.test.id
}
`, rName)
}
