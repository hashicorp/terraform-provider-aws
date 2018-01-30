package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const sagemakerTestAccSagemakerNotebookInstanceResourceNamePrefix = "terraform-testacc-"

func init() {
	resource.AddTestSweepers("aws_sagemaker_notebook_instance", &resource.Sweeper{
		Name: "aws_sagemaker_notebook_instance",
	})
}

func TestAccAWSSagemakerNotebookInstance_basic(t *testing.T) {
	var notebook sagemaker.DescribeNotebookInstanceOutput
	notebookName := resource.PrefixedUniqueId(sagemakerTestAccSagemakerNotebookInstanceResourceNamePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerNotebookInstanceConfig(notebookName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerNotebookInstanceExists("aws_sagemaker_notebook_instance.foo", &notebook),
					testAccCheckSagemakerNotebookInstanceName(&notebook, notebookName),

					resource.TestCheckResourceAttr(
						"aws_sagemaker_notebook_instance.foo", "name", notebookName),
				),
			},
		},
	})
}

func TestAccAWSSagemakerNotebookInstance_update(t *testing.T) {
	var notebook sagemaker.DescribeNotebookInstanceOutput
	notebookName := resource.PrefixedUniqueId(sagemakerTestAccSagemakerNotebookInstanceResourceNamePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerNotebookInstanceConfig(notebookName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerNotebookInstanceExists("aws_sagemaker_notebook_instance.foo", &notebook),

					resource.TestCheckResourceAttr(
						"aws_sagemaker_notebook_instance.foo", "instance_type", "ml.t2.medium"),
				),
			},

			{
				Config: testAccSagemakerNotebookInstanceUpdateConfig(notebookName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerNotebookInstanceExists("aws_sagemaker_notebook_instance.foo", &notebook),

					resource.TestCheckResourceAttr(
						"aws_sagemaker_notebook_instance.foo", "instance_type", "ml.m4.xlarge"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerNotebookInstance_tags(t *testing.T) {
	var notebook sagemaker.DescribeNotebookInstanceOutput
	notebookName := resource.PrefixedUniqueId(sagemakerTestAccSagemakerNotebookInstanceResourceNamePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerNotebookInstanceTagsConfig(notebookName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerNotebookInstanceExists("aws_sagemaker_notebook_instance.foo", &notebook),
					testAccCheckSagemakerNotebookInstanceTags(&notebook, "foo", "bar"),

					resource.TestCheckResourceAttr(
						"aws_sagemaker_notebook_instance.foo", "name", notebookName),
					resource.TestCheckResourceAttr("aws_sagemaker_notebook_instance.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_sagemaker_notebook_instance.foo", "tags.foo", "bar"),
				),
			},

			{
				Config: testAccSagemakerNotebookInstanceTagsUpdateConfig(notebookName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerNotebookInstanceExists("aws_sagemaker_notebook_instance.foo", &notebook),
					testAccCheckSagemakerNotebookInstanceTags(&notebook, "foo", ""),
					testAccCheckSagemakerNotebookInstanceTags(&notebook, "bar", "baz"),

					resource.TestCheckResourceAttr("aws_sagemaker_notebook_instance.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_sagemaker_notebook_instance.foo", "tags.bar", "baz"),
				),
			},
		},
	})
}

func testAccCheckSagemakerNotebookInstanceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_notebook_instance" {
			continue
		}

		resp, err := conn.ListNotebookInstances(&sagemaker.ListNotebookInstancesInput{
			NameContains: aws.String(rs.Primary.ID),
		})
		if err == nil {
			if len(resp.NotebookInstances) > 0 {
				return fmt.Errorf("Sagemaker Notebook Instance still exist.")
			}

			return nil
		}

		sagemakerErr, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if sagemakerErr.Code() != "ResourceNotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckSagemakerNotebookInstanceExists(n string, notebook *sagemaker.DescribeNotebookInstanceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker Notebook Instance ID is set")
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

