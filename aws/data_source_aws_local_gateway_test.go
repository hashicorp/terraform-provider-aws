package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataSourceAwsLocalGateway_basic(t *testing.T) {
	dsResourceName := "data.aws_local_gateway.by_id"

	localGatewayId := os.Getenv("AWS_LOCAL_GATEWAY_ID")
	if localGatewayId == "" {
		t.Skip(
			"Environment variable AWS_LOCAL_GATEWAY_ID is not set. " +
				"This environment variable must be set to the ID of " +
				"a deployed Local Gateway to enable this test.")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsLocalGatewayConfig(localGatewayId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dsResourceName, "id", localGatewayId),
					testAccCheckResourceAttrAccountID(dsResourceName, "owner_id"),
					resource.TestCheckResourceAttrSet(dsResourceName, "state"),
					resource.TestCheckResourceAttrSet(dsResourceName, "outpost_arn"),
				),
			},
		},
	})
}

func testAccDataSourceAwsLocalGatewayConfig(localGatewayId string) string {
	return fmt.Sprintf(`
data "aws_local_gateway" "by_id" {
  id = "%s"
}
`, localGatewayId)
}
