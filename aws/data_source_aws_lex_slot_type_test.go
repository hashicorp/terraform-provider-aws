package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
)

func TestAccDataSourceAwsLexSlotType_basic(t *testing.T) {
	rName := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	dataSourceName := "data.aws_lex_slot_type.test"
	resourceName := "aws_lex_slot_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t) },
		ErrorCheck: acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccAwsLexSlotTypeConfig_basic(rName),
					testAccDataSourceAwsLexSlotTypeConfig_basic(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "checksum", resourceName, "checksum"),
					resource.TestCheckResourceAttrPair(dataSourceName, "created_date", resourceName, "created_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enumeration_value.#", resourceName, "enumeration_value.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "last_updated_date", resourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "value_selection_strategy", resourceName, "value_selection_strategy"),
					resource.TestCheckResourceAttrPair(dataSourceName, "version", resourceName, "version"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsLexSlotType_withVersion(t *testing.T) {
	rName := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	dataSourceName := "data.aws_lex_slot_type.test"
	resourceName := "aws_lex_slot_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t) },
		ErrorCheck: acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccAwsLexSlotTypeConfig_withVersion(rName),
					testAccDataSourceAwsLexSlotTypeConfig_withVersion(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "checksum", resourceName, "checksum"),
					resource.TestCheckResourceAttrPair(dataSourceName, "created_date", resourceName, "created_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "last_updated_date", resourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "value_selection_strategy", resourceName, "value_selection_strategy"),
					resource.TestCheckResourceAttrPair(dataSourceName, "version", resourceName, "version"),
				),
			},
		},
	})
}

func testAccDataSourceAwsLexSlotTypeConfig_basic() string {
	return `
data "aws_lex_slot_type" "test" {
  name = aws_lex_slot_type.test.name
}
`
}

func testAccDataSourceAwsLexSlotTypeConfig_withVersion() string {
	return `
data "aws_lex_slot_type" "test" {
  name    = aws_lex_slot_type.test.name
  version = "1"
}
`
}
