package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSCloudWatchLogGroup_basic(t *testing.T) {
	var lg cloudwatchlogs.LogGroup
	rInt := acctest.RandInt()
	resourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchLogGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchLogGroupConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_in_days"}, //this has a default value
			},
		},
	})
}

func TestAccAWSCloudWatchLogGroup_namePrefix(t *testing.T) {
	var lg cloudwatchlogs.LogGroup
	resourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchLogGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchLogGroup_namePrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile("^tf-test-")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_in_days", "name_prefix"},
			},
		},
	})
}

func TestAccAWSCloudWatchLogGroup_namePrefix_retention(t *testing.T) {
	var lg cloudwatchlogs.LogGroup
	rName := acctest.RandString(5)
	resourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchLogGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchLogGroup_namePrefix_retention(rName, 365),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile("^tf-test-")),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "365"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_in_days", "name_prefix"},
			},
			{
				Config: testAccAWSCloudWatchLogGroup_namePrefix_retention(rName, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile("^tf-test-")),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "7"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchLogGroup_generatedName(t *testing.T) {
	var lg cloudwatchlogs.LogGroup
	resourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchLogGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchLogGroup_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_in_days"},
			},
		},
	})
}

func TestAccAWSCloudWatchLogGroup_retentionPolicy(t *testing.T) {
	var lg cloudwatchlogs.LogGroup
	rInt := acctest.RandInt()
	resourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchLogGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchLogGroupConfig_withRetention(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "365"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_in_days"},
			},
			{
				Config: testAccAWSCloudWatchLogGroupConfigModified_withRetention(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "0"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchLogGroup_multiple(t *testing.T) {
	var lg cloudwatchlogs.LogGroup
	rInt := acctest.RandInt()
	resourceName := "aws_cloudwatch_log_group.alpha"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchLogGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchLogGroupConfig_multiple(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists("aws_cloudwatch_log_group.alpha", &lg),
					resource.TestCheckResourceAttr("aws_cloudwatch_log_group.alpha", "retention_in_days", "14"),
					testAccCheckCloudWatchLogGroupExists("aws_cloudwatch_log_group.beta", &lg),
					resource.TestCheckResourceAttr("aws_cloudwatch_log_group.beta", "retention_in_days", "0"),
					testAccCheckCloudWatchLogGroupExists("aws_cloudwatch_log_group.charlie", &lg),
					resource.TestCheckResourceAttr("aws_cloudwatch_log_group.charlie", "retention_in_days", "3653"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_in_days"},
			},
		},
	})
}

func TestAccAWSCloudWatchLogGroup_disappears(t *testing.T) {
	var lg cloudwatchlogs.LogGroup
	rInt := acctest.RandInt()
	resourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchLogGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchLogGroupConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					testAccCheckCloudWatchLogGroupDisappears(&lg),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCloudWatchLogGroup_tagging(t *testing.T) {
	var lg cloudwatchlogs.LogGroup
	rInt := acctest.RandInt()
	resourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchLogGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchLogGroupConfigWithTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "Production"),
					resource.TestCheckResourceAttr(resourceName, "tags.Foo", "Bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.Empty", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_in_days"},
			},
			{
				Config: testAccAWSCloudWatchLogGroupConfigWithTagsAdded(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "Development"),
					resource.TestCheckResourceAttr(resourceName, "tags.Foo", "Bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.Empty", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.Bar", "baz"),
				),
			},
			{
				Config: testAccAWSCloudWatchLogGroupConfigWithTagsUpdated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "Development"),
					resource.TestCheckResourceAttr(resourceName, "tags.Empty", "NotEmpty"),
					resource.TestCheckResourceAttr(resourceName, "tags.Foo", "UpdatedBar"),
					resource.TestCheckResourceAttr(resourceName, "tags.Bar", "baz"),
				),
			},
			{
				Config: testAccAWSCloudWatchLogGroupConfigWithTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "Production"),
					resource.TestCheckResourceAttr(resourceName, "tags.Foo", "Bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.Empty", ""),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchLogGroup_kmsKey(t *testing.T) {
	var lg cloudwatchlogs.LogGroup
	rInt := acctest.RandInt()
	resourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchLogGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchLogGroupConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_in_days"},
			},
			{
				Config: testAccAWSCloudWatchLogGroupConfigWithKmsKeyId(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					resource.TestCheckResourceAttrSet(resourceName, "kms_key_id"),
				),
			},
		},
	})
}

