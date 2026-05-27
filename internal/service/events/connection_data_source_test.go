// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package events_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEventsConnectionDataSource_Connection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_cloudwatch_event_connection.test"
	resourceName := "aws_cloudwatch_event_connection.api_key"

	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	authorizationType := "API_KEY"
	description := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	key := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	value := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectDataSourceConfig_basic(
					name,
					description,
					authorizationType,
					key,
					value,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "secret_arn", resourceName, "secret_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "authorization_type", resourceName, "authorization_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kms_key_identifier", resourceName, "kms_key_identifier"),
				),
			},
		},
	})
}

func testAccConnectDataSourceConfig_basic(name, description, authorizationType, key, value string) string {
	return acctest.ConfigCompose(
		testAccConnectionConfig_apiKey(name, description, authorizationType, key, value),
		`
data "aws_cloudwatch_event_connection" "test" {
  name = aws_cloudwatch_event_connection.api_key.name
}
`)
}