func testAccCheckSagemakerNotebookInstanceName(notebook *sagemaker.DescribeNotebookInstanceOutput, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		notebookName := notebook.NotebookInstanceName
		if *notebookName != expected {
			return fmt.Errorf("Bad Notebook Instance name: %s", *notebook.NotebookInstanceName)
		}

		return nil
	}
}

func testAccCheckSagemakerNotebookInstanceTags(notebook *sagemaker.DescribeNotebookInstanceOutput, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

		ts, err := conn.ListTags(&sagemaker.ListTagsInput{
			ResourceArn: notebook.NotebookInstanceArn,
		})
		if err != nil {
			return fmt.Errorf("Error listing tags: %s", err)
		}

		m := tagsToMapSagemaker(ts.Tags)
		v, ok := m[key]
		if value != "" && !ok {
			return fmt.Errorf("Missing tag: %s", key)
		} else if value == "" && ok {
			return fmt.Errorf("Extra tag: %s", key)
		}
		if value == "" {
			return nil
		}

		if v != value {
			return fmt.Errorf("%s: bad value: %s", key, v)
		}

		return nil
	}
}

func testAccSagemakerNotebookInstanceConfig(notebookName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "foo" {
	name = "%s"
	role_arn = "${aws_iam_role.foo.arn}"
	instance_type = "ml.t2.medium"
}

resource "aws_iam_role" "foo" {
	name = "terraform-testacc-sagemaker-foo"
	path = "/"
	assume_role_policy = "${data.aws_iam_policy_document.assume_role.json}"
}

data "aws_iam_policy_document" "assume_role" {
	statement {
		actions = [ "sts:AssumeRole" ]
		principals {
			type = "Service"
			identifiers = [ "sagemaker.amazonaws.com" ]
		}
	}
}
`, notebookName)
}

func testAccSagemakerNotebookInstanceUpdateConfig(notebookName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "foo" {
	name = "%s"
	role_arn = "${aws_iam_role.foo.arn}"
	instance_type = "ml.m4.xlarge"
}

resource "aws_iam_role" "foo" {
	name = "terraform-testacc-sagemaker-foo"
	path = "/"
	assume_role_policy = "${data.aws_iam_policy_document.assume_role.json}"
}

data "aws_iam_policy_document" "assume_role" {
	statement {
		actions = [ "sts:AssumeRole" ]
		principals {
			type = "Service"
			identifiers = [ "sagemaker.amazonaws.com" ]
		}
	}
}
`, notebookName)
}

func testAccSagemakerNotebookInstanceTagsConfig(notebookName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "foo" {
	name = "%s"
	role_arn = "${aws_iam_role.foo.arn}"
	instance_type = "ml.t2.medium"
	tags {
		foo = "bar"
	}
}

resource "aws_iam_role" "foo" {
	name = "terraform-testacc-sagemaker-foo"
	path = "/"
	assume_role_policy = "${data.aws_iam_policy_document.assume_role.json}"
}

data "aws_iam_policy_document" "assume_role" {
	statement {
		actions = [ "sts:AssumeRole" ]
		principals {
			type = "Service"
			identifiers = [ "sagemaker.amazonaws.com" ]
		}
	}
}
`, notebookName)
}

func testAccSagemakerNotebookInstanceTagsUpdateConfig(notebookName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "foo" {
	name = "%s"
	role_arn = "${aws_iam_role.foo.arn}"
	instance_type = "ml.t2.medium"
	tags {
		bar = "baz"
	}
}

resource "aws_iam_role" "foo" {
	name = "terraform-testacc-sagemaker-foo"
	path = "/"
	assume_role_policy = "${data.aws_iam_policy_document.assume_role.json}"
}

data "aws_iam_policy_document" "assume_role" {
	statement {
		actions = [ "sts:AssumeRole" ]
		principals {
			type = "Service"
			identifiers = [ "sagemaker.amazonaws.com" ]
		}
	}
}
`, notebookName)
}
