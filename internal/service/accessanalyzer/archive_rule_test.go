package accessanalyzer_test
// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.
//
// Remember to register this new resource in the provider
// (internal/provider/provider.go) once you finish. Otherwise, Terraform won't
// know about it.

import (
	// TIP: ==== IMPORT ====
    // This is a common set of imports but not customized to your code
	// since your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/accessanalyzer"
	"github.com/aws/aws-sdk-go-v2/service/accessanalyzer/types" // TIP: Some v2 packages use a separate package for types while some do not.
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"

	// TIP: You will often need to import the package that this test file lives
    // in. Since it is in the "test" context, it must import the package to use
    // any normal context constants, variables, or functions.
	tfaccessanalyzer "github.com/hashicorp/terraform-provider-aws/internal/service/accessanalyzer"
)

// TIP: File Structure. The basic outline for all test files should be as
// follows. Improve this resource's maintainability by following this
// outline.
//
// 1. Package declaration (add "_test" since this is a test file)
// 2. Imports
// 3. Unit tests
// 4. Basic test
// 5. Disappears test
// 6. All the other tests
// 7. Helper functions (exists, destroy, check, etc.)
// 8. Functions that return Terraform configurations


// TIP: ==== UNIT TESTS ====
// This is an example of a unit test. Its name is not prefixed with
// "TestAcc" like an acceptance test.
//
// Unlike acceptance tests, unit tests do not access AWS and are focused on a
// function (or method). Because of this, they are quick and cheap to run.
//
// In designing a resource's implementation, isolate complex bits from AWS bits
// so that they can be tested through a unit test. We encourage more unit tests
// in the provider.
//
// Cut and dry functions using well-used patterns, like typical flatteners and
// expanders, don't need unit testing. However, if they are complex or
// intricate, they should be unit tested.
func TestArchiveRuleExampleUnitTest(t *testing.T) {
	testCases := []struct {
		TestName string
		Input    string
		Expected string
		Error    bool
	}{
		{
			TestName: "empty",
			Input:    "",
			Expected: "",
			Error:    true,
		},
		{
			TestName: "descriptive name",
			Input:    "some input",
			Expected: "some output",
			Error:    false,
		},
		{
			TestName: "another descriptive name",
			Input:    "more input",
			Expected: "more output",
			Error:    false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			got, err := tfaccessanalyzer.FunctionFromResource(testCase.Input)

			if err != nil && !testCase.Error {
				t.Errorf("got error (%s), expected no error", err)
			}

			if err == nil && testCase.Error {
				t.Errorf("got (%s) and no error, expected error", got)
			}

			if got != testCase.Expected {
				t.Errorf("got %s, expected %s", got, testCase.Expected)
			}
		})
	}
}


// TIP: ==== ACCEPTANCE TESTS ====
// This is an example of a basic acceptance test. This should test as much of
// standard functionality of the resource as possible, and test importing, if
// applicable. We prefix its name with "TestAcc", the service, and the
// resource name.
//
// Acceptance test access AWS and cost money to run.
func TestAccAccessAnalyzerArchiveRule_basic(t *testing.T) {
    // TIP: This is a long-running test guard for tests that run longer than
    // 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var archiverule accessanalyzer.DescribeArchiveRuleResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_accessanalyzer_archiverule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(accessanalyzer.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, accessanalyzer.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckArchiveRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccArchiveRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveRuleExists(resourceName, &archiverule),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "accessanalyzer", regexp.MustCompile(`archiverule:+.`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccAccessAnalyzerArchiveRule_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var archiverule accessanalyzer.DescribeArchiveRuleResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_accessanalyzer_archiverule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(accessanalyzer.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, accessanalyzer.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckArchiveRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccArchiveRuleConfig_basic(rName, testAccArchiveRuleVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveRuleExists(resourceName, &archiverule),
					acctest.CheckResourceDisappears(acctest.Provider, tfaccessanalyzer.ResourceArchiveRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckArchiveRuleDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AccessAnalyzerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_accessanalyzer_archiverule" {
			continue
		}

		input := &accessanalyzer.DescribeArchiveRuleInput{
			ArchiveRuleId: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeArchiveRule(input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, accessanalyzer.ErrCodeNotFoundException) {
				return nil
			}
			return err
		}

		return fmt.Errorf("Expected AccessAnalyzer ArchiveRule to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckArchiveRuleExists(name string, archiverule *accessanalyzer.DescribeArchiveRuleResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No AccessAnalyzer ArchiveRule is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AccessAnalyzerConn
		resp, err := conn.DescribeArchiveRule(&accessanalyzer.DescribeArchiveRuleInput{
			ArchiveRuleId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return fmt.Errorf("Error describing AccessAnalyzer ArchiveRule: %s", err.Error())
		}

		*archiverule = *resp

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AccessAnalyzerConn

	input := &accessanalyzer.ListArchiveRulesInput{}

	_, err := conn.ListArchiveRules(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckArchiveRuleNotRecreated(before, after *accessanalyzer.DescribeArchiveRuleResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.ArchiveRuleId), aws.StringValue(after.ArchiveRuleId); before != after {
			return fmt.Errorf("AccessAnalyzer ArchiveRule (%s/%s) recreated", before, after)
		}

		return nil
	}
}

func testAccArchiveRuleConfig_basic(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_accessanalyzer_archiverule" "test" {
  archiverule_name             = %[1]q
  engine_type             = "ActiveAccessAnalyzer"
  engine_version          = %[2]q
  host_instance_type      = "accessanalyzer.t2.micro"
  security_groups         = [aws_security_group.test.id]
  authentication_strategy = "simple"
  storage_type            = "efs"

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }
}
`, rName, version)
}
