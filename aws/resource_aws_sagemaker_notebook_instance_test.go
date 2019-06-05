package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/acctest"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSSagemakerNotebookInstance_basic(t *testing.T) {
	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_notebook_instance.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerNotebookInstanceConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),

					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "ml.t2.medium"),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
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

//"accelerator_types": {
//Type:     schema.TypeList,
//Optional: true,
//Elem:     &schema.Schema{Type: schema.TypeString},
//},
//
//"additional_code_repositories": {
//Type:     schema.TypeList,
//Optional: true,
//Elem:     &schema.Schema{Type: schema.TypeString},
//},
//
//"default_code_repository": {
//Type:     schema.TypeString,
//Optional: true,
//// TODO min 1
//},

func TestAccAWSSagemakerNotebookInstance_DirectInternetAccess(t *testing.T) {
	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_notebook_instance.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerNotebookInstanceConfig_DirectInternetAccess(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),

					resource.TestCheckResourceAttr(resourceName, "direct_internet_access", "Disabled"),
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

func TestAccAWSSagemakerNotebookInstance_LifeCycleConfigName(t *testing.T) {
	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_notebook_instance.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerNotebookInstanceConfig_LifecycleConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),

					resource.TestCheckResourceAttr(resourceName, "lifecycle_config_name", rName),
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

func TestAccAWSSagemakerNotebookInstance_volumeSize(t *testing.T) {
	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_notebook_instance.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerNotebookInstanceConfig_VolumeSize(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),

					resource.TestCheckResourceAttr(resourceName, "volume_size", "7"),
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

func TestAccAWSSagemakerNotebookInstance_KmsKeyID(t *testing.T) {
	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_notebook_instance.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerNotebookInstanceConfig_KmsKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),

					resource.TestCheckResourceAttr(resourceName, "kms_key_id", "7"),
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

func TestAccAWSSagemakerNotebookInstance_updateInstanceType(t *testing.T) {
	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_notebook_instance.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerNotebookInstanceConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),

					resource.TestCheckResourceAttr(resourceName, "instance_type", "ml.t2.medium"),
				),
			},

			{
				Config: testAccAWSSagemakerNotebookInstanceUpdateInstanceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),

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

func TestAccAWSSagemakerNotebookInstance_LifecycleConfigName(t *testing.T) {
	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := resource.PrefixedUniqueId(sagemakerTestAccSagemakerNotebookInstanceResourceNamePrefix)
	resourceName := "aws_sagemaker_notebook_instance.test"
	sagemakerLifecycleConfigResourceName := "aws_sagemaker_notebook_instance_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerNotebookInstanceConfigLifecycleConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttrPair(resourceName, "lifecycle_config_name", sagemakerLifecycleConfigResourceName, "name"),
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

func TestAccAWSSagemakerNotebookInstance_tags(t *testing.T) {
	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_notebook_instance.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerNotebookInstanceTagsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),

					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
				),
			},

			{
				Config: testAccAWSSagemakerNotebookInstanceTagsUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),

					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.bar", "baz"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerNotebookInstance_disappears(t *testing.T) {
	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_notebook_instance.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerNotebookInstanceConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					testAccCheckAWSSagemakerNotebookInstanceDisappears(&notebook),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSagemakerNotebookInstanceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_notebook_instance" {
			continue
		}

		describeNotebookInput := &sagemaker.DescribeNotebookInstanceInput{
			NotebookInstanceName: aws.String(rs.Primary.ID),
		}
		notebookInstance, err := conn.DescribeNotebookInstance(describeNotebookInput)
		if err != nil {
			if isAWSErr(err, "ValidationException", "RecordNotFound") {
				return nil
			}
			return err
		}

		if *notebookInstance.NotebookInstanceName == rs.Primary.ID {
			return fmt.Errorf("sagemaker notebook instance %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSSagemakerNotebookInstanceExists(n string, notebook *sagemaker.DescribeNotebookInstanceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no sagmaker Notebook Instance ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn
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

func testAccCheckAWSSagemakerNotebookInstanceDisappears(instance *sagemaker.DescribeNotebookInstanceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

		if *instance.NotebookInstanceStatus != sagemaker.NotebookInstanceStatusFailed && *instance.NotebookInstanceStatus != sagemaker.NotebookInstanceStatusStopped {
			if err := stopSagemakerNotebookInstance(conn, *instance.NotebookInstanceName); err != nil {
				return err
			}
		}

		deleteOpts := &sagemaker.DeleteNotebookInstanceInput{
			NotebookInstanceName: instance.NotebookInstanceName,
		}

		if _, err := conn.DeleteNotebookInstance(deleteOpts); err != nil {
			return fmt.Errorf("error trying to delete sagemaker notebook instance (%s): %s", aws.StringValue(instance.NotebookInstanceName), err)
		}

		stateConf := &resource.StateChangeConf{
			Pending: []string{
				sagemaker.NotebookInstanceStatusDeleting,
			},
			Target:  []string{""},
			Refresh: sagemakerNotebookInstanceStateRefreshFunc(conn, *instance.NotebookInstanceName),
			Timeout: 10 * time.Minute,
		}
		_, err := stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("error waiting for sagemaker notebook instance (%s) to delete: %s", aws.StringValue(instance.NotebookInstanceName), err)
		}

		return nil
	}
}

