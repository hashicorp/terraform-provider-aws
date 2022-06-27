package accessanalyzer_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/accessanalyzer"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccAnalyzer_basic(t *testing.T) {
	var analyzer accessanalyzer.AnalyzerSummary

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_accessanalyzer_analyzer.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, accessanalyzer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAnalyzerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAnalyzerConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalyzerExists(resourceName, &analyzer),
					resource.TestCheckResourceAttr(resourceName, "analyzer_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "access-analyzer", fmt.Sprintf("analyzer/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", accessanalyzer.TypeAccount),
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

func testAccAnalyzer_disappears(t *testing.T) {
	var analyzer accessanalyzer.AnalyzerSummary

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_accessanalyzer_analyzer.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, accessanalyzer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAnalyzerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAnalyzerConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalyzerExists(resourceName, &analyzer),
					testAccCheckAnalyzerDisappears(&analyzer),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAnalyzer_Tags(t *testing.T) {
	var analyzer accessanalyzer.AnalyzerSummary

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_accessanalyzer_analyzer.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, accessanalyzer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAnalyzerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAnalyzerConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalyzerExists(resourceName, &analyzer),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAnalyzerConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalyzerExists(resourceName, &analyzer),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAnalyzerConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalyzerExists(resourceName, &analyzer),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccAnalyzer_Type_Organization(t *testing.T) {
	var analyzer accessanalyzer.AnalyzerSummary

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_accessanalyzer_analyzer.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
			acctest.PreCheckOrganizationsAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, accessanalyzer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAnalyzerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAnalyzerConfig_typeOrganization(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalyzerExists(resourceName, &analyzer),
					resource.TestCheckResourceAttr(resourceName, "type", accessanalyzer.TypeOrganization),
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

func testAccCheckAnalyzerDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AccessAnalyzerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_accessanalyzer_analyzer" {
			continue
		}

		input := &accessanalyzer.GetAnalyzerInput{
			AnalyzerName: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetAnalyzer(input)

		if tfawserr.ErrCodeEquals(err, accessanalyzer.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil {
			return fmt.Errorf("Access Analyzer Analyzer (%s) still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCheckAnalyzerDisappears(analyzer *accessanalyzer.AnalyzerSummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AccessAnalyzerConn

		input := &accessanalyzer.DeleteAnalyzerInput{
			AnalyzerName: analyzer.Name,
		}

		_, err := conn.DeleteAnalyzer(input)

		return err
	}
}

func testAccCheckAnalyzerExists(resourceName string, analyzer *accessanalyzer.AnalyzerSummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AccessAnalyzerConn

		input := &accessanalyzer.GetAnalyzerInput{
			AnalyzerName: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetAnalyzer(input)

		if err != nil {
			return err
		}

		*analyzer = *output.Analyzer

		return nil
	}
}

func testAccAnalyzerConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_accessanalyzer_analyzer" "test" {
  analyzer_name = %[1]q
}
`, rName)
}

func testAccAnalyzerConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_accessanalyzer_analyzer" "test" {
  analyzer_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAnalyzerConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_accessanalyzer_analyzer" "test" {
  analyzer_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAnalyzerConfig_typeOrganization(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["access-analyzer.${data.aws_partition.current.dns_suffix}"]
}

resource "aws_accessanalyzer_analyzer" "test" {
  depends_on = [aws_organizations_organization.test]

  analyzer_name = %[1]q
  type          = "ORGANIZATION"
}
`, rName)
}
