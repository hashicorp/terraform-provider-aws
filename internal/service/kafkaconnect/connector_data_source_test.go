// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kafkaconnect_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKafkaConnectConnectorDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_connector.test"
	dataSourceName := "data.aws_mskconnect_connector.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.KafkaConnectEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDescription, dataSourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVersion, dataSourceName, names.AttrVersion),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTags, dataSourceName, names.AttrTags),
				),
			},
		},
	})
}

func testAccConnectorDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccConnectorConfig_basic(rName), `
data "aws_mskconnect_connector" "test" {
  name = aws_mskconnect_connector.test.name
}
`)
}