func testagemakerNotebookInstanceConfig_Base(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "foo" {
  name          = "%s"
  role_arn      = "${aws_iam_role.foo.arn}"
  instance_type = "ml.t2.medium"
}

resource "aws_iam_role" "foo" {
  name               = "%s"
  path               = "/"
  assume_role_policy = "${data.aws_iam_policy_document.assume_role.json}"
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
`, rName)
}

func testAccAWSSagemakerNotebookInstanceConfig_Basic(notebookName string) string {
	return testagemakerNotebookInstanceConfig_Base(notebookName) + fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "foo" {
  name          = "%s"
  role_arn      = "${aws_iam_role.foo.arn}"
  instance_type = "ml.m4.xlarge"
}

resource "aws_iam_role" "foo" {
  name               = "%s"
  path               = "/"
  assume_role_policy = "${data.aws_iam_policy_document.assume_role.json}"
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
`, notebookName)
}

func testAccAWSSagemakerNotebookInstanceConfigLifecycleConfigName(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      identifiers = ["sagemaker.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_iam_role" "test" {
  assume_role_policy = "${data.aws_iam_policy_document.assume_role.json}"
  name               = %[1]q
  path               = "/"
}

resource "aws_sagemaker_notebook_instance_lifecycle_configuration" "test" {
  name = %[1]q
}

resource "aws_sagemaker_notebook_instance" "test" {
  instance_type         = "ml.t2.medium"
  lifecycle_config_name = "${aws_sagemaker_notebook_instance_lifecycle_configuration.test.name}"
  name                  = %[1]q
  role_arn              = "${aws_iam_role.test.arn}"
}
`, rName)
}

func testAccAWSSagemakerNotebookInstanceConfig_DirectInternetAccess(notebookName string) string {
	return testagemakerNotebookInstanceConfig_Base(notebookName) + fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "foo" {
  name          = "%s"
  role_arn      = "${aws_iam_role.foo.arn}"
  instance_type = "ml.t2.medium"

  tags = {
    foo = "bar"
  }
}

resource "aws_iam_role" "foo" {
  name               = "%s"
  path               = "/"
  assume_role_policy = "${data.aws_iam_policy_document.assume_role.json}"
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
`, notebookName, notebookName)
}

func testAccAWSSagemakerNotebookInstanceUpdateInstanceConfig(notebookName string) string {
	return testagemakerNotebookInstanceConfig_Base(notebookName) + fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "foo" {
  name          = "%s"
  role_arn      = "${aws_iam_role.foo.arn}"
  instance_type = "ml.t2.medium"

  tags = {
    bar = "baz"
  }
}

resource "aws_iam_role" "foo" {
  name               = "%s"
  path               = "/"
  assume_role_policy = "${data.aws_iam_policy_document.assume_role.json}"
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
`, notebookName)
}
