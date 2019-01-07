package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const sagemakerTestAccSagemakerTrainingJobResourceNamePrefix = "terraform-testacc-"

func init() {
	resource.AddTestSweepers("aws_sagemaker_training_job", &resource.Sweeper{
		Name: "aws_sagemaker_training_job",
	})
}

func TestAccAWSSagemakerTrainingJob_basic(t *testing.T) {
	var trainingJob sagemaker.DescribeTrainingJobOutput
	trainingJobName := resource.PrefixedUniqueId(sagemakerTestAccSagemakerTrainingJobResourceNamePrefix)
	bucketName := resource.PrefixedUniqueId(sagemakerTestAccSagemakerTrainingJobResourceNamePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerTrainingJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerTrainingJobConfig(trainingJobName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerTrainingJobExists("aws_sagemaker_training_job.foo", &trainingJob),
					testAccCheckSagemakerTrainingJobName(&trainingJob, trainingJobName),

					resource.TestCheckResourceAttr(
						"aws_sagemaker_training_job.foo", "name", trainingJobName),
				),
			},
		},
	})
}

func TestAccAWSSagemakerTrainingJob_update(t *testing.T) {
	var trainingJob sagemaker.DescribeTrainingJobOutput
	trainingJobName := resource.PrefixedUniqueId(sagemakerTestAccSagemakerTrainingJobResourceNamePrefix)
	bucketName := resource.PrefixedUniqueId(sagemakerTestAccSagemakerTrainingJobResourceNamePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerTrainingJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerTrainingJobConfig(trainingJobName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerTrainingJobExists("aws_sagemaker_training_job.foo", &trainingJob),

					resource.TestCheckResourceAttr(
						"aws_sagemaker_training_job.foo", "hyper_parameters.epochs", "3"),
				),
			},

			{
				Config: testAccSagemakerTrainingJobUpdateConfig(trainingJobName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerTrainingJobExists("aws_sagemaker_training_job.foo", &trainingJob),

					resource.TestCheckResourceAttr(
						"aws_sagemaker_training_job.foo", "hyper_parameters.epochs", "3"),
				),
				ExpectError: regexp.MustCompile(`.*existing Training Jobs cannot be updated.*`),
			},
		},
	})
}
func TestAccAWSSagemakerTrainingJob_tags(t *testing.T) {
	var trainingJob sagemaker.DescribeTrainingJobOutput
	trainingJobName := resource.PrefixedUniqueId(sagemakerTestAccSagemakerTrainingJobResourceNamePrefix)
	bucketName := resource.PrefixedUniqueId(sagemakerTestAccSagemakerTrainingJobResourceNamePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerTrainingJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerTrainingJobTagsConfig(trainingJobName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerTrainingJobExists("aws_sagemaker_training_job.foo", &trainingJob),
					testAccCheckSagemakerTrainingJobTags(&trainingJob, "foo", "bar"),

					resource.TestCheckResourceAttr(
						"aws_sagemaker_training_job.foo", "name", trainingJobName),
					resource.TestCheckResourceAttr("aws_sagemaker_training_job.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_sagemaker_training_job.foo", "tags.foo", "bar"),
				),
			},

			{
				Config: testAccSagemakerTrainingJobTagsUpdateConfig(trainingJobName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerTrainingJobExists("aws_sagemaker_training_job.foo", &trainingJob),
					testAccCheckSagemakerTrainingJobTags(&trainingJob, "foo", ""),
					testAccCheckSagemakerTrainingJobTags(&trainingJob, "bar", "baz"),

					resource.TestCheckResourceAttr("aws_sagemaker_training_job.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_sagemaker_training_job.foo", "tags.bar", "baz"),
				),
			},
		},
	})
}

func testAccCheckSagemakerTrainingJobDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_training_job" {
			continue
		}

		resp, err := conn.ListTrainingJobs(&sagemaker.ListTrainingJobsInput{
			NameContains: aws.String(rs.Primary.ID),
			StatusEquals: aws.String("InProgress"),
		})
		if err == nil {
			if len(resp.TrainingJobSummaries) > 0 {
				return fmt.Errorf("Sagemaker Training Job still exist.")
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

func testAccCheckSagemakerTrainingJobExists(n string, trainingJob *sagemaker.DescribeTrainingJobOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker Training Job ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn
		opts := &sagemaker.DescribeTrainingJobInput{
			TrainingJobName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeTrainingJob(opts)
		if err != nil {
			return err
		}

		*trainingJob = *resp

		return nil
	}
}

func testAccCheckSagemakerTrainingJobName(trainingJob *sagemaker.DescribeTrainingJobOutput, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		trainingJobName := trainingJob.TrainingJobName
		if *trainingJobName != expected {
			return fmt.Errorf("Bad Training Job name: %s", *trainingJob.TrainingJobName)
		}

		return nil
	}
}

func testAccCheckSagemakerTrainingJobTags(trainingJob *sagemaker.DescribeTrainingJobOutput, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

		ts, err := conn.ListTags(&sagemaker.ListTagsInput{
			ResourceArn: trainingJob.TrainingJobArn,
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

func testAccSagemakerTrainingJobConfig(trainingJobName string, bucketName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_training_job" "foo" {
	name = "%s"
	role_arn = "${aws_iam_role.foo.arn}"
	algorithm_specification {
		image = "174872318107.dkr.ecr.us-west-2.amazonaws.com/kmeans:1"
		input_mode = "File"
	}
	resource_config {
		instance_type = "ml.c4.8xlarge"
		instance_count = 2
		volume_size_in_gb = 30
	}
	stopping_condition {
		max_runtime_in_seconds = 3600
	}
	hyper_parameters {
		epochs = "3"
		feature_dim = "784"
		force_dense = "True"
		k = "10"
		mini_batch_size = "500"
	}
	input_data_config = [{
		name = "train-1"
		data_source {
			s3_data_source {
				s3_data_type = "ManifestFile"
				s3_uri = "s3://${aws_s3_bucket.foo.id}/kmeans_highlevel_example/data/KMeans-2018-01-10-14-45-00-000/.amazon.manifest"
				s3_data_distribution_type = "ShardedByS3Key"
			}
		}
	}]
	output_data_config {
		s3_output_path = "s3://${aws_s3_bucket.foo.id}/kmeans_example/output"
	}
}

resource "aws_s3_bucket" "foo" {
	bucket = "%s"
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
`, trainingJobName, bucketName)
}

func testAccSagemakerTrainingJobUpdateConfig(trainingJobName string, bucketName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_training_job" "foo" {
	name = "%s"
	role_arn = "${aws_iam_role.foo.arn}"
	algorithm_specification {
		image = "174872318107.dkr.ecr.us-west-2.amazonaws.com/kmeans:1"
		input_mode = "File"
	}
	resource_config {
		instance_type = "ml.c4.8xlarge"
		instance_count = 2
		volume_size_in_gb = 30
	}
	stopping_condition {
		max_runtime_in_seconds = 3600
	}
	hyper_parameters {
		epochs = "6"
		feature_dim = "784"
		force_dense = "True"
		k = "10"
		mini_batch_size = "500"
	}
	input_data_config = [{
		name = "train-1"
		data_source {
			s3_data_source {
				s3_data_type = "ManifestFile"
				s3_uri = "s3://${aws_s3_bucket.foo.id}/kmeans_highlevel_example/data/KMeans-2018-01-10-14-45-00-000/.amazon.manifest"
				s3_data_distribution_type = "ShardedByS3Key"
			}
		}
	}]
	output_data_config {
		s3_output_path = "s3://${aws_s3_bucket.foo.id}/kmeans_example/output"
	}
}

resource "aws_s3_bucket" "foo" {
	bucket = "%s"
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
`, trainingJobName, bucketName)
}

func testAccSagemakerTrainingJobTagsConfig(trainingJobName string, bucketName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_training_job" "foo" {
	name = "%s"
	role_arn = "${aws_iam_role.foo.arn}"
	algorithm_specification {
		image = "174872318107.dkr.ecr.us-west-2.amazonaws.com/kmeans:1"
		input_mode = "File"
	}
	resource_config {
		instance_type = "ml.c4.8xlarge"
		instance_count = 2
		volume_size_in_gb = 30
	}
	stopping_condition {
		max_runtime_in_seconds = 3600
	}
	hyper_parameters {
		epochs = "3"
		feature_dim = "784"
		force_dense = "True"
		k = "10"
		mini_batch_size = "500"
	}
	input_data_config = [{
		name = "train"
		data_source {
			s3_data_source {
				s3_data_type = "ManifestFile"
				s3_uri = "s3://${aws_s3_bucket.foo.id}/kmeans_highlevel_example/data/KMeans-2018-01-10-14-45-00-000/.amazon.manifest"
				s3_data_distribution_type = "ShardedByS3Key"
			}
		}
	}]
	output_data_config {
		s3_output_path = "s3://${aws_s3_bucket.foo.id}/kmeans_example/output"
	}
	tags {
		foo = "bar"
	}
}

resource "aws_s3_bucket" "foo" {
	bucket = "%s"
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
`, trainingJobName, bucketName)
}

func testAccSagemakerTrainingJobTagsUpdateConfig(trainingJobName string, bucketName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_training_job" "foo" {
	name = "%s"
	role_arn = "${aws_iam_role.foo.arn}"
	algorithm_specification {
		image = "174872318107.dkr.ecr.us-west-2.amazonaws.com/kmeans:1"
		input_mode = "File"
	}
	resource_config {
		instance_type = "ml.c4.8xlarge"
		instance_count = 2
		volume_size_in_gb = 30
	}
	stopping_condition {
		max_runtime_in_seconds = 3600
	}
	hyper_parameters {
		epochs = "3"
		feature_dim = "784"
		force_dense = "True"
		k = "10"
		mini_batch_size = "500"
	}
	input_data_config = [{
		name = "train"
		data_source {
			s3_data_source {
				s3_data_type = "ManifestFile"
				s3_uri = "s3://${aws_s3_bucket.foo.id}/kmeans_highlevel_example/data/KMeans-2018-01-10-14-45-00-000/.amazon.manifest"
				s3_data_distribution_type = "ShardedByS3Key"
			}
		}
	}]
	output_data_config {
		s3_output_path = "s3://${aws_s3_bucket.foo.id}/kmeans_example/output"
	}
	tags {
		bar = "baz"
	}
}

resource "aws_s3_bucket" "foo" {
	bucket = "%s"
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
`, trainingJobName, bucketName)
}
