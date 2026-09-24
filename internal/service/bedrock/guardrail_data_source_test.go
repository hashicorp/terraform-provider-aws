// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockGuardrailDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrock_guardrail.test"
	datasourceName := "data.aws_bedrock_guardrail.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "guardrail_arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "guardrail_arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "guardrail_id", resourceName, "guardrail_id"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, "blocked_input_messaging", resourceName, "blocked_input_messaging"),
					resource.TestCheckResourceAttrPair(datasourceName, "blocked_outputs_messaging", resourceName, "blocked_outputs_messaging"),
					resource.TestCheckResourceAttr(datasourceName, names.AttrVersion, "DRAFT"),
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttr(datasourceName, names.AttrStatus, "READY"),
				),
			},
		},
	})
}

func TestAccBedrockGuardrailDataSource_fetchLatest(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	datasourceName := "data.aws_bedrock_guardrail.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailDataSourceConfig_fetchLatest(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Three versions are created sequentially; latest=true must return the highest (3).
					resource.TestCheckResourceAttr(datasourceName, names.AttrVersion, "3"),
					resource.TestCheckResourceAttrSet(datasourceName, "arn"),
				),
			},
		},
	})
}

func TestAccBedrockGuardrailDataSource_fetchLatest_nonePublished(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccGuardrailDataSourceConfig_fetchLatest_nonePublished(rName),
				ExpectError: regexache.MustCompile(`latest is true but no published versions exist`),
			},
		},
	})
}

func TestAccBedrockGuardrailDataSource_fetchLatest_conflictsVersion(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccGuardrailDataSourceConfig_fetchLatest_conflictsVersion(rName),
				ExpectError: regexache.MustCompile(`Attribute "version" cannot be specified when "latest" is specified`),
			},
		},
	})
}

func testAccGuardrailDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccGuardrailConfig_basic(rName), `
data "aws_bedrock_guardrail" "test" {
  guardrail_identifier = aws_bedrock_guardrail.test.guardrail_id
}
`)
}

func testAccGuardrailDataSourceConfig_fetchLatest(rName string) string {
	return acctest.ConfigCompose(testAccGuardrailConfig_basic(rName), `
resource "aws_bedrock_guardrail_version" "v1" {
  guardrail_arn = aws_bedrock_guardrail.test.guardrail_arn
}

resource "aws_bedrock_guardrail_version" "v2" {
  guardrail_arn = aws_bedrock_guardrail.test.guardrail_arn
  depends_on    = [aws_bedrock_guardrail_version.v1]
}

resource "aws_bedrock_guardrail_version" "v3" {
  guardrail_arn = aws_bedrock_guardrail.test.guardrail_arn
  depends_on    = [aws_bedrock_guardrail_version.v2]
}

data "aws_bedrock_guardrail" "test" {
  guardrail_identifier = aws_bedrock_guardrail.test.guardrail_id
  latest               = true
  depends_on           = [aws_bedrock_guardrail_version.v3]
}
`)
}

func testAccGuardrailDataSourceConfig_fetchLatest_nonePublished(rName string) string {
	return acctest.ConfigCompose(testAccGuardrailConfig_basic(rName), `
data "aws_bedrock_guardrail" "test" {
  guardrail_identifier = aws_bedrock_guardrail.test.guardrail_id
  latest               = true
}
`)
}

func testAccGuardrailDataSourceConfig_fetchLatest_conflictsVersion(rName string) string {
	return acctest.ConfigCompose(testAccGuardrailConfig_basic(rName), `
data "aws_bedrock_guardrail" "test" {
  guardrail_identifier = aws_bedrock_guardrail.test.guardrail_id
  latest               = true
  version              = "1"
}
`)
}
