package cloudwatchevidently_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevidently"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudwatchevidently "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatchevidently"
)

func TestAccProject_basic(t *testing.T) {
	var project cloudwatchevidently.GetProjectOutput

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	originalDescription := "original description"
	updatedDescription := "updated description"
	resourceName := "aws_cloudwatchevidently_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudwatchevidently.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_basic(rName, originalDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttrSet(resourceName, "active_experiment_count"),
					resource.TestCheckResourceAttrSet(resourceName, "active_launch_count"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "description", originalDescription),
					resource.TestCheckResourceAttrSet(resourceName, "experiment_count"),
					resource.TestCheckResourceAttrSet(resourceName, "feature_count"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_time"),
					resource.TestCheckResourceAttrSet(resourceName, "launch_count"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Test Project"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_basic(rName, updatedDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "active_experiment_count"),
					resource.TestCheckResourceAttrSet(resourceName, "active_launch_count"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "description", updatedDescription),
					resource.TestCheckResourceAttrSet(resourceName, "experiment_count"),
					resource.TestCheckResourceAttrSet(resourceName, "feature_count"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_time"),
					resource.TestCheckResourceAttrSet(resourceName, "launch_count"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Test Project"),
				),
			},
		},
	})
}

func TestAccProject_updateTags(t *testing.T) {
	var project cloudwatchevidently.GetProjectOutput

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	description := "example description"
	resourceName := "aws_cloudwatchevidently_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudwatchevidently.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_basic(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttrSet(resourceName, "active_experiment_count"),
					resource.TestCheckResourceAttrSet(resourceName, "active_launch_count"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttrSet(resourceName, "experiment_count"),
					resource.TestCheckResourceAttrSet(resourceName, "feature_count"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_time"),
					resource.TestCheckResourceAttrSet(resourceName, "launch_count"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Test Project"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_tags(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttrSet(resourceName, "active_experiment_count"),
					resource.TestCheckResourceAttrSet(resourceName, "active_launch_count"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttrSet(resourceName, "experiment_count"),
					resource.TestCheckResourceAttrSet(resourceName, "feature_count"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_time"),
					resource.TestCheckResourceAttrSet(resourceName, "launch_count"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Test Project"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_tagsUpdated(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttrSet(resourceName, "active_experiment_count"),
					resource.TestCheckResourceAttrSet(resourceName, "active_launch_count"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttrSet(resourceName, "experiment_count"),
					resource.TestCheckResourceAttrSet(resourceName, "feature_count"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_time"),
					resource.TestCheckResourceAttrSet(resourceName, "launch_count"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Test Project"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
				),
			},
		},
	})
}

func TestAccProject_updateDataDelivery(t *testing.T) {
	var project cloudwatchevidently.GetProjectOutput

	rName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	description := "example description"
	resourceName := "aws_cloudwatchevidently_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudwatchevidently.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_dataDeliveryCloudWatchLogs(rName, rName2, rName3, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttrSet(resourceName, "active_experiment_count"),
					resource.TestCheckResourceAttrSet(resourceName, "active_launch_count"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "data_delivery.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_delivery.0.cloudwatch_logs.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "data_delivery.0.cloudwatch_logs.0.log_group", "aws_cloudwatch_log_group.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttrSet(resourceName, "experiment_count"),
					resource.TestCheckResourceAttrSet(resourceName, "feature_count"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_time"),
					resource.TestCheckResourceAttrSet(resourceName, "launch_count"),
					resource.TestCheckResourceAttr(resourceName, "name", rName3),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Test Project"),
				),
			},
			// A bug in the service API for UpdateProjectDataDelivery has been reported
			// {
			// 	ResourceName:      resourceName,
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// },
			// {
			// 	Config: testAccProjectConfig_dataDeliveryS3Bucket(rName, rName2, rName3, description),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		testAccCheckProjectExists(resourceName, &project),
			// 		resource.TestCheckResourceAttrSet(resourceName, "active_experiment_count"),
			// 		resource.TestCheckResourceAttrSet(resourceName, "active_launch_count"),
			// 		resource.TestCheckResourceAttrSet(resourceName, "arn"),
			// 		resource.TestCheckResourceAttrSet(resourceName, "created_time"),
			// 		resource.TestCheckResourceAttr(resourceName, "data_delivery.#", "1"),
			// 		resource.TestCheckResourceAttr(resourceName, "data_delivery.0.s3_destination.#", "1"),
			// 		resource.TestCheckResourceAttrPair(resourceName, "data_delivery.0.s3_destination.0.bucket", "aws_s3_bucket.test", "id"),
			// 		resource.TestCheckResourceAttr(resourceName, "data_delivery.0.s3_destination.0.prefix", "test"),
			// 		resource.TestCheckResourceAttr(resourceName, "description", description),
			// 		resource.TestCheckResourceAttrSet(resourceName, "experiment_count"),
			// 		resource.TestCheckResourceAttrSet(resourceName, "feature_count"),
			// 		resource.TestCheckResourceAttrSet(resourceName, "last_updated_time"),
			// 		resource.TestCheckResourceAttrSet(resourceName, "launch_count"),
			// 		resource.TestCheckResourceAttr(resourceName, "name", rName3),
			// 		resource.TestCheckResourceAttrSet(resourceName, "status"),
			// 		resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
			// 		resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Test Project"),
			// 	),
			// },
		},
	})
}

func TestAccProject_disappears(t *testing.T) {
	var project cloudwatchevidently.GetProjectOutput

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	description := "disappears"
	resourceName := "aws_cloudwatchevidently_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudwatchevidently.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_basic(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudwatchevidently.ResourceProject(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckProjectDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchEvidentlyConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatchevidently_project" {
			continue
		}

		input := &cloudwatchevidently.GetProjectInput{
			Project: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetProject(input)

		if err == nil {
			if aws.StringValue(resp.Project.Arn) == rs.Primary.ID {
				return fmt.Errorf("Project '%s' was not deleted properly", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckProjectExists(name string, project *cloudwatchevidently.GetProjectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchEvidentlyConn
		input := &cloudwatchevidently.GetProjectInput{
			Project: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetProject(input)

		if err != nil {
			return err
		}

		*project = *resp

		return nil
	}
}

func testAccProjectConfig_basic(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatchevidently_project" "test" {
  name        = %[1]q
  description = %[2]q

  tags = {
    "Key1" = "Test Project"
  }
}
`, rName, description)
}

func testAccProjectConfig_tags(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatchevidently_project" "test" {
  name        = %[1]q
  description = %[2]q

  tags = {
    "Key1" = "Test Project"
    "Key2" = "Value2a"
  }
}
`, rName, description)
}

func testAccProjectConfig_tagsUpdated(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatchevidently_project" "test" {
  name        = %[1]q
  description = %[2]q

  tags = {
    "Key1" = "Test Project"
    "Key2" = "Value2b"
    "Key3" = "Value3"
  }
}
`, rName, description)
}

func testAccProjectBaseConfig(rName, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_public_access_block" "test" {
  bucket                  = aws_s3_bucket.test.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[2]q
}
`, rName, rName2)
}

func testAccProjectConfig_dataDeliveryCloudWatchLogs(rName, rName2, rName3, description string) string {
	return acctest.ConfigCompose(
		testAccProjectBaseConfig(rName, rName2),
		fmt.Sprintf(`
resource "aws_cloudwatchevidently_project" "test" {
  name        = %[1]q
  description = %[2]q

  data_delivery {
    cloudwatch_logs {
      log_group = aws_cloudwatch_log_group.test.name
    }
  }

  tags = {
    "Key1" = "Test Project"
  }
}
`, rName3, description))
}

// func testAccProjectConfig_dataDeliveryS3Bucket(rName, rName2, rName3, description string) string {
// 	return acctest.ConfigCompose(
// 		testAccProjectBaseConfig(rName, rName2),
// 		fmt.Sprintf(`
// resource "aws_cloudwatchevidently_project" "test" {
//   name        = %[1]q
//   description = %[2]q

//   data_delivery {
//     s3_destination {
//       bucket = aws_s3_bucket.test.id
//       prefix = "test"
//     }
//   }

//   tags = {
//     "Key1" = "Test Project"
//   }
// }
// `, rName3, description))
// }
