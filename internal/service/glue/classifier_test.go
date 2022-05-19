package glue_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
)

func TestAccGlueClassifier_csvClassifier(t *testing.T) {
	var classifier glue.Classifier

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_classifier.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClassifierDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClassifierConfig_CSVClassifier(rName, false, "PRESENT", "|", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.allow_single_column", "false"),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.contains_header", "PRESENT"),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.disable_value_trimming", "false"),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.header.0", "header_column1"),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.header.1", "header_column2"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", "0"),
				),
			},
			{
				Config: testAccClassifierConfig_CSVClassifier(rName, false, "PRESENT", ",", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.allow_single_column", "false"),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.contains_header", "PRESENT"),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.disable_value_trimming", "false"),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.header.0", "header_column1"),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.header.1", "header_column2"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", "0"),
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

func TestAccGlueClassifier_CSVClassifier_quoteSymbol(t *testing.T) {
	var classifier glue.Classifier

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_classifier.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClassifierDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClassifierCSVClassifierQuoteSymbolConfig(rName, "\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.quote_symbol", "\""),
				),
			},
			{
				Config: testAccClassifierCSVClassifierQuoteSymbolConfig(rName, "'"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.quote_symbol", "'"),
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

func TestAccGlueClassifier_grokClassifier(t *testing.T) {
	var classifier glue.Classifier

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_classifier.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClassifierDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClassifierConfig_GrokClassifier(rName, "classification1", "pattern1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.classification", "classification1"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.custom_patterns", ""),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.grok_pattern", "pattern1"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", "0"),
				),
			},
			{
				Config: testAccClassifierConfig_GrokClassifier(rName, "classification2", "pattern2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.classification", "classification2"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.custom_patterns", ""),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.grok_pattern", "pattern2"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", "0"),
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

func TestAccGlueClassifier_GrokClassifier_customPatterns(t *testing.T) {
	var classifier glue.Classifier

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_classifier.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClassifierDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClassifierConfig_GrokClassifier_CustomPatterns(rName, "custompattern1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.classification", "classification"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.custom_patterns", "custompattern1"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.grok_pattern", "pattern"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", "0"),
				),
			},
			{
				Config: testAccClassifierConfig_GrokClassifier_CustomPatterns(rName, "custompattern2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.classification", "classification"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.custom_patterns", "custompattern2"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.grok_pattern", "pattern"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", "0"),
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

func TestAccGlueClassifier_jsonClassifier(t *testing.T) {
	var classifier glue.Classifier

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_classifier.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClassifierDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClassifierConfig_JSONClassifier(rName, "jsonpath1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.0.json_path", "jsonpath1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", "0"),
				),
			},
			{
				Config: testAccClassifierConfig_JSONClassifier(rName, "jsonpath2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.0.json_path", "jsonpath2"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", "0"),
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

func TestAccGlueClassifier_typeChange(t *testing.T) {
	var classifier glue.Classifier

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_classifier.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClassifierDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClassifierConfig_GrokClassifier(rName, "classification1", "pattern1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.classification", "classification1"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.custom_patterns", ""),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.grok_pattern", "pattern1"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", "0"),
				),
			},
			{
				Config: testAccClassifierConfig_JSONClassifier(rName, "jsonpath1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.0.json_path", "jsonpath1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", "0"),
				),
			},
			{
				Config: testAccClassifierConfig_XmlClassifier(rName, "classification1", "rowtag1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.0.classification", "classification1"),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.0.row_tag", "rowtag1"),
				),
			},
			{
				Config: testAccClassifierConfig_GrokClassifier(rName, "classification1", "pattern1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.classification", "classification1"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.custom_patterns", ""),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.grok_pattern", "pattern1"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", "0"),
				),
			},
		},
	})
}

func TestAccGlueClassifier_xmlClassifier(t *testing.T) {
	var classifier glue.Classifier

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_classifier.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClassifierDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClassifierConfig_XmlClassifier(rName, "classification1", "rowtag1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.0.classification", "classification1"),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.0.row_tag", "rowtag1"),
				),
			},
			{
				Config: testAccClassifierConfig_XmlClassifier(rName, "classification2", "rowtag2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.0.classification", "classification2"),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.0.row_tag", "rowtag2"),
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

func TestAccGlueClassifier_disappears(t *testing.T) {
	var classifier glue.Classifier

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_classifier.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClassifierDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClassifierConfig_CSVClassifier(rName, false, "PRESENT", "|", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(resourceName, &classifier),
					acctest.CheckResourceDisappears(acctest.Provider, tfglue.ResourceClassifier(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckClassifierExists(resourceName string, classifier *glue.Classifier) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Glue Classifier ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn

		output, err := conn.GetClassifier(&glue.GetClassifierInput{
			Name: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if output.Classifier == nil {
			return fmt.Errorf("Glue Classifier (%s) not found", rs.Primary.ID)
		}

		*classifier = *output.Classifier
		return nil
	}
}

func testAccCheckClassifierDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_classifier" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn

		output, err := conn.GetClassifier(&glue.GetClassifierInput{
			Name: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
				return nil
			}

		}

		classifier := output.Classifier
		if classifier != nil {
			return fmt.Errorf("Glue Classifier %s still exists", rs.Primary.ID)
		}

		return err
	}

	return nil
}

func testAccClassifierConfig_CSVClassifier(rName string, allowSingleColumn bool, containsHeader string, delimiter string, disableValueTrimming bool) string {
	return fmt.Sprintf(`
resource "aws_glue_classifier" "test" {
  name = "%s"

  csv_classifier {
    allow_single_column    = "%t"
    contains_header        = "%s"
    delimiter              = "%s"
    disable_value_trimming = "%t"
    header                 = ["header_column1", "header_column2"]
  }
}
`, rName, allowSingleColumn, containsHeader, delimiter, disableValueTrimming)
}

func testAccClassifierCSVClassifierQuoteSymbolConfig(rName, symbol string) string {
	return fmt.Sprintf(`
resource "aws_glue_classifier" "test" {
  name = %[1]q

  csv_classifier {
    allow_single_column = false
    contains_header     = "PRESENT"
    delimiter           = ","
    header              = ["header_column1", "header_column2"]
    quote_symbol        = %[2]q
  }
}
`, rName, symbol)
}

func testAccClassifierConfig_GrokClassifier(rName, classification, grokPattern string) string {
	return fmt.Sprintf(`
resource "aws_glue_classifier" "test" {
  name = "%s"

  grok_classifier {
    classification = "%s"
    grok_pattern   = "%s"
  }
}
`, rName, classification, grokPattern)
}

func testAccClassifierConfig_GrokClassifier_CustomPatterns(rName, customPatterns string) string {
	return fmt.Sprintf(`
resource "aws_glue_classifier" "test" {
  name = "%s"

  grok_classifier {
    classification  = "classification"
    custom_patterns = "%s"
    grok_pattern    = "pattern"
  }
}
`, rName, customPatterns)
}

func testAccClassifierConfig_JSONClassifier(rName, jsonPath string) string {
	return fmt.Sprintf(`
resource "aws_glue_classifier" "test" {
  name = "%s"

  json_classifier {
    json_path = "%s"
  }
}
`, rName, jsonPath)
}

func testAccClassifierConfig_XmlClassifier(rName, classification, rowTag string) string {
	return fmt.Sprintf(`
resource "aws_glue_classifier" "test" {
  name = "%s"

  xml_classifier {
    classification = "%s"
    row_tag        = "%s"
  }
}
`, rName, classification, rowTag)
}
