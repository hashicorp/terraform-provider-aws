package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
)

func TestAccAwsMacie2ClassificationJob_basic(t *testing.T) {
	var macie2Output macie2.DescribeClassificationJobOutput
	resourceName := "aws_macie2_classification_job.test"
	bucketName := "test-bucket-name-aws"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2ClassificationJobDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieClassificationJobconfigNameGenerated(bucketName, macie2.JobTypeOneTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2ClassificationJobExists(resourceName, &macie2Output),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "job_type", macie2.JobTypeOneTime),
					resource.TestCheckResourceAttr(resourceName, "s3_job_definition.0.bucket_definitions.0.buckets.0", bucketName),
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

func TestAccAwsMacie2ClassificationJob_Name_Generated(t *testing.T) {
	var macie2Output macie2.DescribeClassificationJobOutput
	resourceName := "aws_macie2_classification_job.test"
	bucketName := "test-bucket-name-aws"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2ClassificationJobDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieClassificationJobconfigNameGenerated(bucketName, macie2.JobTypeOneTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2ClassificationJobExists(resourceName, &macie2Output),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
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

func TestAccAwsMacie2ClassificationJob_NamePrefix(t *testing.T) {
	var macie2Output macie2.DescribeClassificationJobOutput
	resourceName := "aws_macie2_classification_job.test"
	bucketName := "test-bucket-name-aws"
	namePrefix := "tf-acc-test-prefix-"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2ClassificationJobDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieClassificationJobconfigNamePrefix(bucketName, namePrefix, macie2.JobTypeOneTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2ClassificationJobExists(resourceName, &macie2Output),
					naming.TestCheckResourceAttrNameFromPrefix(resourceName, "name", namePrefix),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", namePrefix),
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

func TestAccAwsMacie2ClassificationJob_disappears(t *testing.T) {
	var macie2Output macie2.DescribeClassificationJobOutput
	resourceName := "aws_macie2_classification_job.test"
	bucketName := "test-bucket-name-aws"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2ClassificationJobDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieClassificationJobconfigNameGenerated(bucketName, macie2.JobTypeOneTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2ClassificationJobExists(resourceName, &macie2Output),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsMacie2ClassificationJob(), resourceName),
				),
			},
		},
	})
}

