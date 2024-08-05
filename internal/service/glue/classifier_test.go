// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlueClassifier_csvClassifier(t *testing.T) {
	ctx := acctest.Context(t)
	var classifier awstypes.Classifier

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_classifier.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClassifierDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClassifierConfig_csv(rName, false, "PRESENT", "|", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(ctx, resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.allow_single_column", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.contains_header", "PRESENT"),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.disable_value_trimming", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.header.0", "header_column1"),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.header.1", "header_column2"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", acctest.Ct0),
				),
			},
			{
				Config: testAccClassifierConfig_csv(rName, false, "PRESENT", ",", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(ctx, resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.allow_single_column", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.contains_header", "PRESENT"),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.disable_value_trimming", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.header.0", "header_column1"),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.header.1", "header_column2"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", acctest.Ct0),
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

func TestAccGlueClassifier_csvClassifierCustomSerde(t *testing.T) {
	ctx := acctest.Context(t)
	var classifier awstypes.Classifier

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_classifier.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClassifierDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClassifierConfig_csvWithSerde(rName, false, "PRESENT", "|", false, "None"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(ctx, resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.serde", "None"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				Config: testAccClassifierConfig_csvWithSerde(rName, false, "PRESENT", ",", false, "OpenCSVSerDe"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(ctx, resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.serde", "OpenCSVSerDe"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
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
	ctx := acctest.Context(t)
	var classifier awstypes.Classifier

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_classifier.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClassifierDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClassifierConfig_csvQuoteSymbol(rName, "\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(ctx, resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.quote_symbol", "\""),
				),
			},
			{
				Config: testAccClassifierConfig_csvQuoteSymbol(rName, "'"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(ctx, resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", acctest.Ct1),
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

func TestAccGlueClassifier_CSVClassifier_custom(t *testing.T) {
	ctx := acctest.Context(t)
	var classifier awstypes.Classifier

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_classifier.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClassifierDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClassifierConfig_csvCustom(rName, "BINARY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(ctx, resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.custom_datatypes.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.custom_datatypes.0", "BINARY"),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.custom_datatypes.1", "SHORT"),
				),
			},
			{
				Config: testAccClassifierConfig_csvCustom(rName, "BOOLEAN"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(ctx, resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.custom_datatypes.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.custom_datatypes.0", "BOOLEAN"),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.0.custom_datatypes.1", "SHORT"),
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
	ctx := acctest.Context(t)
	var classifier awstypes.Classifier

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_classifier.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClassifierDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClassifierConfig_grok(rName, "classification1", "pattern1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(ctx, resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.classification", "classification1"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.custom_patterns", ""),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.grok_pattern", "pattern1"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", acctest.Ct0),
				),
			},
			{
				Config: testAccClassifierConfig_grok(rName, "classification2", "pattern2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(ctx, resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.classification", "classification2"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.custom_patterns", ""),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.grok_pattern", "pattern2"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", acctest.Ct0),
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
	ctx := acctest.Context(t)
	var classifier awstypes.Classifier

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_classifier.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClassifierDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClassifierConfig_grokCustomPatterns(rName, "custompattern1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(ctx, resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.classification", "classification"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.custom_patterns", "custompattern1"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.grok_pattern", "pattern"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", acctest.Ct0),
				),
			},
			{
				Config: testAccClassifierConfig_grokCustomPatterns(rName, "custompattern2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(ctx, resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.classification", "classification"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.custom_patterns", "custompattern2"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.grok_pattern", "pattern"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", acctest.Ct0),
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
	ctx := acctest.Context(t)
	var classifier awstypes.Classifier

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_classifier.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClassifierDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClassifierConfig_json(rName, "jsonpath1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(ctx, resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.0.json_path", "jsonpath1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", acctest.Ct0),
				),
			},
			{
				Config: testAccClassifierConfig_json(rName, "jsonpath2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(ctx, resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.0.json_path", "jsonpath2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", acctest.Ct0),
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
	ctx := acctest.Context(t)
	var classifier awstypes.Classifier

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_classifier.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClassifierDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClassifierConfig_grok(rName, "classification1", "pattern1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(ctx, resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.classification", "classification1"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.custom_patterns", ""),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.grok_pattern", "pattern1"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", acctest.Ct0),
				),
			},
			{
				Config: testAccClassifierConfig_json(rName, "jsonpath1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(ctx, resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.0.json_path", "jsonpath1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", acctest.Ct0),
				),
			},
			{
				Config: testAccClassifierConfig_xml(rName, "classification1", "rowtag1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(ctx, resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.0.classification", "classification1"),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.0.row_tag", "rowtag1"),
				),
			},
			{
				Config: testAccClassifierConfig_grok(rName, "classification1", "pattern1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(ctx, resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.classification", "classification1"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.custom_patterns", ""),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.grok_pattern", "pattern1"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccGlueClassifier_xmlClassifier(t *testing.T) {
	ctx := acctest.Context(t)
	var classifier awstypes.Classifier

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_classifier.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClassifierDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClassifierConfig_xml(rName, "classification1", "rowtag1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(ctx, resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.0.classification", "classification1"),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.0.row_tag", "rowtag1"),
				),
			},
			{
				Config: testAccClassifierConfig_xml(rName, "classification2", "rowtag2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(ctx, resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "csv_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var classifier awstypes.Classifier

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_classifier.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClassifierDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClassifierConfig_csv(rName, false, "PRESENT", "|", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(ctx, resourceName, &classifier),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfglue.ResourceClassifier(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckClassifierExists(ctx context.Context, resourceName string, classifier *awstypes.Classifier) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Glue Classifier ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

		output, err := tfglue.FindClassifierByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*classifier = *output
		return nil
	}
}

func testAccCheckClassifierDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_glue_classifier" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

			_, err := tfglue.FindClassifierByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Glue Classifier %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccClassifierConfig_csv(rName string, allowSingleColumn bool, containsHeader string, delimiter string, disableValueTrimming bool) string {
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

func testAccClassifierConfig_csvWithSerde(rName string, allowSingleColumn bool, containsHeader string, delimiter string, disableValueTrimming bool, serde string) string {
	return fmt.Sprintf(`
resource "aws_glue_classifier" "test" {
  name = %[1]q

  csv_classifier {
    allow_single_column    = %[2]t
    contains_header        = %[3]q
    delimiter              = %[4]q
    disable_value_trimming = %[5]t
    serde                  = %[6]q
    header                 = ["header_column1", "header_column2"]
  }
}
`, rName, allowSingleColumn, containsHeader, delimiter, disableValueTrimming, serde)
}

func testAccClassifierConfig_csvQuoteSymbol(rName, symbol string) string {
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

func testAccClassifierConfig_csvCustom(rName, customType string) string {
	return fmt.Sprintf(`
resource "aws_glue_classifier" "test" {
  name = %[1]q

  csv_classifier {
    allow_single_column        = false
    contains_header            = "PRESENT"
    delimiter                  = ","
    header                     = ["header_column1", "header_column2"]
    custom_datatype_configured = true
    custom_datatypes           = ["%[2]s", "SHORT"]
  }
}
`, rName, customType)
}

func testAccClassifierConfig_grok(rName, classification, grokPattern string) string {
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

func testAccClassifierConfig_grokCustomPatterns(rName, customPatterns string) string {
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

func testAccClassifierConfig_json(rName, jsonPath string) string {
	return fmt.Sprintf(`
resource "aws_glue_classifier" "test" {
  name = "%s"

  json_classifier {
    json_path = "%s"
  }
}
`, rName, jsonPath)
}

func testAccClassifierConfig_xml(rName, classification, rowTag string) string {
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
