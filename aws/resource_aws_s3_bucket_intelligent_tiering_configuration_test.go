package aws

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSS3BucketIntelligentTieringConfiguration_basic(t *testing.T) {
	var ac s3.AnalyticsConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, s3.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketAnalyticsConfiguration(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketAnalyticsConfigurationExists(resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "archive_configuration.#", "0"),
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

func testAccCheckAWSS3BucketIntelligentTieringConfigurationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).s3conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket_intelligent_tiering_configuration" {
			continue
		}

		id = rs.Primary.Id
		bucket = rs.Priamry.Bucket
		if err != nil {
			return err
		}

		return resourceAwsS3IntelligentTieringConfigurationDelete(conn, bucket, name, 1*time.Minute)

	}
	return nil
}