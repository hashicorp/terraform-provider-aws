package aws

import (
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_sagemaker_model", &resource.Sweeper{
		Name: "aws_sagemaker_model",
		F:    testSweepSagemakerModels,
	})
}

func testSweepSagemakerModels(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).sagemakerconn

	req := &sagemaker.ListModelsInput{
		NameContains: aws.String("terraform-testacc-sagemaker-model"),
	}
	resp, err := conn.ListModels(req)
	if err != nil {
		return fmt.Errorf("Error listing models: %s", err)
	}

	if len(resp.Models) == 0 {
		log.Print("[DEBUG] No sagemaker models to sweep")
		return nil
	}

	for _, model := range resp.Models {
		_, err := conn.DeleteModel(&sagemaker.DeleteModelInput{
			ModelName: model.ModelName,
		})
		if err != nil {
			return fmt.Errorf(
				"Error deleting sagemaker model (%s): %s",
				*model.ModelName, err)
		}
	}

	return nil
}

func TestAccAWSSagemakerModel_importBasic(t *testing.T) {
	resourceName := "aws_sagemaker_model.foo"
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerModelConfig(rName),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSagemakerModel_basic(t *testing.T) {
	var model sagemaker.DescribeModelOutput
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerModelConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists("aws_sagemaker_model.foo", &model),

					resource.TestCheckResourceAttr(
						"aws_sagemaker_model.foo", "name", "terraform-testacc-sagemaker-model-foo"),
					resource.TestCheckResourceAttr("aws_sagemaker_model.foo", "primary_container.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_model.foo", "primary_container.0.image",
						"174872318107.dkr.ecr.us-west-2.amazonaws.com/kmeans:1"),
					resource.TestCheckResourceAttrSet("aws_sagemaker_model.foo", "execution_role_arn"),
					resource.TestCheckResourceAttrSet("aws_sagemaker_model.foo", "arn"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerModel_tags(t *testing.T) {
	var model sagemaker.DescribeModelOutput
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerModelConfigTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists("aws_sagemaker_model.foo", &model),
					testAccCheckSagemakerTags(&model, "foo", "bar"),

					resource.TestCheckResourceAttr(
						"aws_sagemaker_model.foo", "name", "terraform-testacc-sagemaker-model-foo"),
					resource.TestCheckResourceAttr("aws_sagemaker_model.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_sagemaker_model.foo", "tags.foo", "bar"),
				),
			},
			{
				Config: testAccSagemakerModelConfigTagsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists("aws_sagemaker_model.foo", &model),
					testAccCheckSagemakerTags(&model, "foo", ""),
					testAccCheckSagemakerTags(&model, "bar", "baz"),

					resource.TestCheckResourceAttr("aws_sagemaker_model.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_sagemaker_model.foo", "tags.bar", "baz"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerModel_primaryContainerModelDataUrl(t *testing.T) {
	var model sagemaker.DescribeModelOutput
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerPrimaryContainerModelDataUrlConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists("aws_sagemaker_model.foo", &model),
					testAccCheckSagemakerModelPrimaryContainerModelDataUrl(&model,
						"https://s3-us-west-2.amazonaws.com/terraform-testacc-sagemaker-model-data-bucket/model.tar.gz"),

					resource.TestCheckResourceAttr("aws_sagemaker_model.foo",
						"primary_container.0.model_data_url",
						"https://s3-us-west-2.amazonaws.com/terraform-testacc-sagemaker-model-data-bucket/model.tar.gz"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerModel_primaryContainerHostname(t *testing.T) {
	var model sagemaker.DescribeModelOutput
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerPrimaryContainerHostnameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists("aws_sagemaker_model.foo", &model),
					testAccCheckSagemakerModelPrimaryContainerHostname(&model, "primary-hostname-foo"),

					resource.TestCheckResourceAttr("aws_sagemaker_model.foo",
						"primary_container.0.container_hostname", "primary-hostname-foo"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerModel_primaryContainerEnvironment(t *testing.T) {
	var model sagemaker.DescribeModelOutput
	environment := map[string]*string{
		"foo": aws.String("bar"),
	}
	rName := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerPrimaryContainerEnvironmentConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists("aws_sagemaker_model.foo", &model),
					testAccCheckSagemakerModelPrimaryContainerEnvironment(&model, environment),

					resource.TestCheckResourceAttr("aws_sagemaker_model.foo",
						"primary_container.0.environment.%", "1"),
					resource.TestCheckResourceAttr("aws_sagemaker_model.foo",
						"primary_container.0.environment.foo", "bar"),
				),
			},
		},
	})
}

func testAccCheckSagemakerModelDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_model" {
			continue
		}

		resp, err := conn.ListModels(&sagemaker.ListModelsInput{
			NameContains: aws.String(rs.Primary.ID),
		})
		if err == nil {
			if len(resp.Models) > 0 {
				return fmt.Errorf("Sagemaker models still exist.")
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

func testAccCheckSagemakerModelExists(n string, model *sagemaker.DescribeModelOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker model ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn
		DescribeModelOpts := &sagemaker.DescribeModelInput{
			ModelName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeModel(DescribeModelOpts)
		if err != nil {
			return err
		}

		*model = *resp
		return nil
	}
}

func testAccCheckSagemakerTags(model *sagemaker.DescribeModelOutput, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

		ts, err := conn.ListTags(&sagemaker.ListTagsInput{
			ResourceArn: model.ModelArn,
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

func testAccCheckSagemakerModelPrimaryContainerModelDataUrl(model *sagemaker.DescribeModelOutput, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		modelDataUrl := model.PrimaryContainer.ModelDataUrl
		if *modelDataUrl != expected {
			return fmt.Errorf("Bad model data url of primary container: %s", *modelDataUrl)
		}

		return nil
	}
}

func testAccCheckSagemakerModelPrimaryContainerHostname(model *sagemaker.DescribeModelOutput, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		hostname := model.PrimaryContainer.ContainerHostname
		if *hostname != expected {
			return fmt.Errorf("Bad container hostname of primary container: %s", *hostname)
		}

		return nil
	}
}

func testAccCheckSagemakerModelPrimaryContainerEnvironment(model *sagemaker.DescribeModelOutput, expected map[string]*string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		environment := model.PrimaryContainer.Environment

		if !reflect.DeepEqual(environment, expected) {
			return fmt.Errorf("Bad environment of primary container.")
		}

		return nil
	}
}

func testAccSagemakerModelConfig(rName string) string {
	return fmt.Sprintf(`
esource "aws_sagemaker_model" "foo" {
	name = "terraform-testacc-sagemaker-model-%s"
	execution_role_arn = "${aws_iam_role.foo.arn}"

	primary_container {
		image = "174872318107.dkr.ecr.us-west-2.amazonaws.com/kmeans:1"
	}
}

resource "aws_iam_role" "foo" {
  name = "terraform-testacc-sagemaker-model-%s"
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
`, rName, rName)
}

func testAccSagemakerModelConfigTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_model" "foo" {
	name = "terraform-testacc-sagemaker-model-%s"
	execution_role_arn = "${aws_iam_role.foo.arn}"

	primary_container {
		image = "174872318107.dkr.ecr.us-west-2.amazonaws.com/kmeans:1"
	}

	tags {
		foo = "bar"
	}
}

resource "aws_iam_role" "foo" {
  name = "terraform-testacc-sagemaker-model-%s"
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
`, rName, rName)
}

func testAccSagemakerModelConfigTagsUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_model" "foo" {
	name = "terraform-testacc-sagemaker-model-%s"
	execution_role_arn = "${aws_iam_role.foo.arn}"

	primary_container {
		image = "174872318107.dkr.ecr.us-west-2.amazonaws.com/kmeans:1"
	}

	tags {
		bar = "baz"
	}
}

resource "aws_iam_role" "foo" {
  name = "terraform-testacc-sagemaker-model-%s"
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
`, rName, rName)
}

func testAccSagemakerPrimaryContainerModelDataUrlConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_model" "foo" {
	name = "terraform-testacc-sagemaker-model-%s"
	execution_role_arn = "${aws_iam_role.foo.arn}"

	primary_container {
		image = "174872318107.dkr.ecr.us-west-2.amazonaws.com/kmeans:1"
		model_data_url = "https://s3-us-west-2.amazonaws.com/${aws_s3_bucket.foo.bucket}/model.tar.gz"
	}
}

resource "aws_iam_role" "foo" {
  name = "terraform-testacc-sagemaker-model-%s"
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

resource "aws_iam_policy" "foo" {
  name = "terraform-testacc-sagemaker-model-%s"
  description = "Allow Sagemaker to create model"
  policy = "${data.aws_iam_policy_document.foo.json}"
}

data "aws_iam_policy_document" "foo" {
  statement {
    effect = "Allow"
    actions = [
      "cloudwatch:PutMetricData",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:CreateLogGroup",
      "logs:DescribeLogStreams",
      "ecr:GetAuthorizationToken",
      "ecr:BatchCheckLayerAvailability",
      "ecr:GetDownloadUrlForLayer",
      "ecr:BatchGetImage"
    ]
    resources = [
      "*"]
  }
  statement {
    effect = "Allow"
    actions = [
      "s3:GetObject"
    ]
    resources = [
      "arn:aws:s3:::${aws_s3_bucket.foo.bucket}",
      "arn:aws:s3:::${aws_s3_bucket.foo.bucket}/*"
    ]
  }
}

resource "aws_iam_role_policy_attachment" "foo" {
  role = "${aws_iam_role.foo.name}"
  policy_arn = "${aws_iam_policy.foo.arn}"
}

resource "aws_s3_bucket" "foo" {
  bucket = "terraform-testacc-sagemaker-model-data-bucket-%s"
  acl    = "private"
  force_destroy = true
}

resource "aws_s3_bucket_object" "object" {
  bucket = "${aws_s3_bucket.foo.bucket}"
  key    = "model.tar.gz"
  content = "some-data"
}
`, rName, rName, rName)
}

func testAccSagemakerPrimaryContainerHostnameConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_model" "foo" {
	name = "terraform-testacc-sagemaker-model-%s"
	execution_role_arn = "${aws_iam_role.foo.arn}"

	primary_container {
		image = "174872318107.dkr.ecr.us-west-2.amazonaws.com/kmeans:1"
		container_hostname = "primary-hostname-foo"
	}
}

resource "aws_iam_role" "foo" {
  name = "terraform-testacc-sagemaker-model-%s"
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
`, rName, rName)
}

func testAccSagemakerPrimaryContainerEnvironmentConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_model" "foo" {
	name = "terraform-testacc-sagemaker-model-%s"
	execution_role_arn = "${aws_iam_role.foo.arn}"

	primary_container {
		image = "174872318107.dkr.ecr.us-west-2.amazonaws.com/kmeans:1"
		environment {
			foo = "bar"
		}
	}
}

resource "aws_iam_role" "foo" {
  name = "terraform-testacc-sagemaker-model-%s"
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
`, rName, rName)
}
