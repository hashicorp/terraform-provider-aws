package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
)

func TestAccAwsMacie2ClassificationJob_Name_Generated(t *testing.T) {
	var macie2Output macie2.DescribeClassificationJobOutput
	resourceName := "aws_macie2_classification_job.test"
	bucketName := os.Getenv("AWS_S3_BUCKET_NAME")
	JobType := "ONE_TIME"
	accountID := os.Getenv("AWS_ACCOUNT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testaccawsmacieClassificationJobconfigNameGenerated(bucketName, accountID, JobType),
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
	bucketName := os.Getenv("AWS_S3_BUCKET_NAME")
	JobType := "ONE_TIME"
	accountID := os.Getenv("AWS_ACCOUNT_ID")
	namePrefix := "tf-acc-test-prefix-"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testaccawsmacieClassificationJobconfigNamePrefix(namePrefix, bucketName, accountID, JobType),
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

func TestAccAwsMacie2ClassificationJob_complete(t *testing.T) {
	var macie2Output macie2.DescribeClassificationJobOutput
	resourceName := "aws_macie2_classification_job.test"
	bucketName := os.Getenv("AWS_S3_BUCKET_NAME")
	accountID := os.Getenv("AWS_ACCOUNT_ID")
	jobStatus := "RUNNING"
	jobStatusUpdated := "USER_PAUSED"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testaccawsmacieClassificationJobconfigComplete(bucketName, accountID, jobStatus),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2ClassificationJobExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttrSet(resourceName, "sampling_percentage"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrSet(resourceName, "initial_run"),
					resource.TestCheckResourceAttrSet(resourceName, "job_type"),
					resource.TestCheckResourceAttrSet(resourceName, "job_id"),
					resource.TestCheckResourceAttrSet(resourceName, "job_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "job_status"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
			{
				Config: testaccawsmacieClassificationJobconfigComplete(bucketName, accountID, jobStatusUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2ClassificationJobExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttrSet(resourceName, "sampling_percentage"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrSet(resourceName, "initial_run"),
					resource.TestCheckResourceAttrSet(resourceName, "job_type"),
					resource.TestCheckResourceAttrSet(resourceName, "job_id"),
					resource.TestCheckResourceAttrSet(resourceName, "job_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "job_status"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
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

func TestAccAwsMacie2ClassificationJob_completeWithTags(t *testing.T) {
	var macie2Output macie2.DescribeClassificationJobOutput
	resourceName := "aws_macie2_classification_job.test"
	bucketName := os.Getenv("AWS_S3_BUCKET_NAME")
	accountID := os.Getenv("AWS_ACCOUNT_ID")
	jobStatus := "RUNNING"
	jobStatusUpdated := "USER_PAUSED"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testaccawsmacieClassificationJobconfigCompleteWithTags(bucketName, accountID, jobStatus),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2ClassificationJobExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttrSet(resourceName, "sampling_percentage"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrSet(resourceName, "initial_run"),
					resource.TestCheckResourceAttrSet(resourceName, "job_type"),
					resource.TestCheckResourceAttrSet(resourceName, "job_id"),
					resource.TestCheckResourceAttrSet(resourceName, "job_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "job_status"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
			{
				Config: testaccawsmacieClassificationJobconfigComplete(bucketName, accountID, jobStatusUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2ClassificationJobExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttrSet(resourceName, "sampling_percentage"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrSet(resourceName, "initial_run"),
					resource.TestCheckResourceAttrSet(resourceName, "job_type"),
					resource.TestCheckResourceAttrSet(resourceName, "job_id"),
					resource.TestCheckResourceAttrSet(resourceName, "job_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "job_status"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
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
			return fmt.Errorf("macie2 ClassificationJob %q does not exist", rs.Primary.ID)
		}

		*macie2Session = *resp

		return nil
	}
}

func testaccawsmacieClassificationJobconfigNameGenerated(nameBucket, accountID, jobType string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {}

resource "aws_macie2_classification_job" "test" {
  job_type = "%s"
  s3_job_definition {
    bucket_definitions {
      account_id = "%s"
      buckets    = ["%s"]
    }
  }
  depends_on = [aws_macie2_account.test]
}
`, jobType, accountID, nameBucket)
}

func testaccawsmacieClassificationJobconfigNamePrefix(namePrefix, nameBucket, accountID, jobType string) string {
	return fmt.Sprintf(`resource "aws_macie2_account" "test" {}

resource "aws_macie2_classification_job" "test" {
	name_prefix = %[1]q
	job_type = "%s"
	s3_job_definition {
		bucket_definitions{
			account_id = "%s"
			buckets = ["%s"]
		}
	}
	depends_on = [aws_macie2_account.test]
}
`, namePrefix, jobType, accountID, nameBucket)
}

func testaccawsmacieClassificationJobconfigComplete(nameBucket, accountID, jobStatus string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {}

resource "aws_macie2_classification_job" "test" {
  job_type = "SCHEDULED"
  s3_job_definition {
    bucket_definitions {
      account_id = "%s"
      buckets    = ["%s"]
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
  job_status          = "%s"

  depends_on = [aws_macie2_account.test]
}
`, accountID, nameBucket, jobStatus)
}

func testaccawsmacieClassificationJobconfigCompleteWithTags(nameBucket, accountID, jobStatus string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {}

resource "aws_macie2_classification_job" "test" {
  job_type = "SCHEDULED"
  s3_job_definition {
    bucket_definitions {
      account_id = "%s"
      buckets    = ["%s"]
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
  job_status          = "%s"
  tags = {
    Key = "value"
  }

  depends_on = [aws_macie2_account.test]
}
`, accountID, nameBucket, jobStatus)
}
