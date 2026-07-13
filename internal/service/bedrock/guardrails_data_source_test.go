// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockGuardrailsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	datasourceName := "data.aws_bedrock_guardrails.test"
	resourceName := "aws_bedrock_guardrail.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrID),
					resource.TestCheckResourceAttr(datasourceName, "guardrails.#", "1"),
					resource.TestCheckResourceAttrPair(datasourceName, "guardrails.0.arn", resourceName, "guardrail_arn"),
					resource.TestCheckResourceAttrSet(datasourceName, "guardrails.0.guardrail_id"),
					resource.TestCheckResourceAttrSet(datasourceName, "guardrails.0.name"),
					resource.TestCheckResourceAttrSet(datasourceName, "guardrails.0.status"),
					resource.TestCheckResourceAttrSet(datasourceName, "guardrails.0.version"),
					resource.TestCheckResourceAttrSet(datasourceName, "guardrails.0.created_at"),
					resource.TestCheckResourceAttrSet(datasourceName, "guardrails.0.updated_at"),
				),
			},
		},
	})
}

func TestAccBedrockGuardrailsDataSource_noFilter(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	datasourceName := "data.aws_bedrock_guardrails.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailsDataSourceConfig_noFilter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrID),
					acctest.CheckResourceAttrGreaterThanValue(datasourceName, "guardrails.#", 0),
				),
			},
		},
	})
}

func TestAccBedrockGuardrailsDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	datasourceName := "data.aws_bedrock_guardrails.test"
	tagKey := "key"
	tagVal := "value"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailsDataSourceConfig_tags(rName, tagKey, tagVal),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrID),
					acctest.CheckResourceAttrGreaterThanValue(datasourceName, "guardrails.#", 0),
					resource.TestCheckResourceAttr(datasourceName, "guardrails.0.tags.%", "1"),
					resource.TestCheckResourceAttr(datasourceName, "guardrails.0.tags."+tagKey, tagVal),
				),
			},
		},
	})
}

func TestAccBedrockGuardrailsDataSource_guardrailIdentifier(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	datasourceName := "data.aws_bedrock_guardrails.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailsDataSourceConfig_guardrailIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrID),
					resource.TestCheckResourceAttr(datasourceName, "guardrails.#", "2"),
				),
			},
		},
	})
}

func TestAccBedrockGuardrailsDataSource_guardrailIdentifierARN(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	datasourceName := "data.aws_bedrock_guardrails.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailsDataSourceConfig_guardrailIdentifierARN(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrID),
					resource.TestCheckResourceAttr(datasourceName, "guardrails.#", "2"),
				),
			},
		},
	})
}

func TestAccBedrockGuardrailsDataSource_nameRegex(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	datasourceName := "data.aws_bedrock_guardrails.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailsDataSourceConfig_nameRegex(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "guardrails.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "guardrails.0.name", rName),
				),
			},
		},
	})
}

func TestAccBedrockGuardrailsDataSource_nameRegex_conflictsWithGuardrailIdentifier(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccGuardrailsDataSourceConfig_nameRegexAndGuardrailIdentifier(),
				ExpectError: regexache.MustCompile(`.`),
			},
		},
	})
}

func TestAccBedrockGuardrailsDataSource_guardrailIdentifier_invalidFormat(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccGuardrailsDataSourceConfig_guardrailIdentifierRaw(`"UPPERCASE-ID"`),
				ExpectError: regexache.MustCompile(`.`),
			},
			{
				Config:      testAccGuardrailsDataSourceConfig_guardrailIdentifierRaw(`"abc-def"`),
				ExpectError: regexache.MustCompile(`.`),
			},
			{
				Config:      testAccGuardrailsDataSourceConfig_guardrailIdentifierRaw(`"arn:not-valid-arn-format"`),
				ExpectError: regexache.MustCompile(`.`),
			},
		},
	})
}

func testAccGuardrailsDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccGuardrailConfig_basic(rName),
		fmt.Sprintf(`
data "aws_bedrock_guardrails" "test" {
  name_regex = "^%[1]s$"

  depends_on = [aws_bedrock_guardrail.test]
}
`, rName),
	)
}

