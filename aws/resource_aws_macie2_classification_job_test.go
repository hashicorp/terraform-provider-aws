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
	accountID := os.Getenv("AWS_ACCOUNT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieClassificationJobconfigNameGenerated(bucketName, accountID, macie2.JobTypeOneTime),
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
	accountID := os.Getenv("AWS_ACCOUNT_ID")
	namePrefix := "tf-acc-test-prefix-"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieClassificationJobconfigNamePrefix(namePrefix, bucketName, accountID, macie2.JobTypeOneTime),
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

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieClassificationJobconfigComplete(bucketName, accountID, macie2.JobStatusRunning),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2ClassificationJobExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "job_type", macie2.JobTypeScheduled),
					resource.TestCheckResourceAttr(resourceName, "job_status", macie2.JobStatusRunning),
					testAccCheckResourceAttrRfc3339(resourceName, "created_at"),
					testAccCheckResourceAttrRfc3339(resourceName, "last_run_time"),
				),
			},
			{
				Config: testAccAwsMacieClassificationJobconfigComplete(bucketName, accountID, macie2.JobStatusUserPaused),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2ClassificationJobExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "job_type", macie2.JobTypeScheduled),
					resource.TestCheckResourceAttr(resourceName, "job_status", macie2.JobStatusUserPaused),
					testAccCheckResourceAttrRfc3339(resourceName, "created_at"),
					testAccCheckResourceAttrRfc3339(resourceName, "last_run_time"),
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
	bucketName := os.Getenv("AWS_S3_BUCKET_NAME")
	accountID := os.Getenv("AWS_ACCOUNT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieClassificationJobconfigCompleteWithTags(bucketName, accountID, macie2.JobStatusRunning),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2ClassificationJobExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "job_type", macie2.JobTypeScheduled),
					resource.TestCheckResourceAttr(resourceName, "job_status", macie2.JobStatusRunning),
					testAccCheckResourceAttrRfc3339(resourceName, "created_at"),
					testAccCheckResourceAttrRfc3339(resourceName, "last_run_time"),
				),
			},
			{
				Config: testAccAwsMacieClassificationJobconfigComplete(bucketName, accountID, macie2.JobStatusUserPaused),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2ClassificationJobExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "job_type", macie2.JobTypeScheduled),
					resource.TestCheckResourceAttr(resourceName, "job_status", macie2.JobStatusUserPaused),
					testAccCheckResourceAttrRfc3339(resourceName, "created_at"),
					testAccCheckResourceAttrRfc3339(resourceName, "last_run_time"),
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

func testAccAwsMacieClassificationJobconfigNameGenerated(nameBucket, accountID, jobType string) string {
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

func testAccAwsMacieClassificationJobconfigNamePrefix(namePrefix, nameBucket, accountID, jobType string) string {
	return fmt.Sprintf(`resource "aws_macie2_account" "test" {}

resource "aws_macie2_classification_job" "test" {
  name_prefix = "%[1]s"
  job_type    = "%s"
  s3_job_definition {
    bucket_definitions {
      account_id = "%s"
      buckets    = ["%s"]
    }
  }
  depends_on = [aws_macie2_account.test]
}
`, namePrefix, jobType, accountID, nameBucket)
}

func testAccAwsMacieClassificationJobconfigComplete(nameBucket, accountID, jobStatus string) string {
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

func testAccAwsMacieClassificationJobconfigCompleteWithTags(nameBucket, accountID, jobStatus string) string {
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
