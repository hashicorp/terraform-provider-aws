package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSS3BucketNotification_importBasic(t *testing.T) {
	rString := acctest.RandString(8)

	topicName := fmt.Sprintf("tf-acc-topic-s3-b-n-import-%s", rString)
	bucketName := fmt.Sprintf("tf-acc-bucket-n-import-%s", rString)
	queueName := fmt.Sprintf("tf-acc-queue-s3-b-n-import-%s", rString)
	roleName := fmt.Sprintf("tf-acc-role-s3-b-n-import-%s", rString)
	lambdaFuncName := fmt.Sprintf("tf-acc-lambda-func-s3-b-n-import-%s", rString)

	resourceName := "aws_s3_bucket_notification.notification"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketNotificationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithTopicNotification(topicName, bucketName),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketNotificationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithQueueNotification(queueName, bucketName),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketNotificationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithLambdaNotification(roleName, lambdaFuncName, bucketName),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
