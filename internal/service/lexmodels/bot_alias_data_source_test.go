package lexmodels_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccBotAliasDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	dataSourceName := "data.aws_lex_bot_alias.test"
	resourceName := "aws_lex_bot_alias.test"

	// If this test runs in parallel with other Lex Bot tests, it loses its description
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(rName),
					testAccBotConfig_createVersion(rName),
					testAccBotAliasConfig_basic(rName),
					testAccBotAliasDataSourceConfig_basic(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bot_name", resourceName, "bot_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bot_version", resourceName, "bot_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "checksum", resourceName, "checksum"),
					resource.TestCheckResourceAttrPair(dataSourceName, "created_date", resourceName, "created_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "last_updated_date", resourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
				),
			},
		},
	})
}

func testAccBotAliasDataSourceConfig_basic() string {
	return `
data "aws_lex_bot_alias" "test" {
  name     = aws_lex_bot_alias.test.name
  bot_name = aws_lex_bot.test.name
}
`
}
