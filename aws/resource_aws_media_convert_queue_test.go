package aws

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/mediaconvert"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSMediaConvertQueue_basic(t *testing.T) {
	var queue mediaconvert.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	endpointURL := os.Getenv("AWS_MEDIA_CONVERT_ENDPOINT")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMediaConvertQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMediaConvertQueueConfig_Basic(endpointURL, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaConvertQueueExists(resourceName, &queue),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "mediaconvert", regexp.MustCompile(`queues/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "pricing_plan", mediaconvert.PricingPlanOnDemand),
					resource.TestCheckResourceAttr(resourceName, "status", mediaconvert.QueueStatusActive),
				),
			},
		},
	})
}

func TestAccAWSMediaConvertQueue_withStatus(t *testing.T) {
	var queue mediaconvert.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	endpointURL := os.Getenv("AWS_MEDIA_CONVERT_ENDPOINT")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMediaConvertQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMediaConvertQueueConfig_withStatus(endpointURL, rName, mediaconvert.QueueStatusPaused),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaConvertQueueExists(resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "status", mediaconvert.QueueStatusPaused),
				),
			},
			{
				Config: testAccMediaConvertQueueConfig_withStatus(endpointURL, rName, mediaconvert.QueueStatusActive),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaConvertQueueExists(resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "status", mediaconvert.QueueStatusActive),
				),
			},
		},
	})
}

func TestAccAWSMediaConvertQueue_withTags(t *testing.T) {
	var queue mediaconvert.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	endpointURL := os.Getenv("AWS_MEDIA_CONVERT_ENDPOINT")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMediaConvertQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMediaConvertQueueConfig_withTags(endpointURL, rName, "foo", "bar", "fizz", "buzz"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaConvertQueueExists(resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.fizz", "buzz"),
				),
			},
			{
				Config: testAccMediaConvertQueueConfig_withTags(endpointURL, rName, "foo", "bar2", "fizz2", "buzz2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaConvertQueueExists(resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar2"),
					resource.TestCheckResourceAttr(resourceName, "tags.fizz2", "buzz2"),
				),
			},
			{
				Config: testAccMediaConvertQueueConfig_Basic(endpointURL, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaConvertQueueExists(resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSMediaConvertQueue_withDescription(t *testing.T) {
	var queue mediaconvert.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	description1 := acctest.RandomWithPrefix("Description: ")
	description2 := acctest.RandomWithPrefix("Description: ")
	endpointURL := os.Getenv("AWS_MEDIA_CONVERT_ENDPOINT")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMediaConvertQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMediaConvertQueueConfig_withDescription(endpointURL, rName, description1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaConvertQueueExists(resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "description", description1),
				),
			},
			{
				Config: testAccMediaConvertQueueConfig_withDescription(endpointURL, rName, description2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaConvertQueueExists(resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "description", description2),
				),
			},
		},
	})
}

func testAccCheckAwsMediaConvertQueueDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_media_convert_queue" {
			continue
		}
		originalConn := testAccProvider.Meta().(*AWSClient).mediaconvertconn
		endpointURL := os.Getenv("AWS_MEDIA_CONVERT_ENDPOINT")
		if endpointURL == "" {
			return fmt.Errorf("No AWS Media Convert Endpoint is set")
		}

		sess, err := session.NewSession(&originalConn.Config)
		if err != nil {
			return fmt.Errorf("Error creating AWS session: %s", err)
		}
		conn := mediaconvert.New(sess.Copy(&aws.Config{Endpoint: aws.String(endpointURL)}))

		_, err = conn.GetQueue(&mediaconvert.GetQueueInput{
			Name: aws.String(rs.Primary.ID),
		})
		if err != nil {
			if isAWSErr(err, mediaconvert.ErrCodeNotFoundException, "") {
				continue
			}
			return err
		}
	}

	return nil
}

func testAccCheckAwsMediaConvertQueueExists(n string, queue *mediaconvert.Queue) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Queue id is set")
		}

		originalConn := testAccProvider.Meta().(*AWSClient).mediaconvertconn
		endpointURL := os.Getenv("AWS_MEDIA_CONVERT_ENDPOINT")
		if endpointURL == "" {
			return fmt.Errorf("No AWS Media Convert Endpoint is set")
		}

		sess, err := session.NewSession(&originalConn.Config)
		if err != nil {
			return fmt.Errorf("Error creating AWS session: %s", err)
		}
		conn := mediaconvert.New(sess.Copy(&aws.Config{Endpoint: aws.String(endpointURL)}))

		resp, err := conn.GetQueue(&mediaconvert.GetQueueInput{
			Name: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("Error getting queue: %s", err)
		}

		*queue = *resp.Queue
		return nil
	}
}

func testAccMediaConvertQueueConfig_Base(endpoint string) string {
	return fmt.Sprintf(`
provider "aws" {
  endpoints {
    mediaconvert = %[1]q
  }
}

`, endpoint)
}

func testAccMediaConvertQueueConfig_Basic(endpoint, rName string) string {
	return testAccMediaConvertQueueConfig_Base(endpoint) + fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name = %[1]q
}

`, rName)
}

func testAccMediaConvertQueueConfig_withStatus(endpoint, rName, status string) string {
	return testAccMediaConvertQueueConfig_Base(endpoint) + fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name   = %[1]q
  status = %[2]q
}

`, rName, status)
}

func testAccMediaConvertQueueConfig_withDescription(endpoint, rName, description string) string {
	return testAccMediaConvertQueueConfig_Base(endpoint) + fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name   = %[1]q
  description = %[2]q
}

`, rName, description)
}

func testAccMediaConvertQueueConfig_withTags(endpoint, rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccMediaConvertQueueConfig_Base(endpoint) + fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name   = %[1]q
  
  tags = {
	  %[2]s = %[3]q
	  %[4]s = %[5]q
  }
}

`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