func testAccGuardrailsDataSourceConfig_noFilter(rName string) string {
	return acctest.ConfigCompose(
		testAccGuardrailConfig_basic(rName),
		`
data "aws_bedrock_guardrails" "test" {
  depends_on = [aws_bedrock_guardrail.test]
}
`)
}

func testAccGuardrailsDataSourceConfig_guardrailIdentifier(rName string) string {
	return acctest.ConfigCompose(
		testAccGuardrailConfig_basic(rName),
		fmt.Sprintf(`
resource "aws_bedrock_guardrail_version" "test" {
  description   = %[1]q
  guardrail_arn = aws_bedrock_guardrail.test.guardrail_arn
}

data "aws_bedrock_guardrails" "test" {
  guardrail_identifier = aws_bedrock_guardrail.test.guardrail_id

  depends_on = [aws_bedrock_guardrail_version.test]
}
`, rName),
	)
}

func testAccGuardrailsDataSourceConfig_guardrailIdentifierARN(rName string) string {
	return acctest.ConfigCompose(
		testAccGuardrailConfig_basic(rName),
		fmt.Sprintf(`
resource "aws_bedrock_guardrail_version" "test" {
  description   = %[1]q
  guardrail_arn = aws_bedrock_guardrail.test.guardrail_arn
}

data "aws_bedrock_guardrails" "test" {
  guardrail_identifier = aws_bedrock_guardrail.test.guardrail_arn

  depends_on = [aws_bedrock_guardrail_version.test]
}
`, rName),
	)
}

func testAccGuardrailsDataSourceConfig_guardrailIdentifierRaw(identifier string) string {
	return fmt.Sprintf(`
data "aws_bedrock_guardrails" "test" {
  guardrail_identifier = %s
}
`, identifier)
}

func testAccGuardrailsDataSourceConfig_nameRegexAndGuardrailIdentifier() string {
	return `
data "aws_bedrock_guardrails" "test" {
  guardrail_identifier = "abc123"
  name_regex           = "^my-guardrail"
}
`
}

func testAccGuardrailsDataSourceConfig_nameRegex(rName string) string {
	return acctest.ConfigCompose(
		testAccGuardrailConfig_basic(rName),
		fmt.Sprintf(`
data "aws_bedrock_guardrails" "test" {
  name_regex = "^%[1]s$"

  depends_on = [aws_bedrock_guardrail.test]
}
`, rName),
	)
}

func testAccGuardrailsDataSourceConfig_tags(rName, tagKey, tagVal string) string {
	return fmt.Sprintf(`
resource "aws_bedrock_guardrail" "test" {
  name                      = %[1]q
  blocked_input_messaging   = "test"
  blocked_outputs_messaging = "test"

  content_policy_config {
    filters_config {
      input_strength  = "MEDIUM"
      output_strength = "MEDIUM"
      type            = "HATE"
    }
    filters_config {
      input_strength  = "HIGH"
      output_strength = "HIGH"
      type            = "VIOLENCE"
    }
  }

  contextual_grounding_policy_config {
    filters_config {
      threshold = 0.4
      type      = "GROUNDING"
    }
  }

  sensitive_information_policy_config {
    pii_entities_config {
      action = "BLOCK"
      type   = "NAME"
    }
    pii_entities_config {
      action = "BLOCK"
      type   = "DRIVER_ID"
    }
    pii_entities_config {
      action = "ANONYMIZE"
      type   = "USERNAME"
    }
    regexes_config {
      action      = "BLOCK"
      description = "example regex"
      name        = "regex_example"
      pattern     = "^\\d{3}-\\d{2}-\\d{4}$"
    }
  }

  topic_policy_config {
    topics_config {
      name       = "investment_topic"
      examples   = ["Where should I invest my money?"]
      type       = "DENY"
      definition = "Investment advice refers to inquiries, guidance, or recommendations regarding the management or allocation of funds or assets with the goal of generating returns."
    }
  }

  word_policy_config {
    managed_word_lists_config {
      type = "PROFANITY"
    }
    words_config {
      text = "self-assured"
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}

data "aws_bedrock_guardrails" "test" {
  name_regex = "^%[1]s$"

  depends_on = [aws_bedrock_guardrail.test]
}
`, rName, tagKey, tagVal)
}
