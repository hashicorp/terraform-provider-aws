// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexmodels_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLexModelsBotDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	dataSourceName := "data.aws_lex_bot.test"
	resourceName := "aws_lex_bot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, lexmodelbuildingservice.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(rName),
					testAccBotConfig_basic(rName),
					testAccBotDataSourceConfig_basic(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "checksum", resourceName, "checksum"),
					resource.TestCheckResourceAttrPair(dataSourceName, "child_directed", resourceName, "child_directed"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrCreatedDate, resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, "detect_sentiment", resourceName, "detect_sentiment"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enable_model_improvements", resourceName, "enable_model_improvements"),
					resource.TestCheckResourceAttrPair(dataSourceName, "failure_reason", resourceName, "failure_reason"),
					resource.TestCheckResourceAttrPair(dataSourceName, "idle_session_ttl_in_seconds", resourceName, "idle_session_ttl_in_seconds"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrLastUpdatedDate, resourceName, names.AttrLastUpdatedDate),
					resource.TestCheckResourceAttrPair(dataSourceName, "locale", resourceName, "locale"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "nlu_intent_confidence_threshold", resourceName, "nlu_intent_confidence_threshold"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStatus, resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVersion, resourceName, names.AttrVersion),
				),
			},
		},
	})
}

func testAccBotDataSource_withVersion(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	dataSourceName := "data.aws_lex_bot.test"
	resourceName := "aws_lex_bot.test"

	// If this test runs in parallel with other Lex Bot tests, it loses its description
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, lexmodelbuildingservice.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(rName),
					testAccBotConfig_createVersion(rName),
					testAccBotDataSourceConfig_withVersion(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "checksum", resourceName, "checksum"),
					resource.TestCheckResourceAttrPair(dataSourceName, "child_directed", resourceName, "child_directed"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, "detect_sentiment", resourceName, "detect_sentiment"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enable_model_improvements", resourceName, "enable_model_improvements"),
					resource.TestCheckResourceAttrPair(dataSourceName, "failure_reason", resourceName, "failure_reason"),
					resource.TestCheckResourceAttrPair(dataSourceName, "idle_session_ttl_in_seconds", resourceName, "idle_session_ttl_in_seconds"),
					//Removed due to race condition when bots build/read/not_ready. Modified the bot status to read true status:
					//https://github.com/hashicorp/terraform-provider-aws/issues/21107
					//resource.TestCheckResourceAttrPair(dataSourceName, "created_date", resourceName, "created_date"),
					//resource.TestCheckResourceAttrPair(dataSourceName, "last_updated_date", resourceName, "last_updated_date"),
					//resource.TestCheckResourceAttrPair(dataSourceName, "status", resourceName, "status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "locale", resourceName, "locale"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "nlu_intent_confidence_threshold", resourceName, "nlu_intent_confidence_threshold"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVersion, resourceName, names.AttrVersion),
				),
			},
		},
	})
}

func testAccBotDataSourceConfig_basic() string {
	return `
data "aws_lex_bot" "test" {
  name = aws_lex_bot.test.name
}
`
}

func testAccBotDataSourceConfig_withVersion() string {
	return `
data "aws_lex_bot" "test" {
  name    = aws_lex_bot.test.name
  version = "1"
}
`
}
