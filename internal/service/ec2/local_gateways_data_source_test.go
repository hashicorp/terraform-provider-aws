package ec2_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2LocalGatewaysDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_ec2_local_gateways.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccLocalGatewaysDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "ids.#", "0"),
				),
			},
		},
	})
}

func testAccLocalGatewaysDataSourceConfig() string {
	return `
data "aws_ec2_local_gateways" "test" {}
`
}