func TestAccAwsMacie2ClassificationJob_Status(t *testing.T) {
	var macie2Output, macie2Output2 macie2.DescribeClassificationJobOutput
	resourceName := "aws_macie2_classification_job.test"
	bucketName := "test-bucket-name-aws"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2ClassificationJobDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieClassificationJobconfigStatus(bucketName, macie2.JobStatusRunning),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2ClassificationJobExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "job_type", macie2.JobTypeScheduled),
					resource.TestCheckResourceAttr(resourceName, "job_status", macie2.JobStatusRunning),
					resource.TestCheckResourceAttr(resourceName, "s3_job_definition.0.bucket_definitions.0.buckets.0", bucketName),
					testAccCheckResourceAttrRfc3339(resourceName, "created_at"),
				),
			},
			{
				Config: testAccAwsMacieClassificationJobconfigStatus(bucketName, macie2.JobStatusUserPaused),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2ClassificationJobExists(resourceName, &macie2Output2),
					testAccCheckAwsMacie2ClassificationJobNotRecreated(&macie2Output, &macie2Output2),
					resource.TestCheckResourceAttr(resourceName, "job_type", macie2.JobTypeScheduled),
					resource.TestCheckResourceAttr(resourceName, "job_status", macie2.JobStatusUserPaused),
					resource.TestCheckResourceAttr(resourceName, "s3_job_definition.0.bucket_definitions.0.buckets.0", bucketName),
					testAccCheckResourceAttrRfc3339(resourceName, "created_at"),
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

func TestAccAwsMacie2ClassificationJob_complete(t *testing.T) {
	var macie2Output macie2.DescribeClassificationJobOutput
	resourceName := "aws_macie2_classification_job.test"
	bucketName := "test-bucket-name-aws"
	description := "Description of a test"
	descriptionUpdated := "Updated Description of a test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2ClassificationJobDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieClassificationJobconfigComplete(bucketName, macie2.JobStatusRunning, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2ClassificationJobExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "job_type", macie2.JobTypeScheduled),
					resource.TestCheckResourceAttr(resourceName, "job_status", macie2.JobStatusRunning),
					resource.TestCheckResourceAttr(resourceName, "s3_job_definition.0.bucket_definitions.0.buckets.0", bucketName),
					testAccCheckResourceAttrRfc3339(resourceName, "created_at"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
				),
			},
			{
				Config: testAccAwsMacieClassificationJobconfigComplete(bucketName, macie2.JobStatusRunning, descriptionUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2ClassificationJobExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "job_type", macie2.JobTypeScheduled),
					resource.TestCheckResourceAttr(resourceName, "job_status", macie2.JobStatusRunning),
					testAccCheckResourceAttrRfc3339(resourceName, "created_at"),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionUpdated),
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

func TestAccAwsMacie2ClassificationJob_WithTags(t *testing.T) {
	var macie2Output macie2.DescribeClassificationJobOutput
	resourceName := "aws_macie2_classification_job.test"
	bucketName := "test-bucket-name-aws"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2ClassificationJobDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieClassificationJobconfigCompleteWithTags(bucketName, macie2.JobStatusRunning),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2ClassificationJobExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "job_type", macie2.JobTypeScheduled),
					resource.TestCheckResourceAttr(resourceName, "job_status", macie2.JobStatusRunning),
					testAccCheckResourceAttrRfc3339(resourceName, "created_at"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", "value"),
				),
			},
			{
				Config: testAccAwsMacieClassificationJobconfigCompleteWithTags(bucketName, macie2.JobStatusUserPaused),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2ClassificationJobExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "job_type", macie2.JobTypeScheduled),
					resource.TestCheckResourceAttr(resourceName, "job_status", macie2.JobStatusUserPaused),
					testAccCheckResourceAttrRfc3339(resourceName, "created_at"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", "value"),
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

func testAccCheckAwsMacie2ClassificationJobExists(resourceName string, macie2Session *macie2.DescribeClassificationJobOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).macie2conn
		input := &macie2.DescribeClassificationJobInput{JobId: aws.String(rs.Primary.ID)}

		resp, err := conn.DescribeClassificationJob(input)

		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("macie ClassificationJob %q does not exist", rs.Primary.ID)
		}

		*macie2Session = *resp

		return nil
	}
}

func testAccCheckAwsMacie2ClassificationJobDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).macie2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_macie2_classification_job" {
			continue
		}
		input := &macie2.DescribeClassificationJobInput{JobId: aws.String(rs.Primary.ID)}

		resp, err := conn.DescribeClassificationJob(input)

		if tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeValidationException, "cannot update cancelled job for job") {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil && aws.StringValue(resp.JobStatus) != macie2.JobStatusCancelled {
			return fmt.Errorf("macie ClassificationJob %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsMacie2ClassificationJobNotRecreated(i, j *macie2.DescribeClassificationJobOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.CreatedAt).Equal(aws.TimeValue(j.CreatedAt)) {
			return fmt.Errorf("Macie Classification Job recreated")
		}

		return nil
	}
}

