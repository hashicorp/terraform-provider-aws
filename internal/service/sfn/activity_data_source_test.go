package sfn_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/sfn"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccSFNActivityDataSource_StepFunctions_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sfn_activity.test"
	dataName := "data.aws_sfn_activity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sfn.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActivityDataSourceConfig_checkActivityARN(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "creation_date", dataName, "creation_date"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataName, "name"),
				),
			},
			{
				Config: testAccActivityDataSourceConfig_checkActivityName(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "creation_date", dataName, "creation_date"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataName, "name"),
				),
			},
		},
	})
}

func testAccActivityDataSourceConfig_checkActivityARN(rName string) string {
	return fmt.Sprintf(`
resource aws_sfn_activity "test" {
  name = "%s"
}

data aws_sfn_activity "test" {
  arn = aws_sfn_activity.test.id
}
`, rName)
}

func testAccActivityDataSourceConfig_checkActivityName(rName string) string {
	return fmt.Sprintf(`
resource aws_sfn_activity "test" {
  name = "%s"
}

data aws_sfn_activity "test" {
  name = aws_sfn_activity.test.name
}
`, rName)
}