func testAccCheckCloudWatchLogGroupDisappears(lg *cloudwatchlogs.LogGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cloudwatchlogsconn
		opts := &cloudwatchlogs.DeleteLogGroupInput{
			LogGroupName: lg.LogGroupName,
		}
		_, err := conn.DeleteLogGroup(opts)
		return err
	}
}

func testAccCheckCloudWatchLogGroupExists(n string, lg *cloudwatchlogs.LogGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudwatchlogsconn
		logGroup, err := lookupCloudWatchLogGroup(conn, rs.Primary.ID)
		if err != nil {
			return err
		}
		if logGroup == nil {
			return fmt.Errorf("Bad: LogGroup %q does not exist", rs.Primary.ID)
		}

		*lg = *logGroup

		return nil
	}
}

func testAccCheckAWSCloudWatchLogGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudwatchlogsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_log_group" {
			continue
		}
		logGroup, err := lookupCloudWatchLogGroup(conn, rs.Primary.ID)
		if err != nil {
			return nil
		}

		if logGroup != nil {
			return fmt.Errorf("Bad: LogGroup still exists: %q", rs.Primary.ID)
		}

	}

	return nil
}

func testAccAWSCloudWatchLogGroupConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = "foo-bar-%d"
}
`, rInt)
}

func testAccAWSCloudWatchLogGroupConfigWithTags(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = "foo-bar-%d"

  tags = {
    Environment = "Production"
    Foo         = "Bar"
    Empty       = ""
  }
}
`, rInt)
}

func testAccAWSCloudWatchLogGroupConfigWithTagsAdded(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = "foo-bar-%d"

  tags = {
    Environment = "Development"
    Foo         = "Bar"
    Empty       = ""
    Bar         = "baz"
  }
}
`, rInt)
}

func testAccAWSCloudWatchLogGroupConfigWithTagsUpdated(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = "foo-bar-%d"

  tags = {
    Environment = "Development"
    Foo         = "UpdatedBar"
    Empty       = "NotEmpty"
    Bar         = "baz"
  }
}
`, rInt)
}

func testAccAWSCloudWatchLogGroupConfig_withRetention(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name              = "foo-bar-%d"
  retention_in_days = 365
}
`, rInt)
}

func testAccAWSCloudWatchLogGroupConfigModified_withRetention(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = "foo-bar-%d"
}
`, rInt)
}

func testAccAWSCloudWatchLogGroupConfig_multiple(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "alpha" {
  name              = "foo-bar-%d"
  retention_in_days = 14
}

resource "aws_cloudwatch_log_group" "beta" {
  name = "foo-bar-%d"
}

resource "aws_cloudwatch_log_group" "charlie" {
  name              = "foo-bar-%d"
  retention_in_days = 3653
}
`, rInt, rInt+1, rInt+2)
}

func testAccAWSCloudWatchLogGroupConfigWithKmsKeyId(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "foo" {
  description             = "Terraform acc test %d"
  deletion_window_in_days = 7

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_cloudwatch_log_group" "test" {
  name       = "foo-bar-%d"
  kms_key_id = "${aws_kms_key.foo.arn}"
}
`, rInt, rInt)
}

const testAccAWSCloudWatchLogGroup_namePrefix = `
resource "aws_cloudwatch_log_group" "test" {
    name_prefix = "tf-test-"
}
`

func testAccAWSCloudWatchLogGroup_namePrefix_retention(rName string, retention int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name_prefix       = "tf-test-%s"
  retention_in_days = %d
}
`, rName, retention)
}

const testAccAWSCloudWatchLogGroup_generatedName = `
resource "aws_cloudwatch_log_group" "test" {}
`
