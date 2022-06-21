package accessanalyzer_test

import (
	"context"
	"fmt"
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

func testAccAccessAnalyzerArchiveRule_basic(t *testing.T) {
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
		},
		ErrorCheck:        acctest.ErrorCheck(t, accessanalyzer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckArchiveRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccArchiveRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveRuleExists(resourceName, &archiveRule),
					resource.TestCheckResourceAttr(resourceName, "filter.0.criteria", "isPublic"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.exists", "false"),
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

func testAccAccessAnalyzerArchiveRule_disappears(t *testing.T) {
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

func testAccArchiveRuleBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_accessanalyzer_analyzer" "test" {
  analyzer_name = %[1]q
}

`, rName)
}

func testAccArchiveRuleConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccArchiveRuleBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_accessanalyzer_archiverule" "test" {
  analyzer_name = aws_accessanalyzer_analyzer.test.analyzer_name
  rule_name     = %[1]q

  filter {
    criteria = "isPublic"
    eq       = ["false"]
  }
}
`, rName))
}
