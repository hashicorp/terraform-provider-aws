package logs_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
)

func TestAccLogsStream_basic(t *testing.T) {
	var ls cloudwatchlogs.LogStream
	resourceName := "aws_cloudwatch_log_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &ls),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccStreamImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLogsStream_disappears(t *testing.T) {
	var ls cloudwatchlogs.LogStream
	resourceName := "aws_cloudwatch_log_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &ls),
					acctest.CheckResourceDisappears(acctest.Provider, tflogs.ResourceStream(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLogsStream_Disappears_logGroup(t *testing.T) {
	var ls cloudwatchlogs.LogStream
	var lg cloudwatchlogs.LogGroup
	resourceName := "aws_cloudwatch_log_stream.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(resourceName, &ls),
					testAccCheckGroupExists(logGroupResourceName, &lg),
					acctest.CheckResourceDisappears(acctest.Provider, tflogs.ResourceGroup(), logGroupResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckStreamExists(n string, ls *cloudwatchlogs.LogStream) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		logGroupName := rs.Primary.Attributes["log_group_name"]
		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsConn
		logGroup, exists, err := tflogs.LookupStream(conn, rs.Primary.ID, logGroupName, nil)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Bad: LogStream %q does not exist", rs.Primary.ID)
		}

		*ls = *logGroup

		return nil
	}
}

func testAccCheckStreamDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LogsConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_log_stream" {
			continue
		}

		logGroupName := rs.Primary.Attributes["log_group_name"]
		_, exists, err := tflogs.LookupStream(conn, rs.Primary.ID, logGroupName, nil)

		if tfawserr.ErrCodeEquals(err, cloudwatchlogs.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading CloudWatch Log Stream (%s): %w", rs.Primary.ID, err)
		}

		if exists {
			return fmt.Errorf("Bad: LogStream still exists: %q", rs.Primary.ID)
		}

	}

	return nil
}

func testAccStreamImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["log_group_name"], rs.Primary.ID), nil
	}
}

func TestValidateStreamName(t *testing.T) {
	validNames := []string{
		"test-log-stream",
		"my_sample_log_stream",
		"012345678",
		"logstream/1234",
	}
	for _, v := range validNames {
		_, errors := tflogs.ValidStreamName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid CloudWatch LogStream name: %q", v, errors)
		}
	}

	invalidNames := []string{
		sdkacctest.RandString(513),
		"",
		"stringwith:colon",
	}
	for _, v := range invalidNames {
		_, errors := tflogs.ValidStreamName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid CloudWatch LogStream name", v)
		}
	}
}

func testAccStreamConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_stream" "test" {
  name           = %[1]q
  log_group_name = aws_cloudwatch_log_group.test.id
}
`, rName)
}
