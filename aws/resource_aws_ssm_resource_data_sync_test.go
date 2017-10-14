package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSSsmResourceDataSync(t *testing.T) {
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSsmResourceDataSyncDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSsmResourceDataSyncConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSsmResourceDataSyncExists("aws_ssm_resource_data_sync.foo"),
				),
			},
		},
	})
}

func testAccCheckAWSSsmResourceDataSyncDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ssmconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssm_resource_data_sync" {
			continue
		}
		nextToken := ""
		found := false
		for {
			input := &ssm.ListResourceDataSyncInput{}
			if nextToken != "" {
				input.NextToken = aws.String(nextToken)
			}
			resp, err := conn.ListResourceDataSync(input)
			if err != nil {
				return err
			}
			for _, v := range resp.ResourceDataSyncItems {
				if *v.SyncName == rs.Primary.Attributes["name"] {
					found = true
				}
			}
			if found || *resp.NextToken == "" {
				break
			}
			nextToken = *resp.NextToken
		}
		if !found {
			return fmt.Errorf("No Resource Data Sync found for SyncName: %s", rs.Primary.Attributes["name"])
		}
	}
	return nil
}

func testAccCheckAWSSsmResourceDataSyncExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		log.Println(s.RootModule().Resources)
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccSsmResourceDataSyncConfig(randInt int) string {
	return fmt.Sprintf(`
    resource "aws_s3_bucket" "hoge" {
      bucket = "tf-test-bucket-%d"
      region = "us-east-1"
      force_destroy = true
    }

    resource "aws_s3_bucket_policy" "hoge" {
    	bucket = "${aws_s3_bucket.hoge.bucket}"
    	policy = <<EOF
{
        "Version": "2012-10-17",
        "Statement": [
            {
                "Sid": "SSMBucketPermissionsCheck",
                "Effect": "Allow",
                "Principal": {
                    "Service": "ssm.amazonaws.com"
                },
                "Action": "s3:GetBucketAcl",
                "Resource": "arn:aws:s3:::tf-test-bucket-%d"
            },
            {
                "Sid": " SSMBucketDelivery",
                "Effect": "Allow",
                "Principal": {
                    "Service": "ssm.amazonaws.com"
                },
                "Action": "s3:PutObject",
                "Resource": ["arn:aws:s3:::tf-test-bucket-%d/*"],
                "Condition": {
                    "StringEquals": {
                        "s3:x-amz-acl": "bucket-owner-full-control"
                    }
                }
            }
          ]
      }
      EOF
    }

    resource "aws_ssm_resource_data_sync" "foo" {
      name = "foo"
      destination = {
        bucket = "${aws_s3_bucket.hoge.bucket}"
        region = "${aws_s3_bucket.hoge.region}"
      }
    }
    `, randInt, randInt, randInt)
}
