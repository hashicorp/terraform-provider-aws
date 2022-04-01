package cloudwatchlogs_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudwatchlogs "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatchlogs"
)

func TestAccCloudWatchLogsGroup_basic(t *testing.T) {
	var lg cloudwatchlogs.LogGroup
	rInt := sdkacctest.RandInt()
	resourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "logs", fmt.Sprintf("log-group:foo-bar-%d", rInt)),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("foo-bar-%d", rInt)),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccCloudWatchLogsGroup_namePrefix(t *testing.T) {
	var lg cloudwatchlogs.LogGroup
	resourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroup_namePrefix,
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

func TestAccCloudWatchLogsGroup_NamePrefix_retention(t *testing.T) {
	var lg cloudwatchlogs.LogGroup
	rName := sdkacctest.RandString(5)
	resourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroup_namePrefix_retention(rName, 365),
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
				Config: testAccGroup_namePrefix_retention(rName, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile("^tf-test-")),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "7"),
				),
			},
		},
	})
}

func TestAccCloudWatchLogsGroup_generatedName(t *testing.T) {
	var lg cloudwatchlogs.LogGroup
	resourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroup_generatedName,
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

func TestAccCloudWatchLogsGroup_retentionPolicy(t *testing.T) {
	var lg cloudwatchlogs.LogGroup
	rInt := sdkacctest.RandInt()
	resourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_withRetention(rInt),
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
				Config: testAccGroupModifiedConfig_withRetention(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "0"),
				),
			},
		},
	})
}

func TestAccCloudWatchLogsGroup_multiple(t *testing.T) {
	var lg cloudwatchlogs.LogGroup
	rInt := sdkacctest.RandInt()
	resourceName := "aws_cloudwatch_log_group.alpha"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_multiple(rInt),
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

func TestAccCloudWatchLogsGroup_disappears(t *testing.T) {
	var lg cloudwatchlogs.LogGroup
	rInt := sdkacctest.RandInt()
	resourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					testAccCheckCloudWatchLogGroupDisappears(&lg),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudWatchLogsGroup_tagging(t *testing.T) {
	var lg cloudwatchlogs.LogGroup
	rInt := sdkacctest.RandInt()
	resourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupWithTagsConfig(rInt),
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
				Config: testAccGroupWithTagsAddedConfig(rInt),
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
				Config: testAccGroupWithTagsUpdatedConfig(rInt),
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
				Config: testAccGroupWithTagsConfig(rInt),
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

func TestAccCloudWatchLogsGroup_kmsKey(t *testing.T) {
	var lg cloudwatchlogs.LogGroup
	rInt := sdkacctest.RandInt()
	resourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig(rInt),
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
				Config: testAccGroupWithKMSKeyIDConfig(rInt),
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
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchLogsConn
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchLogsConn
		logGroup, err := tfcloudwatchlogs.LookupGroup(conn, rs.Primary.ID)
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

func testAccCheckGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchLogsConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_log_group" {
			continue
		}
		logGroup, err := tfcloudwatchlogs.LookupGroup(conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error reading CloudWatch Log Group (%s): %w", rs.Primary.ID, err)
		}

		if logGroup != nil {
			return fmt.Errorf("Bad: LogGroup still exists: %q", rs.Primary.ID)
		}

	}

	return nil
}

func testAccGroupConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = "foo-bar-%d"
}
`, rInt)
}

func testAccGroupWithTagsConfig(rInt int) string {
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

func testAccGroupWithTagsAddedConfig(rInt int) string {
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

func testAccGroupWithTagsUpdatedConfig(rInt int) string {
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

func testAccGroupConfig_withRetention(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name              = "foo-bar-%d"
  retention_in_days = 365
}
`, rInt)
}

func testAccGroupModifiedConfig_withRetention(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = "foo-bar-%d"
}
`, rInt)
}

func testAccGroupConfig_multiple(rInt int) string {
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

func testAccGroupWithKMSKeyIDConfig(rInt int) string {
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
  kms_key_id = aws_kms_key.foo.arn
}
`, rInt, rInt)
}

const testAccGroup_namePrefix = `
resource "aws_cloudwatch_log_group" "test" {
  name_prefix = "tf-test-"
}
`

func testAccGroup_namePrefix_retention(rName string, retention int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name_prefix       = "tf-test-%s"
  retention_in_days = %d
}
`, rName, retention)
}

const testAccGroup_generatedName = `
resource "aws_cloudwatch_log_group" "test" {}
`
