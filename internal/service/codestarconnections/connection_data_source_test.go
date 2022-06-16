package codestarconnections_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/codestarconnections"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccCodeStarConnectionsConnectionDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_codestarconnections_connection.test_arn"
	dataSourceName2 := "data.aws_codestarconnections_connection.test_name"
	resourceName := "aws_codestarconnections_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codestarconnections.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, codestarconnections.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "provider_type", dataSourceName, "provider_type"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "connection_status", dataSourceName, "connection_status"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName2, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName2, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "provider_type", dataSourceName2, "provider_type"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName2, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "connection_status", dataSourceName2, "connection_status"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName2, "tags.%"),
				),
			},
		},
	})
}

func TestAccCodeStarConnectionsConnectionDataSource_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_codestarconnections_connection.test"
	resourceName := "aws_codestarconnections_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codestarconnections.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, codestarconnections.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionDataSourceConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccConnectionDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "Bitbucket"
}

data "aws_codestarconnections_connection" "test_arn" {
  arn = aws_codestarconnections_connection.test.arn
}

data "aws_codestarconnections_connection" "test_name" {
  name = aws_codestarconnections_connection.test.name
}
`, rName)
}

func testAccConnectionDataSourceConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "Bitbucket"

  tags = {
    "key1" = "value1"
    "key2" = "value2"
  }
}

data "aws_codestarconnections_connection" "test" {
  arn = aws_codestarconnections_connection.test.arn
}
`, rName)
}