func testAccAwsMacieClassificationJobconfigNameGenerated(bucketName, jobType string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_macie2_account" "test" {}

resource "aws_s3_bucket" "test" {
  bucket = "%s"
}

resource "aws_macie2_classification_job" "test" {
  depends_on = [aws_macie2_account.test]
  job_type   = "%s"
  s3_job_definition {
    bucket_definitions {
      account_id = data.aws_caller_identity.current.account_id
      buckets    = [aws_s3_bucket.test.bucket]
    }
  }
}
`, bucketName, jobType)
}

func testAccAwsMacieClassificationJobconfigNamePrefix(nameBucket, namePrefix, jobType string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_macie2_account" "test" {}

resource "aws_s3_bucket" "test" {
  bucket = %q
}

resource "aws_macie2_classification_job" "test" {
  name_prefix = %[2]q
  job_type    = %q
  s3_job_definition {
    bucket_definitions {
      account_id = data.aws_caller_identity.current.account_id
      buckets    = [aws_s3_bucket.test.bucket]
    }
  }
  depends_on = [aws_macie2_account.test]
}
`, nameBucket, namePrefix, jobType)
}

func testAccAwsMacieClassificationJobconfigComplete(nameBucket, jobStatus, description string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_macie2_account" "test" {}

resource "aws_s3_bucket" "test" {
  bucket = %q
}

resource "aws_macie2_classification_job" "test" {
  job_type = "SCHEDULED"
  s3_job_definition {
    bucket_definitions {
      account_id = data.aws_caller_identity.current.account_id
      buckets    = [aws_s3_bucket.test.bucket]
    }
    scoping {
      excludes {
        and {
          simple_scope_term {
            comparator = "EQ"
            key        = "OBJECT_EXTENSION"
            values     = ["test"]
          }
        }
      }
      includes {
        and {
          simple_scope_term {
            comparator = "EQ"
            key        = "OBJECT_EXTENSION"
            values     = ["test"]
          }
        }
      }
    }
  }
  schedule_frequency {
    daily_schedule = true
  }
  sampling_percentage = 100
  initial_run         = true
  job_status          = %q
  description         = %q

  depends_on = [aws_macie2_account.test]
}
`, nameBucket, jobStatus, description)
}

func testAccAwsMacieClassificationJobconfigStatus(nameBucket, jobStatus string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_macie2_account" "test" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_macie2_classification_job" "test" {
  job_type = "SCHEDULED"
  s3_job_definition {
    bucket_definitions {
      account_id = data.aws_caller_identity.current.account_id
      buckets    = [aws_s3_bucket.test.bucket]
    }
    scoping {
      excludes {
        and {
          simple_scope_term {
            comparator = "EQ"
            key        = "OBJECT_EXTENSION"
            values     = ["test"]
          }
        }
      }
      includes {
        and {
          simple_scope_term {
            comparator = "EQ"
            key        = "OBJECT_EXTENSION"
            values     = ["test"]
          }
        }
      }
    }
  }
  schedule_frequency {
    daily_schedule = true
  }
  sampling_percentage = 100
  initial_run         = true
  job_status          = %[2]q

  depends_on = [aws_macie2_account.test]
}
`, nameBucket, jobStatus)
}

func testAccAwsMacieClassificationJobconfigCompleteWithTags(nameBucket, jobStatus string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_macie2_account" "test" {}

resource "aws_s3_bucket" "test" {
  bucket = %q
}

resource "aws_macie2_classification_job" "test" {
  job_type = "SCHEDULED"
  s3_job_definition {
    bucket_definitions {
      account_id = data.aws_caller_identity.current.account_id
      buckets    = [aws_s3_bucket.test.bucket]
    }
    scoping {
      excludes {
        and {
          simple_scope_term {
            comparator = "EQ"
            key        = "OBJECT_EXTENSION"
            values     = ["test"]
          }
        }
      }
      includes {
        and {
          simple_scope_term {
            comparator = "EQ"
            key        = "OBJECT_EXTENSION"
            values     = ["test"]
          }
        }
      }
    }
  }
  schedule_frequency {
    daily_schedule = true
  }
  sampling_percentage = 100
  description         = "test"
  initial_run         = true
  job_status          = %q
  tags = {
    Key = "value"
  }

  depends_on = [aws_macie2_account.test]
}
`, nameBucket, jobStatus)
}
