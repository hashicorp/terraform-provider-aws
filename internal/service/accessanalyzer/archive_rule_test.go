package accessanalyzer_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/accessanalyzer"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfaccessanalyzer "github.com/hashicorp/terraform-provider-aws/internal/service/accessanalyzer"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

//func TestArchiveRuleExampleUnitTest(t *testing.T) {
//	testCases := []struct {
//		TestName string
//		Input    string
//		Expected string
//		Error    bool
//	}{
//		{
//			TestName: "empty",
//			Input:    "",
//			Expected: "",
//			Error:    true,
//		},
//		{
//			TestName: "descriptive name",
//			Input:    "some input",
//			Expected: "some output",
//			Error:    false,
//		},
//		{
//			TestName: "another descriptive name",
//			Input:    "more input",
//			Expected: "more output",
//			Error:    false,
//		},
//	}
//
//	for _, testCase := range testCases {
//		t.Run(testCase.TestName, func(t *testing.T) {
//			got, err := tfaccessanalyzer.FunctionFromResource(testCase.Input)
//
//			if err != nil && !testCase.Error {
//				t.Errorf("got error (%s), expected no error", err)
//			}
//
//			if err == nil && testCase.Error {
//				t.Errorf("got (%s) and no error, expected error", got)
//			}
//
//			if got != testCase.Expected {
//				t.Errorf("got %s, expected %s", got, testCase.Expected)
//			}
//		})
//	}
//}

func TestAccAccessAnalyzerArchiveRule_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var archiveRule accessanalyzer.ArchiveRuleSummary
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_accessanalyzer_archiverule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(accessanalyzer.EndpointsID, t)
			testAccArchiveRulePreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, accessanalyzer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckArchiveRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccArchiveRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveRuleExists(resourceName, &archiveRule),
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
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAccessAnalyzerArchiveRule_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var archiveRule accessanalyzer.ArchiveRuleSummary
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_accessanalyzer_archiverule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(accessanalyzer.EndpointsID, t)
			testAccArchiveRulePreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, accessanalyzer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckArchiveRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccArchiveRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveRuleExists(resourceName, &archiveRule),
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

		analyzerName, ruleName, err := tfaccessanalyzer.DecodeRuleID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("unable to decode AccessAnalyzer ArchiveRule ID (%s): %s", rs.Primary.ID, err)
		}

		_, err = tfaccessanalyzer.FindArchiveRule(context.Background(), conn, analyzerName, ruleName)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Expected AccessAnalyzer ArchiveRule to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckArchiveRuleExists(name string, archiveRule *accessanalyzer.ArchiveRuleSummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No AccessAnalyzer ArchiveRule is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AccessAnalyzerConn
		analyzerName, ruleName, err := tfaccessanalyzer.DecodeRuleID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("unable to decode AccessAnalyzer ArchiveRule ID (%s): %s", rs.Primary.ID, err)
		}

		resp, err := tfaccessanalyzer.FindArchiveRule(context.Background(), conn, analyzerName, ruleName)

		if err != nil {
			return fmt.Errorf("Error describing AccessAnalyzer ArchiveRule: %s", err.Error())
		}

		*archiveRule = *resp

		return nil
	}
}

func testAccArchiveRulePreCheck(t *testing.T) {
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

func testAccArchiveRuleConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_accessanalyzer_archiverule" "test" {
  analyzer_name = aws_accessanalyzer_analyzer.test
  rule_name     = %[1]q

  filter {
    criteria = "error"
    exists   = true
  }

  filter {
    criteria = "isPublic"
    eq       = ["false"]
  }
}
`, rName)
}
